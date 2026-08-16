package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/auth"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
)

type fakeUsecase struct {
	accessToken               *auth.AccessToken
	refreshToken              *auth.RefreshToken
	refreshTokenValue         auth.RefreshTokenValue
	logoutedRefreshTokenValue auth.RefreshTokenValue
	loginedEmail              user.Email
	loginedPassword           user.Password
	err                       error
}

func (u *fakeUsecase) ValidateAccessToken(
	ctx context.Context,
	accessToken *auth.AccessToken,
) (*auth.Principal, error) {
	return nil, nil
}

func (u *fakeUsecase) Login(
	ctx context.Context,
	email user.Email,
	password user.Password,
) (*auth.AccessToken, *auth.RefreshToken, error) {
	u.loginedEmail = email
	u.loginedPassword = password
	return u.accessToken, u.refreshToken, u.err
}

func (u *fakeUsecase) Logout(
	ctx context.Context,
	refreshTokenValue auth.RefreshTokenValue,
) error {
	u.logoutedRefreshTokenValue = refreshTokenValue
	return u.err
}

func (u *fakeUsecase) Refresh(
	ctx context.Context,
	refreshTokenValue auth.RefreshTokenValue,
) (*auth.AccessToken, *auth.RefreshToken, error) {
	u.refreshTokenValue = refreshTokenValue
	return u.accessToken, u.refreshToken, u.err
}

