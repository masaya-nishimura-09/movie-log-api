package record

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/record"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
)

type fakeArticleRepository struct {
	detail      *record.ArticleDetail
	listResult  record.ArticleListResult
	listUserID  user.ID
	listQuery   record.ArticleQuery
	created     *record.Article
	updated     *record.Article
	deletedID   record.ArticleID
	deleteCalls int
}

func (r *fakeArticleRepository) GetByID(
	ctx context.Context,
	articleID record.ArticleID,
) (*record.ArticleDetail, error) {
	return r.detail, nil
}

func (r *fakeArticleRepository) ListByUserID(
	ctx context.Context,
	userID user.ID,
	query record.ArticleQuery,
) (record.ArticleListResult, error) {
	r.listUserID = userID
	r.listQuery = query
	return r.listResult, nil
}

func (r *fakeArticleRepository) Create(ctx context.Context, a *record.Article) error {
	r.created = a
	return nil
}

func (r *fakeArticleRepository) Update(ctx context.Context, a *record.Article) error {
	r.updated = a
	return nil
}

func (r *fakeArticleRepository) Delete(ctx context.Context, articleID record.ArticleID) error {
	r.deletedID = articleID
	r.deleteCalls++
	return nil
}

func TestArticleGetByID(t *testing.T) {
	userID := user.ID(1)
	articleID := record.ArticleID(10)

	t.Run(
		"returns the article when it belongs to the user",
		func(t *testing.T) {
			articleRepo := &fakeArticleRepository{
				detail: &record.ArticleDetail{
					Article: record.Article{ID: articleID, UserID: userID},
				},
			}
			au := NewArticleUsecase(articleRepo, &fakeRepository{})

			ctx := context.Background()

			got, err := au.GetByID(ctx, userID, articleID)
			if err != nil {
				t.Fatalf(
					"GetByID(ctx, %v, %v) (*record.ArticleDetail, error) = %v, %v",
					userID, articleID, got, err,
				)
			}
			if got.Article.ID != articleID {
				t.Errorf(
					"GetByID(ctx, %v, %v) ID = %v, want %v",
					userID, articleID, got.Article.ID, articleID,
				)
			}
		},
	)

	t.Run(
		"returns ErrNotFound when the article belongs to another user",
		func(t *testing.T) {
			articleRepo := &fakeArticleRepository{
				detail: &record.ArticleDetail{
					Article: record.Article{ID: articleID, UserID: user.ID(2)},
				},
			}
			au := NewArticleUsecase(articleRepo, &fakeRepository{})

			ctx := context.Background()

			got, err := au.GetByID(ctx, userID, articleID)
			if !errors.Is(err, exception.ErrNotFound) {
				t.Fatalf(
					"GetByID(ctx, %v, %v) (*record.ArticleDetail, error) = %v, %v, want %v",
					userID, articleID, got, err, exception.ErrNotFound,
				)
			}
			if got != nil {
				t.Errorf(
					"GetByID(ctx, %v, %v) (*record.ArticleDetail, error) = %v, %v, want nil",
					userID, articleID, got, err,
				)
			}
		},
	)
}

func TestArticleListByUserID(t *testing.T) {
	t.Run(
		"passes the user ID and the query to the repository and returns its result",
		func(t *testing.T) {
			userID := user.ID(1)
			query := record.NewArticleQuery(
				record.ID(5),
				record.ArticleKeyword("感想"),
				record.ArticleSortFieldTitle,
				record.SortOrderAsc,
				record.Page(2),
				record.PerPage(10),
			)
			want := record.NewArticleListResult(
				[]record.ArticleSummary{{ID: record.ArticleID(1)}},
				record.FilteredCount(1),
				record.TotalCount(3),
			)
			articleRepo := &fakeArticleRepository{listResult: want}
			au := NewArticleUsecase(articleRepo, &fakeRepository{})

			ctx := context.Background()

			got, err := au.ListByUserID(ctx, userID, query)
			if err != nil {
				t.Fatalf(
					"ListByUserID(ctx, %v, %v) (record.ArticleListResult, error) = %v, %v",
					userID, query, got, err,
				)
			}
			if len(got.Articles) != 1 || got.FilteredCount != want.FilteredCount || got.TotalCount != want.TotalCount {
				t.Errorf(
					"ListByUserID(ctx, %v, %v) = %v, want %v",
					userID, query, got, want,
				)
			}
			if articleRepo.listUserID != userID || articleRepo.listQuery != query {
				t.Errorf(
					"ListByUserID(ctx, %v, %v) calls the repository with %v, %v",
					userID, query, articleRepo.listUserID, articleRepo.listQuery,
				)
			}
		},
	)
}

