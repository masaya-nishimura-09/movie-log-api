package record

import (
	"fmt"
	"unicode/utf8"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
)

type ArticleTitle string

func NewArticleTitle(value string) (ArticleTitle, error) {
	if value == "" {
		return "", fmt.Errorf("%w: article title is required", exception.ErrInvalid)
	}

	if utf8.RuneCountInString(value) > 255 {
		return "", fmt.Errorf("%w: article title must be at most 255 characters", exception.ErrInvalid)
	}

	return ArticleTitle(value), nil
}
