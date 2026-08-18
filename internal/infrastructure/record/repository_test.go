package record

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/record"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
	userinfra "github.com/masaya-nishimura-09/movie-log-api/internal/infrastructure/user"
	"github.com/masaya-nishimura-09/movie-log-api/internal/testutil"
	"gorm.io/gorm"
)

var testDB *gorm.DB

func TestMain(m *testing.M) {
	testDB = testutil.NewTestDB()
	code := m.Run()
	os.Exit(code)
}

func newTestUser(t *testing.T, tx *gorm.DB, email string) user.User {
	t.Helper()
	u := user.User{
		Username:       user.Username("Test"),
		Email:          user.Email(email),
		HashedPassword: user.HashedPassword("testpassword"),
		Role:           user.RoleUser,
	}
	if err := userinfra.NewUserRepo(tx).Create(context.Background(), &u); err != nil {
		t.Fatalf(
			"Create(ctx, %v) error = %v",
			u, err,
		)
	}
	return u
}

func newTestRecord(userID user.ID, watchedAt time.Time) record.Record {
	return record.Record{
		UserID:      userID,
		Title:       record.Title("Test Movie"),
		ReleaseYear: record.ReleaseYear(2020),
		Runtime:     record.Runtime(120),
		Genres:      []record.Genre{record.GenreDrama, record.GenreThriller},
		Countries:   []record.Country{record.Country("US"), record.Country("JP")},
		Language:    record.Language("en"),
		Credits: []record.Credit{
			{
				PersonName: record.PersonName("Test Director"),
				CreditRole: record.CreditRoleDirector,
			},
			{
				PersonName: record.PersonName("Test Actor"),
				CreditRole: record.CreditRoleCast,
			},
		},
		PosterURL: record.PosterURL("https://example.com/poster.jpg"),
		WatchedAt: watchedAt,
		Platform:  record.PlatformNetflix,
		Score:     record.Score(4),
		MoodTags:  []record.MoodTag{record.MoodTagMoving, record.MoodTagTense},
		Memo:      record.Memo("test memo"),
	}
}

func equalGenres(got, want []record.Genre) bool {
	g := slices.Clone(got)
	w := slices.Clone(want)
	slices.Sort(g)
	slices.Sort(w)
	return slices.Equal(g, w)
}

func equalCountries(got, want []record.Country) bool {
	g := slices.Clone(got)
	w := slices.Clone(want)
	slices.Sort(g)
	slices.Sort(w)
	return slices.Equal(g, w)
}

func equalMoodTags(got, want []record.MoodTag) bool {
	g := slices.Clone(got)
	w := slices.Clone(want)
	slices.Sort(g)
	slices.Sort(w)
	return slices.Equal(g, w)
}

func equalCredits(got, want []record.Credit) bool {
	compare := func(a, b record.Credit) int {
		if c := cmp.Compare(a.PersonName, b.PersonName); c != 0 {
			return c
		}
		return cmp.Compare(a.CreditRole, b.CreditRole)
	}

	g := slices.Clone(got)
	w := slices.Clone(want)
	slices.SortFunc(g, compare)
	slices.SortFunc(w, compare)
	return slices.Equal(g, w)
}

func assertRecordEqual(t *testing.T, call string, got, want *record.Record) {
	t.Helper()

	if got.ID != want.ID ||
		got.UserID != want.UserID ||
		got.Title != want.Title ||
		got.ReleaseYear != want.ReleaseYear ||
		got.Runtime != want.Runtime ||
		got.Language != want.Language ||
		got.PosterURL != want.PosterURL ||
		got.Platform != want.Platform ||
		got.Score != want.Score ||
		got.Memo != want.Memo {
		t.Errorf(
			"%s = %v, want %v",
			call, got, want,
		)
	}
	if !got.WatchedAt.Equal(want.WatchedAt.Truncate(time.Microsecond)) {
		t.Errorf(
			"%s = %v, want WatchedAt %v",
			call, got, want.WatchedAt.Truncate(time.Microsecond),
		)
	}
	if !equalGenres(got.Genres, want.Genres) {
		t.Errorf(
			"%s = %v, want Genres %v",
			call, got, want.Genres,
		)
	}
	if !equalCountries(got.Countries, want.Countries) {
		t.Errorf(
			"%s = %v, want Countries %v",
			call, got, want.Countries,
		)
	}
	if !equalMoodTags(got.MoodTags, want.MoodTags) {
		t.Errorf(
			"%s = %v, want MoodTags %v",
			call, got, want.MoodTags,
		)
	}
	if !equalCredits(got.Credits, want.Credits) {
		t.Errorf(
			"%s = %v, want Credits %v",
			call, got, want.Credits,
		)
	}
}

