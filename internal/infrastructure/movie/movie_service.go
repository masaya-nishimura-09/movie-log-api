package movie

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/movie"
)

var genreIDName = map[uint]movie.Genre{
	28:    movie.GenreAction,
	12:    movie.GenreAdventure,
	16:    movie.GenreAnimation,
	35:    movie.GenreComedy,
	80:    movie.GenreCrime,
	99:    movie.GenreDocumentary,
	18:    movie.GenreDrama,
	10751: movie.GenreFamily,
	14:    movie.GenreFantasy,
	36:    movie.GenreHistory,
	27:    movie.GenreHorror,
	10402: movie.GenreMusic,
	9648:  movie.GenreMystery,
	10749: movie.GenreRomance,
	878:   movie.GenreScienceFiction,
	10770: movie.GenreTVMovie,
	53:    movie.GenreThriller,
	10752: movie.GenreWar,
	37:    movie.GenreWestern,
}

type movieService struct {
	client        *TMDBClient
	posterBaseURL *url.URL
}

func NewMovieService(
	client *TMDBClient, posterBaseURL *url.URL,
) movie.MovieService {
	return &movieService{client: client, posterBaseURL: posterBaseURL}
}

type searchMovieDTO struct {
	Page         uint                   `json:"page"`
	Results      []searchMovieResultDTO `json:"results"`
	TotalPages   uint                   `json:"total_pages"`
	TotalResults uint                   `json:"total_results"`
}

type searchMovieResultDTO struct {
	ID               uint   `json:"id"`
	Title            string `json:"title"`
	OriginalTitle    string `json:"original_title"`
	OriginalLanguage string `json:"original_language"`
	Overview         string `json:"overview"`
	PosterPath       string `json:"poster_path"`
	ReleaseDate      string `json:"release_date"`
}

type movieDTO struct {
	ID               uint       `json:"id"`
	Title            string     `json:"title"`
	OriginalTitle    string     `json:"original_title"`
	OriginCountry    []string   `json:"origin_country"`
	OriginalLanguage string     `json:"original_language"`
	Genres           []genreDTO `json:"genres"`
	Overview         string     `json:"overview"`
	PosterPath       string     `json:"poster_path"`
	ReleaseDate      string     `json:"release_date"`
	Runtime          uint       `json:"runtime"`
}

type genreDTO struct {
	ID uint `json:"id"`
}

func (ms *movieService) toMovie(dto *movieDTO) *movie.Movie {
	var genres []movie.Genre
	for _, g := range dto.Genres {
		if name, ok := genreIDName[g.ID]; ok {
			genres = append(genres, name)
		}
	}

	var originCountry []movie.OriginCountry
	for _, c := range dto.OriginCountry {
		originCountry = append(originCountry, movie.OriginCountry(c))
	}

	return &movie.Movie{
		ID:               movie.ID(dto.ID),
		Title:            movie.Title(dto.Title),
		OriginalTitle:    movie.OriginalTitle(dto.OriginalTitle),
		Overview:         movie.Overview(dto.Overview),
		Genres:           genres,
		PosterURL:        ms.toPosterURL(dto.PosterPath),
		ReleaseYear:      toReleaseYear(dto.ReleaseDate),
		Runtime:          movie.Runtime(dto.Runtime),
		OriginalLanguage: movie.OriginalLanguage(dto.OriginalLanguage),
		OriginCountry:    originCountry,
	}
}

func (ms *movieService) toSearchResult(dto *searchMovieDTO) *movie.SearchResult {
	var movies []*movie.Movie
	for _, r := range dto.Results {
		m := movie.Movie{
			ID:               movie.ID(r.ID),
			OriginalLanguage: movie.OriginalLanguage(r.OriginalLanguage),
			OriginalTitle:    movie.OriginalTitle(r.OriginalTitle),
			Title:            movie.Title(r.Title),
			Overview:         movie.Overview(r.Overview),
			PosterURL:        ms.toPosterURL(r.PosterPath),
			ReleaseYear:      toReleaseYear(r.ReleaseDate),
		}
		movies = append(movies, &m)
	}

	return &movie.SearchResult{
		Page:         movie.Page(dto.Page),
		Movies:       movies,
		TotalPages:   movie.TotalPages(dto.TotalPages),
		TotalResults: movie.TotalResults(dto.TotalResults),
	}
}

func (ms *movieService) toPosterURL(path string) movie.PosterURL {
	if path == "" {
		return ""
	}

	return movie.PosterURL(ms.posterBaseURL.JoinPath(path).String())
}

func toReleaseYear(date string) *movie.ReleaseYear {
	t, err := time.Parse(time.DateOnly, date)
	if err != nil {
		return nil
	}
	y := movie.ReleaseYear(t.Year())
	return &y
}

func (ms *movieService) GetByID(
	ctx context.Context,
	movieID movie.ID,
	displayLanguage movie.DisplayLanguage,
) (*movie.Movie, error) {
	var dto movieDTO

	query := url.Values{}
	query.Set("language", string(displayLanguage))

	body, err := ms.client.Get(
		ctx, fmt.Sprintf("/movie/%d", movieID), query,
	)
	if err != nil {
		return nil, fmt.Errorf("request TMDB movie detail: %w", err)
	}

	if err := json.Unmarshal(body, &dto); err != nil {
		return nil, fmt.Errorf("unmarshal TMDB get response: %w", err)
	}
	return ms.toMovie(&dto), nil
}

func (ms *movieService) SearchByTitle(
	ctx context.Context,
	title movie.Title,
	page movie.Page,
	displayLanguage movie.DisplayLanguage,
) (*movie.SearchResult, error) {
	var dto searchMovieDTO

	query := url.Values{}
	query.Set("query", string(title))
	query.Set("page", fmt.Sprintf("%d", page))
	query.Set("language", string(displayLanguage))
	query.Set("include_adult", "true")

	body, err := ms.client.Get(ctx, "/search/movie", query)
	if err != nil {
		return nil, fmt.Errorf("request TMDB search: %w", err)
	}

	if err := json.Unmarshal(body, &dto); err != nil {
		return nil, fmt.Errorf("unmarshal TMDB search response: %w", err)
	}
	return ms.toSearchResult(&dto), nil
}
