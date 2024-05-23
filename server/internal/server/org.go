package server

import (
	"context"
	"errors"
	"log"
	"strings"

	gerrors "github.com/llm-operator/common/pkg/gormlib/errors"
	"github.com/llm-operator/common/pkg/id"
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
	if req.Title == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}

	org, err := s.createOrganization(ctx, req.Title, false)
	if err != nil {
		if gerrors.IsUniqueConstraintViolation(err) {
			return nil, status.Errorf(codes.AlreadyExists, "organizatione %q already exists", req.Title)
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return org.ToProto(), nil
}

func (s *S) createOrganization(ctx context.Context, title string, isDefault bool) (*store.Organization, error) {
	orgID, err := id.GenerateID("org-", 24)
	if err != nil {
		return nil, err
	}
	org, err := s.store.CreateOrganization(fakeTenantID, orgID, title, isDefault)
	if err != nil {
		return nil, err
	}
	return org, nil
}

// ListOrganizations lists all organizations.
func (s *S) ListOrganizations(ctx context.Context, req *v1.ListOrganizationsRequest) (*v1.ListOrganizationsResponse, error) {
	orgs, err := s.store.ListOrganizations(fakeTenantID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list organizations: %s", err)
	}

	var orgProtos []*v1.Organization
	for _, org := range orgs {
		orgProtos = append(orgProtos, org.ToProto())
	}
	return &v1.ListOrganizationsResponse{
		Organizations: orgProtos,
	}, nil
}

// DeleteOrganization deletes an organization.
func (s *S) DeleteOrganization(ctx context.Context, req *v1.DeleteOrganizationRequest) (*v1.DeleteOrganizationResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "organization id is required")
	}

	o, err := s.validateOrgID(req.Id)
	if err != nil {
		return nil, err
	}
	if o.IsDefault {
		return nil, status.Errorf(codes.InvalidArgument, "cannot delete a default org")
	}

	// CHeck if the org still has a project.
	ps, err := s.store.ListProjectsByTenantIDAndOrganizationID(fakeTenantID, req.Id)
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

	if err := s.store.DeleteOrganization(fakeTenantID, req.Id); err != nil {
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
	if req.OrganizationId == "" {
		return nil, status.Error(codes.InvalidArgument, "organization id is required")
	}
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}
	if req.Role == v1.OrganizationRole_ORGANIZATION_ROLE_UNSPECIFIED {
		return nil, status.Error(codes.InvalidArgument, "role is required")
	}

	if _, err := s.validateOrgID(req.OrganizationId); err != nil {
		return nil, err
	}

	ou, err := s.store.CreateOrganizationUser(req.OrganizationId, req.UserId, req.Role.String())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.FailedPrecondition, "organization %q not found", req.OrganizationId)
		}
		if gerrors.IsUniqueConstraintViolation(err) {
			return nil, status.Errorf(codes.AlreadyExists, "user %q is already a member of organizatione %qs", req.UserId, req.OrganizationId)
		}
		return nil, status.Errorf(codes.Internal, "add user to organization: %s", err)
	}

	return ou.ToProto(), nil
}

// ListOrganizationUsers lists organization users for the specified organization.
func (s *S) ListOrganizationUsers(ctx context.Context, req *v1.ListOrganizationUsersRequest) (*v1.ListOrganizationUsersResponse, error) {
	if req.OrganizationId == "" {
		return nil, status.Error(codes.InvalidArgument, "organization id is required")
	}

	if _, err := s.validateOrgID(req.OrganizationId); err != nil {
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
	if req.OrganizationId == "" {
		return nil, status.Error(codes.InvalidArgument, "organization id is required")
	}
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}

	if _, err := s.validateOrgID(req.OrganizationId); err != nil {
		return nil, err
	}

	// TODO(kenji): Validate the user ID.

	if err := s.store.DeleteOrganizationUser(req.OrganizationId, req.UserId); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "organization user not found")
		}
		return nil, status.Errorf(codes.Internal, "delete organization user: %s", err)
	}

	return &emptypb.Empty{}, nil
}

func (s *S) validateOrgID(orgID string) (*store.Organization, error) {
	o, err := s.store.GetOrganizationByTenantIDAndOrgID(fakeTenantID, orgID)
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
	existing, err := s.store.GetDefaultOrganization(fakeTenantID)
	if err == nil {
		// Do nothing.
		return existing, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	log.Printf("Creating default org %q", c.Title)
	org, err := s.createOrganization(ctx, c.Title, true)
	if err != nil {
		return nil, err
	}

	for _, uid := range c.UserIDs {
		if _, err := s.CreateOrganizationUser(ctx, &v1.CreateOrganizationUserRequest{
			OrganizationId: org.OrganizationID,
			UserId:         uid,
			Role:           v1.OrganizationRole_ORGANIZATION_ROLE_OWNER,
		}); err != nil {
			return nil, err
		}
	}
	return org, nil
}

// ListOrganizations lists all organizations.
func (s *IS) ListOrganizations(ctx context.Context, req *v1.ListOrganizationsRequest) (*v1.ListOrganizationsResponse, error) {
	orgs, err := s.store.ListAllOrganizations()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list organizations: %s", err)
	}

	var orgProtos []*v1.Organization
	for _, org := range orgs {
		orgProtos = append(orgProtos, org.ToProto())
	}
	return &v1.ListOrganizationsResponse{
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
