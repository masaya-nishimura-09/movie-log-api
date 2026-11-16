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

func emailChangeVerificationURL(
	webBaseURL *url.URL,
	language mail.Language,
	value user.EmailChangeTokenValue,
) string {
	u := webBaseURL.JoinPath(string(language), "verify-email-change")
	u.RawQuery = url.Values{"token": {string(value)}}.Encode()
	return u.String()
}

func passwordResetURL(
	webBaseURL *url.URL,
	language mail.Language,
	value user.PasswordResetTokenValue,
) string {
	u := webBaseURL.JoinPath(string(language), "reset-password")
	u.RawQuery = url.Values{"token": {string(value)}}.Encode()
	return u.String()
}

func contactURL(webBaseURL *url.URL, language mail.Language) string {
	return webBaseURL.JoinPath(string(language), "contact").String()
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

func emailChangeVerificationMailContent(language mail.Language, verificationURL string) mailContent {
	if language == mail.LanguageEN {
		return mailContent{
			subject: "[Cinelier] Verify your new email address",
			heading: "Verify your new email address",
			paragraphs: []string{
				"We received a request to change the email address of your Cinelier account to this address.",
				"Click the button below to complete the change.",
			},
			buttonLabel: "Change email address",
			buttonURL:   verificationURL,
			notes: []string{
				"This link expires in 24 hours. Until then, your current email address stays the same.",
				"If you did not request this, you can ignore this email.",
			},
			footer: "This email was sent automatically by Cinelier.",
		}
	}

	return mailContent{
		subject: "【Cinelier】新しいメールアドレスの確認",
		heading: "新しいメールアドレスの確認",
		paragraphs: []string{
			"Cinelier のアカウントのメールアドレスを、このアドレスに変更する申し込みがありました。",
			"下のボタンを押すと、変更が完了します。",
		},
		buttonLabel: "メールアドレスを変更する",
		buttonURL:   verificationURL,
		notes: []string{
			"このリンクの有効期限は24時間です。変更が完了するまでは、今のメールアドレスのままです。",
			"心当たりがない場合は、このメールを破棄してください。",
		},
		footer: "このメールは、Cinelier から自動で送信されました。",
	}
}

func emailChangeAlreadyRegisteredMailContent(language mail.Language, loginURL string) mailContent {
	if language == mail.LanguageEN {
		return mailContent{
			subject: "[Cinelier] This email address is already registered",
			heading: "This email address is already registered",
			paragraphs: []string{
				"Someone tried to change the email address of a Cinelier account to this address.",
				"This address is already registered, so the change was not made.",
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
			"Cinelier のアカウントのメールアドレスを、このアドレスに変更する申し込みがありました。",
			"このアドレスはすでに登録されているため、変更は行われていません。",
		},
		buttonLabel: "ログイン画面を開く",
		buttonURL:   loginURL,
		notes: []string{
			"心当たりがない場合は、このメールを破棄してください。アカウントには何も変更されていません。",
		},
		footer: "このメールは、Cinelier から自動で送信されました。",
	}
}

func emailChangedMailContent(language mail.Language, newEmail user.Email, contactURL string) mailContent {
	if language == mail.LanguageEN {
		return mailContent{
			subject: "[Cinelier] Your email address has been changed",
			heading: "Your email address has been changed",
			paragraphs: []string{
				"The email address of your Cinelier account has been changed.",
				fmt.Sprintf("New email address: %s", newEmail),
				"From now on, please log in with the new email address.",
			},
			buttonLabel: "Contact us",
			buttonURL:   contactURL,
			notes: []string{
				"If you did not make this change, please contact us from the button above.",
			},
			footer: "This email was sent automatically by Cinelier.",
		}
	}

	return mailContent{
		subject: "【Cinelier】メールアドレスが変更されました",
		heading: "メールアドレスが変更されました",
		paragraphs: []string{
			"Cinelier のアカウントのメールアドレスが変更されました。",
			fmt.Sprintf("新しいメールアドレス: %s", newEmail),
			"今後は、新しいメールアドレスでログインしてください。",
		},
		buttonLabel: "お問い合わせ",
		buttonURL:   contactURL,
		notes: []string{
			"心当たりがない場合は、上のボタンからお問い合わせください。",
		},
		footer: "このメールは、Cinelier から自動で送信されました。",
	}
}

func passwordResetMailContent(language mail.Language, resetURL string) mailContent {
	if language == mail.LanguageEN {
		return mailContent{
			subject: "[Cinelier] Reset your password",
			heading: "Reset your password",
			paragraphs: []string{
				"We received a request to reset the password of your Cinelier account.",
				"Click the button below to set a new password.",
			},
			buttonLabel: "Reset password",
			buttonURL:   resetURL,
			notes: []string{
				"This link expires in 1 hour.",
				"If you did not request this, you can ignore this email. Your password will not be changed.",
			},
			footer: "This email was sent automatically by Cinelier.",
		}
	}

	return mailContent{
		subject: "【Cinelier】パスワードの再設定",
		heading: "パスワードの再設定",
		paragraphs: []string{
			"Cinelier のアカウントのパスワードを再設定する申し込みがありました。",
			"下のボタンを押して、新しいパスワードを設定してください。",
		},
		buttonLabel: "パスワードを再設定する",
		buttonURL:   resetURL,
		notes: []string{
			"このリンクの有効期限は1時間です。",
			"心当たりがない場合は、このメールを破棄してください。パスワードは変更されません。",
		},
		footer: "このメールは、Cinelier から自動で送信されました。",
	}
}

func passwordChangedMailContent(language mail.Language, contactURL string) mailContent {
	if language == mail.LanguageEN {
		return mailContent{
			subject: "[Cinelier] Your password has been changed",
			heading: "Your password has been changed",
			paragraphs: []string{
				"The password of your Cinelier account has been reset.",
				"You have been logged out on all devices. Please log in again with the new password.",
			},
			buttonLabel: "Contact us",
			buttonURL:   contactURL,
			notes: []string{
				"If you did not make this change, please contact us from the button above.",
			},
			footer: "This email was sent automatically by Cinelier.",
		}
	}

	return mailContent{
		subject: "【Cinelier】パスワードが変更されました",
		heading: "パスワードが変更されました",
		paragraphs: []string{
			"Cinelier のアカウントのパスワードが再設定されました。",
			"すべての端末からログアウトしています。新しいパスワードでログインし直してください。",
		},
		buttonLabel: "お問い合わせ",
		buttonURL:   contactURL,
		notes: []string{
			"心当たりがない場合は、上のボタンからお問い合わせください。",
		},
		footer: "このメールは、Cinelier から自動で送信されました。",
	}
}

func withdrawnMailContent(language mail.Language) mailContent {
	if language == mail.LanguageEN {
		return mailContent{
			subject: "[Cinelier] Your account has been deleted",
			heading: "Your account has been deleted",
			paragraphs: []string{
				"Your Cinelier account has been deleted.",
				"All of your records and images have been deleted as well.",
				"Thank you for using Cinelier.",
			},
			footer: "This email was sent automatically by Cinelier.",
		}
	}

	return mailContent{
		subject: "【Cinelier】退会が完了しました",
		heading: "退会が完了しました",
		paragraphs: []string{
			"Cinelier のアカウントを削除しました。",
			"記録や画像も、すべて削除しています。",
			"これまでご利用いただき、ありがとうございました。",
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
