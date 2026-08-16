package movie

import (
	"time"
)

type Movie struct {
	ID          ID
	TMDBID      *MovieTMDBID
	Title       Title
	ReleaseYear ReleaseYear
	Runtime     Runtime
	Countries   []Country
	Language    Language
	PosterURL   PosterURL
	Genres      []Genre
	Credits     []Credit
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ID uint
type MovieTMDBID uint
