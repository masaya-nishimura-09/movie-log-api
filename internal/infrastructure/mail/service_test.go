package mail

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/mail"
	"github.com/resend/resend-go/v4"
)

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

func newTestMessage(replyTo mail.ReplyTo) mail.Message {
	return mail.NewMessage(
		mail.Address("support@example.com"),
		mail.Address("user@example.com"),
		replyTo,
		mail.Subject("Test Subject"),
		mail.HTMLBody("<p>Test Body</p>"),
		mail.TextBody("Test Body"),
	)
}

type sendRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	ReplyTo string   `json:"reply_to"`
	Subject string   `json:"subject"`
	Html    string   `json:"html"`
	Text    string   `json:"text"`
}

func TestSend(t *testing.T) {
	t.Run(
		"sends the message to Resend",
		func(t *testing.T) {
			var (
				gotMethod string
				gotPath   string
				gotAuth   string
				gotBody   sendRequest
			)
			srv := httptest.NewServer(http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					gotMethod = r.Method
					gotPath = r.URL.Path
					gotAuth = r.Header.Get("Authorization")
					if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
						t.Errorf("decode request body: %v", err)
					}
					w.Write([]byte(`{"id": "test-id"}`))
				},
			))
			defer srv.Close()

			s := NewService(newTestClient(t, srv))
			ctx := context.Background()
			m := newTestMessage(mail.ReplyTo("inquirer@example.com"))

			if err := s.Send(ctx, m); err != nil {
				t.Fatalf("Send(ctx, %v) error = %v", m, err)
			}

			if gotMethod != http.MethodPost {
				t.Errorf("method = %q, want %q", gotMethod, http.MethodPost)
			}
			if gotPath != "/emails" {
				t.Errorf("path = %q, want %q", gotPath, "/emails")
			}
			if gotAuth != "Bearer test-api-key" {
				t.Errorf("Authorization = %q, want %q", gotAuth, "Bearer test-api-key")
			}
			want := sendRequest{
				From:    "support@example.com",
				To:      []string{"user@example.com"},
				ReplyTo: "inquirer@example.com",
				Subject: "Test Subject",
				Html:    "<p>Test Body</p>",
				Text:    "Test Body",
			}
			if gotBody.From != want.From ||
				len(gotBody.To) != 1 || gotBody.To[0] != want.To[0] ||
				gotBody.ReplyTo != want.ReplyTo ||
				gotBody.Subject != want.Subject ||
				gotBody.Html != want.Html ||
				gotBody.Text != want.Text {
				t.Errorf("request body = %+v, want %+v", gotBody, want)
			}
		},
	)

	t.Run(
		"omits reply_to when ReplyTo is empty",
		func(t *testing.T) {
			var gotBody map[string]any
			srv := httptest.NewServer(http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
						t.Errorf("decode request body: %v", err)
					}
					w.Write([]byte(`{"id": "test-id"}`))
				},
			))
			defer srv.Close()

			s := NewService(newTestClient(t, srv))
			ctx := context.Background()
			m := newTestMessage(mail.ReplyTo(""))

			if err := s.Send(ctx, m); err != nil {
				t.Fatalf("Send(ctx, %v) error = %v", m, err)
			}

			if _, ok := gotBody["reply_to"]; ok {
				t.Errorf("request body has reply_to = %v, want no reply_to", gotBody["reply_to"])
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

			s := NewService(newTestClient(t, srv))
			ctx := context.Background()
			m := newTestMessage(mail.ReplyTo(""))

			if err := s.Send(ctx, m); err == nil {
				t.Fatalf("Send(ctx, %v) error = nil, want an error", m)
			}
		},
	)
}
