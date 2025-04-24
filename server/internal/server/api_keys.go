package server

import (
	"context"
	"errors"
	"fmt"
	"strings"

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
	return s.CreateProjectAPIKey(ctx, req)
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

	key, err := s.store.GetAPIKeyByID(req.Id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "api key %q not found", req.Id)
		}
		return nil, status.Errorf(codes.Internal, "get api key: %s", err)
	}

	isOwner := s.validateProjectOwner(key.ProjectID, key.OrganizationID, userInfo.UserID) == nil
	if !isOwner && userInfo.UserID != key.UserID {
		return nil, status.Errorf(codes.NotFound, "api key %q not found", req.Id)
	}

	if err := s.store.DeleteAPIKey(req.Id, key.ProjectID); err != nil {
		return nil, status.Errorf(codes.Internal, "delete api key: %s", err)
	}
	return &v1.DeleteAPIKeyResponse{
		Id:      req.Id,
		Object:  "users.api_key",
		Deleted: true,
	}, nil
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

	ks, err := s.store.ListAPIKeysByTenantID(userInfo.TenantID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list api keys: %s", err)
	}

	// Show all API keys if the user is an owner. Otherwise only show API keys owned by the user.
	var filtered []*store.APIKey
	for _, k := range ks {
		isOwner := s.validateProjectOwner(k.ProjectID, k.OrganizationID, userInfo.UserID) == nil
		if isOwner || k.UserID == userInfo.UserID {
			filtered = append(filtered, k)
		}
	}

	orgsByID, projectsByID, err := getOrgsAndProjects(s.store, userInfo.TenantID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get orgs and projects: %s", err)
	}

	var apiKeyProtos []*v1.APIKey
	for _, k := range filtered {
		// Do not populate the internal User ID for non-internal gRPC.
		kp, err := toAPIKeyProto(ctx, s.store, s.dataKey, k, "", false, orgsByID, projectsByID)
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

// UpdateAPIKey updates an API key.
func (s *S) UpdateAPIKey(
	ctx context.Context,
	req *v1.UpdateAPIKeyRequest,
) (*v1.APIKey, error) {
	userInfo, ok := auth.ExtractUserInfoFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("failed to extract user info from context")
	}

	if req.ApiKey == nil {
		return nil, status.Error(codes.InvalidArgument, "api key is required")
	}

	if req.ApiKey.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	if req.ApiKey.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	// Currently only support the update of the name field.
	if req.UpdateMask == nil {
		return nil, status.Error(codes.InvalidArgument, "update mask is required")
	}
	if len(req.UpdateMask.Paths) != 1 || req.UpdateMask.Paths[0] != "name" {
		return nil, status.Error(codes.InvalidArgument, "only name field is supported for update")
	}

	key, err := s.store.GetAPIKeyByID(req.ApiKey.Id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "api key %q not found", req.ApiKey.Id)
		}
		return nil, status.Errorf(codes.Internal, "get api key: %s", err)
	}

	isOwner := s.validateProjectOwner(key.ProjectID, key.OrganizationID, userInfo.UserID) == nil
	if !isOwner && userInfo.UserID != key.UserID {
		return nil, status.Errorf(codes.NotFound, "api key %q not found", req.ApiKey.Id)
	}

	key.Name = req.ApiKey.Name
	if err := s.store.UpdateAPIKey(key); err != nil {
		return nil, status.Errorf(codes.Internal, "update api key: %s", err)
	}

	orgsByID, projectsByID, err := getOrgAndProject(s.store, userInfo.TenantID, key.OrganizationID, key.ProjectID)
	if err != nil {
		return nil, err
	}

	kProto, err := toAPIKeyProto(ctx, s.store, s.dataKey, key, "", false, orgsByID, projectsByID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "to api key proto: %s", err)
	}
	return kProto, nil
}

// CreateProjectAPIKey creates an API key.
func (s *S) CreateProjectAPIKey(
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
	if req.IsServiceAccount {
		if req.Role == v1.OrganizationRole_ORGANIZATION_ROLE_UNSPECIFIED {
			return nil, status.Error(codes.InvalidArgument, "role is required for service account")
		}
		if !s.isOrganizationOwner(req.OrganizationId, userInfo.UserID) {
			return nil, status.Error(codes.PermissionDenied, "only organization owner can create service account")
		}
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

	key, err := s.createProjectAPIKey(
		ctx,
		req.Name,
		secKey,
		userInfo.UserID,
		req.OrganizationId,
		req.ProjectId,
		userInfo.TenantID,
		req.IsServiceAccount,
		req.Role,
		req.ExcludedFromRateLimiting,
	)
	if err != nil {
		if gerrors.IsUniqueConstraintViolation(err) {
			return nil, status.Errorf(codes.AlreadyExists, "api key %q already exists", req.Name)
		}
		return nil, status.Errorf(codes.Internal, "create api key: %s", err)
	}
	orgsByID, projectsByID, err := getOrgAndProject(s.store, userInfo.TenantID, req.OrganizationId, req.ProjectId)
	if err != nil {
		return nil, err
	}
	// Do not populate the internal User ID for non-internal gRPC.
	kProto, err := toAPIKeyProto(ctx, s.store, s.dataKey, key, "", true, orgsByID, projectsByID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "to api key proto: %s", err)
	}
	return kProto, nil
}

