package auth

import "time"

type RefreshToken struct {
    ID        RefreshTokenID
    Principal Principal
    Value     RefreshTokenValue
    Hash      RefreshTokenHash
    ExpiresAt time.Time
    CreatedAt time.Time
    RevokedAt *time.Time
}

type RefreshTokenID uint
type RefreshTokenValue string
type RefreshTokenHash string
