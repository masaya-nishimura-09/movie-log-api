package record

import (
	"fmt"
	"unicode/utf8"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
)

type Overview string

func NewOverview(value string) (Overview, error) {
	if utf8.RuneCountInString(value) > 2000 {
		return "", fmt.Errorf("%w: overview must be at most 2000 characters", exception.ErrInvalid)
	}

	return Overview(value), nil
}
