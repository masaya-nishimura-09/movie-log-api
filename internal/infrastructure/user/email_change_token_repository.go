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

type emailChangeTokenRepository struct {
	db  *gorm.DB
	ttl time.Duration
}

func NewEmailChangeTokenRepository(db *gorm.DB, ttl time.Duration) user.EmailChangeTokenRepository {
	return &emailChangeTokenRepository{db: db, ttl: ttl}
}

type emailChangeTokenDTO struct {
	ID        uint `gorm:"primaryKey"`
	UserID    uint
	NewEmail  string
	Language  string
	Hash      string
	ExpiresAt time.Time
	CreatedAt time.Time
}

func (emailChangeTokenDTO) TableName() string {
	return "email_change_tokens"
}

func toEmailChangeTokenDTO(ect *user.EmailChangeToken) emailChangeTokenDTO {
	return emailChangeTokenDTO{
		UserID:    uint(ect.UserID),
		NewEmail:  string(ect.NewEmail),
		Language:  string(ect.Language),
		Hash:      string(ect.Hash),
		ExpiresAt: ect.ExpiresAt,
		CreatedAt: ect.CreatedAt,
	}
}

func fromEmailChangeTokenDTO(dto *emailChangeTokenDTO) *user.EmailChangeToken {
	return &user.EmailChangeToken{
		ID:        user.EmailChangeTokenID(dto.ID),
		UserID:    user.ID(dto.UserID),
		NewEmail:  user.Email(dto.NewEmail),
		Language:  mail.Language(dto.Language),
		Hash:      user.EmailChangeTokenHash(dto.Hash),
		ExpiresAt: dto.ExpiresAt,
		CreatedAt: dto.CreatedAt,
	}
}

func (ectr *emailChangeTokenRepository) Create(
	ctx context.Context,
	userID user.ID,
	newEmail user.Email,
	language mail.Language,
) (*user.EmailChangeToken, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return nil, fmt.Errorf("generate email change token value: %w", err)
	}
	value := hex.EncodeToString(raw)

	sum := sha256.Sum256([]byte(value))
	hash := hex.EncodeToString(sum[:])

	emailChangeToken := user.EmailChangeToken{
		UserID:    userID,
		NewEmail:  newEmail,
		Language:  language,
		Value:     user.EmailChangeTokenValue(value),
		Hash:      user.EmailChangeTokenHash(hash),
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(ectr.ttl),
	}

	dto := toEmailChangeTokenDTO(&emailChangeToken)
	result := ectr.db.WithContext(ctx).Create(&dto)
	if errors.Is(result.Error, gorm.ErrForeignKeyViolated) {
		return nil, exception.ErrNotFound
	}
	if result.Error != nil {
		return nil, fmt.Errorf("create email change token: %w", result.Error)
	}
	emailChangeToken.ID = user.EmailChangeTokenID(dto.ID)
	return &emailChangeToken, nil
}

func (ectr *emailChangeTokenRepository) FindValidByValue(
	ctx context.Context,
	value user.EmailChangeTokenValue,
) (*user.EmailChangeToken, error) {
	sum := sha256.Sum256([]byte(value))
	hash := hex.EncodeToString(sum[:])

	var dto emailChangeTokenDTO
	result := ectr.db.WithContext(ctx).
		Where("hash = ?", hash).
		Where("expires_at > ?", time.Now()).
		First(&dto)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, exception.ErrInvalid
	}
	if result.Error != nil {
		return nil, fmt.Errorf("find email change token: %w", result.Error)
	}

	return fromEmailChangeTokenDTO(&dto), nil
}

func (ectr *emailChangeTokenRepository) DeleteAllForUser(
	ctx context.Context,
	userID user.ID,
) error {
	result := ectr.db.WithContext(ctx).
		Where("user_id = ?", uint(userID)).
		Delete(&emailChangeTokenDTO{})
	if result.Error != nil {
		return fmt.Errorf("delete email change tokens for user: %w", result.Error)
	}
	return nil
}

func (ectr *emailChangeTokenRepository) DeleteExpired(
	ctx context.Context,
) error {
	result := ectr.db.WithContext(ctx).
		Where("expires_at <= ?", time.Now()).
		Delete(&emailChangeTokenDTO{})
	if result.Error != nil {
		return fmt.Errorf("delete expired email change tokens: %w", result.Error)
	}
	return nil
}
