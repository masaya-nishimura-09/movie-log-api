package contact

import (
	"context"
	"fmt"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/contact"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/mail"
)

type Usecase interface {
	Send(
		ctx context.Context,
		email mail.Address,
		language mail.Language,
		message contact.Message,
	) error
}

type ContactUsecase struct {
	contactService contact.Service
	adminAddress   mail.Address
}

func NewContactUsecase(
	contactService contact.Service,
	adminAddress mail.Address,
) *ContactUsecase {
	return &ContactUsecase{
		contactService: contactService,
		adminAddress:   adminAddress,
	}
}

func (cu *ContactUsecase) Send(
	ctx context.Context,
	email mail.Address,
	language mail.Language,
	message contact.Message,
) error {
	c := contact.NewContact(email, language, message)
	if err := cu.contactService.Send(ctx, cu.adminAddress, c); err != nil {
		return fmt.Errorf("send contact: %w", err)
	}
	return nil
}
