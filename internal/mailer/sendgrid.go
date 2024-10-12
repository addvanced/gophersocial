package mailer

import (
	"bytes"
	"fmt"
	"text/template"
	"time"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"go.uber.org/zap"
)

type sendgridMailer struct {
	client    *sendgrid.Client
	apiKey    string
	fromEmail string
	fromName  string

	logger *zap.SugaredLogger
}

func NewSendgrid(logger *zap.SugaredLogger, fromName, fromEmail, apiKey string) Client {
	client := sendgrid.NewSendClient(apiKey)

	return &sendgridMailer{
		client:    client,
		apiKey:    apiKey,
		fromEmail: fromEmail,
		fromName:  fromName,
		logger:    logger.Named("sendgrid"),
	}
}

func (s *sendgridMailer) Send(templateFile string, username, email string, data any, isSandbox bool) error {
	from := mail.NewEmail(s.fromName, s.fromEmail)
	to := mail.NewEmail(username, email)

	// template parsing and building
	tmpl, err := template.ParseFS(FS, fmt.Sprintf("templates/%s", templateFile))
	if err != nil {
		return err
	}

	subject := new(bytes.Buffer)
	if err = tmpl.ExecuteTemplate(subject, "subject", data); err != nil {
		return err
	}
	body := new(bytes.Buffer)
	if err = tmpl.ExecuteTemplate(body, "body", data); err != nil {
		return err
	}

	message := mail.NewSingleEmail(from, subject.String(), to, "", body.String())

	message.SetMailSettings(&mail.MailSettings{
		SandboxMode: &mail.Setting{
			Enable: &isSandbox,
		},
	})

	for i := 1; i <= maxRetries; i++ {
		response, err := s.client.Send(message)
		if err != nil {
			s.logger.Errorf("Failed to send email to '%s', attempt %d of %d. Received status code '%v'. Error: %v\n", email, i, maxRetries, response.StatusCode, err)
			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}
		s.logger.Infow("Successfully sent email.", "receiver_email", email, "status_code", response.StatusCode, "attempt", i)
		return nil
	}

	return fmt.Errorf("failed to send email after %d attempts", maxRetries)
}
