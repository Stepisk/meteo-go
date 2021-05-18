package email

import (
	"bytes"
	"errors"
	"gitlab.com/peleng-meteo/meteo-go/pkg/logger"
	"text/template"
)

type SendEmailInput struct {
	To      string
	Subject string
	Body    string
}

type Sender interface {
	Send(input SendEmailInput) error
}

func (e *SendEmailInput) GenerateBodyFromHTML(templateFileName string, data interface{}) error {
	t, err := template.ParseFiles(templateFileName)
	if err != nil {
		logger.Errorf("failed to parse file %s:%s", templateFileName, err.Error())
		return err
	}

	buf := new(bytes.Buffer)
	if err = t.Execute(buf, data); err != nil {
		return err
	}

	e.Body = buf.String()

	return nil
}

func (e *SendEmailInput) Validate() error {
	if e.To == "" {
		return errors.New("empty to")
	}

	if e.Subject == "" {
		return errors.New("empty subject")
	}

	if e.Body == "" {
		return errors.New("empty body")
	}

	if !IsEmailValid(e.To) {
		return errors.New("invalid to email")
	}

	return nil
}
