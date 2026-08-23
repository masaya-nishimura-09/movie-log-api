package record

import (
	"fmt"
	"net/http"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
)

type PosterContentType string

const (
	PosterContentTypeJPEG PosterContentType = "image/jpeg"
	PosterContentTypePNG  PosterContentType = "image/png"
	PosterContentTypeWebP PosterContentType = "image/webp"
)

func NewPosterContentType(value string) (PosterContentType, error) {
	switch contentType := PosterContentType(value); contentType {
	case PosterContentTypeJPEG,
		PosterContentTypePNG,
		PosterContentTypeWebP:
		return contentType, nil
	default:
		return "", fmt.Errorf("%w: invalid poster content type", exception.ErrInvalid)
	}
}

type Poster struct {
	Data        []byte
	ContentType PosterContentType
}

func NewPoster(value []byte) (Poster, error) {
	if len(value) == 0 {
		return Poster{}, fmt.Errorf("%w: poster is required", exception.ErrInvalid)
	}

	if len(value) > 5*1024*1024 {
		return Poster{}, fmt.Errorf("%w: poster must be at most 5 megabytes", exception.ErrInvalid)
	}

	contentType, err := NewPosterContentType(http.DetectContentType(value))
	if err != nil {
		return Poster{}, err
	}

	buf := make([]byte, len(value))
	copy(buf, value)
	return Poster{Data: buf, ContentType: contentType}, nil
}
