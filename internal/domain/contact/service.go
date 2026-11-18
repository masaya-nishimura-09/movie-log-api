package contact

import (
	"context"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/mail"
)

type Service interface {
	Send(ctx context.Context, to mail.Address, c Contact) error
}
