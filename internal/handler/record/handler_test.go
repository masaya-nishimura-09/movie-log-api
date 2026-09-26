package record

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	recorddomain "github.com/masaya-nishimura-09/movie-log-api/internal/domain/record"
	userdomain "github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
)

type fakeUsecase struct {
	record          *recorddomain.Record
	listResult      recorddomain.ListResult
	listedQuery     recorddomain.Query
	createdUserID   userdomain.ID
	createdRecord   recorddomain.Record
	updatedUserID   userdomain.ID
	updatedRecordID recorddomain.ID
	updatedRecord   recorddomain.Record
	deletedUserID   userdomain.ID
	deletedRecordID recorddomain.ID
	err             error
}

func (u *fakeUsecase) GetByID(
	ctx context.Context,
	userID userdomain.ID,
	recordID recorddomain.ID,
) (*recorddomain.Record, error) {
	return u.record, u.err
}

func (u *fakeUsecase) ListByUserID(
	ctx context.Context,
	userID userdomain.ID,
	query recorddomain.Query,
) (recorddomain.ListResult, error) {
	u.listedQuery = query

	return u.listResult, u.err
}

func (u *fakeUsecase) Create(
	ctx context.Context,
	userID userdomain.ID,
	r recorddomain.Record,
) (*recorddomain.Record, error) {
	u.createdUserID = userID
	u.createdRecord = r

	return u.record, u.err
}

func (u *fakeUsecase) Update(
	ctx context.Context,
	userID userdomain.ID,
	recordID recorddomain.ID,
	r recorddomain.Record,
) (*recorddomain.Record, error) {
	u.updatedUserID = userID
	u.updatedRecordID = recordID
	u.updatedRecord = r

	return u.record, u.err
}

func (u *fakeUsecase) Delete(
	ctx context.Context,
	userID userdomain.ID,
	recordID recorddomain.ID,
) error {
	u.deletedUserID = userID
	u.deletedRecordID = recordID

	return u.err
}

const validBody = `{
	"title":"Test Movie",
	"release_year":2020,
	"runtime":120,
	"genres":["drama"],
	"countries":["US"],
	"language":"en",
	"credits":[{"person_name":"Test Director","credit_role":"director"}],
	"poster_url":"https://example.com/poster.jpg",
	"watched_at":"2026-01-01T00:00:00Z",
	"platform":"netflix",
	"score":4,
	"mood_tags":["moving"],
	"memo":"test memo"
}`

const wantBody = `{"countries":["US"],` +
	`"credits":[{"credit_role":"director","person_name":"Test Director"}],` +
	`"genres":["drama"],"language":"en","memo":"test memo","mood_tags":["moving"],` +
	`"platform":"netflix","poster_url":"https://example.com/poster.jpg",` +
	`"record_id":"10","release_year":2020,"runtime":120,"score":4,` +
	`"title":"Test Movie","watched_at":"2026-01-01T00:00:00Z"}`

func newTestRecord() recorddomain.Record {
	return recorddomain.Record{
		ID:          recorddomain.ID(10),
		UserID:      userdomain.ID(1),
		Title:       recorddomain.Title("Test Movie"),
		ReleaseYear: recorddomain.ReleaseYear(2020),
		Runtime:     recorddomain.Runtime(120),
		Genres:      []recorddomain.Genre{recorddomain.GenreDrama},
		Countries:   []recorddomain.Country{recorddomain.Country("US")},
		Language:    recorddomain.Language("en"),
		Credits: []recorddomain.Credit{
			{
				PersonName: recorddomain.PersonName("Test Director"),
				CreditRole: recorddomain.CreditRoleDirector,
			},
		},
		PosterURL: recorddomain.PosterURL("https://example.com/poster.jpg"),
		WatchedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		Platform:  recorddomain.PlatformNetflix,
		Score:     recorddomain.Score(4),
		MoodTags:  []recorddomain.MoodTag{recorddomain.MoodTagMoving},
		Memo:      recorddomain.Memo("test memo"),
	}
}

