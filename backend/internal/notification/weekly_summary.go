package notification

import (
	"context"
	"fmt"
	"time"

	"uuid"

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
	// GetPortfolioValuesAsOf gives the digest the figures this week's are
	// compared against, one per portfolio. Portfolios with no history that far
	// back are simply absent.
	GetPortfolioValuesAsOf(ctx context.Context, userID uuid.UUID, asOf time.Time) ([]portfolio.PortfolioValuePoint, error)
	// The two below build the cash block: what the account keeps in cash, and
	// the rate each account earns on it. money.XXX asks for the figures in the
	// owner's preferred currency, which is the one the digest speaks.
	GetCashBalances(ctx context.Context, userID uuid.UUID, displayCurrency money.Currency) ([]portfolio.CashBalance, error)
	GetCashRates(ctx context.Context, userID uuid.UUID) ([]portfolio.CashRate, error)
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

		// One lookup covers both the per-portfolio rows and the account total:
		// the total's baseline is the sum of the same figures, so what the rows
		// say adds up to what the headline says.
		baseline := s.weekBaseline(ctx, u.ID, now)

		totalValue, totalGain := decimal.Zero, decimal.Zero
		baseTotal, baseCount := decimal.Zero, 0
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

			row := mail.WeeklySummaryPortfolio{
				Name:             p.Name,
				Type:             string(p.Type),
				TotalMarketValue: fixed(mv) + " " + p.BaseCurrency.String(),
				TotalGainLoss:    fixed(gl),
				TotalGainLossPct: fixed(glp),
				GainLossColor:    color,
			}

			if was, ok := baseline.values[p.ID]; ok {
				baseTotal = baseTotal.Add(was)
				baseCount++
				applyChange(mv, was, &row.HasWeekChange, &row.WeekChangeValue, &row.WeekChangePct, &row.WeekChangeColor)
			}

			portfolios = append(portfolios, row)
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

		if baseCount > 0 {
			applyChange(totalValue, baseTotal, &data.HasWeekChange, &data.WeekChangeValue, &data.WeekChangePct, &data.WeekChangeColor)
			data.WeekChangeSince = formatDay(baseline.date)
		}

		data.Cash = s.cashBlock(ctx, u.ID, now)

		if err := s.m.SendWeeklySummary(u.Email, data); err != nil {
			errs = append(errs, fmt.Errorf("user %s: %w", u.ID, err))
			continue
		}
		sent++
	}

	return sent, errs
}

// cashBlock reads the account's cash balances and the rates they earn, and
// builds the digest's cash block. A read that fails, or an account with no cash
// at all, leaves the block out: the rest of the digest is worth sending.
func (s *Service) cashBlock(ctx context.Context, userID uuid.UUID, now time.Time) *mail.WeeklySummaryCash {
	balances, err := s.port.GetCashBalances(ctx, userID, money.XXX)
	if err != nil {
		return nil
	}

	rates, err := s.port.GetCashRates(ctx, userID)
	if err != nil {
		return nil
	}

	return cashBlock(balances, rates, now)
}

// cashAccount names what a rate belongs to: a platform, a currency and a pocket
// of it (000047). The main account is the empty pocket, so a platform without
// pockets keys exactly as it always did.
type cashAccount struct {
	sourceID uuid.UUID
	currency money.Currency
	pocketID uuid.UUID
}

// pocketKey is a pocket as the key holds it: the zero UUID for the main
// account.
func pocketKey(pocketID *uuid.UUID) uuid.UUID {
	if pocketID == nil {
		return uuid.UUID{}
	}

	return *pocketID
}

// cashBlock totals the cash and the interest it earned this month, and averages
// the rate the earning balances are on, weighted by what each of them holds.
//
// The average leaves the idle balances out rather than averaging a zero into
// them: what the money that earns is earning and how much money earns nothing
// are two different figures, and the block reports both.
func cashBlock(balances []portfolio.CashBalance, rates []portfolio.CashRate, now time.Time) *mail.WeeklySummaryCash {
	if len(balances) == 0 {
		return nil
	}

	// What each account holds together, in its own currency. A rate with tiers
	// pays in steps of that figure, so what a balance earns depends on the rest
	// of its account.
	accountHeld := make(map[cashAccount]decimal.Decimal, len(balances))

	for _, b := range balances {
		key := cashAccount{sourceID: b.SourceID, currency: b.Currency, pocketID: pocketKey(b.PocketID)}

		sum, ok := accountHeld[key]
		if !ok {
			sum = decimal.Zero
		}

		accountHeld[key] = sum.Add(amount(b.Balance))
	}

	inEffect := make(map[cashAccount]decimal.Decimal, len(rates))

	for _, r := range rates {
		key := cashAccount{sourceID: r.SourceID, currency: r.Currency, pocketID: pocketKey(r.PocketID)}

		held, ok := accountHeld[key]
		if !ok || !r.InEffectOn(now) {
			continue
		}

		// A rate whose figures do not read is still the rate it quotes.
		pct, err := r.EffectiveAnnualPct(held)
		if err != nil {
			pct = amount(r.AnnualRatePct)
		}

		inEffect[key] = pct
	}

	block := mail.WeeklySummaryCash{Accounts: len(balances), Currency: balances[0].DisplayCurrency.String()}

	value, interest := decimal.Zero, decimal.Zero
	earning, weighted := decimal.Zero, decimal.Zero

	for _, b := range balances {
		held := amount(b.Value)
		value = value.Add(held)
		interest = interest.Add(amount(b.InterestThisMonthValue))

		rate, ok := inEffect[cashAccount{sourceID: b.SourceID, currency: b.Currency, pocketID: pocketKey(b.PocketID)}]
		if !ok {
			block.Idle++

			continue
		}

		earning = earning.Add(held)
		weighted = weighted.Add(held.Mul(rate))
	}

	block.Value = fixed(value)
	block.Interest = signed(interest)

	block.InterestColor = gainColor
	if interest.IsNeg() {
		block.InterestColor = lossColor
	}

	if earning.IsPos() {
		if average, err := weighted.Div(earning); err == nil {
			block.AverageRatePct = fixed(average)
		}
	}

	return &block
}

