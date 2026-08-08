// Package testutil provides shared helpers for tests that need a real
// database connection. It is only ever imported from _test.go files.
package testutil

import (
	"os"
	"sync"
	"testing"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var loadEnvOnce sync.Once

// TestDB connects to the dedicated test database identified by the
// TEST_POSTGRES_DSN environment variable. If that variable isn't set
// (in .env or otherwise), the calling test is skipped rather than failed.
func TestDB(t *testing.T) *gorm.DB {
	t.Helper()

	// internal/infrastructure/<context> is 3 directories below the repo
	// root, where .env lives.
	loadEnvOnce.Do(func() {
		_ = godotenv.Load("../../../.env")
	})

	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("TEST_POSTGRES_DSN not set (add it to .env or export it); skipping test that requires a database")
	}

	db, err := gorm.Open(postgres.New(postgres.Config{DSN: dsn}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	return db
}
