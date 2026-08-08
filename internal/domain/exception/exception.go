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
	ErrInvalidToken = errors.New("invalid token")
)

var (
	ErrValidation = errors.New("validation error")
)
