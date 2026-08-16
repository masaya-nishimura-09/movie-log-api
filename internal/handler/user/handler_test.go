package user

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
)

type fakeUsecase struct {
	user               *user.User
	registeredUsername user.Username
	registeredEmail    user.Email
	registeredPassword user.Password
	updatedUserID      user.ID
	updatedUsername    user.Username
	updatedEmail       user.Email
	updatedPassword    user.Password
	deletedUserID      user.ID
	err                error
}

func (u *fakeUsecase) GetByID(
	ctx context.Context,
	userID user.ID,
) (*user.User, error) {
	return nil, nil
}

func (u *fakeUsecase) Register(
	ctx context.Context,
	username user.Username,
	email user.Email,
	password user.Password,
) (*user.User, error) {
	u.registeredUsername = username
	u.registeredEmail = email
	u.registeredPassword = password

	return u.user, u.err
}

func (u *fakeUsecase) UpdateUser(
	ctx context.Context,
	userID user.ID,
	username user.Username,
	email user.Email,
	password user.Password,
) (*user.User, error) {
	u.updatedUserID = userID
	u.updatedUsername = username
	u.updatedEmail = email
	u.updatedPassword = password

	return u.user, u.err
}

func (u *fakeUsecase) DeleteUser(
	ctx context.Context,
	userID user.ID,
) error {
	u.deletedUserID = userID
	return u.err
}

