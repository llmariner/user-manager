package server

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/go-logr/logr"
	"github.com/llmariner/api-usage/pkg/sender"
	"github.com/llmariner/rbac-manager/pkg/auth"
	v1 "github.com/llmariner/user-manager/api/v1"
	"github.com/llmariner/user-manager/server/internal/config"
	"github.com/llmariner/user-manager/server/internal/store"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

const (
	defaultUserID    = "defaultuser"
	defaultProjectID = "defaultProject"
	defaultTenantID  = "default-tenant-id"
)

// New creates a server.
func New(
	store *store.S,
	dataKey []byte,
	log logr.Logger,
) *S {
	return &S{
		store:   store,
		dataKey: dataKey,
		log:     log.WithName("grpc"),
	}
}

// S is a server.
type S struct {
	v1.UnimplementedUsersServiceServer

	srv *grpc.Server

	dataKey []byte
	store   *store.S
	log     logr.Logger

	enableAuth bool
}

// Run starts the gRPC server.
func (s *S) Run(ctx context.Context, port int, authConfig config.AuthConfig, usage sender.UsageSetter) error {
	s.log.Info("Starting gRPC server...", "port", port)

	var opt grpc.ServerOption
	if authConfig.Enable {
		ai, err := auth.NewInterceptor(ctx, auth.Config{
			RBACServerAddr: authConfig.RBACInternalServerAddr,
			GetAccessResourceForGRPCRequest: func(fullMethod string) string {
				// Note that the authorization check peformed by the RBAC server is not sufficient
				// since organizations and projects have more complex authorization rules.
				// The additional checks are performed in the individual handlers.
				ms := strings.Split(fullMethod, "/")
				method := ms[len(ms)-1]
				switch method {
				case "CreateAPIKey", "DeleteAPIKey", "ListAPIKeys", "CreateProjectAPIKey", "DeleteProjectAPIKey", "ListProjectAPIKeys":
					return "api.organizations.projects.api_keys"
				case "CreateOrganization", "DeleteOrganization", "ListOrganizations":
					return "api.organizations"
				case "CreateOrganizationUser", "DeleteOrganizationUser", "ListOrganizationUsers":
					return "api.organizations.users"
				case "CreateProject", "DeleteProject", "ListProjects":
					return "api.organizations.projects"
				case "CreateProjectUser", "DeleteProjectUser", "ListProjectUsers":
					return "api.organizations.projects.users"
				case "GetUserSelf":
					return "api.selfuser"
				default:
					return "unknown"
				}
			},
		})
		if err != nil {
			return err
		}
		opt = grpc.ChainUnaryInterceptor(ai.Unary("/grpc.health.v1.Health/Check"), sender.Unary(usage))
		s.enableAuth = true
	} else {
		fakeAuth := func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
			return handler(fakeAuthInto(ctx), req)
		}
		opt = grpc.ChainUnaryInterceptor(fakeAuth, sender.Unary(usage))
	}

	grpcServer := grpc.NewServer(opt)
	v1.RegisterUsersServiceServer(grpcServer, s)
	reflection.Register(grpcServer)

	healthCheck := health.NewServer()
	healthCheck.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(grpcServer, healthCheck)

	s.srv = grpcServer

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("listen: %s", err)
	}
	if err := grpcServer.Serve(l); err != nil {
		return fmt.Errorf("serve: %s", err)
	}
	return nil
}

// Stop stops the gRPC server.
func (s *S) Stop() {
	s.srv.Stop()
}

// fakeAuthInto sets dummy user info and token into the context.
func fakeAuthInto(ctx context.Context) context.Context {
	return auth.AppendUserInfoToContext(ctx, auth.UserInfo{
		UserID:         defaultUserID,
		OrganizationID: "default",
		ProjectID:      defaultProjectID,
		AssignedKubernetesEnvs: []auth.AssignedKubernetesEnv{
			{
				ClusterID: "default",
				Namespace: "default",
			},
		},
		TenantID: defaultTenantID,
	})
}

