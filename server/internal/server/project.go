package server

import (
	"context"
	"errors"
	"fmt"

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
	"k8s.io/apimachinery/pkg/api/validation"
)

// CreateProject creates a new project.
func (s *S) CreateProject(ctx context.Context, req *v1.CreateProjectRequest) (*v1.Project, error) {
	userInfo, ok := auth.ExtractUserInfoFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("failed to extract user info from context")
	}

	if req.Title == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}
	if req.OrganizationId == "" {
		return nil, status.Error(codes.InvalidArgument, "organization id is required")
	}
	if req.KubernetesNamespace == "" {
		return nil, status.Error(codes.InvalidArgument, "kubernetes namespace is required")
	}

	if err := s.validateOrganizationOwner(req.OrganizationId, userInfo.UserID); err != nil {
		return nil, err
	}

	return s.createProject(ctx,
		req.Title,
		req.OrganizationId,
		req.KubernetesNamespace,
		false,
		userInfo.TenantID,
	)
}

func (s *S) createProject(
	ctx context.Context,
	title string,
	organizationID string,
	kubernetesNamespace string,
	isDefault bool,
	tenantID string,
) (*v1.Project, error) {
	if _, err := s.validateOrganizationID(organizationID, tenantID); err != nil {
		return nil, err
	}

	if errs := validation.ValidateNamespaceName(kubernetesNamespace, false); len(errs) != 0 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid kubernetes namespace: %s", errs)
	}

	projectID, err := id.GenerateID("proj_", 24)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "generate project id: %s", err)
	}

	orgUsers, err := s.store.ListOrganizationUsersByOrganizationID(organizationID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list organization users: %s", err)
	}

	var p *store.Project
	if err := s.store.Transaction(func(tx *gorm.DB) error {
		p, err = store.CreateProjectInTransaction(tx, store.CreateProjectParams{
			TenantID:            tenantID,
			ProjectID:           projectID,
			OrganizationID:      organizationID,
			Title:               title,
			KubernetesNamespace: kubernetesNamespace,
			IsDefault:           isDefault,
		})
		if err != nil {
			return err
		}

		// Add org owners to project owners.
		for _, ou := range orgUsers {
			role := v1.OrganizationRole(v1.OrganizationRole_value[ou.Role])
			if role != v1.OrganizationRole_ORGANIZATION_ROLE_OWNER {
				continue
			}
			_, err := store.CreateProjectUserInTransaction(tx, store.CreateProjectUserParams{
				ProjectID:      p.ProjectID,
				OrganizationID: p.OrganizationID,
				UserID:         ou.UserID,
				Role:           v1.ProjectRole_PROJECT_ROLE_OWNER,
			})
			if err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		if gerrors.IsUniqueConstraintViolation(err) {
			return nil, status.Errorf(codes.AlreadyExists, "project %q already exists", title)
		}
		return nil, status.Errorf(codes.Internal, "create project: %s", err)
	}

	return p.ToProto(), nil
}

// ListProjects lists all projects.
func (s *S) ListProjects(ctx context.Context, req *v1.ListProjectsRequest) (*v1.ListProjectsResponse, error) {
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

	ps, err := s.store.ListProjectsByTenantIDAndOrganizationID(userInfo.TenantID, req.OrganizationId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list projects: %s", err)
	}

	var filtered []*store.Project
	for _, p := range ps {
		if s.validateProjectMember(p.ProjectID, p.OrganizationID, userInfo.UserID) == nil {
			filtered = append(filtered, p)
		}
	}

	var pProtos []*v1.Project
	for _, p := range filtered {
		pProtos = append(pProtos, p.ToProto())
	}
	return &v1.ListProjectsResponse{
		Projects: pProtos,
	}, nil
}

// DeleteProject deletes an project.
func (s *S) DeleteProject(ctx context.Context, req *v1.DeleteProjectRequest) (*v1.DeleteProjectResponse, error) {
	userInfo, ok := auth.ExtractUserInfoFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("failed to extract user info from context")
	}

	if req.OrganizationId == "" {
		return nil, status.Error(codes.InvalidArgument, "organization id is required")
	}
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "project id is required")
	}

	p, err := s.validateProjectID(req.Id, req.OrganizationId, userInfo.TenantID)
	if err != nil {
		return nil, err
	}

	if err := s.validateProjectOwner(req.Id, req.OrganizationId, userInfo.UserID); err != nil {
		return nil, err
	}

	if p.IsDefault {
		return nil, status.Errorf(codes.InvalidArgument, "cannot delete a default project")
	}

	if err := s.store.Transaction(func(tx *gorm.DB) error {
		if err := store.DeleteProjectInTransaction(tx, req.Id); err != nil {
			return fmt.Errorf("delete project: %s", err)
		}
		if err := store.DeleteAllProjectUsersInTransaction(tx, req.Id); err != nil {
			return fmt.Errorf("delete all project users: %s", err)
		}
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "transaction: %s", err)
	}

	return &v1.DeleteProjectResponse{
		Id:      req.Id,
		Object:  "project",
		Deleted: true,
	}, nil
}

