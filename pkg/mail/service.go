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
	m := mail.NewV3Mail()
	m.SetFrom(mail.NewEmail("Soapbox", "no-reply@soapbox.social"))
	m.SetTemplateID("d-94ee80b7ff33499894de719c02f095cf")

	p := mail.NewPersonalization()
	p.AddTos(mail.NewEmail("", recipient))
	p.SetDynamicTemplateData("pin", pin)

	m.AddPersonalizations(p)

	_, err := s.client.Send(m)
	return err
}
