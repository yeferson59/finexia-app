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

type ActivityAlertData struct {
	UserName        string
	AssetTicker     string
	AssetName       string
	TransactionType string
	Quantity        string
	Price           string
	Total           string
	Currency        string
	TransactionDate string
	DashboardURL    string
}

type WeeklySummaryPortfolio struct {
	Name             string
	Type             string
	TotalMarketValue string
	TotalGainLoss    string
	TotalGainLossPct string
	GainLossColor    string
}

type WeeklySummaryData struct {
	UserName         string
	TotalValue       string
	TotalGainLoss    string
	TotalGainLossPct string
	GainLossColor    string
	Portfolios       []WeeklySummaryPortfolio
	DashboardURL     string
	WeekLabel        string
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

func (s *Service) SendActivityAlert(email string, data ActivityAlertData) error {
	var body bytes.Buffer
	if err := s.tmpl.ExecuteTemplate(&body, "activity_alert.html", data); err != nil {
		return fmt.Errorf("mail: render activity_alert template: %w", err)
	}

	subject := fmt.Sprintf("Nueva transacción: %s %s — Finexia", data.TransactionType, data.AssetTicker)
	params := &resend.SendEmailRequest{
		From:    s.from,
		To:      []string{email},
		Subject: subject,
		Html:    body.String(),
	}

	if _, err := s.client.Emails.Send(params); err != nil {
		return fmt.Errorf("mail: send activity alert to %s: %w", email, err)
	}

	return nil
}

func (s *Service) SendWeeklySummary(email string, data WeeklySummaryData) error {
	var body bytes.Buffer
	if err := s.tmpl.ExecuteTemplate(&body, "weekly_summary.html", data); err != nil {
		return fmt.Errorf("mail: render weekly_summary template: %w", err)
	}

	params := &resend.SendEmailRequest{
		From:    s.from,
		To:      []string{email},
		Subject: fmt.Sprintf("Tu resumen semanal — %s — Finexia", data.WeekLabel),
		Html:    body.String(),
	}

	if _, err := s.client.Emails.Send(params); err != nil {
		return fmt.Errorf("mail: send weekly summary to %s: %w", email, err)
	}

	return nil
}
