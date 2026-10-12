package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

func CleanupEnable() (bool, error) {
	cleanupEnableStr := os.Getenv("CLEANUP_ENABLE")
	cleanupEnableBool, err := strconv.ParseBool(cleanupEnableStr)
	if err != nil {
		return false, fmt.Errorf("invalid CLEANUP_ENABLE: %w", err)
	}
	return cleanupEnableBool, nil
}

func CleanupInterval() (time.Duration, error) {
	cleanupIntervalStr := os.Getenv("CLEANUP_INTERVAL_SECOND")
	if cleanupIntervalStr == "" {
		return 0, fmt.Errorf("environment variable CLEANUP_INTERVAL_SECOND is required")
	}

	cleanupIntervalInt, err := strconv.Atoi(cleanupIntervalStr)
	if err != nil {
		return 0, fmt.Errorf("invalid CLEANUP_INTERVAL_SECOND: %w", err)
	}
	if cleanupIntervalInt < 1 {
		return 0, fmt.Errorf("environment variable CLEANUP_INTERVAL_SECOND must be over 0")
	}

	cleanupInterval := time.Duration(cleanupIntervalInt) * time.Second
	return cleanupInterval, nil
}
