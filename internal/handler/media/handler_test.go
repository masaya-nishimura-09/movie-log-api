package media

import (
	"bytes"
	"context"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	mediadomain "github.com/masaya-nishimura-09/movie-log-api/internal/domain/media"
	userdomain "github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
)

type fakeUsecase struct {
	url            mediadomain.URL
	err            error
	uploadedUserID userdomain.ID
	uploadedMedia  mediadomain.Media
}

func (u *fakeUsecase) Upload(
	ctx context.Context,
	userID userdomain.ID,
	m mediadomain.Media,
) (mediadomain.URL, error) {
	u.uploadedUserID = userID
	u.uploadedMedia = m

	return u.url, u.err
}

func (u *fakeUsecase) Delete(
	ctx context.Context,
	userID userdomain.ID,
	url mediadomain.URL,
) error {
	return u.err
}

func newTestMediaRequest(t *testing.T, field string, data []byte) *http.Request {
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

func TestUpload(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jpeg := []byte("\xFF\xD8\xFF")

	t.Run(
		"passes the converted media to the usecase and returns 201 when the request is valid",
		func(t *testing.T) {
			userID := userdomain.ID(1)
			url := mediadomain.URL("https://example.com/1/image.jpg")
			usecase := &fakeUsecase{url: url}
			mediaHandler := NewMediaHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userID)
			c.Request = newTestMediaRequest(t, "file", jpeg)

			mediaHandler.Upload(c)
			if rec.Code != http.StatusCreated {
				t.Errorf(
					"Upload(c) code = %v, want %v",
					rec.Code, http.StatusCreated,
				)
			}
			want := `{"url":"https://example.com/1/image.jpg"}`
			if rec.Body.String() != want {
				t.Errorf(
					"Upload(c) body = %v, want %v",
					rec.Body.String(), want,
				)
			}

			if usecase.uploadedUserID != userID {
				t.Errorf(
					"Upload(c) usecase user id = %v, want %v",
					usecase.uploadedUserID, userID,
				)
			}
			if !bytes.Equal(usecase.uploadedMedia.Data, jpeg) ||
				usecase.uploadedMedia.ContentType != mediadomain.ContentTypeJPEG {
				t.Errorf(
					"Upload(c) usecase media = %v, want %v with %v",
					usecase.uploadedMedia, jpeg, mediadomain.ContentTypeJPEG,
				)
			}
		},
	)

	t.Run(
		"returns 400 when the request has no file",
		func(t *testing.T) {
			usecase := &fakeUsecase{}
			mediaHandler := NewMediaHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Request = newTestMediaRequest(t, "image", jpeg)

			mediaHandler.Upload(c)
			if rec.Code != http.StatusBadRequest {
				t.Errorf(
					"Upload(c) code = %v, want %v",
					rec.Code, http.StatusBadRequest,
				)
			}
			want := `"code":"INVALID_INPUT","message":"malformed request body"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"Upload(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)

	t.Run(
		"returns 400 when the file is not a supported image",
		func(t *testing.T) {
			usecase := &fakeUsecase{}
			mediaHandler := NewMediaHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Request = newTestMediaRequest(t, "file", []byte("not an image"))

			mediaHandler.Upload(c)
			if rec.Code != http.StatusBadRequest {
				t.Errorf(
					"Upload(c) code = %v, want %v",
					rec.Code, http.StatusBadRequest,
				)
			}
			want := `"message":"invalid: invalid content type"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"Upload(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)

	t.Run(
		"returns 400 when the file is larger than 5 megabytes",
		func(t *testing.T) {
			usecase := &fakeUsecase{}
			mediaHandler := NewMediaHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			data := append(
				[]byte("\xFF\xD8\xFF"),
				make([]byte, mediadomain.MaxBytes)...,
			)
			c.Request = newTestMediaRequest(t, "file", data)

			mediaHandler.Upload(c)
			if rec.Code != http.StatusBadRequest {
				t.Errorf(
					"Upload(c) code = %v, want %v",
					rec.Code, http.StatusBadRequest,
				)
			}
			want := `"message":"invalid: media must be at most 5 megabytes"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"Upload(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)

	t.Run(
		"returns 500 when the usecase returns an unexpected error",
		func(t *testing.T) {
			usecase := &fakeUsecase{err: errors.New("upload media")}
			mediaHandler := NewMediaHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userdomain.ID(1))
			c.Request = newTestMediaRequest(t, "file", jpeg)

			mediaHandler.Upload(c)
			if rec.Code != http.StatusInternalServerError {
				t.Errorf(
					"Upload(c) code = %v, want %v",
					rec.Code, http.StatusInternalServerError,
				)
			}
		},
	)

	t.Run(
		"returns 500 when the user ID is missing from the context",
		func(t *testing.T) {
			usecase := &fakeUsecase{}
			mediaHandler := NewMediaHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = newTestMediaRequest(t, "file", jpeg)

			mediaHandler.Upload(c)
			if rec.Code != http.StatusInternalServerError {
				t.Errorf(
					"Upload(c) code = %v, want %v",
					rec.Code, http.StatusInternalServerError,
				)
			}
		},
	)
}
