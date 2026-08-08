package exception

import (
	"errors"
)

var (
	// Resource
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")

	// Invalid
	ErrInvalid = errors.New("invalid")

	// Authentication
	ErrUnauthenticated = errors.New("unauthenticated")
)
