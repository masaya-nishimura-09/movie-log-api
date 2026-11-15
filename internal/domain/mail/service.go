package mail

import (
	"context"
)

type Service interface {
	Send(ctx context.Context, m Message) error
}
