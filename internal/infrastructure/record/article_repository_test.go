package record

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/record"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
	"github.com/masaya-nishimura-09/movie-log-api/internal/testutil"
	"gorm.io/gorm"
)

func createTestRecord(t *testing.T, tx *gorm.DB, userID user.ID) record.ID {
	t.Helper()
	rec := newTestRecord(userID, time.Now())
	if err := NewRepository(tx).Create(context.Background(), &rec); err != nil {
		t.Fatalf("Create(ctx, %v) error = %v", rec, err)
	}
	return rec.ID
}

func newTestArticleContent(text string) record.ArticleContent {
	return record.ArticleContent(fmt.Sprintf(
		`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":%q}]}]}`,
		text,
	))
}

func newTestArticle(userID user.ID, recordID record.ID, title string) record.Article {
	return record.NewArticle(
		userID,
		recordID,
		record.ArticleTitle(title),
		newTestArticleContent("感想です"),
	)
}

func createTestArticle(t *testing.T, tx *gorm.DB, a *record.Article) {
	t.Helper()
	if err := NewArticleRepository(tx).Create(context.Background(), a); err != nil {
		t.Fatalf("Create(ctx, %v) error = %v", *a, err)
	}
}

func equalJSON(got, want []byte) bool {
	var g, w any
	if json.Unmarshal(got, &g) != nil || json.Unmarshal(want, &w) != nil {
		return false
	}
	return reflect.DeepEqual(g, w)
}

func assertArticleEqual(t *testing.T, call string, got, want *record.Article) {
	t.Helper()

	if got.ID != want.ID ||
		got.UserID != want.UserID ||
		got.RecordID != want.RecordID ||
		got.Title != want.Title {
		t.Errorf("%s = %v, want %v", call, got, want)
	}
	if !equalJSON(got.Content, want.Content) {
		t.Errorf("%s = %v, want Content %s", call, got, want.Content)
	}
}

func assertSummariesEqual(t *testing.T, call string, got []record.ArticleSummary, want []record.Article) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("%s = %v, want %d summaries", call, got, len(want))
	}
	for i := range want {
		if got[i].ID != want[i].ID ||
			got[i].Record.ID != want[i].RecordID ||
			got[i].Title != want[i].Title {
			t.Errorf("%s[%d] = %v, want %v", call, i, got[i], want[i])
		}
	}
}

func contentTextOf(t *testing.T, tx *gorm.DB, articleID record.ArticleID) string {
	t.Helper()

	var text string
	if err := tx.Raw(
		"SELECT content_text FROM record_articles WHERE id = ?", uint(articleID),
	).Scan(&text).Error; err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	return text
}

func TestArticleGetByID(t *testing.T) {
	t.Run(
		"returns the article and its record when the ID exists",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			r := NewArticleRepository(tx)

			ctx := context.Background()
			userID := newTestUser(t, tx, "test@example.com").ID
			rec := newTestRecord(userID, time.Now())
			if err := NewRepository(tx).Create(ctx, &rec); err != nil {
				t.Fatalf("Create(ctx, %v) error = %v", rec, err)
			}
			want := newTestArticle(userID, rec.ID, "Test Article")
			createTestArticle(t, tx, &want)

			got, err := r.GetByID(ctx, want.ID)
			if err != nil {
				t.Fatalf(
					"GetByID(ctx, %d) (record.ArticleDetail, error) = %v, %v",
					want.ID, got, err,
				)
			}
			assertArticleEqual(
				t,
				fmt.Sprintf("GetByID(ctx, %d) (record.ArticleDetail, error)", want.ID),
				&got.Article, &want,
			)
			wantRecord := record.ArticleRecordSummary{
				ID:          rec.ID,
				Title:       rec.Title,
				ReleaseYear: rec.ReleaseYear,
				PosterURL:   rec.PosterURL,
			}
			if got.Record != wantRecord {
				t.Errorf(
					"GetByID(ctx, %d) Record = %v, want %v",
					want.ID, got.Record, wantRecord,
				)
			}
		},
	)

	t.Run(
		"returns ErrNotFound when the ID does not exist",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			r := NewArticleRepository(tx)

			ctx := context.Background()
			fakeID := record.ArticleID(999999)

			got, err := r.GetByID(ctx, fakeID)
			if !errors.Is(err, exception.ErrNotFound) {
				t.Fatalf(
					"GetByID(ctx, %d) (record.ArticleDetail, error) = %v, %v, want %v",
					fakeID, got, err, exception.ErrNotFound,
				)
			}
			if got != nil {
				t.Errorf(
					"GetByID(ctx, %d) (record.ArticleDetail, error) = %v, %v, want nil",
					fakeID, got, err,
				)
			}
		},
	)

	t.Run(
		"returns a wrapped error when the context is canceled",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			r := NewArticleRepository(tx)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			id := record.ArticleID(1)
			got, err := r.GetByID(ctx, id)

			if !errors.Is(err, context.Canceled) {
				t.Fatalf(
					"GetByID(ctx, %d) (record.ArticleDetail, error) = %v, %v, want %v",
					id, got, err, context.Canceled,
				)
			}
			if got != nil {
				t.Errorf(
					"GetByID(ctx, %d) (record.ArticleDetail, error) = %v, %v, want nil",
					id, got, err,
				)
			}
		},
	)
}

