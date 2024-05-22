package store

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestAPIKey(t *testing.T) {
	st, tearDown := NewTest(t)
	defer tearDown()

	_, err := st.GetAPIKeyByNameAndProjectID("n0", "p0")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))

	k0, err := st.CreateAPIKey(APIKeySpec{
		APIKeyID:       "k0",
		TenantID:       "t0",
		OrganizationID: "o0",
		ProjectID:      "p0",
		UserID:         "u0",
		Name:           "n0",
		Secret:         "s0",
	})
	assert.NoError(t, err)
	assert.Equal(t, "k0", k0.APIKeyID)
	assert.Equal(t, "t0", k0.TenantID)
	assert.Equal(t, "n0", k0.Name)
	assert.Equal(t, "s0", k0.Secret)
	assert.Equal(t, "o0", k0.OrganizationID)
	assert.Equal(t, "p0", k0.ProjectID)
	assert.Equal(t, "u0", k0.UserID)

	k, err := st.GetAPIKeyByNameAndProjectID("n0", "p0")
	assert.NoError(t, err)
	assert.Equal(t, "k0", k.APIKeyID)

	got, err := st.ListAPIKeysByProjectID("p0")
	assert.NoError(t, err)
	assert.Len(t, got, 1)
	assert.Equal(t, "k0", got[0].APIKeyID)

	got, err = st.ListAPIKeysByProjectID("p1")
	assert.NoError(t, err)
	assert.Empty(t, got)

	got, err = st.ListAllAPIKeys()
	assert.NoError(t, err)
	assert.Len(t, got, 1)
	assert.Equal(t, "k0", got[0].APIKeyID)

	k1, err := st.CreateAPIKey(APIKeySpec{
		APIKeyID:       "k1",
		TenantID:       "t1",
		OrganizationID: "o1",
		ProjectID:      "p1",
		UserID:         "u1",
		Name:           "n1",
		Secret:         "s1",
	})
	assert.NoError(t, err)
	assert.Equal(t, "k1", k1.APIKeyID)
	assert.Equal(t, "t1", k1.TenantID)
	assert.Equal(t, "n1", k1.Name)
	assert.Equal(t, "s1", k1.Secret)
	assert.Equal(t, "o1", k1.OrganizationID)
	assert.Equal(t, "p1", k1.ProjectID)
	assert.Equal(t, "u1", k1.UserID)

	got, err = st.ListAPIKeysByProjectID("p0")
	assert.NoError(t, err)
	assert.Len(t, got, 1)
	assert.Equal(t, "k0", got[0].APIKeyID)

	got, err = st.ListAPIKeysByProjectID("p1")
	assert.NoError(t, err)
	assert.Len(t, got, 1)
	assert.Equal(t, "k1", got[0].APIKeyID)

	got, err = st.ListAllAPIKeys()
	assert.NoError(t, err)
	assert.Len(t, got, 2)

	err = st.DeleteAPIKey("k0", "p0")
	assert.NoError(t, err)

	got, err = st.ListAPIKeysByProjectID("p0")
	assert.NoError(t, err)
	assert.Empty(t, got)

	got, err = st.ListAPIKeysByProjectID("p1")
	assert.NoError(t, err)
	assert.Len(t, got, 1)
	assert.Equal(t, "k1", got[0].APIKeyID)
}

func TestAPIKeySameName(t *testing.T) {
	st, tearDown := NewTest(t)
	defer tearDown()

	_, err := st.CreateAPIKey(APIKeySpec{
		APIKeyID: "k1",
		TenantID: "t1",
		Name:     "n1",
		Secret:   "s1",
	})
	assert.NoError(t, err)

	_, err = st.CreateAPIKey(APIKeySpec{
		APIKeyID: "k2",
		TenantID: "t1",
		Name:     "n1",
		Secret:   "s2",
	})
	assert.Error(t, err)

	// A different tenant can have the same name.
	_, err = st.CreateAPIKey(APIKeySpec{
		APIKeyID: "k3",
		TenantID: "t2",
		Name:     "n1",
		Secret:   "s3",
	})
	assert.NoError(t, err)
}
