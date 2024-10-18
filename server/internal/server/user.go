package server

import (
	"context"

	gerrors "github.com/llmariner/common/pkg/gormlib/errors"
	"github.com/llmariner/common/pkg/id"
	v1 "github.com/llmariner/user-manager/api/v1"
	"github.com/llmariner/user-manager/pkg/userid"
	"github.com/llmariner/user-manager/server/internal/store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// CreateInternalUser creates a new user and related organization and project, and generates an API key.
func (s *IS) CreateInternalUser(ctx context.Context, req *v1.CreateInternalUserRequest) (*emptypb.Empty, error) {
	if req.TenantId == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant id is required")
	}
	if req.Title == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}

	users, err := s.store.ListAllUsers()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list users: %s", err)
	}
	for _, u := range users {
		if u.UserID == req.UserId {
			return nil, status.Errorf(codes.AlreadyExists, "user %q already exists", req.UserId)
		}
	}

	orgs, err := s.store.ListAllOrganizations()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list organizations: %s", err)
	}

	var foundTenant bool
	for _, org := range orgs {
		if org.TenantID == req.TenantId {
			foundTenant = true
		}
	}

	var proj *v1.Project
	var org *store.Organization
	if !foundTenant {
		org, err = createOrganization(ctx, s.store, req.Title, false, req.TenantId, []string{userid.Normalize(req.UserId)})
		if err != nil {
			if gerrors.IsUniqueConstraintViolation(err) {
				return nil, status.Errorf(codes.AlreadyExists, "organizatione %q already exists", req.Title)
			}
			return nil, status.Error(codes.Internal, err.Error())
		}

		proj, err = createProject(ctx, s.store, req.Title, org.OrganizationID, "default", false, org.TenantID)
		if err != nil {
			return nil, err
		}
	}

	// TODO(kenji): Make sure this gives sufficient randomness.
	secKey, err := id.GenerateID("sk-", 48)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "generate api key id: %s", err)
	}

	_, err = createAPIKey(ctx, s.store, s.dataKey, "default", secKey, req.UserId, org.OrganizationID, proj.Id, org.TenantID)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
