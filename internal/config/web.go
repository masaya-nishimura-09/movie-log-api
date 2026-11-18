package config

import (
	"fmt"
	"net/url"
	"os"
)

func WebBaseURL() (*url.URL, error) {
	baseURL := os.Getenv("WEB_BASE_URL")
	if baseURL == "" {
		return nil, fmt.Errorf("environment variable WEB_BASE_URL is required")
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid WEB_BASE_URL: %w", err)
	}
	if (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return nil, fmt.Errorf("environment variable WEB_BASE_URL must be an http or https url")
	}

	return u, nil
}
