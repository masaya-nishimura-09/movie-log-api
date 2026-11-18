package mail

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/mail"
	"github.com/resend/resend-go/v4"
)

var testWebBaseURL = &url.URL{Scheme: "https", Host: "web.example.com"}

func newTestClient(t *testing.T, srv *httptest.Server) *resend.Client {
	t.Helper()

	baseURL, err := url.Parse(srv.URL + "/")
	if err != nil {
		t.Fatalf("url.Parse(%q) error = %v", srv.URL, err)
	}
	client := resend.NewClient("test-api-key")
	client.BaseURL = baseURL
	return client
}

type sendRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	ReplyTo string   `json:"reply_to"`
	Subject string   `json:"subject"`
	Html    string   `json:"html"`
	Text    string   `json:"text"`
}

func newTestServer(t *testing.T, got *sendRequest) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if err := json.NewDecoder(r.Body).Decode(got); err != nil {
				t.Errorf("decode request body: %v", err)
			}
			w.Write([]byte(`{"id": "test-id"}`))
		},
	))
}

func assertSent(t *testing.T, call string, got sendRequest, to mail.Address, wantSubject string, wantText string) {
	t.Helper()

	if len(got.To) != 1 || got.To[0] != string(to) {
		t.Errorf("%s to = %v, want %v", call, got.To, to)
	}
	if got.Subject != wantSubject {
		t.Errorf("%s subject = %q, want %q", call, got.Subject, wantSubject)
	}
	if !strings.Contains(got.Text, wantText) {
		t.Errorf("%s text body = %s, want to contain %s", call, got.Text, wantText)
	}
	if got.Html == "" {
		t.Errorf("%s html body is empty", call)
	}
}

func TestSendEmailVerification(t *testing.T) {
	t.Run(
		"sends the verification mail from the support address with the logo and the link",
		func(t *testing.T) {
			var gotMethod, gotPath, gotAuth string
			var got sendRequest
			srv := httptest.NewServer(http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					gotMethod = r.Method
					gotPath = r.URL.Path
					gotAuth = r.Header.Get("Authorization")
					if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
						t.Errorf("decode request body: %v", err)
					}
					w.Write([]byte(`{"id": "test-id"}`))
				},
			))
			defer srv.Close()

			s := NewService(newTestClient(t, srv), mail.Address("support@example.com"), testWebBaseURL)
			ctx := context.Background()
			to := mail.Address("user@example.com")
			verificationURL := mail.URL("https://web.example.com/ja/verify-email?token=abc")

			if err := s.SendEmailVerification(ctx, to, mail.LanguageJA, verificationURL); err != nil {
				t.Fatalf("SendEmailVerification(ctx, %v, %v, %v) error = %v", to, mail.LanguageJA, verificationURL, err)
			}

			if gotMethod != http.MethodPost || gotPath != "/emails" || gotAuth != "Bearer test-api-key" {
				t.Errorf(
					"request = %s %s with Authorization %q, want POST /emails with %q",
					gotMethod, gotPath, gotAuth, "Bearer test-api-key",
				)
			}
			if got.From != "support@example.com" ||
				len(got.To) != 1 || got.To[0] != string(to) ||
				got.ReplyTo != "" ||
				got.Subject != "【Cinelier】メールアドレスの確認" {
				t.Errorf("request body = %+v, want from support@example.com to %v with the Japanese subject", got, to)
			}
			for _, want := range []string{
				`lang="ja"`,
				`src="https://web.example.com/mail/logo.png"`,
				`href="https://web.example.com/ja/verify-email?token=abc"`,
				"このメールは、Cinelier から自動で送信されました。",
			} {
				if !strings.Contains(got.Html, want) {
					t.Errorf("html body = %s, want to contain %s", got.Html, want)
				}
			}
			for _, want := range []string{
				string(verificationURL),
				"このメールは、Cinelier から自動で送信されました。",
			} {
				if !strings.Contains(got.Text, want) {
					t.Errorf("text body = %s, want to contain %s", got.Text, want)
				}
			}
		},
	)

	t.Run(
		"sends the verification mail in English",
		func(t *testing.T) {
			var got sendRequest
			srv := newTestServer(t, &got)
			defer srv.Close()

			s := NewService(newTestClient(t, srv), mail.Address("support@example.com"), testWebBaseURL)
			to := mail.Address("user@example.com")
			link := mail.URL("https://web.example.com/en/verify-email?token=abc")

			if err := s.SendEmailVerification(context.Background(), to, mail.LanguageEN, link); err != nil {
				t.Fatalf("SendEmailVerification(...) error = %v", err)
			}
			assertSent(t, "SendEmailVerification(...)", got, to, "[Cinelier] Verify your email address", string(link))
		},
	)

	t.Run(
		"escapes the link in the html body",
		func(t *testing.T) {
			var got sendRequest
			srv := newTestServer(t, &got)
			defer srv.Close()

			s := NewService(newTestClient(t, srv), mail.Address("support@example.com"), testWebBaseURL)
			ctx := context.Background()
			verificationURL := mail.URL("https://web.example.com/ja/verify-email?token=abc&x=<y>")

			if err := s.SendEmailVerification(ctx, mail.Address("user@example.com"), mail.LanguageJA, verificationURL); err != nil {
				t.Fatalf("SendEmailVerification(...) error = %v", err)
			}
			if strings.Contains(got.Html, "<y>") {
				t.Errorf("html body = %s, want the link to be escaped", got.Html)
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

			s := NewService(newTestClient(t, srv), mail.Address("support@example.com"), testWebBaseURL)
			ctx := context.Background()

			if err := s.SendEmailVerification(
				ctx, mail.Address("user@example.com"), mail.LanguageJA, mail.URL("https://web.example.com/"),
			); err == nil {
				t.Fatalf("SendEmailVerification(...) error = nil, want an error")
			}
		},
	)
}

