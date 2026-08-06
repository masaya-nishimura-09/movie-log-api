package user

import (
	"fmt"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
)

type Password string

func NewPassword(value string) (Password, error) {
	if value == "" {
		return "", fmt.Errorf("%w: password is required", exception.ErrValidation)
	}

	if len(value) < 8 {
		return "", fmt.Errorf("%w: password must be at least 8 characters", exception.ErrValidation)
	}

	if len(value) > 72 {
		return "", fmt.Errorf("%w: password must be at most 72 characters", exception.ErrValidation)
	}

	return Password(value), nil
}