func newEmptyArticleQuery() record.ArticleQuery {
	return record.NewArticleQuery(
		0,
		"",
		record.ArticleSortFieldCreatedAt,
		record.SortOrderDesc,
		record.Page(1),
		record.PerPage(20),
	)
}

func TestArticleListByUserID(t *testing.T) {
	t.Run(
		"returns only the user's article summaries, newest first, with counts",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			r := NewArticleRepository(tx)

			ctx := context.Background()
			userID := newTestUser(t, tx, "test@example.com").ID
			otherUserID := newTestUser(t, tx, "other@example.com").ID
			recordID := createTestRecord(t, tx, userID)
			otherRecordID := createTestRecord(t, tx, otherUserID)

			first := newTestArticle(userID, recordID, "First")
			second := newTestArticle(userID, recordID, "Second")
			third := newTestArticle(userID, recordID, "Third")
			other := newTestArticle(otherUserID, otherRecordID, "Other")
			for _, a := range []*record.Article{&first, &second, &third, &other} {
				createTestArticle(t, tx, a)
			}

			got, err := r.ListByUserID(ctx, userID, newEmptyArticleQuery())
			if err != nil {
				t.Fatalf(
					"ListByUserID(ctx, %d, query) (record.ArticleListResult, error) = %v, %v",
					userID, got, err,
				)
			}
			assertSummariesEqual(
				t,
				fmt.Sprintf("ListByUserID(ctx, %d, query)", userID),
				got.Articles, []record.Article{third, second, first},
			)
			if got.FilteredCount != 3 || got.TotalCount != 3 {
				t.Errorf(
					"ListByUserID(ctx, %d, query) = FilteredCount %d, TotalCount %d, want 3, 3",
					userID, got.FilteredCount, got.TotalCount,
				)
			}
		},
	)

	t.Run(
		"includes the title, release year, and poster URL of the record",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			r := NewArticleRepository(tx)

			ctx := context.Background()
			userID := newTestUser(t, tx, "test@example.com").ID
			rec := newTestRecord(userID, time.Now())
			if err := NewRepository(tx).Create(ctx, &rec); err != nil {
				t.Fatalf("Create(ctx, %v) error = %v", rec, err)
			}
			a := newTestArticle(userID, rec.ID, "Test Article")
			createTestArticle(t, tx, &a)

			got, err := r.ListByUserID(ctx, userID, newEmptyArticleQuery())
			if err != nil || len(got.Articles) != 1 {
				t.Fatalf(
					"ListByUserID(ctx, %d, query) (record.ArticleListResult, error) = %v, %v, want 1 summary",
					userID, got, err,
				)
			}
			want := record.ArticleRecordSummary{
				ID:          rec.ID,
				Title:       rec.Title,
				ReleaseYear: rec.ReleaseYear,
				PosterURL:   rec.PosterURL,
			}
			if got.Articles[0].Record != want {
				t.Errorf(
					"ListByUserID(ctx, %d, query) Record = %v, want %v",
					userID, got.Articles[0].Record, want,
				)
			}
		},
	)

	t.Run(
		"filters by record ID",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			r := NewArticleRepository(tx)

			ctx := context.Background()
			userID := newTestUser(t, tx, "test@example.com").ID
			recordID := createTestRecord(t, tx, userID)
			otherRecordID := createTestRecord(t, tx, userID)

			first := newTestArticle(userID, recordID, "First")
			second := newTestArticle(userID, recordID, "Second")
			other := newTestArticle(userID, otherRecordID, "Other")
			for _, a := range []*record.Article{&first, &second, &other} {
				createTestArticle(t, tx, a)
			}

			query := newEmptyArticleQuery()
			query.RecordID = recordID

			got, err := r.ListByUserID(ctx, userID, query)
			if err != nil {
				t.Fatalf(
					"ListByUserID(ctx, %d, query) (record.ArticleListResult, error) = %v, %v",
					userID, got, err,
				)
			}
			assertSummariesEqual(
				t,
				fmt.Sprintf("ListByUserID(ctx, %d, query)", userID),
				got.Articles, []record.Article{second, first},
			)
			if got.FilteredCount != 2 || got.TotalCount != 3 {
				t.Errorf(
					"ListByUserID(ctx, %d, query) = FilteredCount %d, TotalCount %d, want 2, 3",
					userID, got.FilteredCount, got.TotalCount,
				)
			}
		},
	)

	t.Run(
		"filters by keyword in the title or the plain text",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			r := NewArticleRepository(tx)

			ctx := context.Background()
			userID := newTestUser(t, tx, "test@example.com").ID
			recordID := createTestRecord(t, tx, userID)

			inTitle := newTestArticle(userID, recordID, "Matrix の感想")
			inTitle.Content = newTestArticleContent("面白かった")
			inText := newTestArticle(userID, recordID, "メモ")
			inText.Content = newTestArticleContent("この映画の感想を書く")
			unrelated := newTestArticle(userID, recordID, "別の記事")
			unrelated.Content = newTestArticleContent("面白かった")
			for _, a := range []*record.Article{&inTitle, &inText, &unrelated} {
				createTestArticle(t, tx, a)
			}

			query := newEmptyArticleQuery()
			query.Keyword = record.ArticleKeyword("感想")

			got, err := r.ListByUserID(ctx, userID, query)
			if err != nil {
				t.Fatalf(
					"ListByUserID(ctx, %d, query) (record.ArticleListResult, error) = %v, %v",
					userID, got, err,
				)
			}
			assertSummariesEqual(
				t,
				fmt.Sprintf("ListByUserID(ctx, %d, query)", userID),
				got.Articles, []record.Article{inText, inTitle},
			)
			if got.FilteredCount != 2 || got.TotalCount != 3 {
				t.Errorf(
					"ListByUserID(ctx, %d, query) = FilteredCount %d, TotalCount %d, want 2, 3",
					userID, got.FilteredCount, got.TotalCount,
				)
			}
		},
	)

	t.Run(
		"treats LIKE wildcards in the keyword as literal characters",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			r := NewArticleRepository(tx)

			ctx := context.Background()
			userID := newTestUser(t, tx, "test@example.com").ID
			recordID := createTestRecord(t, tx, userID)

			percent := newTestArticle(userID, recordID, "割合")
			percent.Content = newTestArticleContent("満足度は100%")
			plain := newTestArticle(userID, recordID, "普通")
			plain.Content = newTestArticleContent("満足度は高い")
			for _, a := range []*record.Article{&percent, &plain} {
				createTestArticle(t, tx, a)
			}

			query := newEmptyArticleQuery()
			query.Keyword = record.ArticleKeyword("%")

			got, err := r.ListByUserID(ctx, userID, query)
			if err != nil {
				t.Fatalf(
					"ListByUserID(ctx, %d, query) (record.ArticleListResult, error) = %v, %v",
					userID, got, err,
				)
			}
			assertSummariesEqual(
				t,
				fmt.Sprintf("ListByUserID(ctx, %d, query)", userID),
				got.Articles, []record.Article{percent},
			)
		},
	)

	t.Run(
		"sorts by the given field and order",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			r := NewArticleRepository(tx)

			ctx := context.Background()
			userID := newTestUser(t, tx, "test@example.com").ID
			recordID := createTestRecord(t, tx, userID)

			b := newTestArticle(userID, recordID, "B")
			a := newTestArticle(userID, recordID, "A")
			c := newTestArticle(userID, recordID, "C")
			for _, article := range []*record.Article{&b, &a, &c} {
				createTestArticle(t, tx, article)
			}

			query := newEmptyArticleQuery()
			query.SortField = record.ArticleSortFieldTitle
			query.SortOrder = record.SortOrderAsc

			got, err := r.ListByUserID(ctx, userID, query)
			if err != nil {
				t.Fatalf(
					"ListByUserID(ctx, %d, query) (record.ArticleListResult, error) = %v, %v",
					userID, got, err,
				)
			}
			assertSummariesEqual(
				t,
				fmt.Sprintf("ListByUserID(ctx, %d, query)", userID),
				got.Articles, []record.Article{a, b, c},
			)
		},
	)

	t.Run(
		"paginates while reporting the counts before pagination",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			r := NewArticleRepository(tx)

			ctx := context.Background()
			userID := newTestUser(t, tx, "test@example.com").ID
			recordID := createTestRecord(t, tx, userID)

			first := newTestArticle(userID, recordID, "First")
			second := newTestArticle(userID, recordID, "Second")
			third := newTestArticle(userID, recordID, "Third")
			for _, a := range []*record.Article{&first, &second, &third} {
				createTestArticle(t, tx, a)
			}

			query := newEmptyArticleQuery()
			query.Page = record.Page(2)
			query.PerPage = record.PerPage(2)

			got, err := r.ListByUserID(ctx, userID, query)
			if err != nil {
				t.Fatalf(
					"ListByUserID(ctx, %d, query) (record.ArticleListResult, error) = %v, %v",
					userID, got, err,
				)
			}
			assertSummariesEqual(
				t,
				fmt.Sprintf("ListByUserID(ctx, %d, query)", userID),
				got.Articles, []record.Article{first},
			)
			if got.FilteredCount != 3 || got.TotalCount != 3 {
				t.Errorf(
					"ListByUserID(ctx, %d, query) = FilteredCount %d, TotalCount %d, want 3, 3",
					userID, got.FilteredCount, got.TotalCount,
				)
			}
		},
	)

	t.Run(
		"returns an empty list when the user has no articles",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			r := NewArticleRepository(tx)

			ctx := context.Background()
			userID := newTestUser(t, tx, "test@example.com").ID

			got, err := r.ListByUserID(ctx, userID, newEmptyArticleQuery())
			if err != nil || len(got.Articles) != 0 || got.FilteredCount != 0 || got.TotalCount != 0 {
				t.Fatalf(
					"ListByUserID(ctx, %d, query) (record.ArticleListResult, error) = %v, %v, want empty",
					userID, got, err,
				)
			}
		},
	)

	t.Run(
		"returns a wrapped error when the context is canceled",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			r := NewArticleRepository(tx)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			userID := user.ID(1)
			got, err := r.ListByUserID(ctx, userID, newEmptyArticleQuery())

			if !errors.Is(err, context.Canceled) {
				t.Fatalf(
					"ListByUserID(ctx, %d, query) (record.ArticleListResult, error) = %v, %v, want %v",
					userID, got, err, context.Canceled,
				)
			}
		},
	)
}

