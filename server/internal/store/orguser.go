package store

import (
	v1 "github.com/llmariner/user-manager/api/v1"
	"gorm.io/gorm"
)

// OrganizationUser is a model for user_organization.
type OrganizationUser struct {
	gorm.Model

	OrganizationID string `gorm:"uniqueIndex:user_id_org_id"`
	UserID         string `gorm:"uniqueIndex:user_id_org_id"`

	Role string

	// Hidden is set to true if the user is not visible from the list/get API call.
	Hidden bool
}

// ToProto converts the model to Porto.
func (o *OrganizationUser) ToProto(internalUserID string) *v1.OrganizationUser {
	return &v1.OrganizationUser{
		OrganizationId: o.OrganizationID,
		UserId:         o.UserID,
		InternalUserId: internalUserID,
		Role:           v1.OrganizationRole(v1.OrganizationRole_value[o.Role]),
	}
}

// CreateOrganizationUser creates a organization user.
func (s *S) CreateOrganizationUser(orgID, userID, role string) (*OrganizationUser, error) {
	return CreateOrganizationUserInTransaction(s.db, orgID, userID, role)
}

// CreateOrganizationUserInTransaction creates a organization user in a transaction.
func CreateOrganizationUserInTransaction(tx *gorm.DB, orgID, userID, role string) (*OrganizationUser, error) {
	// TODO(aya): rethink user validation: retrieving user information from dex?
	orgusr := &OrganizationUser{
		OrganizationID: orgID,
		UserID:         userID,
		Role:           role,
	}
	if err := tx.Create(orgusr).Error; err != nil {
		return nil, err
	}
	return orgusr, nil
}

// GetOrganizationUser gets a organization user.
func (s *S) GetOrganizationUser(orgID, userID string) (*OrganizationUser, error) {
	var orgusr OrganizationUser
	if err := s.db.Where("organization_id = ? AND user_id = ?", orgID, userID).First(&orgusr).Error; err != nil {
		return nil, err
	}
	return &orgusr, nil
}

// ListOrganizationUsersByOrganizationID lists organization users in the specified organization.
func (s *S) ListOrganizationUsersByOrganizationID(orgID string) ([]OrganizationUser, error) {
	var users []OrganizationUser
	if err := s.db.Where("organization_id = ?", orgID).Order("user_id").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// ListOrganizationNonHiddenUsersByOrganizationID lists organization users in the specified organization.
func (s *S) ListOrganizationNonHiddenUsersByOrganizationID(orgID string) ([]OrganizationUser, error) {
	var users []OrganizationUser
	if err := s.db.Where("organization_id = ?", orgID).Where("hidden = ?", false).Order("user_id").Find(&users).Error; err != nil {
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

// HideOrganizationUser sets the hide field of the organization user to true.
func (s *S) HideOrganizationUser(orgID, userID string) error {
	result := s.db.Model(&OrganizationUser{}).
		Where("organization_id = ?", orgID).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"hidden": true,
		})
	if err := result.Error; err != nil {
		return err
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// DeleteOrganizationUser deletes a organization user.
func (s *S) DeleteOrganizationUser(orgID, userID string) error {
	return DeleteOrganizationUserInTransaction(s.db, orgID, userID)
}

// DeleteOrganizationUserInTransaction deletes a organization user in a transaction.
func DeleteOrganizationUserInTransaction(tx *gorm.DB, orgID, userID string) error {
	res := tx.Unscoped().Where("organization_id = ? AND user_id = ?", orgID, userID).Delete(&OrganizationUser{})
	if err := res.Error; err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// DeleteAllOrganizationUsersInTransaction deletes all organization users in the specified organization.
func DeleteAllOrganizationUsersInTransaction(tx *gorm.DB, orgID string) error {
	if err := tx.Unscoped().Where("organization_id = ?", orgID).Delete(&OrganizationUser{}).Error; err != nil {
		return err
	}
	return nil
}
