package contact

import (
	"context"
	"errors"
	"testing"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/contact"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/mail"
)

type fakeService struct {
	err    error
	sentTo mail.Address
	sent   contact.Contact
}

func (s *fakeService) Send(
	ctx context.Context,
	to mail.Address,
	c contact.Contact,
) error {
	s.sentTo = to
	s.sent = c
	return s.err
}

func TestSend(t *testing.T) {
	adminAddress := mail.Address("admin@example.com")
	email := mail.Address("user@example.com")
	language := mail.LanguageEN
	message := contact.Message("退会できません")

	t.Run(
		"sends the contact to the admin address",
		func(t *testing.T) {
			service := &fakeService{}
			cu := NewContactUsecase(service, adminAddress)

			ctx := context.Background()

			if err := cu.Send(ctx, email, language, message); err != nil {
				t.Fatalf(
					"Send(ctx, %v, %v, %v) error = %v",
					email, language, message, err,
				)
			}

			want := contact.NewContact(email, language, message)
			if service.sentTo != adminAddress || service.sent != want {
				t.Errorf(
					"Send(ctx, %v, %v, %v) sent %v to %v, want %v to %v",
					email, language, message, service.sent, service.sentTo, want, adminAddress,
				)
			}
		},
	)

	t.Run(
		"returns an error when sending fails",
		func(t *testing.T) {
			service := &fakeService{err: errors.New("send mail")}
			cu := NewContactUsecase(service, adminAddress)

			ctx := context.Background()

			if err := cu.Send(ctx, email, language, message); err == nil {
				t.Fatalf(
					"Send(ctx, %v, %v, %v) error = nil, want an error",
					email, language, message,
				)
			}
		},
	)
}
