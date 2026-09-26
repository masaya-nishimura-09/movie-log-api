package config

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
)

func GinMode() (string, error) {
	mode := os.Getenv("GIN_MODE")
	switch mode {
	case "", gin.DebugMode, gin.ReleaseMode, gin.TestMode:
		return mode, nil
	default:
		return "", fmt.Errorf("invalid GIN_MODE: %s", mode)
	}
}
