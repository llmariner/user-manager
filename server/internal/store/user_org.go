package store

import "gorm.io/gorm"

// UserOrganization is a model for user_organization.
type UserOrganization struct {
	gorm.Model

	User         uint `gorm:"uniqueIndex:user_org"`
	Organization uint `gorm:"uniqueIndex:user_org"`
}
