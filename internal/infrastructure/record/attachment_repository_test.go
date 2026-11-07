package record

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/record"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
	"github.com/masaya-nishimura-09/movie-log-api/internal/testutil"
	"gorm.io/gorm"
)

func newTestAttachment(userID user.ID, recordID record.ID, name string) record.Attachment {
	return record.NewAttachment(
		userID,
		recordID,
		record.AttachmentImageURL(fmt.Sprintf("https://example.com/%d/%s.webp", uint(userID), name)),
	)
}

func createTestAttachment(t *testing.T, tx *gorm.DB, a *record.Attachment) {
	t.Helper()
	if err := NewAttachmentRepository(tx).Create(context.Background(), a); err != nil {
		t.Fatalf("Create(ctx, %v) error = %v", *a, err)
	}
}

func assertAttachmentEqual(t *testing.T, call string, got, want *record.Attachment) {
	t.Helper()

	if got.ID != want.ID ||
		got.UserID != want.UserID ||
		got.RecordID != want.RecordID ||
		got.ImageURL != want.ImageURL {
		t.Errorf("%s = %v, want %v", call, got, want)
	}
}

func TestAttachmentGetByID(t *testing.T) {
	t.Run(
		"returns ErrNotFound when the attachment does not exist",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			r := NewAttachmentRepository(tx)

			ctx := context.Background()
			id := record.AttachmentID(999999)

			got, err := r.GetByID(ctx, id)
			if !errors.Is(err, exception.ErrNotFound) {
				t.Fatalf(
					"GetByID(ctx, %d) (*record.Attachment, error) = %v, %v, want %v",
					id, got, err, exception.ErrNotFound,
				)
			}
		},
	)
}

func TestAttachmentListByRecordID(t *testing.T) {
	t.Run(
		"returns the attachments of the record, oldest first",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			r := NewAttachmentRepository(tx)

			ctx := context.Background()
			userID := newTestUser(t, tx, "test@example.com").ID
			recordID := createTestRecord(t, tx, userID)
			otherRecordID := createTestRecord(t, tx, userID)

			first := newTestAttachment(userID, recordID, "first")
			createTestAttachment(t, tx, &first)
			other := newTestAttachment(userID, otherRecordID, "other")
			createTestAttachment(t, tx, &other)
			second := newTestAttachment(userID, recordID, "second")
			createTestAttachment(t, tx, &second)

			got, err := r.ListByRecordID(ctx, recordID)
			if err != nil {
				t.Fatalf(
					"ListByRecordID(ctx, %d) ([]record.Attachment, error) = %v, %v",
					recordID, got, err,
				)
			}
			want := []record.Attachment{first, second}
			if len(got) != len(want) {
				t.Fatalf(
					"ListByRecordID(ctx, %d) = %v, want %d attachments",
					recordID, got, len(want),
				)
			}
			for i := range want {
				assertAttachmentEqual(
					t,
					fmt.Sprintf("ListByRecordID(ctx, %d)[%d]", recordID, i),
					&got[i], &want[i],
				)
			}
		},
	)

	t.Run(
		"returns an empty list when the record has no attachments",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			r := NewAttachmentRepository(tx)

			ctx := context.Background()
			userID := newTestUser(t, tx, "test@example.com").ID
			recordID := createTestRecord(t, tx, userID)

			got, err := r.ListByRecordID(ctx, recordID)
			if err != nil || len(got) != 0 {
				t.Fatalf(
					"ListByRecordID(ctx, %d) ([]record.Attachment, error) = %v, %v, want empty",
					recordID, got, err,
				)
			}
		},
	)
}

func TestAttachmentCountByRecordID(t *testing.T) {
	t.Run(
		"counts only the attachments of the record",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			r := NewAttachmentRepository(tx)

			ctx := context.Background()
			userID := newTestUser(t, tx, "test@example.com").ID
			recordID := createTestRecord(t, tx, userID)
			otherRecordID := createTestRecord(t, tx, userID)

			for _, name := range []string{"a", "b"} {
				a := newTestAttachment(userID, recordID, name)
				createTestAttachment(t, tx, &a)
			}
			other := newTestAttachment(userID, otherRecordID, "other")
			createTestAttachment(t, tx, &other)

			got, err := r.CountByRecordID(ctx, recordID)
			if err != nil || got != 2 {
				t.Fatalf(
					"CountByRecordID(ctx, %d) (uint, error) = %v, %v, want 2, nil",
					recordID, got, err,
				)
			}
		},
	)
}

