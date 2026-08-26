package testutil

import (
	"log"
	"path/filepath"
	"runtime"

	"github.com/joho/godotenv"
)

func loadTestEnv() {
	_, file, _, _ := runtime.Caller(0)
	repoRoot := filepath.Join(filepath.Dir(file), "..", "..")
	envPath := filepath.Join(repoRoot, ".env.test")

	if err := godotenv.Load(envPath); err != nil {
		log.Println(".env.test file not found, using environment variables")
	}
}
