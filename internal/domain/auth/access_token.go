package auth

import "time"

type AccessToken struct {
    Value AccessTokenValue
	Principal Principal
	ExpiresAt time.Time
}

type AccessTokenValue string
