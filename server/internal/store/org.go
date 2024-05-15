package store

import (
	v1 "github.com/llm-operator/user-manager/api/v1"
	"gorm.io/gorm"
)

// Organization is a model for organization
type Organization struct {
	gorm.Model

	TenantID       string `gorm:"index"`
	OrganizationID string `gorm:"uniqueIndex"`

	Title string
}

// ToProto converts the organization to proto.
func (o *Organization) ToProto() *v1.Organization {
	return &v1.Organization{
		Id:        o.OrganizationID,
		Title:     o.Title,
		CreatedAt: o.CreatedAt.UTC().Unix(),
	}
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

// ListOrganizations lists all organizations in the tenant.
func (s *S) ListOrganizations(tenantID string) ([]*Organization, error) {
	var orgs []*Organization
	if err := s.db.Where("tenant_id = ?", tenantID).Find(&orgs).Error; err != nil {
		return nil, err
	}
	return orgs, nil
}

// DeleteOrganization deletes an organization.
func (s *S) DeleteOrganization(tenantID, orgID string) error {
	res := s.db.Unscoped().Where("organization_id = ? AND tenant_id = ?", orgID, tenantID).Delete(&Organization{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return s.db.Unscoped().Where("organization_id = ?", orgID).Delete(&OrganizationUser{}).Error
}
