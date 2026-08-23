package record

import (
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/masaya-nishimura-09/movie-log-api/internal/testutil"
	"gorm.io/gorm"
)

var (
	testDB       *gorm.DB
	testS3Client *s3.Client
	testBucket   string
	testBaseURL  string
)

func TestMain(m *testing.M) {
	testDB = testutil.NewTestDB()
	testS3Client = testutil.NewTestS3Client()
	testBucket = testutil.NewTestS3Bucket()
	testBaseURL = testutil.NewTestS3BaseURL()
	code := m.Run()
	os.Exit(code)
}
