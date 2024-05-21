package store

import (
	"errors"
	"testing"

	gerrors "github.com/llm-operator/common/pkg/gormlib/errors"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
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

	gotOrg, err = s.GetOrganizationByTenantIDAndTitle(org.TenantID, org.Title)
	assert.NoError(t, err)
	assert.NotNil(t, gotOrg)
	assert.Equal(t, org.TenantID, gotOrg.TenantID)
	assert.Equal(t, org.OrganizationID, gotOrg.OrganizationID)
	assert.Equal(t, org.Title, gotOrg.Title)

	_, err = s.GetOrganizationByTenantIDAndTitle("different", org.Title)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))

	_, err = s.GetOrganizationByTenantIDAndTitle(org.TenantID, "different")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))

	_, err = s.CreateOrganization(org.TenantID, "o2", "Test Organization 2", false)
	assert.NoError(t, err)
	gotOrgs, err := s.ListOrganizations(org.TenantID)
	assert.NoError(t, err)
	assert.Len(t, gotOrgs, 2)

	err = s.DeleteOrganization(org.TenantID, org.OrganizationID)
	assert.NoError(t, err)
	gotOrg, err = s.GetOrganizationByTenantIDAndOrgID(org.TenantID, org.OrganizationID)
	assert.Error(t, err)
	assert.Nil(t, gotOrg)
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
