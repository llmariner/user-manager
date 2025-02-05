package store

import (
	"gorm.io/gorm"
)

// APIKey represents an API key.
type APIKey struct {
	gorm.Model

	APIKeyID string `gorm:"uniqueIndex"`

	Name string `gorm:"uniqueIndex:idx_api_key_name_user_id"`

	TenantID string

	OrganizationID string
	ProjectID      string
	UserID         string `gorm:"uniqueIndex:idx_api_key_name_user_id"`

	IsServiceAccount bool

	// Secret is set when kms encryption is disabled.
	Secret string
	// EncryptedSecret is encrypted by data key, and it is set when kms encryption is enabled.
	EncryptedSecret []byte

	// TODO(kenji): Associate roles.
}

// APIKeySpec is a spec of the API key.
type APIKeySpec struct {
	APIKeyID       string
	TenantID       string
	OrganizationID string
	ProjectID      string
	UserID         string

	IsServiceAccount bool

	Name   string
	Secret string
	// EncryptedSecret is encrypted by data key.
	EncryptedSecret []byte
}

// CreateAPIKey creates a new API key.
func (s *S) CreateAPIKey(spec APIKeySpec) (*APIKey, error) {
	return CreateAPIKeyInTransaction(s.db, spec)
}

// CreateAPIKeyInTransaction creates a new API key in a transaction.
func CreateAPIKeyInTransaction(db *gorm.DB, spec APIKeySpec) (*APIKey, error) {
	k := &APIKey{
		APIKeyID:         spec.APIKeyID,
		TenantID:         spec.TenantID,
		OrganizationID:   spec.OrganizationID,
		ProjectID:        spec.ProjectID,
		UserID:           spec.UserID,
		IsServiceAccount: spec.IsServiceAccount,

		Name:            spec.Name,
		Secret:          spec.Secret,
		EncryptedSecret: spec.EncryptedSecret,
	}
	if err := db.Create(k).Error; err != nil {
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

// ListAPIKeysByTenantID lists API keys by a tenant ID.
func (s *S) ListAPIKeysByTenantID(tenantID string) ([]*APIKey, error) {
	var ks []*APIKey
	if err := s.db.Where("tenant_id = ?", tenantID).Find(&ks).Error; err != nil {
		return nil, err
	}
	return ks, nil
}

// GetAPIKeyByNameAndUserID gets an API key by its name and user ID.
func (s *S) GetAPIKeyByNameAndUserID(name, userID string) (*APIKey, error) {
	var k APIKey
	if err := s.db.Where("name = ? AND user_id = ?", name, userID).Take(&k).Error; err != nil {
		return nil, err
	}
	return &k, nil
}

// GetAPIKey gets an API key by its ID and tenant ID.
func (s *S) GetAPIKey(id, projectID string) (*APIKey, error) {
	var k APIKey
	if err := s.db.Where("api_key_id = ? AND project_id = ?", id, projectID).Take(&k).Error; err != nil {
		return nil, err
	}
	return &k, nil
}

// GetAPIKeyByID gets an API key by its ID.
func (s *S) GetAPIKeyByID(id string) (*APIKey, error) {
	var k APIKey
	if err := s.db.Where("api_key_id = ?", id).Take(&k).Error; err != nil {
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
	return DeleteAPIKeyInTransaction(s.db, apiKeyID, projectID)
}

// DeleteAPIKeyInTransaction deletes an APIKey in a transaction.
func DeleteAPIKeyInTransaction(tx *gorm.DB, apiKeyID, projectID string) error {
	res := tx.Unscoped().Where("api_key_id = ? AND project_id = ?", apiKeyID, projectID).Delete(&APIKey{})
	if err := res.Error; err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// UpdateAPIKey updates an API key.
func (s *S) UpdateAPIKey(key *APIKey) error {
	if err := s.db.Save(key).Error; err != nil {
		return err
	}
	return nil
}
