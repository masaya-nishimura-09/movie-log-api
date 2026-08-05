package user

import (
	"fmt"
	"regexp"
	"time"
)

type User struct {
	ID             UserID
	Username       Username
	Email          Email
	HashedPassword HashedPassword
	Role           Role
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type UserID uint
type Username string
type Email string
type Password string
type HashedPassword string
type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)+$")

func NewUsername(value string) (Username, error) {
	if value == "" {
		return "", fmt.Errorf("%w: username is required", ErrValidation)
	}
	if len(value) > 100 {
		return "", fmt.Errorf("%w: username must be at most 100 characters", ErrValidation)
	}
	return Username(value), nil
}

func ValidateEmail(email Email) error {
	if email == "" {
		return fmt.Errorf("%w: email is required", ErrValidation)
	}
	if !emailRegex.MatchString(string(email)) {
		return fmt.Errorf("%w: invalid email", ErrValidation)
	}
	return nil
}

func ValidatePassword(password Password) error {
	if password == "" {
		return fmt.Errorf("%w: password is required", ErrValidation)
	}
	if len(password) < 8 {
		return fmt.Errorf("%w: password must be at least 8 characters", ErrValidation)
	}
	if len(password) > 72 {
		return fmt.Errorf("%w: password must be at most 72 characters", ErrValidation)
	}
	return nil
}
