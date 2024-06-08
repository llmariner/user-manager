package server

import (
	"context"
	"errors"
	"strings"

	gerrors "github.com/llm-operator/common/pkg/gormlib/errors"
	"github.com/llm-operator/common/pkg/id"
	"github.com/llm-operator/rbac-manager/pkg/auth"
	v1 "github.com/llm-operator/user-manager/api/v1"
	"github.com/llm-operator/user-manager/server/internal/config"
	"github.com/llm-operator/user-manager/server/internal/store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

// CreateOrganization creates a new organization.
func (s *S) CreateOrganization(ctx context.Context, req *v1.CreateOrganizationRequest) (*v1.Organization, error) {
	userInfo, err := s.extractUserInfoFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.Title == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}

	isAllowed, err := s.canCreateOrganization(userInfo)
	if err != nil {
		return nil, err
	}
	if !isAllowed {
		return nil, status.Error(codes.PermissionDenied, "user is not allowed to create an organization")
	}

	org, err := s.createOrganization(ctx, req.Title, false, userInfo.TenantID)
	if err != nil {
		if gerrors.IsUniqueConstraintViolation(err) {
			return nil, status.Errorf(codes.AlreadyExists, "organizatione %q already exists", req.Title)
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Add a creator as an owner. Othewise, there is no owner in the org, and no one can access.
	if _, err := s.store.CreateOrganizationUser(org.OrganizationID, userInfo.UserID, v1.OrganizationRole_ORGANIZATION_ROLE_OWNER.String()); err != nil {
		return nil, err
	}

	return org.ToProto(), nil
}

// canCreateOrganization checks if the user can create an organization.
// Currently it checks if the user is the owner of the default organization.
//
// TODO(kenji): Should we have some more restriction?
func (s *S) canCreateOrganization(userInfo *auth.UserInfo) (bool, error) {
	if !s.enableAuth {
		return true, nil
	}

	org, err := s.store.GetDefaultOrganization(userInfo.TenantID)
	if err != nil {
		return false, status.Errorf(codes.Internal, "get default organizations: %s", err)
	}
	return s.organizationRole(org.OrganizationID, userInfo.UserID) == v1.OrganizationRole_ORGANIZATION_ROLE_OWNER, nil
}

func (s *S) createOrganization(ctx context.Context, title string, isDefault bool, tenantID string) (*store.Organization, error) {
	orgID, err := id.GenerateID("org-", 24)
	if err != nil {
		return nil, err
	}
	org, err := s.store.CreateOrganization(tenantID, orgID, title, isDefault)
	if err != nil {
		return nil, err
	}
	return org, nil
}

// ListOrganizations lists all organizations.
func (s *S) ListOrganizations(ctx context.Context, req *v1.ListOrganizationsRequest) (*v1.ListOrganizationsResponse, error) {
	userInfo, err := s.extractUserInfoFromContext(ctx)
	if err != nil {
		return nil, err
	}

	orgs, err := s.store.ListOrganizations(userInfo.TenantID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list organizations: %s", err)
	}

	// Only show orgs that the user is a owner/reader of.
	var filtered []*store.Organization
	for _, org := range orgs {
		if s.organizationRole(org.OrganizationID, userInfo.UserID) != v1.OrganizationRole_ORGANIZATION_ROLE_UNSPECIFIED {
			filtered = append(filtered, org)
		}
	}

	var orgProtos []*v1.Organization
	for _, org := range filtered {
		orgProtos = append(orgProtos, org.ToProto())
	}
	return &v1.ListOrganizationsResponse{
		Organizations: orgProtos,
	}, nil
}

// DeleteOrganization deletes an organization.
func (s *S) DeleteOrganization(ctx context.Context, req *v1.DeleteOrganizationRequest) (*v1.DeleteOrganizationResponse, error) {
	userInfo, err := s.extractUserInfoFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "organization id is required")
	}

	o, err := s.validateOrganizationID(req.Id, userInfo.TenantID)
	if err != nil {
		return nil, err
	}

	if err := s.validateOrganizationOwner(req.Id, userInfo.UserID); err != nil {
		return nil, err
	}

	if o.IsDefault {
		return nil, status.Errorf(codes.InvalidArgument, "cannot delete a default org")
	}

	// Check if the org still has a project.
	ps, err := s.store.ListProjectsByTenantIDAndOrganizationID(userInfo.TenantID, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list projects: %s", err)
	}
	if len(ps) > 0 {
		var s []string
		for _, p := range ps {
			s = append(s, p.Title)
		}
		return nil, status.Errorf(codes.FailedPrecondition, "organization %q still has projects: %q", req.Id, strings.Join(s, ", "))
	}

	if err := s.store.DeleteOrganization(userInfo.TenantID, req.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "delete organization: %s", err)
	}

	return &v1.DeleteOrganizationResponse{
		Id:      req.Id,
		Object:  "organization",
		Deleted: true,
	}, nil
}