func (s *S) createProjectAPIKey(
	ctx context.Context,
	name string,
	secKey string,
	userID string,
	organizationID string,
	projectID string,
	tenantID string,
	isServiceAccount bool,
	role v1.OrganizationRole,
	excludedFromRateLimiting bool,
) (*store.APIKey, error) {
	if isServiceAccount {
		var key *store.APIKey
		err := s.store.Transaction(func(tx *gorm.DB) error {
			userID := toServiceAccountUserID(name)
			if _, err := findOrCreateUserInTransaction(tx, userID); err != nil {
				return err
			}
			if _, err := store.CreateOrganizationUserInTransaction(
				tx,
				organizationID,
				userID,
				role.String(),
			); err != nil {
				return err
			}
			if _, err := store.CreateProjectUserInTransaction(tx, store.CreateProjectUserParams{
				ProjectID:      projectID,
				OrganizationID: organizationID,
				UserID:         userID,
				Role:           v1.ProjectRole_PROJECT_ROLE_OWNER,
			}); err != nil {
				return err
			}
			spec, err := createAPIKeySpec(ctx, s.dataKey, name, secKey, userID, organizationID, projectID, tenantID, true, excludedFromRateLimiting)
			if err != nil {
				return err
			}
			key, err = store.CreateAPIKeyInTransaction(tx, spec)
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		return key, nil
	}

	spec, err := createAPIKeySpec(ctx, s.dataKey, name, secKey, userID, organizationID, projectID, tenantID, false, excludedFromRateLimiting)
	if err != nil {
		return nil, err
	}
	return s.store.CreateAPIKey(spec)
}

func createAPIKeySpec(
	ctx context.Context,
	dataKey []byte,
	name string,
	secKey string,
	userID string,
	organizationID string,
	projectID string,
	tenantID string,
	isServiceAccount bool,
	excludedFromRateLimiting bool,
) (store.APIKeySpec, error) {
	trackID, err := id.GenerateID("key_", 16)
	if err != nil {
		return store.APIKeySpec{}, fmt.Errorf("generate api key id: %s", err)
	}
	spec := store.APIKeySpec{
		APIKeyID:                 trackID,
		TenantID:                 tenantID,
		OrganizationID:           organizationID,
		ProjectID:                projectID,
		UserID:                   userID,
		Name:                     name,
		IsServiceAccount:         isServiceAccount,
		ExcludedFromRateLimiting: excludedFromRateLimiting,
	}
	if len(dataKey) > 0 {
		encryptedAPIKey, err := aws.Encrypt(ctx, secKey, trackID, dataKey)
		if err != nil {
			return store.APIKeySpec{}, err
		}
		spec.EncryptedSecret = encryptedAPIKey
	} else {
		spec.Secret = secKey
	}
	return spec, nil
}

// ListProjectAPIKeys lists API keys.
func (s *S) ListProjectAPIKeys(
	ctx context.Context,
	req *v1.ListProjectAPIKeysRequest,
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

	orgsByID, projectsByID, err := getOrgAndProject(s.store, userInfo.TenantID, req.OrganizationId, req.ProjectId)
	if err != nil {
		return nil, err
	}

	var apiKeyProtos []*v1.APIKey
	for _, k := range filtered {
		// Do not populate the internal User ID for non-internal gRPC.
		kp, err := toAPIKeyProto(ctx, s.store, s.dataKey, k, "", false, orgsByID, projectsByID)
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

// DeleteProjectAPIKey deletes an API key.
func (s *S) DeleteProjectAPIKey(
	ctx context.Context,
	req *v1.DeleteProjectAPIKeyRequest,
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

	if key.IsServiceAccount {
		if err := s.store.Transaction(func(tx *gorm.DB) error {
			if err := store.DeleteProjectUserInTransaction(tx, key.ProjectID, key.UserID); err != nil {
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					return fmt.Errorf("delete project user %s: %s", key.UserID, err)
				}
			}
			if err := store.DeleteOrganizationUserInTransaction(tx, req.OrganizationId, key.UserID); err != nil {
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					return fmt.Errorf("delete project user %s: %s", key.UserID, err)
				}
			}
			if err := store.DeleteUserInTransaction(tx, key.UserID); err != nil {
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					return fmt.Errorf("delete user %s: %s", key.UserID, err)
				}
			}
			return store.DeleteAPIKeyInTransaction(tx, req.Id, req.ProjectId)
		}); err != nil {
			return nil, status.Errorf(codes.Internal, "delete api key: %s", err)
		}
	} else {
		if err := s.store.DeleteAPIKey(req.Id, req.ProjectId); err != nil {
			return nil, status.Errorf(codes.Internal, "delete api key: %s", err)
		}
	}

	return &v1.DeleteAPIKeyResponse{
		Id:      req.Id,
		Object:  "users.api_key",
		Deleted: true,
	}, nil
}

// CreateDefaultAPIKey creates a default API key.
func (s *S) CreateDefaultAPIKey(ctx context.Context, c *config.DefaultAPIKeyConfig, orgID, projectID, tenantID string) error {
	userID := c.UserID
	if c.IsServiceAccount {
		userID = toServiceAccountUserID(c.Name)
	}
	if _, err := s.store.GetAPIKeyByNameAndUserID(c.Name, userID); err == nil {
		// Do nothing.
		return nil
	}

	_, err := s.createProjectAPIKey(
		ctx,
		c.Name,
		c.Secret,
		c.UserID,
		orgID,
		projectID,
		tenantID,
		c.IsServiceAccount,
		v1.OrganizationRole_ORGANIZATION_ROLE_TENANT_SYSTEM,
		c.ExcludedFromRateLimiting,
	)
	return err
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
		kp, err := toAPIKeyProto(ctx, s.store, s.dataKey, k, id, true, nil, nil)
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

func getOrgsAndProjects(s *store.S, tenantID string) (map[string]*store.Organization, map[string]*store.Project, error) {
	orgs, err := s.ListOrganizations(tenantID)
	if err != nil {
		return nil, nil, err
	}
	orgsByID := map[string]*store.Organization{}
	for _, o := range orgs {
		orgsByID[o.OrganizationID] = o
	}

	projects, err := s.ListProjectsByTenantID(tenantID)
	if err != nil {
		return nil, nil, err
	}
	projectsByID := map[string]*store.Project{}
	for _, p := range projects {
		projectsByID[p.ProjectID] = p
	}

	return orgsByID, projectsByID, nil
}

func getOrgAndProject(st *store.S, tenantID, organizationID, projectID string) (map[string]*store.Organization, map[string]*store.Project, error) {
	org, err := st.GetOrganizationByTenantIDAndOrgID(tenantID, organizationID)
	if err != nil {
		return nil, nil, status.Errorf(codes.Internal, "get organization: %s", err)
	}
	orgsByID := map[string]*store.Organization{organizationID: org}
	project, err := st.GetProject(store.GetProjectParams{
		TenantID:       tenantID,
		OrganizationID: organizationID,
		ProjectID:      projectID,
	})
	if err != nil {
		return nil, nil, status.Errorf(codes.Internal, "get project: %s", err)
	}
	projectsByID := map[string]*store.Project{projectID: project}
	return orgsByID, projectsByID, nil
}

func toAPIKeyProto(
	ctx context.Context,
	s *store.S,
	dataKey []byte,
	k *store.APIKey,
	internalUserID string,
	showFullSecret bool,
	orgsByID map[string]*store.Organization,
	projectsByID map[string]*store.Project,
) (*v1.APIKey, error) {
	orgRole, err := findOrgRole(s, k.OrganizationID, k.UserID)
	if err != nil {
		return nil, err
	}
	projectRole, err := findProjectRole(s, k.ProjectID, k.UserID)
	if err != nil {
		return nil, err
	}

	var secret string
	if len(dataKey) > 0 {
		secret, err = aws.Decrypt(ctx, k.EncryptedSecret, k.APIKeyID, dataKey)
		if err != nil {
			return nil, err
		}
	} else {
		secret = k.Secret
	}

	if !showFullSecret {
		secret = obfuscateSecret(secret)
	}

	var orgTitle string
	if orgsByID != nil {
		org, ok := orgsByID[k.OrganizationID]
		if !ok {
			return nil, fmt.Errorf("organization %q not found", k.OrganizationID)
		}
		orgTitle = org.Title
	}
	var projectTitle string
	if projectsByID != nil {
		project, ok := projectsByID[k.ProjectID]
		if !ok {
			return nil, fmt.Errorf("project %q not found", k.ProjectID)
		}
		projectTitle = project.Title
	}

	return &v1.APIKey{
		Id:        k.APIKeyID,
		CreatedAt: k.CreatedAt.UTC().Unix(),
		Name:      k.Name,
		Object:    "user.api_key",
		User: &v1.User{
			Id:               k.UserID,
			InternalId:       internalUserID,
			IsServiceAccount: k.IsServiceAccount,
		},
		Organization: &v1.Organization{
			Id:    k.OrganizationID,
			Title: orgTitle,
		},
		Project: &v1.Project{
			Id:    k.ProjectID,
			Title: projectTitle,
		},
		OrganizationRole:         orgRole,
		ProjectRole:              projectRole,
		Secret:                   secret,
		ExcludedFromRateLimiting: k.ExcludedFromRateLimiting,
	}, nil
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

// obfuscateSecret obfuscates the secret.
// The function returns the first 5 characters and the last 2 characters of the secret
// The rest of the characters are replaced with '*' (at most 5 characters are shown).
func obfuscateSecret(secret string) string {
	prefix := secret[0:5]
	suffix := secret[len(secret)-2:]
	starLen := len(secret) - 7
	if starLen > 5 {
		starLen = 5
	}
	obfuscated := strings.Repeat("*", starLen)
	return prefix + obfuscated + suffix
}

func toServiceAccountUserID(name string) string {
	return fmt.Sprintf("system:serviceaccount:%s", name)
}
