package store

import (
	"fmt"

	"gorm.io/gorm"
)

// User provides a mapping between the external user ID (= email address)
// and the internal user ID.
//
// Internal UserIDs can be used to uniquely idently users without
// exposing PII.
type User struct {
	gorm.Model

	UserID         string `gorm:"uniqueIndex"`
	InternalUserID string `gorm:"uniqueIndex"`
}

// FindOrCreateUserInTransaction creates a new user.
func FindOrCreateUserInTransaction(tx *gorm.DB, userID, internalUserID string) (*User, error) {
	var us []*User
	if err := tx.Where("user_id = ?", userID).Find(&us).Error; err != nil {
		return nil, err
	}
	if len(us) > 1 {
		return nil, fmt.Errorf("unexpected number of users found: %v", us)
	}
	if len(us) == 1 {
		return us[0], nil
	}

	u := &User{
		UserID:         userID,
		InternalUserID: internalUserID,
	}
	if err := tx.Create(u).Error; err != nil {
		return nil, err
	}
	return u, nil
}
