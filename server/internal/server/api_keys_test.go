package server

import (
	"context"
	"testing"

	"github.com/go-logr/logr/testr"
	"github.com/llmariner/common/pkg/aws"
	"github.com/llmariner/rbac-manager/pkg/auth"
	v1 "github.com/llmariner/user-manager/api/v1"
	"github.com/llmariner/user-manager/server/internal/store"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"gorm.io/gorm"
)

func TestAPIKey(t *testing.T) {
	tcs := []struct {
		name      string
		enableKMS bool
		secret    string
	}{
		{
			name:      "enable kms",
			enableKMS: true,
			secret:    "secret",
		},
		{
			name:      "disable kms",
			enableKMS: false,
			secret:    "secret",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			st, tearDown := store.NewTest(t)
			defer tearDown()

			kmsClient := aws.NewMockKMSClient()
			var dataKey []byte
			if tc.enableKMS {
				dataKey = kmsClient.DataKey
			}
			srv := New(st, dataKey, testr.New(t))

			ctx := fakeAuthInto(context.Background())
			org, err := srv.CreateOrganization(ctx, &v1.CreateOrganizationRequest{
				Title: "Test organization",
			})
			assert.NoError(t, err)

			proj, err := srv.CreateProject(ctx, &v1.CreateProjectRequest{
				Title:               "Test project",
				OrganizationId:      org.Id,
				KubernetesNamespace: "test",
			})
			assert.NoError(t, err)

			// Test default value of excluded_from_rate_limiting (should be false)
			cresp, err := srv.CreateAPIKey(ctx, &v1.CreateAPIKeyRequest{
				Name:           "dummy",
				OrganizationId: org.Id,
				ProjectId:      proj.Id,
			})
			assert.NoError(t, err)
			assert.Equal(t, "dummy", cresp.Name)
			assert.False(t, cresp.ExcludedFromRateLimiting, "excluded_from_rate_limiting should default to false")

			apiKey, err := st.GetAPIKey(cresp.Id, proj.Id)
			assert.NoError(t, err)
			assert.False(t, apiKey.ExcludedFromRateLimiting, "excluded_from_rate_limiting in database should be false")

			if tc.enableKMS {
				assert.NotEmpty(t, apiKey.EncryptedSecret)
				assert.Empty(t, apiKey.Secret)
				secret, err := aws.Decrypt(ctx, apiKey.EncryptedSecret, apiKey.APIKeyID, kmsClient.DataKey)
				assert.NoError(t, err)
				assert.Equal(t, secret, cresp.Secret)
			} else {
				assert.Empty(t, apiKey.EncryptedSecret)
				assert.NotEmpty(t, apiKey.Secret)
				assert.Equal(t, apiKey.Secret, cresp.Secret)
			}

			// Test creating an API key with excluded_from_rate_limiting set to true
			cresp2, err := srv.CreateAPIKey(ctx, &v1.CreateAPIKeyRequest{
				Name:                     "excluded-key",
				OrganizationId:           org.Id,
				ProjectId:                proj.Id,
				ExcludedFromRateLimiting: true,
			})
			assert.NoError(t, err)
			assert.Equal(t, "excluded-key", cresp2.Name)
			assert.True(t, cresp2.ExcludedFromRateLimiting, "excluded_from_rate_limiting should be true")

			apiKey2, err := st.GetAPIKey(cresp2.Id, proj.Id)
			assert.NoError(t, err)
			assert.True(t, apiKey2.ExcludedFromRateLimiting, "excluded_from_rate_limiting in database should be true")

			_, err = srv.CreateAPIKey(ctx, &v1.CreateAPIKeyRequest{
				Name:           "dummy",
				OrganizationId: org.Id,
				ProjectId:      proj.Id,
			})
			assert.Error(t, err)
			assert.Equal(t, codes.AlreadyExists, status.Code(err))

			lresp, err := srv.ListAPIKeys(ctx, &v1.ListAPIKeysRequest{})
			assert.NoError(t, err)
			assert.Len(t, lresp.Data, 2)

			// Check that the field value is preserved in list responses
			for _, key := range lresp.Data {
				if key.Name == "dummy" {
					assert.False(t, key.ExcludedFromRateLimiting, "regular key should have excluded_from_rate_limiting=false")
				} else if key.Name == "excluded-key" {
					assert.True(t, key.ExcludedFromRateLimiting, "excluded key should have excluded_from_rate_limiting=true")
				}
			}

			_, err = srv.DeleteAPIKey(ctx, &v1.DeleteAPIKeyRequest{Id: cresp.Id})
			assert.NoError(t, err)
			_, err = srv.DeleteAPIKey(ctx, &v1.DeleteAPIKeyRequest{Id: cresp2.Id})
			assert.NoError(t, err)

			lresp, err = srv.ListAPIKeys(ctx, &v1.ListAPIKeysRequest{})
			assert.NoError(t, err)
			assert.Empty(t, lresp.Data)
		})
	}
}

