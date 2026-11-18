package contact

import (
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/mail"
)

type Contact struct {
	Email    mail.Address
	Language mail.Language
	Message  Message
}

func NewContact(
	email mail.Address,
	language mail.Language,
	message Message,
) Contact {
	return Contact{
		Email:    email,
		Language: language,
		Message:  message,
	}
}
