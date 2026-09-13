package movie

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	moviedomain "github.com/masaya-nishimura-09/movie-log-api/internal/domain/movie"
)

type fakeUsecase struct {
	movie        *moviedomain.Movie
	searchResult *moviedomain.SearchResult
	err          error

	receivedMovieID         moviedomain.ID
	receivedTitle           moviedomain.Title
	receivedPage            moviedomain.Page
	receivedDisplayLanguage moviedomain.DisplayLanguage
}

func (u *fakeUsecase) GetByID(
	ctx context.Context,
	movieID moviedomain.ID,
	displayLanguage moviedomain.DisplayLanguage,
) (*moviedomain.Movie, error) {
	u.receivedMovieID = movieID
	u.receivedDisplayLanguage = displayLanguage
	return u.movie, u.err
}

func (u *fakeUsecase) SearchByTitle(
	ctx context.Context,
	title moviedomain.Title,
	page moviedomain.Page,
	displayLanguage moviedomain.DisplayLanguage,
) (*moviedomain.SearchResult, error) {
	u.receivedTitle = title
	u.receivedPage = page
	u.receivedDisplayLanguage = displayLanguage
	return u.searchResult, u.err
}

const wantGetByIDBody = `{"genres":["drama"],` +
	`"id":1,` +
	`"origin_country":["US"],` +
	`"original_language":"en",` +
	`"original_title":"Test Original Title",` +
	`"overview":"Test Overview",` +
	`"poster_url":"https://example.com/poster.jpg",` +
	`"release_year":2020,` +
	`"runtime":120,` +
	`"title":"Test Movie"}`

const wantSearchByTitleBody = `{"movies":[` +
	`{"id":1,"original_language":"en","original_title":"Test Original Title",` +
	`"title":"Test Movie","overview":"Test Overview",` +
	`"poster_url":"https://example.com/poster.jpg","release_year":2020},` +
	`{"id":2,"original_language":"fr","original_title":"Test Original Title 2",` +
	`"title":"Test Movie 2","overview":"Test Overview 2",` +
	`"poster_url":"https://example.com/poster2.jpg","release_year":2025}` +
	`],"page":2,"total_pages":1,"total_results":2}`

func newTestMovie() moviedomain.Movie {
	releaseYear := moviedomain.ReleaseYear(2020)

	return moviedomain.Movie{
		ID:               moviedomain.ID(1),
		Title:            moviedomain.Title("Test Movie"),
		OriginalTitle:    moviedomain.OriginalTitle("Test Original Title"),
		Overview:         moviedomain.Overview("Test Overview"),
		Genres:           []moviedomain.Genre{moviedomain.GenreDrama},
		PosterURL:        moviedomain.PosterURL("https://example.com/poster.jpg"),
		ReleaseYear:      &releaseYear,
		Runtime:          moviedomain.Runtime(120),
		OriginalLanguage: moviedomain.OriginalLanguage("en"),
		OriginCountry: []moviedomain.OriginCountry{
			moviedomain.OriginCountry("US"),
		},
	}
}

func newTestSearchResult() moviedomain.SearchResult {
	releaseYear := moviedomain.ReleaseYear(2025)

	movie1 := newTestMovie()
	movie2 := moviedomain.Movie{
		ID:               moviedomain.ID(2),
		Title:            moviedomain.Title("Test Movie 2"),
		OriginalTitle:    moviedomain.OriginalTitle("Test Original Title 2"),
		Overview:         moviedomain.Overview("Test Overview 2"),
		Genres:           []moviedomain.Genre{moviedomain.GenreAction},
		PosterURL:        moviedomain.PosterURL("https://example.com/poster2.jpg"),
		ReleaseYear:      &releaseYear,
		Runtime:          moviedomain.Runtime(180),
		OriginalLanguage: moviedomain.OriginalLanguage("fr"),
		OriginCountry: []moviedomain.OriginCountry{
			moviedomain.OriginCountry("FR"),
		},
	}

	return moviedomain.SearchResult{
		Page:         moviedomain.Page(2),
		TotalPages:   moviedomain.TotalPages(1),
		TotalResults: moviedomain.TotalResults(2),
		Movies:       []*moviedomain.Movie{&movie1, &movie2},
	}
}