func TestCreate(t *testing.T) {
	t.Run(
		"passes the converted values to the usecase and returns 201 when the request is valid",
		func(t *testing.T) {
			userID := userdomain.ID(1)
			r := newTestRecord()
			usecase := &fakeUsecase{record: &r}
			recordHandler := NewRecordHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userID)
			c.Request = httptest.NewRequest(
				http.MethodPost, "/", strings.NewReader(validBody),
			)

			recordHandler.Create(c)
			if rec.Code != http.StatusCreated {
				t.Errorf(
					"Create(c) code = %v, want %v",
					rec.Code, http.StatusCreated,
				)
			}
			if rec.Body.String() != wantBody {
				t.Errorf(
					"Create(c) body = %v, want %v",
					rec.Body.String(), wantBody,
				)
			}

			if usecase.createdUserID != userID {
				t.Errorf(
					"Create(c) usecase user id = %v, want %v",
					usecase.createdUserID, userID,
				)
			}
			if usecase.createdRecord.Title != r.Title ||
				usecase.createdRecord.Platform != r.Platform ||
				usecase.createdRecord.Score != r.Score {
				t.Errorf(
					"Create(c) usecase record = %v, want %v",
					usecase.createdRecord, r,
				)
			}
		},
	)

	t.Run(
		"returns 500 when the authenticated user ID is missing from the context",
		func(t *testing.T) {
			usecase := &fakeUsecase{}
			recordHandler := NewRecordHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(
				http.MethodPost, "/", strings.NewReader(validBody),
			)

			recordHandler.Create(c)
			if rec.Code != http.StatusInternalServerError {
				t.Errorf(
					"Create(c) code = %v, want %v",
					rec.Code, http.StatusInternalServerError,
				)
			}
			want := `"code":"INTERNAL_SERVER_ERROR"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"Create(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)

	t.Run(
		"returns 400 when the request body is not valid JSON",
		func(t *testing.T) {
			usecase := &fakeUsecase{}
			recordHandler := NewRecordHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Request = httptest.NewRequest(
				http.MethodPost, "/", strings.NewReader(`{"title":"Test Movie",}`),
			)

			recordHandler.Create(c)
			if rec.Code != http.StatusBadRequest {
				t.Errorf(
					"Create(c) code = %v, want %v",
					rec.Code, http.StatusBadRequest,
				)
			}
			want := `"code":"INVALID_INPUT","message":"malformed request body"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"Create(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)

	t.Run(
		"returns 400 when the genre is not supported",
		func(t *testing.T) {
			usecase := &fakeUsecase{}
			recordHandler := NewRecordHandler(usecase)

			body := strings.Replace(validBody, `["drama"]`, `["invalid"]`, 1)
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Request = httptest.NewRequest(
				http.MethodPost, "/", strings.NewReader(body),
			)

			recordHandler.Create(c)
			if rec.Code != http.StatusBadRequest {
				t.Errorf(
					"Create(c) code = %v, want %v",
					rec.Code, http.StatusBadRequest,
				)
			}
			want := `"code":"INVALID_INPUT"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"Create(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)

	t.Run(
		"returns 500 when the usecase returns an unexpected error",
		func(t *testing.T) {
			usecase := &fakeUsecase{err: errors.New("unexpected")}
			recordHandler := NewRecordHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Request = httptest.NewRequest(
				http.MethodPost, "/", strings.NewReader(validBody),
			)

			recordHandler.Create(c)
			if rec.Code != http.StatusInternalServerError {
				t.Errorf(
					"Create(c) code = %v, want %v",
					rec.Code, http.StatusInternalServerError,
				)
			}
			want := `"code":"INTERNAL_SERVER_ERROR"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"Create(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)
}

func TestGetByID(t *testing.T) {
	t.Run(
		"returns the record and 200 when the request is valid",
		func(t *testing.T) {
			r := newTestRecord()
			usecase := &fakeUsecase{record: &r}
			recordHandler := NewRecordHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Params = gin.Params{{Key: "id", Value: "10"}}
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

			recordHandler.GetByID(c)
			if rec.Code != http.StatusOK {
				t.Errorf(
					"GetByID(c) code = %v, want %v",
					rec.Code, http.StatusOK,
				)
			}
			if rec.Body.String() != wantBody {
				t.Errorf(
					"GetByID(c) body = %v, want %v",
					rec.Body.String(), wantBody,
				)
			}
		},
	)

	t.Run(
		"returns 400 when the path parameter is not a number",
		func(t *testing.T) {
			usecase := &fakeUsecase{}
			recordHandler := NewRecordHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Params = gin.Params{{Key: "id", Value: "invalid"}}
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

			recordHandler.GetByID(c)
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
		"returns 404 when the record does not exist",
		func(t *testing.T) {
			usecase := &fakeUsecase{err: exception.ErrNotFound}
			recordHandler := NewRecordHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Params = gin.Params{{Key: "id", Value: "10"}}
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

			recordHandler.GetByID(c)
			if rec.Code != http.StatusNotFound {
				t.Errorf(
					"GetByID(c) code = %v, want %v",
					rec.Code, http.StatusNotFound,
				)
			}
			want := `"code":"RECORD_NOT_FOUND"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"GetByID(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)

	t.Run(
		"returns 500 when the authenticated user ID is missing from the context",
		func(t *testing.T) {
			usecase := &fakeUsecase{}
			recordHandler := NewRecordHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Params = gin.Params{{Key: "id", Value: "10"}}
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

			recordHandler.GetByID(c)
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

func TestListByUserID(t *testing.T) {
	t.Run(
		"passes the converted query to the usecase and returns the records and 200 when the request is valid",
		func(t *testing.T) {
			r := newTestRecord()
			usecase := &fakeUsecase{
				listResult: recorddomain.ListResult{
					Records:    []*recorddomain.Record{&r},
					TotalCount: recorddomain.TotalCount(1),
				},
			}
			recordHandler := NewRecordHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Request = httptest.NewRequest(
				http.MethodGet,
				"/?scores=4&scores=5&platforms=netflix&moodTags=moving&genres=drama"+
					"&title=test+movie&sortField=title&sortOrder=asc&page=2&perPage=10",
				nil,
			)

			recordHandler.ListByUserID(c)
			if rec.Code != http.StatusOK {
				t.Errorf(
					"ListByUserID(c) code = %v, want %v",
					rec.Code, http.StatusOK,
				)
			}
			want := `{"records":[` + wantBody + `],"total_count":1}`
			if rec.Body.String() != want {
				t.Errorf(
					"ListByUserID(c) body = %v, want %v",
					rec.Body.String(), want,
				)
			}

			wantScores := []recorddomain.Score{4, 5}
			if !slices.Equal(usecase.listedQuery.Scores, wantScores) {
				t.Errorf(
					"ListByUserID(c) scores = %v, want %v",
					usecase.listedQuery.Scores, wantScores,
				)
			}
			wantPlatforms := []recorddomain.Platform{recorddomain.PlatformNetflix}
			if !slices.Equal(usecase.listedQuery.Platforms, wantPlatforms) {
				t.Errorf(
					"ListByUserID(c) platforms = %v, want %v",
					usecase.listedQuery.Platforms, wantPlatforms,
				)
			}
			wantMoodTags := []recorddomain.MoodTag{recorddomain.MoodTagMoving}
			if !slices.Equal(usecase.listedQuery.MoodTags, wantMoodTags) {
				t.Errorf(
					"ListByUserID(c) moodTags = %v, want %v",
					usecase.listedQuery.MoodTags, wantMoodTags,
				)
			}
			wantGenres := []recorddomain.Genre{recorddomain.GenreDrama}
			if !slices.Equal(usecase.listedQuery.Genres, wantGenres) {
				t.Errorf(
					"ListByUserID(c) genres = %v, want %v",
					usecase.listedQuery.Genres, wantGenres,
				)
			}
			if usecase.listedQuery.TitleKeyword != "test movie" {
				t.Errorf(
					"ListByUserID(c) title = %v, want %v",
					usecase.listedQuery.TitleKeyword, "test movie",
				)
			}
			if usecase.listedQuery.SortField != recorddomain.SortFieldTitle {
				t.Errorf(
					"ListByUserID(c) sortField = %v, want %v",
					usecase.listedQuery.SortField, recorddomain.SortFieldTitle,
				)
			}
			if usecase.listedQuery.SortOrder != recorddomain.SortOrderAsc {
				t.Errorf(
					"ListByUserID(c) sortOrder = %v, want %v",
					usecase.listedQuery.SortOrder, recorddomain.SortOrderAsc,
				)
			}
			if usecase.listedQuery.Page != 2 {
				t.Errorf(
					"ListByUserID(c) page = %v, want %v",
					usecase.listedQuery.Page, 2,
				)
			}
			if usecase.listedQuery.PerPage != 10 {
				t.Errorf(
					"ListByUserID(c) perPage = %v, want %v",
					usecase.listedQuery.PerPage, 10,
				)
			}
		},
	)

	t.Run(
		"passes the default query to the usecase when the query parameters are omitted",
		func(t *testing.T) {
			usecase := &fakeUsecase{}
			recordHandler := NewRecordHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

			recordHandler.ListByUserID(c)
			if rec.Code != http.StatusOK {
				t.Errorf(
					"ListByUserID(c) code = %v, want %v",
					rec.Code, http.StatusOK,
				)
			}
			if usecase.listedQuery.SortField != recorddomain.SortFieldWatchedAt {
				t.Errorf(
					"ListByUserID(c) sortField = %v, want %v",
					usecase.listedQuery.SortField, recorddomain.SortFieldWatchedAt,
				)
			}
			if usecase.listedQuery.SortOrder != recorddomain.SortOrderDesc {
				t.Errorf(
					"ListByUserID(c) sortOrder = %v, want %v",
					usecase.listedQuery.SortOrder, recorddomain.SortOrderDesc,
				)
			}
			if usecase.listedQuery.Page != 1 {
				t.Errorf(
					"ListByUserID(c) page = %v, want %v",
					usecase.listedQuery.Page, 1,
				)
			}
			if usecase.listedQuery.PerPage != 20 {
				t.Errorf(
					"ListByUserID(c) perPage = %v, want %v",
					usecase.listedQuery.PerPage, 20,
				)
			}
		},
	)

	t.Run(
		"returns an empty array when the user has no records",
		func(t *testing.T) {
			usecase := &fakeUsecase{}
			recordHandler := NewRecordHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

			recordHandler.ListByUserID(c)
			if rec.Code != http.StatusOK {
				t.Errorf(
					"ListByUserID(c) code = %v, want %v",
					rec.Code, http.StatusOK,
				)
			}
			want := `{"records":[],"total_count":0}`
			if rec.Body.String() != want {
				t.Errorf(
					"ListByUserID(c) body = %v, want %v",
					rec.Body.String(), want,
				)
			}
		},
	)

	t.Run(
		"returns 400 when the score is not a number",
		func(t *testing.T) {
			usecase := &fakeUsecase{}
			recordHandler := NewRecordHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Request = httptest.NewRequest(
				http.MethodGet,
				"/?scores=invalid",
				nil,
			)

			recordHandler.ListByUserID(c)
			if rec.Code != http.StatusBadRequest {
				t.Errorf(
					"ListByUserID(c) code = %v, want %v",
					rec.Code, http.StatusBadRequest,
				)
			}
			want := `"code":"INVALID_INPUT"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"ListByUserID(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)

	t.Run(
		"returns 400 when the platform is invalid",
		func(t *testing.T) {
			usecase := &fakeUsecase{}
			recordHandler := NewRecordHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Request = httptest.NewRequest(
				http.MethodGet,
				"/?platforms=invalid",
				nil,
			)

			recordHandler.ListByUserID(c)
			if rec.Code != http.StatusBadRequest {
				t.Errorf(
					"ListByUserID(c) code = %v, want %v",
					rec.Code, http.StatusBadRequest,
				)
			}
			want := `"code":"INVALID_INPUT"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"ListByUserID(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)

	t.Run(
		"returns 400 when the mood tag is invalid",
		func(t *testing.T) {
			usecase := &fakeUsecase{}
			recordHandler := NewRecordHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Request = httptest.NewRequest(
				http.MethodGet,
				"/?moodTags=invalid",
				nil,
			)

			recordHandler.ListByUserID(c)
			if rec.Code != http.StatusBadRequest {
				t.Errorf(
					"ListByUserID(c) code = %v, want %v",
					rec.Code, http.StatusBadRequest,
				)
			}
			want := `"code":"INVALID_INPUT"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"ListByUserID(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)

	t.Run(
		"returns 400 when the genre is invalid",
		func(t *testing.T) {
			usecase := &fakeUsecase{}
			recordHandler := NewRecordHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Request = httptest.NewRequest(
				http.MethodGet,
				"/?genres=invalid",
				nil,
			)

			recordHandler.ListByUserID(c)
			if rec.Code != http.StatusBadRequest {
				t.Errorf(
					"ListByUserID(c) code = %v, want %v",
					rec.Code, http.StatusBadRequest,
				)
			}
			want := `"code":"INVALID_INPUT"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"ListByUserID(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)

	t.Run(
		"returns 400 when the title is invalid",
		func(t *testing.T) {
			usecase := &fakeUsecase{}
			recordHandler := NewRecordHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Request = httptest.NewRequest(
				http.MethodGet,
				"/?title="+strings.Repeat("a", 256),
				nil,
			)

			recordHandler.ListByUserID(c)
			if rec.Code != http.StatusBadRequest {
				t.Errorf(
					"ListByUserID(c) code = %v, want %v",
					rec.Code, http.StatusBadRequest,
				)
			}
			want := `"code":"INVALID_INPUT"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"ListByUserID(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)

	t.Run(
		"returns 400 when the sort field is invalid",
		func(t *testing.T) {
			usecase := &fakeUsecase{}
			recordHandler := NewRecordHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Request = httptest.NewRequest(
				http.MethodGet,
				"/?sortField=invalid",
				nil,
			)

			recordHandler.ListByUserID(c)
			if rec.Code != http.StatusBadRequest {
				t.Errorf(
					"ListByUserID(c) code = %v, want %v",
					rec.Code, http.StatusBadRequest,
				)
			}
			want := `"code":"INVALID_INPUT"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"ListByUserID(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)

	t.Run(
		"returns 400 when the sort order is invalid",
		func(t *testing.T) {
			usecase := &fakeUsecase{}
			recordHandler := NewRecordHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Request = httptest.NewRequest(
				http.MethodGet,
				"/?sortOrder=invalid",
				nil,
			)

			recordHandler.ListByUserID(c)
			if rec.Code != http.StatusBadRequest {
				t.Errorf(
					"ListByUserID(c) code = %v, want %v",
					rec.Code, http.StatusBadRequest,
				)
			}
			want := `"code":"INVALID_INPUT"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"ListByUserID(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)

	t.Run(
		"returns 400 when the page is not a number",
		func(t *testing.T) {
			usecase := &fakeUsecase{}
			recordHandler := NewRecordHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Request = httptest.NewRequest(
				http.MethodGet,
				"/?page=invalid",
				nil,
			)

			recordHandler.ListByUserID(c)
			if rec.Code != http.StatusBadRequest {
				t.Errorf(
					"ListByUserID(c) code = %v, want %v",
					rec.Code, http.StatusBadRequest,
				)
			}
			want := `"code":"INVALID_INPUT"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"ListByUserID(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)

	t.Run(
		"returns 400 when the per page is not a number",
		func(t *testing.T) {
			usecase := &fakeUsecase{}
			recordHandler := NewRecordHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Request = httptest.NewRequest(
				http.MethodGet,
				"/?perPage=invalid",
				nil,
			)

			recordHandler.ListByUserID(c)
			if rec.Code != http.StatusBadRequest {
				t.Errorf(
					"ListByUserID(c) code = %v, want %v",
					rec.Code, http.StatusBadRequest,
				)
			}
			want := `"code":"INVALID_INPUT"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"ListByUserID(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)

	t.Run(
		"returns 500 when the authenticated user ID is missing from the context",
		func(t *testing.T) {
			usecase := &fakeUsecase{}
			recordHandler := NewRecordHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

			recordHandler.ListByUserID(c)
			if rec.Code != http.StatusInternalServerError {
				t.Errorf(
					"ListByUserID(c) code = %v, want %v",
					rec.Code, http.StatusInternalServerError,
				)
			}
			want := `"code":"INTERNAL_SERVER_ERROR"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"ListByUserID(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)

	t.Run(
		"returns 500 when the usecase returns an unexpected error",
		func(t *testing.T) {
			usecase := &fakeUsecase{err: errors.New("unexpected")}
			recordHandler := NewRecordHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

			recordHandler.ListByUserID(c)
			if rec.Code != http.StatusInternalServerError {
				t.Errorf(
					"ListByUserID(c) code = %v, want %v",
					rec.Code, http.StatusInternalServerError,
				)
			}
			want := `"code":"INTERNAL_SERVER_ERROR"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"ListByUserID(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)
}

func TestUpdate(t *testing.T) {
	t.Run(
		"passes the converted values to the usecase and returns 200 when the request is valid",
		func(t *testing.T) {
			userID := userdomain.ID(1)
			recordID := recorddomain.ID(10)
			r := newTestRecord()
			usecase := &fakeUsecase{record: &r}
			recordHandler := NewRecordHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userID)
			c.Params = gin.Params{{Key: "id", Value: "10"}}
			c.Request = httptest.NewRequest(
				http.MethodPut, "/", strings.NewReader(validBody),
			)

			recordHandler.Update(c)
			if rec.Code != http.StatusOK {
				t.Errorf(
					"Update(c) code = %v, want %v",
					rec.Code, http.StatusOK,
				)
			}
			if rec.Body.String() != wantBody {
				t.Errorf(
					"Update(c) body = %v, want %v",
					rec.Body.String(), wantBody,
				)
			}

			if usecase.updatedUserID != userID || usecase.updatedRecordID != recordID {
				t.Errorf(
					"Update(c) usecase args = %v, %v, want %v, %v",
					usecase.updatedUserID, usecase.updatedRecordID, userID, recordID,
				)
			}
			if usecase.updatedRecord.Title != r.Title {
				t.Errorf(
					"Update(c) usecase record = %v, want %v",
					usecase.updatedRecord, r,
				)
			}
		},
	)

	t.Run(
		"returns 404 when the record does not exist",
		func(t *testing.T) {
			usecase := &fakeUsecase{err: exception.ErrNotFound}
			recordHandler := NewRecordHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Params = gin.Params{{Key: "id", Value: "10"}}
			c.Request = httptest.NewRequest(
				http.MethodPut, "/", strings.NewReader(validBody),
			)

			recordHandler.Update(c)
			if rec.Code != http.StatusNotFound {
				t.Errorf(
					"Update(c) code = %v, want %v",
					rec.Code, http.StatusNotFound,
				)
			}
			want := `"code":"RECORD_NOT_FOUND"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"Update(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)

	t.Run(
		"returns 400 when the request body is not valid JSON",
		func(t *testing.T) {
			usecase := &fakeUsecase{}
			recordHandler := NewRecordHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Params = gin.Params{{Key: "id", Value: "10"}}
			c.Request = httptest.NewRequest(
				http.MethodPut, "/", strings.NewReader(`{"title":"Test Movie",}`),
			)

			recordHandler.Update(c)
			if rec.Code != http.StatusBadRequest {
				t.Errorf(
					"Update(c) code = %v, want %v",
					rec.Code, http.StatusBadRequest,
				)
			}
			want := `"code":"INVALID_INPUT","message":"malformed request body"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"Update(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)

	t.Run(
		"returns 500 when the authenticated user ID is missing from the context",
		func(t *testing.T) {
			usecase := &fakeUsecase{}
			recordHandler := NewRecordHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Params = gin.Params{{Key: "id", Value: "10"}}
			c.Request = httptest.NewRequest(
				http.MethodPut, "/", strings.NewReader(validBody),
			)

			recordHandler.Update(c)
			if rec.Code != http.StatusInternalServerError {
				t.Errorf(
					"Update(c) code = %v, want %v",
					rec.Code, http.StatusInternalServerError,
				)
			}
			want := `"code":"INTERNAL_SERVER_ERROR"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"Update(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)
}

func TestDelete(t *testing.T) {
	t.Run(
		"passes the ids to the usecase and returns 204 when the request is valid",
		func(t *testing.T) {
			userID := userdomain.ID(1)
			recordID := recorddomain.ID(10)
			usecase := &fakeUsecase{}
			recordHandler := NewRecordHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userID)
			c.Params = gin.Params{{Key: "id", Value: "10"}}
			c.Request = httptest.NewRequest(http.MethodDelete, "/", nil)

			recordHandler.Delete(c)
			c.Writer.WriteHeaderNow()
			if rec.Code != http.StatusNoContent {
				t.Errorf(
					"Delete(c) code = %v, want %v",
					rec.Code, http.StatusNoContent,
				)
			}
			want := ``
			if rec.Body.String() != want {
				t.Errorf(
					"Delete(c) body = %v, want %v",
					rec.Body.String(), want,
				)
			}

			if usecase.deletedUserID != userID || usecase.deletedRecordID != recordID {
				t.Errorf(
					"Delete(c) usecase args = %v, %v, want %v, %v",
					usecase.deletedUserID, usecase.deletedRecordID, userID, recordID,
				)
			}
		},
	)

	t.Run(
		"returns 404 when the record does not exist",
		func(t *testing.T) {
			usecase := &fakeUsecase{err: exception.ErrNotFound}
			recordHandler := NewRecordHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Params = gin.Params{{Key: "id", Value: "10"}}
			c.Request = httptest.NewRequest(http.MethodDelete, "/", nil)

			recordHandler.Delete(c)
			if rec.Code != http.StatusNotFound {
				t.Errorf(
					"Delete(c) code = %v, want %v",
					rec.Code, http.StatusNotFound,
				)
			}
			want := `"code":"RECORD_NOT_FOUND"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"Delete(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)

	t.Run(
		"returns 500 when the authenticated user ID is missing from the context",
		func(t *testing.T) {
			usecase := &fakeUsecase{}
			recordHandler := NewRecordHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Params = gin.Params{{Key: "id", Value: "10"}}
			c.Request = httptest.NewRequest(http.MethodDelete, "/", nil)

			recordHandler.Delete(c)
			if rec.Code != http.StatusInternalServerError {
				t.Errorf(
					"Delete(c) code = %v, want %v",
					rec.Code, http.StatusInternalServerError,
				)
			}
			want := `"code":"INTERNAL_SERVER_ERROR"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"Delete(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)
}
