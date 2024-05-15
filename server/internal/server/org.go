package server

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strings"

	v1 "github.com/llm-operator/user-manager/api/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// CreateOrganization creates a new organization.
func (s *S) CreateOrganization(ctx context.Context, req *v1.CreateOrganizationRequest) (*v1.Organization, error) {
	if req.Title == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}

	orgID, err := generateOrgID()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "generate organization id: %s", err)
	}
	org, err := s.store.CreateOrganization(fakeTenantID, orgID, req.Title)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "create organization: %s", err)
	}

	return &v1.Organization{
		Id:        org.OrganizationID,
		Title:     org.Title,
		CreatedAt: org.CreatedAt.UTC().Unix(),
	}, nil
}

// ListOrganizations lists all organizations.
func (s *S) ListOrganizations(ctx context.Context, req *v1.ListOrganizationsRequest) (*v1.ListOrganizationsResponse, error) {
	orgs, err := s.store.ListOrganizations(fakeTenantID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list organizations: %s", err)
	}

	var orgProtos []*v1.Organization
	for _, org := range orgs {
		orgProtos = append(orgProtos, &v1.Organization{
			Id:    org.OrganizationID,
			Title: org.Title,
		})
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

	return &v1.DeleteOrganizationResponse{}, nil
}

// AddUserToOrganization adds a user to an organization.
func (s *S) AddUserToOrganization(ctx context.Context, req *v1.AddUserToOrganizationRequest) (*v1.AddUserToOrganizationResponse, error) {
	if req.OrganizationId == "" {
		return nil, status.Error(codes.InvalidArgument, "organization id is required")
	}
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}

	if _, err := s.store.CreateUserOrganization(fakeTenantID, req.OrganizationId, req.UserId); err != nil {
		return nil, status.Errorf(codes.Internal, "add user to organization: %s", err)
	}

	return &v1.AddUserToOrganizationResponse{}, nil
}

func generateRandomString(n int) (string, error) {
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
	return randStr, nil
}

func generateOrgID() (string, error) {
	const (
		prefix = "org-"
		length = 22
	)
	randStr, err := generateRandomString(length)
	if err != nil {
		return "", err
	}
	return prefix + randStr, nil
}
