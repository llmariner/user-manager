package store

import (
	"testing"

	gerrors "github.com/llmariner/common/pkg/gormlib/errors"
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

	gotOrg, err := s.CreateOrganization(org.TenantID, org.OrganizationID, org.Title, false)
	assert.NoError(t, err)
	assert.NotNil(t, gotOrg)
	assert.Equal(t, org.TenantID, gotOrg.TenantID)
	assert.Equal(t, org.OrganizationID, gotOrg.OrganizationID)
	assert.Equal(t, org.Title, gotOrg.Title)

	gotOrg, err = s.GetOrganizationByTenantIDAndOrgID(org.TenantID, org.OrganizationID)
	assert.NoError(t, err)
	assert.NotNil(t, gotOrg)
	assert.Equal(t, org.TenantID, gotOrg.TenantID)
	assert.Equal(t, org.OrganizationID, gotOrg.OrganizationID)
	assert.Equal(t, org.Title, gotOrg.Title)

	_, err = s.CreateOrganization(org.TenantID, "o2", "Test Organization 2", false)
	assert.NoError(t, err)
	gotOrgs, err := s.ListOrganizations(org.TenantID)
	assert.NoError(t, err)
	assert.Len(t, gotOrgs, 2)

	err = s.DeleteOrganization(org.OrganizationID)
	assert.NoError(t, err)
	gotOrg, err = s.GetOrganizationByTenantIDAndOrgID(org.TenantID, org.OrganizationID)
	assert.Error(t, err)
	assert.Nil(t, gotOrg)
}

func TestGetDefaultOrganization(t *testing.T) {
	s, tearDown := NewTest(t)
	defer tearDown()

	orgs := []*Organization{
		{
			TenantID:       "t0",
			OrganizationID: "o0",
			Title:          "Test Organization 0",
			IsDefault:      true,
		},
		{
			TenantID:       "t0",
			OrganizationID: "o1",
			Title:          "Test Organization 1",
			IsDefault:      false,
		},
		{
			TenantID:       "t2",
			OrganizationID: "o2",
			Title:          "Test Organization 2",
			IsDefault:      true,
		},
	}
	for _, org := range orgs {
		_, err := s.CreateOrganization(org.TenantID, org.OrganizationID, org.Title, org.IsDefault)
		assert.NoError(t, err)
	}

	tcs := []struct {
		tenantID  string
		wantOrgID string
	}{
		{
			tenantID:  "t0",
			wantOrgID: "o0",
		},
		{
			tenantID:  "t2",
			wantOrgID: "o2",
		},
	}

	for _, tc := range tcs {
		got, err := s.GetDefaultOrganization(tc.tenantID)
		assert.NoError(t, err)
		assert.Equal(t, tc.wantOrgID, got.OrganizationID)
	}
}

func TestListAllOrganizations(t *testing.T) {
	s, tearDown := NewTest(t)
	defer tearDown()

	_, err := s.CreateOrganization("t0", "o0", "Test Organization 0", false)
	assert.NoError(t, err)

	_, err = s.CreateOrganization("t1", "o1", "Test Organization 1", false)
	assert.NoError(t, err)

	gotOrgs, err := s.ListAllOrganizations()
	assert.NoError(t, err)
	assert.Len(t, gotOrgs, 2)
}

func TestOrganization_UniqueConstraint(t *testing.T) {
	s, tearDown := NewTest(t)
	defer tearDown()

	_, err := s.CreateOrganization("t1", "o1", "Test Organization", false)
	assert.NoError(t, err)

	_, err = s.CreateOrganization("t1", "o2", "Test Organization", false)
	assert.Error(t, err)
	assert.True(t, gerrors.IsUniqueConstraintViolation(err))

	_, err = s.CreateOrganization("t2", "o3", "Test Organization", false)
	assert.NoError(t, err)
}
