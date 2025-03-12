package store

import (
	v1 "github.com/llmariner/user-manager/api/v1"
	"gorm.io/gorm"
)

// ProjectUser is a model for user_project.
type ProjectUser struct {
	gorm.Model

	ProjectID      string `gorm:"uniqueIndex:user_id_project_id"`
	OrganizationID string
	UserID         string `gorm:"uniqueIndex:user_id_project_id"`

	Role string

	// Hidden is set to true if the user is not visible from the list/get API call.
	Hidden bool
}

// ToProto converts the model to Porto.
func (p *ProjectUser) ToProto() *v1.ProjectUser {
	return &v1.ProjectUser{
		UserId:         p.UserID,
		ProjectId:      p.ProjectID,
		OrganizationId: p.OrganizationID,
		Role:           v1.ProjectRole(v1.ProjectRole_value[p.Role]),
	}
}

// CreateProjectUserParams is the parameters for CreateProjectUser.
type CreateProjectUserParams struct {
	ProjectID      string
	OrganizationID string
	UserID         string
	Role           v1.ProjectRole
}

// CreateProjectUser creates a project user.
func (s *S) CreateProjectUser(p CreateProjectUserParams) (*ProjectUser, error) {
	return CreateProjectUserInTransaction(s.db, p)
}

// CreateProjectUserInTransaction creates a project user in a transaction.
func CreateProjectUserInTransaction(tx *gorm.DB, p CreateProjectUserParams) (*ProjectUser, error) {
	// TODO(aya): rethink user validation: retrieving user information from dex?
	pusr := &ProjectUser{
		ProjectID:      p.ProjectID,
		OrganizationID: p.OrganizationID,
		UserID:         p.UserID,
		Role:           p.Role.String(),
	}
	if err := tx.Create(pusr).Error; err != nil {
		return nil, err
	}
	return pusr, nil
}

// GetProjectUser gets a project user.
func (s *S) GetProjectUser(projectID, userID string) (*ProjectUser, error) {
	var pusr ProjectUser
	if err := s.db.Where("project_id = ? AND user_id = ?", projectID, userID).First(&pusr).Error; err != nil {
		return nil, err
	}
	return &pusr, nil
}

// ListProjectUsersByProjectID lists project users in the specified project.
func (s *S) ListProjectUsersByProjectID(projectID string) ([]ProjectUser, error) {
	var users []ProjectUser
	if err := s.db.Where("project_id = ?", projectID).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// CountProjectUsersByProjectID counts the number of project users in the specified project.
func (s *S) CountProjectUsersByProjectID(projectID string) (int64, error) {
	var numUsers int64
	if err := s.db.Model(&ProjectUser{}).Where("project_id = ?", projectID).Count(&numUsers).Error; err != nil {
		return 0, err
	}
	return numUsers, nil
}

// ListAllProjectUsers lists all project users.
func (s *S) ListAllProjectUsers() ([]ProjectUser, error) {
	var users []ProjectUser
	if err := s.db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// HideProjectUser sets the hide field of the project user to true.
func (s *S) HideProjectUser(projectID, userID string) error {
	result := s.db.Model(&ProjectUser{}).
		Where("project_id = ?", projectID).
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

// DeleteProjectUser deletes a project user.
func (s *S) DeleteProjectUser(projectID, userID string) error {
	return DeleteProjectUserInTransaction(s.db, projectID, userID)
}

// DeleteProjectUserInTransaction deletes a project user in a transaction.
func DeleteProjectUserInTransaction(tx *gorm.DB, projectID, userID string) error {
	res := tx.Unscoped().Where("project_id = ? AND user_id = ?", projectID, userID).Delete(&ProjectUser{})
	if err := res.Error; err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// DeleteAllProjectUsersInTransaction deletes all project users in a transaction.
func DeleteAllProjectUsersInTransaction(tx *gorm.DB, projectID string) error {
	if err := tx.Unscoped().Where("project_id = ?", projectID).Delete(&ProjectUser{}).Error; err != nil {
		return err
	}

	return nil
}
