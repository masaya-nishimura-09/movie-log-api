package auth

import (
	"os"
	"testing"

	"github.com/masaya-nishimura-09/movie-log-api/internal/testutil"
	"gorm.io/gorm"
)

var testDB *gorm.DB

func TestMain(m *testing.M) {
	testDB = testutil.NewTestDB()
	code := m.Run()
	os.Exit(code)
}
