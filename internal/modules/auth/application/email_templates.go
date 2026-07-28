package application

import (
	"bytes"
	"embed"
	"fmt"
	htemplate "html/template"
	"io"
	ttemplate "text/template"
)

// emailTemplatesFS embeds all .tmpl files under templates/ into the binary so
// the deployable artifact stays single-file while keeping markup out of Go.
//
//go:embed templates/*.tmpl
var emailTemplatesFS embed.FS

// passwordResetSubject is the email subject line for password reset emails.
const passwordResetSubject = "Reset your password"

// PasswordResetData is the data model passed to the password reset templates.
type PasswordResetData struct {
	ResetLink string
}

// Templates are parsed once at startup. A malformed template panics the app
// immediately (fail-fast) rather than failing at first email send.
var (
	passwordResetHTMLTmpl = htemplate.Must(
		htemplate.New("password_reset.html.tmpl").
			ParseFS(emailTemplatesFS, "templates/password_reset.html.tmpl"),
	)
	passwordResetPlainTmpl = ttemplate.Must(
		ttemplate.New("password_reset.txt.tmpl").
			ParseFS(emailTemplatesFS, "templates/password_reset.txt.tmpl"),
	)
)

// passwordResetHTML renders the HTML email body for a password reset.
func passwordResetHTML(data PasswordResetData) (string, error) {
	return renderTemplate(passwordResetHTMLTmpl, data)
}

// passwordResetPlain renders the plain-text email body for a password reset.
func passwordResetPlain(data PasswordResetData) (string, error) {
	return renderTemplate(passwordResetPlainTmpl, data)
}

// renderTemplate executes a parsed template into a string. Both html/template
// and text/template implement Execute(io.Writer, any) error, so this helper
// works for either via the small renderer interface.
func renderTemplate(t renderer, data any) (string, error) {
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("render email template: %w", err)
	}
	return buf.String(), nil
}

// renderer is the subset of *template.Template both text and html packages
// satisfy, allowing a single render helper.
type renderer interface {
	Execute(wr io.Writer, data any) error
}