func countChildRows(t *testing.T, tx *gorm.DB, model any, recordID record.ID) int64 {
	t.Helper()

	var count int64
	if err := tx.Model(model).
		Where("record_id = ?", uint(recordID)).
		Count(&count).Error; err != nil {
		t.Fatalf(
			"Count() error = %v",
			err,
		)
	}
	return count
}

func countAssociations(t *testing.T, tx *gorm.DB, recordID record.ID) int64 {
	t.Helper()

	return countChildRows(t, tx, &genreDTO{}, recordID) +
		countChildRows(t, tx, &countryDTO{}, recordID) +
		countChildRows(t, tx, &creditDTO{}, recordID) +
		countChildRows(t, tx, &moodTagDTO{}, recordID)
}

func TestGetByID(t *testing.T) {
	t.Run(
		"returns the record with its associations when the ID exists",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			rr := NewRecordRepo(tx)

			ctx := context.Background()
			userID := newTestUser(t, tx, "test@example.com").ID
			want := newTestRecord(userID, time.Now())

			if err := rr.Create(ctx, &want); err != nil {
				t.Fatalf("Create(ctx, %v) error = %v", want, err)
			}

			got, err := rr.GetByID(ctx, want.ID)
			if err != nil {
				t.Fatalf(
					"GetByID(ctx, %d) (record.Record, error) = %v, %v",
					want.ID, got, err,
				)
			}
			assertRecordEqual(
				t,
				fmt.Sprintf("GetByID(ctx, %d) (record.Record, error)", want.ID),
				got, &want,
			)
		},
	)

	t.Run(
		"returns ErrNotFound when the ID does not exist",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			rr := NewRecordRepo(tx)

			ctx := context.Background()
			fakeID := record.ID(999999)

			got, err := rr.GetByID(ctx, fakeID)
			if !errors.Is(err, exception.ErrNotFound) {
				t.Fatalf(
					"GetByID(ctx, %d) (record.Record, error) = %v, %v, want %v",
					fakeID, got, err, exception.ErrNotFound,
				)
			}
			if got != nil {
				t.Errorf(
					"GetByID(ctx, %d) (record.Record, error) = %v, %v, want nil",
					fakeID, got, err,
				)
			}
		},
	)

	t.Run(
		"returns a wrapped error when the context is canceled",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			rr := NewRecordRepo(tx)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			id := record.ID(1)
			got, err := rr.GetByID(ctx, id)

			if !errors.Is(err, context.Canceled) {
				t.Fatalf(
					"GetByID(ctx, %d) (record.Record, error) = %v, %v, want %v",
					id, got, err, context.Canceled,
				)
			}
			if got != nil {
				t.Errorf(
					"GetByID(ctx, %d) (record.Record, error) = %v, %v, want nil",
					id, got, err,
				)
			}
		},
	)
}

