package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/llmariner/common/pkg/aws"
	gerrors "github.com/llmariner/common/pkg/gormlib/errors"
	"github.com/llmariner/common/pkg/id"
	"github.com/llmariner/rbac-manager/pkg/auth"
	v1 "github.com/llmariner/user-manager/api/v1"
	"github.com/llmariner/user-manager/server/internal/config"
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
		return nil, status.Errorf(codes.Internal, "failed to extract user info from context")
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

	if _, err := validateProjectID(s.store, req.ProjectId, req.OrganizationId, userInfo.TenantID); err != nil {
		return nil, err
	}

	if err := s.validateProjectMember(req.ProjectId, req.OrganizationId, userInfo.UserID); err != nil {
		return nil, err
	}

	// TODO(kenji): Make sure this gives sufficient randomness.
	secKey, err := id.GenerateID("sk-", 48)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "generate api key id: %s", err)
	}

	k, err := createAPIKey(ctx, s.store, s.dataKey, req.Name, secKey, userInfo.UserID, req.OrganizationId, req.ProjectId, userInfo.TenantID)
	if err != nil {
		return nil, err
	}
	return k, nil
}

func createAPIKey(
	ctx context.Context,
	st *store.S,
	dataKey []byte,
	name string,
	secKey string,
	userID string,
	organizationID string,
	projectID string,
	tenantID string,
) (*v1.APIKey, error) {
	trackID, err := id.GenerateID("key_", 16)
	if err != nil {
		return nil, fmt.Errorf("generate api key id: %s", err)
	}

	spec := store.APIKeySpec{
		APIKeyID:       trackID,
		TenantID:       tenantID,
		OrganizationID: organizationID,
		ProjectID:      projectID,
		UserID:         userID,
		Name:           name,
	}
	if len(dataKey) > 0 {
		encryptedAPIKey, err := aws.Encrypt(ctx, secKey, trackID, dataKey)
		if err != nil {
			return nil, err
		}
		spec.EncryptedSecret = encryptedAPIKey
	} else {
		spec.Secret = secKey
	}

	k, err := st.CreateAPIKey(spec)
	if err != nil {
		if gerrors.IsUniqueConstraintViolation(err) {
			return nil, status.Errorf(codes.AlreadyExists, "api key %q already exists", name)
		}
		return nil, status.Errorf(codes.Internal, "create api key: %s", err)
	}
	// Do not populate the internal User ID for non-internal gRPC.
	kProto, err := toAPIKeyProto(ctx, st, dataKey, k, "", true)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "to api key proto: %s", err)
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
		return nil, status.Errorf(codes.Internal, "failed to extract user info from context")
	}

	if req.ProjectId == "" {
		return nil, status.Error(codes.InvalidArgument, "project id is required")
	}
	if req.OrganizationId == "" {
		return nil, status.Error(codes.InvalidArgument, "organization id is required")
	}

	if _, err := validateProjectID(s.store, req.ProjectId, req.OrganizationId, userInfo.TenantID); err != nil {
		return nil, err
	}

	// TODO(kenji): Do not allow a project member to read other users' API keys.
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
		kp, err := toAPIKeyProto(ctx, s.store, s.dataKey, k, "", false)
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

	if _, err := validateProjectID(s.store, req.ProjectId, req.OrganizationId, userInfo.TenantID); err != nil {
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

// CreateDefaultAPIKey creates a default API key.
func (s *S) CreateDefaultAPIKey(ctx context.Context, c *config.DefaultAPIKeyConfig, orgID, projectID, tenantID string) error {
	if _, err := s.store.GetAPIKeyByNameAndUserID(c.Name, c.UserID); err == nil {
		// Do nothing.
		return nil
	}

	if _, err := createAPIKey(ctx, s.store, s.dataKey, c.Name, c.Secret, c.UserID, orgID, projectID, tenantID); err != nil {
		return fmt.Errorf("create api key: %s", err)
	}

	return nil
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
		id, ok := internalUserIDs[k.UserID]
		if !ok {
			return nil, status.Errorf(codes.Internal, "internal user ID not found for user %q", k.UserID)
		}
		kp, err := toAPIKeyProto(ctx, s.store, s.dataKey, k, id, true)
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
	ctx context.Context,
	s *store.S,
	dataKey []byte,
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
		if len(dataKey) > 0 {
			secret, err := aws.Decrypt(ctx, k.EncryptedSecret, k.APIKeyID, dataKey)
			if err != nil {
				return nil, err
			}
			kp.Secret = secret
		} else {
			kp.Secret = k.Secret
		}
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
