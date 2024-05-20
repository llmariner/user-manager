package store

import (
	v1 "github.com/llm-operator/user-manager/api/v1"
	"gorm.io/gorm"
)

// OrganizationUser is a model for user_organization.
type OrganizationUser struct {
	gorm.Model

	OrganizationID string `gorm:"uniqueIndex:user_id_org_id"`
	UserID         string `gorm:"uniqueIndex:user_id_org_id"`

	Role string
}

// ToProto converts the model to Porto.
func (o *OrganizationUser) ToProto() *v1.OrganizationUser {
	return &v1.OrganizationUser{
		OrganizationId: o.OrganizationID,
		UserId:         o.UserID,
		Role:           v1.Role(v1.Role_value[o.Role]),
	}
}

// CreateOrganizationUser creates a organization user.
func (s *S) CreateOrganizationUser(orgID, userID, role string) (*OrganizationUser, error) {
	// TODO(aya): rethink user validation: retrieving user information from dex?
	orgusr := &OrganizationUser{
		OrganizationID: orgID,
		UserID:         userID,
		Role:           role,
	}
	if err := s.db.Create(orgusr).Error; err != nil {
		return nil, err
	}
	return orgusr, nil
}

// ListOrganizationUsersByOrganizationID lists organization users in the specified organization.
func (s *S) ListOrganizationUsersByOrganizationID(orgID string) ([]OrganizationUser, error) {
	var users []OrganizationUser
	if err := s.db.Where("organization_id = ?", orgID).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// ListAllOrganizationUsers lists all organization users.
func (s *S) ListAllOrganizationUsers() ([]OrganizationUser, error) {
	var users []OrganizationUser
	if err := s.db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}