func TestListByUserID(t *testing.T) {
	t.Run(
		"returns the records of the user ordered by watched_at descending",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			rr := NewRecordRepo(tx)

			ctx := context.Background()
			userID := newTestUser(t, tx, "test@example.com").ID
			otherUserID := newTestUser(t, tx, "other@example.com").ID

			watchedAts := []time.Time{
				time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC),
				time.Date(2026, 8, 15, 12, 0, 0, 0, time.UTC),
				time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC),
			}
			for _, watchedAt := range watchedAts {
				r := newTestRecord(userID, watchedAt)
				if err := rr.Create(ctx, &r); err != nil {
					t.Fatalf("Create(ctx, %v) error = %v", r, err)
				}
			}

			other := newTestRecord(otherUserID, time.Date(2026, 8, 20, 12, 0, 0, 0, time.UTC))
			if err := rr.Create(ctx, &other); err != nil {
				t.Fatalf("Create(ctx, %v) error = %v", other, err)
			}

			got, err := rr.ListByUserID(ctx, userID)
			if err != nil {
				t.Fatalf(
					"ListByUserID(ctx, %d) ([]*record.Record, error) = %v, %v",
					userID, got, err,
				)
			}
			if len(got) != len(watchedAts) {
				t.Fatalf(
					"ListByUserID(ctx, %d) returns %d records, want %d",
					userID, len(got), len(watchedAts),
				)
			}

			wantOrder := []time.Time{watchedAts[1], watchedAts[2], watchedAts[0]}
			for i, r := range got {
				if r.UserID != userID {
					t.Errorf(
						"ListByUserID(ctx, %d) returns a record of user %d",
						userID, r.UserID,
					)
				}
				if !r.WatchedAt.Equal(wantOrder[i]) {
					t.Errorf(
						"ListByUserID(ctx, %d) record[%d] WatchedAt = %v, want %v",
						userID, i, r.WatchedAt, wantOrder[i],
					)
				}
				if len(r.Genres) == 0 || len(r.Credits) == 0 {
					t.Errorf(
						"ListByUserID(ctx, %d) record[%d] = %v, want preloaded associations",
						userID, i, r,
					)
				}
			}
		},
	)

	t.Run(
		"returns an empty slice when the user has no records",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			rr := NewRecordRepo(tx)

			ctx := context.Background()
			userID := newTestUser(t, tx, "test@example.com").ID

			got, err := rr.ListByUserID(ctx, userID)
			if err != nil {
				t.Fatalf(
					"ListByUserID(ctx, %d) ([]*record.Record, error) = %v, %v",
					userID, got, err,
				)
			}
			if len(got) != 0 {
				t.Errorf(
					"ListByUserID(ctx, %d) returns %d records, want 0",
					userID, len(got),
				)
			}
		},
	)

	t.Run(
		"returns a wrapped error when the context is canceled",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			rr := NewRecordRepo(tx)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			userID := user.ID(1)
			got, err := rr.ListByUserID(ctx, userID)

			if !errors.Is(err, context.Canceled) {
				t.Fatalf(
					"ListByUserID(ctx, %d) ([]*record.Record, error) = %v, %v, want %v",
					userID, got, err, context.Canceled,
				)
			}
			if got != nil {
				t.Errorf(
					"ListByUserID(ctx, %d) ([]*record.Record, error) = %v, %v, want nil",
					userID, got, err,
				)
			}
		},
	)
}

func TestCreate(t *testing.T) {
	t.Run(
		"persists the record and its associations",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			rr := NewRecordRepo(tx)

			ctx := context.Background()
			userID := newTestUser(t, tx, "test@example.com").ID
			want := newTestRecord(userID, time.Now())

			if err := rr.Create(ctx, &want); err != nil {
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

			got, err := rr.GetByID(ctx, want.ID)
			if err != nil {
				t.Fatalf(
					"GetByID(ctx, %d) (record.Record, error) = %v, %v",
					want.ID, got, err,
				)
			}
			assertRecordEqual(
				t,
				fmt.Sprintf("GetByID(ctx, %d) (record.Record, error)", want.ID),
				got, &want,
			)

			wantCount := int64(
				len(want.Genres) + len(want.Countries) + len(want.Credits) + len(want.MoodTags),
			)
			if count := countAssociations(t, tx, want.ID); count != wantCount {
				t.Errorf(
					"Create(ctx, %v) persists %d association rows, want %d",
					want, count, wantCount,
				)
			}
		},
	)

	t.Run(
		"returns a wrapped error when the context is canceled",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			rr := NewRecordRepo(tx)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			r := newTestRecord(user.ID(1), time.Now())
			err := rr.Create(ctx, &r)
			if !errors.Is(err, context.Canceled) {
				t.Fatalf(
					"Create(ctx, %v) error = %v, want %v",
					r, err, context.Canceled,
				)
			}
		},
	)
}

