package server

import (
	"context"
	"errors"
	"fmt"

	gerrors "github.com/llmariner/common/pkg/gormlib/errors"
	"github.com/llmariner/common/pkg/id"
	"github.com/llmariner/rbac-manager/pkg/auth"
	v1 "github.com/llmariner/user-manager/api/v1"
	"github.com/llmariner/user-manager/server/internal/store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// CreateAPIKey creates an API key.
func (s *S) CreateAPIKey(
	ctx context.Context,
	req *v1.CreateAPIKeyRequest,
) (*v1.APIKey, error) {
	userInfo, ok := auth.ExtractUserInfoFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("failed to extract user info from context")
	}

	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
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
		TenantID:       userInfo.TenantID,
		OrganizationID: req.OrganizationId,
		ProjectID:      req.ProjectId,
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
	// Do not populate the internal User ID for non-internal gRPC.
	kProto, err := toAPIKeyProto(s.store, k, "", true)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "to api key proto")
	}
	return kProto, nil
}

// ListAPIKeys lists API keys.
func (s *S) ListAPIKeys(
	ctx context.Context,
	req *v1.ListAPIKeysRequest,
) (*v1.ListAPIKeysResponse, error) {
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

	// TODO(kenji): Do not allow a project member to delete other users' API keys.
	if err := s.validateProjectMember(req.ProjectId, req.OrganizationId, userInfo.UserID); err != nil {
		return nil, err
	}

	ks, err := s.store.ListAPIKeysByProjectID(req.ProjectId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list api keys: %s", err)
	}

	// Show all API keys if the user is an owner. Otherwise only show API keys owned by the user.
	var filtered []*store.APIKey
	isOwner := s.validateProjectOwner(req.ProjectId, req.OrganizationId, userInfo.UserID) == nil
	for _, k := range ks {
		if isOwner || k.UserID == userInfo.UserID {
			filtered = append(filtered, k)
		}
	}

	var apiKeyProtos []*v1.APIKey
	for _, k := range filtered {
		// Do not populate the internal User ID for non-internal gRPC.
		kp, err := toAPIKeyProto(s.store, k, "", false)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "to api key proto")
		}
		apiKeyProtos = append(apiKeyProtos, kp)
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
	userInfo, ok := auth.ExtractUserInfoFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("failed to extract user info from context")
	}

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
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

	// TODO(kenji): Do not allow a project member to delete other users' API keys.
	if err := s.validateProjectMember(req.ProjectId, req.OrganizationId, userInfo.UserID); err != nil {
		return nil, err
	}

	// Do not allow a project member to delete other users' API keys.

	key, err := s.store.GetAPIKey(req.Id, req.ProjectId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "api key %q not found", req.Id)
		}
		return nil, status.Errorf(codes.Internal, "get api key: %s", err)
	}

	isOwner := s.validateProjectOwner(req.ProjectId, req.OrganizationId, userInfo.UserID) == nil
	if !isOwner && userInfo.UserID != key.UserID {
		return nil, status.Errorf(codes.NotFound, "api key %q not found", req.Id)
	}

	if err := s.store.DeleteAPIKey(req.Id, req.ProjectId); err != nil {
		return nil, status.Errorf(codes.Internal, "delete api key: %s", err)
	}
	return &v1.DeleteAPIKeyResponse{
		Id:      req.Id,
		Object:  "users.api_key",
		Deleted: true,
	}, nil
}

// ListInternalAPIKeys lists all API keys.
func (s *IS) ListInternalAPIKeys(
	ctx context.Context,
	req *v1.ListInternalAPIKeysRequest,
) (*v1.ListInternalAPIKeysResponse, error) {
	ks, err := s.store.ListAllAPIKeys()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list api keys: %s", err)
	}

	us, err := s.store.ListAllUsers()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list users: %s", err)
	}
	internalUserIDs := map[string]string{}
	for _, u := range us {
		internalUserIDs[u.UserID] = u.InternalUserID
	}

	var apiKeyProtos []*v1.InternalAPIKey
	for _, k := range ks {
		// Gracefully handle a case where the user is not found for backward compatibility.
		// TODO(kenji): Remove once all users are backfilled.
		id := internalUserIDs[k.UserID]
		kp, err := toAPIKeyProto(s.store, k, id, true)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "to api key proto")
		}
		apiKeyProtos = append(apiKeyProtos, &v1.InternalAPIKey{
			ApiKey:   kp,
			TenantId: k.TenantID,
		})
	}
	return &v1.ListInternalAPIKeysResponse{
		ApiKeys: apiKeyProtos,
	}, nil
}

func toAPIKeyProto(
	s *store.S,
	k *store.APIKey,
	internalUserID string,
	includeSecret bool,
) (*v1.APIKey, error) {
	orgRole, err := findOrgRole(s, k.OrganizationID, k.UserID)
	if err != nil {
		return nil, err
	}
	projectRole, err := findProjectRole(s, k.ProjectID, k.UserID)
	if err != nil {
		return nil, err
	}

	kp := &v1.APIKey{
		Id:        k.APIKeyID,
		CreatedAt: k.CreatedAt.UTC().Unix(),
		Name:      k.Name,
		Object:    "user.api_key",
		User: &v1.User{
			Id:         k.UserID,
			InternalId: internalUserID,
		},
		Organization: &v1.Organization{
			Id: k.OrganizationID,
		},
		Project: &v1.Project{
			Id: k.ProjectID,
		},
		OrganizationRole: orgRole,
		ProjectRole:      projectRole,
	}
	if includeSecret {
		kp.Secret = k.Secret
	}
	return kp, nil
}

func findOrgRole(s *store.S, orgID, userID string) (v1.OrganizationRole, error) {
	ou, err := s.GetOrganizationUser(orgID, userID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return v1.OrganizationRole_ORGANIZATION_ROLE_UNSPECIFIED, err
		}
		return v1.OrganizationRole_ORGANIZATION_ROLE_UNSPECIFIED, nil
	}

	return v1.OrganizationRole(v1.OrganizationRole_value[ou.Role]), nil
}

func findProjectRole(s *store.S, projectID, userID string) (v1.ProjectRole, error) {
	pu, err := s.GetProjectUser(projectID, userID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return v1.ProjectRole_PROJECT_ROLE_UNSPECIFIED, err
		}
		return v1.ProjectRole_PROJECT_ROLE_UNSPECIFIED, nil
	}

	return v1.ProjectRole(v1.ProjectRole_value[pu.Role]), nil
}
