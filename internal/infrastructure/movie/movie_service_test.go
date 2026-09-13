package movie

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"testing"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/movie"
)

func newTestMovie() movie.Movie {
	releaseYear := movie.ReleaseYear(2020)

	movie := movie.Movie{
		ID:               movie.ID(1),
		Title:            movie.Title("Test Movie"),
		OriginalTitle:    movie.OriginalTitle("Test Original Title"),
		Overview:         movie.Overview("Test Overview"),
		Genres:           []movie.Genre{movie.GenreDrama, movie.GenreThriller},
		PosterURL:        movie.PosterURL(posterBaseURL.JoinPath("/poster.jpg").String()),
		ReleaseYear:      &releaseYear,
		Runtime:          movie.Runtime(120),
		OriginalLanguage: movie.OriginalLanguage("en"),
		OriginCountry: []movie.OriginCountry{
			movie.OriginCountry("US"), movie.OriginCountry("JP"),
		},
	}

	return movie
}

func newTestSearchResult() movie.SearchResult {
	releaseYear := movie.ReleaseYear(2020)
	releaseYear2 := movie.ReleaseYear(2025)

	movie1 := movie.Movie{
		ID:               movie.ID(1),
		Title:            movie.Title("Test Movie"),
		OriginalTitle:    movie.OriginalTitle("Test Original Title"),
		Overview:         movie.Overview("Test Overview"),
		PosterURL:        movie.PosterURL(posterBaseURL.JoinPath("/poster.jpg").String()),
		ReleaseYear:      &releaseYear,
		OriginalLanguage: movie.OriginalLanguage("en"),
	}

	movie2 := movie.Movie{
		ID:               movie.ID(2),
		Title:            movie.Title("Test Movie 2"),
		OriginalTitle:    movie.OriginalTitle("Test Original Title 2"),
		Overview:         movie.Overview("Test Overview 2"),
		PosterURL:        movie.PosterURL(posterBaseURL.JoinPath("/poster2.jpg").String()),
		ReleaseYear:      &releaseYear2,
		OriginalLanguage: movie.OriginalLanguage("fr"),
	}

	searchResult := movie.SearchResult{
		Page:         movie.Page(1),
		Movies:       []*movie.Movie{&movie1, &movie2},
		TotalPages:   movie.TotalPages(1),
		TotalResults: movie.TotalResults(2),
	}

	return searchResult
}

func equalGenres(got, want []movie.Genre) bool {
	g := slices.Clone(got)
	w := slices.Clone(want)
	slices.Sort(g)
	slices.Sort(w)
	return slices.Equal(g, w)
}

func equalReleaseYear(got, want *movie.ReleaseYear) bool {
	if got == nil || want == nil {
		return got == want
	}
	return *got == *want
}

func equalOriginCountry(got, want []movie.OriginCountry) bool {
	g := slices.Clone(got)
	w := slices.Clone(want)
	slices.Sort(g)
	slices.Sort(w)
	return slices.Equal(g, w)
}

func assertMovieEqual(t *testing.T, call string, got, want *movie.Movie) {
	t.Helper()

	if got.ID != want.ID ||
		got.Title != want.Title ||
		got.OriginalTitle != want.OriginalTitle ||
		got.Overview != want.Overview ||
		got.PosterURL != want.PosterURL ||
		got.Runtime != want.Runtime ||
		got.OriginalLanguage != want.OriginalLanguage {
		t.Errorf(
			"%s = %v, want %v",
			call, got, want,
		)
	}
	if !equalGenres(got.Genres, want.Genres) {
		t.Errorf(
			"%s = %v, want Genres %v",
			call, got, want.Genres,
		)
	}
	if !equalReleaseYear(got.ReleaseYear, want.ReleaseYear) {
		t.Errorf(
			"%s = %v, want ReleaseYear %v",
			call, got, want.ReleaseYear,
		)
	}
	if !equalOriginCountry(got.OriginCountry, want.OriginCountry) {
		t.Errorf(
			"%s = %v, want OriginCountry %v",
			call, got, want.OriginCountry,
		)
	}
}

func assertSearchResultEqual(t *testing.T, call string, got, want *movie.SearchResult) {
	t.Helper()

	if got.Page != want.Page ||
		got.TotalPages != want.TotalPages ||
		got.TotalResults != want.TotalResults {
		t.Errorf(
			"%s = %v, want %v",
			call, got, want,
		)
	}

	for i, m := range got.Movies {
		if m.ID != want.Movies[i].ID ||
			m.Title != want.Movies[i].Title ||
			m.OriginalTitle != want.Movies[i].OriginalTitle ||
			m.Overview != want.Movies[i].Overview ||
			m.PosterURL != want.Movies[i].PosterURL ||
			m.OriginalLanguage != want.Movies[i].OriginalLanguage {
			t.Errorf(
				"%s = %v, want %v",
				call, got, want,
			)
		}
		if !equalReleaseYear(m.ReleaseYear, want.Movies[i].ReleaseYear) {
			t.Errorf(
				"%s = %v, want ReleaseYear %v",
				call, got, want.Movies[i].ReleaseYear,
			)
		}
	}
}