func TestAPIKey_Update(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	srv := New(st, nil, testr.New(t))

	ctx := fakeAuthInto(context.Background())
	org, err := srv.CreateOrganization(ctx, &v1.CreateOrganizationRequest{
		Title: "Test organization",
	})
	assert.NoError(t, err)

	proj, err := srv.CreateProject(ctx, &v1.CreateProjectRequest{
		Title:               "Test project",
		OrganizationId:      org.Id,
		KubernetesNamespace: "test",
	})
	assert.NoError(t, err)

	key, err := srv.CreateAPIKey(ctx, &v1.CreateAPIKeyRequest{
		Name:           "dummy",
		OrganizationId: org.Id,
		ProjectId:      proj.Id,
	})
	assert.NoError(t, err)
	assert.Equal(t, "dummy", key.Name)

	key.Name = "dummy2"
	_, err = srv.UpdateAPIKey(ctx, &v1.UpdateAPIKeyRequest{
		ApiKey: key,
		UpdateMask: &fieldmaskpb.FieldMask{
			Paths: []string{"name"},
		},
	})
	assert.NoError(t, err)

	resp, err := srv.ListAPIKeys(ctx, &v1.ListAPIKeysRequest{})
	assert.NoError(t, err)
	assert.Len(t, resp.Data, 1)
	assert.Equal(t, "dummy2", resp.Data[0].Name)
}

