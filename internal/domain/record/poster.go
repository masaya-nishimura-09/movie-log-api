package record

import (
	"fmt"
	"net/http"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
)

type PosterData []byte

func NewPosterData(value []byte) (PosterData, error) {
	if len(value) == 0 {
		return nil, fmt.Errorf("%w: poster is required", exception.ErrInvalid)
	}

	if len(value) > 5*1024*1024 {
		return nil, fmt.Errorf("%w: poster must be at most 5 megabytes", exception.ErrInvalid)
	}

	data := make(PosterData, len(value))
	copy(data, value)
	return data, nil
}

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
	Data        PosterData
	ContentType PosterContentType
}

func NewPoster(data PosterData) (Poster, error) {
	contentType, err := NewPosterContentType(http.DetectContentType(data))
	if err != nil {
		return Poster{}, err
	}

	return Poster{Data: data, ContentType: contentType}, nil
}