// digestPeriod is how far back the digest looks for the value it compares
// against. The job runs weekly, so a week back is the previous digest's figure.
const digestPeriod = 7 * 24 * time.Hour

// weekBaseline is what each portfolio was worth at the last digest, keyed by
// portfolio, plus the day those figures come from.
type weekBaseline struct {
	values map[uuid.UUID]decimal.Decimal
	date   time.Time
}

// weekBaseline reads the values this week's figures are compared against.
//
// The comparison is against the stored daily snapshots from a week ago, not
// against anything the mailer remembers, so a digest that failed to send — or
// a user who enabled the summary midway — still gets a truthful "since" date
// instead of a gap.
//
// A lookup that fails returns an empty baseline rather than an error: the
// digest is worth sending without the comparison. Callers then leave
// HasWeekChange false and the blocks stay hidden, because an account with no
// history has nothing to compare against and a 0.00% would claim the portfolio
// stood still when it simply was not being watched yet.
//
// The date reported is the most recent one across the portfolios found. They
// share a date in practice — SyncPortfolioSnapshots writes them in one pass —
// but a portfolio that missed a day would otherwise date the whole digest to
// its own older snapshot.
func (s *Service) weekBaseline(ctx context.Context, userID uuid.UUID, now time.Time) weekBaseline {
	out := weekBaseline{values: map[uuid.UUID]decimal.Decimal{}}

	points, err := s.port.GetPortfolioValuesAsOf(ctx, userID, now.Add(-digestPeriod))
	if err != nil {
		return out
	}

	for _, point := range points {
		out.values[point.PortfolioID] = amount(point.TotalValue)
		if point.Date.After(out.date) {
			out.date = point.Date
		}
	}

	return out
}

// applyChange fills in one movement — a portfolio row's or the account's —
// from what a figure is worth now and what it was worth at the baseline.
//
// The percentage is returns.ROI: the same "profit over amount invested" as the
// all-time figure, with last week's value standing in for the amount invested.
// ROI refuses a non-positive base, which is exactly the case where a
// percentage means nothing — there was no value to grow from, so any gain is
// an infinite one — and the pct is then left empty. The absolute change still
// says what happened.
func applyChange(now, before decimal.Decimal, has *bool, value, pct, color *string) {
	change := now.Sub(before)

	*has = true
	*value = signed(change)

	*color = gainColor
	if change.IsNeg() {
		*color = lossColor
	}

	roi, err := returns.ROI(money.NewFromDecimal(before, digestUnit), money.NewFromDecimal(now, digestUnit))
	if err != nil {
		return
	}

	*pct = signed(roi.Mul(oneHundred))
}

// signed renders a movement with an explicit sign, so a gain reads as "+12.50"
// rather than as an amount that could be either.
func signed(d decimal.Decimal) string {
	out := fixed(d)
	if !d.IsNeg() {
		return "+" + out
	}

	return out
}

// spanishMonths are the abbreviations used to date the comparison. Go's own
// month names are English, and the digest is written in Spanish.
var spanishMonths = [...]string{
	"ene", "feb", "mar", "abr", "may", "jun",
	"jul", "ago", "sep", "oct", "nov", "dic",
}

// formatDay renders a snapshot date as "29 jul".
func formatDay(t time.Time) string {
	return fmt.Sprintf("%d %s", t.Day(), spanishMonths[int(t.Month())-1])
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
	costBase := money.NewFromDecimal(totalValue.Sub(totalGain), digestUnit)
	current := money.NewFromDecimal(totalValue, digestUnit)

	// ROI refuses a non-positive cost base, which is what stops a user whose
	// holdings net out to nothing from dividing by zero.
	roi, err := returns.ROI(costBase, current)
	if err != nil {
		return decimal.Zero
	}

	return roi.Mul(oneHundred)
}
