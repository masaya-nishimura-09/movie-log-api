package record

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/record"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
	"gorm.io/gorm"
)

type attachmentRepository struct {
	db *gorm.DB
}

func NewAttachmentRepository(db *gorm.DB) record.AttachmentRepository {
	return &attachmentRepository{db}
}

type attachmentDTO struct {
	ID        uint `gorm:"primaryKey"`
	UserID    uint
	RecordID  uint
	ImageURL  string
	CreatedAt time.Time
}

func (attachmentDTO) TableName() string {
	return "record_attachments"
}

func toAttachmentDTO(a *record.Attachment) attachmentDTO {
	return attachmentDTO{
		ID:        uint(a.ID),
		UserID:    uint(a.UserID),
		RecordID:  uint(a.RecordID),
		ImageURL:  string(a.ImageURL),
		CreatedAt: a.CreatedAt,
	}
}

func fromAttachmentDTO(dto *attachmentDTO) record.Attachment {
	return record.Attachment{
		ID:        record.AttachmentID(dto.ID),
		UserID:    user.ID(dto.UserID),
		RecordID:  record.ID(dto.RecordID),
		ImageURL:  record.AttachmentImageURL(dto.ImageURL),
		CreatedAt: dto.CreatedAt,
	}
}

func (r *attachmentRepository) GetByID(
	ctx context.Context,
	attachmentID record.AttachmentID,
) (*record.Attachment, error) {
	var dto attachmentDTO
	result := r.db.WithContext(ctx).First(&dto, uint(attachmentID))
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, exception.ErrNotFound
	}
	if result.Error != nil {
		return nil, fmt.Errorf("get attachment by id: %w", result.Error)
	}
	a := fromAttachmentDTO(&dto)
	return &a, nil
}

func (r *attachmentRepository) ListByRecordID(
	ctx context.Context,
	recordID record.ID,
) ([]record.Attachment, error) {
	var dtos []attachmentDTO
	result := r.db.WithContext(ctx).
		Where("record_id = ?", uint(recordID)).
		Order("created_at ASC, id ASC").
		Find(&dtos)
	if result.Error != nil {
		return nil, fmt.Errorf("list attachments by record id: %w", result.Error)
	}

	attachments := make([]record.Attachment, 0, len(dtos))
	for i := range dtos {
		attachments = append(attachments, fromAttachmentDTO(&dtos[i]))
	}
	return attachments, nil
}

func (r *attachmentRepository) CountByRecordID(
	ctx context.Context,
	recordID record.ID,
) (uint, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&attachmentDTO{}).
		Where("record_id = ?", uint(recordID)).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count attachments by record id: %w", err)
	}
	return uint(count), nil
}

func (r *attachmentRepository) Create(
	ctx context.Context,
	a *record.Attachment,
) error {
	a.CreatedAt = time.Now()

	dto := toAttachmentDTO(a)
	result := r.db.WithContext(ctx).Create(&dto)
	if errors.Is(result.Error, gorm.ErrForeignKeyViolated) {
		return exception.ErrNotFound
	}
	if result.Error != nil {
		return fmt.Errorf("create attachment: %w", result.Error)
	}
	a.ID = record.AttachmentID(dto.ID)
	return nil
}

func (r *attachmentRepository) Delete(
	ctx context.Context,
	attachmentID record.AttachmentID,
) error {
	result := r.db.WithContext(ctx).Delete(&attachmentDTO{}, uint(attachmentID))
	if result.Error != nil {
		return fmt.Errorf("delete attachment: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return exception.ErrNotFound
	}
	return nil
}