func TestSendAlreadyRegistered(t *testing.T) {
	to := mail.Address("user@example.com")
	link := mail.URL("https://web.example.com/ja/login")

	t.Run(
		"sends the already registered mail in Japanese",
		func(t *testing.T) {
			var got sendRequest
			srv := newTestServer(t, &got)
			defer srv.Close()

			s := NewService(newTestClient(t, srv), mail.Address("support@example.com"), testWebBaseURL)

			if err := s.SendAlreadyRegistered(context.Background(), to, mail.LanguageJA, link); err != nil {
				t.Fatalf("SendAlreadyRegistered(...) error = %v", err)
			}
			assertSent(t, "SendAlreadyRegistered(...)", got, to, "【Cinelier】このメールアドレスは登録済みです", "新規登録が試みられました")
		},
	)

	t.Run(
		"sends the already registered mail in English",
		func(t *testing.T) {
			var got sendRequest
			srv := newTestServer(t, &got)
			defer srv.Close()

			s := NewService(newTestClient(t, srv), mail.Address("support@example.com"), testWebBaseURL)

			if err := s.SendAlreadyRegistered(context.Background(), to, mail.LanguageEN, link); err != nil {
				t.Fatalf("SendAlreadyRegistered(...) error = %v", err)
			}
			assertSent(t, "SendAlreadyRegistered(...)", got, to, "[Cinelier] This email address is already registered", "Someone tried to sign up")
		},
	)
}

func TestSendUserVerified(t *testing.T) {
	t.Run(
		"sends the user verified notice with the verified time in JST",
		func(t *testing.T) {
			var got sendRequest
			srv := newTestServer(t, &got)
			defer srv.Close()

			s := NewService(newTestClient(t, srv), mail.Address("support@example.com"), testWebBaseURL)
			to := mail.Address("admin@example.com")
			verifiedAt := time.Date(2026, 11, 16, 3, 4, 0, 0, time.UTC)

			if err := s.SendUserVerified(context.Background(), to, verifiedAt); err != nil {
				t.Fatalf("SendUserVerified(...) error = %v", err)
			}
			assertSent(t, "SendUserVerified(...)", got, to, "[Cinelier] 新しいユーザーの確認", "確認日時: 2026年11月16日 12:04（日本時間）")
		},
	)
}

