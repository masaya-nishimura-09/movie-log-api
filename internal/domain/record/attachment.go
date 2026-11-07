package record

import (
	"fmt"
	"time"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
)

const MaxAttachmentsPerRecord = 10

type Attachment struct {
	ID        AttachmentID
	UserID    user.ID
	RecordID  ID
	ImageURL  AttachmentImageURL
	CreatedAt time.Time
}

type AttachmentID uint

type AttachmentImageURL string

func NewAttachment(
	userID user.ID,
	recordID ID,
	imageURL AttachmentImageURL,
) Attachment {
	return Attachment{
		UserID:   userID,
		RecordID: recordID,
		ImageURL: imageURL,
	}
}

func CheckAttachmentLimit(current uint) error {
	if current >= MaxAttachmentsPerRecord {
		return fmt.Errorf(
			"%w: a record can have at most %d attachments",
			exception.ErrInvalid, MaxAttachmentsPerRecord,
		)
	}

	return nil
}
