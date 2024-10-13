package mailer

import (
	"fmt"
	"strings"
)

type SendError struct {
	Message string
	Errors  []error
}

func NewSendError(errors []error, message string, args ...any) error {
	return &SendError{
		Message: fmt.Sprintf(message, args...),
		Errors:  errors,
	}
}

func (e *SendError) Error() string {
	msg := strings.TrimSpace(e.Message)

	if msg != "" && len(e.Errors) > 0 {
		msg += ": "
	}
	for i, err := range e.Errors {
		if i > 0 {
			msg += ", "
		}
		msg += fmt.Sprintf("Error #%d: [%v]", i+1, err)
	}
	return msg
}

func (e *SendError) AddError(err error) {
	if e.Errors == nil {
		e.Errors = make([]error, 0)
	}
	e.Errors = append(e.Errors, err)
}
