package jwt

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
    secret []byte
)

func init() {
    s := os.Getenv("JWT_SECRET")
    if s == "" {
        panic("environment variable JWT_SECRET is required")
    }
    secret = []byte(s)
}

type Claims struct {
    UserID uint   
    Role   string 
    jwt.RegisteredClaims
}

func Generate(userID uint, role string) (string, error) {
    c := Claims{
        UserID: userID,
        Role:   role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }
    t := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
    return t.SignedString(secret)
}

func Validate(tokenStr string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
        return secret, nil
    })
    if err != nil || !token.Valid {
        return nil, err
    }
    return token.Claims.(*Claims), nil
}
