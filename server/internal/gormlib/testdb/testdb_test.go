package testdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestNew(t *testing.T) {
	db, tearDown := New(t)
	defer tearDown()

	type testRecord struct {
		gorm.Model
		Name string
	}
	err := db.AutoMigrate(&testRecord{})
	assert.NoError(t, err)

	r := &testRecord{Name: "name"}
	result := db.Create(r)
	assert.NoError(t, result.Error)

	var rs []*testRecord
	result = db.Find(&rs)
	assert.NoError(t, result.Error)
}
