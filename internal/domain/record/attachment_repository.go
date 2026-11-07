package record

import "context"

type AttachmentRepository interface {
	GetByID(ctx context.Context, attachmentID AttachmentID) (*Attachment, error)
	ListByRecordID(ctx context.Context, recordID ID) ([]Attachment, error)
	CountByRecordID(ctx context.Context, recordID ID) (uint, error)
	Create(ctx context.Context, a *Attachment) error
	Delete(ctx context.Context, attachmentID AttachmentID) error
}
