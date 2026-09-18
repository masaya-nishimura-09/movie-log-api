package record

import (
	"context"
	"errors"
	"testing"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/media"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/record"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
)

type fakeRepository struct {
	record    *record.Record
	created   *record.Record
	updated   *record.Record
	deletedID record.ID
}

func (r *fakeRepository) GetByID(
	ctx context.Context,
	recordID record.ID,
) (*record.Record, error) {
	return r.record, nil
}

func (r *fakeRepository) ListByUserID(
	ctx context.Context,
	userID user.ID,
) ([]*record.Record, error) {
	return nil, nil
}

func (r *fakeRepository) Create(ctx context.Context, rec *record.Record) error {
	r.created = rec
	return nil
}

func (r *fakeRepository) Update(ctx context.Context, rec *record.Record) error {
	r.updated = rec
	return nil
}

func (r *fakeRepository) Delete(ctx context.Context, recordID record.ID) error {
	r.deletedID = recordID
	return nil
}

type fakeMediaService struct {
	deletedURL media.URL
	err        error
}

func (s *fakeMediaService) Upload(
	ctx context.Context,
	userID user.ID,
	m media.Media,
) (media.URL, error) {
	return "", s.err
}

func (s *fakeMediaService) Delete(
	ctx context.Context,
	userID user.ID,
	url media.URL,
) error {
	s.deletedURL = url
	return s.err
}

func TestGetByID(t *testing.T) {
	userID := user.ID(1)
	recordID := record.ID(10)

	t.Run(
		"returns the record when it belongs to the user",
		func(t *testing.T) {
			repo := &fakeRepository{
				record: &record.Record{ID: recordID, UserID: userID},
			}
			mediaService := &fakeMediaService{}
			ru := NewRecordUsecase(repo, mediaService)

			ctx := context.Background()

			got, err := ru.GetByID(ctx, userID, recordID)
			if err != nil {
				t.Fatalf(
					"GetByID(ctx, %v, %v) (*record.Record, error) = %v, %v",
					userID, recordID, got, err,
				)
			}
			if got.ID != recordID {
				t.Errorf(
					"GetByID(ctx, %v, %v) ID = %v, want %v",
					userID, recordID, got.ID, recordID,
				)
			}
		},
	)

	t.Run(
		"returns ErrNotFound when the record belongs to another user",
		func(t *testing.T) {
			repo := &fakeRepository{
				record: &record.Record{ID: recordID, UserID: user.ID(2)},
			}
			mediaService := &fakeMediaService{}
			ru := NewRecordUsecase(repo, mediaService)

			ctx := context.Background()

			got, err := ru.GetByID(ctx, userID, recordID)
			if !errors.Is(err, exception.ErrNotFound) {
				t.Fatalf(
					"GetByID(ctx, %v, %v) (*record.Record, error) = %v, %v, want %v",
					userID, recordID, got, err, exception.ErrNotFound,
				)
			}
			if got != nil {
				t.Errorf(
					"GetByID(ctx, %v, %v) (*record.Record, error) = %v, %v, want nil",
					userID, recordID, got, err,
				)
			}
		},
	)
}

func TestCreate(t *testing.T) {
	t.Run(
		"overwrites the user id with the given user id",
		func(t *testing.T) {
			repo := &fakeRepository{}
			mediaService := &fakeMediaService{}
			ru := NewRecordUsecase(repo, mediaService)

			ctx := context.Background()
			userID := user.ID(1)
			rec := record.Record{UserID: user.ID(2)}

			got, err := ru.Create(ctx, userID, rec)
			if err != nil {
				t.Fatalf(
					"Create(ctx, %v, %v) (*record.Record, error) = %v, %v",
					userID, rec, got, err,
				)
			}
			if got.UserID != userID {
				t.Errorf(
					"Create(ctx, %v, %v) UserID = %v, want %v",
					userID, rec, got.UserID, userID,
				)
			}
			if repo.created.UserID != userID {
				t.Errorf(
					"Create(ctx, %v, %v) creates a record of user %v, want %v",
					userID, rec, repo.created.UserID, userID,
				)
			}
		},
	)
}

