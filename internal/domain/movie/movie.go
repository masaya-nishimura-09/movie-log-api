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
	OriginCountries  []OriginCountry
	Casts            []Cast
}

type ID uint
type OriginalTitle string
type Overview string
type PosterURL string
type ReleaseYear uint
type Runtime uint
type OriginalLanguage string
type OriginCountry string
type Genre string

const (
	GenreAction         Genre = "action"
	GenreAdventure      Genre = "adventure"
	GenreAnimation      Genre = "animation"
	GenreComedy         Genre = "comedy"
	GenreCrime          Genre = "crime"
	GenreDocumentary    Genre = "documentary"
	GenreDrama          Genre = "drama"
	GenreFamily         Genre = "family"
	GenreFantasy        Genre = "fantasy"
	GenreHistory        Genre = "history"
	GenreHorror         Genre = "horror"
	GenreMusic          Genre = "music"
	GenreMystery        Genre = "mystery"
	GenreRomance        Genre = "romance"
	GenreScienceFiction Genre = "science_fiction"
	GenreTVMovie        Genre = "tv_movie"
	GenreThriller       Genre = "thriller"
	GenreWar            Genre = "war"
	GenreWestern        Genre = "western"
)
