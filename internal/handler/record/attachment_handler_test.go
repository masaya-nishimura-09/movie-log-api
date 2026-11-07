package record

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	mediadomain "github.com/masaya-nishimura-09/movie-log-api/internal/domain/media"
	recorddomain "github.com/masaya-nishimura-09/movie-log-api/internal/domain/record"
	userdomain "github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
)

type fakeAttachmentUsecase struct {
	attachments []recorddomain.Attachment
	attachment  *recorddomain.Attachment
	err         error

	userID       userdomain.ID
	recordID     recorddomain.ID
	media        mediadomain.Media
	attachmentID recorddomain.AttachmentID
}

func (u *fakeAttachmentUsecase) ListByRecordID(
	ctx context.Context,
	userID userdomain.ID,
	recordID recorddomain.ID,
) ([]recorddomain.Attachment, error) {
	u.userID = userID
	u.recordID = recordID
	return u.attachments, u.err
}

func (u *fakeAttachmentUsecase) Create(
	ctx context.Context,
	userID userdomain.ID,
	recordID recorddomain.ID,
	m mediadomain.Media,
) (*recorddomain.Attachment, error) {
	u.userID = userID
	u.recordID = recordID
	u.media = m
	return u.attachment, u.err
}

func (u *fakeAttachmentUsecase) Delete(
	ctx context.Context,
	userID userdomain.ID,
	recordID recorddomain.ID,
	attachmentID recorddomain.AttachmentID,
) error {
	u.userID = userID
	u.recordID = recordID
	u.attachmentID = attachmentID
	return u.err
}

func newTestAttachment() recorddomain.Attachment {
	return recorddomain.Attachment{
		ID:        recorddomain.AttachmentID(5),
		UserID:    userdomain.ID(1),
		RecordID:  recorddomain.ID(10),
		ImageURL:  recorddomain.AttachmentImageURL("https://example.com/1/image.jpg"),
		CreatedAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
	}
}

const wantAttachmentBody = `{"attachment_id":"5","created_at":"2026-01-02T00:00:00Z",` +
	`"image_url":"https://example.com/1/image.jpg"}`

