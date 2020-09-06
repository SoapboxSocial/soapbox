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
		mail.NewEmail("Soapbox", "no-reply@soapbox.social"),
		"Your Soapbox Login Pin",
		mail.NewEmail("", recipient),
		mail.NewContent("text/plain", "Hey, \n Here is your login pin: " + pin + "\n Have fun!"),
	)

	_, err := s.client.Send(message)
	return err
}