// CreateProjectUser adds a user to an project.
func (s *S) CreateProjectUser(ctx context.Context, req *v1.CreateProjectUserRequest) (*v1.ProjectUser, error) {
	userInfo, ok := auth.ExtractUserInfoFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("failed to extract user info from context")
	}

	if req.ProjectId == "" {
		return nil, status.Error(codes.InvalidArgument, "project id is required")
	}
	if req.OrganizationId == "" {
		return nil, status.Error(codes.InvalidArgument, "organization id is required")
	}
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}
	if req.Role == v1.ProjectRole_PROJECT_ROLE_UNSPECIFIED {
		return nil, status.Error(codes.InvalidArgument, "role is required")
	}

	if _, err := s.validateProjectID(req.ProjectId, req.OrganizationId, userInfo.TenantID); err != nil {
		return nil, err
	}

	if err := s.validateProjectOwner(req.ProjectId, req.OrganizationId, userInfo.UserID); err != nil {
		return nil, err
	}

	userID := userid.Normalize(req.UserId)
	if s.organizationRole(req.OrganizationId, userID) == v1.OrganizationRole_ORGANIZATION_ROLE_UNSPECIFIED {
		return nil, status.Errorf(codes.FailedPrecondition, "user %q is not a member of the organization", userID)
	}

	pu, err := s.store.CreateProjectUser(store.CreateProjectUserParams{
		ProjectID:      req.ProjectId,
		OrganizationID: req.OrganizationId,
		UserID:         userID,
		Role:           req.Role,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.FailedPrecondition, "project %q not found", req.ProjectId)
		}
		if gerrors.IsUniqueConstraintViolation(err) {
			return nil, status.Errorf(codes.AlreadyExists, "user %q is already a member of project %q", userID, req.ProjectId)
		}
		return nil, status.Errorf(codes.Internal, "add user to project: %s", err)
	}

	return pu.ToProto(), nil
}

// ListProjectUsers lists project users for the specified project.
func (s *S) ListProjectUsers(ctx context.Context, req *v1.ListProjectUsersRequest) (*v1.ListProjectUsersResponse, error) {
	userInfo, ok := auth.ExtractUserInfoFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("failed to extract user info from context")
	}

	if req.ProjectId == "" {
		return nil, status.Error(codes.InvalidArgument, "project id is required")
	}
	if req.OrganizationId == "" {
		return nil, status.Error(codes.InvalidArgument, "organization id is required")
	}

	if _, err := s.validateProjectID(req.ProjectId, req.OrganizationId, userInfo.TenantID); err != nil {
		return nil, err
	}

	if err := s.validateProjectMember(req.ProjectId, req.OrganizationId, userInfo.UserID); err != nil {
		return nil, err
	}

	users, err := s.store.ListProjectUsersByProjectID(req.ProjectId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list project users: %s", err)
	}

	var userProtos []*v1.ProjectUser
	for _, user := range users {
		userProtos = append(userProtos, user.ToProto())
	}
	return &v1.ListProjectUsersResponse{
		Users: userProtos,
	}, nil
}

// DeleteProjectUser deletes an project user.
func (s *S) DeleteProjectUser(ctx context.Context, req *v1.DeleteProjectUserRequest) (*emptypb.Empty, error) {
	userInfo, ok := auth.ExtractUserInfoFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("failed to extract user info from context")
	}

	if req.ProjectId == "" {
		return nil, status.Error(codes.InvalidArgument, "project id is required")
	}
	if req.OrganizationId == "" {
		return nil, status.Error(codes.InvalidArgument, "organization id is required")
	}
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}

	if _, err := s.validateProjectID(req.ProjectId, req.OrganizationId, userInfo.TenantID); err != nil {
		return nil, err
	}

	if err := s.validateProjectOwner(req.ProjectId, req.OrganizationId, userInfo.UserID); err != nil {
		return nil, err
	}

	userID := userid.Normalize(req.UserId)
	if err := s.store.DeleteProjectUser(req.ProjectId, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "project user not found")
		}
		return nil, status.Errorf(codes.Internal, "delete project user: %s", err)
	}

	return &emptypb.Empty{}, nil
}

func (s *S) validateProjectID(projectID, orgID, tenantID string) (*store.Project, error) {
	if _, err := s.validateOrganizationID(orgID, tenantID); err != nil {
		return nil, err
	}

	p, err := s.store.GetProject(store.GetProjectParams{
		TenantID:       tenantID,
		OrganizationID: orgID,
		ProjectID:      projectID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.FailedPrecondition, "project %q not found", orgID)
		}
		return nil, status.Errorf(codes.Internal, "get project: %s", err)
	}

	return p, nil
}

// CreateDefaultProject creates the default org.
// TODO(kenji): This is not the best place for this function as there is nothing related to
// the server itself.
func (s *S) CreateDefaultProject(ctx context.Context, c *config.DefaultProjectConfig, orgID, tenantID string) (*store.Project, error) {
	p, err := s.store.GetDefaultProject(tenantID)
	if err == nil {
		// Do nothing.
		return p, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	s.log.Info("Creating default project", "title", c.Title)
	if _, err := s.createProject(ctx,
		c.Title,
		orgID,
		c.KubernetesNamespace,
		true,
		tenantID,
	); err != nil {
		return nil, err
	}

	return p, nil
}

// ListProjects lists all projects.
func (s *IS) ListProjects(ctx context.Context, req *v1.ListProjectsRequest) (*v1.ListProjectsResponse, error) {
	orgs, err := s.store.ListAllProjects()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list projects: %s", err)
	}

	var orgProtos []*v1.Project
	for _, org := range orgs {
		orgProtos = append(orgProtos, org.ToProto())
	}
	return &v1.ListProjectsResponse{
		Projects: orgProtos,
	}, nil
}

// ListProjectUsers lists project users for all projects.
func (s *IS) ListProjectUsers(ctx context.Context, req *v1.ListProjectUsersRequest) (*v1.ListProjectUsersResponse, error) {
	users, err := s.store.ListAllProjectUsers()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list project users: %s", err)
	}

	var userProtos []*v1.ProjectUser
	for _, user := range users {
		userProtos = append(userProtos, user.ToProto())
	}
	return &v1.ListProjectUsersResponse{
		Users: userProtos,
	}, nil
}
