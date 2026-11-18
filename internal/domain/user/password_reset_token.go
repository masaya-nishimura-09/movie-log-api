package user

import (
	"fmt"
	"time"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/mail"
)

const (
	PasswordResetTokenTTL        = 1 * time.Hour
	PasswordResetRequestInterval = 60 * time.Second
)

type PasswordResetToken struct {
	ID        PasswordResetTokenID
	UserID    ID
	Language  mail.Language
	Value     PasswordResetTokenValue
	Hash      PasswordResetTokenHash
	ExpiresAt time.Time
	CreatedAt time.Time
}

type PasswordResetTokenID uint
type PasswordResetTokenValue string
type PasswordResetTokenHash string

func NewPasswordResetTokenValue(value string) (PasswordResetTokenValue, error) {
	if value == "" {
		return "", fmt.Errorf("%w: token is required", exception.ErrInvalid)
	}
	return PasswordResetTokenValue(value), nil
}