func TestSendEmailChangeVerification(t *testing.T) {
	to := mail.Address("new@example.com")
	link := mail.URL("https://web.example.com/ja/verify-email-change?token=abc")

	t.Run(
		"sends the email change verification mail in Japanese",
		func(t *testing.T) {
			var got sendRequest
			srv := newTestServer(t, &got)
			defer srv.Close()

			s := NewService(newTestClient(t, srv), mail.Address("support@example.com"), testWebBaseURL)

			if err := s.SendEmailChangeVerification(context.Background(), to, mail.LanguageJA, link); err != nil {
				t.Fatalf("SendEmailChangeVerification(...) error = %v", err)
			}
			assertSent(t, "SendEmailChangeVerification(...)", got, to, "【Cinelier】新しいメールアドレスの確認", string(link))
		},
	)

	t.Run(
		"sends the email change verification mail in English",
		func(t *testing.T) {
			var got sendRequest
			srv := newTestServer(t, &got)
			defer srv.Close()

			s := NewService(newTestClient(t, srv), mail.Address("support@example.com"), testWebBaseURL)

			if err := s.SendEmailChangeVerification(context.Background(), to, mail.LanguageEN, link); err != nil {
				t.Fatalf("SendEmailChangeVerification(...) error = %v", err)
			}
			assertSent(t, "SendEmailChangeVerification(...)", got, to, "[Cinelier] Verify your new email address", string(link))
		},
	)
}

func TestSendEmailChangeAlreadyRegistered(t *testing.T) {
	to := mail.Address("new@example.com")
	link := mail.URL("https://web.example.com/ja/login")

	t.Run(
		"sends the email change already registered mail in Japanese",
		func(t *testing.T) {
			var got sendRequest
			srv := newTestServer(t, &got)
			defer srv.Close()

			s := NewService(newTestClient(t, srv), mail.Address("support@example.com"), testWebBaseURL)

			if err := s.SendEmailChangeAlreadyRegistered(context.Background(), to, mail.LanguageJA, link); err != nil {
				t.Fatalf("SendEmailChangeAlreadyRegistered(...) error = %v", err)
			}
			assertSent(t, "SendEmailChangeAlreadyRegistered(...)", got, to, "【Cinelier】このメールアドレスは登録済みです", "変更する申し込みがありました")
		},
	)

	t.Run(
		"sends the email change already registered mail in English",
		func(t *testing.T) {
			var got sendRequest
			srv := newTestServer(t, &got)
			defer srv.Close()

			s := NewService(newTestClient(t, srv), mail.Address("support@example.com"), testWebBaseURL)

			if err := s.SendEmailChangeAlreadyRegistered(context.Background(), to, mail.LanguageEN, link); err != nil {
				t.Fatalf("SendEmailChangeAlreadyRegistered(...) error = %v", err)
			}
			assertSent(t, "SendEmailChangeAlreadyRegistered(...)", got, to, "[Cinelier] This email address is already registered", "Someone tried to change the email address")
		},
	)
}

func TestSendEmailChanged(t *testing.T) {
	to := mail.Address("old@example.com")
	newEmail := mail.Address("new@example.com")
	link := mail.URL("https://web.example.com/ja/contact")

	t.Run(
		"sends the email changed notice with the new email in Japanese",
		func(t *testing.T) {
			var got sendRequest
			srv := newTestServer(t, &got)
			defer srv.Close()

			s := NewService(newTestClient(t, srv), mail.Address("support@example.com"), testWebBaseURL)

			if err := s.SendEmailChanged(context.Background(), to, mail.LanguageJA, newEmail, link); err != nil {
				t.Fatalf("SendEmailChanged(...) error = %v", err)
			}
			assertSent(t, "SendEmailChanged(...)", got, to, "【Cinelier】メールアドレスが変更されました", "新しいメールアドレス: new@example.com")
		},
	)

	t.Run(
		"sends the email changed notice with the new email in English",
		func(t *testing.T) {
			var got sendRequest
			srv := newTestServer(t, &got)
			defer srv.Close()

			s := NewService(newTestClient(t, srv), mail.Address("support@example.com"), testWebBaseURL)

			if err := s.SendEmailChanged(context.Background(), to, mail.LanguageEN, newEmail, link); err != nil {
				t.Fatalf("SendEmailChanged(...) error = %v", err)
			}
			assertSent(t, "SendEmailChanged(...)", got, to, "[Cinelier] Your email address has been changed", "New email address: new@example.com")
		},
	)
}

