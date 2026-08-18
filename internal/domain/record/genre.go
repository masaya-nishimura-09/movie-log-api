package record

import (
	"fmt"
	"slices"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
)

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

func NewGenre(value string) (Genre, error) {
	if value == "" {
		return "", fmt.Errorf("%w: genre is required", exception.ErrInvalid)
	}

	switch genre := Genre(value); genre {
	case GenreAction,
		GenreAdventure,
		GenreAnimation,
		GenreComedy,
		GenreCrime,
		GenreDocumentary,
		GenreDrama,
		GenreFamily,
		GenreFantasy,
		GenreHistory,
		GenreHorror,
		GenreMusic,
		GenreMystery,
		GenreRomance,
		GenreScienceFiction,
		GenreTVMovie,
		GenreThriller,
		GenreWar,
		GenreWestern:
		return genre, nil
	default:
		return "", fmt.Errorf("%w: invalid genre", exception.ErrInvalid)
	}
}

func NewGenres(values []string) ([]Genre, error) {
	genres := make([]Genre, 0, len(values))

	for _, value := range values {
		genre, err := NewGenre(value)
		if err != nil {
			return nil, err
		}

		if slices.Contains(genres, genre) {
			return nil, fmt.Errorf("%w: duplicate genre", exception.ErrInvalid)
		}

		genres = append(genres, genre)
	}

	return genres, nil
}
