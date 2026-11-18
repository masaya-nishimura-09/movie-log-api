package mail

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/contact"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/mail"
)

func TestContactSend(t *testing.T) {
	t.Run(
		"sends the contact to the admin with the receipt number and the reply-to set to the sender",
		func(t *testing.T) {
			var got sendRequest
			srv := newTestServer(t, &got)
			defer srv.Close()

			cs := NewContactService(newTestClient(t, srv), mail.Address("support@example.com"), testWebBaseURL)
			to := mail.Address("admin@example.com")
			c := contact.NewContact(
				mail.Address("user@example.com"),
				mail.LanguageEN,
				contact.Message("1行目\n2行目"),
			)

			if err := cs.Send(context.Background(), to, c); err != nil {
				t.Fatalf("Send(ctx, %v, %v) error = %v", to, c, err)
			}

			if got.From != "support@example.com" ||
				len(got.To) != 1 || got.To[0] != string(to) ||
				got.ReplyTo != "user@example.com" {
				t.Errorf(
					"request body = %+v, want from support@example.com to %v with reply-to user@example.com",
					got, to,
				)
			}
			if !strings.HasPrefix(got.Subject, "[Cinelier お問い合わせ] #CL-") ||
				!strings.HasSuffix(got.Subject, " user@example.com") {
				t.Errorf("subject = %q, want the receipt number and the sender's email", got.Subject)
			}
			for _, want := range []string{
				"メールアドレス: user@example.com",
				"表示言語: English",
				"1行目\n2行目",
			} {
				if !strings.Contains(got.Text, want) {
					t.Errorf("text body = %s, want to contain %s", got.Text, want)
				}
			}
			if !strings.Contains(got.Html, "1行目\n2行目") || !strings.Contains(got.Html, "white-space:pre-wrap") {
				t.Errorf("html body = %s, want the message with its line breaks kept", got.Html)
			}
		},
	)

	t.Run(
		"escapes the message in the html body",
		func(t *testing.T) {
			var got sendRequest
			srv := newTestServer(t, &got)
			defer srv.Close()

			cs := NewContactService(newTestClient(t, srv), mail.Address("support@example.com"), testWebBaseURL)
			c := contact.NewContact(
				mail.Address("user@example.com"),
				mail.LanguageJA,
				contact.Message("<script>alert(1)</script>"),
			)

			if err := cs.Send(context.Background(), mail.Address("admin@example.com"), c); err != nil {
				t.Fatalf("Send(...) error = %v", err)
			}
			if strings.Contains(got.Html, "<script>") {
				t.Errorf("html body = %s, want the message to be escaped", got.Html)
			}
			if !strings.Contains(got.Text, "表示言語: 日本語") {
				t.Errorf("text body = %s, want the display language in Japanese", got.Text)
			}
		},
	)

	t.Run(
		"returns an error when Resend returns an error",
		func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`{"statusCode": 500, "name": "internal_server_error", "message": "test error"}`))
				},
			))
			defer srv.Close()

			cs := NewContactService(newTestClient(t, srv), mail.Address("support@example.com"), testWebBaseURL)
			c := contact.NewContact(
				mail.Address("user@example.com"),
				mail.LanguageJA,
				contact.Message("退会できません"),
			)

			if err := cs.Send(context.Background(), mail.Address("admin@example.com"), c); err == nil {
				t.Fatalf("Send(...) error = nil, want an error")
			}
		},
	)
}
