package mail

import (
	"fmt"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
)

type Language string

const (
	LanguageJA Language = "ja"
	LanguageEN Language = "en"
)

func NewLanguage(value string) (Language, error) {
	switch language := Language(value); language {
	case LanguageJA,
		LanguageEN:
		return language, nil
	case "":
		return "", fmt.Errorf("%w: language is required", exception.ErrInvalid)
	default:
		return "", fmt.Errorf("%w: invalid language", exception.ErrInvalid)
	}
}
