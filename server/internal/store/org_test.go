package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrganization(t *testing.T) {
	s, tearDown := NewTest(t)
	defer tearDown()

	org := &Organization{
		TenantID:       "t1",
		OrganizationID: "o1",
		Title:          "Test Organization",
	}

	gotOrg, err := s.CreateOrganization(org.TenantID, org.OrganizationID, org.Title)
	assert.NoError(t, err)
	assert.NotNil(t, gotOrg)
	assert.Equal(t, org.TenantID, gotOrg.TenantID)
	assert.Equal(t, org.OrganizationID, gotOrg.OrganizationID)
	assert.Equal(t, org.Title, gotOrg.Title)

	gotOrg, err = s.GetOrganization(org.OrganizationID)
	assert.NoError(t, err)
	assert.NotNil(t, gotOrg)
	assert.Equal(t, org.TenantID, gotOrg.TenantID)
	assert.Equal(t, org.OrganizationID, gotOrg.OrganizationID)
	assert.Equal(t, org.Title, gotOrg.Title)

	_, err = s.CreateOrganization(org.TenantID, "o2", "Test Organization 2")
	assert.NoError(t, err)
	gotOrgs, err := s.ListOrganization(org.TenantID)
	assert.NoError(t, err)
	assert.Len(t, gotOrgs, 2)

	err = s.DeleteOrganization(org.OrganizationID)
	assert.NoError(t, err)
	gotOrg, err = s.GetOrganization(org.OrganizationID)
	assert.Error(t, err)
	assert.Nil(t, gotOrg)
}
