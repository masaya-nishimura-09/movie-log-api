package record

import (
	"fmt"
	"net/url"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
)

type PosterURL string

func NewPosterURL(value string) (PosterURL, error) {
	if value == "" {
		return "", nil
	}

	u, err := url.Parse(value)
	if err != nil {
		return "", fmt.Errorf("%w: invalid poster url", exception.ErrInvalid)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("%w: poster url must be http or https", exception.ErrInvalid)
	}

	if u.Host == "" {
		return "", fmt.Errorf("%w: poster url must have a host", exception.ErrInvalid)
	}

	return PosterURL(value), nil
}
