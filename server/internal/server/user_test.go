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

func TestCreateUserInternal(t *testing.T) {
	tcs := []struct {
		name          string
		enableKMS     bool
		isExistTenant bool
	}{
		{
			name:          "new sso user",
			enableKMS:     true,
			isExistTenant: false,
		},
		{
			name:          "add user to existing org and project",
			enableKMS:     true,
			isExistTenant: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			st, tearDown := store.NewTest(t)
			defer tearDown()

			if tc.isExistTenant {
				org, err := st.CreateOrganization("tenant-1", "org-1", "Title-1", false)
				assert.NoError(t, err)
				_, err = st.CreateProject(store.CreateProjectParams{
					TenantID:            "tenant-1",
					OrganizationID:      org.OrganizationID,
					ProjectID:           "proj-1",
					Title:               "Title-1",
					KubernetesNamespace: "ns-1",
				})
				assert.NoError(t, err)
			}

			kmsClient := aws.NewMockKMSClient()
			var dataKey []byte
			if tc.enableKMS {
				dataKey = kmsClient.DataKey
			}
			isrv := NewInternal(st, dataKey, testr.New(t))

			ctx := fakeAuthInto(context.Background())
			_, err := isrv.CreateUserInternal(ctx, &v1.CreateUserInternalRequest{
				TenantId:           "tenant-1",
				Title:              "Title-1",
				UserId:             "user-1",
				KubernetesNamespac: "ns-1",
			})
			assert.NoError(t, err)

			us, err := st.ListAllUsers()
			assert.NoError(t, err)
			assert.Len(t, us, 1)
			assert.Equal(t, "user-1", us[0].UserID)

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

		})
	}
}
