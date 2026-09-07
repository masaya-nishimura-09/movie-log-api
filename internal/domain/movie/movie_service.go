package movie

import (
	"context"
)

type MovieService interface {
	GetByID(
		ctx context.Context,
		movieID ID,
		displayLanguage DisplayLanguage,
	) (*Movie, error)
	SearchByTitle(
		ctx context.Context,
		title Title,
		page Page,
		displayLanguage DisplayLanguage,
	) (*SearchResult, error)
}
