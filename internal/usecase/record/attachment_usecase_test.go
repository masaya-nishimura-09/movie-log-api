package record

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/media"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/record"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
)

type fakeAttachmentRepository struct {
	attachment  *record.Attachment
	attachments []record.Attachment
	count       uint
	createErr   error
	created     *record.Attachment
	deletedID   record.AttachmentID
	deleteCalls int
}

func (r *fakeAttachmentRepository) GetByID(
	ctx context.Context,
	attachmentID record.AttachmentID,
) (*record.Attachment, error) {
	return r.attachment, nil
}

func (r *fakeAttachmentRepository) ListByRecordID(
	ctx context.Context,
	recordID record.ID,
) ([]record.Attachment, error) {
	return r.attachments, nil
}

func (r *fakeAttachmentRepository) CountByRecordID(
	ctx context.Context,
	recordID record.ID,
) (uint, error) {
	return r.count, nil
}

func (r *fakeAttachmentRepository) Create(ctx context.Context, a *record.Attachment) error {
	r.created = a
	return r.createErr
}

func (r *fakeAttachmentRepository) Delete(ctx context.Context, attachmentID record.AttachmentID) error {
	r.deletedID = attachmentID
	r.deleteCalls++
	return nil
}

func TestAttachmentListByRecordID(t *testing.T) {
	userID := user.ID(1)
	recordID := record.ID(10)

	t.Run(
		"returns the attachments when the record belongs to the user",
		func(t *testing.T) {
			want := []record.Attachment{{ID: 1, UserID: userID, RecordID: recordID}}
			repo := &fakeRepository{record: &record.Record{ID: recordID, UserID: userID}}
			attachmentRepo := &fakeAttachmentRepository{attachments: want}
			au := NewAttachmentUsecase(attachmentRepo, repo, &fakeMediaService{})

			ctx := context.Background()

			got, err := au.ListByRecordID(ctx, userID, recordID)
			if err != nil || !slices.Equal(got, want) {
				t.Fatalf(
					"ListByRecordID(ctx, %v, %v) ([]record.Attachment, error) = %v, %v, want %v, nil",
					userID, recordID, got, err, want,
				)
			}
		},
	)

	t.Run(
		"returns ErrNotFound when the record belongs to another user",
		func(t *testing.T) {
			repo := &fakeRepository{record: &record.Record{ID: recordID, UserID: user.ID(2)}}
			au := NewAttachmentUsecase(&fakeAttachmentRepository{}, repo, &fakeMediaService{})

			ctx := context.Background()

			got, err := au.ListByRecordID(ctx, userID, recordID)
			if !errors.Is(err, exception.ErrNotFound) {
				t.Fatalf(
					"ListByRecordID(ctx, %v, %v) ([]record.Attachment, error) = %v, %v, want %v",
					userID, recordID, got, err, exception.ErrNotFound,
				)
			}
		},
	)
}

func TestAttachmentCreate(t *testing.T) {
	userID := user.ID(1)
	recordID := record.ID(10)
	m := media.Media{Data: media.Data("\xFF\xD8\xFF"), ContentType: media.ContentTypeJPEG}
	url := media.URL("https://example.com/1/image.jpg")

	t.Run(
		"uploads the image and saves the attachment",
		func(t *testing.T) {
			repo := &fakeRepository{record: &record.Record{ID: recordID, UserID: userID}}
			attachmentRepo := &fakeAttachmentRepository{count: record.MaxAttachmentsPerRecord - 1}
			mediaService := &fakeMediaService{uploadURL: url}
			au := NewAttachmentUsecase(attachmentRepo, repo, mediaService)

			ctx := context.Background()

			got, err := au.Create(ctx, userID, recordID, m)
			if err != nil {
				t.Fatalf(
					"Create(ctx, %v, %v, m) (*record.Attachment, error) = %v, %v",
					userID, recordID, got, err,
				)
			}
			want := record.NewAttachment(userID, recordID, record.AttachmentImageURL(url))
			if attachmentRepo.created == nil || *attachmentRepo.created != want {
				t.Errorf(
					"Create(ctx, %v, %v, m) saves %v, want %v",
					userID, recordID, attachmentRepo.created, want,
				)
			}
		},
	)

	t.Run(
		"returns ErrInvalid without uploading when the record already has the maximum number of attachments",
		func(t *testing.T) {
			repo := &fakeRepository{record: &record.Record{ID: recordID, UserID: userID}}
			attachmentRepo := &fakeAttachmentRepository{count: record.MaxAttachmentsPerRecord}
			mediaService := &fakeMediaService{uploadURL: url}
			au := NewAttachmentUsecase(attachmentRepo, repo, mediaService)

			ctx := context.Background()

			got, err := au.Create(ctx, userID, recordID, m)
			if !errors.Is(err, exception.ErrInvalid) {
				t.Fatalf(
					"Create(ctx, %v, %v, m) (*record.Attachment, error) = %v, %v, want %v",
					userID, recordID, got, err, exception.ErrInvalid,
				)
			}
			if mediaService.uploadCalls != 0 || attachmentRepo.created != nil {
				t.Errorf(
					"Create(ctx, %v, %v, m) uploads %d times and saves %v, want nothing",
					userID, recordID, mediaService.uploadCalls, attachmentRepo.created,
				)
			}
		},
	)

	t.Run(
		"returns ErrNotFound without uploading when the record belongs to another user",
		func(t *testing.T) {
			repo := &fakeRepository{record: &record.Record{ID: recordID, UserID: user.ID(2)}}
			mediaService := &fakeMediaService{uploadURL: url}
			au := NewAttachmentUsecase(&fakeAttachmentRepository{}, repo, mediaService)

			ctx := context.Background()

			got, err := au.Create(ctx, userID, recordID, m)
			if !errors.Is(err, exception.ErrNotFound) {
				t.Fatalf(
					"Create(ctx, %v, %v, m) (*record.Attachment, error) = %v, %v, want %v",
					userID, recordID, got, err, exception.ErrNotFound,
				)
			}
			if mediaService.uploadCalls != 0 {
				t.Errorf(
					"Create(ctx, %v, %v, m) uploads %d times, want 0",
					userID, recordID, mediaService.uploadCalls,
				)
			}
		},
	)

	t.Run(
		"deletes the uploaded image when saving the attachment fails",
		func(t *testing.T) {
			repo := &fakeRepository{record: &record.Record{ID: recordID, UserID: userID}}
			attachmentRepo := &fakeAttachmentRepository{createErr: errors.New("create attachment")}
			mediaService := &fakeMediaService{uploadURL: url}
			au := NewAttachmentUsecase(attachmentRepo, repo, mediaService)

			ctx := context.Background()

			if _, err := au.Create(ctx, userID, recordID, m); err == nil {
				t.Fatalf("Create(ctx, %v, %v, m) error = nil, want an error", userID, recordID)
			}
			if mediaService.deletedURL != url {
				t.Errorf(
					"Create(ctx, %v, %v, m) deletes media %v, want %v",
					userID, recordID, mediaService.deletedURL, url,
				)
			}
		},
	)
}

