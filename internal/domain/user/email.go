package user

import (
	"fmt"
	"regexp"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
)

type Email string

var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)+$")

func NewEmail(value string) (Email, error) {
	if value == "" {
		return "", fmt.Errorf("%w: email is required", exception.ErrInvalid)
	}
	if !emailRegex.MatchString(string(value)) {
		return "", fmt.Errorf("%w: invalid email", exception.ErrInvalid)
	}
	return Email(value), nil
}
