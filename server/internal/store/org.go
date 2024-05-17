package store

import (
	v1 "github.com/llm-operator/user-manager/api/v1"
	"gorm.io/gorm"
)

// Organization is a model for organization
type Organization struct {
	gorm.Model

	TenantID       string `gorm:"index;uniqueIndex:idx_orgs_tenant_id_title"`
	OrganizationID string `gorm:"uniqueIndex"`

	Title string `gorm:"uniqueIndex:idx_orgs_tenant_id_title"`
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

// GetOrganizationByTenantIDAndTitle gets an organization ID and a title.
func (s *S) GetOrganizationByTenantIDAndTitle(tenantID, title string) (*Organization, error) {
	var org Organization
	if err := s.db.Where("tenant_id = ? AND title = ?", tenantID, title).First(&org).Error; err != nil {
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
	return s.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Unscoped().Where("organization_id = ? AND tenant_id = ?", orgID, tenantID).Delete(&Organization{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return tx.Unscoped().Where("organization_id = ?", orgID).Delete(&OrganizationUser{}).Error
	})
}
