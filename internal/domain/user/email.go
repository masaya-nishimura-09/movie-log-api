package user

import (
	"fmt"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/mail"
)

type Email string

func NewEmail(value string) (Email, error) {
	if value == "" {
		return "", fmt.Errorf("%w: email is required", exception.ErrInvalid)
	}
	address, err := mail.NewAddress(value)
	if err != nil {
		return "", fmt.Errorf("%w: invalid email", exception.ErrInvalid)
	}
	return Email(address), nil
}
