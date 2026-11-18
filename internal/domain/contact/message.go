package contact

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
)

type Message string

func NewMessage(value string) (Message, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", fmt.Errorf("%w: message is required", exception.ErrInvalid)
	}

	if utf8.RuneCountInString(trimmed) > 2000 {
		return "", fmt.Errorf("%w: message must be at most 2000 characters", exception.ErrInvalid)
	}

	return Message(trimmed), nil
}
