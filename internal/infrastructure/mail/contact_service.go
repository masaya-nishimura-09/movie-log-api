package mail

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/contact"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/mail"
	"github.com/resend/resend-go/v4"
)

type contactService struct {
	client  *resend.Client
	from    mail.Address
	logoURL string
}

func NewContactService(client *resend.Client, from mail.Address, webBaseURL *url.URL) contact.Service {
	return &contactService{
		client:  client,
		from:    from,
		logoURL: webBaseURL.JoinPath("mail", "logo.png").String(),
	}
}

func (cs *contactService) Send(
	ctx context.Context,
	to mail.Address,
	c contact.Contact,
) error {
	receivedAt := time.Now()

	raw := make([]byte, 2)
	if _, err := rand.Read(raw); err != nil {
		return fmt.Errorf("generate receipt number: %w", err)
	}
	receiptNumber := fmt.Sprintf(
		"CL-%s-%s",
		receivedAt.In(jst).Format("20060102"),
		strings.ToUpper(hex.EncodeToString(raw)),
	)

	displayLanguage := "日本語"
	if c.Language == mail.LanguageEN {
		displayLanguage = "English"
	}

	return sendTemplate(ctx, cs.client, cs.from, cs.logoURL, to, mail.ReplyTo(c.Email), templateContact, templateData{
		Language:        mail.LanguageJA,
		Email:           c.Email,
		ReceiptNumber:   receiptNumber,
		ReceivedAt:      receivedAt.In(jst).Format("2006年1月2日 15:04"),
		DisplayLanguage: displayLanguage,
		Message:         string(c.Message),
	})
}
