package mailer

import (
	"bytes"
	"fmt"
	"text/template"
	"time"

	"github.com/resend/resend-go/v2"
	"go.uber.org/zap"
)

type resendMailer struct {
	client *resend.Client
	from   string

	logger *zap.SugaredLogger
}

func NewResend(logger *zap.SugaredLogger, fromName, fromEmail, apiKey string) Client {
	client := resend.NewClient(apiKey)

	return &resendMailer{
		client: client,
		from:   formatEmail(fromName, fromEmail),
		logger: logger.Named("resend"),
	}
}

func (s *resendMailer) Send(templateFile string, username, email string, data any, _ bool) error {

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

	message := &resend.SendEmailRequest{
		From:    s.from,
		To:      []string{formatEmail(username, email)},
		Subject: subject.String(),
		Html:    body.String(),
	}

	for i := 1; i <= maxRetries; i++ {
		response, err := s.client.Emails.Send(message)
		if err != nil {
			s.logger.Errorf("Failed to send email to '%s', attempt %d of %d. Error: %v\n", email, i, maxRetries, err)
			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}
		s.logger.Infow("Successfully sent email.", "receiver_email", email, "resend_message_id", response.Id, "attempt", i)
		return nil
	}

	return fmt.Errorf("failed to send email after %d attempts", maxRetries)
}
