package media

import (
	"fmt"
	"net/http"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
)

type Data []byte

const MaxBytes = 5 * 1024 * 1024

func NewData(value []byte) (Data, error) {
	if len(value) == 0 {
		return nil, fmt.Errorf("%w: media is required", exception.ErrInvalid)
	}

	if len(value) > MaxBytes {
		return nil, fmt.Errorf("%w: media must be at most 5 megabytes", exception.ErrInvalid)
	}

	data := make(Data, len(value))
	copy(data, value)
	return data, nil
}

type ContentType string

const (
	ContentTypeJPEG ContentType = "image/jpeg"
	ContentTypePNG  ContentType = "image/png"
	ContentTypeWebP ContentType = "image/webp"
)

func NewContentType(value string) (ContentType, error) {
	switch contentType := ContentType(value); contentType {
	case ContentTypeJPEG,
		ContentTypePNG,
		ContentTypeWebP:
		return contentType, nil
	default:
		return "", fmt.Errorf("%w: invalid content type", exception.ErrInvalid)
	}
}

type Media struct {
	Data        Data
	ContentType ContentType
}

func NewMedia(data Data) (Media, error) {
	contentType, err := NewContentType(http.DetectContentType(data))
	if err != nil {
		return Media{}, err
	}

	return Media{Data: data, ContentType: contentType}, nil
}
