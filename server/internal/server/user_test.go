package server

import (
	"context"
	"testing"

	"github.com/go-logr/logr/testr"
	"github.com/llmariner/common/pkg/aws"
	v1 "github.com/llmariner/user-manager/api/v1"
	"github.com/llmariner/user-manager/server/internal/store"
	"github.com/stretchr/testify/assert"
)

func TestCreateInternalUser(t *testing.T) {
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
			isrv := NewInternal(st, dataKey, testr.New(t))

			ctx := fakeAuthInto(context.Background())
			_, err := isrv.CreateInternalUser(ctx, &v1.CreateInternalUserRequest{
				TenantId: "tenant-1",
				Title:    "Title-1",
				UserId:   "user-1",
			})
			assert.NoError(t, err)

			oResp, err := isrv.ListInternalOrganizations(ctx, &v1.ListInternalOrganizationsRequest{})
			assert.NoError(t, err)
			assert.Len(t, oResp.Organizations, 1)
			for _, org := range oResp.Organizations {
				assert.Equal(t, "tenant-1", org.TenantId)
			}
			org := oResp.Organizations[0].Organization
			orgUsers, err := isrv.ListOrganizationUsers(ctx, &v1.ListOrganizationUsersRequest{
				OrganizationId: org.Id,
			})
			assert.NoError(t, err)
			assert.Len(t, orgUsers.Users, 1)
			assert.Equal(t, "user-1", orgUsers.Users[0].UserId)

			projs, err := isrv.ListProjects(ctx, &v1.ListProjectsRequest{
				OrganizationId: org.Id,
			})
			assert.NoError(t, err)
			assert.Len(t, projs.Projects, 1)
			proj := projs.Projects[0]
			projUsers, err := isrv.ListProjectUsers(ctx, &v1.ListProjectUsersRequest{
				OrganizationId: org.Id,
				ProjectId:      proj.Id,
			})
			assert.NoError(t, err)
			assert.Len(t, projUsers.Users, 1)
			assert.Equal(t, "user-1", projUsers.Users[0].UserId)

			aResp, err := isrv.ListInternalAPIKeys(ctx, &v1.ListInternalAPIKeysRequest{})
			assert.NoError(t, err)
			assert.Len(t, aResp.ApiKeys, 1)
			key := aResp.ApiKeys[0].ApiKey
			tenantID := aResp.ApiKeys[0].TenantId
			assert.NotEmpty(t, key.User.InternalId)
			assert.Equal(t, v1.OrganizationRole_ORGANIZATION_ROLE_OWNER, key.OrganizationRole)
			assert.Equal(t, v1.ProjectRole_PROJECT_ROLE_OWNER, key.ProjectRole)
			assert.Equal(t, "user-1", key.User.Id)
			assert.Equal(t, org.Id, key.Organization.Id)
			assert.Equal(t, proj.Id, key.Project.Id)
			assert.Equal(t, "tenant-1", tenantID)
		})
	}
}
