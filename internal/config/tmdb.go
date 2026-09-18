package config

import (
	"fmt"
	"net/url"
	"os"
)

func TMDBAccessToken() (string, error) {
	token := os.Getenv("TMDB_ACCESS_TOKEN")
	if token == "" {
		return "", fmt.Errorf("environment variable TMDB_ACCESS_TOKEN is required")
	}
	return token, nil
}

func TMDBEndpoint() (*url.URL, error) {
	endpoint := os.Getenv("TMDB_ENDPOINT")
	if endpoint == "" {
		return nil, fmt.Errorf("environment variable TMDB_ENDPOINT is required")
	}

	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid TMDB_ENDPOINT: %w", err)
	}
	if (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return nil, fmt.Errorf("environment variable TMDB_ENDPOINT must be an http or https url")
	}

	return u, nil
}

func TMDBPosterBaseURL() (*url.URL, error) {
	baseURL := os.Getenv("TMDB_POSTER_BASE_URL")
	if baseURL == "" {
		return nil, fmt.Errorf("environment variable TMDB_POSTER_BASE_URL is required")
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid TMDB_POSTER_BASE_URL: %w", err)
	}
	if (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return nil, fmt.Errorf("environment variable TMDB_POSTER_BASE_URL must be an http or https url")
	}

	return u, nil
}