func TestLogin(t *testing.T) {
	t.Run(
		"passes the converted values to the usecase and returns 200 when the request is valid",
		func(t *testing.T) {
			email := user.Email("test@example.com")
			password := user.Password("testpassword")
			atv := auth.AccessTokenValue("validaccesstoken")
			rtv := auth.RefreshTokenValue("validrefreshtoken")
			at := auth.AccessToken{
				Value: atv,
			}
			rt := auth.RefreshToken{
				Value: rtv,
			}

			usecase := &fakeUsecase{accessToken: &at, refreshToken: &rt}
			authHandler := NewAuthHandler(usecase)

			body := `{
				"email":"test@example.com",
				"password":"testpassword"
			}`
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(
				http.MethodPost, "/", strings.NewReader(body),
			)

			authHandler.Login(c)
			if rec.Code != http.StatusOK {
				t.Errorf(
					"Login(c) code = %v, want %v",
					rec.Code, http.StatusOK,
				)
			}
			want := `{"access_token":"validaccesstoken","refresh_token":"validrefreshtoken"}`
			if rec.Body.String() != want {
				t.Errorf("Login(c) body = %v, want %v", rec.Body.String(), want)
			}

			if usecase.loginedEmail != email ||
				usecase.loginedPassword != password {
				t.Errorf(
					"Login(c) usecase args = %v, %v, want %v, %v",
					usecase.loginedEmail,
					usecase.loginedPassword,
					email, password,
				)
			}
		},
	)

	t.Run(
		"returns 400 when the request body is not valid JSON",
		func(t *testing.T) {
			usecase := &fakeUsecase{}
			authHandler := NewAuthHandler(usecase)

			body := `{
				"email":"test@example.com",
				"password":"testpassword",
			}`
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(
				http.MethodPost, "/", strings.NewReader(body),
			)

			authHandler.Login(c)
			if rec.Code != http.StatusBadRequest {
				t.Errorf(
					"Login(c) code = %v, want %v",
					rec.Code, http.StatusBadRequest,
				)
			}
			want := `"code":"INVALID_INPUT"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"Login(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)

	t.Run(
		"returns 400 when the email is not a valid address",
		func(t *testing.T) {
			usecase := &fakeUsecase{}
			authHandler := NewAuthHandler(usecase)

			body := `{
				"email":"invalid",
				"password":"testpassword"
			}`
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(
				http.MethodPost, "/", strings.NewReader(body),
			)

			authHandler.Login(c)
			if rec.Code != http.StatusBadRequest {
				t.Errorf(
					"Login(c) code = %v, want %v",
					rec.Code, http.StatusBadRequest,
				)
			}
			want := `"code":"INVALID_INPUT"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"Login(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)

	t.Run(
		"returns 401 when the credentials do not match",
		func(t *testing.T) {
			usecase := &fakeUsecase{err: exception.ErrInvalid}
			authHandler := NewAuthHandler(usecase)

			body := `{
				"email":"test@example.com",
				"password":"testpassword"
			}`
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(
				http.MethodPost, "/", strings.NewReader(body),
			)

			authHandler.Login(c)
			if rec.Code != http.StatusUnauthorized {
				t.Errorf(
					"Login(c) code = %v, want %v",
					rec.Code, http.StatusUnauthorized,
				)
			}
			want := `"code":"INVALID_CREDENTIALS"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"Login(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)

	t.Run(
		"returns 500 when the usecase returns an unexpected error",
		func(t *testing.T) {
			usecase := &fakeUsecase{err: errors.New("unexpected")}
			authHandler := NewAuthHandler(usecase)

			body := `{
				"email":"test@example.com",
				"password":"testpassword"
			}`
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(
				http.MethodPost, "/", strings.NewReader(body),
			)

			authHandler.Login(c)
			if rec.Code != http.StatusInternalServerError {
				t.Errorf(
					"Login(c) code = %v, want %v",
					rec.Code, http.StatusInternalServerError,
				)
			}
			want := `"code":"INTERNAL_SERVER_ERROR"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"Login(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)
}

func TestLogout(t *testing.T) {
	t.Run(
		"passes the refresh token to the usecase and returns 204 when the request is valid",
		func(t *testing.T) {
			logoutedRefreshTokenValue := auth.RefreshTokenValue("validrefreshtoken")

			usecase := &fakeUsecase{}
			authHandler := NewAuthHandler(usecase)

			body := `{
				"refresh_token":"validrefreshtoken"
			}`
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(
				http.MethodPost, "/", strings.NewReader(body),
			)

			authHandler.Logout(c)
			c.Writer.WriteHeaderNow()
			if rec.Code != http.StatusNoContent {
				t.Errorf(
					"Logout(c) code = %v, want %v",
					rec.Code, http.StatusNoContent,
				)
			}
			want := ``
			if rec.Body.String() != want {
				t.Errorf("Logout(c) body = %v, want %v", rec.Body.String(), want)
			}

			if usecase.logoutedRefreshTokenValue != logoutedRefreshTokenValue {
				t.Errorf(
					"Logout(c) usecase args = %v, want %v",
					usecase.logoutedRefreshTokenValue,
					logoutedRefreshTokenValue,
				)
			}
		},
	)

	t.Run(
		"returns 400 when the request body is not valid JSON",
		func(t *testing.T) {
			usecase := &fakeUsecase{}
			authHandler := NewAuthHandler(usecase)

			body := `{
				"refresh_token":"validrefreshtoken",
			}`
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(
				http.MethodPost, "/", strings.NewReader(body),
			)

			authHandler.Logout(c)
			if rec.Code != http.StatusBadRequest {
				t.Errorf(
					"Logout(c) code = %v, want %v",
					rec.Code, http.StatusBadRequest,
				)
			}
			want := `"code":"INVALID_INPUT"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"Logout(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)

	t.Run(
		"returns 204 when the refresh token is invalid",
		func(t *testing.T) {
			usecase := &fakeUsecase{err: exception.ErrInvalid}
			authHandler := NewAuthHandler(usecase)

			body := `{
				"refresh_token":"invalidrefreshtoken"
			}`
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(
				http.MethodPost, "/", strings.NewReader(body),
			)

			authHandler.Logout(c)
			c.Writer.WriteHeaderNow()
			if rec.Code != http.StatusNoContent {
				t.Errorf(
					"Logout(c) code = %v, want %v",
					rec.Code, http.StatusNoContent,
				)
			}
			want := ``
			if rec.Body.String() != want {
				t.Errorf(
					"Logout(c) body = %v, want %v",
					rec.Body.String(), want,
				)
			}
		},
	)

	t.Run(
		"returns 500 when the usecase returns an unexpected error",
		func(t *testing.T) {
			usecase := &fakeUsecase{err: errors.New("unexpected")}
			authHandler := NewAuthHandler(usecase)

			body := `{
				"refresh_token":"validrefreshtoken"
			}`
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(
				http.MethodPost, "/", strings.NewReader(body),
			)

			authHandler.Logout(c)
			if rec.Code != http.StatusInternalServerError {
				t.Errorf(
					"Logout(c) code = %v, want %v",
					rec.Code, http.StatusInternalServerError,
				)
			}
			want := `"code":"INTERNAL_SERVER_ERROR"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"Logout(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)
}

func TestRefresh(t *testing.T) {
	t.Run(
		"passes the refresh token to the usecase and returns 200 when the request is valid",
		func(t *testing.T) {
			atv := auth.AccessTokenValue("validaccesstoken")
			rtv := auth.RefreshTokenValue("validrefreshtoken")
			at := auth.AccessToken{
				Value: atv,
			}
			rt := auth.RefreshToken{
				Value: rtv,
			}

			usecase := &fakeUsecase{accessToken: &at, refreshToken: &rt}
			authHandler := NewAuthHandler(usecase)

			body := `{
				"refresh_token":"validrefreshtoken"
			}`
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(
				http.MethodPost, "/", strings.NewReader(body),
			)

			authHandler.Refresh(c)
			if rec.Code != http.StatusOK {
				t.Errorf(
					"Refresh(c) code = %v, want %v",
					rec.Code, http.StatusOK,
				)
			}
			want := `{"access_token":"validaccesstoken","refresh_token":"validrefreshtoken"}`
			if rec.Body.String() != want {
				t.Errorf("Refresh(c) body = %v, want %v", rec.Body.String(), want)
			}

			if usecase.refreshTokenValue != rtv {
				t.Errorf(
					"Refresh(c) usecase args = %v, want %v",
					usecase.refreshTokenValue,
					rtv,
				)
			}
		},
	)

	t.Run(
		"returns 400 when the request body is not valid JSON",
		func(t *testing.T) {
			usecase := &fakeUsecase{}
			authHandler := NewAuthHandler(usecase)

			body := `{
				"refresh_token":"validrefreshtoken",
			}`
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(
				http.MethodPost, "/", strings.NewReader(body),
			)

			authHandler.Refresh(c)
			if rec.Code != http.StatusBadRequest {
				t.Errorf(
					"Refresh(c) code = %v, want %v",
					rec.Code, http.StatusBadRequest,
				)
			}
			want := `"code":"INVALID_INPUT"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"Refresh(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)

	t.Run(
		"returns 401 when the refresh token is invalid",
		func(t *testing.T) {
			usecase := &fakeUsecase{err: exception.ErrInvalid}
			authHandler := NewAuthHandler(usecase)

			body := `{
				"refresh_token":"invalidrefreshtoken"
			}`
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(
				http.MethodPost, "/", strings.NewReader(body),
			)

			authHandler.Refresh(c)
			if rec.Code != http.StatusUnauthorized {
				t.Errorf(
					"Refresh(c) code = %v, want %v",
					rec.Code, http.StatusUnauthorized,
				)
			}
			want := `"code":"INVALID_REFRESH_TOKEN"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"Refresh(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)

	t.Run(
		"returns 500 when the usecase returns an unexpected error",
		func(t *testing.T) {
			usecase := &fakeUsecase{err: errors.New("unexpected")}
			authHandler := NewAuthHandler(usecase)

			body := `{
				"refresh_token":"invalidrefreshtoken"
			}`
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(
				http.MethodPost, "/", strings.NewReader(body),
			)

			authHandler.Refresh(c)
			if rec.Code != http.StatusInternalServerError {
				t.Errorf(
					"Refresh(c) code = %v, want %v",
					rec.Code, http.StatusInternalServerError,
				)
			}
			want := `"code":"INTERNAL_SERVER_ERROR"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"Refresh(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)
}
