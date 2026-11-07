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

type AttachmentUsecase interface {
	ListByRecordID(
		ctx context.Context,
		userID user.ID,
		recordID record.ID,
	) ([]record.Attachment, error)
	Create(
		ctx context.Context,
		userID user.ID,
		recordID record.ID,
		m media.Media,
	) (*record.Attachment, error)
	Delete(
		ctx context.Context,
		userID user.ID,
		recordID record.ID,
		attachmentID record.AttachmentID,
	) error
}

type attachmentUsecase struct {
	attachmentRepo record.AttachmentRepository
	recordRepo     record.Repository
	mediaService   media.Service
}

func NewAttachmentUsecase(
	attachmentRepo record.AttachmentRepository,
	recordRepo record.Repository,
	mediaService media.Service,
) AttachmentUsecase {
	return &attachmentUsecase{
		attachmentRepo: attachmentRepo,
		recordRepo:     recordRepo,
		mediaService:   mediaService,
	}
}

func (au *attachmentUsecase) checkRecordOwner(
	ctx context.Context,
	userID user.ID,
	recordID record.ID,
) error {
	r, err := au.recordRepo.GetByID(ctx, recordID)
	if err != nil {
		return fmt.Errorf("get record by id: %w", err)
	}

	if r.UserID != userID {
		return exception.ErrNotFound
	}

	return nil
}

func (au *attachmentUsecase) ListByRecordID(
	ctx context.Context,
	userID user.ID,
	recordID record.ID,
) ([]record.Attachment, error) {
	if err := au.checkRecordOwner(ctx, userID, recordID); err != nil {
		return nil, err
	}

	attachments, err := au.attachmentRepo.ListByRecordID(ctx, recordID)
	if err != nil {
		return nil, fmt.Errorf("list attachments by record id: %w", err)
	}

	return attachments, nil
}

func (au *attachmentUsecase) Create(
	ctx context.Context,
	userID user.ID,
	recordID record.ID,
	m media.Media,
) (*record.Attachment, error) {
	if err := au.checkRecordOwner(ctx, userID, recordID); err != nil {
		return nil, err
	}

	count, err := au.attachmentRepo.CountByRecordID(ctx, recordID)
	if err != nil {
		return nil, fmt.Errorf("count attachments by record id: %w", err)
	}

	if err := record.CheckAttachmentLimit(count); err != nil {
		return nil, err
	}

	url, err := au.mediaService.Upload(ctx, userID, m)
	if err != nil {
		return nil, fmt.Errorf("upload media: %w", err)
	}

	a := record.NewAttachment(userID, recordID, record.AttachmentImageURL(url))
	if err := au.attachmentRepo.Create(ctx, &a); err != nil {
		if err := au.mediaService.Delete(ctx, userID, url); err != nil {
			slog.WarnContext(ctx, "media_delete_failed", "userId", fmt.Sprint(userID), "message", err.Error())
		}
		return nil, fmt.Errorf("create attachment: %w", err)
	}

	return &a, nil
}

func (au *attachmentUsecase) Delete(
	ctx context.Context,
	userID user.ID,
	recordID record.ID,
	attachmentID record.AttachmentID,
) error {
	a, err := au.attachmentRepo.GetByID(ctx, attachmentID)
	if err != nil {
		return fmt.Errorf("get attachment by id: %w", err)
	}

	if a.UserID != userID || a.RecordID != recordID {
		return exception.ErrNotFound
	}

	if err := au.attachmentRepo.Delete(ctx, attachmentID); err != nil {
		return fmt.Errorf("delete attachment: %w", err)
	}

	if err := au.mediaService.Delete(ctx, userID, media.URL(a.ImageURL)); err != nil {
		slog.WarnContext(ctx, "media_delete_failed", "userId", fmt.Sprint(userID), "message", err.Error())
	}

	return nil
}
