package auth

import "time"

type RefreshToken struct {
    ID        string
    UserID    string
    Hash      string
    ExpiresAt time.Time
    RevokedAt *time.Time
}
