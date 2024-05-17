package server

import (
	"context"
	"fmt"
	"testing"

	v1 "github.com/llm-operator/user-manager/api/v1"
	"github.com/llm-operator/user-manager/server/internal/config"
	"github.com/llm-operator/user-manager/server/internal/store"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

func TestOrganization(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	srv := New(st)
	isrv := NewInternal(st)
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("Authorization", "dummy"))

	for i := 0; i < 2; i++ {
		title := fmt.Sprintf("test %d", i)
		cresp, err := srv.CreateOrganization(ctx, &v1.CreateOrganizationRequest{
			Title:               title,
			KubernetesNamespace: fmt.Sprintf("namespace%d", i),
		})
		assert.NoError(t, err)
		assert.Equal(t, title, cresp.Title)

		_, err = srv.CreateOrganizationUser(ctx, &v1.CreateOrganizationUserRequest{
			OrganizationId: cresp.Id,
			UserId:         "user 1",
			Role:           v1.Role_OWNER,
		})
		assert.NoError(t, err)
	}

	lresp, err := srv.ListOrganizations(ctx, &v1.ListOrganizationsRequest{})
	assert.NoError(t, err)
	assert.Len(t, lresp.Organizations, 2)

	laresp, err := isrv.store.ListAllOrganizationUsers()
	assert.NoError(t, err)
	assert.Len(t, laresp, 2)

	_, err = srv.DeleteOrganization(ctx, &v1.DeleteOrganizationRequest{
		Id: lresp.Organizations[0].Id,
	})
	assert.NoError(t, err)

	lresp2, err := srv.ListOrganizations(ctx, &v1.ListOrganizationsRequest{})
	assert.NoError(t, err)
	assert.Len(t, lresp2.Organizations, 1)
	assert.Equal(t, lresp2.Organizations[0].Id, lresp.Organizations[1].Id)

	laresp2, err := isrv.ListOrganizationUsers(ctx, &v1.ListOrganizationUsersRequest{})
	assert.NoError(t, err)
	assert.Len(t, laresp2.Users, 1)

	_, err = srv.DeleteOrganizationUser(ctx, &v1.DeleteOrganizationUserRequest{
		OrganizationId: lresp2.Organizations[0].Id,
		UserId:         "user 1",
	})
	assert.NoError(t, err)

	laresp3, err := isrv.ListOrganizationUsers(ctx, &v1.ListOrganizationUsersRequest{})
	assert.NoError(t, err)
	assert.Empty(t, laresp3.Users)
}

func TestListOrganizationUsers(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	srv := New(st)
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("Authorization", "dummy"))

	var orgs []*v1.Organization
	for i := 0; i < 2; i++ {
		org, err := srv.CreateOrganization(ctx, &v1.CreateOrganizationRequest{
			Title:               fmt.Sprintf("title %d", i),
			KubernetesNamespace: fmt.Sprintf("namespace%d", i),
		})
		assert.NoError(t, err)
		orgs = append(orgs, org)

		_, err = srv.CreateOrganizationUser(ctx, &v1.CreateOrganizationUserRequest{
			OrganizationId: org.Id,
			UserId:         fmt.Sprintf("user %d", i),
			Role:           v1.Role_OWNER,
		})
		assert.NoError(t, err)
	}

	for i := 0; i < 2; i++ {
		resp, err := srv.ListOrganizationUsers(ctx, &v1.ListOrganizationUsersRequest{OrganizationId: orgs[i].Id})
		assert.NoError(t, err)
		assert.Len(t, resp.Users, 1)
		assert.Equal(t, fmt.Sprintf("user %d", i), resp.Users[0].UserId)
	}
}

func TestCreateDefaultOrganization(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	srv := New(st)
	c := &config.DefaultOrganizationConfig{
		Title:               "default",
		KubernetesNamespace: "default",
		UserIDs: []string{
			"admin",
		},
	}
	err := srv.CreateDefaultOrganization(context.Background(), c)
	assert.NoError(t, err)

	o, err := st.GetOrganizationByTenantIDAndTitle(fakeTenantID, "default")
	assert.NoError(t, err)

	users, err := st.ListAllOrganizationUsers()
	assert.NoError(t, err)
	assert.Len(t, users, 1)
	u := users[0]
	assert.Equal(t, o.OrganizationID, u.OrganizationID)
	assert.Equal(t, "admin", u.UserID)
	assert.Equal(t, v1.Role_OWNER.String(), u.Role)

	// Calling again is no-op.
	err = srv.CreateDefaultOrganization(context.Background(), c)
	assert.NoError(t, err)
}
