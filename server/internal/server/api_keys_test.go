package server

import (
	"context"
	"testing"

	"github.com/llmariner/rbac-manager/pkg/auth"
	v1 "github.com/llm-operator/user-manager/api/v1"
	"github.com/llm-operator/user-manager/server/internal/store"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestAPIKey(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	srv := New(st)
	isrv := NewInternal(st)

	ctx := context.Background()
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

	cresp, err := srv.CreateAPIKey(ctx, &v1.CreateAPIKeyRequest{
		Name:           "dummy",
		OrganizationId: org.Id,
		ProjectId:      proj.Id,
	})
	assert.NoError(t, err)
	assert.Equal(t, "dummy", cresp.Name)

	_, err = srv.CreateAPIKey(ctx, &v1.CreateAPIKeyRequest{
		Name:           "dummy",
		OrganizationId: org.Id,
		ProjectId:      proj.Id,
	})
	assert.Error(t, err)
	assert.Equal(t, codes.AlreadyExists, status.Code(err))

	lresp, err := srv.ListAPIKeys(ctx, &v1.ListAPIKeysRequest{
		OrganizationId: org.Id,
		ProjectId:      proj.Id,
	})
	assert.NoError(t, err)
	assert.Len(t, lresp.Data, 1)

	ilresp, err := isrv.ListInternalAPIKeys(ctx, &v1.ListInternalAPIKeysRequest{})
	assert.NoError(t, err)
	assert.Len(t, ilresp.ApiKeys, 1)

	_, err = srv.DeleteAPIKey(ctx, &v1.DeleteAPIKeyRequest{
		Id:             cresp.Id,
		OrganizationId: org.Id,
		ProjectId:      proj.Id,
	})
	assert.NoError(t, err)

	lresp, err = srv.ListAPIKeys(ctx, &v1.ListAPIKeysRequest{
		OrganizationId: org.Id,
		ProjectId:      proj.Id,
	})
	assert.NoError(t, err)
	assert.Empty(t, lresp.Data)
}

func TestAPIKey_EnableAuth(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	srv := New(st)
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

	key0, err := srv.CreateAPIKey(u0Ctx, req)
	assert.NoError(t, err)

	_, err = srv.CreateAPIKey(u1Ctx, req)
	assert.Error(t, err)
	assert.Equal(t, codes.FailedPrecondition, status.Code(err))

	key2, err := srv.CreateAPIKey(u2Ctx, req)
	assert.NoError(t, err)

	// List API keys.

	resp, err := srv.ListAPIKeys(u0Ctx, &v1.ListAPIKeysRequest{
		OrganizationId: org.OrganizationID,
		ProjectId:      proj.Id,
	})
	assert.NoError(t, err)
	assert.Len(t, resp.Data, 2)

	_, err = srv.ListAPIKeys(u1Ctx, &v1.ListAPIKeysRequest{
		OrganizationId: org.OrganizationID,
		ProjectId:      proj.Id,
	})
	assert.Error(t, err)
	assert.Equal(t, codes.FailedPrecondition, status.Code(err))

	resp, err = srv.ListAPIKeys(u2Ctx, &v1.ListAPIKeysRequest{
		OrganizationId: org.OrganizationID,
		ProjectId:      proj.Id,
	})
	assert.NoError(t, err)
	assert.Len(t, resp.Data, 1)
	assert.Equal(t, "u2", resp.Data[0].User.Id)

	// Delete API keys.

	// "u2" cannot delete the API key.
	_, err = srv.DeleteAPIKey(u1Ctx, &v1.DeleteAPIKeyRequest{
		OrganizationId: org.OrganizationID,
		ProjectId:      proj.Id,
		Id:             key0.Id,
	})
	assert.Error(t, err)
	assert.Equal(t, codes.FailedPrecondition, status.Code(err))

	// "u2" cannot delete the API key created by "u0".
	_, err = srv.DeleteAPIKey(u2Ctx, &v1.DeleteAPIKeyRequest{
		OrganizationId: org.OrganizationID,
		ProjectId:      proj.Id,
		Id:             key0.Id,
	})
	assert.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))

	// "u2" can delete the API key created by "u2".
	_, err = srv.DeleteAPIKey(u2Ctx, &v1.DeleteAPIKeyRequest{
		OrganizationId: org.OrganizationID,
		ProjectId:      proj.Id,
		Id:             key2.Id,
	})
	assert.NoError(t, err)
}
