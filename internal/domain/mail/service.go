package mail

import (
	"context"
	"time"
)

type Service interface {
	SendEmailVerification(ctx context.Context, to Address, language Language, verificationURL URL) error
	SendAlreadyRegistered(ctx context.Context, to Address, language Language, loginURL URL) error
	SendUserVerified(ctx context.Context, to Address, verifiedAt time.Time) error
	SendEmailChangeVerification(ctx context.Context, to Address, language Language, verificationURL URL) error
	SendEmailChangeAlreadyRegistered(ctx context.Context, to Address, language Language, loginURL URL) error
	SendEmailChanged(ctx context.Context, to Address, language Language, newEmail Address, contactURL URL) error
	SendPasswordReset(ctx context.Context, to Address, language Language, resetURL URL) error
	SendPasswordChanged(ctx context.Context, to Address, language Language, contactURL URL) error
	SendWithdrawn(ctx context.Context, to Address, language Language) error
}
