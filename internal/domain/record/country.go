package record

import (
	"fmt"
	"slices"

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
		return "", fmt.Errorf("%w: malformed country code", exception.ErrInvalid)
	}

	if !region.IsCountry() {
		return "", fmt.Errorf("%w: unknown country code", exception.ErrInvalid)
	}

	return Country(region.String()), nil
}

func NewCountries(values []string) ([]Country, error) {
	countries := make([]Country, 0, len(values))

	for _, value := range values {
		country, err := NewCountry(value)
		if err != nil {
			return nil, err
		}

		if slices.Contains(countries, country) {
			return nil, fmt.Errorf("%w: duplicate country", exception.ErrInvalid)
		}

		countries = append(countries, country)
	}

	return countries, nil
}
