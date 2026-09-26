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
	// The two below give the digest the week's movement: each portfolio's in
	// its own currency, and the account's in the preferred one (money.XXX). Both
	// net out what the owner paid in or took out, so a deposit is not reported
	// as the week's gain.
	GetPortfolioTrailingReturns(ctx context.Context, userID, portfolioID uuid.UUID) ([]portfolio.TrailingReturn, error)
	GetTrailingReturns(ctx context.Context, userID uuid.UUID, currency money.Currency) ([]portfolio.TrailingReturn, error)
	// The two below build the cash block: what the account keeps in cash, and
	// the rate each account earns on it. money.XXX asks for the figures in the
	// owner's preferred currency, which is the one the digest speaks.
	GetCashBalances(ctx context.Context, userID uuid.UUID, displayCurrency money.Currency) ([]portfolio.CashBalance, error)
	GetCashRates(ctx context.Context, userID uuid.UUID) ([]portfolio.CashRate, error)
	// The two below build the funds block: the funds the account follows, and
	// how each one did.
	GetFunds(ctx context.Context, userID uuid.UUID) ([]portfolio.Fund, error)
	GetFundPerformance(ctx context.Context, userID, assetID uuid.UUID) (portfolio.FundPerformance, error)
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

		// The account's week is the sum of the rows' weeks, so what the rows say
		// adds up to what the headline says; weekCount is how many rows had a
		// week to report, and weekSince the most recent day any was measured
		// from.
		totalValue, totalGain := decimal.Zero, decimal.Zero
		weekGain, weekCount := decimal.Zero, 0

		var weekSince time.Time
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

			if week, ok := s.portfolioWeek(ctx, u.ID, p.ID); ok {
				weekGain = weekGain.Add(week.Gain)
				weekCount++

				if week.From.After(weekSince) {
					weekSince = week.From
				}

				row.HasWeekChange = true
				row.WeekChangeValue = signed(week.Gain)
				row.WeekChangeColor = signColor(week.Gain)
				row.WeekChangePct = weekPct(week)
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
			DashboardURL:     s.cfg.FrontendURL + "/dashboard",
			WeekLabel:        weekLabel,
		}

		if weekCount > 0 {
			data.HasWeekChange = true
			data.WeekChangeValue = signed(weekGain)
			data.WeekChangeColor = signColor(weekGain)
			data.WeekChangePct = s.accountWeekPct(ctx, u.ID)
			data.WeekChangeSince = formatDay(weekSince)
		}

		data.Cash = s.cashBlock(ctx, u.ID, now)
		data.Funds = s.fundsBlock(ctx, u.ID, now)

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

// fundsBlock reads the funds the account follows and how each did. A read that
// fails leaves the block, or that fund, out: the rest of the digest is worth
// sending.
func (s *Service) fundsBlock(ctx context.Context, userID uuid.UUID, now time.Time) []mail.WeeklySummaryFund {
	funds, err := s.port.GetFunds(ctx, userID)
	if err != nil || len(funds) == 0 {
		return nil
	}

	rows := make([]mail.WeeklySummaryFund, 0, len(funds))

	for _, f := range funds {
		perf, err := s.port.GetFundPerformance(ctx, userID, f.AssetID)
		if err != nil {
			continue
		}

		rows = append(rows, fundRow(f, perf, now))
	}

	return rows
}

// fundRow is one fund of the digest. The return is the last 30 days', or the
// one since the fund opened when it has no mark that far back: a fund with a
// month of history says how the month went either way. A fund whose latest
// value is older than portfolio.FundStaleDays — or that has none — is flagged,
// because Finexia does not estimate between marks and the figure is as old as
// the statement it came from.
func fundRow(f portfolio.Fund, perf portfolio.FundPerformance, now time.Time) mail.WeeklySummaryFund {
	row := mail.WeeklySummaryFund{
		Name:     f.Name,
		Value:    fixed(amount(f.Value)),
		Currency: f.Currency.String(),
		Stale:    f.ValuedOn == nil,
	}

	if f.ValuedOn != nil {
		row.ValuedOn = formatDay(*f.ValuedOn)
		row.Stale = now.Sub(*f.ValuedOn) > time.Duration(portfolio.FundStaleDays)*24*time.Hour
	}

	for _, key := range []string{"30d", "inception"} {
		for _, p := range perf.Periods {
			if p.Key != key || p.Pct == nil || row.ReturnPct != "" {
				continue
			}

			pct := amount(*p.Pct)
			row.ReturnPct = signed(pct)
			row.ReturnLabel = "30 días"

			if key == "inception" {
				row.ReturnLabel = "desde el inicio"
			}

			row.ReturnColor = gainColor
			if pct.IsNeg() {
				row.ReturnColor = lossColor
			}
		}
	}

	return row
}

// weekOf picks the last week out of a set of trailing returns, if the history
// reaches that far back. A portfolio opened this week has no week to report
// and shows none, rather than a gain from nothing.
func weekOf(trailing []portfolio.TrailingReturn) (portfolio.TrailingReturn, bool) {
	for _, t := range trailing {
		if t.Period == portfolio.TrailingWeek && t.Available {
			return t, true
		}
	}

	return portfolio.TrailingReturn{}, false
}

// portfolioWeek is one portfolio's last week. A lookup that fails leaves that
// row without a comparison: the digest is worth sending without it.
//
// The week is the growth series' own (portfolio.BuildTrailingReturns): it
// nets out the money the owner moved, where the comparison it replaces set
// this week's value against last week's and reported a deposit as the week's
// gain. It is measured from the last snapshot on or before a week ago — the
// stored series, not anything the mailer remembers — so a digest that failed
// to send, or a user who enabled the summary midway, still gets a truthful
// "since" date instead of a gap.
func (s *Service) portfolioWeek(ctx context.Context, userID, portfolioID uuid.UUID) (portfolio.TrailingReturn, bool) {
	trailing, err := s.port.GetPortfolioTrailingReturns(ctx, userID, portfolioID)
	if err != nil {
		return portfolio.TrailingReturn{}, false
	}

	return weekOf(trailing)
}

// accountWeekPct is the account's return over the week, or empty when there is
// none to give.
//
// It comes from the account-wide series rather than from the rows: returns do
// not add up the way amounts do, and chaining the account's own subperiods is
// what weighs each portfolio by what it held. That series is in the preferred
// currency, and a percentage is the one figure the unit does not change.
func (s *Service) accountWeekPct(ctx context.Context, userID uuid.UUID) string {
	trailing, err := s.port.GetTrailingReturns(ctx, userID, money.XXX)
	if err != nil {
		return ""
	}

	week, ok := weekOf(trailing)
	if !ok {
		return ""
	}

	return weekPct(week)
}

// weekPct renders a week's time-weighted return, or nothing when it has none:
// a portfolio that held nothing all week has no capital a percentage could be
// of, and the amount stands on its own.
func weekPct(week portfolio.TrailingReturn) string {
	if !week.HasRate {
		return ""
	}

	return signed(week.Rate.Mul(oneHundred))
}

// signColor is the color a movement is painted by its sign.
func signColor(d decimal.Decimal) string {
	if d.IsNeg() {
		return lossColor
	}

	return gainColor
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
