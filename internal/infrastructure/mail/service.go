package mail

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/mail"
	"github.com/resend/resend-go/v4"
)

var jst = time.FixedZone("JST", 9*60*60)

type service struct {
	client  *resend.Client
	from    mail.Address
	logoURL string
}

func NewService(client *resend.Client, from mail.Address, webBaseURL *url.URL) mail.Service {
	return &service{
		client:  client,
		from:    from,
		logoURL: webBaseURL.JoinPath("mail", "logo.png").String(),
	}
}

func (s *service) SendEmailVerification(
	ctx context.Context,
	to mail.Address,
	language mail.Language,
	verificationURL mail.URL,
) error {
	return s.send(ctx, to, "", templateEmailVerification, templateData{
		Language: language,
		URL:      verificationURL,
	})
}

func (s *service) SendAlreadyRegistered(
	ctx context.Context,
	to mail.Address,
	language mail.Language,
	loginURL mail.URL,
) error {
	return s.send(ctx, to, "", templateAlreadyRegistered, templateData{
		Language: language,
		URL:      loginURL,
	})
}

func (s *service) SendUserVerified(
	ctx context.Context,
	to mail.Address,
	verifiedAt time.Time,
) error {
	return s.send(ctx, to, "", templateUserVerified, templateData{
		Language:   mail.LanguageJA,
		VerifiedAt: verifiedAt.In(jst).Format("2006年1月2日 15:04"),
	})
}

func (s *service) SendEmailChangeVerification(
	ctx context.Context,
	to mail.Address,
	language mail.Language,
	verificationURL mail.URL,
) error {
	return s.send(ctx, to, "", templateEmailChangeVerification, templateData{
		Language: language,
		URL:      verificationURL,
	})
}

func (s *service) SendEmailChangeAlreadyRegistered(
	ctx context.Context,
	to mail.Address,
	language mail.Language,
	loginURL mail.URL,
) error {
	return s.send(ctx, to, "", templateEmailChangeAlreadyRegistered, templateData{
		Language: language,
		URL:      loginURL,
	})
}

func (s *service) SendEmailChanged(
	ctx context.Context,
	to mail.Address,
	language mail.Language,
	newEmail mail.Address,
	contactURL mail.URL,
) error {
	return s.send(ctx, to, "", templateEmailChanged, templateData{
		Language: language,
		URL:      contactURL,
		NewEmail: newEmail,
	})
}

func (s *service) SendPasswordReset(
	ctx context.Context,
	to mail.Address,
	language mail.Language,
	resetURL mail.URL,
) error {
	return s.send(ctx, to, "", templatePasswordReset, templateData{
		Language: language,
		URL:      resetURL,
	})
}

func (s *service) SendPasswordChanged(
	ctx context.Context,
	to mail.Address,
	language mail.Language,
	contactURL mail.URL,
) error {
	return s.send(ctx, to, "", templatePasswordChanged, templateData{
		Language: language,
		URL:      contactURL,
	})
}

func (s *service) SendWithdrawn(
	ctx context.Context,
	to mail.Address,
	language mail.Language,
) error {
	return s.send(ctx, to, "", templateWithdrawn, templateData{
		Language: language,
	})
}

func (s *service) send(
	ctx context.Context,
	to mail.Address,
	replyTo mail.ReplyTo,
	name string,
	data templateData,
) error {
	data.LogoURL = s.logoURL

	subject, htmlBody, textBody, err := render(name, data)
	if err != nil {
		return fmt.Errorf("render mail: %w", err)
	}

	_, err = s.client.Emails.SendWithContext(ctx, &resend.SendEmailRequest{
		From:    string(s.from),
		To:      []string{string(to)},
		ReplyTo: string(replyTo),
		Subject: subject,
		Html:    htmlBody,
		Text:    textBody,
	})
	if err != nil {
		return fmt.Errorf("send mail: %w", err)
	}
	return nil
}
