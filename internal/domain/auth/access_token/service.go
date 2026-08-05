package auth

import (
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/model"
)

type TokenRepository interface {
	Generate(principal *model.Principal) (model.Token, error)
	Validate(token model.Token) (*model.Principal, error)
}
