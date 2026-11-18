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
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
	"gorm.io/gorm"
)

type emailVerificationTokenRepository struct {
	db  *gorm.DB
	ttl time.Duration
}

func NewEmailVerificationTokenRepository(db *gorm.DB, ttl time.Duration) user.EmailVerificationTokenRepository {
	return &emailVerificationTokenRepository{db: db, ttl: ttl}
}

type emailVerificationTokenDTO struct {
	ID        uint `gorm:"primaryKey"`
	UserID    uint
	Hash      string
	ExpiresAt time.Time
	CreatedAt time.Time
}

func (emailVerificationTokenDTO) TableName() string {
	return "email_verification_tokens"
}

func toEmailVerificationTokenDTO(evt *user.EmailVerificationToken) emailVerificationTokenDTO {
	return emailVerificationTokenDTO{
		UserID:    uint(evt.UserID),
		Hash:      string(evt.Hash),
		ExpiresAt: evt.ExpiresAt,
		CreatedAt: evt.CreatedAt,
	}
}

func fromEmailVerificationTokenDTO(dto *emailVerificationTokenDTO) *user.EmailVerificationToken {
	return &user.EmailVerificationToken{
		ID:        user.EmailVerificationTokenID(dto.ID),
		UserID:    user.ID(dto.UserID),
		Hash:      user.EmailVerificationTokenHash(dto.Hash),
		ExpiresAt: dto.ExpiresAt,
		CreatedAt: dto.CreatedAt,
	}
}

func (evtr *emailVerificationTokenRepository) Create(
	ctx context.Context,
	userID user.ID,
) (*user.EmailVerificationToken, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return nil, fmt.Errorf("generate email verification token value: %w", err)
	}
	value := hex.EncodeToString(raw)

	sum := sha256.Sum256([]byte(value))
	hash := hex.EncodeToString(sum[:])

	emailVerificationToken := user.EmailVerificationToken{
		UserID:    userID,
		Value:     user.EmailVerificationTokenValue(value),
		Hash:      user.EmailVerificationTokenHash(hash),
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(evtr.ttl),
	}

	dto := toEmailVerificationTokenDTO(&emailVerificationToken)
	result := evtr.db.WithContext(ctx).Create(&dto)
	if errors.Is(result.Error, gorm.ErrForeignKeyViolated) {
		return nil, exception.ErrNotFound
	}
	if result.Error != nil {
		return nil, fmt.Errorf("create email verification token: %w", result.Error)
	}
	emailVerificationToken.ID = user.EmailVerificationTokenID(dto.ID)
	return &emailVerificationToken, nil
}

func (evtr *emailVerificationTokenRepository) FindValidByValue(
	ctx context.Context,
	value user.EmailVerificationTokenValue,
) (*user.EmailVerificationToken, error) {
	sum := sha256.Sum256([]byte(value))
	hash := hex.EncodeToString(sum[:])

	var dto emailVerificationTokenDTO
	result := evtr.db.WithContext(ctx).
		Where("hash = ?", hash).
		Where("expires_at > ?", time.Now()).
		First(&dto)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, exception.ErrInvalid
	}
	if result.Error != nil {
		return nil, fmt.Errorf("find email verification token: %w", result.Error)
	}

	return fromEmailVerificationTokenDTO(&dto), nil
}

func (evtr *emailVerificationTokenRepository) FindLatestByUserID(
	ctx context.Context,
	userID user.ID,
) (*user.EmailVerificationToken, error) {
	var dto emailVerificationTokenDTO
	result := evtr.db.WithContext(ctx).
		Where("user_id = ?", uint(userID)).
		Order("created_at DESC").
		First(&dto)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, exception.ErrNotFound
	}
	if result.Error != nil {
		return nil, fmt.Errorf("find latest email verification token: %w", result.Error)
	}

	return fromEmailVerificationTokenDTO(&dto), nil
}

func (evtr *emailVerificationTokenRepository) DeleteAllForUser(
	ctx context.Context,
	userID user.ID,
) error {
	result := evtr.db.WithContext(ctx).
		Where("user_id = ?", uint(userID)).
		Delete(&emailVerificationTokenDTO{})
	if result.Error != nil {
		return fmt.Errorf("delete email verification tokens for user: %w", result.Error)
	}
	return nil
}
