package mail

import (
	"fmt"
	"net/url"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
)

type URL string

func NewURL(value string) (URL, error) {
	if value == "" {
		return "", fmt.Errorf("%w: url is required", exception.ErrInvalid)
	}

	u, err := url.Parse(value)
	if err != nil {
		return "", fmt.Errorf("%w: invalid url", exception.ErrInvalid)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("%w: url must be http or https", exception.ErrInvalid)
	}

	if u.Host == "" {
		return "", fmt.Errorf("%w: url must have a host", exception.ErrInvalid)
	}

	return URL(value), nil
}
