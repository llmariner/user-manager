package server

import (
	"context"
	"errors"

	"github.com/google/uuid"
	v1 "github.com/llm-operator/user-manager/api/v1"
	"github.com/llm-operator/user-manager/server/internal/store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

const (
	fakeTenantID = "fake-tenant-id"
)

// CreateAPIKey creates an API key.
func (s *S) CreateAPIKey(
	ctx context.Context,
	req *v1.CreateAPIKeyRequest,
) (*v1.APIKey, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	spec := store.APIKeySpec{
		Key: store.APIKeyKey{
			APIKeyID: newAPIKeyID(),
			TenantID: fakeTenantID,
		},
		Name:   req.Name,
		Secret: newSecret(),
	}
	k, err := s.store.CreateAPIKey(spec)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "create api key: %s", err)
	}
	return toAPIKeyProto(k, true), nil
}

// ListAPIKeys lists API keys.
func (s *S) ListAPIKeys(
	ctx context.Context,
	req *v1.ListAPIKeysRequest,
) (*v1.ListAPIKeysResponse, error) {
	ks, err := s.store.ListAPIKeysByTenantID(fakeTenantID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list api keys: %s", err)
	}

	var apiKeyProtos []*v1.APIKey
	for _, k := range ks {
		apiKeyProtos = append(apiKeyProtos, toAPIKeyProto(k, false))
	}
	return &v1.ListAPIKeysResponse{
		Object: "list",
		Data:   apiKeyProtos,
	}, nil
}

// DeleteAPIKey deletes an API key.
func (s *S) DeleteAPIKey(
	ctx context.Context,
	req *v1.DeleteAPIKeyRequest,
) (*v1.DeleteAPIKeyResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	if err := s.store.DeleteAPIKey(store.APIKeyKey{
		APIKeyID: req.Id,
		TenantID: fakeTenantID,
	}); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "api key %q not found", req.Id)
		}
		return nil, status.Errorf(codes.Internal, "delete api key: %s", err)
	}
	return &v1.DeleteAPIKeyResponse{
		Id:      req.Id,
		Object:  "users.api_key",
		Deleted: true,
	}, nil
}

// ListAPIKeys lists all API keys.
func (s *IS) ListAPIKeys(
	ctx context.Context,
	req *v1.ListAPIKeysRequest,
) (*v1.ListAPIKeysResponse, error) {
	ks, err := s.store.ListAllAPIKeys()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list api keys: %s", err)
	}

	var apiKeyProtos []*v1.APIKey
	for _, k := range ks {
		apiKeyProtos = append(apiKeyProtos, toAPIKeyProto(k, true))
	}
	return &v1.ListAPIKeysResponse{
		Object: "list",
		Data:   apiKeyProtos,
	}, nil
}

func toAPIKeyProto(k *store.APIKey, includeSecret bool) *v1.APIKey {
	kp := &v1.APIKey{
		Id:        k.APIKeyID,
		CreatedAt: k.CreatedAt.UTC().Unix(),
		Name:      k.Name,
		Object:    "user.api_key",
	}
	if includeSecret {
		kp.Secret = k.Secret
	}
	return kp
}

func newAPIKeyID() string {
	return uuid.New().String()
}

// TODO(kenji): Revisit.
func newSecret() string {
	return uuid.New().String()
}
