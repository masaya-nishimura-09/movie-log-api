package user

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/mail"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
	"gorm.io/gorm"
)

type passwordResetTokenRepository struct {
	db  *gorm.DB
	ttl time.Duration
}

func NewPasswordResetTokenRepository(db *gorm.DB, ttl time.Duration) user.PasswordResetTokenRepository {
	return &passwordResetTokenRepository{db: db, ttl: ttl}
}

type passwordResetTokenDTO struct {
	ID        uint `gorm:"primaryKey"`
	UserID    uint
	Language  string
	Hash      string
	ExpiresAt time.Time
	CreatedAt time.Time
}

func (passwordResetTokenDTO) TableName() string {
	return "password_reset_tokens"
}

func toPasswordResetTokenDTO(prt *user.PasswordResetToken) passwordResetTokenDTO {
	return passwordResetTokenDTO{
		UserID:    uint(prt.UserID),
		Language:  string(prt.Language),
		Hash:      string(prt.Hash),
		ExpiresAt: prt.ExpiresAt,
		CreatedAt: prt.CreatedAt,
	}
}

func fromPasswordResetTokenDTO(dto *passwordResetTokenDTO) *user.PasswordResetToken {
	return &user.PasswordResetToken{
		ID:        user.PasswordResetTokenID(dto.ID),
		UserID:    user.ID(dto.UserID),
		Language:  mail.Language(dto.Language),
		Hash:      user.PasswordResetTokenHash(dto.Hash),
		ExpiresAt: dto.ExpiresAt,
		CreatedAt: dto.CreatedAt,
	}
}

func (prtr *passwordResetTokenRepository) Create(
	ctx context.Context,
	userID user.ID,
	language mail.Language,
) (*user.PasswordResetToken, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return nil, fmt.Errorf("generate password reset token value: %w", err)
	}
	value := hex.EncodeToString(raw)

	sum := sha256.Sum256([]byte(value))
	hash := hex.EncodeToString(sum[:])

	passwordResetToken := user.PasswordResetToken{
		UserID:    userID,
		Language:  language,
		Value:     user.PasswordResetTokenValue(value),
		Hash:      user.PasswordResetTokenHash(hash),
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(prtr.ttl),
	}

	dto := toPasswordResetTokenDTO(&passwordResetToken)
	result := prtr.db.WithContext(ctx).Create(&dto)
	if errors.Is(result.Error, gorm.ErrForeignKeyViolated) {
		return nil, exception.ErrNotFound
	}
	if result.Error != nil {
		return nil, fmt.Errorf("create password reset token: %w", result.Error)
	}
	passwordResetToken.ID = user.PasswordResetTokenID(dto.ID)
	return &passwordResetToken, nil
}

func (prtr *passwordResetTokenRepository) FindValidByValue(
	ctx context.Context,
	value user.PasswordResetTokenValue,
) (*user.PasswordResetToken, error) {
	sum := sha256.Sum256([]byte(value))
	hash := hex.EncodeToString(sum[:])

	var dto passwordResetTokenDTO
	result := prtr.db.WithContext(ctx).
		Where("hash = ?", hash).
		Where("expires_at > ?", time.Now()).
		First(&dto)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, exception.ErrInvalid
	}
	if result.Error != nil {
		return nil, fmt.Errorf("find password reset token: %w", result.Error)
	}

	return fromPasswordResetTokenDTO(&dto), nil
}

func (prtr *passwordResetTokenRepository) FindLatestByUserID(
	ctx context.Context,
	userID user.ID,
) (*user.PasswordResetToken, error) {
	var dto passwordResetTokenDTO
	result := prtr.db.WithContext(ctx).
		Where("user_id = ?", uint(userID)).
		Order("created_at DESC").
		First(&dto)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, exception.ErrNotFound
	}
	if result.Error != nil {
		return nil, fmt.Errorf("find latest password reset token: %w", result.Error)
	}

	return fromPasswordResetTokenDTO(&dto), nil
}

func (prtr *passwordResetTokenRepository) DeleteAllForUser(
	ctx context.Context,
	userID user.ID,
) error {
	result := prtr.db.WithContext(ctx).
		Where("user_id = ?", uint(userID)).
		Delete(&passwordResetTokenDTO{})
	if result.Error != nil {
		return fmt.Errorf("delete password reset tokens for user: %w", result.Error)
	}
	return nil
}

func (prtr *passwordResetTokenRepository) DeleteExpired(
	ctx context.Context,
) error {
	result := prtr.db.WithContext(ctx).
		Where("expires_at <= ?", time.Now()).
		Delete(&passwordResetTokenDTO{})
	if result.Error != nil {
		return fmt.Errorf("delete expired password reset tokens: %w", result.Error)
	}
	return nil
}
