package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
	"gorm.io/gorm"
)

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) user.Repository {
	return &repository{db}
}

type userDTO struct {
	ID             uint `gorm:"primaryKey"`
	Username       string
	Email          string
	HashedPassword string
	Role           string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (userDTO) TableName() string {
	return "users"
}

func toDTO(u *user.User) userDTO {
	return userDTO{
		ID:             uint(u.ID),
		Username:       string(u.Username),
		Email:          string(u.Email),
		HashedPassword: string(u.HashedPassword),
		Role:           string(u.Role),
		CreatedAt:      u.CreatedAt,
		UpdatedAt:      u.UpdatedAt,
	}
}

func fromDTO(dto *userDTO) *user.User {
	return &user.User{
		ID:             user.ID(dto.ID),
		Username:       user.Username(dto.Username),
		Email:          user.Email(dto.Email),
		HashedPassword: user.HashedPassword(dto.HashedPassword),
		Role:           user.Role(dto.Role),
		CreatedAt:      dto.CreatedAt,
		UpdatedAt:      dto.UpdatedAt,
	}
}

func (r *repository) GetByID(
	ctx context.Context,
	userID user.ID,
) (*user.User, error) {
	var dto userDTO
	result := r.db.WithContext(ctx).First(&dto, userID)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, exception.ErrNotFound
	}
	if result.Error != nil {
		return nil, fmt.Errorf("get user by id: %w", result.Error)
	}
	return fromDTO(&dto), nil
}

func (r *repository) GetByEmail(
	ctx context.Context,
	email user.Email,
) (*user.User, error) {
	var dto userDTO
	result := r.db.WithContext(ctx).Where("email = ?", string(email)).First(&dto)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, exception.ErrNotFound
	}
	if result.Error != nil {
		return nil, fmt.Errorf("get user by email: %w", result.Error)
	}
	return fromDTO(&dto), nil
}

func (r *repository) Create(
	ctx context.Context,
	u *user.User,
) error {
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()

	dto := toDTO(u)
	result := r.db.WithContext(ctx).Create(&dto)
	if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
		return exception.ErrAlreadyExists
	}
	if result.Error != nil {
		return fmt.Errorf("create user: %w", result.Error)
	}
	u.ID = user.ID(dto.ID)
	return nil
}

func (r *repository) UpdateUsername(
	ctx context.Context,
	u *user.User,
) error {
	u.UpdatedAt = time.Now()

	dto := toDTO(u)
	result := r.db.WithContext(ctx).
		Model(&userDTO{}).
		Where("id = ?", u.ID).
		Select("username", "updated_at").
		Updates(userDTO{
			Username:  dto.Username,
			UpdatedAt: dto.UpdatedAt,
		})
	if result.Error != nil {
		return fmt.Errorf("update username: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return exception.ErrNotFound
	}
	return nil
}

func (r *repository) UpdateEmail(
	ctx context.Context,
	u *user.User,
) error {
	u.UpdatedAt = time.Now()

	dto := toDTO(u)
	result := r.db.WithContext(ctx).
		Model(&userDTO{}).
		Where("id = ?", u.ID).
		Select("email", "updated_at").
		Updates(userDTO{
			Email:     dto.Email,
			UpdatedAt: dto.UpdatedAt,
		})
	if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
		return exception.ErrAlreadyExists
	}
	if result.Error != nil {
		return fmt.Errorf("update email: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return exception.ErrNotFound
	}
	return nil
}

func (r *repository) UpdatePassword(
	ctx context.Context,
	u *user.User,
) error {
	u.UpdatedAt = time.Now()

	dto := toDTO(u)
	result := r.db.WithContext(ctx).
		Model(&userDTO{}).
		Where("id = ?", u.ID).
		Select("hashed_password", "updated_at").
		Updates(userDTO{
			HashedPassword: dto.HashedPassword,
			UpdatedAt:      dto.UpdatedAt,
		})
	if result.Error != nil {
		return fmt.Errorf("update password: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return exception.ErrNotFound
	}
	return nil
}

func (r *repository) Delete(ctx context.Context, userID user.ID) error {
	result := r.db.WithContext(ctx).Delete(&userDTO{}, userID)
	if result.Error != nil {
		return fmt.Errorf("delete user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return exception.ErrNotFound
	}
	return nil
}
