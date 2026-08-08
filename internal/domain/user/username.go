package user

import (
	"fmt"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
)

type Username string

func NewUsername(value string) (Username, error) {
	if value == "" {
		return "", fmt.Errorf("%w: username is required", exception.ErrInvalid)
	}

	if len(value) > 100 {
		return "", fmt.Errorf("%w: username must be at most 100 characters", exception.ErrInvalid)
	}

	return Username(value), nil
}
