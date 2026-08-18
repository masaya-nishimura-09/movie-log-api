package record

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
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
	records         []*recorddomain.Record
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
) ([]*recorddomain.Record, error) {
	return u.records, u.err
}

func (u *fakeUsecase) CreateRecord(
	ctx context.Context,
	userID userdomain.ID,
	r recorddomain.Record,
) (*recorddomain.Record, error) {
	u.createdUserID = userID
	u.createdRecord = r

	return u.record, u.err
}

func (u *fakeUsecase) UpdateRecord(
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

func (u *fakeUsecase) DeleteRecord(
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

func TestCreateRecord(t *testing.T) {
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

			recordHandler.CreateRecord(c)
			if rec.Code != http.StatusCreated {
				t.Errorf(
					"CreateRecord(c) code = %v, want %v",
					rec.Code, http.StatusCreated,
				)
			}
			if rec.Body.String() != wantBody {
				t.Errorf(
					"CreateRecord(c) body = %v, want %v",
					rec.Body.String(), wantBody,
				)
			}

			if usecase.createdUserID != userID {
				t.Errorf(
					"CreateRecord(c) usecase user id = %v, want %v",
					usecase.createdUserID, userID,
				)
			}
			if usecase.createdRecord.Title != r.Title ||
				usecase.createdRecord.Platform != r.Platform ||
				usecase.createdRecord.Score != r.Score {
				t.Errorf(
					"CreateRecord(c) usecase record = %v, want %v",
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

			recordHandler.CreateRecord(c)
			if rec.Code != http.StatusInternalServerError {
				t.Errorf(
					"CreateRecord(c) code = %v, want %v",
					rec.Code, http.StatusInternalServerError,
				)
			}
			want := `"code":"INTERNAL_SERVER_ERROR"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"CreateRecord(c) body = %v, want to contain %v",
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

			recordHandler.CreateRecord(c)
			if rec.Code != http.StatusBadRequest {
				t.Errorf(
					"CreateRecord(c) code = %v, want %v",
					rec.Code, http.StatusBadRequest,
				)
			}
			want := `"code":"INVALID_INPUT","message":"malformed request body"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"CreateRecord(c) body = %v, want to contain %v",
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

			recordHandler.CreateRecord(c)
			if rec.Code != http.StatusBadRequest {
				t.Errorf(
					"CreateRecord(c) code = %v, want %v",
					rec.Code, http.StatusBadRequest,
				)
			}
			want := `"code":"INVALID_INPUT"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"CreateRecord(c) body = %v, want to contain %v",
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

			recordHandler.CreateRecord(c)
			if rec.Code != http.StatusInternalServerError {
				t.Errorf(
					"CreateRecord(c) code = %v, want %v",
					rec.Code, http.StatusInternalServerError,
				)
			}
			want := `"code":"INTERNAL_SERVER_ERROR"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"CreateRecord(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)
}

func TestGetRecord(t *testing.T) {
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

			recordHandler.GetRecord(c)
			if rec.Code != http.StatusOK {
				t.Errorf(
					"GetRecord(c) code = %v, want %v",
					rec.Code, http.StatusOK,
				)
			}
			if rec.Body.String() != wantBody {
				t.Errorf(
					"GetRecord(c) body = %v, want %v",
					rec.Body.String(), wantBody,
				)
			}
		},
	)

	t.Run(
		"returns 404 when the path parameter is not a number",
		func(t *testing.T) {
			usecase := &fakeUsecase{}
			recordHandler := NewRecordHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Params = gin.Params{{Key: "id", Value: "invalid"}}
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

			recordHandler.GetRecord(c)
			if rec.Code != http.StatusNotFound {
				t.Errorf(
					"GetRecord(c) code = %v, want %v",
					rec.Code, http.StatusNotFound,
				)
			}
			want := `"code":"RECORD_NOT_FOUND"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"GetRecord(c) body = %v, want to contain %v",
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

			recordHandler.GetRecord(c)
			if rec.Code != http.StatusNotFound {
				t.Errorf(
					"GetRecord(c) code = %v, want %v",
					rec.Code, http.StatusNotFound,
				)
			}
			want := `"code":"RECORD_NOT_FOUND"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"GetRecord(c) body = %v, want to contain %v",
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

			recordHandler.GetRecord(c)
			if rec.Code != http.StatusInternalServerError {
				t.Errorf(
					"GetRecord(c) code = %v, want %v",
					rec.Code, http.StatusInternalServerError,
				)
			}
			want := `"code":"INTERNAL_SERVER_ERROR"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"GetRecord(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)
}

func TestListRecords(t *testing.T) {
	t.Run(
		"returns the records of the authenticated user and 200",
		func(t *testing.T) {
			r := newTestRecord()
			usecase := &fakeUsecase{records: []*recorddomain.Record{&r}}
			recordHandler := NewRecordHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

			recordHandler.ListRecords(c)
			if rec.Code != http.StatusOK {
				t.Errorf(
					"ListRecords(c) code = %v, want %v",
					rec.Code, http.StatusOK,
				)
			}
			want := `{"records":[` + wantBody + `]}`
			if rec.Body.String() != want {
				t.Errorf(
					"ListRecords(c) body = %v, want %v",
					rec.Body.String(), want,
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

			recordHandler.ListRecords(c)
			if rec.Code != http.StatusOK {
				t.Errorf(
					"ListRecords(c) code = %v, want %v",
					rec.Code, http.StatusOK,
				)
			}
			want := `{"records":[]}`
			if rec.Body.String() != want {
				t.Errorf(
					"ListRecords(c) body = %v, want %v",
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

			recordHandler.ListRecords(c)
			if rec.Code != http.StatusInternalServerError {
				t.Errorf(
					"ListRecords(c) code = %v, want %v",
					rec.Code, http.StatusInternalServerError,
				)
			}
			want := `"code":"INTERNAL_SERVER_ERROR"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"ListRecords(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)
}

func TestUpdateRecord(t *testing.T) {
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

			recordHandler.UpdateRecord(c)
			if rec.Code != http.StatusOK {
				t.Errorf(
					"UpdateRecord(c) code = %v, want %v",
					rec.Code, http.StatusOK,
				)
			}
			if rec.Body.String() != wantBody {
				t.Errorf(
					"UpdateRecord(c) body = %v, want %v",
					rec.Body.String(), wantBody,
				)
			}

			if usecase.updatedUserID != userID || usecase.updatedRecordID != recordID {
				t.Errorf(
					"UpdateRecord(c) usecase args = %v, %v, want %v, %v",
					usecase.updatedUserID, usecase.updatedRecordID, userID, recordID,
				)
			}
			if usecase.updatedRecord.Title != r.Title {
				t.Errorf(
					"UpdateRecord(c) usecase record = %v, want %v",
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

			recordHandler.UpdateRecord(c)
			if rec.Code != http.StatusNotFound {
				t.Errorf(
					"UpdateRecord(c) code = %v, want %v",
					rec.Code, http.StatusNotFound,
				)
			}
			want := `"code":"RECORD_NOT_FOUND"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"UpdateRecord(c) body = %v, want to contain %v",
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

			recordHandler.UpdateRecord(c)
			if rec.Code != http.StatusBadRequest {
				t.Errorf(
					"UpdateRecord(c) code = %v, want %v",
					rec.Code, http.StatusBadRequest,
				)
			}
			want := `"code":"INVALID_INPUT","message":"malformed request body"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"UpdateRecord(c) body = %v, want to contain %v",
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

			recordHandler.UpdateRecord(c)
			if rec.Code != http.StatusInternalServerError {
				t.Errorf(
					"UpdateRecord(c) code = %v, want %v",
					rec.Code, http.StatusInternalServerError,
				)
			}
			want := `"code":"INTERNAL_SERVER_ERROR"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"UpdateRecord(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)
}

func TestDeleteRecord(t *testing.T) {
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

			recordHandler.DeleteRecord(c)
			c.Writer.WriteHeaderNow()
			if rec.Code != http.StatusNoContent {
				t.Errorf(
					"DeleteRecord(c) code = %v, want %v",
					rec.Code, http.StatusNoContent,
				)
			}
			want := ``
			if rec.Body.String() != want {
				t.Errorf(
					"DeleteRecord(c) body = %v, want %v",
					rec.Body.String(), want,
				)
			}

			if usecase.deletedUserID != userID || usecase.deletedRecordID != recordID {
				t.Errorf(
					"DeleteRecord(c) usecase args = %v, %v, want %v, %v",
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

			recordHandler.DeleteRecord(c)
			if rec.Code != http.StatusNotFound {
				t.Errorf(
					"DeleteRecord(c) code = %v, want %v",
					rec.Code, http.StatusNotFound,
				)
			}
			want := `"code":"RECORD_NOT_FOUND"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"DeleteRecord(c) body = %v, want to contain %v",
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

			recordHandler.DeleteRecord(c)
			if rec.Code != http.StatusInternalServerError {
				t.Errorf(
					"DeleteRecord(c) code = %v, want %v",
					rec.Code, http.StatusInternalServerError,
				)
			}
			want := `"code":"INTERNAL_SERVER_ERROR"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"DeleteRecord(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)
}
