package record

import (
	"fmt"
	"unicode/utf8"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
)

type OriginalTitle string

func NewOriginalTitle(value string) (OriginalTitle, error) {
	if utf8.RuneCountInString(value) > 255 {
		return "", fmt.Errorf("%w: original title must be at most 255 characters", exception.ErrInvalid)
	}

	return OriginalTitle(value), nil
}