func newTestAttachmentRequest(t *testing.T, field string, data []byte) *http.Request {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(field, "image.jpg")
	if err != nil {
		t.Fatalf("CreateFormFile(%q, %q) error = %v", field, "image.jpg", err)
	}
	if _, err := part.Write(data); err != nil {
		t.Fatalf("Write(len=%d) error = %v", len(data), err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func TestAttachmentListByRecordID(t *testing.T) {
	t.Run(
		"passes the IDs to the usecase and returns 200 with the attachments",
		func(t *testing.T) {
			userID := userdomain.ID(1)
			usecase := &fakeAttachmentUsecase{
				attachments: []recorddomain.Attachment{newTestAttachment()},
			}
			attachmentHandler := NewAttachmentHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userID)
			c.Params = gin.Params{{Key: "id", Value: "10"}}
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

			attachmentHandler.ListByRecordID(c)
			if rec.Code != http.StatusOK {
				t.Errorf("ListByRecordID(c) code = %v, want %v", rec.Code, http.StatusOK)
			}
			want := `{"attachments":[` + wantAttachmentBody + `]}`
			if rec.Body.String() != want {
				t.Errorf("ListByRecordID(c) body = %v, want %v", rec.Body.String(), want)
			}
			if usecase.userID != userID || usecase.recordID != recorddomain.ID(10) {
				t.Errorf(
					"ListByRecordID(c) usecase user id, record id = %v, %v, want %v, %v",
					usecase.userID, usecase.recordID, userID, recorddomain.ID(10),
				)
			}
		},
	)

	t.Run(
		"returns an empty array when the record has no attachments",
		func(t *testing.T) {
			usecase := &fakeAttachmentUsecase{attachments: []recorddomain.Attachment{}}
			attachmentHandler := NewAttachmentHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Params = gin.Params{{Key: "id", Value: "10"}}
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

			attachmentHandler.ListByRecordID(c)
			want := `{"attachments":[]}`
			if rec.Body.String() != want {
				t.Errorf("ListByRecordID(c) body = %v, want %v", rec.Body.String(), want)
			}
		},
	)

	t.Run(
		"returns 404 when the record is not found",
		func(t *testing.T) {
			usecase := &fakeAttachmentUsecase{err: exception.ErrNotFound}
			attachmentHandler := NewAttachmentHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Params = gin.Params{{Key: "id", Value: "10"}}
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

			attachmentHandler.ListByRecordID(c)
			if rec.Code != http.StatusNotFound {
				t.Errorf("ListByRecordID(c) code = %v, want %v", rec.Code, http.StatusNotFound)
			}
			if !strings.Contains(rec.Body.String(), `"code":"RECORD_NOT_FOUND"`) {
				t.Errorf("ListByRecordID(c) body = %v, want RECORD_NOT_FOUND", rec.Body.String())
			}
		},
	)
}

func TestAttachmentCreate(t *testing.T) {
	jpeg := []byte("\xFF\xD8\xFF")

	t.Run(
		"passes the converted media to the usecase and returns 201 when the request is valid",
		func(t *testing.T) {
			userID := userdomain.ID(1)
			a := newTestAttachment()
			usecase := &fakeAttachmentUsecase{attachment: &a}
			attachmentHandler := NewAttachmentHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userID)
			c.Params = gin.Params{{Key: "id", Value: "10"}}
			c.Request = newTestAttachmentRequest(t, "file", jpeg)

			attachmentHandler.Create(c)
			if rec.Code != http.StatusCreated {
				t.Errorf("Create(c) code = %v, want %v", rec.Code, http.StatusCreated)
			}
			if rec.Body.String() != wantAttachmentBody {
				t.Errorf("Create(c) body = %v, want %v", rec.Body.String(), wantAttachmentBody)
			}
			if usecase.userID != userID || usecase.recordID != recorddomain.ID(10) {
				t.Errorf(
					"Create(c) usecase user id, record id = %v, %v, want %v, %v",
					usecase.userID, usecase.recordID, userID, recorddomain.ID(10),
				)
			}
			if !bytes.Equal(usecase.media.Data, jpeg) ||
				usecase.media.ContentType != mediadomain.ContentTypeJPEG {
				t.Errorf(
					"Create(c) usecase media = %v, want %v with %v",
					usecase.media, jpeg, mediadomain.ContentTypeJPEG,
				)
			}
		},
	)

	t.Run(
		"returns 400 when the request has no file",
		func(t *testing.T) {
			attachmentHandler := NewAttachmentHandler(&fakeAttachmentUsecase{})

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Params = gin.Params{{Key: "id", Value: "10"}}
			c.Request = newTestAttachmentRequest(t, "image", jpeg)

			attachmentHandler.Create(c)
			if rec.Code != http.StatusBadRequest {
				t.Errorf("Create(c) code = %v, want %v", rec.Code, http.StatusBadRequest)
			}
		},
	)

	t.Run(
		"returns 400 when the file is not a supported image",
		func(t *testing.T) {
			attachmentHandler := NewAttachmentHandler(&fakeAttachmentUsecase{})

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Params = gin.Params{{Key: "id", Value: "10"}}
			c.Request = newTestAttachmentRequest(t, "file", []byte("not an image"))

			attachmentHandler.Create(c)
			if rec.Code != http.StatusBadRequest {
				t.Errorf("Create(c) code = %v, want %v", rec.Code, http.StatusBadRequest)
			}
			want := `"message":"invalid: invalid content type"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf("Create(c) body = %v, want to contain %v", rec.Body.String(), want)
			}
		},
	)

	t.Run(
		"returns 400 when the request body is larger than 6 megabytes",
		func(t *testing.T) {
			attachmentHandler := NewAttachmentHandler(&fakeAttachmentUsecase{})

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Params = gin.Params{{Key: "id", Value: "10"}}
			data := append(
				[]byte("\xFF\xD8\xFF"),
				make([]byte, mediadomain.MaxBytes+1024*1024)...,
			)
			c.Request = newTestAttachmentRequest(t, "file", data)

			attachmentHandler.Create(c)
			if rec.Code != http.StatusBadRequest {
				t.Errorf("Create(c) code = %v, want %v", rec.Code, http.StatusBadRequest)
			}
			want := `"message":"invalid: media must be at most 5 megabytes"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf("Create(c) body = %v, want to contain %v", rec.Body.String(), want)
			}
		},
	)

	t.Run(
		"returns 400 when the record already has the maximum number of attachments",
		func(t *testing.T) {
			usecase := &fakeAttachmentUsecase{
				err: fmt.Errorf("%w: a record can have at most 10 attachments", exception.ErrInvalid),
			}
			attachmentHandler := NewAttachmentHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Params = gin.Params{{Key: "id", Value: "10"}}
			c.Request = newTestAttachmentRequest(t, "file", jpeg)

			attachmentHandler.Create(c)
			if rec.Code != http.StatusBadRequest {
				t.Errorf("Create(c) code = %v, want %v", rec.Code, http.StatusBadRequest)
			}
			if !strings.Contains(rec.Body.String(), `"code":"INVALID_INPUT"`) {
				t.Errorf("Create(c) body = %v, want INVALID_INPUT", rec.Body.String())
			}
		},
	)

	t.Run(
		"returns 404 when the record is not found",
		func(t *testing.T) {
			usecase := &fakeAttachmentUsecase{err: exception.ErrNotFound}
			attachmentHandler := NewAttachmentHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Params = gin.Params{{Key: "id", Value: "10"}}
			c.Request = newTestAttachmentRequest(t, "file", jpeg)

			attachmentHandler.Create(c)
			if rec.Code != http.StatusNotFound {
				t.Errorf("Create(c) code = %v, want %v", rec.Code, http.StatusNotFound)
			}
			if !strings.Contains(rec.Body.String(), `"code":"RECORD_NOT_FOUND"`) {
				t.Errorf("Create(c) body = %v, want RECORD_NOT_FOUND", rec.Body.String())
			}
		},
	)

	t.Run(
		"returns 500 when the usecase returns an unexpected error",
		func(t *testing.T) {
			usecase := &fakeAttachmentUsecase{err: errors.New("upload media")}
			attachmentHandler := NewAttachmentHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Params = gin.Params{{Key: "id", Value: "10"}}
			c.Request = newTestAttachmentRequest(t, "file", jpeg)

			attachmentHandler.Create(c)
			if rec.Code != http.StatusInternalServerError {
				t.Errorf("Create(c) code = %v, want %v", rec.Code, http.StatusInternalServerError)
			}
		},
	)
}