func TestGetByID(t *testing.T) {
	t.Run(
		"returns the movie and 200 when the request is valid",
		func(t *testing.T) {
			m := newTestMovie()
			usecase := &fakeUsecase{movie: &m}
			movieHandler := NewMovieHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Params = gin.Params{{Key: "id", Value: "1"}}
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

			movieHandler.GetByID(c)
			if rec.Code != http.StatusOK {
				t.Errorf(
					"GetByID(c) code = %v, want %v",
					rec.Code, http.StatusOK,
				)
			}
			if rec.Body.String() != wantGetByIDBody {
				t.Errorf(
					"GetByID(c) body = %v, want %v",
					rec.Body.String(), wantGetByIDBody,
				)
			}
			if usecase.receivedMovieID != 1 {
				t.Errorf(
					"GetByID(c) movieID = %v, want %v",
					usecase.receivedMovieID, 1,
				)
			}
			if usecase.receivedDisplayLanguage != "en" {
				t.Errorf(
					"GetByID(c) displayLanguage = %v, want %v",
					usecase.receivedDisplayLanguage, "en",
				)
			}
		},
	)

	t.Run(
		"returns 400 when the path parameter is not a number",
		func(t *testing.T) {
			usecase := &fakeUsecase{}
			movieHandler := NewMovieHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Params = gin.Params{{Key: "id", Value: "invalid"}}
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

			movieHandler.GetByID(c)
			if rec.Code != http.StatusBadRequest {
				t.Errorf(
					"GetByID(c) code = %v, want %v",
					rec.Code, http.StatusBadRequest,
				)
			}
			want := `"code":"INVALID_INPUT"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"GetByID(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)

	t.Run(
		"returns 400 when display language is invalid",
		func(t *testing.T) {
			usecase := &fakeUsecase{}
			movieHandler := NewMovieHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Params = gin.Params{{Key: "id", Value: "1"}}
			c.Request = httptest.NewRequest(
				http.MethodGet, "/?language=invalid", nil,
			)

			movieHandler.GetByID(c)
			if rec.Code != http.StatusBadRequest {
				t.Errorf(
					"GetByID(c) code = %v, want %v",
					rec.Code, http.StatusBadRequest,
				)
			}
			want := `"code":"INVALID_INPUT"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"GetByID(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)

	t.Run(
		"returns 404 when the movie does not exist",
		func(t *testing.T) {
			usecase := &fakeUsecase{err: exception.ErrNotFound}
			movieHandler := NewMovieHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Params = gin.Params{{Key: "id", Value: "2"}}
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

			movieHandler.GetByID(c)
			if rec.Code != http.StatusNotFound {
				t.Errorf(
					"GetByID(c) code = %v, want %v",
					rec.Code, http.StatusNotFound,
				)
			}
			want := `"code":"MOVIE_NOT_FOUND"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"GetByID(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)

	t.Run(
		"returns 500 when the usecase returns an unexpected error",
		func(t *testing.T) {
			usecase := &fakeUsecase{err: errors.New("unexpected")}
			movieHandler := NewMovieHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Params = gin.Params{{Key: "id", Value: "1"}}
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

			movieHandler.GetByID(c)
			if rec.Code != http.StatusInternalServerError {
				t.Errorf(
					"GetByID(c) code = %v, want %v",
					rec.Code, http.StatusInternalServerError,
				)
			}
			want := `"code":"INTERNAL_SERVER_ERROR"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"GetByID(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)
}

func TestSearchByTitle(t *testing.T) {
	t.Run(
		"returns the search result and 200 when the request is valid",
		func(t *testing.T) {
			sr := newTestSearchResult()
			usecase := &fakeUsecase{searchResult: &sr}
			movieHandler := NewMovieHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(
				http.MethodGet,
				"/?title=test+movie&page=2&language=en",
				nil,
			)

			movieHandler.SearchByTitle(c)
			if rec.Code != http.StatusOK {
				t.Errorf(
					"SearchByTitle(c) code = %v, want %v",
					rec.Code, http.StatusOK,
				)
			}
			if rec.Body.String() != wantSearchByTitleBody {
				t.Errorf(
					"SearchByTitle(c) body = %v, want %v",
					rec.Body.String(), wantSearchByTitleBody,
				)
			}
			if usecase.receivedTitle != "test movie" {
				t.Errorf(
					"SearchByTitle(c) title = %v, want %v",
					usecase.receivedTitle, "test movie",
				)
			}
			if usecase.receivedPage != 2 {
				t.Errorf(
					"SearchByTitle(c) page = %v, want %v",
					usecase.receivedPage, 2,
				)
			}
			if usecase.receivedDisplayLanguage != "en" {
				t.Errorf(
					"SearchByTitle(c) displayLanguage = %v, want %v",
					usecase.receivedDisplayLanguage, "en",
				)
			}
		},
	)

	t.Run(
		"returns 400 when the title is invalid",
		func(t *testing.T) {
			usecase := &fakeUsecase{}
			movieHandler := NewMovieHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(
				http.MethodGet,
				"/?title=",
				nil,
			)

			movieHandler.SearchByTitle(c)
			if rec.Code != http.StatusBadRequest {
				t.Errorf(
					"SearchByTitle(c) code = %v, want %v",
					rec.Code, http.StatusBadRequest,
				)
			}
			want := `"code":"INVALID_INPUT"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"SearchByTitle(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)

	t.Run(
		"returns 400 when the page is not a number",
		func(t *testing.T) {
			usecase := &fakeUsecase{}
			movieHandler := NewMovieHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(
				http.MethodGet,
				"/?title=test+movie&page=invalid&language=en",
				nil,
			)

			movieHandler.SearchByTitle(c)
			if rec.Code != http.StatusBadRequest {
				t.Errorf(
					"SearchByTitle(c) code = %v, want %v",
					rec.Code, http.StatusBadRequest,
				)
			}
			want := `"code":"INVALID_INPUT"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"SearchByTitle(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)

	t.Run(
		"returns 400 when display language is invalid",
		func(t *testing.T) {
			usecase := &fakeUsecase{}
			movieHandler := NewMovieHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(
				http.MethodGet,
				"/?title=test+movie&page=2&language=invalid",
				nil,
			)

			movieHandler.SearchByTitle(c)
			if rec.Code != http.StatusBadRequest {
				t.Errorf(
					"SearchByTitle(c) code = %v, want %v",
					rec.Code, http.StatusBadRequest,
				)
			}
			want := `"code":"INVALID_INPUT"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"SearchByTitle(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)

	t.Run(
		"returns 500 when the usecase returns an unexpected error",
		func(t *testing.T) {
			usecase := &fakeUsecase{err: errors.New("unexpected")}
			movieHandler := NewMovieHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(
				http.MethodGet,
				"/?title=test+movie&page=2&language=en",
				nil,
			)

			movieHandler.SearchByTitle(c)
			if rec.Code != http.StatusInternalServerError {
				t.Errorf(
					"SearchByTitle(c) code = %v, want %v",
					rec.Code, http.StatusInternalServerError,
				)
			}
			want := `"code":"INTERNAL_SERVER_ERROR"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"SearchByTitle(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)
}
