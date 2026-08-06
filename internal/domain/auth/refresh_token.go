package auth

import "time"

type RefreshToken struct {
    Value      string
    ID        string
    UserID    string
    Hash      string
    ExpiresAt time.Time
    RevokedAt *time.Time
}
