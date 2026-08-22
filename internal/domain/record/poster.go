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

type Poster struct {
	Data        []byte
	ContentType PosterContentType
}

func NewPoster(data []byte) (Poster, error) {
	if len(data) == 0 {
		return Poster{}, fmt.Errorf("%w: poster is required", exception.ErrInvalid)
	}

	if len(data) > 5*1024*1024 {
		return Poster{}, fmt.Errorf("%w: poster must be at most 5 megabytes", exception.ErrInvalid)
	}

	switch contentType := PosterContentType(http.DetectContentType(data)); contentType {
	case PosterContentTypeJPEG,
		PosterContentTypePNG,
		PosterContentTypeWebP:
		return Poster{Data: data, ContentType: contentType}, nil
	default:
		return Poster{}, fmt.Errorf("%w: invalid poster content type", exception.ErrInvalid)
	}
}
