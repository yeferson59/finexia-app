package mail

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"html/template"

	"github.com/resend/resend-go/v3"
)

//go:embed templates/*.html
var templates embed.FS

type Service struct {
	client *resend.Client
	from   string
	tmpl   *template.Template
}

type WaitlistData struct {
	Email string
}

func New(apiKey, from string) (*Service, error) {
	tmpl, err := template.ParseFS(templates, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("mail: parse templates: %w", err)
	}

	return &Service{
		client: resend.NewClient(apiKey),
		from:   from,
		tmpl:   tmpl,
	}, nil
}

func (s *Service) SendWaitlistConfirmation(email string) error {
	var body bytes.Buffer
	if err := s.tmpl.ExecuteTemplate(&body, "waitlist_confirmation.html", WaitlistData{Email: email}); err != nil {
		return errors.New("mail: render waitlist template: " + err.Error())
	}

	params := &resend.SendEmailRequest{
		From:    s.from,
		To:      []string{email},
		Subject: "¡Tu lugar está reservado — Finexia acceso anticipado",
		Html:    body.String(),
	}

	if _, err := s.client.Emails.Send(params); err != nil {
		return fmt.Errorf("mail: send waitlist confirmation to %s: %w", email, err)
	}

	return nil
}
