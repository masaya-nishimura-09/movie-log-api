package movie

type Movie struct {
	ID               ID
	Title            Title
	OriginalTitle    Title
	Overview         Overview
	Genres           []Genre
	PosterURL        PosterURL
	ReleaseYear      *ReleaseYear
	Runtime          Runtime
	OriginalLanguage OriginalLanguage
	OriginCountry    []Country
}

type ID uint
type Overview string
type PosterURL string
type ReleaseYear uint
type Runtime uint
type OriginalLanguage string
type Country string
