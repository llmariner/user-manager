package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAPIKey(t *testing.T) {
	st, tearDown := NewTest(t)
	defer tearDown()

	k0, err := st.CreateAPIKey(APIKeySpec{
		Key: APIKeyKey{
			APIKeyID:       "k0",
			TenantID:       "t0",
			OrganizationID: "o0",
			UserID:         "u0",
		},
		Name:   "n0",
		Secret: "s0",
	})
	assert.NoError(t, err)
	assert.Equal(t, "k0", k0.APIKeyID)
	assert.Equal(t, "t0", k0.TenantID)
	assert.Equal(t, "n0", k0.Name)
	assert.Equal(t, "s0", k0.Secret)
	assert.Equal(t, "o0", k0.OrganizationID)
	assert.Equal(t, "u0", k0.UserID)

	got, err := st.ListAPIKeysByTenantID("t0")
	assert.NoError(t, err)
	assert.Len(t, got, 1)
	assert.Equal(t, "k0", got[0].APIKeyID)

	got, err = st.ListAPIKeysByTenantID("t1")
	assert.NoError(t, err)
	assert.Empty(t, got)

	got, err = st.ListAllAPIKeys()
	assert.NoError(t, err)
	assert.Len(t, got, 1)
	assert.Equal(t, "k0", got[0].APIKeyID)

	k1, err := st.CreateAPIKey(APIKeySpec{
		Key: APIKeyKey{
			APIKeyID: "k1",
			TenantID: "t1",
		},
		Name:   "n1",
		Secret: "s1",
	})
	assert.NoError(t, err)
	assert.Equal(t, "k1", k1.APIKeyID)
	assert.Equal(t, "t1", k1.TenantID)
	assert.Equal(t, "n1", k1.Name)
	assert.Equal(t, "s1", k1.Secret)

	got, err = st.ListAPIKeysByTenantID("t0")
	assert.NoError(t, err)
	assert.Len(t, got, 1)
	assert.Equal(t, "k0", got[0].APIKeyID)

	got, err = st.ListAPIKeysByTenantID("t1")
	assert.NoError(t, err)
	assert.Len(t, got, 1)
	assert.Equal(t, "k1", got[0].APIKeyID)

	got, err = st.ListAllAPIKeys()
	assert.NoError(t, err)
	assert.Len(t, got, 2)

	err = st.DeleteAPIKey(APIKeyKey{
		APIKeyID: "k0",
		TenantID: "t0",
	})
	assert.NoError(t, err)

	got, err = st.ListAPIKeysByTenantID("t0")
	assert.NoError(t, err)
	assert.Empty(t, got)

	got, err = st.ListAPIKeysByTenantID("t1")
	assert.NoError(t, err)
	assert.Len(t, got, 1)
	assert.Equal(t, "k1", got[0].APIKeyID)
}

func TestAPIKeySameName(t *testing.T) {
	st, tearDown := NewTest(t)
	defer tearDown()

	_, err := st.CreateAPIKey(APIKeySpec{
		Key: APIKeyKey{
			APIKeyID: "k1",
			TenantID: "t1",
		},
		Name:   "n1",
		Secret: "s1",
	})
	assert.NoError(t, err)

	_, err = st.CreateAPIKey(APIKeySpec{
		Key: APIKeyKey{
			APIKeyID: "k2",
			TenantID: "t1",
		},
		Name:   "n1",
		Secret: "s2",
	})
	assert.Error(t, err)

	// A different tenant can have the same name.
	_, err = st.CreateAPIKey(APIKeySpec{
		Key: APIKeyKey{
			APIKeyID: "k3",
			TenantID: "t2",
		},
		Name:   "n1",
		Secret: "s3",
	})
	assert.NoError(t, err)
}
