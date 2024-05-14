package server

import (
	"context"

	v1 "github.com/llm-operator/user-manager/api/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateOrganization creates a new organization.
func (s *S) CreateOrganization(ctx context.Context, req *v1.CreateOrganizationRequest) (*v1.Organization, error) {
	if req.Title == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}

	orgID := ""
	org, err := s.store.CreateOrganization(fakeTenantID, orgID, req.Title)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "create organization: %s", err)
	}

	return &v1.Organization{
		Id:    org.OrganizationID,
		Title: org.Title,
	}, nil
}

// ListOrganization lists all organizations.
func (s *S) ListOrganization(ctx context.Context, req *v1.ListOrganizationRequest) (*v1.ListOrganizationResponse, error) {
	orgs, err := s.store.ListOrganization(fakeTenantID)
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
	return &v1.ListOrganizationResponse{
		Organizations: orgProtos,
	}, nil
}

// DeleteOrganization deletes an organization.
func (s *S) DeleteOrganization(ctx context.Context, req *v1.DeleteOrganizationRequest) (*v1.DeleteOrganizationResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "organization id is required")
	}

	if err := s.store.DeleteOrganization(req.Id); err != nil {
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

	if _, err := s.store.CreateUserOrganization(req.OrganizationId, req.UserId); err != nil {
		return nil, status.Errorf(codes.Internal, "add user to organization: %s", err)
	}

	return &v1.AddUserToOrganizationResponse{}, nil
}
