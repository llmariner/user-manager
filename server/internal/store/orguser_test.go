package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
}
