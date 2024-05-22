package store

import (
	"gorm.io/gorm"
)

// APIKey represents an API key.
type APIKey struct {
	gorm.Model

	APIKeyID string `gorm:"uniqueIndex"`

	Name string `gorm:"uniqueIndex:idx_api_key_name_tenant_id"`

	TenantID string `gorm:"uniqueIndex:idx_api_key_name_tenant_id"`

	OrganizationID string
	ProjectID      string
	UserID         string

	Secret string

	// TODO(kenji): Associate roles.
}

// APIKeySpec is a spec of the API key.
type APIKeySpec struct {
	APIKeyID       string
	TenantID       string
	OrganizationID string
	ProjectID      string
	UserID         string

	Name   string
	Secret string
}

// CreateAPIKey creates a new API key.
func (s *S) CreateAPIKey(spec APIKeySpec) (*APIKey, error) {
	k := &APIKey{
		APIKeyID:       spec.APIKeyID,
		TenantID:       spec.TenantID,
		OrganizationID: spec.OrganizationID,
		ProjectID:      spec.ProjectID,
		UserID:         spec.UserID,

		Name:   spec.Name,
		Secret: spec.Secret,
	}
	if err := s.db.Create(k).Error; err != nil {
		return nil, err
	}
	return k, nil
}

// ListAPIKeysByProjectID lists API keys by a tenant ID.
func (s *S) ListAPIKeysByProjectID(projectID string) ([]*APIKey, error) {
	var ks []*APIKey
	if err := s.db.Where("project_id = ?", projectID).Find(&ks).Error; err != nil {
		return nil, err
	}
	return ks, nil
}

// GetAPIKeyByNameAndProjectID gets an API key by its name and tenant ID.
func (s *S) GetAPIKeyByNameAndProjectID(name, projectID string) (*APIKey, error) {
	var k APIKey
	if err := s.db.Where("name = ? AND project_id = ?", name, projectID).Take(&k).Error; err != nil {
		return nil, err
	}
	return &k, nil
}

// ListAllAPIKeys lists all API keys.
func (s *S) ListAllAPIKeys() ([]*APIKey, error) {
	var ks []*APIKey
	if err := s.db.Find(&ks).Error; err != nil {
		return nil, err
	}
	return ks, nil
}

// DeleteAPIKey deletes an APIKey by APIKey ID and tenant ID.
func (s *S) DeleteAPIKey(apiKeyID, projectID string) error {
	res := s.db.Unscoped().Where("api_key_id = ? AND project_id = ?", apiKeyID, projectID).Delete(&APIKey{})
	if err := res.Error; err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
