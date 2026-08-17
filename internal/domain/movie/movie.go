package movie

import (
	"time"
)

type Movie struct {
	ID          ID
	TMDBID      MovieTMDBID
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

func NewMovie(
	tmdbID MovieTMDBID,
	title Title,
	releaseYear ReleaseYear,
	runtime Runtime,
	countries []Country,
	language Language,
	posterURL PosterURL,
	genres []Genre,
	credits []Credit,
) Movie {
	return Movie{
		TMDBID:      tmdbID,
		Title:       title,
		ReleaseYear: releaseYear,
		Runtime:     runtime,
		Countries:   countries,
		Language:    language,
		PosterURL:   posterURL,
		Genres:      genres,
		Credits:     credits,
	}
}