func TestGetByID(t *testing.T) {
	t.Run(
		"returns the movie with its associations when the ID exists",
		func(t *testing.T) {
			json := `{
				"id": 1,
				"title": "Test Movie",
				"original_title": "Test Original Title",
				"original_language": "en",
				"overview": "Test Overview",
				"genres": [
					{"id": 18, "name": "Drama"},
					{"id": 53, "name": "Thriller"}
				],
				"origin_country": ["US", "JP"],
				"poster_path": "/poster.jpg",
				"release_date": "2020-05-01",
				"runtime": 120
			}`
			srv := httptest.NewServer(http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					w.Write([]byte(json))
				},
			))
			defer srv.Close()
			endpoint, _ := url.Parse(srv.URL)

			ms := NewMovieService(
				NewTMDBClient(endpoint, "test-token"), posterBaseURL,
			)
			ctx := context.Background()

			displayLanguage := movie.DisplayLanguage("en")
			want := newTestMovie()

			got, err := ms.GetByID(ctx, want.ID, displayLanguage)
			if err != nil {
				t.Fatalf(
					"GetByID(ctx, %d, %v) (*movie.Movie, error) = %v, %v",
					want.ID, displayLanguage, got, err,
				)
			}
			assertMovieEqual(
				t,
				fmt.Sprintf(
					"GetByID(ctx, %d, %v) (*movie.Movie, error)",
					want.ID,
					displayLanguage,
				),
				got, &want,
			)
		},
	)

	t.Run(
		"returns ErrNotFound when the ID does not exist",
		func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
				},
			))
			defer srv.Close()
			endpoint, _ := url.Parse(srv.URL)

			ms := NewMovieService(
				NewTMDBClient(endpoint, "test-token"), posterBaseURL,
			)
			ctx := context.Background()
			fakeID := movie.ID(999999)
			displayLanguage := movie.DisplayLanguage("en")

			got, err := ms.GetByID(ctx, fakeID, displayLanguage)
			if !errors.Is(err, exception.ErrNotFound) {
				t.Fatalf(
					"GetByID(ctx, %d, %v) (*movie.Movie, error) = %v, %v, want %v",
					fakeID, displayLanguage, got, err, exception.ErrNotFound,
				)
			}
			if got != nil {
				t.Errorf(
					"GetByID(ctx, %d, %v) (*movie.Movie, error) = %v, want nil",
					fakeID, displayLanguage, got,
				)
			}
		},
	)

	t.Run(
		"returns an error when the response is invalid JSON",
		func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					w.Write([]byte("invalid json"))
				},
			))
			defer srv.Close()
			endpoint, _ := url.Parse(srv.URL)

			ms := NewMovieService(
				NewTMDBClient(endpoint, "test-token"), posterBaseURL,
			)
			ctx := context.Background()
			id := movie.ID(1)
			displayLanguage := movie.DisplayLanguage("en")

			got, err := ms.GetByID(ctx, id, displayLanguage)
			if err == nil {
				t.Fatalf(
					"GetByID(ctx, %d, %v) (*movie.Movie, error) = %v, nil, want error",
					id, displayLanguage, got,
				)
			}
			if got != nil {
				t.Errorf(
					"GetByID(ctx, %d, %v) (*movie.Movie, error) = %v, want nil",
					id, displayLanguage, got,
				)
			}
		},
	)

	t.Run(
		"returns a wrapped error when the context is canceled",
		func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {},
			))
			defer srv.Close()
			endpoint, _ := url.Parse(srv.URL)

			ms := NewMovieService(
				NewTMDBClient(endpoint, "test-token"), posterBaseURL,
			)
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			fakeID := movie.ID(999999)
			displayLanguage := movie.DisplayLanguage("en")

			got, err := ms.GetByID(ctx, fakeID, displayLanguage)
			if !errors.Is(err, context.Canceled) {
				t.Fatalf(
					"GetByID(ctx, %d, %v) (*movie.Movie, error) = %v, %v, want %v",
					fakeID, displayLanguage, got, err, context.Canceled,
				)
			}
			if got != nil {
				t.Errorf(
					"GetByID(ctx, %d, %v) (*movie.Movie, error) = %v, want nil",
					fakeID, displayLanguage, got,
				)
			}
		},
	)
}