func TestArticleCreate(t *testing.T) {
	t.Run(
		"persists the article and its plain text",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			r := NewArticleRepository(tx)

			ctx := context.Background()
			userID := newTestUser(t, tx, "test@example.com").ID
			recordID := createTestRecord(t, tx, userID)
			want := newTestArticle(userID, recordID, "Test Article")

			if err := r.Create(ctx, &want); err != nil {
				t.Fatalf("Create(ctx, %v) error = %v", want, err)
			}
			if want.ID == 0 {
				t.Fatalf("Create(ctx, %v) sets ID = 0, want non-zero", want)
			}
			if want.CreatedAt.IsZero() || want.UpdatedAt.IsZero() {
				t.Fatalf(
					"Create(ctx, %v) sets CreatedAt = %v, UpdatedAt = %v, want non-zero",
					want, want.CreatedAt, want.UpdatedAt,
				)
			}

			got, err := r.GetByID(ctx, want.ID)
			if err != nil {
				t.Fatalf(
					"GetByID(ctx, %d) (record.ArticleDetail, error) = %v, %v",
					want.ID, got, err,
				)
			}
			assertArticleEqual(
				t,
				fmt.Sprintf("GetByID(ctx, %d) (record.ArticleDetail, error)", want.ID),
				&got.Article, &want,
			)

			if text := contentTextOf(t, tx, want.ID); text != "感想です" {
				t.Errorf(
					"Create(ctx, %v) persists content_text = %q, want %q",
					want, text, "感想です",
				)
			}
		},
	)

	t.Run(
		"returns ErrNotFound when the record does not exist",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			r := NewArticleRepository(tx)

			ctx := context.Background()
			userID := newTestUser(t, tx, "test@example.com").ID
			a := newTestArticle(userID, record.ID(999999), "Test Article")

			if err := r.Create(ctx, &a); !errors.Is(err, exception.ErrNotFound) {
				t.Fatalf(
					"Create(ctx, %v) error = %v, want %v",
					a, err, exception.ErrNotFound,
				)
			}
		},
	)

	t.Run(
		"returns a wrapped error when the context is canceled",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			r := NewArticleRepository(tx)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			a := newTestArticle(user.ID(1), record.ID(1), "Test Article")
			err := r.Create(ctx, &a)
			if !errors.Is(err, context.Canceled) {
				t.Fatalf(
					"Create(ctx, %v) error = %v, want %v",
					a, err, context.Canceled,
				)
			}
		},
	)
}

