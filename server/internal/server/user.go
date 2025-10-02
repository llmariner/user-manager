package server

import (
	"context"
	"fmt"
	"sort"

	gerrors "github.com/llmariner/common/pkg/gormlib/errors"
	"github.com/llmariner/rbac-manager/pkg/auth"
	v1 "github.com/llmariner/user-manager/api/v1"
	"github.com/llmariner/user-manager/pkg/userid"
	"github.com/llmariner/user-manager/server/internal/store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

// GetUserSelf gets a self-user.
func (s *S) GetUserSelf(ctx context.Context, req *v1.GetUserSelfRequest) (*v1.User, error) {
	userInfo, ok := auth.ExtractUserInfoFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("failed to extract user info from context")
	}

	orgUsers, err := s.store.ListOrganizationUsersByUserID(userInfo.UserID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list organization users: %s", err)
	}
	projUsers, err := s.store.ListProjectUsersByUserID(userInfo.UserID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list project users: %s", err)
	}

	return &v1.User{
		Id:                       userInfo.UserID,
		OrganizationRoleBindings: toOrganizationRoleBindings(orgUsers),
		ProjectRoleBindings:      toProjectRoleBindings(projUsers),
	}, nil
}

// ListUsers lists all users.
//   - An organization owner can view users in their organizations.
//   - An organization reader can view users in their organizations if they belong to the projects
//     where the organization member also belongs to.
func (s *S) ListUsers(ctx context.Context, req *v1.ListUsersRequest) (*v1.ListUsersResponse, error) {
	userInfo, ok := auth.ExtractUserInfoFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("failed to extract user info from context")
	}

	orgUsers, err := s.store.ListAllOrganizationUsers()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list organization users: %s", err)
	}
	orgUsersByUserID := make(map[string][]store.OrganizationUser)
	for _, ou := range orgUsers {
		orgUsersByUserID[ou.UserID] = append(orgUsersByUserID[ou.UserID], ou)
	}

	projUsers, err := s.store.ListAllProjectUsers()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list project users: %s", err)
	}
	projUsersByUserID := make(map[string][]store.ProjectUser)
	for _, pu := range projUsers {
		projUsersByUserID[pu.UserID] = append(projUsersByUserID[pu.UserID], pu)
	}

	// Identify visible orgs and projects.

	userByUserID := make(map[string]*v1.User)
	for _, ou := range orgUsersByUserID[userInfo.UserID] {
		r, ok := v1.OrganizationRole_value[ou.Role]
		if !ok {
			return nil, status.Errorf(codes.Internal, "unknown organization role: %q", ou.Role)
		}
		if v1.OrganizationRole(r) != v1.OrganizationRole_ORGANIZATION_ROLE_OWNER {
			continue
		}

		users, err := s.store.ListOrganizationNonHiddenUsersByOrganizationID(ou.OrganizationID)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "list organization users: %s", err)
		}
		for _, u := range users {
			if _, ok := userByUserID[u.UserID]; ok {
				continue
			}

			userByUserID[u.UserID] = &v1.User{
				Id: u.UserID,
				// TODO(kenji): Consider restricting the visibility of organization role bindings
				// (i.e., do not show org role bindings if the requester is not an org owner).
				OrganizationRoleBindings: toOrganizationRoleBindings(orgUsersByUserID[u.UserID]),
				ProjectRoleBindings:      toProjectRoleBindings(projUsersByUserID[u.UserID]),
			}
		}
	}

	for _, pu := range projUsersByUserID[userInfo.UserID] {
		if _, ok := userByUserID[pu.UserID]; ok {
			// Already visible with org owner role.
			continue
		}

		users, err := s.store.ListProjectUsersByProjectID(pu.ProjectID)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "list project users: %s", err)
		}
		for _, u := range users {
			if _, ok := userByUserID[u.UserID]; ok {
				continue
			}

			userByUserID[u.UserID] = &v1.User{
				Id: u.UserID,
				// Do not expose the organization role bindings.
				ProjectRoleBindings: toProjectRoleBindings(projUsersByUserID[u.UserID]),
			}
		}
	}

	var users []*v1.User
	for _, u := range userByUserID {
		users = append(users, u)
	}

	sort.Slice(users, func(i, j int) bool {
		return users[i].Id < users[j].Id
	})

	return &v1.ListUsersResponse{
		Users: users,
	}, nil
}

// ListUsers lists all users.
func (s *IS) ListUsers(ctx context.Context, req *v1.ListUsersRequest) (*v1.ListUsersResponse, error) {
	users, err := s.store.ListAllUsers()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list users: %s", err)
	}

	var res v1.ListUsersResponse
	for _, u := range users {
		res.Users = append(res.Users, &v1.User{
			Id:               u.UserID,
			InternalId:       u.InternalUserID,
			IsServiceAccount: false,
			Hidden:           u.Hidden,
			// Do not populate organization_role_bindings and project_role_bindings as they are not used.
		})
	}
	return &res, nil
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

				var found *store.Project
				for _, p := range ps {
					if p.OrganizationID != org.OrganizationID {
						continue
					}

					kn := p.KubernetesNamespace
					if kn == "" {
						var asp v1.ProjectAssignments
						if err := proto.Unmarshal(p.Assignments, &asp); err != nil {
							return status.Errorf(codes.Internal, "unmarshal project assignments: %s", err)
						}
						for _, a := range asp.Assignments {
							if a.ClusterId == "" {
								kn = a.Namespace
								break
							}
						}
						if kn == "" {
							return status.Errorf(codes.Internal, "kubernetes namespace not found")
						}
					}

					if kn != req.KubernetesNamespace {
						continue
					}

					found = p
					break
				}

				if found == nil {
					return status.Errorf(codes.NotFound, "project not found: %s", req.KubernetesNamespace)
				}

				if _, err := store.CreateProjectUserInTransaction(tx, store.CreateProjectUserParams{
					ProjectID:      found.ProjectID,
					OrganizationID: found.OrganizationID,
					UserID:         userID,
					Role:           v1.ProjectRole_PROJECT_ROLE_OWNER,
				}); err != nil {
					return err
				}
				return nil
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

		as := []*v1.ProjectAssignment{
			{
				Namespace: req.KubernetesNamespace,
			},
		}
		if _, err = createProject(s.store, req.Title, org.OrganizationID, as, false, org.TenantID); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	s.log.Info("Created orgnization, project, and user", "user_id", userID)
	return &emptypb.Empty{}, nil
}

func toOrganizationRoleBindings(ous []store.OrganizationUser) []*v1.User_OrganizationRoleBinding {
	var bindings []*v1.User_OrganizationRoleBinding
	for _, ou := range ous {
		bindings = append(bindings, &v1.User_OrganizationRoleBinding{
			OrganizationId: ou.OrganizationID,
			Role:           v1.OrganizationRole(v1.OrganizationRole_value[ou.Role]),
		})
	}
	return bindings
}

func toProjectRoleBindings(pus []store.ProjectUser) []*v1.User_ProjectRoleBinding {
	var bindings []*v1.User_ProjectRoleBinding
	for _, pu := range pus {
		bindings = append(bindings, &v1.User_ProjectRoleBinding{
			ProjectId: pu.ProjectID,
			Role:      v1.ProjectRole(v1.ProjectRole_value[pu.Role]),
		})
	}
	return bindings
}
