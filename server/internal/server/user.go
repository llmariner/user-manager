package server

import (
	"context"
	"fmt"

	gerrors "github.com/llmariner/common/pkg/gormlib/errors"
	"github.com/llmariner/rbac-manager/pkg/auth"
	v1 "github.com/llmariner/user-manager/api/v1"
	"github.com/llmariner/user-manager/pkg/userid"
	"github.com/llmariner/user-manager/server/internal/store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

// GetUserSelf gets a self-user.
func (s *S) GetUserSelf(ctx context.Context, req *v1.GetUserSelfRequest) (*v1.User, error) {
	userInfo, ok := auth.ExtractUserInfoFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("failed to extract user info from context")
	}

	return &v1.User{
		Id: userInfo.UserID,
	}, nil
}

// CreateUserInternal creates a new user and related organization and project.
func (s *IS) CreateUserInternal(ctx context.Context, req *v1.CreateUserInternalRequest) (*emptypb.Empty, error) {
	if req.TenantId == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant id is required")
	}
	if req.Title == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}
	userID := userid.Normalize(req.UserId)
	if req.KubernetesNamespace == "" {
		return nil, status.Error(codes.InvalidArgument, "kubernets namespace is required")
	}

	users, err := s.store.ListAllUsers()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list users: %s", err)
	}
	for _, u := range users {
		if u.UserID == userID {
			// no-op if user already exists.
			return &emptypb.Empty{}, nil
		}
	}

	orgs, err := s.store.ListAllOrganizations()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list organizations: %s", err)
	}

	for _, org := range orgs {
		// add user to existing organization and project.
		if org.TenantID == req.TenantId && org.Title == req.Title {
			if err := s.store.Transaction(func(tx *gorm.DB) error {
				if _, err := findOrCreateUserInTransaction(tx, userID); err != nil {
					return err
				}
				if _, err := store.CreateOrganizationUserInTransaction(
					tx,
					org.OrganizationID,
					userID,
					v1.OrganizationRole_ORGANIZATION_ROLE_OWNER.String(),
				); err != nil {
					return err
				}

				ps, err := s.store.ListProjectsByTenantIDAndOrganizationID(org.TenantID, org.OrganizationID)
				if err != nil {
					return status.Errorf(codes.Internal, "list projects: %s", err)
				}
				for _, p := range ps {
					if p.OrganizationID == org.OrganizationID && p.KubernetesNamespace == req.KubernetesNamespace {
						if _, err := store.CreateProjectUserInTransaction(tx, store.CreateProjectUserParams{
							ProjectID:      p.ProjectID,
							OrganizationID: p.OrganizationID,
							UserID:         userID,
							Role:           v1.ProjectRole_PROJECT_ROLE_OWNER,
						}); err != nil {
							return err
						}
						return nil
					}
				}
				return status.Errorf(codes.NotFound, "project not found: %s", req.KubernetesNamespace)
			}); err != nil {
				return nil, err
			}
			s.log.Info("Created user", "user_id", userID)
			return &emptypb.Empty{}, nil
		}
	}

	// create new organization and project, and add user to them.
	if err := s.store.Transaction(func(tx *gorm.DB) error {
		org, err := createOrganization(s.store, req.Title, false, req.TenantId, []string{userID})
		if err != nil {
			if gerrors.IsUniqueConstraintViolation(err) {
				return status.Errorf(codes.AlreadyExists, "organization %q already exists", req.Title)
			}
			return status.Error(codes.Internal, err.Error())
		}

		if _, err = createProject(s.store, req.Title, org.OrganizationID, req.KubernetesNamespace, false, org.TenantID); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	s.log.Info("Created orgnization, project, and user", "user_id", userID)
	return &emptypb.Empty{}, nil
}
