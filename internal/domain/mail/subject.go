package mail

import (
	"fmt"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
)

type Subject string

func NewSubject(value string) (Subject, error) {
	if value == "" {
		return "", fmt.Errorf("%w: subject is required", exception.ErrInvalid)
	}
	return Subject(value), nil
}
