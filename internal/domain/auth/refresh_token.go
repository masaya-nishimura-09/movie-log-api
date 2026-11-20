package auth

import "time"

const RefreshTokenReuseInterval = 10 * time.Second

type RefreshToken struct {
	ID        RefreshTokenID
	Principal Principal
	Value     RefreshTokenValue
	Hash      RefreshTokenHash
	ExpiresAt time.Time
	CreatedAt time.Time
	RevokedAt *time.Time
	RotatedAt *time.Time
}

type RefreshTokenID uint
type RefreshTokenValue string
type RefreshTokenHash string
