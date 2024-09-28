package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindOrCreateUserInTransaction(t *testing.T) {
	st, tearDown := NewTest(t)
	defer tearDown()

	u1, err := FindOrCreateUserInTransaction(st.db, "user1", "iuser1")
	assert.NoError(t, err)

	_, err = FindOrCreateUserInTransaction(st.db, "user2", "iuser2")
	assert.NoError(t, err)

	u1Again, err := FindOrCreateUserInTransaction(st.db, "user1", "iuser1")
	assert.NoError(t, err)
	assert.Equal(t, u1.ID, u1Again.ID)
}
