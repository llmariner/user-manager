package store

import (
	"gorm.io/gorm"
)

// APIKey represents an API key.
type APIKey struct {
	gorm.Model

	APIKeyID string `gorm:"uniqueIndex:idx_api_key_api_key_id_tenant_id"`

	Name string `gorm:"uniqueIndex:idx_api_key_name_tenant_id"`

	TenantID string `gorm:"uniqueIndex:idx_api_key_api_key_id_tenant_id;uniqueIndex:idx_api_key_name_tenant_id"`

	OrganizationID string
	UserID         string

	Secret string

	// TODO(kenji): Associate roles.
}

// APIKeyKey represents a key of an API key.
type APIKeyKey struct {
	APIKeyID       string
	TenantID       string
	OrganizationID string
	UserID         string
}

// APIKeySpec is a spec of the API key.
type APIKeySpec struct {
	Key APIKeyKey

	Name   string
	Secret string
}

// CreateAPIKey creates a new API key.
func (s *S) CreateAPIKey(spec APIKeySpec) (*APIKey, error) {
	k := &APIKey{
		APIKeyID:       spec.Key.APIKeyID,
		TenantID:       spec.Key.TenantID,
		OrganizationID: spec.Key.OrganizationID,
		UserID:         spec.Key.UserID,

		Name:   spec.Name,
		Secret: spec.Secret,
	}
	if err := s.db.Create(k).Error; err != nil {
		return nil, err
	}
	return k, nil
}

// ListAPIKeysByTenantID lists API keys by a tenant ID.
func (s *S) ListAPIKeysByTenantID(tenantID string) ([]*APIKey, error) {
	var ks []*APIKey
	if err := s.db.Where("tenant_id = ?", tenantID).Find(&ks).Error; err != nil {
		return nil, err
	}
	return ks, nil
}

// GetAPIKeyByNameAndTenantID gets an API key by its name and tenant ID.
func (s *S) GetAPIKeyByNameAndTenantID(name, tenantID string) (*APIKey, error) {
	var k APIKey
	if err := s.db.Where("name = ? AND tenant_id = ?", name, tenantID).Take(&k).Error; err != nil {
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
func (s *S) DeleteAPIKey(k APIKeyKey) error {
	res := s.db.Unscoped().Where("api_key_id = ? AND tenant_id = ?", k.APIKeyID, k.TenantID).Delete(&APIKey{})
	if err := res.Error; err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
