package mail

import (
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type Service struct {
	client *sendgrid.Client
}

func NewMailService(client *sendgrid.Client) *Service {
	return &Service{client: client}
}

func (s *Service) SendPinEmail(recipient, pin string) error {
	message := mail.NewV3MailInit(
		mail.NewEmail("Voicely", "no-reply@spksy.app"),
		"Voicely Login Pin",
		mail.NewEmail("", recipient),
		mail.NewContent("text/plain", "Your login pin is: "+pin),
	)

	_, err := s.client.Send(message)
	return err
}