func TestUpdate(t *testing.T) {
	t.Run(
		"updates the record and replaces its associations",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			rr := NewRecordRepo(tx)

			ctx := context.Background()
			userID := newTestUser(t, tx, "test@example.com").ID

			created := newTestRecord(userID, time.Now())
			if err := rr.Create(ctx, &created); err != nil {
				t.Fatalf("Create(ctx, %v) error = %v", created, err)
			}

			want := record.Record{
				ID:          created.ID,
				UserID:      userID,
				Title:       record.Title("Updated Movie"),
				ReleaseYear: record.ReleaseYear(1999),
				Runtime:     record.Runtime(90),
				Genres:      []record.Genre{record.GenreHorror},
				Countries:   []record.Country{record.Country("FR")},
				Language:    record.Language("fr"),
				Credits: []record.Credit{
					{
						PersonName: record.PersonName("Updated Director"),
						CreditRole: record.CreditRoleDirector,
					},
				},
				PosterURL: record.PosterURL("https://example.com/updated.jpg"),
				WatchedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				Platform:  record.PlatformTheater,
				Score:     record.Score(5),
				MoodTags:  []record.MoodTag{record.MoodTagDark},
				Memo:      record.Memo("updated memo"),
			}
			if err := rr.Update(ctx, &want); err != nil {
				t.Fatalf("Update(ctx, %v) error = %v", want, err)
			}

			got, err := rr.GetByID(ctx, created.ID)
			if err != nil {
				t.Fatalf(
					"GetByID(ctx, %d) (record.Record, error) = %v, %v",
					created.ID, got, err,
				)
			}
			assertRecordEqual(
				t,
				fmt.Sprintf("GetByID(ctx, %d) (record.Record, error)", created.ID),
				got, &want,
			)

			if !got.CreatedAt.Equal(created.CreatedAt.Truncate(time.Microsecond)) {
				t.Errorf(
					"GetByID(ctx, %d) returns CreatedAt = %v, want %v",
					created.ID, got.CreatedAt, created.CreatedAt.Truncate(time.Microsecond),
				)
			}
			if !got.UpdatedAt.After(created.UpdatedAt) {
				t.Errorf(
					"GetByID(ctx, %d) returns UpdatedAt = %v, want after %v",
					created.ID, got.UpdatedAt, created.UpdatedAt,
				)
			}

			wantCount := int64(
				len(want.Genres) + len(want.Countries) + len(want.Credits) + len(want.MoodTags),
			)
			if count := countAssociations(t, tx, created.ID); count != wantCount {
				t.Errorf(
					"Update(ctx, %v) leaves %d association rows, want %d",
					want, count, wantCount,
				)
			}
		},
	)

	t.Run(
		"returns ErrNotFound when the record does not exist",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			rr := NewRecordRepo(tx)

			ctx := context.Background()
			userID := newTestUser(t, tx, "test@example.com").ID

			r := newTestRecord(userID, time.Now())
			r.ID = record.ID(999999)

			if err := rr.Update(ctx, &r); !errors.Is(err, exception.ErrNotFound) {
				t.Fatalf(
					"Update(ctx, %v) error = %v, want %v",
					r, err, exception.ErrNotFound,
				)
			}
		},
	)

	t.Run(
		"returns ErrNotFound when the record belongs to another user",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			rr := NewRecordRepo(tx)

			ctx := context.Background()
			userID := newTestUser(t, tx, "test@example.com").ID
			otherUserID := newTestUser(t, tx, "other@example.com").ID

			created := newTestRecord(userID, time.Now())
			if err := rr.Create(ctx, &created); err != nil {
				t.Fatalf("Create(ctx, %v) error = %v", created, err)
			}

			hijacked := newTestRecord(otherUserID, time.Now())
			hijacked.ID = created.ID
			hijacked.Title = record.Title("Hijacked Movie")

			if err := rr.Update(ctx, &hijacked); !errors.Is(err, exception.ErrNotFound) {
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
			rr := NewRecordRepo(tx)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			r := newTestRecord(user.ID(1), time.Now())
			r.ID = record.ID(1)

			err := rr.Update(ctx, &r)
			if !errors.Is(err, context.Canceled) {
				t.Fatalf(
					"Update(ctx, %v) error = %v, want %v",
					r, err, context.Canceled,
				)
			}
		},
	)
}

func TestDelete(t *testing.T) {
	t.Run(
		"deletes the record and its associations",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			rr := NewRecordRepo(tx)

			ctx := context.Background()
			userID := newTestUser(t, tx, "test@example.com").ID

			r := newTestRecord(userID, time.Now())
			if err := rr.Create(ctx, &r); err != nil {
				t.Fatalf("Create(ctx, %v) error = %v", r, err)
			}

			if err := rr.Delete(ctx, r.ID); err != nil {
				t.Fatalf("Delete(ctx, %d) error = %v", r.ID, err)
			}

			got, err := rr.GetByID(ctx, r.ID)
			if !errors.Is(err, exception.ErrNotFound) {
				t.Fatalf(
					"GetByID(ctx, %d) (record.Record, error) = %v, %v, want %v",
					r.ID, got, err, exception.ErrNotFound,
				)
			}
			if count := countAssociations(t, tx, r.ID); count != 0 {
				t.Errorf(
					"Delete(ctx, %d) leaves %d association rows, want 0",
					r.ID, count,
				)
			}
		},
	)

	t.Run(
		"returns ErrNotFound when the record does not exist",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			rr := NewRecordRepo(tx)

			ctx := context.Background()
			id := record.ID(999999)

			if err := rr.Delete(ctx, id); !errors.Is(err, exception.ErrNotFound) {
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
			rr := NewRecordRepo(tx)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			id := record.ID(1)
			err := rr.Delete(ctx, id)
			if !errors.Is(err, context.Canceled) {
				t.Fatalf(
					"Delete(ctx, %d) error = %v, want %v",
					id, err, context.Canceled,
				)
			}
		},
	)
}