func TestAttachmentDelete(t *testing.T) {
	userID := user.ID(1)
	recordID := record.ID(10)
	attachmentID := record.AttachmentID(5)
	imageURL := record.AttachmentImageURL("https://example.com/1/image.jpg")

	t.Run(
		"deletes the attachment and its image",
		func(t *testing.T) {
			attachmentRepo := &fakeAttachmentRepository{
				attachment: &record.Attachment{
					ID: attachmentID, UserID: userID, RecordID: recordID, ImageURL: imageURL,
				},
			}
			mediaService := &fakeMediaService{}
			au := NewAttachmentUsecase(attachmentRepo, &fakeRepository{}, mediaService)

			ctx := context.Background()

			if err := au.Delete(ctx, userID, recordID, attachmentID); err != nil {
				t.Fatalf(
					"Delete(ctx, %v, %v, %v) error = %v",
					userID, recordID, attachmentID, err,
				)
			}
			if attachmentRepo.deletedID != attachmentID {
				t.Errorf(
					"Delete(ctx, %v, %v, %v) deletes attachment %v, want %v",
					userID, recordID, attachmentID, attachmentRepo.deletedID, attachmentID,
				)
			}
			if mediaService.deletedURL != media.URL(imageURL) {
				t.Errorf(
					"Delete(ctx, %v, %v, %v) deletes media %v, want %v",
					userID, recordID, attachmentID, mediaService.deletedURL, imageURL,
				)
			}
		},
	)

	t.Run(
		"succeeds even when deleting the image fails",
		func(t *testing.T) {
			attachmentRepo := &fakeAttachmentRepository{
				attachment: &record.Attachment{
					ID: attachmentID, UserID: userID, RecordID: recordID, ImageURL: imageURL,
				},
			}
			mediaService := &fakeMediaService{err: errors.New("delete media")}
			au := NewAttachmentUsecase(attachmentRepo, &fakeRepository{}, mediaService)

			ctx := context.Background()

			if err := au.Delete(ctx, userID, recordID, attachmentID); err != nil {
				t.Fatalf(
					"Delete(ctx, %v, %v, %v) error = %v, want nil",
					userID, recordID, attachmentID, err,
				)
			}
		},
	)

	for _, tt := range []struct {
		name       string
		attachment record.Attachment
	}{
		{
			"returns ErrNotFound when the attachment belongs to another user",
			record.Attachment{ID: attachmentID, UserID: user.ID(2), RecordID: recordID},
		},
		{
			"returns ErrNotFound when the attachment belongs to another record",
			record.Attachment{ID: attachmentID, UserID: userID, RecordID: record.ID(11)},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			attachmentRepo := &fakeAttachmentRepository{attachment: &tt.attachment}
			mediaService := &fakeMediaService{}
			au := NewAttachmentUsecase(attachmentRepo, &fakeRepository{}, mediaService)

			ctx := context.Background()

			err := au.Delete(ctx, userID, recordID, attachmentID)
			if !errors.Is(err, exception.ErrNotFound) {
				t.Fatalf(
					"Delete(ctx, %v, %v, %v) error = %v, want %v",
					userID, recordID, attachmentID, err, exception.ErrNotFound,
				)
			}
			if attachmentRepo.deleteCalls != 0 || len(mediaService.deletedURLs) != 0 {
				t.Errorf(
					"Delete(ctx, %v, %v, %v) deletes the attachment or its image, want no deletion",
					userID, recordID, attachmentID,
				)
			}
		})
	}
}