func TestProjectAPIKey(t *testing.T) {
	tcs := []struct {
		name      string
		enableKMS bool
		secret    string
	}{
		{
			name:      "enable kms",
			enableKMS: true,
			secret:    "secret",
		},
		{
			name:      "disable kms",
			enableKMS: false,
			secret:    "secret",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			st, tearDown := store.NewTest(t)
			defer tearDown()

			kmsClient := aws.NewMockKMSClient()
			var dataKey []byte
			if tc.enableKMS {
				dataKey = kmsClient.DataKey
			}
			srv := New(st, dataKey, testr.New(t))
			isrv := NewInternal(st, dataKey, testr.New(t))

			ctx := fakeAuthInto(context.Background())
			org, err := srv.CreateOrganization(ctx, &v1.CreateOrganizationRequest{
				Title: "Test organization",
			})
			assert.NoError(t, err)

			proj, err := srv.CreateProject(ctx, &v1.CreateProjectRequest{
				Title:               "Test project",
				OrganizationId:      org.Id,
				KubernetesNamespace: "test",
			})
			assert.NoError(t, err)

			// Test default value for excluded_from_rate_limiting (should be false)
			cresp, err := srv.CreateProjectAPIKey(ctx, &v1.CreateAPIKeyRequest{
				Name:           "dummy",
				OrganizationId: org.Id,
				ProjectId:      proj.Id,
			})
			assert.NoError(t, err)
			assert.Equal(t, "dummy", cresp.Name)
			assert.False(t, cresp.ExcludedFromRateLimiting, "excluded_from_rate_limiting should default to false")

			apiKey, err := st.GetAPIKey(cresp.Id, proj.Id)
			assert.NoError(t, err)
			assert.False(t, apiKey.ExcludedFromRateLimiting, "excluded_from_rate_limiting in database should be false")

			if tc.enableKMS {
				assert.NotEmpty(t, apiKey.EncryptedSecret)
				assert.Empty(t, apiKey.Secret)
				secret, err := aws.Decrypt(ctx, apiKey.EncryptedSecret, apiKey.APIKeyID, kmsClient.DataKey)
				assert.NoError(t, err)
				assert.Equal(t, secret, cresp.Secret)
			} else {
				assert.Empty(t, apiKey.EncryptedSecret)
				assert.NotEmpty(t, apiKey.Secret)
				assert.Equal(t, apiKey.Secret, cresp.Secret)
			}

			// Test creating an API key with excluded_from_rate_limiting set to true
			gotKeyResp, err := srv.CreateProjectAPIKey(ctx, &v1.CreateAPIKeyRequest{
				Name:                     "excluded-key",
				OrganizationId:           org.Id,
				ProjectId:                proj.Id,
				ExcludedFromRateLimiting: true,
			})
			assert.NoError(t, err)
			assert.Equal(t, "excluded-key", gotKeyResp.Name)
			assert.True(t, gotKeyResp.ExcludedFromRateLimiting, "excluded_from_rate_limiting should be true")

			expAPIKey, err := st.GetAPIKey(gotKeyResp.Id, proj.Id)
			assert.NoError(t, err)
			assert.True(t, expAPIKey.ExcludedFromRateLimiting, "excluded_from_rate_limiting in database should be true")

			_, err = srv.CreateProjectAPIKey(ctx, &v1.CreateAPIKeyRequest{
				Name:           "dummy",
				OrganizationId: org.Id,
				ProjectId:      proj.Id,
			})
			assert.Error(t, err)
			assert.Equal(t, codes.AlreadyExists, status.Code(err))

			lresp, err := srv.ListProjectAPIKeys(ctx, &v1.ListProjectAPIKeysRequest{
				OrganizationId: org.Id,
				ProjectId:      proj.Id,
			})
			assert.NoError(t, err)
			assert.Len(t, lresp.Data, 2)

			// Check that the field value is preserved in list responses
			var regularKeyFound, excludedKeyFound bool
			for _, key := range lresp.Data {
				if key.Name == "dummy" {
					regularKeyFound = true
					assert.False(t, key.ExcludedFromRateLimiting, "regular key should have excluded_from_rate_limiting=false")
					assert.Empty(t, key.User.InternalId)
					assert.Equal(t, v1.OrganizationRole_ORGANIZATION_ROLE_OWNER, key.OrganizationRole)
					assert.Equal(t, v1.ProjectRole_PROJECT_ROLE_OWNER, key.ProjectRole)
					assert.Equal(t, org.Id, key.Organization.Id)
					assert.Equal(t, org.Title, key.Organization.Title)
					assert.Equal(t, proj.Id, key.Project.Id)
					assert.Equal(t, proj.Title, key.Project.Title)
				} else if key.Name == "excluded-key" {
					excludedKeyFound = true
					assert.True(t, key.ExcludedFromRateLimiting, "excluded key should have excluded_from_rate_limiting=true")
				}
			}
			assert.True(t, regularKeyFound, "regular key should be in the list response")
			assert.True(t, excludedKeyFound, "excluded key should be in the list response")

			ilresp, err := isrv.ListInternalAPIKeys(ctx, &v1.ListInternalAPIKeysRequest{})
			assert.NoError(t, err)
			assert.Len(t, ilresp.ApiKeys, 2)

			// Check that the field value is preserved in internal list responses
			regularKeyFound, excludedKeyFound = false, false
			for _, internalKey := range ilresp.ApiKeys {
				key := internalKey.ApiKey
				if key.Name == "dummy" {
					regularKeyFound = true
					assert.False(t, key.ExcludedFromRateLimiting, "regular key should have excluded_from_rate_limiting=false")
					u, err := st.GetUserByUserID(key.User.Id)
					assert.NoError(t, err, "failed to get user by user id", key.User.Id)
					assert.Equal(t, u.InternalUserID, key.User.InternalId)
				} else if key.Name == "excluded-key" {
					excludedKeyFound = true
					assert.True(t, key.ExcludedFromRateLimiting, "excluded key should have excluded_from_rate_limiting=true")
				}
			}
			assert.True(t, regularKeyFound, "regular key should be in the internal list response")
			assert.True(t, excludedKeyFound, "excluded key should be in the internal list response")

			// Clean up
			_, err = srv.DeleteProjectAPIKey(ctx, &v1.DeleteProjectAPIKeyRequest{
				Id:             cresp.Id,
				OrganizationId: org.Id,
				ProjectId:      proj.Id,
			})
			assert.NoError(t, err)

			_, err = srv.DeleteProjectAPIKey(ctx, &v1.DeleteProjectAPIKeyRequest{
				Id:             gotKeyResp.Id,
				OrganizationId: org.Id,
				ProjectId:      proj.Id,
			})
			assert.NoError(t, err)

			lresp, err = srv.ListProjectAPIKeys(ctx, &v1.ListProjectAPIKeysRequest{
				OrganizationId: org.Id,
				ProjectId:      proj.Id,
			})
			assert.NoError(t, err)
			assert.Empty(t, lresp.Data)
		})
	}
}

