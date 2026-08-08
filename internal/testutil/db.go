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

func TestDB(t *testing.T) *gorm.DB {
	t.Helper()

	loadEnvOnce.Do(func() {
		_ = godotenv.Load("../../../.env")
	})

	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("TEST_POSTGRES_DSN not set (add it to .env or export it); skipping test that requires a database")
	}

	db, err := gorm.Open(postgres.New(postgres.Config{DSN: dsn}), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	return db
}
