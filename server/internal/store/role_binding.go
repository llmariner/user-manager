package store

import "gorm.io/gorm"

// SubjectType is a type for subject.
type SubjectType int32

const (
	// SubjectTypeUser is subject type of user.
	SubjectTypeUser SubjectType = iota
	// SubjectTypeOrganization is subject type of organization.
	SubjectTypeOrganization
	// SubjectTypeAPIKey is subject type of API key.
	SubjectTypeAPIKey
)

// RoleBinding is a model for role_binding
type RoleBinding struct {
	gorm.Model

	SubjectID   uint        `gorm:"uniqueIndex:subject_id_type"`
	SubjectType SubjectType `gorm:"uniqueIndex:subject_id_type"`

	Role string
}
