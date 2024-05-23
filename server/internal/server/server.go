package server

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/llm-operator/rbac-manager/pkg/auth"
	v1 "github.com/llm-operator/user-manager/api/v1"
	"github.com/llm-operator/user-manager/server/internal/config"
	"github.com/llm-operator/user-manager/server/internal/store"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

const (
	defaultProjectID = "default"
)

// New creates a server.
func New(store *store.S) *S {
	return &S{
		store: store,
	}
}

// S is a server.
type S struct {
	v1.UnimplementedUsersServiceServer

	srv *grpc.Server

	store *store.S

	enableAuth bool
}

// Run starts the gRPC server.
func (s *S) Run(ctx context.Context, port int, authConfig config.AuthConfig) error {
	log.Printf("Starting server on port %d\n", port)

	var opts []grpc.ServerOption
	if authConfig.Enable {
		// TODO(kenji): Change the scope depending on RPC methods.
		ai, err := auth.NewInterceptor(ctx, auth.Config{
			RBACServerAddr: authConfig.RBACInternalServerAddr,
			AccessResource: "api.users.api_keys",
		})
		if err != nil {
			return err
		}
		opts = append(opts, grpc.ChainUnaryInterceptor(ai.Unary()))
		s.enableAuth = true
	}

	grpcServer := grpc.NewServer(opts...)
	v1.RegisterUsersServiceServer(grpcServer, s)
	reflection.Register(grpcServer)

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

func (s *S) extractUserInfoFromContext(ctx context.Context) (*auth.UserInfo, error) {
	if !s.enableAuth {
		return &auth.UserInfo{
			OrganizationID:      "default",
			ProjectID:           defaultProjectID,
			KubernetesNamespace: "default",
		}, nil
	}
	var ok bool
	userInfo, ok := auth.ExtractUserInfoFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user info not found")
	}
	return userInfo, nil
}

// orgRole returns a role that the given user has for the given organization.
func (s *S) orgRole(orgID, userID string) v1.OrganizationRole {
	if !s.enableAuth {
		return v1.OrganizationRole_ORGANIZATION_ROLE_OWNER
	}

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

// projectRole returns a role that the given user has for the given project.
func (s *S) projectRole(projectID, userID string) v1.ProjectRole {
	if !s.enableAuth {
		return v1.ProjectRole_PROJECT_ROLE_OWNER
	}

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
