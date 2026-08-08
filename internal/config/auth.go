package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

func Secret() ([]byte, error) {
	s := os.Getenv("JWT_SECRET")
	if s == "" {
		return nil, fmt.Errorf("environment variable JWT_SECRET is required")
	}
	return []byte(s), nil
}

func AccessTokenTTL() (time.Duration, error) {
	raw := os.Getenv("ACCESS_TOKEN_TTL_HOURS")
	if raw == "" {
		return 24 * time.Hour, nil
	}
	hours, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid ACCESS_TOKEN_TTL_HOURS: %w", err)
	}
	return time.Duration(hours) * time.Hour, nil
}

func RefreshTokenTTL() (time.Duration, error) {
	raw := os.Getenv("REFRESH_TOKEN_TTL_HOURS")
	if raw == "" {
		return 30 * 24 * time.Hour, nil
	}
	hours, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid REFRESH_TOKEN_TTL_HOURS: %w", err)
	}
	return time.Duration(hours) * time.Hour, nil
}
