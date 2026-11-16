package user

import (
	"fmt"
	"time"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/mail"
)

const EmailChangeTokenTTL = 24 * time.Hour

type EmailChangeToken struct {
	ID        EmailChangeTokenID
	UserID    ID
	NewEmail  Email
	Language  mail.Language
	Value     EmailChangeTokenValue
	Hash      EmailChangeTokenHash
	ExpiresAt time.Time
	CreatedAt time.Time
}

type EmailChangeTokenID uint
type EmailChangeTokenValue string
type EmailChangeTokenHash string

func NewEmailChangeTokenValue(value string) (EmailChangeTokenValue, error) {
	if value == "" {
		return "", fmt.Errorf("%w: token is required", exception.ErrInvalid)
	}
	return EmailChangeTokenValue(value), nil
}