// organizationRole returns a role that the given user has for the given organization.
func (s *S) organizationRole(orgID, userID string) v1.OrganizationRole {
	ou, err := s.store.GetOrganizationUser(orgID, userID)
	if err != nil {
		return v1.OrganizationRole_ORGANIZATION_ROLE_UNSPECIFIED
	}
	r, ok := v1.OrganizationRole_value[ou.Role]
	if !ok {
		return v1.OrganizationRole_ORGANIZATION_ROLE_UNSPECIFIED
	}
	return v1.OrganizationRole(r)
}

func (s *S) isOrganizationMember(orgID, userID string) bool {
	if !s.enableAuth {
		return true
	}

	return s.organizationRole(orgID, userID) != v1.OrganizationRole_ORGANIZATION_ROLE_UNSPECIFIED
}

func (s *S) isOrganizationOwner(orgID, userID string) bool {
	if !s.enableAuth {
		return true
	}

	return s.organizationRole(orgID, userID) == v1.OrganizationRole_ORGANIZATION_ROLE_OWNER
}

// validateOrganizationOwner checks if the user has the permission to manage the organization.
//
// If the authorization is enabled, this passes only when the user is an owner of the organization.
//
// If the authorization is disabled, this passes.
// Note that in that case, the existence of the organization is not checked.
func (s *S) validateOrganizationOwner(orgID, userID string) error {
	if !s.enableAuth {
		return nil
	}

	r := s.organizationRole(orgID, userID)
	switch r {
	case v1.OrganizationRole_ORGANIZATION_ROLE_UNSPECIFIED:
		// The org shouldn't be visible to the user.
		return status.Errorf(codes.FailedPrecondition, "organization %q not found", orgID)
	case v1.OrganizationRole_ORGANIZATION_ROLE_OWNER:
		return nil
	case v1.OrganizationRole_ORGANIZATION_ROLE_READER:
		return status.Errorf(codes.PermissionDenied, "user %q is not the owner of organization %q", userID, orgID)
	default:
		return status.Errorf(codes.Internal, "unknown role %q", r.String())
	}
}

// projectRole returns a role that the given user has for the given project.
func (s *S) projectRole(projectID, userID string) v1.ProjectRole {
	pu, err := s.store.GetProjectUser(projectID, userID)
	if err != nil {
		return v1.ProjectRole_PROJECT_ROLE_UNSPECIFIED
	}
	r, ok := v1.ProjectRole_value[pu.Role]
	if !ok {
		return v1.ProjectRole_PROJECT_ROLE_UNSPECIFIED
	}
	return v1.ProjectRole(r)
}

// validateProjectOwner checks if the user has the permission to manage the project.
// If the user can manage the specified organization, this passes.
// If the user can manage the specified project, this passes.
//
// Note that this function doesn't check if the specified project belongs to the specified organization.
// If the organization ID is coming from an external source, it needs to be validated first.
func (s *S) validateProjectOwner(projectID, orgID, userID string) error {
	if !s.enableAuth {
		return nil
	}

	if s.isOrganizationOwner(orgID, userID) {
		return nil
	}

	r := s.projectRole(projectID, userID)
	switch r {
	case v1.ProjectRole_PROJECT_ROLE_UNSPECIFIED:
		// The project shouldn't be visible to the user.
		return status.Errorf(codes.FailedPrecondition, "project %q not found", projectID)
	case v1.ProjectRole_PROJECT_ROLE_OWNER:
		return nil
	case v1.ProjectRole_PROJECT_ROLE_MEMBER:
		return status.Errorf(codes.PermissionDenied, "user %q is not the owner of project %q", userID, projectID)
	default:
		return status.Errorf(codes.Internal, "unknown role %q", r.String())
	}
}

func (s *S) validateProjectMember(projectID, orgID, userID string) error {
	if !s.enableAuth {
		return nil
	}

	if s.isOrganizationOwner(orgID, userID) {
		return nil
	}

	r := s.projectRole(projectID, userID)
	switch r {
	case v1.ProjectRole_PROJECT_ROLE_UNSPECIFIED:
		// The project shouldn't be visible to the user.
		return status.Errorf(codes.FailedPrecondition, "project %q not found", projectID)
	case v1.ProjectRole_PROJECT_ROLE_OWNER, v1.ProjectRole_PROJECT_ROLE_MEMBER:
		return nil
	default:
		return status.Errorf(codes.Internal, "unknown role %q", r.String())
	}
}
