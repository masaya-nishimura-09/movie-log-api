package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/model"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/repository"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) repository.UserRepository {
	return &userRepository{db}
}

func (r *userRepository) GetByID(ctx context.Context, userID model.UserID) (*model.User, error) {
	var user model.User
	result := r.db.WithContext(ctx).First(&user, userID)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, model.ErrUserNotFound
	}
	if result.Error != nil {
		return nil, fmt.Errorf("get user by id: %w", result.Error)
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email model.Email) (*model.User, error) {
	var user model.User
	result := r.db.WithContext(ctx).Where("email = ?", email).First(&user)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, model.ErrUserNotFound
	}
	if result.Error != nil {
		return nil, fmt.Errorf("get user by email: %w", result.Error)
	}
	return &user, nil
}

func (r *userRepository) Create(ctx context.Context, u *model.User) error {
	result := r.db.WithContext(ctx).Create(u)
	if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
		return model.ErrUserAlreadyExists
	}
	if result.Error != nil {
		return fmt.Errorf("create user: %w", result.Error)
	}
	return nil
}

func (r *userRepository) Update(ctx context.Context, u *model.User) error {
	result := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ?", u.ID).
		Select("username", "email").
		Updates(model.User{Username: u.Username, Email: u.Email})
	if result.Error != nil {
		return fmt.Errorf("update user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return model.ErrUserNotFound
	}
	return nil
}

func (r *userRepository) UpdatePassword(ctx context.Context, u *model.User) error {
	result := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ?", u.ID).
		Select("hashed_password").
		Update("hashed_password", u.HashedPassword)
	if result.Error != nil {
		return fmt.Errorf("update password: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return model.ErrUserNotFound
	}
	return nil
}

func (r *userRepository) Delete(ctx context.Context, userID model.UserID) error {
	result := r.db.WithContext(ctx).Delete(&model.User{}, userID)
	if result.Error != nil {
		return fmt.Errorf("delete user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return model.ErrUserNotFound
	}
	return nil
}
