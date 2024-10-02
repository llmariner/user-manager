package server

import (
	"context"
	"errors"
	"fmt"
	"strings"

	gerrors "github.com/llmariner/common/pkg/gormlib/errors"
	"github.com/llmariner/common/pkg/id"
	"github.com/llmariner/rbac-manager/pkg/auth"
	v1 "github.com/llmariner/user-manager/api/v1"
	"github.com/llmariner/user-manager/pkg/userid"
	"github.com/llmariner/user-manager/server/internal/config"
	"github.com/llmariner/user-manager/server/internal/store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

// CreateOrganization creates a new organization.
func (s *S) CreateOrganization(ctx context.Context, req *v1.CreateOrganizationRequest) (*v1.Organization, error) {
	userInfo, ok := auth.ExtractUserInfoFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("failed to extract user info from context")
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

	// Create a new organization. Add a creator as an owner. Othewise, there is no owner in the org, and no one can access.

	org, err := s.createOrganization(ctx, req.Title, false, userInfo.TenantID, []string{userid.Normalize(userInfo.UserID)})
	if err != nil {
		if gerrors.IsUniqueConstraintViolation(err) {
			return nil, status.Errorf(codes.AlreadyExists, "organizatione %q already exists", req.Title)
		}
		return nil, status.Error(codes.Internal, err.Error())
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

func (s *S) createOrganization(ctx context.Context, title string, isDefault bool, tenantID string, userIDs []string) (*store.Organization, error) {
	orgID, err := id.GenerateID("org-", 24)
	if err != nil {
		return nil, err
	}

	var org *store.Organization
	if err := s.store.Transaction(func(tx *gorm.DB) error {
		org, err = store.CreateOrganizationInTransaction(tx, tenantID, orgID, title, isDefault)
		if err != nil {
			return err
		}

		for _, uid := range userIDs {
			if _, err := findOrCreateUserInTransaction(tx, uid); err != nil {
				return err
			}
			if _, err := store.CreateOrganizationUserInTransaction(tx, org.OrganizationID, uid, v1.OrganizationRole_ORGANIZATION_ROLE_OWNER.String()); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return org, nil
}

// ListOrganizations lists all organizations.
func (s *S) ListOrganizations(ctx context.Context, req *v1.ListOrganizationsRequest) (*v1.ListOrganizationsResponse, error) {
	userInfo, ok := auth.ExtractUserInfoFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("failed to extract user info from context")
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
	userInfo, ok := auth.ExtractUserInfoFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("failed to extract user info from context")
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

	// TODO(kenji): There is a slight chance that a new project is created just after the check above. Handle such a race condition.

	if err := s.store.Transaction(func(tx *gorm.DB) error {
		if err := store.DeleteOrganizationInTransaction(tx, req.Id); err != nil {
			return fmt.Errorf("delete organization: %s", err)
		}

		if err := store.DeleteAllOrganizationUsersInTransaction(tx, req.Id); err != nil {
			return fmt.Errorf("delete all organization users: %s", err)
		}
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "transaction: %s", err)
	}

	return &v1.DeleteOrganizationResponse{
		Id:      req.Id,
		Object:  "organization",
		Deleted: true,
	}, nil
}

// CreateOrganizationUser adds a user to an organization.
func (s *S) CreateOrganizationUser(ctx context.Context, req *v1.CreateOrganizationUserRequest) (*v1.OrganizationUser, error) {
	userInfo, ok := auth.ExtractUserInfoFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("failed to extract user info from context")
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

	userID := userid.Normalize(req.UserId)
	var ou *store.OrganizationUser
	var err error
	if err := s.store.Transaction(func(tx *gorm.DB) error {
		if _, err = findOrCreateUserInTransaction(tx, userID); err != nil {
			return status.Errorf(codes.Internal, "create new user: %s", err)
		}

		ou, err = store.CreateOrganizationUserInTransaction(tx, req.OrganizationId, userID, req.Role.String())
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return status.Errorf(codes.FailedPrecondition, "organization %q not found", req.OrganizationId)
			}
			if gerrors.IsUniqueConstraintViolation(err) {
				return status.Errorf(codes.AlreadyExists, "user %q is already a member of organization %qs", userID, req.OrganizationId)
			}
			return status.Errorf(codes.Internal, "add user to organization: %s", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	// Do not populate the internal User ID for non-internal RPC.
	return ou.ToProto(""), nil
}

// ListOrganizationUsers lists organization users for the specified organization.
func (s *S) ListOrganizationUsers(ctx context.Context, req *v1.ListOrganizationUsersRequest) (*v1.ListOrganizationUsersResponse, error) {
	userInfo, ok := auth.ExtractUserInfoFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("failed to extract user info from context")
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
		// Do not populate the internal User ID for non-internal RPC.
		userProtos = append(userProtos, user.ToProto(""))
	}
	return &v1.ListOrganizationUsersResponse{
		Users: userProtos,
	}, nil
}

// DeleteOrganizationUser deletes an organization user.
func (s *S) DeleteOrganizationUser(ctx context.Context, req *v1.DeleteOrganizationUserRequest) (*emptypb.Empty, error) {
	userInfo, ok := auth.ExtractUserInfoFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("failed to extract user info from context")
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

	userID := userid.Normalize(req.UserId)
	if _, err := s.store.GetOrganizationUser(req.OrganizationId, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "organization user %q not found", userID)
		}
		return nil, status.Errorf(codes.Internal, "get organization user: %s", err)
	}

	// TODO(kenji): Validate the user ID.

	// Delete the user from all projects in the organization as well as from the organization.
	projects, err := s.store.ListProjectsByTenantIDAndOrganizationID(userInfo.TenantID, req.OrganizationId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list projects: %s", err)
	}

	if err := s.store.Transaction(func(tx *gorm.DB) error {
		for _, p := range projects {
			if err := store.DeleteProjectUserInTransaction(tx, p.ProjectID, userID); err != nil {
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					return fmt.Errorf("delete project user: %s", err)
				}
				// Ignore.
			}
		}

		if err := store.DeleteOrganizationUserInTransaction(tx, req.OrganizationId, userID); err != nil {
			return fmt.Errorf("delete organization user: %s", err)
		}
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "transaction: %s", err)
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
	org, err := s.createOrganization(ctx, c.Title, true, c.TenantID, c.UserIDs)
	if err != nil {
		return nil, err
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
	ous, err := s.store.ListAllOrganizationUsers()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list organization users: %s", err)
	}

	us, err := s.store.ListAllUsers()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list users: %s", err)
	}
	internalUserIDs := map[string]string{}
	for _, u := range us {
		internalUserIDs[u.UserID] = u.InternalUserID
	}

	var userProtos []*v1.OrganizationUser
	for _, ou := range ous {
		id, ok := internalUserIDs[ou.UserID]
		if !ok {
			return nil, status.Errorf(codes.Internal, "internal user ID not found for user %q", ou.UserID)
		}
		userProtos = append(userProtos, ou.ToProto(id))
	}

	return &v1.ListOrganizationUsersResponse{
		Users: userProtos,
	}, nil
}

func findOrCreateUserInTransaction(tx *gorm.DB, userID string) (*store.User, error) {
	internalUserID, err := id.GenerateID("user-", 24)
	if err != nil {
		return nil, err
	}
	u, err := store.FindOrCreateUserInTransaction(tx, userID, internalUserID)
	if err != nil {
		return nil, err
	}
	return u, nil
}