func TestAPIKey_EnableAuth(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	srv := New(st, nil, testr.New(t))
	srv.enableAuth = true
	org := createDefaultOrg(t, srv, "u0")

	// "u0" is an owner of the organization and the project.
	u0Ctx := auth.AppendUserInfoToContext(context.Background(), auth.UserInfo{
		UserID: "u0",
	})

	proj, err := srv.CreateProject(u0Ctx, &v1.CreateProjectRequest{
		Title:               "title",
		OrganizationId:      org.OrganizationID,
		KubernetesNamespace: "n0",
	})
	assert.NoError(t, err)

	// "u1" is not a member of the project.
	u1Ctx := auth.AppendUserInfoToContext(context.Background(), auth.UserInfo{
		UserID: "u1",
	})

	// "u2" is a member of the project, but not the owner.
	u2Ctx := auth.AppendUserInfoToContext(context.Background(), auth.UserInfo{
		UserID: "u2",
	})
	_, err = srv.CreateOrganizationUser(u0Ctx, &v1.CreateOrganizationUserRequest{
		OrganizationId: org.OrganizationID,
		UserId:         "u2",
		Role:           v1.OrganizationRole_ORGANIZATION_ROLE_READER,
	})
	assert.NoError(t, err)
	_, err = srv.CreateProjectUser(u0Ctx, &v1.CreateProjectUserRequest{
		OrganizationId: org.OrganizationID,
		ProjectId:      proj.Id,
		UserId:         "u2",
		Role:           v1.ProjectRole_PROJECT_ROLE_MEMBER,
	})
	assert.NoError(t, err)

	// Create API keys.

	req := &v1.CreateAPIKeyRequest{
		Name:           "title",
		OrganizationId: org.OrganizationID,
		ProjectId:      proj.Id,
	}

	key0, err := srv.CreateProjectAPIKey(u0Ctx, req)
	assert.NoError(t, err)

	_, err = srv.CreateProjectAPIKey(u1Ctx, req)
	assert.Error(t, err)
	assert.Equal(t, codes.FailedPrecondition, status.Code(err))

	key2, err := srv.CreateProjectAPIKey(u2Ctx, req)
	assert.NoError(t, err)

	saKeyReq := &v1.CreateAPIKeyRequest{
		Name:             "sa",
		OrganizationId:   org.OrganizationID,
		ProjectId:        proj.Id,
		IsServiceAccount: true,
		Role:             v1.OrganizationRole_ORGANIZATION_ROLE_TENANT_SYSTEM,
	}
	key3, err := srv.CreateProjectAPIKey(u0Ctx, saKeyReq)
	assert.NoError(t, err)
	_, err = srv.CreateProjectAPIKey(u0Ctx, saKeyReq)
	assert.Error(t, err)
	assert.Equal(t, codes.AlreadyExists, status.Code(err))
	_, err = st.GetUserByUserID(key3.User.Id)
	assert.NoError(t, err, "user")
	_, err = st.GetOrganizationUser(key3.Organization.Id, key3.User.Id)
	assert.NoError(t, err, "org user")
	_, err = st.GetProjectUser(key3.Project.Id, key3.User.Id)
	assert.NoError(t, err, "project user")

	// List API keys.

	resp, err := srv.ListProjectAPIKeys(u0Ctx, &v1.ListProjectAPIKeysRequest{
		OrganizationId: org.OrganizationID,
		ProjectId:      proj.Id,
	})
	assert.NoError(t, err)
	assert.Len(t, resp.Data, 3)

	_, err = srv.ListProjectAPIKeys(u1Ctx, &v1.ListProjectAPIKeysRequest{
		OrganizationId: org.OrganizationID,
		ProjectId:      proj.Id,
	})
	assert.Error(t, err)
	assert.Equal(t, codes.FailedPrecondition, status.Code(err))

	resp, err = srv.ListProjectAPIKeys(u2Ctx, &v1.ListProjectAPIKeysRequest{
		OrganizationId: org.OrganizationID,
		ProjectId:      proj.Id,
	})
	assert.NoError(t, err)
	assert.Len(t, resp.Data, 1)
	assert.Equal(t, "u2", resp.Data[0].User.Id)

	// Delete API keys.

	// "u2" cannot delete the API key.
	_, err = srv.DeleteProjectAPIKey(u1Ctx, &v1.DeleteProjectAPIKeyRequest{
		OrganizationId: org.OrganizationID,
		ProjectId:      proj.Id,
		Id:             key0.Id,
	})
	assert.Error(t, err)
	assert.Equal(t, codes.FailedPrecondition, status.Code(err))

	// "u2" cannot delete the API key created by "u0".
	_, err = srv.DeleteProjectAPIKey(u2Ctx, &v1.DeleteProjectAPIKeyRequest{
		OrganizationId: org.OrganizationID,
		ProjectId:      proj.Id,
		Id:             key0.Id,
	})
	assert.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))

	// "u2" can delete the API key created by "u2".
	_, err = srv.DeleteProjectAPIKey(u2Ctx, &v1.DeleteProjectAPIKeyRequest{
		OrganizationId: org.OrganizationID,
		ProjectId:      proj.Id,
		Id:             key2.Id,
	})
	assert.NoError(t, err)

	// delete service account key
	_, err = srv.DeleteProjectAPIKey(u0Ctx, &v1.DeleteProjectAPIKeyRequest{
		OrganizationId: org.OrganizationID,
		ProjectId:      proj.Id,
		Id:             key3.Id,
	})
	assert.NoError(t, err)
	_, err = st.GetUserByUserID(key3.User.Id)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	_, err = st.GetOrganizationUser(key3.Organization.Id, key3.User.Id)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	_, err = st.GetProjectUser(key3.Project.Id, key3.User.Id)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestObfuscateSecret(t *testing.T) {
	tcs := []struct {
		secret string
		want   string
	}{
		{
			secret: "sk-1234567890abcdef",
			want:   "sk-12*****ef",
		},
	}
	for _, tc := range tcs {
		t.Run(tc.secret, func(t *testing.T) {
			assert.Equal(t, tc.want, obfuscateSecret(tc.secret))
		})
	}
}
