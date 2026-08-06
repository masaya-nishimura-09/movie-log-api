package auth

import (
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
)

type Principal struct {
	UserID user.ID
	Role user.Role
}
