package mail

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
)

type Address string

var addressRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)+$")

func NewAddress(value string) (Address, error) {
	if value == "" {
		return "", fmt.Errorf("%w: address is required", exception.ErrInvalid)
	}
	normalized := strings.ToLower(value)
	if !addressRegex.MatchString(normalized) {
		return "", fmt.Errorf("%w: invalid address", exception.ErrInvalid)
	}
	return Address(normalized), nil
}
