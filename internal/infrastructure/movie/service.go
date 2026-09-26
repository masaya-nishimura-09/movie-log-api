package movie

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
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

var genderIDName = map[uint]movie.Gender{
	0: movie.GenderOther,
	1: movie.GenderFemale,
	2: movie.GenderMale,
}

type service struct {
	client        *TMDBClient
	posterBaseURL *url.URL
}

func NewService(
	client *TMDBClient, posterBaseURL *url.URL,
) movie.Service {
	return &service{client: client, posterBaseURL: posterBaseURL}
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

type creditsDTO struct {
	Casts []castDTO `json:"cast"`
}

type castDTO struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	OriginalName string `json:"original_name"`
	Character    string `json:"character"`
	Role         string `json:"known_for_department"`
	Gender       uint   `json:"gender"`
}

func (s *service) toMovie(movieDto *movieDTO, castDtos []castDTO) *movie.Movie {
	var genres []movie.Genre
	for _, g := range movieDto.Genres {
		if name, ok := genreIDName[g.ID]; ok {
			genres = append(genres, name)
		}
	}

	var originCountry []movie.OriginCountry
	for _, c := range movieDto.OriginCountry {
		originCountry = append(originCountry, movie.OriginCountry(c))
	}

	var casts []movie.Cast
	for _, c := range castDtos {
		gender := movie.GenderOther
		if name, ok := genderIDName[c.Gender]; ok {
			gender = name
		}

		cast := movie.Cast{
			ID:           movie.CastID(c.ID),
			Name:         movie.CastName(c.Name),
			OriginalName: movie.OriginalCastName(c.OriginalName),
			Character:    movie.Character(c.Character),
			Role:         movie.Role(c.Role),
			Gender:       gender,
		}
		casts = append(casts, cast)
	}

	return &movie.Movie{
		ID:               movie.ID(movieDto.ID),
		Title:            movie.Title(movieDto.Title),
		OriginalTitle:    movie.OriginalTitle(movieDto.OriginalTitle),
		Overview:         movie.Overview(movieDto.Overview),
		Genres:           genres,
		PosterURL:        s.toPosterURL(movieDto.PosterPath),
		ReleaseYear:      toReleaseYear(movieDto.ReleaseDate),
		Runtime:          movie.Runtime(movieDto.Runtime),
		OriginalLanguage: movie.OriginalLanguage(movieDto.OriginalLanguage),
		OriginCountry:    originCountry,
		Casts:            casts,
	}
}

func (s *service) toSearchResult(dto *searchMovieDTO) *movie.SearchResult {
	var movies []*movie.Movie
	for _, r := range dto.Results {
		m := movie.Movie{
			ID:               movie.ID(r.ID),
			OriginalLanguage: movie.OriginalLanguage(r.OriginalLanguage),
			OriginalTitle:    movie.OriginalTitle(r.OriginalTitle),
			Title:            movie.Title(r.Title),
			Overview:         movie.Overview(r.Overview),
			PosterURL:        s.toPosterURL(r.PosterPath),
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

func (s *service) toPosterURL(path string) movie.PosterURL {
	if path == "" {
		return ""
	}

	return movie.PosterURL(s.posterBaseURL.JoinPath(path).String())
}

func toReleaseYear(date string) *movie.ReleaseYear {
	t, err := time.Parse(time.DateOnly, date)
	if err != nil {
		return nil
	}
	y := movie.ReleaseYear(t.Year())
	return &y
}

func (s *service) GetByID(
	ctx context.Context,
	movieID movie.ID,
	displayLanguage movie.DisplayLanguage,
) (*movie.Movie, error) {
	var movieDto movieDTO
	var creditsDto creditsDTO

	query := url.Values{}
	query.Set("language", string(displayLanguage))

	movieBody, err := s.client.Get(
		ctx, fmt.Sprintf("/movie/%d", movieID), query,
	)
	if err != nil {
		return nil, fmt.Errorf("request TMDB movie detail: %w", err)
	}

	if err := json.Unmarshal(movieBody, &movieDto); err != nil {
		return nil, fmt.Errorf("unmarshal TMDB get response: %w", err)
	}

	creditsBody, err := s.client.Get(
		ctx, fmt.Sprintf("/movie/%d/credits", movieID), query,
	)
	if err != nil {
		if errors.Is(err, exception.ErrNotFound) {
			return s.toMovie(&movieDto, nil), nil
		}
		return nil, fmt.Errorf("request TMDB casts: %w", err)
	}

	if err := json.Unmarshal(creditsBody, &creditsDto); err != nil {
		return nil, fmt.Errorf("unmarshal TMDB get response: %w", err)
	}

	return s.toMovie(&movieDto, creditsDto.Casts), nil
}

func (s *service) SearchByTitle(
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

	body, err := s.client.Get(ctx, "/search/movie", query)
	if err != nil {
		return nil, fmt.Errorf("request TMDB search: %w", err)
	}

	if err := json.Unmarshal(body, &dto); err != nil {
		return nil, fmt.Errorf("unmarshal TMDB search response: %w", err)
	}
	return s.toSearchResult(&dto), nil
}
