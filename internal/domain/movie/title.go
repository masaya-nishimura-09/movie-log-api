package movie

import (
	"fmt"
	"unicode/utf8"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
)

type Title string

func NewTitle(value string) (Title, error) {
	if value == "" {
		return "", fmt.Errorf("%w: title is required", exception.ErrInvalid)
	}

	if utf8.RuneCountInString(value) > 255 {
		return "", fmt.Errorf("%w: title must be at most 255 characters", exception.ErrInvalid)
	}

	return Title(value), nil
}