func TestAttachmentCreate(t *testing.T) {
	t.Run(
		"persists the attachment",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			r := NewAttachmentRepository(tx)

			ctx := context.Background()
			userID := newTestUser(t, tx, "test@example.com").ID
			recordID := createTestRecord(t, tx, userID)
			want := newTestAttachment(userID, recordID, "pamphlet")

			if err := r.Create(ctx, &want); err != nil {
				t.Fatalf("Create(ctx, %v) error = %v", want, err)
			}
			if want.ID == 0 {
				t.Fatalf("Create(ctx, %v) sets ID = 0, want non-zero", want)
			}
			if want.CreatedAt.IsZero() {
				t.Fatalf("Create(ctx, %v) sets CreatedAt = %v, want non-zero", want, want.CreatedAt)
			}

			got, err := r.GetByID(ctx, want.ID)
			if err != nil {
				t.Fatalf(
					"GetByID(ctx, %d) (*record.Attachment, error) = %v, %v",
					want.ID, got, err,
				)
			}
			assertAttachmentEqual(
				t,
				fmt.Sprintf("GetByID(ctx, %d) (*record.Attachment, error)", want.ID),
				got, &want,
			)
		},
	)

	t.Run(
		"returns ErrNotFound when the record does not exist",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			r := NewAttachmentRepository(tx)

			ctx := context.Background()
			userID := newTestUser(t, tx, "test@example.com").ID
			a := newTestAttachment(userID, record.ID(999999), "pamphlet")

			if err := r.Create(ctx, &a); !errors.Is(err, exception.ErrNotFound) {
				t.Fatalf(
					"Create(ctx, %v) error = %v, want %v",
					a, err, exception.ErrNotFound,
				)
			}
		},
	)

	t.Run(
		"returns a wrapped error when the context is canceled",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			r := NewAttachmentRepository(tx)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			a := newTestAttachment(user.ID(1), record.ID(1), "pamphlet")
			err := r.Create(ctx, &a)
			if !errors.Is(err, context.Canceled) {
				t.Fatalf(
					"Create(ctx, %v) error = %v, want %v",
					a, err, context.Canceled,
				)
			}
		},
	)
}

func TestAttachmentDelete(t *testing.T) {
	t.Run(
		"deletes the attachment",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			r := NewAttachmentRepository(tx)

			ctx := context.Background()
			userID := newTestUser(t, tx, "test@example.com").ID
			recordID := createTestRecord(t, tx, userID)
			a := newTestAttachment(userID, recordID, "pamphlet")
			createTestAttachment(t, tx, &a)

			if err := r.Delete(ctx, a.ID); err != nil {
				t.Fatalf("Delete(ctx, %d) error = %v", a.ID, err)
			}

			got, err := r.GetByID(ctx, a.ID)
			if !errors.Is(err, exception.ErrNotFound) {
				t.Fatalf(
					"GetByID(ctx, %d) (*record.Attachment, error) = %v, %v, want %v",
					a.ID, got, err, exception.ErrNotFound,
				)
			}
		},
	)

	t.Run(
		"is deleted together with its record",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			r := NewAttachmentRepository(tx)

			ctx := context.Background()
			userID := newTestUser(t, tx, "test@example.com").ID
			recordID := createTestRecord(t, tx, userID)
			a := newTestAttachment(userID, recordID, "pamphlet")
			createTestAttachment(t, tx, &a)

			if err := NewRepository(tx).Delete(ctx, recordID); err != nil {
				t.Fatalf("Delete(ctx, %d) error = %v", recordID, err)
			}

			got, err := r.GetByID(ctx, a.ID)
			if !errors.Is(err, exception.ErrNotFound) {
				t.Fatalf(
					"GetByID(ctx, %d) (*record.Attachment, error) = %v, %v, want %v",
					a.ID, got, err, exception.ErrNotFound,
				)
			}
		},
	)

	t.Run(
		"returns ErrNotFound when the attachment does not exist",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			r := NewAttachmentRepository(tx)

			ctx := context.Background()
			id := record.AttachmentID(999999)

			if err := r.Delete(ctx, id); !errors.Is(err, exception.ErrNotFound) {
				t.Fatalf(
					"Delete(ctx, %d) error = %v, want %v",
					id, err, exception.ErrNotFound,
				)
			}
		},
	)
}
