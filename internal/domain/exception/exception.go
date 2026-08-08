package exception

import (
	"errors"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

var (
	ErrInvalidToken              = errors.New("invalid token")
	ErrRefreshTokenAlreadyExists = errors.New("refresh token already exists")
)

var (
	ErrValidation = errors.New("validation error")
)
