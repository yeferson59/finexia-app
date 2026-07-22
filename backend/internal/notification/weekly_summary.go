package notification

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/yeferson59/finexia-app/internal/identity"
	"github.com/yeferson59/finexia-app/internal/platform/config"
	"github.com/yeferson59/finexia-app/internal/platform/mail"
	"github.com/yeferson59/finexia-app/internal/portfolio"
)

type user interface {
	GetUsersWithWeeklySummary(ctx context.Context) ([]identity.User, error)
}

type port interface {
	GetPortfoliosSummary(ctx context.Context, userID uuid.UUID) ([]portfolio.SummaryView, error)
}

type m interface {
	SendWeeklySummary(email string, data mail.WeeklySummaryData) error
}

type Service struct {
	user user
	port port
	m    m
	cfg  *config.Env
}

func NewService(user user, portfolio port, m m, cfg *config.Env) *Service {
	return new(Service{
		user: user,
		port: portfolio,
		m:    m,
		cfg:  cfg,
	})
}

// SendWeeklySummaryEmails belongs to the notification domain and stays in the
// legacy services until Fase 7; it consumes the portfolio module through the
// PortfolioService interface.
func (s *Service) SendWeeklySummaryEmails(ctx context.Context) (int, []error) {
	users, err := s.user.GetUsersWithWeeklySummary(ctx)
	if err != nil {
		return 0, []error{err}
	}

	now := time.Now()
	year, week := now.ISOWeek()
	weekLabel := fmt.Sprintf("Semana %d — %d", week, year)

	var errs []error
	sent := 0

	for _, u := range users {
		summaries, err := s.port.GetPortfoliosSummary(ctx, u.ID)
		if err != nil || len(summaries) == 0 {
			continue
		}

		var totalValue, totalGain, totalGainPct float64
		portfolios := make([]mail.WeeklySummaryPortfolio, 0, len(summaries))

		for _, p := range summaries {
			mv, _ := strconv.ParseFloat(p.TotalMarketValue, 64)
			gl, _ := strconv.ParseFloat(p.TotalGainLoss, 64)
			glp, _ := strconv.ParseFloat(p.TotalGainLossPct, 64)
			totalValue += mv
			totalGain += gl

			color := "#22c97e"
			if glp < 0 {
				color = "#e05a5a"
			}

			portfolios = append(portfolios, mail.WeeklySummaryPortfolio{
				Name:             p.Name,
				Type:             string(p.Type),
				TotalMarketValue: fmt.Sprintf("%.2f %s", mv, p.BaseCurrency),
				TotalGainLoss:    fmt.Sprintf("%.2f", gl),
				TotalGainLossPct: fmt.Sprintf("%.2f", glp),
				GainLossColor:    color,
			})
		}

		if totalValue > 0 {
			totalGainPct = (totalGain / (totalValue - totalGain)) * 100
		}

		gainColor := "#22c97e"
		if totalGain < 0 {
			gainColor = "#e05a5a"
		}

		data := mail.WeeklySummaryData{
			UserName:         u.Name,
			TotalValue:       fmt.Sprintf("%.2f", totalValue),
			TotalGainLoss:    fmt.Sprintf("%.2f", totalGain),
			TotalGainLossPct: fmt.Sprintf("%.2f", totalGainPct),
			GainLossColor:    gainColor,
			Portfolios:       portfolios,
			DashboardURL:     s.cfg.PublicURL + "/dashboard",
			WeekLabel:        weekLabel,
		}

		if err := s.m.SendWeeklySummary(u.Email, data); err != nil {
			errs = append(errs, fmt.Errorf("user %s: %w", u.ID, err))
			continue
		}
		sent++
	}

	return sent, errs
}