func TestSendPasswordReset(t *testing.T) {
	to := mail.Address("user@example.com")
	link := mail.URL("https://web.example.com/ja/reset-password?token=abc")

	t.Run(
		"sends the password reset mail in Japanese",
		func(t *testing.T) {
			var got sendRequest
			srv := newTestServer(t, &got)
			defer srv.Close()

			s := NewService(newTestClient(t, srv), mail.Address("support@example.com"), testWebBaseURL)

			if err := s.SendPasswordReset(context.Background(), to, mail.LanguageJA, link); err != nil {
				t.Fatalf("SendPasswordReset(...) error = %v", err)
			}
			assertSent(t, "SendPasswordReset(...)", got, to, "【Cinelier】パスワードの再設定", string(link))
		},
	)

	t.Run(
		"sends the password reset mail in English",
		func(t *testing.T) {
			var got sendRequest
			srv := newTestServer(t, &got)
			defer srv.Close()

			s := NewService(newTestClient(t, srv), mail.Address("support@example.com"), testWebBaseURL)

			if err := s.SendPasswordReset(context.Background(), to, mail.LanguageEN, link); err != nil {
				t.Fatalf("SendPasswordReset(...) error = %v", err)
			}
			assertSent(t, "SendPasswordReset(...)", got, to, "[Cinelier] Reset your password", string(link))
		},
	)
}

func TestSendPasswordChanged(t *testing.T) {
	to := mail.Address("user@example.com")
	link := mail.URL("https://web.example.com/ja/contact")

	t.Run(
		"sends the password changed notice in Japanese",
		func(t *testing.T) {
			var got sendRequest
			srv := newTestServer(t, &got)
			defer srv.Close()

			s := NewService(newTestClient(t, srv), mail.Address("support@example.com"), testWebBaseURL)

			if err := s.SendPasswordChanged(context.Background(), to, mail.LanguageJA, link); err != nil {
				t.Fatalf("SendPasswordChanged(...) error = %v", err)
			}
			assertSent(t, "SendPasswordChanged(...)", got, to, "【Cinelier】パスワードが変更されました", string(link))
		},
	)

	t.Run(
		"sends the password changed notice in English",
		func(t *testing.T) {
			var got sendRequest
			srv := newTestServer(t, &got)
			defer srv.Close()

			s := NewService(newTestClient(t, srv), mail.Address("support@example.com"), testWebBaseURL)

			if err := s.SendPasswordChanged(context.Background(), to, mail.LanguageEN, link); err != nil {
				t.Fatalf("SendPasswordChanged(...) error = %v", err)
			}
			assertSent(t, "SendPasswordChanged(...)", got, to, "[Cinelier] Your password has been changed", string(link))
		},
	)
}

func TestSendWithdrawn(t *testing.T) {
	to := mail.Address("user@example.com")

	t.Run(
		"sends the withdrawn notice in Japanese",
		func(t *testing.T) {
			var got sendRequest
			srv := newTestServer(t, &got)
			defer srv.Close()

			s := NewService(newTestClient(t, srv), mail.Address("support@example.com"), testWebBaseURL)

			if err := s.SendWithdrawn(context.Background(), to, mail.LanguageJA); err != nil {
				t.Fatalf("SendWithdrawn(...) error = %v", err)
			}
			assertSent(t, "SendWithdrawn(...)", got, to, "【Cinelier】退会が完了しました", "これまでご利用いただき、ありがとうございました。")
		},
	)

	t.Run(
		"sends the withdrawn notice in English",
		func(t *testing.T) {
			var got sendRequest
			srv := newTestServer(t, &got)
			defer srv.Close()

			s := NewService(newTestClient(t, srv), mail.Address("support@example.com"), testWebBaseURL)

			if err := s.SendWithdrawn(context.Background(), to, mail.LanguageEN); err != nil {
				t.Fatalf("SendWithdrawn(...) error = %v", err)
			}
			assertSent(t, "SendWithdrawn(...)", got, to, "[Cinelier] Your account has been deleted", "Thank you for using Cinelier.")
		},
	)
}
