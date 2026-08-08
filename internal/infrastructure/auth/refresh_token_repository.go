package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/auth"
	"gorm.io/gorm"
)

type refreshTokenRepository struct {
	db *gorm.DB
	ttl time.Duration
}

func NewRefreshTokenRepo(db *gorm.DB, ttl time.Duration) auth.RefreshTokenRepository {
	return &refreshTokenRepository{db: db, ttl: ttl}
}

type refreshTokenDTO struct {
	ID uint `gorm:"primaryKey"`
	UserID uint
	Role string
    Hash string
    ExpiresAt time.Time
    CreatedAt time.Time
    RevokedAt *time.Time
}

func (refreshTokenDTO) TableName() string {
	return "refresh_tokens"
}

func toDTO(rt *auth.RefreshToken) refreshTokenDTO {
	return refreshTokenDTO {
		UserID: uint(rt.Principal.UserID),
		Role: string(rt.Principal.Role),
		Hash: string(rt.Hash),
		ExpiresAt: rt.ExpiresAt,
		RevokedAt: rt.RevokedAt,
	}
}

func (r *refreshTokenRepository) Create(
	ctx context.Context, 
	principal *auth.Principal, 
) (*auth.RefreshToken, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return nil, fmt.Errorf("generate refresh token value: %w", err)
	}
	value := hex.EncodeToString(raw)

	sum := sha256.Sum256([]byte(value))
	hash := hex.EncodeToString(sum[:])

	refreshToken := auth.RefreshToken{
		Principal: *principal,
		Value: auth.RefreshTokenValue(value),
		Hash: auth.RefreshTokenHash(hash),
		ExpiresAt: time.Now().Add(r.ttl),
	}

	dto := toDTO(&refreshToken)
	result := r.db.WithContext(ctx).Create(&dto)
	if result.Error != nil {
		return nil, fmt.Errorf("create refresh token: %w", result.Error)
	}
	return &refreshToken, nil
}
