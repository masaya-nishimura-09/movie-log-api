package media

import (
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/masaya-nishimura-09/movie-log-api/internal/testutil"
)

var (
	testS3Client *s3.Client
	testBucket   string
	testBaseURL  string
)

func TestMain(m *testing.M) {
	testS3Client = testutil.NewTestS3Client()
	testBucket = testutil.S3Bucket()
	testBaseURL = testutil.S3BaseURL()
	code := m.Run()
	os.Exit(code)
}