// CreateOrganizationUser adds a user to an organization.
func (s *S) CreateOrganizationUser(ctx context.Context, req *v1.CreateOrganizationUserRequest) (*v1.OrganizationUser, error) {
	userInfo, err := s.extractUserInfoFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.OrganizationId == "" {
		return nil, status.Error(codes.InvalidArgument, "organization id is required")
	}
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}
	if req.Role == v1.OrganizationRole_ORGANIZATION_ROLE_UNSPECIFIED {
		return nil, status.Error(codes.InvalidArgument, "role is required")
	}

	if _, err := s.validateOrganizationID(req.OrganizationId, userInfo.TenantID); err != nil {
		return nil, err
	}

	if err := s.validateOrganizationOwner(req.OrganizationId, userInfo.UserID); err != nil {
		return nil, err
	}

	ou, err := s.store.CreateOrganizationUser(req.OrganizationId, req.UserId, req.Role.String())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.FailedPrecondition, "organization %q not found", req.OrganizationId)
		}
		if gerrors.IsUniqueConstraintViolation(err) {
			return nil, status.Errorf(codes.AlreadyExists, "user %q is already a member of organization %qs", req.UserId, req.OrganizationId)
		}
		return nil, status.Errorf(codes.Internal, "add user to organization: %s", err)
	}

	return ou.ToProto(), nil
}

// ListOrganizationUsers lists organization users for the specified organization.
func (s *S) ListOrganizationUsers(ctx context.Context, req *v1.ListOrganizationUsersRequest) (*v1.ListOrganizationUsersResponse, error) {
	userInfo, err := s.extractUserInfoFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.OrganizationId == "" {
		return nil, status.Error(codes.InvalidArgument, "organization id is required")
	}

	if _, err := s.validateOrganizationID(req.OrganizationId, userInfo.TenantID); err != nil {
		return nil, err
	}

	if err := s.validateOrganizationOwner(req.OrganizationId, userInfo.UserID); err != nil {
		return nil, err
	}

	users, err := s.store.ListOrganizationUsersByOrganizationID(req.OrganizationId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list organization users: %s", err)
	}

	var userProtos []*v1.OrganizationUser
	for _, user := range users {
		userProtos = append(userProtos, user.ToProto())
	}
	return &v1.ListOrganizationUsersResponse{
		Users: userProtos,
	}, nil
}

// DeleteOrganizationUser deletes an organization user.
func (s *S) DeleteOrganizationUser(ctx context.Context, req *v1.DeleteOrganizationUserRequest) (*emptypb.Empty, error) {
	userInfo, err := s.extractUserInfoFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.OrganizationId == "" {
		return nil, status.Error(codes.InvalidArgument, "organization id is required")
	}
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}

	if _, err := s.validateOrganizationID(req.OrganizationId, userInfo.TenantID); err != nil {
		return nil, err
	}

	if err := s.validateOrganizationOwner(req.OrganizationId, userInfo.UserID); err != nil {
		return nil, err
	}

	// TODO(kenji): Validate the user ID.

	// TODO(kenji): Delete all records in a single transaction.

	// Delete the user from all projects in the organization as well as from the organization.
	projects, err := s.store.ListProjectsByTenantIDAndOrganizationID(userInfo.TenantID, req.OrganizationId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list projects: %s", err)
	}
	for _, p := range projects {
		if err := s.store.DeleteProjectUser(p.ProjectID, req.UserId); err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, status.Errorf(codes.Internal, "delete project user: %s", err)
			}
			// Ignore.
		}
	}

	if err := s.store.DeleteOrganizationUser(req.OrganizationId, req.UserId); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "organization user %q not found", req.UserId)
		}
		return nil, status.Errorf(codes.Internal, "delete organization user: %s", err)
	}

	return &emptypb.Empty{}, nil
}

func (s *S) validateOrganizationID(orgID, tenantID string) (*store.Organization, error) {
	o, err := s.store.GetOrganizationByTenantIDAndOrgID(tenantID, orgID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.FailedPrecondition, "organization %q not found", orgID)
		}
		return nil, status.Errorf(codes.Internal, "get organization: %s", err)
	}

	return o, nil
}

// CreateDefaultOrganization creates the default org.
// TODO(kenji): This is not the best place for this function as there is nothing related to
// the server itself.
func (s *S) CreateDefaultOrganization(ctx context.Context, c *config.DefaultOrganizationConfig) (*store.Organization, error) {
	existing, err := s.store.GetDefaultOrganization(c.TenantID)
	if err == nil {
		// Do nothing.
		return existing, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	org, err := s.createOrganization(ctx, c.Title, true, c.TenantID)
	if err != nil {
		return nil, err
	}

	for _, uid := range c.UserIDs {
		if _, err := s.store.CreateOrganizationUser(org.OrganizationID, uid, v1.OrganizationRole_ORGANIZATION_ROLE_OWNER.String()); err != nil {
			return nil, err
		}
	}
	return org, nil
}

// ListInternalOrganizations lists all organizations.
func (s *IS) ListInternalOrganizations(ctx context.Context, req *v1.ListInternalOrganizationsRequest) (*v1.ListInternalOrganizationsResponse, error) {
	orgs, err := s.store.ListAllOrganizations()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list organizations: %s", err)
	}

	var orgProtos []*v1.InternalOrganization
	for _, org := range orgs {
		orgProtos = append(orgProtos, &v1.InternalOrganization{
			Organization: org.ToProto(),
			TenantId:     org.TenantID,
		})
	}
	return &v1.ListInternalOrganizationsResponse{
		Organizations: orgProtos,
	}, nil
}

// ListOrganizationUsers lists organization users for all organizations.
func (s *IS) ListOrganizationUsers(ctx context.Context, req *v1.ListOrganizationUsersRequest) (*v1.ListOrganizationUsersResponse, error) {
	users, err := s.store.ListAllOrganizationUsers()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list organization users: %s", err)
	}

	var userProtos []*v1.OrganizationUser
	for _, user := range users {
		userProtos = append(userProtos, user.ToProto())
	}
	return &v1.ListOrganizationUsersResponse{
		Users: userProtos,
	}, nil
}
