package auth

import (
	"context"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/auth"
	"gorm.io/gorm"
)

type refreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepo(db *gorm.DB) auth.RefreshTokenRepository {
	return &refreshTokenRepository{db}
}

func (r *refreshTokenRepository) Create(
	ctx context.Context, 
) error {
	return nil
}