func TestSearchByTitle(t *testing.T) {
	t.Run(
		"returns the search result when movies match the title",
		func(t *testing.T) {
			json := `{
				"page": 1,
				"results": [
					{
						"id": 1,
						"title": "Test Movie",
						"original_title": "Test Original Title",
						"original_language": "en",
						"overview": "Test Overview",
						"poster_path": "/poster.jpg",
						"release_date": "2020-05-01"
					},
					{
						"id": 2,
						"title": "Test Movie 2",
						"original_title": "Test Original Title 2",
						"original_language": "fr",
						"overview": "Test Overview 2",
						"poster_path": "/poster2.jpg",
						"release_date": "2025-05-01"
					}
				],
				"total_pages": 1,
				"total_results": 2
				
			}`

			srv := httptest.NewServer(http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					w.Write([]byte(json))
				},
			))
			defer srv.Close()
			endpoint, _ := url.Parse(srv.URL)

			ms := NewMovieService(
				NewTMDBClient(endpoint, "test-token"), posterBaseURL,
			)
			ctx := context.Background()
			title := movie.Title("Test")
			page := movie.Page(1)
			displayLanguage := movie.DisplayLanguage("en")

			want := newTestSearchResult()

			got, err := ms.SearchByTitle(ctx, title, page, displayLanguage)
			if err != nil {
				t.Fatalf(
					"SearchByTitle(ctx, %v, %d, %v) = %v, %v",
					title, page, displayLanguage, got, err,
				)
			}

			assertSearchResultEqual(
				t,
				fmt.Sprintf(
					"SearchByTitle(ctx, %v, %d, %v)",
					title, page, displayLanguage,
				),
				got, &want,
			)
		},
	)

	t.Run(
		"returns ErrNotFound when the title does not exist",
		func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
				},
			))
			defer srv.Close()
			endpoint, _ := url.Parse(srv.URL)

			ms := NewMovieService(
				NewTMDBClient(endpoint, "test-token"), posterBaseURL,
			)
			ctx := context.Background()
			fakeTitle := movie.Title("fake title")
			page := movie.Page(1)
			displayLanguage := movie.DisplayLanguage("en")

			got, err := ms.SearchByTitle(ctx, fakeTitle, page, displayLanguage)
			if !errors.Is(err, exception.ErrNotFound) {
				t.Fatalf(
					"SearchByTitle(ctx, %v, %d, %v) = %v, %v, want %v",
					fakeTitle, page, displayLanguage, got, err, exception.ErrNotFound,
				)
			}
			if got != nil {
				t.Errorf(
					"SearchByTitle(ctx, %v, %d, %v) = %v, want nil",
					fakeTitle, page, displayLanguage, got,
				)
			}
		},
	)

	t.Run(
		"returns an error when the response is invalid JSON",
		func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					w.Write([]byte("invalid json"))
				},
			))
			defer srv.Close()
			endpoint, _ := url.Parse(srv.URL)

			ms := NewMovieService(
				NewTMDBClient(endpoint, "test-token"), posterBaseURL,
			)
			ctx := context.Background()
			title := movie.Title("title")
			page := movie.Page(1)
			displayLanguage := movie.DisplayLanguage("en")

			got, err := ms.SearchByTitle(ctx, title, page, displayLanguage)
			if err == nil {
				t.Fatalf(
					"SearchByTitle(ctx, %v, %d, %v) = %v, want error",
					title, page, displayLanguage, got,
				)
			}
			if got != nil {
				t.Errorf(
					"SearchByTitle(ctx, %v, %d, %v) = %v, want nil",
					title, page, displayLanguage, got,
				)
			}
		},
	)

	t.Run(
		"returns a wrapped error when the context is canceled",
		func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {},
			))
			defer srv.Close()
			endpoint, _ := url.Parse(srv.URL)

			ms := NewMovieService(
				NewTMDBClient(endpoint, "test-token"), posterBaseURL,
			)
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			title := movie.Title("title")
			page := movie.Page(1)
			displayLanguage := movie.DisplayLanguage("en")

			got, err := ms.SearchByTitle(ctx, title, page, displayLanguage)
			if !errors.Is(err, context.Canceled) {
				t.Fatalf(
					"SearchByTitle(ctx, %v, %d, %v) = %v, %v, want %v",
					title, page, displayLanguage, got, err, context.Canceled,
				)
			}
			if got != nil {
				t.Errorf(
					"SearchByTitle(ctx, %v, %d, %v) = %v, want nil",
					title, page, displayLanguage, got,
				)
			}
		},
	)
}
