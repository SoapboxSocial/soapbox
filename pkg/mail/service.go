package mail

import (
	"fmt"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type Service struct {
	client sendgrid.Client
}

func NewMailService(client sendgrid.Client) *Service {
	return &Service{client: client}
}

func (s *Service) SendPinEmail(recipient, pin string) error {
	from := mail.NewEmail("Voicely", "no-reply@spksy.app")
	subject := "Voicely Pin"
	to := mail.NewEmail("", recipient)
	plainTextContent := "Your login pin is: " + pin
	htmlContent := fmt.Sprintf("Your login pin is: <strong>%s</strong>", pin)
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	_, err := s.client.Send(message)
	if err != nil {
		return err
	}

	return nil
}