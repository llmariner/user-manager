package server

import (
	"testing"

	"github.com/go-logr/logr/testr"
	v1 "github.com/llmariner/user-manager/api/v1"
	"github.com/llmariner/user-manager/server/internal/store"
	"github.com/stretchr/testify/assert"
)

func TestOrgRole(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	srv := New(st, testr.New(t))
	srv.enableAuth = true

	_, err := st.CreateOrganizationUser("org1", "user1", v1.OrganizationRole_ORGANIZATION_ROLE_OWNER.String())
	assert.NoError(t, err)

	got := srv.organizationRole("org1", "user1")
	assert.Equal(t, v1.OrganizationRole_ORGANIZATION_ROLE_OWNER, got)

	got = srv.organizationRole("org1", "user2")
	assert.Equal(t, v1.OrganizationRole_ORGANIZATION_ROLE_UNSPECIFIED, got)

	got = srv.organizationRole("org2", "user1")
	assert.Equal(t, v1.OrganizationRole_ORGANIZATION_ROLE_UNSPECIFIED, got)
}

func TestProjectRole(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	srv := New(st, testr.New(t))
	srv.enableAuth = true

	_, err := st.CreateProjectUser(store.CreateProjectUserParams{
		OrganizationID: "org1",
		ProjectID:      "proj1",
		UserID:         "user1",
		Role:           v1.ProjectRole_PROJECT_ROLE_OWNER,
	})
	assert.NoError(t, err)

	got := srv.projectRole("proj1", "user1")
	assert.Equal(t, v1.ProjectRole_PROJECT_ROLE_OWNER, got)

	got = srv.projectRole("proj1", "user2")
	assert.Equal(t, v1.ProjectRole_PROJECT_ROLE_UNSPECIFIED, got)

	got = srv.projectRole("proj2", "user1")
	assert.Equal(t, v1.ProjectRole_PROJECT_ROLE_UNSPECIFIED, got)
}
