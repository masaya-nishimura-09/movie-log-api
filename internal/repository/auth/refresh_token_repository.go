package auth

import (
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/auth"
	"gorm.io/gorm"
)

type refreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepo(db *gorm.DB) auth.RefreshTokenRepository {
	return &refreshTokenRepository{db}
}
