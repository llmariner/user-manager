package store

import (
	v1 "github.com/llmariner/user-manager/api/v1"
	"gorm.io/gorm"
)

// Organization is a model for organization
type Organization struct {
	gorm.Model

	OrganizationID string `gorm:"uniqueIndex"`

	TenantID string `gorm:"uniqueIndex:idx_orgs_tenant_id_title"`
	Title    string `gorm:"uniqueIndex:idx_orgs_tenant_id_title"`

	IsDefault bool
}

// ToProto converts the organization to proto.
func (o *Organization) ToProto() *v1.Organization {
	return &v1.Organization{
		Id:        o.OrganizationID,
		Title:     o.Title,
		CreatedAt: o.CreatedAt.UTC().Unix(),
		IsDefault: o.IsDefault,
	}
}

// CreateOrganization creates a new organization.
func (s *S) CreateOrganization(tenantID, orgID, title string, isDefault bool) (*Organization, error) {
	return CreateOrganizationInTransaction(s.db, tenantID, orgID, title, isDefault)
}

// CreateOrganizationInTransaction creates a new organization in a transaction.
func CreateOrganizationInTransaction(tx *gorm.DB, tenantID, orgID, title string, isDefault bool) (*Organization, error) {
	org := &Organization{
		TenantID:       tenantID,
		OrganizationID: orgID,
		Title:          title,
		IsDefault:      isDefault,
	}
	if err := tx.Create(org).Error; err != nil {
		return nil, err
	}
	return org, nil
}

// GetOrganizationByTenantIDAndOrgID gets an organization.
func (s *S) GetOrganizationByTenantIDAndOrgID(tenantID, orgID string) (*Organization, error) {
	var org Organization
	if err := s.db.Where("tenant_id = ? AND organization_id = ?", tenantID, orgID).First(&org).Error; err != nil {
		return nil, err
	}
	return &org, nil
}

// GetDefaultOrganization gets a default organization.
func (s *S) GetDefaultOrganization(tenantID string) (*Organization, error) {
	var org Organization
	if err := s.db.Where("tenant_id = ? AND is_default = ?", tenantID, true).First(&org).Error; err != nil {
		return nil, err
	}
	return &org, nil
}

// ListOrganizations lists all organizations in the tenant.
func (s *S) ListOrganizations(tenantID string) ([]*Organization, error) {
	var orgs []*Organization
	if err := s.db.Where("tenant_id = ?", tenantID).Find(&orgs).Order("title").Error; err != nil {
		return nil, err
	}
	return orgs, nil
}

// ListAllOrganizations lists all organizations.
func (s *S) ListAllOrganizations() ([]*Organization, error) {
	var orgs []*Organization
	if err := s.db.Find(&orgs).Error; err != nil {
		return nil, err
	}
	return orgs, nil
}

// DeleteOrganization deletes an organization.
func (s *S) DeleteOrganization(orgID string) error {
	return DeleteOrganizationInTransaction(s.db, orgID)
}

// DeleteOrganizationInTransaction deletes an organization in a transaction.
func DeleteOrganizationInTransaction(tx *gorm.DB, orgID string) error {
	res := tx.Unscoped().Where("organization_id = ?", orgID).Delete(&Organization{})
	if err := res.Error; err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
