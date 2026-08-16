package middleware

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

type fakeAuthUsecase struct {
	principal *auth.Principal
	err       error
}

func (u *fakeAuthUsecase) ValidateAccessToken(
	ctx context.Context,
	accessToken *auth.AccessToken,
) (*auth.Principal, error) {
	return u.principal, u.err
}

func (u *fakeAuthUsecase) Login(
	ctx context.Context,
	email user.Email,
	password user.Password,
) (*auth.AccessToken, *auth.RefreshToken, error) {
	return nil, nil, nil
}

func (u *fakeAuthUsecase) Logout(
	ctx context.Context,
	refreshTokenValue auth.RefreshTokenValue,
) error {
	return nil
}

func (u *fakeAuthUsecase) Refresh(
	ctx context.Context,
	refreshTokenValue auth.RefreshTokenValue,
) (*auth.AccessToken, *auth.RefreshToken, error) {
	return nil, nil, nil
}

type fakeUserUsecase struct {
	err error
}

func (u *fakeUserUsecase) GetByID(
	ctx context.Context,
	userID user.ID,
) (*user.User, error) {
	return nil, u.err
}

func (u *fakeUserUsecase) Register(
	ctx context.Context,
	username user.Username,
	email user.Email,
	password user.Password,
) (*user.User, error) {
	return nil, nil
}

func (u *fakeUserUsecase) UpdateUser(
	ctx context.Context,
	userID user.ID,
	username user.Username,
	email user.Email,
	password user.Password,
) (*user.User, error) {
	return nil, nil
}

func (u *fakeUserUsecase) DeleteUser(
	ctx context.Context,
	userID user.ID,
) error {
	return nil
}

func TestJWTAuth(t *testing.T) {
	t.Run(
		"calls the next handler with the user ID and role in the context when the access token is valid",
		func(t *testing.T) {
			userID := user.ID(1)
			role := user.RoleUser
			principal := auth.Principal{
				UserID: userID,
				Role:   role,
			}
			au := &fakeAuthUsecase{principal: &principal}
			uu := &fakeUserUsecase{}
			r := gin.New()

			called := false
			var gotUserID any
			var gotRole any

			r.Use(JWTAuth(au, uu))
			r.GET("/", func(c *gin.Context) {
				called = true
				gotUserID, _ = c.Get("userID")
				gotRole, _ = c.Get("role")
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Authorization", "Bearer validateaccesstoken")
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if !called {
				t.Errorf(
					"JWTAuth() called = %v, want %v",
					called, true,
				)
			}

			if gotUserID != userID ||
				gotRole != role {
				t.Errorf(
					"JWTAuth() userID = %v, role = %v, want %v, %v",
					gotUserID, gotRole, userID, role,
				)
			}
		},
	)

	t.Run(
		"returns 401 without calling the next handler when the Authorization header is not a Bearer token",
		func(t *testing.T) {
			au := &fakeAuthUsecase{}
			uu := &fakeUserUsecase{}
			r := gin.New()

			called := false

			r.Use(JWTAuth(au, uu))
			r.GET("/", func(c *gin.Context) {
				called = true
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if called {
				t.Errorf(
					"JWTAuth() called = %v, want %v",
					called, false,
				)
			}

			if rec.Code != http.StatusUnauthorized {
				t.Errorf(
					"JWTAuth() code = %v, want %v",
					rec.Code, http.StatusUnauthorized,
				)
			}
			want := `"code":"UNAUTHENTICATED"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"JWTAuth() body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)

	t.Run(
		"returns 401 without calling the next handler when the access token is invalid",
		func(t *testing.T) {
			au := &fakeAuthUsecase{err: exception.ErrInvalid}
			uu := &fakeUserUsecase{}
			r := gin.New()

			called := false

			r.Use(JWTAuth(au, uu))
			r.GET("/", func(c *gin.Context) {
				called = true
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Authorization", "Bearer invalidateaccesstoken")
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if called {
				t.Errorf(
					"JWTAuth() called = %v, want %v",
					called, false,
				)
			}

			if rec.Code != http.StatusUnauthorized {
				t.Errorf(
					"JWTAuth() code = %v, want %v",
					rec.Code, http.StatusUnauthorized,
				)
			}
			want := `"code":"INVALID_ACCESS_TOKEN"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"JWTAuth() body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)

	t.Run(
		"returns 401 without calling the next handler when the user does not exist",
		func(t *testing.T) {
			userID := user.ID(1)
			role := user.RoleUser
			principal := auth.Principal{
				UserID: userID,
				Role:   role,
			}
			au := &fakeAuthUsecase{principal: &principal}
			uu := &fakeUserUsecase{err: exception.ErrNotFound}
			r := gin.New()

			called := false

			r.Use(JWTAuth(au, uu))
			r.GET("/", func(c *gin.Context) {
				called = true
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Authorization", "Bearer validateaccesstoken")
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if called {
				t.Errorf(
					"JWTAuth() called = %v, want %v",
					called, false,
				)
			}

			if rec.Code != http.StatusUnauthorized {
				t.Errorf(
					"JWTAuth() code = %v, want %v",
					rec.Code, http.StatusUnauthorized,
				)
			}
			want := `"code":"UNAUTHENTICATED"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"JWTAuth() body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)

	t.Run(
		"returns 500 without calling the next handler when the user usecase returns an unexpected error",
		func(t *testing.T) {
			userID := user.ID(1)
			role := user.RoleUser
			principal := auth.Principal{
				UserID: userID,
				Role:   role,
			}
			au := &fakeAuthUsecase{principal: &principal}
			uu := &fakeUserUsecase{err: errors.New("unexpected")}
			r := gin.New()

			called := false

			r.Use(JWTAuth(au, uu))
			r.GET("/", func(c *gin.Context) {
				called = true
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Authorization", "Bearer validateaccesstoken")
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if called {
				t.Errorf(
					"JWTAuth() called = %v, want %v",
					called, false,
				)
			}

			if rec.Code != http.StatusInternalServerError {
				t.Errorf(
					"JWTAuth() code = %v, want %v",
					rec.Code, http.StatusInternalServerError,
				)
			}
			want := `"code":"INTERNAL_SERVER_ERROR"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"JWTAuth() body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)
}
