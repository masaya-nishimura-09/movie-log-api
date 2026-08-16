package movie

import (
	"fmt"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	"golang.org/x/text/language"
)

type Country string

func NewCountry(value string) (Country, error) {
	if value == "" {
		return "", fmt.Errorf("%w: country is required", exception.ErrInvalid)
	}

	region, err := language.ParseRegion(value)
	if err != nil {
		return "", fmt.Errorf("%w: invalid country", exception.ErrInvalid)
	}

	if !region.IsCountry() {
		return "", fmt.Errorf("%w: invalid country", exception.ErrInvalid)
	}

	return Country(region.String()), nil
}
