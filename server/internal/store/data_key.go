package store

import (
	"context"

	"gorm.io/gorm"
)

// DataKey represents an data key, which is generated from AWS KMS master key.
type DataKey struct {
	gorm.Model

	EncryptedDataKey []byte
}

// DataKeyManagementClient contains the methods to manage data keys for users.
type DataKeyManagementClient interface {
	CreateDataKey(ctx context.Context) ([]byte, []byte, error)
	DecryptDataKey(ctx context.Context, encryptedKey []byte) ([]byte, error)
}

// CreateDataKey creates a data key.
func (s *S) CreateDataKey(ctx context.Context, kmsClient DataKeyManagementClient) ([]byte, error) {
	dataKey, encryptedDataKey, err := kmsClient.CreateDataKey(ctx)
	if err != nil {
		return nil, err
	}

	k := &DataKey{
		EncryptedDataKey: encryptedDataKey,
	}
	if err := s.db.Create(k).Error; err != nil {
		return nil, err
	}
	return dataKey, nil
}

// GetDataKey gets the data key.
func (s *S) GetDataKey(ctx context.Context, kmsClient DataKeyManagementClient) ([]byte, error) {
	var k DataKey
	if err := s.db.Take(&k).Error; err != nil {
		return nil, err
	}
	dataKey, err := kmsClient.DecryptDataKey(ctx, k.EncryptedDataKey)
	if err != nil {
		return nil, err
	}
	return dataKey, nil
}
