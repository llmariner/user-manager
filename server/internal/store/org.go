package store

import "gorm.io/gorm"

// Organization is a model for organization
type Organization struct {
	gorm.Model

	TenantID       string `gorm:"uniqueIndex:tenant_id_org_id"`
	OrganizationID string `gorm:"uniqueIndex:tenant_id_org_id"`

	Title string
}