func TestArticleUpdate(t *testing.T) {
	t.Run(
		"updates the title, content, and plain text",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			r := NewArticleRepository(tx)

			ctx := context.Background()
			userID := newTestUser(t, tx, "test@example.com").ID
			recordID := createTestRecord(t, tx, userID)
			created := newTestArticle(userID, recordID, "Test Article")
			createTestArticle(t, tx, &created)

			want := record.Article{
				ID:       created.ID,
				UserID:   userID,
				RecordID: recordID,
				Title:    record.ArticleTitle("Updated Article"),
				Content:  newTestArticleContent("更新後の本文"),
			}
			if err := r.Update(ctx, &want); err != nil {
				t.Fatalf("Update(ctx, %v) error = %v", want, err)
			}

			got, err := r.GetByID(ctx, created.ID)
			if err != nil {
				t.Fatalf(
					"GetByID(ctx, %d) (record.ArticleDetail, error) = %v, %v",
					created.ID, got, err,
				)
			}
			assertArticleEqual(
				t,
				fmt.Sprintf("GetByID(ctx, %d) (record.ArticleDetail, error)", created.ID),
				&got.Article, &want,
			)

			if !got.Article.CreatedAt.Equal(created.CreatedAt.Truncate(time.Microsecond)) {
				t.Errorf(
					"GetByID(ctx, %d) returns CreatedAt = %v, want %v",
					created.ID, got.Article.CreatedAt, created.CreatedAt.Truncate(time.Microsecond),
				)
			}
			if !got.Article.UpdatedAt.After(created.UpdatedAt) {
				t.Errorf(
					"GetByID(ctx, %d) returns UpdatedAt = %v, want after %v",
					created.ID, got.Article.UpdatedAt, created.UpdatedAt,
				)
			}
			if text := contentTextOf(t, tx, created.ID); text != "更新後の本文" {
				t.Errorf(
					"Update(ctx, %v) leaves content_text = %q, want %q",
					want, text, "更新後の本文",
				)
			}
		},
	)

	t.Run(
		"returns ErrNotFound when the article does not exist",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			r := NewArticleRepository(tx)

			ctx := context.Background()
			userID := newTestUser(t, tx, "test@example.com").ID
			recordID := createTestRecord(t, tx, userID)

			a := newTestArticle(userID, recordID, "Test Article")
			a.ID = record.ArticleID(999999)

			if err := r.Update(ctx, &a); !errors.Is(err, exception.ErrNotFound) {
				t.Fatalf(
					"Update(ctx, %v) error = %v, want %v",
					a, err, exception.ErrNotFound,
				)
			}
		},
	)

	t.Run(
		"returns ErrNotFound when the article belongs to another user",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			r := NewArticleRepository(tx)

			ctx := context.Background()
			userID := newTestUser(t, tx, "test@example.com").ID
			otherUserID := newTestUser(t, tx, "other@example.com").ID
			recordID := createTestRecord(t, tx, userID)

			created := newTestArticle(userID, recordID, "Test Article")
			createTestArticle(t, tx, &created)

			hijacked := newTestArticle(otherUserID, recordID, "Hijacked Article")
			hijacked.ID = created.ID

			if err := r.Update(ctx, &hijacked); !errors.Is(err, exception.ErrNotFound) {
				t.Fatalf(
					"Update(ctx, %v) error = %v, want %v",
					hijacked, err, exception.ErrNotFound,
				)
			}
		},
	)

	t.Run(
		"returns a wrapped error when the context is canceled",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			r := NewArticleRepository(tx)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			a := newTestArticle(user.ID(1), record.ID(1), "Test Article")
			a.ID = record.ArticleID(1)

			err := r.Update(ctx, &a)
			if !errors.Is(err, context.Canceled) {
				t.Fatalf(
					"Update(ctx, %v) error = %v, want %v",
					a, err, context.Canceled,
				)
			}
		},
	)
}

