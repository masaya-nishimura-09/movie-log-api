package record

import (
	"context"
	"errors"
	"testing"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
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

type fakePosterService struct {
	uploadedPoster record.Poster
	uploadedURL    record.PosterURL
	deletedURL     record.PosterURL
	err            error
}

func (r *fakePosterService) Upload(
	ctx context.Context,
	userID user.ID,
	poster record.Poster,
) (record.PosterURL, error) {
	r.uploadedPoster = poster
	return r.uploadedURL, r.err
}

func (r *fakePosterService) Delete(
	ctx context.Context,
	userID user.ID,
	url record.PosterURL,
) error {
	r.deletedURL = url
	return r.err
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
			posterService := &fakePosterService{}
			ru := NewRecordUsecase(repo, posterService)

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
			posterService := &fakePosterService{}
			ru := NewRecordUsecase(repo, posterService)

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

func TestCreateRecord(t *testing.T) {
	t.Run(
		"overwrites the user id with the given user id",
		func(t *testing.T) {
			repo := &fakeRepository{}
			posterService := &fakePosterService{}
			ru := NewRecordUsecase(repo, posterService)

			ctx := context.Background()
			userID := user.ID(1)
			rec := record.Record{UserID: user.ID(2)}

			got, err := ru.CreateRecord(ctx, userID, rec)
			if err != nil {
				t.Fatalf(
					"CreateRecord(ctx, %v, %v) (*record.Record, error) = %v, %v",
					userID, rec, got, err,
				)
			}
			if got.UserID != userID {
				t.Errorf(
					"CreateRecord(ctx, %v, %v) UserID = %v, want %v",
					userID, rec, got.UserID, userID,
				)
			}
			if repo.created.UserID != userID {
				t.Errorf(
					"CreateRecord(ctx, %v, %v) creates a record of user %v, want %v",
					userID, rec, repo.created.UserID, userID,
				)
			}
		},
	)
}

func TestUpdateRecord(t *testing.T) {
	userID := user.ID(1)
	recordID := record.ID(10)

	t.Run(
		"overwrites the record id and the user id with the given ids",
		func(t *testing.T) {
			repo := &fakeRepository{
				record: &record.Record{ID: recordID, UserID: userID},
			}
			posterService := &fakePosterService{}
			ru := NewRecordUsecase(repo, posterService)

			ctx := context.Background()
			rec := record.Record{ID: record.ID(99), UserID: user.ID(2)}

			got, err := ru.UpdateRecord(ctx, userID, recordID, rec)
			if err != nil {
				t.Fatalf(
					"UpdateRecord(ctx, %v, %v, %v) (*record.Record, error) = %v, %v",
					userID, recordID, rec, got, err,
				)
			}
			if got.ID != recordID || got.UserID != userID {
				t.Errorf(
					"UpdateRecord(ctx, %v, %v, %v) ID = %v, UserID = %v, want %v, %v",
					userID, recordID, rec, got.ID, got.UserID, recordID, userID,
				)
			}
			if repo.updated.ID != recordID || repo.updated.UserID != userID {
				t.Errorf(
					"UpdateRecord(ctx, %v, %v, %v) updates ID = %v, UserID = %v, want %v, %v",
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
			posterService := &fakePosterService{}
			ru := NewRecordUsecase(repo, posterService)

			ctx := context.Background()
			rec := record.Record{
				PosterURL: record.PosterURL("https://example.com/1/new.jpg"),
			}

			got, err := ru.UpdateRecord(ctx, userID, recordID, rec)
			if err != nil {
				t.Fatalf(
					"UpdateRecord(ctx, %v, %v, %v) (*record.Record, error) = %v, %v",
					userID, recordID, rec, got, err,
				)
			}
			if posterService.deletedURL != oldPosterURL {
				t.Errorf(
					"UpdateRecord(ctx, %v, %v, %v) deletes poster %v, want %v",
					userID, recordID, rec, posterService.deletedURL, oldPosterURL,
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
			posterService := &fakePosterService{}
			ru := NewRecordUsecase(repo, posterService)

			ctx := context.Background()
			rec := record.Record{PosterURL: posterURL}

			got, err := ru.UpdateRecord(ctx, userID, recordID, rec)
			if err != nil {
				t.Fatalf(
					"UpdateRecord(ctx, %v, %v, %v) (*record.Record, error) = %v, %v",
					userID, recordID, rec, got, err,
				)
			}
			if posterService.deletedURL != "" {
				t.Errorf(
					"UpdateRecord(ctx, %v, %v, %v) deletes poster %v, want no deletion",
					userID, recordID, rec, posterService.deletedURL,
				)
			}
		},
	)
}

func TestDeleteRecord(t *testing.T) {
	userID := user.ID(1)
	recordID := record.ID(10)

	t.Run(
		"deletes the record when it belongs to the user",
		func(t *testing.T) {
			repo := &fakeRepository{
				record: &record.Record{ID: recordID, UserID: userID},
			}
			posterService := &fakePosterService{}
			ru := NewRecordUsecase(repo, posterService)

			ctx := context.Background()

			if err := ru.DeleteRecord(ctx, userID, recordID); err != nil {
				t.Fatalf(
					"DeleteRecord(ctx, %v, %v) error = %v",
					userID, recordID, err,
				)
			}
			if repo.deletedID != recordID {
				t.Errorf(
					"DeleteRecord(ctx, %v, %v) deletes record %v, want %v",
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
			posterService := &fakePosterService{}
			ru := NewRecordUsecase(repo, posterService)

			ctx := context.Background()

			err := ru.DeleteRecord(ctx, userID, recordID)
			if !errors.Is(err, exception.ErrNotFound) {
				t.Fatalf(
					"DeleteRecord(ctx, %v, %v) error = %v, want %v",
					userID, recordID, err, exception.ErrNotFound,
				)
			}
			if repo.deletedID != 0 {
				t.Errorf(
					"DeleteRecord(ctx, %v, %v) deletes record %v, want no deletion",
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
			posterService := &fakePosterService{}
			ru := NewRecordUsecase(repo, posterService)

			ctx := context.Background()

			if err := ru.DeleteRecord(ctx, userID, recordID); err != nil {
				t.Fatalf(
					"DeleteRecord(ctx, %v, %v) error = %v",
					userID, recordID, err,
				)
			}
			if posterService.deletedURL != posterURL {
				t.Errorf(
					"DeleteRecord(ctx, %v, %v) deletes poster %v, want %v",
					userID, recordID, posterService.deletedURL, posterURL,
				)
			}
		},
	)

	t.Run(
		"returns an error when deleting the poster fails",
		func(t *testing.T) {
			repo := &fakeRepository{
				record: &record.Record{ID: recordID, UserID: userID},
			}
			posterService := &fakePosterService{
				err: errors.New("delete poster"),
			}
			ru := NewRecordUsecase(repo, posterService)

			ctx := context.Background()

			if err := ru.DeleteRecord(ctx, userID, recordID); err == nil {
				t.Fatalf(
					"DeleteRecord(ctx, %v, %v) error = nil, want an error",
					userID, recordID,
				)
			}
		},
	)
}

func TestUploadPoster(t *testing.T) {
	t.Run(
		"returns the url of the uploaded poster",
		func(t *testing.T) {
			posterURL := record.PosterURL("https://example.com/1/poster.jpg")
			repo := &fakeRepository{}
			posterService := &fakePosterService{uploadedURL: posterURL}
			ru := NewRecordUsecase(repo, posterService)

			ctx := context.Background()
			userID := user.ID(1)
			poster := record.Poster{
				Data:        record.PosterData("poster"),
				ContentType: record.PosterContentTypeJPEG,
			}

			got, err := ru.UploadPoster(ctx, userID, poster)
			if err != nil {
				t.Fatalf(
					"UploadPoster(ctx, %v, %v) (record.PosterURL, error) = %v, %v",
					userID, poster, got, err,
				)
			}
			if got != posterURL {
				t.Errorf(
					"UploadPoster(ctx, %v, %v) = %v, want %v",
					userID, poster, got, posterURL,
				)
			}
		},
	)
}
