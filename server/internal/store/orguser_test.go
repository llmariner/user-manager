package store

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestOrganizationUser(t *testing.T) {
	s, tearDown := NewTest(t)
	defer tearDown()

	userOrg, err := s.CreateOrganizationUser("o1", "user1", "r1")
	assert.NoError(t, err)
	assert.NotNil(t, userOrg)

	users, err := s.ListAllOrganizationUsers()
	assert.NoError(t, err)
	assert.Len(t, users, 1)

	_, err = s.GetOrganizationUser("o1", "user1")
	assert.NoError(t, err)

	userOrg, err = s.CreateOrganizationUser("o2", "user2", "r1")
	assert.NoError(t, err)
	assert.NotNil(t, userOrg)

	users, err = s.ListAllOrganizationUsers()
	assert.NoError(t, err)
	assert.Len(t, users, 2)

	users, err = s.ListOrganizationUsersByOrganizationID("o1")
	assert.NoError(t, err)
	assert.Len(t, users, 1)
	assert.Equal(t, "user1", users[0].UserID)

	users, err = s.ListOrganizationUsersByOrganizationID("o2")
	assert.NoError(t, err)
	assert.Len(t, users, 1)
	assert.Equal(t, "user2", users[0].UserID)

	err = s.DeleteOrganizationUser("o1", "user1")
	assert.NoError(t, err)

	users, err = s.ListAllOrganizationUsers()
	assert.NoError(t, err)
	assert.Len(t, users, 1)

	err = s.DeleteOrganizationUser("o1", "user1")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
}
