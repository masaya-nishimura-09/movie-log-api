package contact

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	contactdomain "github.com/masaya-nishimura-09/movie-log-api/internal/domain/contact"
	maildomain "github.com/masaya-nishimura-09/movie-log-api/internal/domain/mail"
)

type fakeUsecase struct {
	err          error
	sentEmail    maildomain.Address
	sentLanguage maildomain.Language
	sentMessage  contactdomain.Message
}

func (u *fakeUsecase) Send(
	ctx context.Context,
	email maildomain.Address,
	language maildomain.Language,
	message contactdomain.Message,
) error {
	u.sentEmail = email
	u.sentLanguage = language
	u.sentMessage = message
	return u.err
}

func TestCreate(t *testing.T) {
	t.Run(
		"passes the converted values to the usecase and returns 204 when the request is valid",
		func(t *testing.T) {
			usecase := &fakeUsecase{}
			contactHandler := NewContactHandler(usecase)

			body := `{
				"email":"User@Example.com",
				"message":"  退会できません\n"
			}`
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(
				http.MethodPost, "/?language=ja", strings.NewReader(body),
			)

			contactHandler.Create(c)
			c.Writer.WriteHeaderNow()
			if rec.Code != http.StatusNoContent {
				t.Errorf(
					"Create(c) code = %v, want %v",
					rec.Code, http.StatusNoContent,
				)
			}

			email := maildomain.Address("user@example.com")
			message := contactdomain.Message("退会できません")
			if usecase.sentEmail != email ||
				usecase.sentLanguage != maildomain.LanguageJA ||
				usecase.sentMessage != message {
				t.Errorf(
					"Create(c) usecase args = %v, %v, %v, want %v, %v, %v",
					usecase.sentEmail, usecase.sentLanguage, usecase.sentMessage,
					email, maildomain.LanguageJA, message,
				)
			}
		},
	)

	t.Run(
		"returns 400 when the request body is not valid JSON",
		func(t *testing.T) {
			usecase := &fakeUsecase{}
			contactHandler := NewContactHandler(usecase)

			body := `{"email":"user@example.com",}`
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(
				http.MethodPost, "/?language=ja", strings.NewReader(body),
			)

			contactHandler.Create(c)
			if rec.Code != http.StatusBadRequest {
				t.Errorf(
					"Create(c) code = %v, want %v",
					rec.Code, http.StatusBadRequest,
				)
			}
			want := `"code":"INVALID_INPUT","message":"malformed request body"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"Create(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)

	t.Run(
		"returns 400 without calling the usecase when the email is not a valid address",
		func(t *testing.T) {
			usecase := &fakeUsecase{}
			contactHandler := NewContactHandler(usecase)

			body := `{"email":"invalid","message":"退会できません"}`
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(
				http.MethodPost, "/?language=ja", strings.NewReader(body),
			)

			contactHandler.Create(c)
			if rec.Code != http.StatusBadRequest {
				t.Errorf(
					"Create(c) code = %v, want %v",
					rec.Code, http.StatusBadRequest,
				)
			}
			want := `"code":"INVALID_INPUT"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"Create(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
			if usecase.sentEmail != "" {
				t.Errorf("Create(c) usecase args = %v, want no call", usecase.sentEmail)
			}
		},
	)

	t.Run(
		"returns 400 without calling the usecase when the message is empty",
		func(t *testing.T) {
			usecase := &fakeUsecase{}
			contactHandler := NewContactHandler(usecase)

			body := `{"email":"user@example.com","message":"  "}`
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(
				http.MethodPost, "/?language=ja", strings.NewReader(body),
			)

			contactHandler.Create(c)
			if rec.Code != http.StatusBadRequest {
				t.Errorf(
					"Create(c) code = %v, want %v",
					rec.Code, http.StatusBadRequest,
				)
			}
			if usecase.sentEmail != "" {
				t.Errorf("Create(c) usecase args = %v, want no call", usecase.sentEmail)
			}
		},
	)

	t.Run(
		"returns 400 when the language is not supported",
		func(t *testing.T) {
			usecase := &fakeUsecase{}
			contactHandler := NewContactHandler(usecase)

			body := `{"email":"user@example.com","message":"退会できません"}`
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(
				http.MethodPost, "/?language=fr", strings.NewReader(body),
			)

			contactHandler.Create(c)
			if rec.Code != http.StatusBadRequest {
				t.Errorf(
					"Create(c) code = %v, want %v",
					rec.Code, http.StatusBadRequest,
				)
			}
		},
	)

	t.Run(
		"returns 500 when the usecase returns an unexpected error",
		func(t *testing.T) {
			usecase := &fakeUsecase{err: errors.New("unexpected")}
			contactHandler := NewContactHandler(usecase)

			body := `{"email":"user@example.com","message":"退会できません"}`
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(
				http.MethodPost, "/?language=ja", strings.NewReader(body),
			)

			contactHandler.Create(c)
			if rec.Code != http.StatusInternalServerError {
				t.Errorf(
					"Create(c) code = %v, want %v",
					rec.Code, http.StatusInternalServerError,
				)
			}
			want := `"code":"INTERNAL_SERVER_ERROR"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"Create(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)
}
