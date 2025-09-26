package store

import (
	"errors"
	"testing"

	v1 "github.com/llmariner/user-manager/api/v1"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestProjectUser(t *testing.T) {
	s, tearDown := NewTest(t)
	defer tearDown()

	userOrg, err := s.CreateProjectUser(CreateProjectUserParams{
		ProjectID:      "p1",
		OrganizationID: "o1",
		UserID:         "user1",
		Role:           v1.ProjectRole_PROJECT_ROLE_OWNER,
	})
	assert.NoError(t, err)
	assert.NotNil(t, userOrg)

	users, err := s.ListAllProjectUsers()
	assert.NoError(t, err)
	assert.Len(t, users, 1)

	_, err = s.GetProjectUser("p1", "user1")
	assert.NoError(t, err)

	userOrg, err = s.CreateProjectUser(CreateProjectUserParams{
		ProjectID:      "p2",
		OrganizationID: "o2",
		UserID:         "user2",
		Role:           v1.ProjectRole_PROJECT_ROLE_OWNER,
	})
	assert.NoError(t, err)
	assert.NotNil(t, userOrg)

	users, err = s.ListAllProjectUsers()
	assert.NoError(t, err)
	assert.Len(t, users, 2)

	users, err = s.ListProjectUsersByProjectID("p1")
	assert.NoError(t, err)
	assert.Len(t, users, 1)
	assert.Equal(t, "user1", users[0].UserID)

	users, err = s.ListProjectUsersByProjectID("p2")
	assert.NoError(t, err)
	assert.Len(t, users, 1)
	assert.Equal(t, "user2", users[0].UserID)

	users, err = s.ListProjectUsersByUserID("user1")
	assert.NoError(t, err)
	assert.Len(t, users, 1)
	assert.Equal(t, "user1", users[0].UserID)
	assert.Equal(t, "p1", users[0].ProjectID)

	err = s.DeleteProjectUser("p1", "user1")
	assert.NoError(t, err)

	users, err = s.ListAllProjectUsers()
	assert.NoError(t, err)
	assert.Len(t, users, 1)

	err = s.DeleteProjectUser("p1", "user1")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
}

func TestHideProjectUser(t *testing.T) {
	s, tearDown := NewTest(t)
	defer tearDown()

	pu, err := s.CreateProjectUser(CreateProjectUserParams{
		ProjectID:      "p1",
		OrganizationID: "o1",
		UserID:         "user1",
		Role:           v1.ProjectRole_PROJECT_ROLE_OWNER,
	})
	assert.NoError(t, err)
	assert.False(t, pu.Hidden)

	err = s.HideProjectUser("p1", "user1")
	assert.NoError(t, err)

	pu, err = s.GetProjectUser("p1", "user1")
	assert.NoError(t, err)
	assert.True(t, pu.Hidden)
}
