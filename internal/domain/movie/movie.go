package movie

type Movie struct {
	ID              ID
	Title           Title
	OriginalTitle   Title
	Overview        Overview
	Genres          []Genre
	PosterURL       PosterURL
	ReleaseYear     ReleaseYear
	Runtime         Runtime
	Language        Language
	OriginCountries []Country
}

type ID uint
type Overview string
type Genre string
type PosterURL string
type ReleaseYear uint
type Runtime uint
type Language string
type DisplayLanguage string
type Country string
