package testdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// New creates a new test database and returns a cleanup function.
//
// Please note that a caller needs to import "gorm.io/driver/sqlite"; to avoid
// "gorm.io/driver/sqlite" being in caller's realse, need to manually import it by
// *_test.go files only.
func New(t *testing.T) (*gorm.DB, func()) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	assert.NoError(t, err, "failed to connect database: %s", err)

	deferF := func() {
		sqldb, err := db.DB()
		assert.NoError(t, err, "failed to get sql database to close: %s", err)
		err = sqldb.Close()
		assert.NoError(t, err, "failed to close database: %s", err)
	}

	return db, deferF
}
