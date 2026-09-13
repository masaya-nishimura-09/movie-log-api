package movie

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	moviedomain "github.com/masaya-nishimura-09/movie-log-api/internal/domain/movie"
	"github.com/masaya-nishimura-09/movie-log-api/internal/handler/response"
	movieusecase "github.com/masaya-nishimura-09/movie-log-api/internal/usecase/movie"
)

type MovieResponse struct {
	ID               uint   `json:"id"`
	OriginalLanguage string `json:"original_language"`
	OriginalTitle    string `json:"original_title"`
	Title            string `json:"title"`
	Overview         string `json:"overview"`
	PosterURL        string `json:"poster_url"`
	ReleaseYear      uint   `json:"release_year"`
}

func toGetByIDResponse(m *moviedomain.Movie) gin.H {
	genres := make([]string, 0, len(m.Genres))
	for _, genre := range m.Genres {
		genres = append(genres, string(genre))
	}

	countries := make([]string, 0, len(m.OriginCountry))
	for _, country := range m.OriginCountry {
		countries = append(countries, string(country))
	}

	var releaseYear uint
	if m.ReleaseYear != nil {
		releaseYear = uint(*m.ReleaseYear)
	}

	return gin.H{
		"id":                uint(m.ID),
		"title":             string(m.Title),
		"original_title":    string(m.OriginalTitle),
		"overview":          string(m.Overview),
		"genres":            genres,
		"poster_url":        string(m.PosterURL),
		"release_year":      releaseYear,
		"runtime":           uint(m.Runtime),
		"original_language": string(m.OriginalLanguage),
		"origin_country":    countries,
	}
}

func toMovieResponse(m *moviedomain.Movie) MovieResponse {
	var releaseYear uint
	if m.ReleaseYear != nil {
		releaseYear = uint(*m.ReleaseYear)
	}

	return MovieResponse{
		ID:               uint(m.ID),
		OriginalLanguage: string(m.OriginalLanguage),
		OriginalTitle:    string(m.OriginalTitle),
		Title:            string(m.Title),
		Overview:         string(m.Overview),
		PosterURL:        string(m.PosterURL),
		ReleaseYear:      releaseYear,
	}
}

func toSearchMovieResponse(sr *moviedomain.SearchResult) gin.H {
	movies := make([]MovieResponse, 0, len(sr.Movies))
	for _, m := range sr.Movies {
		movies = append(movies, toMovieResponse(m))
	}

	return gin.H{
		"page":          uint(sr.Page),
		"total_pages":   uint(sr.TotalPages),
		"total_results": uint(sr.TotalResults),
		"movies":        movies,
	}
}

type MovieHandler struct {
	movieUsecase movieusecase.Usecase
}

func NewMovieHandler(movieUsecase movieusecase.Usecase) *MovieHandler {
	return &MovieHandler{movieUsecase: movieUsecase}
}

func getMovieID(c *gin.Context) (moviedomain.ID, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidInput(c, err)
		return 0, false
	}
	return moviedomain.ID(id), true
}

func getDisplayLanguage(c *gin.Context) (moviedomain.DisplayLanguage, bool) {
	l := c.Query("language")
	if l == "" {
		l = "en"
	}
	dl, err := moviedomain.NewDisplayLanguage(l)
	if err != nil {
		response.InvalidInput(c, err)
		return "", false
	}
	return dl, true
}

func getTitle(c *gin.Context) (moviedomain.Title, bool) {
	t := c.Query("title")
	dt, err := moviedomain.NewTitle(t)
	if err != nil {
		response.InvalidInput(c, err)
		return "", false
	}
	return dt, true
}

func getPage(c *gin.Context) (moviedomain.Page, bool) {
	p := c.Query("page")
	if p == "" {
		p = "1"
	}
	page, err := strconv.ParseUint(p, 10, 64)
	if err != nil {
		response.InvalidInput(c, err)
		return 0, false
	}
	return moviedomain.Page(page), true
}

func (mh *MovieHandler) GetByID(c *gin.Context) {
	ctx := c.Request.Context()

	movieID, ok := getMovieID(c)
	if !ok {
		return
	}

	displayLanguage, ok := getDisplayLanguage(c)
	if !ok {
		return
	}

	m, err := mh.movieUsecase.GetByID(ctx, movieID, displayLanguage)
	if errors.Is(err, exception.ErrNotFound) {
		response.MovieNotFound(c)
		return
	}
	if err != nil {
		log.Println(err)
		response.InternalServerError(c)
		return
	}

	c.JSON(http.StatusOK, toGetByIDResponse(m))
}

func (mh *MovieHandler) SearchByTitle(c *gin.Context) {
	ctx := c.Request.Context()

	title, ok := getTitle(c)
	if !ok {
		return
	}

	page, ok := getPage(c)
	if !ok {
		return
	}

	displayLanguage, ok := getDisplayLanguage(c)
	if !ok {
		return
	}

	sr, err := mh.movieUsecase.SearchByTitle(
		ctx, title, page, displayLanguage,
	)
	if err != nil {
		log.Println(err)
		response.InternalServerError(c)
		return
	}

	c.JSON(http.StatusOK, toSearchMovieResponse(sr))
}