func TestUpdate(t *testing.T) {
	userID := user.ID(1)
	recordID := record.ID(10)

	t.Run(
		"overwrites the record id and the user id with the given ids",
		func(t *testing.T) {
			repo := &fakeRepository{
				record: &record.Record{ID: recordID, UserID: userID},
			}
			mediaService := &fakeMediaService{}
			ru := NewRecordUsecase(repo, mediaService)

			ctx := context.Background()
			rec := record.Record{ID: record.ID(99), UserID: user.ID(2)}

			got, err := ru.Update(ctx, userID, recordID, rec)
			if err != nil {
				t.Fatalf(
					"Update(ctx, %v, %v, %v) (*record.Record, error) = %v, %v",
					userID, recordID, rec, got, err,
				)
			}
			if got.ID != recordID || got.UserID != userID {
				t.Errorf(
					"Update(ctx, %v, %v, %v) ID = %v, UserID = %v, want %v, %v",
					userID, recordID, rec, got.ID, got.UserID, recordID, userID,
				)
			}
			if repo.updated.ID != recordID || repo.updated.UserID != userID {
				t.Errorf(
					"Update(ctx, %v, %v, %v) updates ID = %v, UserID = %v, want %v, %v",
					userID, recordID, rec,
					repo.updated.ID, repo.updated.UserID, recordID, userID,
				)
			}
		},
	)

	t.Run(
		"deletes the old poster when the poster url changes",
		func(t *testing.T) {
			oldPosterURL := record.PosterURL("https://example.com/1/old.jpg")
			repo := &fakeRepository{
				record: &record.Record{
					ID:        recordID,
					UserID:    userID,
					PosterURL: oldPosterURL,
				},
			}
			mediaService := &fakeMediaService{}
			ru := NewRecordUsecase(repo, mediaService)

			ctx := context.Background()
			rec := record.Record{
				PosterURL: record.PosterURL("https://example.com/1/new.jpg"),
			}

			got, err := ru.Update(ctx, userID, recordID, rec)
			if err != nil {
				t.Fatalf(
					"Update(ctx, %v, %v, %v) (*record.Record, error) = %v, %v",
					userID, recordID, rec, got, err,
				)
			}
			if mediaService.deletedURL != media.URL(oldPosterURL) {
				t.Errorf(
					"Update(ctx, %v, %v, %v) deletes media %v, want %v",
					userID, recordID, rec, mediaService.deletedURL, oldPosterURL,
				)
			}
		},
	)

	t.Run(
		"does not delete the poster when the poster url does not change",
		func(t *testing.T) {
			posterURL := record.PosterURL("https://example.com/1/poster.jpg")
			repo := &fakeRepository{
				record: &record.Record{
					ID:        recordID,
					UserID:    userID,
					PosterURL: posterURL,
				},
			}
			mediaService := &fakeMediaService{}
			ru := NewRecordUsecase(repo, mediaService)

			ctx := context.Background()
			rec := record.Record{PosterURL: posterURL}

			got, err := ru.Update(ctx, userID, recordID, rec)
			if err != nil {
				t.Fatalf(
					"Update(ctx, %v, %v, %v) (*record.Record, error) = %v, %v",
					userID, recordID, rec, got, err,
				)
			}
			if mediaService.deletedURL != "" {
				t.Errorf(
					"Update(ctx, %v, %v, %v) deletes poster %v, want no deletion",
					userID, recordID, rec, mediaService.deletedURL,
				)
			}
		},
	)
}

func TestDelete(t *testing.T) {
	userID := user.ID(1)
	recordID := record.ID(10)

	t.Run(
		"deletes the record when it belongs to the user",
		func(t *testing.T) {
			repo := &fakeRepository{
				record: &record.Record{ID: recordID, UserID: userID},
			}
			mediaService := &fakeMediaService{}
			ru := NewRecordUsecase(repo, mediaService)

			ctx := context.Background()

			if err := ru.Delete(ctx, userID, recordID); err != nil {
				t.Fatalf(
					"Delete(ctx, %v, %v) error = %v",
					userID, recordID, err,
				)
			}
			if repo.deletedID != recordID {
				t.Errorf(
					"Delete(ctx, %v, %v) deletes record %v, want %v",
					userID, recordID, repo.deletedID, recordID,
				)
			}
		},
	)

	t.Run(
		"returns ErrNotFound when the record belongs to another user",
		func(t *testing.T) {
			repo := &fakeRepository{
				record: &record.Record{ID: recordID, UserID: user.ID(2)},
			}
			mediaService := &fakeMediaService{}
			ru := NewRecordUsecase(repo, mediaService)

			ctx := context.Background()

			err := ru.Delete(ctx, userID, recordID)
			if !errors.Is(err, exception.ErrNotFound) {
				t.Fatalf(
					"Delete(ctx, %v, %v) error = %v, want %v",
					userID, recordID, err, exception.ErrNotFound,
				)
			}
			if repo.deletedID != 0 {
				t.Errorf(
					"Delete(ctx, %v, %v) deletes record %v, want no deletion",
					userID, recordID, repo.deletedID,
				)
			}
		},
	)

	t.Run(
		"deletes the poster of the record",
		func(t *testing.T) {
			posterURL := record.PosterURL("https://example.com/1/poster.jpg")
			repo := &fakeRepository{
				record: &record.Record{
					ID:        recordID,
					UserID:    userID,
					PosterURL: posterURL,
				},
			}
			mediaService := &fakeMediaService{}
			ru := NewRecordUsecase(repo, mediaService)

			ctx := context.Background()

			if err := ru.Delete(ctx, userID, recordID); err != nil {
				t.Fatalf(
					"Delete(ctx, %v, %v) error = %v",
					userID, recordID, err,
				)
			}
			if mediaService.deletedURL != media.URL(posterURL) {
				t.Errorf(
					"Delete(ctx, %v, %v) deletes media %v, want %v",
					userID, recordID, mediaService.deletedURL, posterURL,
				)
			}
		},
	)

	t.Run(
		"succeeds even when deleting the media fails",
		func(t *testing.T) {
			repo := &fakeRepository{
				record: &record.Record{ID: recordID, UserID: userID},
			}
			mediaService := &fakeMediaService{
				err: errors.New("delete media"),
			}
			ru := NewRecordUsecase(repo, mediaService)

			ctx := context.Background()

			if err := ru.Delete(ctx, userID, recordID); err != nil {
				t.Fatalf(
					"Delete(ctx, %v, %v) error = %v, want nil",
					userID, recordID, err,
				)
			}
		},
	)
}
