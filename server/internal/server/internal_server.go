package server

import (
	"fmt"
	"log"
	"net"

	v1 "github.com/llm-operator/user-manager/api/v1"
	"github.com/llm-operator/user-manager/server/internal/store"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// NewInternal creates an internal server.
func NewInternal(store *store.S) *IS {
	return &IS{
		store: store,
	}
}

// IS is an internal server.
type IS struct {
	v1.UnimplementedUsersInternalServiceServer

	srv *grpc.Server

	store *store.S
}

// Run starts the gRPC server.
func (s *IS) Run(port int) error {
	log.Printf("Starting server on port %d\n", port)

	grpcServer := grpc.NewServer()
	v1.RegisterUsersInternalServiceServer(grpcServer, s)
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
