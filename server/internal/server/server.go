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
	"google.golang.org/grpc/reflection"
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
}

// Run starts the gRPC server.
func (s *S) Run(ctx context.Context, port int, authConfig config.AuthConfig) error {
	log.Printf("Starting server on port %d\n", port)

	var opts []grpc.ServerOption
	if authConfig.Enable {
		ai, err := auth.NewInterceptor(ctx, authConfig.RBACInternalServerAddr, "api.users.api_keys")
		if err != nil {
			return err
		}
		opts = append(opts, grpc.ChainUnaryInterceptor(ai.Unary()))
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
