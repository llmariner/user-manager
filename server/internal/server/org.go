package server

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"log"
	"strings"

	v1 "github.com/llm-operator/user-manager/api/v1"
	"github.com/llm-operator/user-manager/server/internal/config"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/api/validation"
)

// CreateOrganization creates a new organization.
func (s *S) CreateOrganization(ctx context.Context, req *v1.CreateOrganizationRequest) (*v1.Organization, error) {
	if req.Title == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}
	if req.KubernetesNamespace == "" {
		return nil, status.Error(codes.InvalidArgument, "kubernetes namespace is required")
	}

	if errs := validation.ValidateNamespaceName(req.KubernetesNamespace, false); len(errs) != 0 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid kubernetes namespace: %s", errs)
	}

	orgID, err := generateRandomString("org-", 22)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "generate organization id: %s", err)
	}
	org, err := s.store.CreateOrganization(fakeTenantID, orgID, req.Title, req.KubernetesNamespace)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "create organization: %s", err)
	}

	return org.ToProto(), nil
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

	if err := s.store.DeleteOrganization(fakeTenantID, req.Id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "organization %q not found", req.Id)
		}
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

	if err := s.validateOrgID(req.OrganizationId); err != nil {
		return nil, err
	}

	ou, err := s.store.CreateOrganizationUser(req.OrganizationId, req.UserId, req.Role.String())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.FailedPrecondition, "organization %q not found", req.OrganizationId)
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

	if err := s.validateOrgID(req.OrganizationId); err != nil {
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

	if err := s.validateOrgID(req.OrganizationId); err != nil {
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

func (s *S) validateOrgID(orgID string) error {
	if _, err := s.store.GetOrganizationByTenantIDAndOrgID(fakeTenantID, orgID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return status.Errorf(codes.FailedPrecondition, "organization %q not found", orgID)
		}
		return status.Errorf(codes.Internal, "get organization: %s", err)
	}

	return nil
}

// CreateDefaultOrganization creates the default org.
// TODO(kenji): This is not the best place for this function as there is nothing related to
// the server itself.
func (s *S) CreateDefaultOrganization(ctx context.Context, c *config.DefaultOrganizationConfig) error {
	log.Printf("Creating default org %q", c.Title)
	_, err := s.store.GetOrganizationByTenantIDAndTitle(fakeTenantID, c.Title)
	if err == nil {
		// Do nothing.
		return nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	org, err := s.CreateOrganization(ctx, &v1.CreateOrganizationRequest{
		Title:               c.Title,
		KubernetesNamespace: c.KubernetesNamespace,
	})
	if err != nil {
		return err
	}

	for _, uid := range c.UserIDs {
		if _, err := s.CreateOrganizationUser(ctx, &v1.CreateOrganizationUserRequest{
			OrganizationId: org.Id,
			UserId:         uid,
			Role:           v1.OrganizationRole_ORGANIZATION_ROLE_OWNER,
		}); err != nil {
			return err
		}
	}
	return nil
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

func generateRandomString(prefix string, n int) (string, error) {
	numBytes := (n * 3) / 4
	randBytes := make([]byte, numBytes)
	if _, err := rand.Read(randBytes); err != nil {
		return "", err
	}

	randStr := base64.URLEncoding.EncodeToString(randBytes)
	randStr = strings.TrimRight(randStr, "=")

	if len(randStr) > n {
		randStr = randStr[:n]
	}
	return prefix + randStr, nil
}
