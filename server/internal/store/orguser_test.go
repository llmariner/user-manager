package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrganizationUser(t *testing.T) {
	s, tearDown := NewTest(t)
	defer tearDown()

	_, err := s.CreateOrganization("t1", "o1", "Test Organization")
	assert.NoError(t, err)

	userOrg, err := s.CreateOrganizationUser("t1", "o1", "user1", "r1")
	assert.NoError(t, err)
	assert.NotNil(t, userOrg)

	users, err := s.ListAllOrganizationUsers()
	assert.NoError(t, err)
	assert.Len(t, users, 1)
}