func TestAttachmentDelete(t *testing.T) {
	t.Run(
		"passes the IDs to the usecase and returns 204 when the request is valid",
		func(t *testing.T) {
			userID := userdomain.ID(1)
			usecase := &fakeAttachmentUsecase{}
			attachmentHandler := NewAttachmentHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userID)
			c.Params = gin.Params{
				{Key: "id", Value: "10"},
				{Key: "attachmentId", Value: "5"},
			}
			c.Request = httptest.NewRequest(http.MethodDelete, "/", nil)

			attachmentHandler.Delete(c)
			c.Writer.WriteHeaderNow()
			if rec.Code != http.StatusNoContent {
				t.Errorf("Delete(c) code = %v, want %v", rec.Code, http.StatusNoContent)
			}
			if usecase.userID != userID ||
				usecase.recordID != recorddomain.ID(10) ||
				usecase.attachmentID != recorddomain.AttachmentID(5) {
				t.Errorf(
					"Delete(c) usecase user id, record id, attachment id = %v, %v, %v, want %v, %v, %v",
					usecase.userID, usecase.recordID, usecase.attachmentID,
					userID, recorddomain.ID(10), recorddomain.AttachmentID(5),
				)
			}
		},
	)

	t.Run(
		"returns 400 when the attachment ID is not a number",
		func(t *testing.T) {
			attachmentHandler := NewAttachmentHandler(&fakeAttachmentUsecase{})

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Params = gin.Params{
				{Key: "id", Value: "10"},
				{Key: "attachmentId", Value: "invalid"},
			}
			c.Request = httptest.NewRequest(http.MethodDelete, "/", nil)

			attachmentHandler.Delete(c)
			if rec.Code != http.StatusBadRequest {
				t.Errorf("Delete(c) code = %v, want %v", rec.Code, http.StatusBadRequest)
			}
		},
	)

	t.Run(
		"returns 404 when the attachment is not found",
		func(t *testing.T) {
			usecase := &fakeAttachmentUsecase{err: exception.ErrNotFound}
			attachmentHandler := NewAttachmentHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Params = gin.Params{
				{Key: "id", Value: "10"},
				{Key: "attachmentId", Value: "5"},
			}
			c.Request = httptest.NewRequest(http.MethodDelete, "/", nil)

			attachmentHandler.Delete(c)
			if rec.Code != http.StatusNotFound {
				t.Errorf("Delete(c) code = %v, want %v", rec.Code, http.StatusNotFound)
			}
			if !strings.Contains(rec.Body.String(), `"code":"ATTACHMENT_NOT_FOUND"`) {
				t.Errorf("Delete(c) body = %v, want ATTACHMENT_NOT_FOUND", rec.Body.String())
			}
		},
	)
}
