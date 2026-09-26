package record

import (
	"cmp"
	"context"
	"errors"
	"fmt"
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

func newEmptyQuery() record.Query {
	return record.NewQuery(
		nil,
		nil,
		nil,
		nil,
		"",
		record.SortFieldWatchedAt,
		record.SortOrderDesc,
		record.Page(1),
		record.PerPage(20),
	)
}

func TestListByUserID(t *testing.T) {
	t.Run(
		"filters, sorts, paginates, and excludes other users",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			rr := NewRecordRepo(tx)

			ctx := context.Background()
			userID := newTestUser(t, tx, "test@example.com").ID
			otherUserID := newTestUser(t, tx, "other@example.com").ID

			// Records that match all filters
			matching := []record.Record{
				{
					UserID:      userID,
					Title:       record.Title("The Matrix"),
					ReleaseYear: record.ReleaseYear(1999),
					Runtime:     record.Runtime(136),
					Genres:      []record.Genre{record.GenreDrama},
					Countries:   []record.Country{record.Country("US")},
					Language:    record.Language("en"),
					Credits:     []record.Credit{{PersonName: "Lana Wachowski", CreditRole: record.CreditRoleDirector}},
					PosterURL:   record.PosterURL("https://example.com/matrix.jpg"),
					WatchedAt:   time.Date(2026, 8, 15, 12, 0, 0, 0, time.UTC),
					Platform:    record.PlatformNetflix,
					Score:       record.Score(5),
					MoodTags:    []record.MoodTag{record.MoodTagTense},
					Memo:        record.Memo(""),
				},
				{
					UserID:      userID,
					Title:       record.Title("The Matrix Reloaded"),
					ReleaseYear: record.ReleaseYear(2003),
					Runtime:     record.Runtime(138),
					Genres:      []record.Genre{record.GenreDrama},
					Countries:   []record.Country{record.Country("US")},
					Language:    record.Language("en"),
					Credits:     []record.Credit{{PersonName: "Lana Wachowski", CreditRole: record.CreditRoleDirector}},
					PosterURL:   record.PosterURL("https://example.com/matrix2.jpg"),
					WatchedAt:   time.Date(2026, 8, 10, 12, 0, 0, 0, time.UTC),
					Platform:    record.PlatformNetflix,
					Score:       record.Score(5),
					MoodTags:    []record.MoodTag{record.MoodTagTense},
					Memo:        record.Memo(""),
				},
				{
					UserID:      userID,
					Title:       record.Title("The Matrix Revolutions"),
					ReleaseYear: record.ReleaseYear(2003),
					Runtime:     record.Runtime(129),
					Genres:      []record.Genre{record.GenreDrama},
					Countries:   []record.Country{record.Country("US")},
					Language:    record.Language("en"),
					Credits:     []record.Credit{{PersonName: "Lana Wachowski", CreditRole: record.CreditRoleDirector}},
					PosterURL:   record.PosterURL("https://example.com/matrix3.jpg"),
					WatchedAt:   time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC),
					Platform:    record.PlatformNetflix,
					Score:       record.Score(5),
					MoodTags:    []record.MoodTag{record.MoodTagTense},
					Memo:        record.Memo(""),
				},
			}
			for i := range matching {
				if err := rr.Create(ctx, &matching[i]); err != nil {
					t.Fatalf("Create(ctx, matching[%d]) error = %v", i, err)
				}
			}

			// Records that fail each filter
			excluded := []record.Record{
				{ // TitleKeyword miss
					UserID:    userID,
					Title:     record.Title("Inception"),
					Genres:    []record.Genre{record.GenreDrama},
					Platform:  record.PlatformNetflix,
					Score:     record.Score(5),
					MoodTags:  []record.MoodTag{record.MoodTagTense},
					WatchedAt: time.Date(2026, 8, 20, 12, 0, 0, 0, time.UTC),
				},
				{ // Score miss
					UserID:    userID,
					Title:     record.Title("The Matrix 4"),
					Genres:    []record.Genre{record.GenreDrama},
					Platform:  record.PlatformNetflix,
					Score:     record.Score(3),
					MoodTags:  []record.MoodTag{record.MoodTagTense},
					WatchedAt: time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC),
				},
				{ // Platform miss
					UserID:    userID,
					Title:     record.Title("The Matrix 5"),
					Genres:    []record.Genre{record.GenreDrama},
					Platform:  record.PlatformTheater,
					Score:     record.Score(5),
					MoodTags:  []record.MoodTag{record.MoodTagTense},
					WatchedAt: time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC),
				},
				{ // MoodTag miss
					UserID:    userID,
					Title:     record.Title("The Matrix 6"),
					Genres:    []record.Genre{record.GenreDrama},
					Platform:  record.PlatformNetflix,
					Score:     record.Score(5),
					MoodTags:  []record.MoodTag{record.MoodTagMoving},
					WatchedAt: time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC),
				},
				{ // Genre miss
					UserID:    userID,
					Title:     record.Title("The Matrix 7"),
					Genres:    []record.Genre{record.GenreHorror},
					Platform:  record.PlatformNetflix,
					Score:     record.Score(5),
					MoodTags:  []record.MoodTag{record.MoodTagTense},
					WatchedAt: time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC),
				},
			}
			for i := range excluded {
				if err := rr.Create(ctx, &excluded[i]); err != nil {
					t.Fatalf("Create(ctx, excluded[%d]) error = %v", i, err)
				}
			}

			// Other user's record (matches all filters but should be excluded)
			other := record.Record{
				UserID:    otherUserID,
				Title:     record.Title("The Matrix"),
				Genres:    []record.Genre{record.GenreDrama},
				Platform:  record.PlatformNetflix,
				Score:     record.Score(5),
				MoodTags:  []record.MoodTag{record.MoodTagTense},
				WatchedAt: time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC),
			}
			if err := rr.Create(ctx, &other); err != nil {
				t.Fatalf("Create(ctx, other) error = %v", err)
			}

			query := record.NewQuery(
				[]record.Score{record.Score(5)},
				[]record.Platform{record.PlatformNetflix},
				[]record.MoodTag{record.MoodTagTense},
				[]record.Genre{record.GenreDrama},
				record.TitleKeyword("Matrix"),
				record.SortFieldWatchedAt,
				record.SortOrderDesc,
				record.Page(1),
				record.PerPage(2),
			)

			got, err := rr.ListByUserID(ctx, userID, query)
			if err != nil {
				t.Fatalf("ListByUserID(ctx, %d, query) error = %v", userID, err)
			}

			if got.TotalCount != 3 {
				t.Errorf("ListByUserID TotalCount = %d, want 3", got.TotalCount)
			}
			if len(got.Records) != 2 {
				t.Fatalf("ListByUserID returns %d records, want 2", len(got.Records))
			}

			// Check sort order (watched_at desc)
			if got.Records[0].ID != matching[0].ID {
				t.Errorf("ListByUserID Records[0].ID = %d, want %d", got.Records[0].ID, matching[0].ID)
			}
			if got.Records[1].ID != matching[1].ID {
				t.Errorf("ListByUserID Records[1].ID = %d, want %d", got.Records[1].ID, matching[1].ID)
			}

			// Check associations are preloaded
			for i, r := range got.Records {
				if len(r.Genres) == 0 || len(r.Credits) == 0 {
					t.Errorf("ListByUserID Records[%d] = %v, want preloaded associations", i, r)
				}
			}
		},
	)

	t.Run(
		"orders records with the same sort value by id desc",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			rr := NewRecordRepo(tx)

			ctx := context.Background()
			userID := newTestUser(t, tx, "test@example.com").ID
			watchedAt := time.Date(2026, 8, 15, 12, 0, 0, 0, time.UTC)

			first := newTestRecord(userID, watchedAt)
			if err := rr.Create(ctx, &first); err != nil {
				t.Fatalf("Create(ctx, %v) error = %v", first, err)
			}
			second := newTestRecord(userID, watchedAt)
			if err := rr.Create(ctx, &second); err != nil {
				t.Fatalf("Create(ctx, %v) error = %v", second, err)
			}

			got, err := rr.ListByUserID(ctx, userID, newEmptyQuery())
			if err != nil {
				t.Fatalf("ListByUserID(ctx, %d, query) error = %v", userID, err)
			}
			if len(got.Records) != 2 {
				t.Fatalf("ListByUserID returns %d records, want 2", len(got.Records))
			}

			if got.Records[0].ID != second.ID {
				t.Errorf("ListByUserID Records[0].ID = %d, want %d", got.Records[0].ID, second.ID)
			}
			if got.Records[1].ID != first.ID {
				t.Errorf("ListByUserID Records[1].ID = %d, want %d", got.Records[1].ID, first.ID)
			}
		},
	)

	t.Run(
		"returns the records of the given page",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			rr := NewRecordRepo(tx)

			ctx := context.Background()
			userID := newTestUser(t, tx, "test@example.com").ID

			records := []record.Record{
				newTestRecord(userID, time.Date(2026, 8, 15, 12, 0, 0, 0, time.UTC)),
				newTestRecord(userID, time.Date(2026, 8, 10, 12, 0, 0, 0, time.UTC)),
				newTestRecord(userID, time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)),
			}
			for i := range records {
				if err := rr.Create(ctx, &records[i]); err != nil {
					t.Fatalf("Create(ctx, records[%d]) error = %v", i, err)
				}
			}

			query := record.NewQuery(
				nil,
				nil,
				nil,
				nil,
				"",
				record.SortFieldWatchedAt,
				record.SortOrderDesc,
				record.Page(2),
				record.PerPage(2),
			)

			got, err := rr.ListByUserID(ctx, userID, query)
			if err != nil {
				t.Fatalf("ListByUserID(ctx, %d, query) error = %v", userID, err)
			}

			if got.TotalCount != 3 {
				t.Errorf("ListByUserID TotalCount = %d, want 3", got.TotalCount)
			}
			if len(got.Records) != 1 {
				t.Fatalf("ListByUserID returns %d records, want 1", len(got.Records))
			}
			if got.Records[0].ID != records[2].ID {
				t.Errorf("ListByUserID Records[0].ID = %d, want %d", got.Records[0].ID, records[2].ID)
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

			got, err := rr.ListByUserID(ctx, userID, newEmptyQuery())
			if err != nil {
				t.Fatalf(
					"ListByUserID(ctx, %d, query) error = %v",
					userID, err,
				)
			}
			if len(got.Records) != 0 {
				t.Errorf(
					"ListByUserID(ctx, %d, query) returns %d records, want 0",
					userID, len(got.Records),
				)
			}
			if got.TotalCount != 0 {
				t.Errorf(
					"ListByUserID(ctx, %d, query) TotalCount = %d, want 0",
					userID, got.TotalCount,
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
			got, err := rr.ListByUserID(ctx, userID, newEmptyQuery())

			if !errors.Is(err, context.Canceled) {
				t.Fatalf(
					"ListByUserID(ctx, %d, query) error = %v, want %v",
					userID, err, context.Canceled,
				)
			}
			if len(got.Records) != 0 {
				t.Errorf(
					"ListByUserID(ctx, %d, query) Records = %v, want empty",
					userID, got.Records,
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
