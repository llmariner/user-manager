package store

import "gorm.io/gorm"

// UserOrganization is a model for user_organization.
type UserOrganization struct {
	gorm.Model

	User         uint `gorm:"uniqueIndex:user_org"`
	Organization uint `gorm:"uniqueIndex:user_org"`
}

// CreateUserOrganization adds a user to an organization.
func (s *S) CreateUserOrganization(orgID, userID string) (*UserOrganization, error) {
	org, err := s.GetOrganization(orgID)
	if err != nil {
		return nil, err
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

	userOrg := &UserOrganization{
		User:         user.ID,
		Organization: org.ID,
	}
	if err := s.db.Create(userOrg).Error; err != nil {
		return nil, err
	}
	return userOrg, nil
}
