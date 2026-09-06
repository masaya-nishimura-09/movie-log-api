package movie

import (
	"net/url"
	"os"
	"testing"

	"github.com/masaya-nishimura-09/movie-log-api/internal/testutil"
)

var (
	posterBaseURL *url.URL
)

func TestMain(m *testing.M) {
	posterBaseURL = testutil.TMDBPosterBaseURL()
	code := m.Run()
	os.Exit(code)
}
