package movie

import (
	"fmt"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	"golang.org/x/text/language"
)

type DisplayLanguage string

func NewDisplayLanguage(value string) (DisplayLanguage, error) {
	if value == "" {
		return "", fmt.Errorf("%w: language is required", exception.ErrInvalid)
	}

	base, err := language.ParseBase(value)
	if err != nil {
		return "", fmt.Errorf("%w: invalid language", exception.ErrInvalid)
	}

	return DisplayLanguage(base.String()), nil
}
