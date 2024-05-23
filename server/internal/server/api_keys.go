package server

import (
	"context"
	"errors"

	gerrors "github.com/llm-operator/common/pkg/gormlib/errors"
	"github.com/llm-operator/common/pkg/id"
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
	userInfo, err := s.extractUserInfoFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	trackID, err := id.GenerateID("key_", 16)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "generate api key id: %s", err)
	}
	// TODO(kenji): Make sure this gives sufficient randomness.
	secKey, err := id.GenerateID("sk-", 48)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "generate secret key: %s", err)
	}

	spec := store.APIKeySpec{
		APIKeyID:       trackID,
		TenantID:       fakeTenantID,
		OrganizationID: userInfo.OrganizationID,
		ProjectID:      userInfo.ProjectID,
		UserID:         userInfo.UserID,
		Name:           req.Name,
		Secret:         secKey,
	}
	k, err := s.store.CreateAPIKey(spec)
	if err != nil {
		if gerrors.IsUniqueConstraintViolation(err) {
			return nil, status.Errorf(codes.AlreadyExists, "api key %q already exists", req.Name)
		}
		return nil, status.Errorf(codes.Internal, "create api key: %s", err)
	}
	return toAPIKeyProto(k, true), nil
}

// ListAPIKeys lists API keys.
func (s *S) ListAPIKeys(
	ctx context.Context,
	req *v1.ListAPIKeysRequest,
) (*v1.ListAPIKeysResponse, error) {
	userInfo, err := s.extractUserInfoFromContext(ctx)
	if err != nil {
		return nil, err
	}

	ks, err := s.store.ListAPIKeysByProjectID(userInfo.ProjectID)
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
	userInfo, err := s.extractUserInfoFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	if err := s.store.DeleteAPIKey(req.Id, userInfo.ProjectID); err != nil {
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
		User: &v1.User{
			Id: k.UserID,
		},
		Organization: &v1.Organization{
			Id: k.OrganizationID,
		},
		Project: &v1.Project{
			Id: k.ProjectID,
		},
	}
	if includeSecret {
		kp.Secret = k.Secret
	}
	return kp
}
