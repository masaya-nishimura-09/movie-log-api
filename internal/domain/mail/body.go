package mail

import (
	"fmt"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
)

type HTMLBody string

func NewHTMLBody(value string) (HTMLBody, error) {
	if value == "" {
		return "", fmt.Errorf("%w: html body is required", exception.ErrInvalid)
	}
	return HTMLBody(value), nil
}

type TextBody string

func NewTextBody(value string) (TextBody, error) {
	if value == "" {
		return "", fmt.Errorf("%w: text body is required", exception.ErrInvalid)
	}
	return TextBody(value), nil
}
