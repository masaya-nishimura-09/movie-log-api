package mail

import (
	"context"
	"fmt"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/mail"
	"github.com/resend/resend-go/v4"
)

type service struct {
	client *resend.Client
}

func NewService(client *resend.Client) mail.Service {
	return &service{client: client}
}

func (s *service) Send(ctx context.Context, m mail.Message) error {
	_, err := s.client.Emails.SendWithContext(ctx, &resend.SendEmailRequest{
		From:    string(m.From),
		To:      []string{string(m.To)},
		ReplyTo: string(m.ReplyTo),
		Subject: string(m.Subject),
		Html:    string(m.HTMLBody),
		Text:    string(m.TextBody),
	})
	if err != nil {
		return fmt.Errorf("send mail: %w", err)
	}
	return nil
}
