package movie

import (
	"context"
	"fmt"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/movie"
)

type Usecase interface {
	GetByID(
		ctx context.Context,
		movieID movie.ID,
		language movie.DisplayLanguage,
	) (*movie.Movie, error)
	SearchByTitle(
		ctx context.Context,
		title movie.Title,
		page movie.Page,
		language movie.DisplayLanguage,
	) (*movie.SearchResult, error)
}

type MovieUsecase struct {
	movieService movie.MovieService
}

func NewMovieUsecase(
	movieService movie.MovieService,
) *MovieUsecase {
	return &MovieUsecase{
		movieService: movieService,
	}
}

func (mu *MovieUsecase) GetByID(
	ctx context.Context,
	movieID movie.ID,
	displayLanguage movie.DisplayLanguage,
) (*movie.Movie, error) {
	m, err := mu.movieService.GetByID(ctx, movieID, displayLanguage)
	if err != nil {
		return nil, fmt.Errorf("get movie by id: %w", err)
	}

	return m, nil
}

func (mu *MovieUsecase) SearchByTitle(
	ctx context.Context,
	title movie.Title,
	page movie.Page,
	displayLanguage movie.DisplayLanguage,
) (*movie.SearchResult, error) {
	sr, err := mu.movieService.SearchByTitle(ctx, title, page, displayLanguage)
	if err != nil {
		return nil, fmt.Errorf("search movies by title: %w", err)
	}

	return sr, nil
}
