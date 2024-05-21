package server

import (
	"context"
	"fmt"
	"testing"

	v1 "github.com/llm-operator/user-manager/api/v1"
	"github.com/llm-operator/user-manager/server/internal/config"
	"github.com/llm-operator/user-manager/server/internal/store"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
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
			Title: title,
		})
		assert.NoError(t, err)
		assert.Equal(t, title, cresp.Title)

		_, err = srv.CreateOrganizationUser(ctx, &v1.CreateOrganizationUserRequest{
			OrganizationId: cresp.Id,
			UserId:         "user 1",
			Role:           v1.OrganizationRole_ORGANIZATION_ROLE_OWNER,
		})
		assert.NoError(t, err)
	}

	lresp, err := srv.ListOrganizations(ctx, &v1.ListOrganizationsRequest{})
	assert.NoError(t, err)
	assert.Len(t, lresp.Organizations, 2)

	lresp, err = isrv.ListOrganizations(ctx, &v1.ListOrganizationsRequest{})
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

func TestDeleteOrganization(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	srv := New(st)
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("Authorization", "dummy"))

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

	_, err = srv.DeleteOrganization(ctx, &v1.DeleteOrganizationRequest{
		Id: org.Id,
	})
	assert.Error(t, err)
	assert.Equal(t, codes.FailedPrecondition, status.Code(err))

	// Delete the project and try again.
	_, err = srv.DeleteProject(ctx, &v1.DeleteProjectRequest{
		Id:             proj.Id,
		OrganizationId: org.Id,
	})
	assert.NoError(t, err)

	_, err = srv.DeleteOrganization(ctx, &v1.DeleteOrganizationRequest{
		Id: org.Id,
	})
	assert.NoError(t, err)
}

func TestListOrganizationUsers(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	srv := New(st)
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("Authorization", "dummy"))

	var orgs []*v1.Organization
	for i := 0; i < 2; i++ {
		org, err := srv.CreateOrganization(ctx, &v1.CreateOrganizationRequest{
			Title: fmt.Sprintf("title %d", i),
		})
		assert.NoError(t, err)
		orgs = append(orgs, org)

		_, err = srv.CreateOrganizationUser(ctx, &v1.CreateOrganizationUserRequest{
			OrganizationId: org.Id,
			UserId:         fmt.Sprintf("user %d", i),
			Role:           v1.OrganizationRole_ORGANIZATION_ROLE_OWNER,
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
		Title: "default",
		UserIDs: []string{
			"admin",
		},
	}
	_, err := srv.CreateDefaultOrganization(context.Background(), c)
	assert.NoError(t, err)

	o, err := st.GetOrganizationByTenantIDAndTitle(fakeTenantID, "default")
	assert.NoError(t, err)

	users, err := st.ListAllOrganizationUsers()
	assert.NoError(t, err)
	assert.Len(t, users, 1)
	u := users[0]
	assert.Equal(t, o.OrganizationID, u.OrganizationID)
	assert.Equal(t, "admin", u.UserID)
	assert.Equal(t, v1.OrganizationRole_ORGANIZATION_ROLE_OWNER.String(), u.Role)

	// Calling again is no-op.
	_, err = srv.CreateDefaultOrganization(context.Background(), c)
	assert.NoError(t, err)
}
