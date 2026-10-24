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

type fakeArticleUsecase struct {
	detail           *recorddomain.ArticleDetail
	listResult       recorddomain.ArticleListResult
	listedQuery      recorddomain.ArticleQuery
	createdUserID    userdomain.ID
	createdArticle   recorddomain.Article
	updatedUserID    userdomain.ID
	updatedArticleID recorddomain.ArticleID
	updatedArticle   recorddomain.Article
	deletedUserID    userdomain.ID
	deletedArticleID recorddomain.ArticleID
	err              error
}

func (u *fakeArticleUsecase) GetByID(
	ctx context.Context,
	userID userdomain.ID,
	articleID recorddomain.ArticleID,
) (*recorddomain.ArticleDetail, error) {
	return u.detail, u.err
}

func (u *fakeArticleUsecase) ListByUserID(
	ctx context.Context,
	userID userdomain.ID,
	query recorddomain.ArticleQuery,
) (recorddomain.ArticleListResult, error) {
	u.listedQuery = query

	return u.listResult, u.err
}

func (u *fakeArticleUsecase) Create(
	ctx context.Context,
	userID userdomain.ID,
	a recorddomain.Article,
) (*recorddomain.ArticleDetail, error) {
	u.createdUserID = userID
	u.createdArticle = a

	return u.detail, u.err
}

func (u *fakeArticleUsecase) Update(
	ctx context.Context,
	userID userdomain.ID,
	articleID recorddomain.ArticleID,
	a recorddomain.Article,
) (*recorddomain.ArticleDetail, error) {
	u.updatedUserID = userID
	u.updatedArticleID = articleID
	u.updatedArticle = a

	return u.detail, u.err
}

func (u *fakeArticleUsecase) Delete(
	ctx context.Context,
	userID userdomain.ID,
	articleID recorddomain.ArticleID,
) error {
	u.deletedUserID = userID
	u.deletedArticleID = articleID

	return u.err
}

const validCreateArticleBody = `{
	"record_id":"20",
	"title":"Test Article",
	"content":{"type":"doc","content":[{"type":"paragraph"}]}
}`

const validUpdateArticleBody = `{
	"title":"Test Article",
	"content":{"type":"doc","content":[{"type":"paragraph"}]}
}`

const wantArticleRecordBody = `{"poster_url":"https://example.com/poster.jpg","record_id":"20",` +
	`"release_year":2020,"title":"Test Movie"}`

const wantArticleBody = `{"article_id":"10",` +
	`"content":{"type":"doc","content":[{"type":"paragraph"}]},` +
	`"created_at":"2026-01-02T00:00:00Z","record":` + wantArticleRecordBody + `,` +
	`"title":"Test Article","updated_at":"2026-01-03T00:00:00Z"}`

const wantArticleSummaryBody = `{"article_id":"10","created_at":"2026-01-02T00:00:00Z",` +
	`"record":` + wantArticleRecordBody + `,` +
	`"title":"Test Article","updated_at":"2026-01-03T00:00:00Z"}`

func newTestArticleRecordSummary() recorddomain.ArticleRecordSummary {
	return recorddomain.ArticleRecordSummary{
		ID:          recorddomain.ID(20),
		Title:       recorddomain.Title("Test Movie"),
		ReleaseYear: recorddomain.ReleaseYear(2020),
		PosterURL:   recorddomain.PosterURL("https://example.com/poster.jpg"),
	}
}

func newTestArticleDetail(a recorddomain.Article) *recorddomain.ArticleDetail {
	return &recorddomain.ArticleDetail{
		Article: a,
		Record:  newTestArticleRecordSummary(),
	}
}

func newTestArticle() recorddomain.Article {
	return recorddomain.Article{
		ID:        recorddomain.ArticleID(10),
		UserID:    userdomain.ID(1),
		RecordID:  recorddomain.ID(20),
		Title:     recorddomain.ArticleTitle("Test Article"),
		Content:   recorddomain.ArticleContent(`{"type":"doc","content":[{"type":"paragraph"}]}`),
		CreatedAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC),
	}
}

