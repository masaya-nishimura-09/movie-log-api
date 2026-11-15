package user

import (
	"fmt"
	"html"
	"net/url"
	"strings"
	"time"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/mail"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
)

var jst = time.FixedZone("JST", 9*60*60)

type mailContent struct {
	subject     string
	heading     string
	paragraphs  []string
	buttonLabel string
	buttonURL   string
	notes       []string
	footer      string
}

func buildMessage(
	from mail.Address,
	to mail.Address,
	language mail.Language,
	logoURL string,
	content mailContent,
) mail.Message {
	return mail.NewMessage(
		from,
		to,
		mail.ReplyTo(""),
		mail.Subject(content.subject),
		mail.HTMLBody(buildHTMLBody(language, logoURL, content)),
		mail.TextBody(buildTextBody(content)),
	)
}

func buildHTMLBody(language mail.Language, logoURL string, content mailContent) string {
	var b strings.Builder

	fmt.Fprintf(&b, `<!doctype html><html lang="%s"><head><meta charset="utf-8"></head>
<body style="margin:0;padding:40px 16px;background:#faf4e9;font-family:-apple-system,'Hiragino Sans','Noto Sans JP',sans-serif;color:#0a2947">
<div style="max-width:444px;margin:0 auto">
<div style="margin:0 0 32px"><img src="%s" alt="Cinelier" height="40" style="height:40px;width:auto;border:0"></div>
<div style="font-size:18px;font-weight:700;margin:0 0 16px">%s</div>
<div style="font-size:14px;line-height:1.8">`,
		html.EscapeString(string(language)),
		html.EscapeString(logoURL),
		html.EscapeString(content.heading),
	)

	for _, p := range content.paragraphs {
		fmt.Fprintf(&b, `<p style="margin:0 0 16px">%s</p>`, html.EscapeString(p))
	}

	if content.buttonURL != "" {
		fmt.Fprintf(
			&b,
			`<p style="margin:24px 0"><a href="%s" style="display:inline-block;background:#8b5e3c;color:#faf4e9;text-decoration:none;font-weight:700;padding:12px 24px;border-radius:8px">%s</a></p>`,
			html.EscapeString(content.buttonURL),
			html.EscapeString(content.buttonLabel),
		)
	}

	for _, n := range content.notes {
		fmt.Fprintf(&b, `<p style="margin:0 0 8px;font-size:13px;color:#566674">%s</p>`, html.EscapeString(n))
	}

	fmt.Fprintf(&b, `</div>
</div>
<div style="max-width:444px;margin:32px auto 0;padding:16px 0 0;border-top:1px solid #d3d4c0;font-size:11px;color:#566674;line-height:1.6">%s</div>
</body></html>`,
		html.EscapeString(content.footer),
	)

	return b.String()
}

func buildTextBody(content mailContent) string {
	lines := []string{content.heading, ""}
	lines = append(lines, content.paragraphs...)
	if content.buttonURL != "" {
		lines = append(lines, "", content.buttonLabel, content.buttonURL)
	}
	if len(content.notes) > 0 {
		lines = append(lines, "")
		lines = append(lines, content.notes...)
	}
	lines = append(lines, "", content.footer)
	return strings.Join(lines, "\n")
}

func verificationURL(
	webBaseURL *url.URL,
	language mail.Language,
	value user.EmailVerificationTokenValue,
) string {
	u := webBaseURL.JoinPath(string(language), "verify-email")
	u.RawQuery = url.Values{"token": {string(value)}}.Encode()
	return u.String()
}

func loginURL(webBaseURL *url.URL, language mail.Language) string {
	return webBaseURL.JoinPath(string(language), "login").String()
}

func logoURL(webBaseURL *url.URL) string {
	return webBaseURL.JoinPath("mail", "logo.png").String()
}

func verificationMailContent(language mail.Language, verificationURL string) mailContent {
	if language == mail.LanguageEN {
		return mailContent{
			subject: "[Cinelier] Verify your email address",
			heading: "Verify your email address",
			paragraphs: []string{
				"Thank you for signing up for Cinelier.",
				"Click the button below to verify your email address.",
			},
			buttonLabel: "Verify email address",
			buttonURL:   verificationURL,
			notes: []string{
				"This link expires in 24 hours.",
				"If you did not sign up, you can ignore this email.",
			},
			footer: "This email was sent automatically by Cinelier.",
		}
	}

	return mailContent{
		subject: "【Cinelier】メールアドレスの確認",
		heading: "メールアドレスの確認",
		paragraphs: []string{
			"Cinelier にご登録いただき、ありがとうございます。",
			"下のボタンを押して、メールアドレスの確認を済ませてください。",
		},
		buttonLabel: "メールアドレスを確認する",
		buttonURL:   verificationURL,
		notes: []string{
			"このリンクの有効期限は24時間です。",
			"心当たりがない場合は、このメールを破棄してください。",
		},
		footer: "このメールは、Cinelier から自動で送信されました。",
	}
}

func alreadyRegisteredMailContent(language mail.Language, loginURL string) mailContent {
	if language == mail.LanguageEN {
		return mailContent{
			subject: "[Cinelier] This email address is already registered",
			heading: "This email address is already registered",
			paragraphs: []string{
				"Someone tried to sign up for Cinelier with this email address.",
				"This address is already registered. Please log in from the login page.",
			},
			buttonLabel: "Open the login page",
			buttonURL:   loginURL,
			notes: []string{
				"If this was not you, you can ignore this email. No changes have been made to your account.",
			},
			footer: "This email was sent automatically by Cinelier.",
		}
	}

	return mailContent{
		subject: "【Cinelier】このメールアドレスは登録済みです",
		heading: "このメールアドレスは登録済みです",
		paragraphs: []string{
			"このメールアドレスで、Cinelier への新規登録が試みられました。",
			"このアドレスはすでに登録されています。ログイン画面からログインしてください。",
		},
		buttonLabel: "ログイン画面を開く",
		buttonURL:   loginURL,
		notes: []string{
			"心当たりがない場合は、このメールを破棄してください。アカウントには何も変更されていません。",
		},
		footer: "このメールは、Cinelier から自動で送信されました。",
	}
}

func userVerifiedMailContent(verifiedAt time.Time) mailContent {
	return mailContent{
		subject: "[Cinelier] 新しいユーザーの確認",
		heading: "新しいユーザーの確認",
		paragraphs: []string{
			"新しいユーザーのメールアドレスの確認が済みました。",
			fmt.Sprintf("確認日時: %s（日本時間）", verifiedAt.In(jst).Format("2006年1月2日 15:04")),
		},
		footer: "このメールは、Cinelier から自動で送信されました。",
	}
}
