package mail

import (
	"bytes"
	"embed"
	"fmt"
	htmltemplate "html/template"
	texttemplate "text/template"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/mail"
)

//go:embed templates
var templateFiles embed.FS

const (
	templateEmailVerification            = "email_verification"
	templateAlreadyRegistered            = "already_registered"
	templateUserVerified                 = "user_verified"
	templateEmailChangeVerification      = "email_change_verification"
	templateEmailChangeAlreadyRegistered = "email_change_already_registered"
	templateEmailChanged                 = "email_changed"
	templatePasswordReset                = "password_reset"
	templatePasswordChanged              = "password_changed"
	templateWithdrawn                    = "withdrawn"
)

type mailTemplate struct {
	html *htmltemplate.Template
	text *texttemplate.Template
}

type templateData struct {
	Language   mail.Language
	LogoURL    string
	URL        mail.URL
	NewEmail   mail.Address
	VerifiedAt string
}

type buttonData struct {
	Label string
	URL   mail.URL
}

var templateFuncs = htmltemplate.FuncMap{
	"button": func(label string, url mail.URL) buttonData {
		return buttonData{Label: label, URL: url}
	},
}

var templates = mustLoadTemplates()

func mustLoadTemplates() map[string]mailTemplate {
	userTemplates := []string{
		templateEmailVerification,
		templateAlreadyRegistered,
		templateEmailChangeVerification,
		templateEmailChangeAlreadyRegistered,
		templateEmailChanged,
		templatePasswordReset,
		templatePasswordChanged,
		templateWithdrawn,
	}

	loaded := map[string]mailTemplate{}
	for _, language := range []mail.Language{mail.LanguageJA, mail.LanguageEN} {
		for _, name := range userTemplates {
			loaded[templateKey(language, name)] = mustLoadTemplate(language, name)
		}
	}
	loaded[templateKey(mail.LanguageJA, templateUserVerified)] = mustLoadTemplate(mail.LanguageJA, templateUserVerified)

	return loaded
}

func mustLoadTemplate(language mail.Language, name string) mailTemplate {
	return mailTemplate{
		html: htmltemplate.Must(
			htmltemplate.New("").Funcs(templateFuncs).ParseFS(
				templateFiles,
				"templates/layout.html",
				fmt.Sprintf("templates/%s/%s.html", language, name),
			),
		),
		text: texttemplate.Must(
			texttemplate.New("").ParseFS(
				templateFiles,
				"templates/layout.txt",
				fmt.Sprintf("templates/%s/%s.txt", language, name),
			),
		),
	}
}

func templateKey(language mail.Language, name string) string {
	return fmt.Sprintf("%s/%s", language, name)
}

func render(name string, data templateData) (subject string, htmlBody string, textBody string, err error) {
	t, ok := templates[templateKey(data.Language, name)]
	if !ok {
		return "", "", "", fmt.Errorf("template %s for language %s is not found", name, data.Language)
	}

	var s, h, x bytes.Buffer
	if err := t.text.ExecuteTemplate(&s, "subject", data); err != nil {
		return "", "", "", fmt.Errorf("render subject: %w", err)
	}
	if err := t.html.ExecuteTemplate(&h, "layout", data); err != nil {
		return "", "", "", fmt.Errorf("render html body: %w", err)
	}
	if err := t.text.ExecuteTemplate(&x, "layout", data); err != nil {
		return "", "", "", fmt.Errorf("render text body: %w", err)
	}

	return s.String(), h.String(), x.String(), nil
}
