package service

import (
	"fmt"
	"gitlab.com/peleng-meteo/meteo-go/internal/config"
	emailProvider "gitlab.com/peleng-meteo/meteo-go/pkg/email"
)

const (
	nameField            = "name"
	verificationLinkTmpl = "%s/verification?code=%s"
)

type EmailService struct {
	provider    emailProvider.Provider
	sender      emailProvider.Sender
	config      config.EmailConfig
	frontendUrl string
}

// Structures used for templates
type verificationEmailInput struct {
	VerificationLink string
}

func NewEmailsService(provider emailProvider.Provider, sender emailProvider.Sender, config config.EmailConfig, frontendUrl string) *EmailService {
	return &EmailService{provider: provider, sender: sender, config: config, frontendUrl: frontendUrl}
}

func (s *EmailService) AddToList(name, email string) error {
	return s.provider.AddEmailToList(emailProvider.AddEmailInput{
		Email:  email,
		ListID: s.config.SendPulse.ListID,
		Variables: map[string]string{
			nameField: name,
		},
	})
}

func (s *EmailService) SendVerificationEmail(input SendVerificationEmailInput) error {
	subject := fmt.Sprintf(s.config.Subjects.Verification, input.Name)

	templateInput := verificationEmailInput{s.createVerificationLink(input.VerificationCode)}
	sendInput := emailProvider.SendEmailInput{Subject: subject, To: input.Email}
	if err := sendInput.GenerateBodyFromHTML(s.config.Templates.Verification, templateInput); err != nil {
		return err
	}

	return s.sender.Send(sendInput)
}

func (s *EmailService) createVerificationLink(code string) string {
	return fmt.Sprintf(verificationLinkTmpl, s.frontendUrl, code)
}
