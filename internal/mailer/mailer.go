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
	Send(templateFile string, username, email string, data any, isSandbox bool) error
}

func formatEmail(username, email string) string {
	return fmt.Sprintf("%s <%s>", strings.TrimSpace(username), strings.TrimSpace(strings.ToLower(email)))
}
