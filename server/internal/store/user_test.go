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

func TestGetAndList(t *testing.T) {
	st, tearDown := NewTest(t)
	defer tearDown()

	_, err := FindOrCreateUserInTransaction(st.db, "user1", "iuser1")
	assert.NoError(t, err)

	_, err = FindOrCreateUserInTransaction(st.db, "user2", "iuser2")
	assert.NoError(t, err)

	u, err := st.GetUserByUserID("user1")
	assert.NoError(t, err)
	assert.Equal(t, "user1", u.UserID)
	assert.Equal(t, "iuser1", u.InternalUserID)

	users, err := st.ListAllUsers()
	assert.NoError(t, err)
	assert.Len(t, users, 2)
}

func TestDeleteUserInTransaction(t *testing.T) {
	st, tearDown := NewTest(t)
	defer tearDown()

	u1, err := FindOrCreateUserInTransaction(st.db, "user1", "iuser1")
	assert.NoError(t, err)

	err = DeleteUserInTransaction(st.db, u1.UserID)
	assert.NoError(t, err)

	u1Again, err := FindOrCreateUserInTransaction(st.db, "user1", "iuser1")
	assert.NoError(t, err)
	assert.NotEqual(t, u1.ID, u1Again.ID)
}
