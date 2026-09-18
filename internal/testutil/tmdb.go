package testutil

import (
	"log"
	"net/url"

	"github.com/masaya-nishimura-09/movie-log-api/internal/config"
)

func TMDBPosterBaseURL() *url.URL {
	loadTestEnv()

	tmdbPosterBaseURL, err := config.TMDBPosterBaseURL()
	if err != nil {
		log.Fatalf("%v", err)
	}

	return tmdbPosterBaseURL
}