func TestArticleCreate(t *testing.T) {
	userID := user.ID(1)
	recordID := record.ID(10)

	t.Run(
		"creates the article for the authenticated user and returns it with its record",
		func(t *testing.T) {
			articleRepo := &fakeArticleRepository{}
			recordRepo := &fakeRepository{
				record: &record.Record{
					ID:          recordID,
					UserID:      userID,
					Title:       record.Title("Test Movie"),
					ReleaseYear: record.ReleaseYear(2020),
					PosterURL:   record.PosterURL("https://example.com/poster.jpg"),
				},
			}
			au := NewArticleUsecase(articleRepo, recordRepo)

			ctx := context.Background()
			a := record.Article{UserID: user.ID(2), RecordID: recordID, Title: "Test Article"}

			got, err := au.Create(ctx, userID, a)
			if err != nil {
				t.Fatalf(
					"Create(ctx, %v, %v) (*record.ArticleDetail, error) = %v, %v",
					userID, a, got, err,
				)
			}
			if got.Article.UserID != userID {
				t.Errorf(
					"Create(ctx, %v, %v) UserID = %v, want %v",
					userID, a, got.Article.UserID, userID,
				)
			}
			wantRecord := record.ArticleRecordSummary{
				ID:          recordID,
				Title:       record.Title("Test Movie"),
				ReleaseYear: record.ReleaseYear(2020),
				PosterURL:   record.PosterURL("https://example.com/poster.jpg"),
			}
			if got.Record != wantRecord {
				t.Errorf(
					"Create(ctx, %v, %v) Record = %v, want %v",
					userID, a, got.Record, wantRecord,
				)
			}
			if articleRepo.created == nil || articleRepo.created.UserID != userID {
				t.Errorf(
					"Create(ctx, %v, %v) persists %v, want UserID %v",
					userID, a, articleRepo.created, userID,
				)
			}
		},
	)

	t.Run(
		"returns ErrNotFound when the record belongs to another user",
		func(t *testing.T) {
			articleRepo := &fakeArticleRepository{}
			recordRepo := &fakeRepository{
				record: &record.Record{ID: recordID, UserID: user.ID(2)},
			}
			au := NewArticleUsecase(articleRepo, recordRepo)

			ctx := context.Background()
			a := record.Article{RecordID: recordID, Title: "Test Article"}

			got, err := au.Create(ctx, userID, a)
			if !errors.Is(err, exception.ErrNotFound) {
				t.Fatalf(
					"Create(ctx, %v, %v) (*record.ArticleDetail, error) = %v, %v, want %v",
					userID, a, got, err, exception.ErrNotFound,
				)
			}
			if articleRepo.created != nil {
				t.Errorf(
					"Create(ctx, %v, %v) persists %v, want nothing",
					userID, a, articleRepo.created,
				)
			}
		},
	)
}