func TestArticleDelete(t *testing.T) {
	t.Run(
		"deletes the article",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			r := NewArticleRepository(tx)

			ctx := context.Background()
			userID := newTestUser(t, tx, "test@example.com").ID
			recordID := createTestRecord(t, tx, userID)
			a := newTestArticle(userID, recordID, "Test Article")
			createTestArticle(t, tx, &a)

			if err := r.Delete(ctx, a.ID); err != nil {
				t.Fatalf("Delete(ctx, %d) error = %v", a.ID, err)
			}

			got, err := r.GetByID(ctx, a.ID)
			if !errors.Is(err, exception.ErrNotFound) {
				t.Fatalf(
					"GetByID(ctx, %d) (record.ArticleDetail, error) = %v, %v, want %v",
					a.ID, got, err, exception.ErrNotFound,
				)
			}
		},
	)

	t.Run(
		"is deleted together with its record",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			r := NewArticleRepository(tx)

			ctx := context.Background()
			userID := newTestUser(t, tx, "test@example.com").ID
			recordID := createTestRecord(t, tx, userID)
			a := newTestArticle(userID, recordID, "Test Article")
			createTestArticle(t, tx, &a)

			if err := NewRepository(tx).Delete(ctx, recordID); err != nil {
				t.Fatalf("Delete(ctx, %d) error = %v", recordID, err)
			}

			got, err := r.GetByID(ctx, a.ID)
			if !errors.Is(err, exception.ErrNotFound) {
				t.Fatalf(
					"GetByID(ctx, %d) (record.ArticleDetail, error) = %v, %v, want %v",
					a.ID, got, err, exception.ErrNotFound,
				)
			}
		},
	)

	t.Run(
		"returns ErrNotFound when the article does not exist",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			r := NewArticleRepository(tx)

			ctx := context.Background()
			id := record.ArticleID(999999)

			if err := r.Delete(ctx, id); !errors.Is(err, exception.ErrNotFound) {
				t.Fatalf(
					"Delete(ctx, %d) error = %v, want %v",
					id, err, exception.ErrNotFound,
				)
			}
		},
	)

	t.Run(
		"returns a wrapped error when the context is canceled",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			r := NewArticleRepository(tx)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			id := record.ArticleID(1)
			err := r.Delete(ctx, id)
			if !errors.Is(err, context.Canceled) {
				t.Fatalf(
					"Delete(ctx, %d) error = %v, want %v",
					id, err, context.Canceled,
				)
			}
		},
	)
}
