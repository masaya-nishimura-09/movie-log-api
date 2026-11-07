package record

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/media"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/record"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
)

type Usecase interface {
	GetByID(
		ctx context.Context,
		userID user.ID,
		recordID record.ID,
	) (*record.Record, error)
	ListByUserID(
		ctx context.Context,
		userID user.ID,
		query record.Query,
	) (record.ListResult, error)
	CountMoodTagsByUserID(
		ctx context.Context,
		userID user.ID,
	) (record.MoodTagCounts, error)
	Create(
		ctx context.Context,
		userID user.ID,
		r record.Record,
	) (*record.Record, error)
	Update(
		ctx context.Context,
		userID user.ID,
		recordID record.ID,
		r record.Record,
	) (*record.Record, error)
	Delete(ctx context.Context, userID user.ID, recordID record.ID) error
}

type RecordUsecase struct {
	recordRepo     record.Repository
	attachmentRepo record.AttachmentRepository
	mediaService   media.Service
}

func NewRecordUsecase(
	recordRepo record.Repository,
	attachmentRepo record.AttachmentRepository,
	mediaService media.Service,
) *RecordUsecase {
	return &RecordUsecase{
		recordRepo:     recordRepo,
		attachmentRepo: attachmentRepo,
		mediaService:   mediaService,
	}
}

func (ru *RecordUsecase) GetByID(
	ctx context.Context,
	userID user.ID,
	recordID record.ID,
) (*record.Record, error) {
	r, err := ru.recordRepo.GetByID(ctx, recordID)
	if err != nil {
		return nil, fmt.Errorf("get record by id: %w", err)
	}

	if r.UserID != userID {
		return nil, exception.ErrNotFound
	}

	return r, nil
}

func (ru *RecordUsecase) ListByUserID(
	ctx context.Context,
	userID user.ID,
	query record.Query,
) (record.ListResult, error) {
	listResult, err := ru.recordRepo.ListByUserID(ctx, userID, query)
	if err != nil {
		return record.ListResult{}, fmt.Errorf("list records by user id: %w", err)
	}

	return listResult, nil
}

func (ru *RecordUsecase) CountMoodTagsByUserID(
	ctx context.Context,
	userID user.ID,
) (record.MoodTagCounts, error) {
	moodTagCounts, err := ru.recordRepo.CountMoodTagsByUserID(ctx, userID)
	if err != nil {
		return record.MoodTagCounts{}, fmt.Errorf("count mood tags by user id: %w", err)
	}

	return moodTagCounts, nil
}

func (ru *RecordUsecase) Create(
	ctx context.Context,
	userID user.ID,
	r record.Record,
) (*record.Record, error) {
	r.UserID = userID

	if err := ru.recordRepo.Create(ctx, &r); err != nil {
		return nil, fmt.Errorf("create record: %w", err)
	}

	return &r, nil
}

func (ru *RecordUsecase) Update(
	ctx context.Context,
	userID user.ID,
	recordID record.ID,
	r record.Record,
) (*record.Record, error) {
	current, err := ru.recordRepo.GetByID(ctx, recordID)
	if err != nil {
		return nil, fmt.Errorf("get record by id: %w", err)
	}

	if current.UserID != userID {
		return nil, exception.ErrNotFound
	}

	r.ID = recordID
	r.UserID = userID
	r.CreatedAt = current.CreatedAt

	if err := ru.recordRepo.Update(ctx, &r); err != nil {
		return nil, fmt.Errorf("update record: %w", err)
	}

	if current.PosterURL != r.PosterURL {
		if err := ru.mediaService.Delete(ctx, userID, media.URL(current.PosterURL)); err != nil {
			slog.WarnContext(ctx, "media_delete_failed", "userId", fmt.Sprint(userID), "message", err.Error())
		}
	}

	return &r, nil
}

func (ru *RecordUsecase) Delete(
	ctx context.Context,
	userID user.ID,
	recordID record.ID,
) error {
	r, err := ru.recordRepo.GetByID(ctx, recordID)
	if err != nil {
		return fmt.Errorf("get record by id: %w", err)
	}

	if r.UserID != userID {
		return exception.ErrNotFound
	}

	attachments, err := ru.attachmentRepo.ListByRecordID(ctx, recordID)
	if err != nil {
		return fmt.Errorf("list attachments by record id: %w", err)
	}

	if err := ru.recordRepo.Delete(ctx, recordID); err != nil {
		return fmt.Errorf("delete record: %w", err)
	}

	urls := make([]media.URL, 0, len(attachments)+1)
	urls = append(urls, media.URL(r.PosterURL))
	for _, a := range attachments {
		urls = append(urls, media.URL(a.ImageURL))
	}
	for _, url := range urls {
		if err := ru.mediaService.Delete(ctx, userID, url); err != nil {
			slog.WarnContext(ctx, "media_delete_failed", "userId", fmt.Sprint(userID), "message", err.Error())
		}
	}

	return nil
}
