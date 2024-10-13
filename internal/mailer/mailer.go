package mailer

import (
	"embed"
	"fmt"
	"strings"
)

const (
	maxRetries = 3

	UserWelcomeTemplate = "user_invitation.tmpl"
)

//go:embed templates
var FS embed.FS

type Client interface {
	Send(templateFile string, email EmailData, data any, isSandbox bool) (responseCode string, err error)
}

type EmailData struct {
	Name  string
	Email string
}

func (e *EmailData) Format() string {
	return fmt.Sprintf("%s <%s>", strings.TrimSpace(e.Name), strings.TrimSpace(strings.ToLower(e.Email)))
}
