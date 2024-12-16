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

			cresp, err := srv.CreateProjectAPIKey(ctx, &v1.CreateAPIKeyRequest{
				Name:           "dummy",
				OrganizationId: org.Id,
				ProjectId:      proj.Id,
			})
			assert.NoError(t, err)
			assert.Equal(t, "dummy", cresp.Name)
			if tc.enableKMS {
				apiKey, err := st.GetAPIKey(cresp.Id, proj.Id)
				assert.NoError(t, err)
				assert.NotEmpty(t, apiKey.EncryptedSecret)
				assert.Empty(t, apiKey.Secret)
				secret, err := aws.Decrypt(ctx, apiKey.EncryptedSecret, apiKey.APIKeyID, kmsClient.DataKey)
				assert.NoError(t, err)
				assert.Equal(t, secret, cresp.Secret)
			} else {
				apiKey, err := st.GetAPIKey(cresp.Id, proj.Id)
				assert.NoError(t, err)
				assert.Empty(t, apiKey.EncryptedSecret)
				assert.NotEmpty(t, apiKey.Secret)
				assert.Equal(t, apiKey.Secret, cresp.Secret)
			}

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
			assert.Len(t, lresp.Data, 1)
			key := lresp.Data[0]
			assert.Empty(t, key.User.InternalId)
			assert.Equal(t, v1.OrganizationRole_ORGANIZATION_ROLE_OWNER, key.OrganizationRole)
			assert.Equal(t, v1.ProjectRole_PROJECT_ROLE_OWNER, key.ProjectRole)

			ilresp, err := isrv.ListInternalAPIKeys(ctx, &v1.ListInternalAPIKeysRequest{})
			assert.NoError(t, err)
			assert.Len(t, ilresp.ApiKeys, 1)
			key = ilresp.ApiKeys[0].ApiKey
			u, err := st.GetUserByUserID(key.User.Id)
			assert.NoError(t, err, "failed to get user by user id", key.User.Id)
			assert.Equal(t, u.InternalUserID, key.User.InternalId)
			assert.Equal(t, v1.OrganizationRole_ORGANIZATION_ROLE_OWNER, key.OrganizationRole)
			assert.Equal(t, v1.ProjectRole_PROJECT_ROLE_OWNER, key.ProjectRole)

			_, err = srv.DeleteProjectAPIKey(ctx, &v1.DeleteProjectAPIKeyRequest{
				Id:             cresp.Id,
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

	// List API keys.

	resp, err := srv.ListProjectAPIKeys(u0Ctx, &v1.ListProjectAPIKeysRequest{
		OrganizationId: org.OrganizationID,
		ProjectId:      proj.Id,
	})
	assert.NoError(t, err)
	assert.Len(t, resp.Data, 2)

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
}

func TestObfuscateSecret(t *testing.T) {
	tcs := []struct {
		secret string
		want   string
	}{
		{
			secret: "sk-1234567890abcdef",
			want:   "sk-12************ef",
		},
	}
	for _, tc := range tcs {
		t.Run(tc.secret, func(t *testing.T) {
			assert.Equal(t, tc.want, obfuscateSecret(tc.secret))
		})
	}
}