func TestCreateUser(t *testing.T) {
	t.Run(
		"passes the converted values to the usecase and returns 201 when the request is valid",
		func(t *testing.T) {
			userID := user.ID(1)
			username := user.Username("Test")
			email := user.Email("test@example.com")
			password := user.Password("testpassword")
			u := user.User{
				ID:       userID,
				Username: username,
				Email:    email,
			}
			usecase := &fakeUsecase{user: &u}
			userHandler := NewUserHandler(usecase)

			body := `{
				"name":"Test",
				"email":"test@example.com",
				"password":"testpassword"
			}`
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(
				http.MethodPost, "/", strings.NewReader(body),
			)

			userHandler.CreateUser(c)
			if rec.Code != http.StatusCreated {
				t.Errorf(
					"CreateUser(c) code = %v, want %v",
					rec.Code, http.StatusCreated,
				)
			}
			want := `{"email":"test@example.com","user_id":"1","username":"Test"}`
			if rec.Body.String() != want {
				t.Errorf("CreateUser(c) body = %v, want %v", rec.Body.String(), want)
			}

			if usecase.registeredUsername != username ||
				usecase.registeredEmail != email ||
				usecase.registeredPassword != password {
				t.Errorf(
					"CreateUser(c) usecase args = %v, %v, %v, want %v, %v, %v",
					usecase.registeredUsername,
					usecase.registeredEmail,
					usecase.registeredPassword,
					username, email, password,
				)
			}
		},
	)

	t.Run(
		"returns 400 when the request body is not valid JSON",
		func(t *testing.T) {
			usecase := &fakeUsecase{}
			userHandler := NewUserHandler(usecase)

			body := `{
				"name":"Test",
				"email":"test@example.com",
				"password":"testpassword",
			}`
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(
				http.MethodPost, "/", strings.NewReader(body),
			)

			userHandler.CreateUser(c)
			if rec.Code != http.StatusBadRequest {
				t.Errorf(
					"CreateUser(c) code = %v, want %v",
					rec.Code, http.StatusBadRequest,
				)
			}
			want := `"code":"INVALID_INPUT","message":"malformed request body"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"CreateUser(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)

	t.Run(
		"returns 400 when the email is not a valid address",
		func(t *testing.T) {
			usecase := &fakeUsecase{}
			userHandler := NewUserHandler(usecase)

			body := `{
				"name":"Test",
				"email":"invalid",
				"password":"testpassword"
			}`
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(
				http.MethodPost, "/", strings.NewReader(body),
			)

			userHandler.CreateUser(c)
			if rec.Code != http.StatusBadRequest {
				t.Errorf(
					"CreateUser(c) code = %v, want %v",
					rec.Code, http.StatusBadRequest,
				)
			}
			want := `"code":"INVALID_INPUT"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"CreateUser(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)

	t.Run(
		"returns 409 when the email is already registered",
		func(t *testing.T) {
			usecase := &fakeUsecase{err: exception.ErrAlreadyExists}
			userHandler := NewUserHandler(usecase)

			body := `{
				"name":"Test",
				"email":"test@example.com",
				"password":"testpassword"
			}`
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(
				http.MethodPost, "/", strings.NewReader(body),
			)

			userHandler.CreateUser(c)
			if rec.Code != http.StatusConflict {
				t.Errorf(
					"CreateUser(c) code = %v, want %v",
					rec.Code, http.StatusConflict,
				)
			}
			want := `"code":"USER_ALREADY_EXISTS"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"CreateUser(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)

	t.Run(
		"returns 500 when the usecase returns an unexpected error",
		func(t *testing.T) {
			usecase := &fakeUsecase{err: errors.New("unexpected")}
			userHandler := NewUserHandler(usecase)

			body := `{
				"name":"Test",
				"email":"test@example.com",
				"password":"testpassword"
			}`
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(
				http.MethodPost, "/", strings.NewReader(body),
			)

			userHandler.CreateUser(c)
			if rec.Code != http.StatusInternalServerError {
				t.Errorf(
					"CreateUser(c) code = %v, want %v",
					rec.Code, http.StatusInternalServerError,
				)
			}
			want := `"code":"INTERNAL_SERVER_ERROR"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"CreateUser(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)
}

func TestUpdateUser(t *testing.T) {
	t.Run(
		"passes the converted values to the usecase and returns 200 when the request is valid",
		func(t *testing.T) {
			userID := user.ID(1)
			username := user.Username("Test")
			email := user.Email("test@example.com")
			password := user.Password("testpassword")
			u := user.User{
				ID:       userID,
				Username: username,
				Email:    email,
			}
			usecase := &fakeUsecase{user: &u}
			userHandler := NewUserHandler(usecase)

			body := `{
				"name":"Test",
				"email":"test@example.com",
				"password":"testpassword"
			}`
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userID)
			c.Request = httptest.NewRequest(
				http.MethodPut, "/", strings.NewReader(body),
			)

			userHandler.UpdateUser(c)
			if rec.Code != http.StatusOK {
				t.Errorf(
					"UpdateUser(c) code = %v, want %v",
					rec.Code, http.StatusOK,
				)
			}
			want := `{"email":"test@example.com","user_id":"1","username":"Test"}`
			if rec.Body.String() != want {
				t.Errorf("UpdateUser(c) body = %v, want %v", rec.Body.String(), want)
			}

			if usecase.updatedUserID != userID ||
				usecase.updatedUsername != username ||
				usecase.updatedEmail != email ||
				usecase.updatedPassword != password {
				t.Errorf(
					"UpdateUser(c) usecase args = %v, %v, %v, %v, want %v, %v, %v, %v",
					usecase.updatedUserID,
					usecase.updatedUsername,
					usecase.updatedEmail,
					usecase.updatedPassword,
					userID, username, email, password,
				)
			}
		},
	)

	t.Run(
		"returns 500 when the authenticated user ID is missing from the context",
		func(t *testing.T) {
			usecase := &fakeUsecase{}
			userHandler := NewUserHandler(usecase)

			body := `{
				"name":"Test",
				"email":"test@example.com",
				"password":"testpassword"
			}`
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(
				http.MethodPut, "/", strings.NewReader(body),
			)

			userHandler.UpdateUser(c)
			if rec.Code != http.StatusInternalServerError {
				t.Errorf(
					"UpdateUser(c) code = %v, want %v",
					rec.Code, http.StatusInternalServerError,
				)
			}
			want := `"code":"INTERNAL_SERVER_ERROR"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"UpdateUser(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)

	t.Run(
		"returns 400 when the request body is not valid JSON",
		func(t *testing.T) {
			usecase := &fakeUsecase{}
			userHandler := NewUserHandler(usecase)

			userID := user.ID(1)
			body := `{
				"name":"Test",
				"email":"test@example.com",
				"password":"testpassword",
			}`
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userID)
			c.Request = httptest.NewRequest(
				http.MethodPut, "/", strings.NewReader(body),
			)

			userHandler.UpdateUser(c)
			if rec.Code != http.StatusBadRequest {
				t.Errorf(
					"UpdateUser(c) code = %v, want %v",
					rec.Code, http.StatusBadRequest,
				)
			}
			want := `"code":"INVALID_INPUT","message":"malformed request body"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"UpdateUser(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)

	t.Run(
		"returns 400 when the email is not a valid address",
		func(t *testing.T) {
			usecase := &fakeUsecase{}
			userHandler := NewUserHandler(usecase)

			userID := user.ID(1)
			body := `{
				"name":"Test",
				"email":"invalid",
				"password":"testpassword"
			}`
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userID)
			c.Request = httptest.NewRequest(
				http.MethodPut, "/", strings.NewReader(body),
			)

			userHandler.UpdateUser(c)
			if rec.Code != http.StatusBadRequest {
				t.Errorf(
					"UpdateUser(c) code = %v, want %v",
					rec.Code, http.StatusBadRequest,
				)
			}
			want := `"code":"INVALID_INPUT"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"UpdateUser(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)

	t.Run(
		"returns 404 when the user does not exist",
		func(t *testing.T) {
			usecase := &fakeUsecase{err: exception.ErrNotFound}
			userHandler := NewUserHandler(usecase)

			userID := user.ID(1)
			body := `{
				"name":"Test",
				"email":"test@example.com",
				"password":"testpassword"
			}`
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userID)
			c.Request = httptest.NewRequest(
				http.MethodPut, "/", strings.NewReader(body),
			)

			userHandler.UpdateUser(c)
			if rec.Code != http.StatusNotFound {
				t.Errorf(
					"UpdateUser(c) code = %v, want %v",
					rec.Code, http.StatusNotFound,
				)
			}
			want := `"code":"USER_NOT_FOUND"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"UpdateUser(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)

	t.Run(
		"returns 500 when the usecase returns an unexpected error",
		func(t *testing.T) {
			usecase := &fakeUsecase{err: errors.New("unexpected")}
			userHandler := NewUserHandler(usecase)

			userID := user.ID(1)
			body := `{
				"name":"Test",
				"email":"test@example.com",
				"password":"testpassword"
			}`
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userID)
			c.Request = httptest.NewRequest(
				http.MethodPut, "/", strings.NewReader(body),
			)

			userHandler.UpdateUser(c)
			if rec.Code != http.StatusInternalServerError {
				t.Errorf(
					"UpdateUser(c) code = %v, want %v",
					rec.Code, http.StatusInternalServerError,
				)
			}
			want := `"code":"INTERNAL_SERVER_ERROR"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"UpdateUser(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)
}

func TestDeleteUser(t *testing.T) {
	t.Run(
		"passes the authenticated user ID to the usecase and returns 204 when the request is valid",
		func(t *testing.T) {
			userID := user.ID(1)

			usecase := &fakeUsecase{}
			userHandler := NewUserHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userID)
			c.Request = httptest.NewRequest(
				http.MethodDelete, "/", nil,
			)

			userHandler.DeleteUser(c)
			c.Writer.WriteHeaderNow()
			if rec.Code != http.StatusNoContent {
				t.Errorf(
					"DeleteUser(c) code = %v, want %v",
					rec.Code, http.StatusNoContent,
				)
			}
			want := ``
			if rec.Body.String() != want {
				t.Errorf("DeleteUser(c) body = %v, want %v", rec.Body.String(), want)
			}

			if usecase.deletedUserID != userID {
				t.Errorf(
					"DeleteUser(c) usecase args = %v, want %v",
					usecase.deletedUserID,
					userID,
				)
			}
		},
	)

	t.Run(
		"returns 500 when the authenticated user ID is missing from the context",
		func(t *testing.T) {
			usecase := &fakeUsecase{}
			userHandler := NewUserHandler(usecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(
				http.MethodDelete, "/", nil,
			)

			userHandler.DeleteUser(c)
			if rec.Code != http.StatusInternalServerError {
				t.Errorf(
					"DeleteUser(c) code = %v, want %v",
					rec.Code, http.StatusInternalServerError,
				)
			}
			want := `"code":"INTERNAL_SERVER_ERROR"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"DeleteUser(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)

	t.Run(
		"returns 404 when the user does not exist",
		func(t *testing.T) {
			usecase := &fakeUsecase{err: exception.ErrNotFound}
			userHandler := NewUserHandler(usecase)

			userID := user.ID(1)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userID)
			c.Request = httptest.NewRequest(
				http.MethodDelete, "/", nil,
			)

			userHandler.DeleteUser(c)
			if rec.Code != http.StatusNotFound {
				t.Errorf(
					"DeleteUser(c) code = %v, want %v",
					rec.Code, http.StatusNotFound,
				)
			}
			want := `"code":"USER_NOT_FOUND"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"DeleteUser(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)

	t.Run(
		"returns 500 when the usecase returns an unexpected error",
		func(t *testing.T) {
			usecase := &fakeUsecase{err: errors.New("unexpected")}
			userHandler := NewUserHandler(usecase)

			userID := user.ID(1)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set("userID", userID)
			c.Request = httptest.NewRequest(
				http.MethodDelete, "/", nil,
			)

			userHandler.DeleteUser(c)
			if rec.Code != http.StatusInternalServerError {
				t.Errorf(
					"DeleteUser(c) code = %v, want %v",
					rec.Code, http.StatusInternalServerError,
				)
			}
			want := `"code":"INTERNAL_SERVER_ERROR"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf(
					"DeleteUser(c) body = %v, want to contain %v",
					rec.Body.String(), want,
				)
			}
		},
	)
}
