package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/auth"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
	"gorm.io/gorm"
)

type refreshTokenRepository struct {
	db  *gorm.DB
	ttl time.Duration
}

func NewRefreshTokenRepo(db *gorm.DB, ttl time.Duration) auth.RefreshTokenRepository {
	return &refreshTokenRepository{db: db, ttl: ttl}
}

type refreshTokenDTO struct {
	ID        uint `gorm:"primaryKey"`
	UserID    uint
	Role      string
	Hash      string
	ExpiresAt time.Time
	CreatedAt time.Time
	RevokedAt *time.Time
}

func (refreshTokenDTO) TableName() string {
	return "refresh_tokens"
}

func toDTO(rt *auth.RefreshToken) refreshTokenDTO {
	return refreshTokenDTO{
		UserID:    uint(rt.Principal.UserID),
		Role:      string(rt.Principal.Role),
		Hash:      string(rt.Hash),
		ExpiresAt: rt.ExpiresAt,
		CreatedAt: rt.CreatedAt,
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
		Value:     auth.RefreshTokenValue(value),
		Hash:      auth.RefreshTokenHash(hash),
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(r.ttl),
	}

	dto := toDTO(&refreshToken)
	result := r.db.WithContext(ctx).Create(&dto)
	if result.Error != nil {
		return nil, fmt.Errorf("create refresh token: %w", result.Error)
	}
	refreshToken.ID = auth.RefreshTokenID(dto.ID)
	return &refreshToken, nil
}

func (r *refreshTokenRepository) FindValidByValue(
	ctx context.Context,
	value auth.RefreshTokenValue,
) (*auth.RefreshToken, error) {
	sum := sha256.Sum256([]byte(value))
	hash := hex.EncodeToString(sum[:])

	var dto refreshTokenDTO
	result := r.db.WithContext(ctx).
		Where("hash = ?", hash).
		Where("expires_at > ?", time.Now()).
		Where("revoked_at IS NULL").
		First(&dto)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, exception.ErrInvalid
	}
	if result.Error != nil {
		return nil, fmt.Errorf("find refresh token: %w", result.Error)
	}

	return &auth.RefreshToken{
		ID: auth.RefreshTokenID(dto.ID),
		Principal: auth.Principal{
			UserID: user.ID(dto.UserID),
			Role:   user.Role(dto.Role),
		},
		Hash:      auth.RefreshTokenHash(dto.Hash),
		ExpiresAt: dto.ExpiresAt,
		CreatedAt: dto.CreatedAt,
		RevokedAt: dto.RevokedAt,
	}, nil
}

func (r *refreshTokenRepository) Revoke(
	ctx context.Context,
	id auth.RefreshTokenID,
) error {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&refreshTokenDTO{}).
		Where("id = ?", id).
		Where("revoked_at IS NULL").
		Update("revoked_at", now)
	if result.Error != nil {
		return fmt.Errorf("revoke refresh token: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return exception.ErrInvalid
	}
	return nil
}

func (r *refreshTokenRepository) RevokeAllForUser(
	ctx context.Context,
	userID user.ID,
) error {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&refreshTokenDTO{}).
		Where("user_id = ?", uint(userID)).
		Where("revoked_at IS NULL").
		Update("revoked_at", now)
	if result.Error != nil {
		return fmt.Errorf("revoke all refresh tokens for user: %w", result.Error)
	}
	return nil
}
