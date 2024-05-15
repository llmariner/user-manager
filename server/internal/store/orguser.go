package store

import (
	"errors"

	v1 "github.com/llm-operator/user-manager/api/v1"
	"gorm.io/gorm"
)

// OrganizationUser is a model for user_organization.
type OrganizationUser struct {
	gorm.Model

	UserID         string `gorm:"uniqueIndex:user_id_org_id"`
	OrganizationID string `gorm:"uniqueIndex:user_id_org_id"`

	Role string
}

// ToProto converts the model to proto.
func (o *OrganizationUser) ToProto() *v1.OrganizationUser {
	return &v1.OrganizationUser{
		UserId:         o.UserID,
		OrganizationId: o.OrganizationID,
		Role:           v1.Role(v1.Role_value[o.Role]),
	}
}

// CreateOrganizationUser creates a organization user.
func (s *S) CreateOrganizationUser(tenantID, orgID, userID, role string) (*OrganizationUser, error) {
	org, err := s.GetOrganization(orgID)
	if err != nil {
		return nil, err
	}
	if org.TenantID != tenantID {
		return nil, errors.New("organization not found")
	}

	// TODO(aya): Revisit user creation:
	// create users when proxying dex create-password API
	// or when retrieving user information from dex.
	var user User
	if err := s.db.FirstOrCreate(&user, User{
		TenantID: org.TenantID,
		UserID:   userID,
	}).Error; err != nil {
		return nil, err
	}

	userOrg := &OrganizationUser{
		UserID:         userID,
		OrganizationID: orgID,
		Role:           role,
	}
	if err := s.db.Create(userOrg).Error; err != nil {
		return nil, err
	}
	return userOrg, nil
}

// ListAllOrganizationUsers lists all organization users.
func (s *S) ListAllOrganizationUsers() ([]OrganizationUser, error) {
	var users []OrganizationUser
	if err := s.db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}
