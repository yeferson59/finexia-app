package notification

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/finance/returns"
	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/identity"
	"github.com/yeferson59/finexia-app/internal/platform/mail"
	"github.com/yeferson59/finexia-app/internal/portfolio"
)

// Colors the email template paints a figure with, by sign.
const (
	gainColor = "#22c97e"
	lossColor = "#e05a5a"
)

// digestUnit is the placeholder currency the cross-portfolio total is measured
// in; see overallReturn for why the digest has no currency of its own.
const digestUnit = money.USD

// oneHundred turns the fraction returns.ROI works in into a percentage.
var oneHundred = decimal.MustFromString("100")

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
	cfg  Config
}

func NewService(user user, portfolio port, m m, cfg Config) *Service {
	return new(Service{
		user: user,
		port: portfolio,
		m:    m,
		cfg:  cfg,
	})
}

// SendWeeklySummaryEmails aggregates each subscriber's portfolios into a
// weekly digest email. It reads users and portfolio summaries through the
// module's local consumer interfaces (user/port) rather than owning that data.
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

		totalValue, totalGain := decimal.Zero, decimal.Zero
		portfolios := make([]mail.WeeklySummaryPortfolio, 0, len(summaries))

		for _, p := range summaries {
			mv := amount(p.TotalMarketValue)
			gl := amount(p.TotalGainLoss)
			glp := amount(p.TotalGainLossPct)
			totalValue = totalValue.Add(mv)
			totalGain = totalGain.Add(gl)

			color := gainColor
			if glp.IsNeg() {
				color = lossColor
			}

			portfolios = append(portfolios, mail.WeeklySummaryPortfolio{
				Name:             p.Name,
				Type:             string(p.Type),
				TotalMarketValue: fixed(mv) + " " + p.BaseCurrency,
				TotalGainLoss:    fixed(gl),
				TotalGainLossPct: fixed(glp),
				GainLossColor:    color,
			})
		}

		color := gainColor
		if totalGain.IsNeg() {
			color = lossColor
		}

		data := mail.WeeklySummaryData{
			UserName:         u.Name,
			TotalValue:       fixed(totalValue),
			TotalGainLoss:    fixed(totalGain),
			TotalGainLossPct: fixed(overallReturn(totalValue, totalGain)),
			GainLossColor:    color,
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

// amount reads one figure off a portfolio summary. The summaries arrive as
// Postgres numerics rendered to text, so they are parsed onto gofinance's
// decimal engine rather than into float64: a week's worth of positions summed
// as binary floats drifts from the total the same rows produce in SQL. An
// unparsable figure counts as zero, as it did when strconv dropped the error.
func amount(raw string) decimal.Decimal {
	d, err := decimal.NewFromString(raw)
	if err != nil {
		return decimal.Zero
	}

	return d
}

// fixed renders a figure with the two decimals the email template expects.
func fixed(d decimal.Decimal) string {
	return d.RoundBank(2).StringFixed(2)
}

// overallReturn is the user's return across every portfolio in the digest:
// the gain measured against what the holdings cost, which is the current value
// less that gain. It is returns.ROI, gofinance's own definition of profit over
// amount invested.
//
// The digest deliberately keeps summing portfolios that are denominated in
// different currencies — the template has one total and no rate to convert
// with — so the pair is handed to ROI in a single unit. The ratio is the same
// whatever unit both ends share; only the sum above mixes them.
func overallReturn(totalValue, totalGain decimal.Decimal) decimal.Decimal {
	costBase := money.FromDecimal(totalValue.Sub(totalGain), digestUnit)
	current := money.FromDecimal(totalValue, digestUnit)

	// ROI refuses a non-positive cost base, which is what stops a user whose
	// holdings net out to nothing from dividing by zero.
	roi, err := returns.ROI(costBase, current)
	if err != nil {
		return decimal.Zero
	}

	return roi.Mul(oneHundred)
}
