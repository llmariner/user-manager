package server

import (
	"fmt"
	"net"

	"github.com/go-logr/logr"
	v1 "github.com/llmariner/user-manager/api/v1"
	"github.com/llmariner/user-manager/server/internal/store"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// NewInternal creates an internal server.
func NewInternal(
	store *store.S,
	dataKey []byte,
	log logr.Logger,
) *IS {
	return &IS{
		dataKey: dataKey,
		store:   store,
		log:     log.WithName("internal"),
	}
}

// IS is an internal server.
type IS struct {
	v1.UnimplementedUsersInternalServiceServer
	srv *grpc.Server

	dataKey []byte
	store   *store.S
	log     logr.Logger
}

// Run starts the gRPC server.
func (s *IS) Run(port int) error {
	s.log.Info("Starting internal server...", "port", port)

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

// GracefulStop gracefully stops the gRPC server.
func (s *IS) GracefulStop() {
	s.srv.GracefulStop()
}
