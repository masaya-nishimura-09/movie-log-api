package testutil

import (
	"log"
	"testing"

	"github.com/masaya-nishimura-09/movie-log-api/internal/config"
	"gorm.io/gorm"
)

func NewTestDB() *gorm.DB {
	loadTestEnv()

	db, err := config.NewDB()
	if err != nil {
		log.Fatalf("%v", err)
	}

	return db
}

func BeginTx(t *testing.T, db *gorm.DB) *gorm.DB {
	t.Helper()
	tx := db.Begin()
	if tx.Error != nil {
		t.Fatalf("Begin() error = %v", tx.Error)
	}
	t.Cleanup(func() { tx.Rollback() })
	return tx
}
