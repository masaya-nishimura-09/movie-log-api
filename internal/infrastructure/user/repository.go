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

type userRepository struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) user.UserRepository {
	return &userRepository{db}
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

func (ur *userRepository) GetByID(
	ctx context.Context,
	userID user.ID,
) (*user.User, error) {
	var dto userDTO
	result := ur.db.WithContext(ctx).First(&dto, userID)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, exception.ErrNotFound
	}
	if result.Error != nil {
		return nil, fmt.Errorf("get user by id: %w", result.Error)
	}
	return fromDTO(&dto), nil
}

func (ur *userRepository) GetByEmail(
	ctx context.Context,
	email user.Email,
) (*user.User, error) {
	var dto userDTO
	result := ur.db.WithContext(ctx).Where("email = ?", string(email)).First(&dto)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, exception.ErrNotFound
	}
	if result.Error != nil {
		return nil, fmt.Errorf("get user by email: %w", result.Error)
	}
	return fromDTO(&dto), nil
}

func (ur *userRepository) Create(
	ctx context.Context,
	u *user.User,
) error {
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()

	dto := toDTO(u)
	result := ur.db.WithContext(ctx).Create(&dto)
	if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
		return exception.ErrAlreadyExists
	}
	if result.Error != nil {
		return fmt.Errorf("create user: %w", result.Error)
	}
	u.ID = user.ID(dto.ID)
	return nil
}

func (ur *userRepository) Update(
	ctx context.Context,
	u *user.User,
) error {
	u.UpdatedAt = time.Now()

	dto := toDTO(u)
	result := ur.db.WithContext(ctx).
		Model(&userDTO{}).
		Where("id = ?", u.ID).
		Select("username", "email", "hashed_password", "updated_at").
		Updates(userDTO{
			Username:       dto.Username,
			Email:          dto.Email,
			HashedPassword: dto.HashedPassword,
			UpdatedAt:      dto.UpdatedAt,
		})
	if result.Error != nil {
		return fmt.Errorf("update user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return exception.ErrNotFound
	}
	return nil
}

func (ur *userRepository) Delete(ctx context.Context, userID user.ID) error {
	result := ur.db.WithContext(ctx).Delete(&userDTO{}, userID)
	if result.Error != nil {
		return fmt.Errorf("delete user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return exception.ErrNotFound
	}
	return nil
}
