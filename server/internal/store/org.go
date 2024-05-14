package store

import (
	"gorm.io/gorm"
)

// Organization is a model for organization
type Organization struct {
	gorm.Model

	TenantID       string `gorm:"index"`
	OrganizationID string `gorm:"uniqueIndex"`

	Title string
}

// CreateOrganization creates a new organization.
func (s *S) CreateOrganization(tenantID, orgID, title string) (*Organization, error) {
	org := &Organization{
		TenantID:       tenantID,
		OrganizationID: orgID,
		Title:          title,
	}
	if err := s.db.Create(org).Error; err != nil {
		return nil, err
	}
	return org, nil
}

// GetOrganization gets an organization.
func (s *S) GetOrganization(orgID string) (*Organization, error) {
	var org Organization
	if err := s.db.Where("organization_id = ?", orgID).First(&org).Error; err != nil {
		return nil, err
	}
	return &org, nil
}

// ListOrganization lists all organizations in the tenant.
func (s *S) ListOrganization(tenantID string) ([]*Organization, error) {
	var orgs []*Organization
	if err := s.db.Where("tenant_id = ?", tenantID).Find(&orgs).Error; err != nil {
		return nil, err
	}
	return orgs, nil
}

// DeleteOrganization deletes an organization.
func (s *S) DeleteOrganization(orgID string) error {
	res := s.db.Unscoped().Where("organization_id = ?", orgID).Delete(&Organization{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
