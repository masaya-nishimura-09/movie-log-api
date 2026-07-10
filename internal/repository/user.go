package repository

import (
	"context"

	"github.com/masaya-nishimura-09/movie-log-api/internal/model"
	"gorm.io/gorm"
)

type UserRepo interface {
	Login(ctx context.Context, u *model.User) error
	Logout(ctx context.Context) error
    CreateUser(ctx context.Context, user *model.User) error
	UpdateUser(ctx context.Context, user *model.User) error
	DeleteUser(ctx context.Context) error
}

type userRepo struct { 
    db *gorm.DB 
}

func NewUserRepo(db *gorm.DB) UserRepo { return userRepo{db} }

func (r *userRepo) Login(ctx context.Context, u *model.User) error {
    return r.db.WithContext(ctx).Create(u).Error
}

func (r *userRepo) Logout(ctx context.Context) error {
    return nil
}

func (r *userRepo) CreateUser(ctx context.Context, u *model.User) error {
    return r.db.WithContext(ctx).Create(u).Error
}

func (r *userRepo) UpdateUser(ctx context.Context, u *model.User) error {
    return r.db.WithContext(ctx).Create(u).Error
}

func (r *userRepo) DeleteUser(ctx context.Context) error {
    return nil
}
