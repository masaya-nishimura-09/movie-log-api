package user

import (
	"context"
	"errors"
	"fmt"

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

func (r *userRepository) GetByID(
	ctx context.Context, 
	userID user.ID,
) (*user.User, error) {
	var u user.User
	result := r.db.WithContext(ctx).First(&u, userID)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, exception.ErrUserNotFound
	}
	if result.Error != nil {
		return nil, fmt.Errorf("get user by id: %w", result.Error)
	}
	return &u, nil
}

func (r *userRepository) GetByEmail(
	ctx context.Context, 
	email user.Email,
) (*user.User, error) {
	var u user.User
	result := r.db.WithContext(ctx).Where("email = ?", email).First(&u)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, exception.ErrUserNotFound
	}
	if result.Error != nil {
		return nil, fmt.Errorf("get user by email: %w", result.Error)
	}
	return &u, nil
}

func (r *userRepository) Create(
	ctx context.Context, 
	u *user.User,
) error {
	result := r.db.WithContext(ctx).Create(u)
	if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
		return exception.ErrUserAlreadyExists
	}
	if result.Error != nil {
		return fmt.Errorf("create user: %w", result.Error)
	}
	return nil
}

func (r *userRepository) Update(
	ctx context.Context, 
	u *user.User,
) error {
	result := r.db.WithContext(ctx).
		Model(&user.User{}).
		Where("id = ?", u.ID).
		Select("username", "email", "hashed_password").
		Updates(user.User{
			Username: u.Username, 
			Email: u.Email,
			HashedPassword: u.HashedPassword,
		})
	if result.Error != nil {
		return fmt.Errorf("update user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return exception.ErrUserNotFound
	}
	return nil
}

func (r *userRepository) Delete(ctx context.Context, userID user.ID) error {
	result := r.db.WithContext(ctx).Delete(&user.User{}, userID)
	if result.Error != nil {
		return fmt.Errorf("delete user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return exception.ErrUserNotFound
	}
	return nil
}
