package store

import "gorm.io/gorm"

// Organization is a model for organization
type Organization struct {
	gorm.Model

	TenantID       string `gorm:"index"`
	OrganizationID string `gorm:"uniqueIndex"`

	Title string
}
