package store

import "gorm.io/gorm"

// User is a model for user
type User struct {
	gorm.Model

	TenantID string `gorm:"uniqueIndex:tenant_id_user_id"`
	UserID   string `gorm:"uniqueIndex:tenant_id_user_id"`
}