func TestArticleUpdate(t *testing.T) {
	userID := user.ID(1)
	articleID := record.ArticleID(10)
	recordID := record.ID(20)

	t.Run(
		"updates the article while keeping its record and creation time",
		func(t *testing.T) {
			createdAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
			currentRecord := record.ArticleRecordSummary{
				ID:          recordID,
				Title:       record.Title("Test Movie"),
				ReleaseYear: record.ReleaseYear(2020),
				PosterURL:   record.PosterURL("https://example.com/poster.jpg"),
			}
			articleRepo := &fakeArticleRepository{
				detail: &record.ArticleDetail{
					Article: record.Article{
						ID:        articleID,
						UserID:    userID,
						RecordID:  recordID,
						CreatedAt: createdAt,
					},
					Record: currentRecord,
				},
			}
			au := NewArticleUsecase(articleRepo, &fakeRepository{})

			ctx := context.Background()
			a := record.Article{
				ID:       record.ArticleID(999),
				UserID:   user.ID(2),
				RecordID: record.ID(999),
				Title:    "Updated Article",
			}

			got, err := au.Update(ctx, userID, articleID, a)
			if err != nil {
				t.Fatalf(
					"Update(ctx, %v, %v, %v) (*record.ArticleDetail, error) = %v, %v",
					userID, articleID, a, got, err,
				)
			}
			if got.Article.ID != articleID ||
				got.Article.UserID != userID ||
				got.Article.RecordID != recordID ||
				!got.Article.CreatedAt.Equal(createdAt) {
				t.Errorf(
					"Update(ctx, %v, %v, %v) = %v, want ID %v, UserID %v, RecordID %v, CreatedAt %v",
					userID, articleID, a, got, articleID, userID, recordID, createdAt,
				)
			}
			if got.Record != currentRecord {
				t.Errorf(
					"Update(ctx, %v, %v, %v) Record = %v, want %v",
					userID, articleID, a, got.Record, currentRecord,
				)
			}
			if articleRepo.updated == nil || articleRepo.updated.Title != "Updated Article" {
				t.Errorf(
					"Update(ctx, %v, %v, %v) persists %v, want Title %q",
					userID, articleID, a, articleRepo.updated, "Updated Article",
				)
			}
		},
	)

	t.Run(
		"returns ErrNotFound when the article belongs to another user",
		func(t *testing.T) {
			articleRepo := &fakeArticleRepository{
				detail: &record.ArticleDetail{
					Article: record.Article{ID: articleID, UserID: user.ID(2)},
				},
			}
			au := NewArticleUsecase(articleRepo, &fakeRepository{})

			ctx := context.Background()
			a := record.Article{Title: "Updated Article"}

			got, err := au.Update(ctx, userID, articleID, a)
			if !errors.Is(err, exception.ErrNotFound) {
				t.Fatalf(
					"Update(ctx, %v, %v, %v) (*record.ArticleDetail, error) = %v, %v, want %v",
					userID, articleID, a, got, err, exception.ErrNotFound,
				)
			}
			if articleRepo.updated != nil {
				t.Errorf(
					"Update(ctx, %v, %v, %v) persists %v, want nothing",
					userID, articleID, a, articleRepo.updated,
				)
			}
		},
	)
}

func TestArticleDelete(t *testing.T) {
	userID := user.ID(1)
	articleID := record.ArticleID(10)

	t.Run(
		"deletes the article when it belongs to the user",
		func(t *testing.T) {
			articleRepo := &fakeArticleRepository{
				detail: &record.ArticleDetail{
					Article: record.Article{ID: articleID, UserID: userID},
				},
			}
			au := NewArticleUsecase(articleRepo, &fakeRepository{})

			ctx := context.Background()

			if err := au.Delete(ctx, userID, articleID); err != nil {
				t.Fatalf("Delete(ctx, %v, %v) error = %v", userID, articleID, err)
			}
			if articleRepo.deletedID != articleID {
				t.Errorf(
					"Delete(ctx, %v, %v) deletes %v, want %v",
					userID, articleID, articleRepo.deletedID, articleID,
				)
			}
		},
	)

	t.Run(
		"returns ErrNotFound when the article belongs to another user",
		func(t *testing.T) {
			articleRepo := &fakeArticleRepository{
				detail: &record.ArticleDetail{
					Article: record.Article{ID: articleID, UserID: user.ID(2)},
				},
			}
			au := NewArticleUsecase(articleRepo, &fakeRepository{})

			ctx := context.Background()

			err := au.Delete(ctx, userID, articleID)
			if !errors.Is(err, exception.ErrNotFound) {
				t.Fatalf(
					"Delete(ctx, %v, %v) error = %v, want %v",
					userID, articleID, err, exception.ErrNotFound,
				)
			}
			if articleRepo.deleteCalls != 0 {
				t.Errorf(
					"Delete(ctx, %v, %v) calls the repository %d times, want 0",
					userID, articleID, articleRepo.deleteCalls,
				)
			}
		},
	)
}
