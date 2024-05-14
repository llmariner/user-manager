package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateUserOrganization(t *testing.T) {
	s, tearDown := NewTest(t)
	defer tearDown()

	_, err := s.CreateOrganization("t1", "o1", "Test Organization")
	assert.NoError(t, err)

	userOrg, err := s.CreateUserOrganization("o1", "user1")
	assert.NoError(t, err)
	assert.NotNil(t, userOrg)
}
