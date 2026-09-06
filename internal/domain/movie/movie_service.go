package movie

import (
	"context"
)

type MovieService interface {
	GetByID(
		ctx context.Context,
		movieID ID,
		language DisplayLanguage,
	) (*Movie, error)
	SearchByTitle(
		ctx context.Context,
		title Title,
		page Page,
		language DisplayLanguage,
	) (*SearchResult, error)
}
