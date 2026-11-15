package user

import (
	"fmt"
	"time"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
)

const (
	EmailVerificationTokenTTL       = 24 * time.Hour
	EmailVerificationResendInterval = 60 * time.Second
)

type EmailVerificationToken struct {
	ID        EmailVerificationTokenID
	UserID    ID
	Value     EmailVerificationTokenValue
	Hash      EmailVerificationTokenHash
	ExpiresAt time.Time
	CreatedAt time.Time
}

type EmailVerificationTokenID uint
type EmailVerificationTokenValue string
type EmailVerificationTokenHash string

func NewEmailVerificationTokenValue(value string) (EmailVerificationTokenValue, error) {
	if value == "" {
		return "", fmt.Errorf("%w: token is required", exception.ErrInvalid)
	}
	return EmailVerificationTokenValue(value), nil
}
