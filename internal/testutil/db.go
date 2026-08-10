package testutil

import (
	"log"
	"path/filepath"
	"runtime"

	"github.com/joho/godotenv"
	"github.com/masaya-nishimura-09/movie-log-api/internal/config"
	"gorm.io/gorm"
)

func NewTestDB() *gorm.DB {
	_, file, _, _ := runtime.Caller(0)
	repoRoot := filepath.Join(filepath.Dir(file), "..", "..")
	envPath := filepath.Join(repoRoot, ".env.test")

	if err := godotenv.Load(envPath); err != nil {
		log.Println(".env.test file not found, using environment variables")
	}

	db, err := config.NewDB()
	if err != nil {
		log.Fatalf("%v", err)
	}

	return db
}