func TestArticleCreate(t *testing.T) {
	t.Run(
		"passes the converted values to the usecase and returns 201 when the request is valid",
		func(t *testing.T) {
			userID := userdomain.ID(1)
			a := newTestArticle()
			usecase := &fakeArticleUsecase{detail: newTestArticleDetail(a)}
			articleHandler := NewArticleHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userID)
			c.Request = httptest.NewRequest(
				http.MethodPost, "/", strings.NewReader(validCreateArticleBody),
			)

			articleHandler.Create(c)
			if rec.Code != http.StatusCreated {
				t.Errorf(
					"Create(c) code = %v, want %v",
					rec.Code, http.StatusCreated,
				)
			}
			if rec.Body.String() != wantArticleBody {
				t.Errorf(
					"Create(c) body = %v, want %v",
					rec.Body.String(), wantArticleBody,
				)
			}

			if usecase.createdUserID != userID {
				t.Errorf(
					"Create(c) usecase user id = %v, want %v",
					usecase.createdUserID, userID,
				)
			}
			if usecase.createdArticle.RecordID != a.RecordID ||
				usecase.createdArticle.Title != a.Title ||
				string(usecase.createdArticle.Content) != string(a.Content) {
				t.Errorf(
					"Create(c) usecase article = %v, want %v",
					usecase.createdArticle, a,
				)
			}
		},
	)

	t.Run(
		"returns 500 when the authenticated user ID is missing from the context",
		func(t *testing.T) {
			usecase := &fakeArticleUsecase{}
			articleHandler := NewArticleHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(
				http.MethodPost, "/", strings.NewReader(validCreateArticleBody),
			)

			articleHandler.Create(c)
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
			usecase := &fakeArticleUsecase{}
			articleHandler := NewArticleHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Request = httptest.NewRequest(
				http.MethodPost, "/", strings.NewReader(`{"title":"Test Article",}`),
			)

			articleHandler.Create(c)
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
		"returns 400 when the record ID is missing",
		func(t *testing.T) {
			usecase := &fakeArticleUsecase{}
			articleHandler := NewArticleHandler(usecase)

			body := strings.Replace(validCreateArticleBody, `"record_id":"20",`, "", 1)
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Request = httptest.NewRequest(
				http.MethodPost, "/", strings.NewReader(body),
			)

			articleHandler.Create(c)
			if rec.Code != http.StatusBadRequest {
				t.Errorf(
					"Create(c) code = %v, want %v",
					rec.Code, http.StatusBadRequest,
				)
			}
			want := `"code":"INVALID_INPUT","message":"invalid: record id is required"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"Create(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)

	t.Run(
		"returns 400 when the record ID is not a number",
		func(t *testing.T) {
			usecase := &fakeArticleUsecase{}
			articleHandler := NewArticleHandler(usecase)

			body := strings.Replace(validCreateArticleBody, `"record_id":"20"`, `"record_id":"abc"`, 1)
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Request = httptest.NewRequest(
				http.MethodPost, "/", strings.NewReader(body),
			)

			articleHandler.Create(c)
			if rec.Code != http.StatusBadRequest {
				t.Errorf(
					"Create(c) code = %v, want %v",
					rec.Code, http.StatusBadRequest,
				)
			}
			want := `"code":"INVALID_INPUT","message":"invalid: record id must be a number"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"Create(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)

	t.Run(
		"returns 400 when the title is empty",
		func(t *testing.T) {
			usecase := &fakeArticleUsecase{}
			articleHandler := NewArticleHandler(usecase)

			body := strings.Replace(validCreateArticleBody, `"Test Article"`, `""`, 1)
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Request = httptest.NewRequest(
				http.MethodPost, "/", strings.NewReader(body),
			)

			articleHandler.Create(c)
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
		"returns 400 when the content is not a document",
		func(t *testing.T) {
			usecase := &fakeArticleUsecase{}
			articleHandler := NewArticleHandler(usecase)

			body := strings.Replace(
				validCreateArticleBody,
				`{"type":"doc","content":[{"type":"paragraph"}]}`,
				`[1,2]`,
				1,
			)
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Request = httptest.NewRequest(
				http.MethodPost, "/", strings.NewReader(body),
			)

			articleHandler.Create(c)
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
		"returns 404 when the record does not exist",
		func(t *testing.T) {
			usecase := &fakeArticleUsecase{err: exception.ErrNotFound}
			articleHandler := NewArticleHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Request = httptest.NewRequest(
				http.MethodPost, "/", strings.NewReader(validCreateArticleBody),
			)

			articleHandler.Create(c)
			if rec.Code != http.StatusNotFound {
				t.Errorf(
					"Create(c) code = %v, want %v",
					rec.Code, http.StatusNotFound,
				)
			}
			want := `"code":"RECORD_NOT_FOUND"`
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
			usecase := &fakeArticleUsecase{err: errors.New("unexpected")}
			articleHandler := NewArticleHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Request = httptest.NewRequest(
				http.MethodPost, "/", strings.NewReader(validCreateArticleBody),
			)

			articleHandler.Create(c)
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

func TestArticleGetByID(t *testing.T) {
	t.Run(
		"returns the article and 200 when the request is valid",
		func(t *testing.T) {
			a := newTestArticle()
			usecase := &fakeArticleUsecase{detail: newTestArticleDetail(a)}
			articleHandler := NewArticleHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Params = gin.Params{{Key: "id", Value: "10"}}
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

			articleHandler.GetByID(c)
			if rec.Code != http.StatusOK {
				t.Errorf(
					"GetByID(c) code = %v, want %v",
					rec.Code, http.StatusOK,
				)
			}
			if rec.Body.String() != wantArticleBody {
				t.Errorf(
					"GetByID(c) body = %v, want %v",
					rec.Body.String(), wantArticleBody,
				)
			}
		},
	)

	t.Run(
		"returns 400 when the article ID is not a number",
		func(t *testing.T) {
			usecase := &fakeArticleUsecase{}
			articleHandler := NewArticleHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Params = gin.Params{{Key: "id", Value: "invalid"}}
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

			articleHandler.GetByID(c)
			if rec.Code != http.StatusBadRequest {
				t.Errorf(
					"GetByID(c) code = %v, want %v",
					rec.Code, http.StatusBadRequest,
				)
			}
			want := `"code":"INVALID_INPUT","message":"invalid: article id must be a number"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"GetByID(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)

	t.Run(
		"returns 404 when the article does not exist",
		func(t *testing.T) {
			usecase := &fakeArticleUsecase{err: exception.ErrNotFound}
			articleHandler := NewArticleHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Params = gin.Params{{Key: "id", Value: "10"}}
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

			articleHandler.GetByID(c)
			if rec.Code != http.StatusNotFound {
				t.Errorf(
					"GetByID(c) code = %v, want %v",
					rec.Code, http.StatusNotFound,
				)
			}
			want := `"code":"ARTICLE_NOT_FOUND"`
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
			usecase := &fakeArticleUsecase{err: errors.New("unexpected")}
			articleHandler := NewArticleHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Params = gin.Params{{Key: "id", Value: "10"}}
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

			articleHandler.GetByID(c)
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

func TestArticleListByUserID(t *testing.T) {
	t.Run(
		"passes the converted query to the usecase and returns the summaries and 200 when the request is valid",
		func(t *testing.T) {
			usecase := &fakeArticleUsecase{
				listResult: recorddomain.ArticleListResult{
					Articles: []recorddomain.ArticleSummary{
						{
							ID:        recorddomain.ArticleID(10),
							Record:    newTestArticleRecordSummary(),
							Title:     recorddomain.ArticleTitle("Test Article"),
							CreatedAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
							UpdatedAt: time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC),
						},
					},
					FilteredCount: recorddomain.FilteredCount(1),
					TotalCount:    recorddomain.TotalCount(5),
				},
			}
			articleHandler := NewArticleHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Request = httptest.NewRequest(
				http.MethodGet,
				"/?record_id=20&keyword=感想&sort_field=title&sort_order=asc&page=2&per_page=10",
				nil,
			)

			articleHandler.ListByUserID(c)
			if rec.Code != http.StatusOK {
				t.Errorf(
					"ListByUserID(c) code = %v, want %v",
					rec.Code, http.StatusOK,
				)
			}
			want := `{"articles":[` + wantArticleSummaryBody + `],"filtered_count":1,"total_count":5}`
			if rec.Body.String() != want {
				t.Errorf(
					"ListByUserID(c) body = %v, want %v",
					rec.Body.String(), want,
				)
			}

			wantQuery := recorddomain.NewArticleQuery(
				recorddomain.ID(20),
				recorddomain.ArticleKeyword("感想"),
				recorddomain.ArticleSortFieldTitle,
				recorddomain.SortOrderAsc,
				recorddomain.Page(2),
				recorddomain.PerPage(10),
			)
			if usecase.listedQuery != wantQuery {
				t.Errorf(
					"ListByUserID(c) query = %v, want %v",
					usecase.listedQuery, wantQuery,
				)
			}
		},
	)

	t.Run(
		"passes the default query to the usecase when the query parameters are omitted",
		func(t *testing.T) {
			usecase := &fakeArticleUsecase{}
			articleHandler := NewArticleHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

			articleHandler.ListByUserID(c)
			if rec.Code != http.StatusOK {
				t.Errorf(
					"ListByUserID(c) code = %v, want %v",
					rec.Code, http.StatusOK,
				)
			}

			wantQuery := recorddomain.NewArticleQuery(
				0,
				"",
				recorddomain.ArticleSortFieldCreatedAt,
				recorddomain.SortOrderDesc,
				recorddomain.Page(1),
				recorddomain.PerPage(20),
			)
			if usecase.listedQuery != wantQuery {
				t.Errorf(
					"ListByUserID(c) query = %v, want %v",
					usecase.listedQuery, wantQuery,
				)
			}
		},
	)

	t.Run(
		"returns an empty array when the user has no articles",
		func(t *testing.T) {
			usecase := &fakeArticleUsecase{}
			articleHandler := NewArticleHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

			articleHandler.ListByUserID(c)
			want := `{"articles":[],"filtered_count":0,"total_count":0}`
			if rec.Body.String() != want {
				t.Errorf(
					"ListByUserID(c) body = %v, want %v",
					rec.Body.String(), want,
				)
			}
		},
	)

	t.Run(
		"returns 400 when the record ID is not a number",
		func(t *testing.T) {
			usecase := &fakeArticleUsecase{}
			articleHandler := NewArticleHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Request = httptest.NewRequest(http.MethodGet, "/?record_id=invalid", nil)

			articleHandler.ListByUserID(c)
			if rec.Code != http.StatusBadRequest {
				t.Errorf(
					"ListByUserID(c) code = %v, want %v",
					rec.Code, http.StatusBadRequest,
				)
			}
			want := `"code":"INVALID_INPUT","message":"invalid: record id must be a number"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"ListByUserID(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)

	t.Run(
		"returns 400 when the record ID is 0",
		func(t *testing.T) {
			usecase := &fakeArticleUsecase{}
			articleHandler := NewArticleHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Request = httptest.NewRequest(http.MethodGet, "/?record_id=0", nil)

			articleHandler.ListByUserID(c)
			if rec.Code != http.StatusBadRequest {
				t.Errorf(
					"ListByUserID(c) code = %v, want %v",
					rec.Code, http.StatusBadRequest,
				)
			}
			want := `"code":"INVALID_INPUT","message":"invalid: record id must be greater than 0"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"ListByUserID(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)

	t.Run(
		"returns 400 when the keyword is too long",
		func(t *testing.T) {
			usecase := &fakeArticleUsecase{}
			articleHandler := NewArticleHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Request = httptest.NewRequest(
				http.MethodGet, "/?keyword="+strings.Repeat("a", 256), nil,
			)

			articleHandler.ListByUserID(c)
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
			usecase := &fakeArticleUsecase{}
			articleHandler := NewArticleHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Request = httptest.NewRequest(http.MethodGet, "/?sort_field=invalid", nil)

			articleHandler.ListByUserID(c)
			if rec.Code != http.StatusBadRequest {
				t.Errorf(
					"ListByUserID(c) code = %v, want %v",
					rec.Code, http.StatusBadRequest,
				)
			}
			want := `"code":"INVALID_INPUT","message":"invalid: invalid article sort field"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"ListByUserID(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)

	t.Run(
		"returns 400 when the page is less than 1",
		func(t *testing.T) {
			usecase := &fakeArticleUsecase{}
			articleHandler := NewArticleHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Request = httptest.NewRequest(http.MethodGet, "/?page=0", nil)

			articleHandler.ListByUserID(c)
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
		"returns 500 when the usecase returns an unexpected error",
		func(t *testing.T) {
			usecase := &fakeArticleUsecase{err: errors.New("unexpected")}
			articleHandler := NewArticleHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

			articleHandler.ListByUserID(c)
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

func TestArticleUpdate(t *testing.T) {
	t.Run(
		"passes the converted values to the usecase and returns 200 when the request is valid",
		func(t *testing.T) {
			userID := userdomain.ID(1)
			a := newTestArticle()
			usecase := &fakeArticleUsecase{detail: newTestArticleDetail(a)}
			articleHandler := NewArticleHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userID)
			c.Params = gin.Params{{Key: "id", Value: "10"}}
			c.Request = httptest.NewRequest(
				http.MethodPut, "/", strings.NewReader(validUpdateArticleBody),
			)

			articleHandler.Update(c)
			if rec.Code != http.StatusOK {
				t.Errorf(
					"Update(c) code = %v, want %v",
					rec.Code, http.StatusOK,
				)
			}
			if rec.Body.String() != wantArticleBody {
				t.Errorf(
					"Update(c) body = %v, want %v",
					rec.Body.String(), wantArticleBody,
				)
			}

			if usecase.updatedUserID != userID {
				t.Errorf(
					"Update(c) usecase user id = %v, want %v",
					usecase.updatedUserID, userID,
				)
			}
			if usecase.updatedArticleID != a.ID {
				t.Errorf(
					"Update(c) usecase article id = %v, want %v",
					usecase.updatedArticleID, a.ID,
				)
			}
			if usecase.updatedArticle.Title != a.Title ||
				string(usecase.updatedArticle.Content) != string(a.Content) {
				t.Errorf(
					"Update(c) usecase article = %v, want %v",
					usecase.updatedArticle, a,
				)
			}
		},
	)

	t.Run(
		"returns 400 when the article ID is not a number",
		func(t *testing.T) {
			usecase := &fakeArticleUsecase{}
			articleHandler := NewArticleHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Params = gin.Params{{Key: "id", Value: "invalid"}}
			c.Request = httptest.NewRequest(
				http.MethodPut, "/", strings.NewReader(validUpdateArticleBody),
			)

			articleHandler.Update(c)
			if rec.Code != http.StatusBadRequest {
				t.Errorf(
					"Update(c) code = %v, want %v",
					rec.Code, http.StatusBadRequest,
				)
			}
		},
	)

	t.Run(
		"returns 400 when the content is not a document",
		func(t *testing.T) {
			usecase := &fakeArticleUsecase{}
			articleHandler := NewArticleHandler(usecase)

			body := strings.Replace(
				validUpdateArticleBody,
				`{"type":"doc","content":[{"type":"paragraph"}]}`,
				`{"type":"paragraph"}`,
				1,
			)
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Params = gin.Params{{Key: "id", Value: "10"}}
			c.Request = httptest.NewRequest(
				http.MethodPut, "/", strings.NewReader(body),
			)

			articleHandler.Update(c)
			if rec.Code != http.StatusBadRequest {
				t.Errorf(
					"Update(c) code = %v, want %v",
					rec.Code, http.StatusBadRequest,
				)
			}
			want := `"code":"INVALID_INPUT"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"Update(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)

	t.Run(
		"returns 404 when the article does not exist",
		func(t *testing.T) {
			usecase := &fakeArticleUsecase{err: exception.ErrNotFound}
			articleHandler := NewArticleHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Params = gin.Params{{Key: "id", Value: "10"}}
			c.Request = httptest.NewRequest(
				http.MethodPut, "/", strings.NewReader(validUpdateArticleBody),
			)

			articleHandler.Update(c)
			if rec.Code != http.StatusNotFound {
				t.Errorf(
					"Update(c) code = %v, want %v",
					rec.Code, http.StatusNotFound,
				)
			}
			want := `"code":"ARTICLE_NOT_FOUND"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"Update(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)

	t.Run(
		"returns 500 when the usecase returns an unexpected error",
		func(t *testing.T) {
			usecase := &fakeArticleUsecase{err: errors.New("unexpected")}
			articleHandler := NewArticleHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Params = gin.Params{{Key: "id", Value: "10"}}
			c.Request = httptest.NewRequest(
				http.MethodPut, "/", strings.NewReader(validUpdateArticleBody),
			)

			articleHandler.Update(c)
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

func TestArticleDelete(t *testing.T) {
	t.Run(
		"passes the IDs to the usecase and returns 204 when the request is valid",
		func(t *testing.T) {
			userID := userdomain.ID(1)
			usecase := &fakeArticleUsecase{}
			articleHandler := NewArticleHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userID)
			c.Params = gin.Params{{Key: "id", Value: "10"}}
			c.Request = httptest.NewRequest(http.MethodDelete, "/", nil)

			articleHandler.Delete(c)
			c.Writer.WriteHeaderNow()
			if rec.Code != http.StatusNoContent {
				t.Errorf(
					"Delete(c) code = %v, want %v",
					rec.Code, http.StatusNoContent,
				)
			}
			if usecase.deletedUserID != userID || usecase.deletedArticleID != recorddomain.ArticleID(10) {
				t.Errorf(
					"Delete(c) usecase user id, article id = %v, %v, want %v, %v",
					usecase.deletedUserID, usecase.deletedArticleID, userID, recorddomain.ArticleID(10),
				)
			}
		},
	)

	t.Run(
		"returns 400 when the article ID is not a number",
		func(t *testing.T) {
			usecase := &fakeArticleUsecase{}
			articleHandler := NewArticleHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Params = gin.Params{{Key: "id", Value: "invalid"}}
			c.Request = httptest.NewRequest(http.MethodDelete, "/", nil)

			articleHandler.Delete(c)
			if rec.Code != http.StatusBadRequest {
				t.Errorf(
					"Delete(c) code = %v, want %v",
					rec.Code, http.StatusBadRequest,
				)
			}
		},
	)

	t.Run(
		"returns 404 when the article does not exist",
		func(t *testing.T) {
			usecase := &fakeArticleUsecase{err: exception.ErrNotFound}
			articleHandler := NewArticleHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Params = gin.Params{{Key: "id", Value: "10"}}
			c.Request = httptest.NewRequest(http.MethodDelete, "/", nil)

			articleHandler.Delete(c)
			if rec.Code != http.StatusNotFound {
				t.Errorf(
					"Delete(c) code = %v, want %v",
					rec.Code, http.StatusNotFound,
				)
			}
			want := `"code":"ARTICLE_NOT_FOUND"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"Delete(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)

	t.Run(
		"returns 500 when the usecase returns an unexpected error",
		func(t *testing.T) {
			usecase := &fakeArticleUsecase{err: errors.New("unexpected")}
			articleHandler := NewArticleHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Params = gin.Params{{Key: "id", Value: "10"}}
			c.Request = httptest.NewRequest(http.MethodDelete, "/", nil)

			articleHandler.Delete(c)
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
