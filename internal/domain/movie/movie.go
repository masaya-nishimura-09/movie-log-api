package movie

type Movie struct {
	ID               ID
	Title            Title
	OriginalTitle    OriginalTitle
	Overview         Overview
	Genres           []Genre
	PosterURL        PosterURL
	ReleaseYear      *ReleaseYear
	Runtime          Runtime
	OriginalLanguage OriginalLanguage
	OriginCountry    []OriginCountry
}

type ID uint
type OriginalTitle string
type Overview string
type PosterURL string
type ReleaseYear uint
type Runtime uint
type OriginalLanguage string
type OriginCountry string
