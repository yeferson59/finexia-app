package notification

import (
	"context"
	"errors"
	"testing"
	"time"

	"uuid"

	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/identity"
	"github.com/yeferson59/finexia-app/internal/platform/mail"
	"github.com/yeferson59/finexia-app/internal/portfolio"
)

// fakeUserReader stubs the users the weekly summary iterates over.
type fakeUserReader struct {
	getUsers func(ctx context.Context) ([]identity.User, error)
}

func (f *fakeUserReader) GetUsersWithWeeklySummary(ctx context.Context) ([]identity.User, error) {
	return f.getUsers(ctx)
}

// fakePortfolioReader stubs the per-user portfolio summaries and the week each
// portfolio, and the account, earned.
type fakePortfolioReader struct {
	getSummary func(ctx context.Context, userID uuid.UUID) ([]portfolio.SummaryView, error)
	// The two below are optional: left unset, the account holds no cash and
	// the digest carries no cash block.
	getCashBalances func(ctx context.Context, userID uuid.UUID, display money.Currency) ([]portfolio.CashBalance, error)
	getCashRates    func(ctx context.Context, userID uuid.UUID) ([]portfolio.CashRate, error)
	// Optional as well: left unset, the account follows no fund.
	getFunds           func(ctx context.Context, userID uuid.UUID) ([]portfolio.Fund, error)
	getFundPerformance func(ctx context.Context, userID, assetID uuid.UUID) (portfolio.FundPerformance, error)
	// The last two are optional as well: left unset, the account reads as having no
	// history, which is what an unstubbed test wants.
	getPortfolioTrailingReturns func(ctx context.Context, userID, portfolioID uuid.UUID) ([]portfolio.TrailingReturn, error)
	getTrailingReturns          func(ctx context.Context, userID uuid.UUID, cur money.Currency) ([]portfolio.TrailingReturn, error)
}

func (f *fakePortfolioReader) GetFunds(ctx context.Context, userID uuid.UUID) ([]portfolio.Fund, error) {
	if f.getFunds == nil {
		return nil, nil
	}

	return f.getFunds(ctx, userID)
}

func (f *fakePortfolioReader) GetFundPerformance(ctx context.Context, userID, assetID uuid.UUID) (portfolio.FundPerformance, error) {
	if f.getFundPerformance == nil {
		return portfolio.FundPerformance{}, nil
	}

	return f.getFundPerformance(ctx, userID, assetID)
}

func (f *fakePortfolioReader) GetPortfoliosSummary(ctx context.Context, userID uuid.UUID) ([]portfolio.SummaryView, error) {
	return f.getSummary(ctx, userID)
}

func (f *fakePortfolioReader) GetPortfolioTrailingReturns(ctx context.Context, userID, portfolioID uuid.UUID) ([]portfolio.TrailingReturn, error) {
	if f.getPortfolioTrailingReturns == nil {
		return nil, nil
	}

	return f.getPortfolioTrailingReturns(ctx, userID, portfolioID)
}

func (f *fakePortfolioReader) GetTrailingReturns(ctx context.Context, userID uuid.UUID, cur money.Currency) ([]portfolio.TrailingReturn, error) {
	if f.getTrailingReturns == nil {
		return nil, nil
	}

	return f.getTrailingReturns(ctx, userID, cur)
}

func (f *fakePortfolioReader) GetCashBalances(ctx context.Context, userID uuid.UUID, display money.Currency) ([]portfolio.CashBalance, error) {
	if f.getCashBalances == nil {
		return nil, nil
	}

	return f.getCashBalances(ctx, userID, display)
}

func (f *fakePortfolioReader) GetCashRates(ctx context.Context, userID uuid.UUID) ([]portfolio.CashRate, error) {
	if f.getCashRates == nil {
		return nil, nil
	}

	return f.getCashRates(ctx, userID)
}

// sentWeekly records one SendWeeklySummary call.
type sentWeekly struct {
	To   string
	Data mail.WeeklySummaryData
}

// fakeMailer captures the weekly summary emails; weeklyErr makes every send fail.
type fakeMailer struct {
	weekly    []sentWeekly
	weeklyErr error
}

func (f *fakeMailer) SendWeeklySummary(email string, data mail.WeeklySummaryData) error {
	if f.weeklyErr != nil {
		return f.weeklyErr
	}
	f.weekly = append(f.weekly, sentWeekly{To: email, Data: data})
	return nil
}

func newTestService(users *fakeUserReader, ports *fakePortfolioReader, mailer *fakeMailer) *Service {
	return NewService(users, ports, mailer, Config{FrontendURL: "https://finexia.test"})
}

func TestSendWeeklySummaryEmails(t *testing.T) {
	t.Run("aggregates portfolios and emails each subscriber", func(t *testing.T) {
		u := identity.User{ID: uuid.New(), Name: "Ada", Email: "ada@example.com"}
		summaries := []portfolio.SummaryView{
			{Name: "Growth", Type: portfolio.TypeStocks, BaseCurrency: money.USD, TotalMarketValue: "600.00", TotalGainLoss: "100.00", TotalGainLossPct: "20.00"},
			{Name: "Crypto", Type: portfolio.TypeCryptos, BaseCurrency: money.USD, TotalMarketValue: "500.00", TotalGainLoss: "-50.00", TotalGainLossPct: "-9.09"},
		}
		users := new(fakeUserReader{getUsers: func(context.Context) ([]identity.User, error) {
			return []identity.User{u}, nil
		}})
		ports := new(fakePortfolioReader{getSummary: func(_ context.Context, uid uuid.UUID) ([]portfolio.SummaryView, error) {
			if uid != u.ID {
				t.Errorf("userID = %s, want %s", uid, u.ID)
			}
			return summaries, nil
		}})
		mailer := new(fakeMailer{})
		svc := newTestService(users, ports, mailer)

		sent, errs := svc.SendWeeklySummaryEmails(context.Background())
		if sent != 1 || len(errs) != 0 {
			t.Fatalf("sent/errs = %d/%v, want 1/none", sent, errs)
		}
		if len(mailer.weekly) != 1 {
			t.Fatalf("weekly emails = %d, want 1", len(mailer.weekly))
		}

		msg := mailer.weekly[0]
		if msg.To != "ada@example.com" {
			t.Errorf("recipient = %q", msg.To)
		}
		data := msg.Data
		if data.UserName != "Ada" {
			t.Errorf("UserName = %q", data.UserName)
		}
		if data.TotalValue != "1100.00" {
			t.Errorf("TotalValue = %q, want 1100.00", data.TotalValue)
		}
		if data.TotalGainLoss != "50.00" {
			t.Errorf("TotalGainLoss = %q, want 50.00", data.TotalGainLoss)
		}
		// 50 gain over a 1050 cost base => 4.76%
		if data.TotalGainLossPct != "4.76" {
			t.Errorf("TotalGainLossPct = %q, want 4.76", data.TotalGainLossPct)
		}
		if data.GainLossColor != "#22c97e" {
			t.Errorf("GainLossColor = %q, want green for a positive total", data.GainLossColor)
		}
		if len(data.Portfolios) != 2 {
			t.Fatalf("portfolios in email = %d, want 2", len(data.Portfolios))
		}
		if data.Portfolios[0].TotalMarketValue != "600.00 USD" {
			t.Errorf("portfolio 1 market value = %q, want '600.00 USD'", data.Portfolios[0].TotalMarketValue)
		}
		if data.Portfolios[0].GainLossColor != "#22c97e" {
			t.Errorf("portfolio 1 color = %q, want green", data.Portfolios[0].GainLossColor)
		}
		if data.Portfolios[1].GainLossColor != "#e05a5a" {
			t.Errorf("portfolio 2 color = %q, want red for a losing portfolio", data.Portfolios[1].GainLossColor)
		}
	})

	t.Run("negative overall gain uses the red color", func(t *testing.T) {
		users := new(fakeUserReader{getUsers: func(context.Context) ([]identity.User, error) {
			return []identity.User{{ID: uuid.New(), Email: "x@example.com"}}, nil
		}})
		ports := new(fakePortfolioReader{getSummary: func(context.Context, uuid.UUID) ([]portfolio.SummaryView, error) {
			return []portfolio.SummaryView{
				{Name: "Down", BaseCurrency: money.USD, TotalMarketValue: "900.00", TotalGainLoss: "-100.00", TotalGainLossPct: "-10.00"},
			}, nil
		}})
		mailer := new(fakeMailer{})
		svc := newTestService(users, ports, mailer)

		sent, errs := svc.SendWeeklySummaryEmails(context.Background())
		if sent != 1 || len(errs) != 0 {
			t.Fatalf("sent/errs = %d/%v", sent, errs)
		}
		data := mailer.weekly[0].Data
		if data.GainLossColor != "#e05a5a" {
			t.Errorf("GainLossColor = %q, want red", data.GainLossColor)
		}
		// -100 over a 1000 cost base => -10%
		if data.TotalGainLossPct != "-10.00" {
			t.Errorf("TotalGainLossPct = %q, want -10.00", data.TotalGainLossPct)
		}
	})

	t.Run("users without portfolios are skipped", func(t *testing.T) {
		users := new(fakeUserReader{getUsers: func(context.Context) ([]identity.User, error) {
			return []identity.User{{ID: uuid.New(), Email: "empty@example.com"}}, nil
		}})
		ports := new(fakePortfolioReader{getSummary: func(context.Context, uuid.UUID) ([]portfolio.SummaryView, error) {
			return []portfolio.SummaryView{}, nil
		}})
		mailer := new(fakeMailer{})
		svc := newTestService(users, ports, mailer)

		sent, errs := svc.SendWeeklySummaryEmails(context.Background())
		if sent != 0 || len(errs) != 0 {
			t.Errorf("sent/errs = %d/%v, want 0/none", sent, errs)
		}
		if len(mailer.weekly) != 0 {
			t.Errorf("no email should be sent for users without portfolios")
		}
	})

	t.Run("summary lookup failure skips the user without failing the batch", func(t *testing.T) {
		okUser := identity.User{ID: uuid.New(), Email: "ok@example.com"}
		badUser := identity.User{ID: uuid.New(), Email: "bad@example.com"}
		users := new(fakeUserReader{getUsers: func(context.Context) ([]identity.User, error) {
			return []identity.User{badUser, okUser}, nil
		}})
		ports := new(fakePortfolioReader{getSummary: func(_ context.Context, uid uuid.UUID) ([]portfolio.SummaryView, error) {
			if uid == badUser.ID {
				return nil, errors.New("summary view broken")
			}
			return []portfolio.SummaryView{{Name: "P", BaseCurrency: money.USD, TotalMarketValue: "10.00", TotalGainLoss: "1.00", TotalGainLossPct: "11.11"}}, nil
		}})
		mailer := new(fakeMailer{})
		svc := newTestService(users, ports, mailer)

		sent, errs := svc.SendWeeklySummaryEmails(context.Background())
		if sent != 1 || len(errs) != 0 {
			t.Errorf("sent/errs = %d/%v, want 1/none", sent, errs)
		}
		if len(mailer.weekly) != 1 || mailer.weekly[0].To != "ok@example.com" {
			t.Errorf("weekly = %+v, want a single email to ok@example.com", mailer.weekly)
		}
	})

	t.Run("mail failures are collected per user", func(t *testing.T) {
		users := new(fakeUserReader{getUsers: func(context.Context) ([]identity.User, error) {
			return []identity.User{{ID: uuid.New(), Email: "x@example.com"}}, nil
		}})
		ports := new(fakePortfolioReader{getSummary: func(context.Context, uuid.UUID) ([]portfolio.SummaryView, error) {
			return []portfolio.SummaryView{{Name: "P", BaseCurrency: money.USD, TotalMarketValue: "10.00", TotalGainLoss: "0.00", TotalGainLossPct: "0.00"}}, nil
		}})
		mailer := new(fakeMailer{weeklyErr: errors.New("smtp down")})
		svc := newTestService(users, ports, mailer)

		sent, errs := svc.SendWeeklySummaryEmails(context.Background())
		if sent != 0 || len(errs) != 1 {
			t.Errorf("sent/errs = %d/%v, want 0 and one error", sent, errs)
		}
	})

	t.Run("subscriber query failure aborts", func(t *testing.T) {
		users := new(fakeUserReader{getUsers: func(context.Context) ([]identity.User, error) {
			return nil, errors.New("db down")
		}})
		svc := newTestService(users, new(fakePortfolioReader{}), new(fakeMailer{}))

		sent, errs := svc.SendWeeklySummaryEmails(context.Background())
		if sent != 0 || len(errs) != 1 {
			t.Errorf("sent/errs = %d/%v, want 0 and one error", sent, errs)
		}
	})
}

// The digest's arithmetic runs on gofinance's decimal engine and returns.ROI
// rather than float64, so long runs of positions no longer drift from the
// figures the same rows produce in SQL.

func TestWeeklySummaryArithmetic(t *testing.T) {
	send := func(t *testing.T, summaries []portfolio.SummaryView) mail.WeeklySummaryData {
		t.Helper()
		users := new(fakeUserReader{getUsers: func(context.Context) ([]identity.User, error) {
			return []identity.User{{ID: uuid.New(), Name: "Ada", Email: "ada@example.com"}}, nil
		}})
		ports := new(fakePortfolioReader{getSummary: func(context.Context, uuid.UUID) ([]portfolio.SummaryView, error) {
			return summaries, nil
		}})
		mailer := new(fakeMailer{})
		sent, errs := newTestService(users, ports, mailer).SendWeeklySummaryEmails(context.Background())
		if sent != 1 || len(errs) != 0 {
			t.Fatalf("sent/errs = %d/%v, want 1/none", sent, errs)
		}
		return mailer.weekly[0].Data
	}

	t.Run("many positions sum exactly", func(t *testing.T) {
		// 0.07 is not representable in binary; summing it a hundred times in
		// float64 lands on 7.000000000000005, not 7.
		summaries := make([]portfolio.SummaryView, 0, 100)
		for range 100 {
			summaries = append(summaries, portfolio.SummaryView{
				Name: "P", BaseCurrency: money.USD,
				TotalMarketValue: "0.07", TotalGainLoss: "0.00", TotalGainLossPct: "0.00",
			})
		}

		data := send(t, summaries)
		if data.TotalValue != "7.00" {
			t.Errorf("TotalValue = %q, want 7.00", data.TotalValue)
		}
	})

	t.Run("holdings that net out to their cost report no return", func(t *testing.T) {
		// Value equals gain, so the cost base is zero: returns.ROI refuses it
		// rather than dividing by it.
		data := send(t, []portfolio.SummaryView{{
			Name: "P", BaseCurrency: money.USD,
			TotalMarketValue: "500.00", TotalGainLoss: "500.00", TotalGainLossPct: "0.00",
		}})
		if data.TotalGainLossPct != "0.00" {
			t.Errorf("TotalGainLossPct = %q, want 0.00", data.TotalGainLossPct)
		}
	})

	t.Run("a total wiped out to zero reports no return", func(t *testing.T) {
		data := send(t, []portfolio.SummaryView{{
			Name: "P", BaseCurrency: money.USD,
			TotalMarketValue: "0", TotalGainLoss: "0", TotalGainLossPct: "0",
		}})
		if data.TotalGainLossPct != "0.00" || data.TotalValue != "0.00" {
			t.Errorf("data = %+v, want zeroes", data)
		}
	})

	t.Run("an unparsable figure counts as zero instead of breaking the digest", func(t *testing.T) {
		data := send(t, []portfolio.SummaryView{
			{Name: "Broken", BaseCurrency: money.USD, TotalMarketValue: "n/a", TotalGainLoss: "n/a", TotalGainLossPct: "n/a"},
			{Name: "Fine", BaseCurrency: money.USD, TotalMarketValue: "100.00", TotalGainLoss: "10.00", TotalGainLossPct: "11.11"},
		})
		if data.TotalValue != "100.00" || data.TotalGainLoss != "10.00" {
			t.Errorf("totals = %q/%q, want 100.00/10.00", data.TotalValue, data.TotalGainLoss)
		}
		if data.Portfolios[0].TotalMarketValue != "0.00 USD" {
			t.Errorf("broken row = %q, want '0.00 USD'", data.Portfolios[0].TotalMarketValue)
		}
	})

	t.Run("figures are rendered with two decimals", func(t *testing.T) {
		data := send(t, []portfolio.SummaryView{{
			Name: "P", BaseCurrency: money.COP,
			TotalMarketValue: "4123456.789", TotalGainLoss: "1.005", TotalGainLossPct: "0.5",
		}})
		if data.Portfolios[0].TotalMarketValue != "4123456.79 COP" {
			t.Errorf("market value = %q, want '4123456.79 COP'", data.Portfolios[0].TotalMarketValue)
		}
		if data.Portfolios[0].TotalGainLossPct != "0.50" {
			t.Errorf("pct = %q, want 0.50", data.Portfolios[0].TotalGainLossPct)
		}
	})
}

func TestWeeklySummaryWeekOverWeekChange(t *testing.T) {
	userID := uuid.New()
	stocksID, cryptoID := uuid.New(), uuid.New()
	lastMonday := time.Date(2026, time.July, 29, 0, 0, 0, 0, time.UTC)

	// weekFrom is a history whose last week opens on from and earned gain; pct
	// is its time-weighted return in percent, "" for a week with none.
	weekFrom := func(from time.Time, gain, pct string) []portfolio.TrailingReturn {
		week := portfolio.TrailingReturn{
			Period:    portfolio.TrailingWeek,
			Available: true,
			From:      from,
			Gain:      decimal.MustFromString(gain),
		}

		if pct != "" {
			rate, err := decimal.MustFromString(pct).Div(decimal.MustFromString("100"))
			if err != nil {
				panic(err)
			}

			week.Rate, week.HasRate = rate, true
		}

		// The day before, which a week's digest must not pick up instead.
		day := portfolio.TrailingReturn{Period: portfolio.TrailingDay, Available: true, Gain: decimal.MustFromString("999")}

		return []portfolio.TrailingReturn{day, week}
	}
	week := func(gain, pct string) []portfolio.TrailingReturn { return weekFrom(lastMonday, gain, pct) }

	// byPortfolio stubs each portfolio's trailing returns; one absent from the
	// map has no history.
	byPortfolio := func(weeks map[uuid.UUID][]portfolio.TrailingReturn) func(context.Context, uuid.UUID, uuid.UUID) ([]portfolio.TrailingReturn, error) {
		return func(_ context.Context, _, portfolioID uuid.UUID) ([]portfolio.TrailingReturn, error) {
			return weeks[portfolioID], nil
		}
	}

	account := func(trailing []portfolio.TrailingReturn) func(context.Context, uuid.UUID, money.Currency) ([]portfolio.TrailingReturn, error) {
		return func(context.Context, uuid.UUID, money.Currency) ([]portfolio.TrailingReturn, error) {
			return trailing, nil
		}
	}

	// send runs the digest with stubbed weeks and returns the email data.
	send := func(t *testing.T, summaries []portfolio.SummaryView, ports *fakePortfolioReader) mail.WeeklySummaryData {
		t.Helper()
		users := new(fakeUserReader{getUsers: func(context.Context) ([]identity.User, error) {
			return []identity.User{{ID: userID, Name: "Ada", Email: "ada@example.com"}}, nil
		}})
		ports.getSummary = func(context.Context, uuid.UUID) ([]portfolio.SummaryView, error) {
			return summaries, nil
		}
		mailer := new(fakeMailer{})
		sent, errs := newTestService(users, ports, mailer).SendWeeklySummaryEmails(context.Background())
		if sent != 1 || len(errs) != 0 {
			t.Fatalf("sent/errs = %d/%v, want 1/none", sent, errs)
		}
		return mailer.weekly[0].Data
	}

	twoPortfolios := []portfolio.SummaryView{
		{ID: stocksID, Name: "Acciones", BaseCurrency: money.USD, TotalMarketValue: "1100.00", TotalGainLoss: "50.00", TotalGainLossPct: "4.76"},
		{ID: cryptoID, Name: "Cripto", BaseCurrency: money.USD, TotalMarketValue: "400.00", TotalGainLoss: "-20.00", TotalGainLossPct: "-4.76"},
	}

	onePortfolio := []portfolio.SummaryView{twoPortfolios[0]}

	t.Run("each portfolio reports its own week", func(t *testing.T) {
		data := send(t, twoPortfolios, &fakePortfolioReader{
			getPortfolioTrailingReturns: byPortfolio(map[uuid.UUID][]portfolio.TrailingReturn{
				stocksID: week("100", "10"),
				cryptoID: week("-100", "-20"),
			}),
		})

		if len(data.Portfolios) != 2 {
			t.Fatalf("portfolios = %d, want 2", len(data.Portfolios))
		}

		stocks := data.Portfolios[0]
		if !stocks.HasWeekChange {
			t.Fatal("the stocks row should carry a comparison")
		}
		if stocks.WeekChangeValue != "+100.00" || stocks.WeekChangePct != "+10.00" {
			t.Errorf("stocks change = %q/%q, want +100.00/+10.00", stocks.WeekChangeValue, stocks.WeekChangePct)
		}
		if stocks.WeekChangeColor != gainColor {
			t.Errorf("stocks color = %q, want the gain color", stocks.WeekChangeColor)
		}

		crypto := data.Portfolios[1]
		if crypto.WeekChangeValue != "-100.00" || crypto.WeekChangePct != "-20.00" {
			t.Errorf("crypto change = %q/%q, want -100.00/-20.00", crypto.WeekChangeValue, crypto.WeekChangePct)
		}
		if crypto.WeekChangeColor != lossColor {
			t.Errorf("crypto color = %q, want the loss color", crypto.WeekChangeColor)
		}
	})

	// The reason the digest reads the growth series: the stocks portfolio is
	// worth 1100 against 1000 a week ago, but 100 of that was a deposit. The
	// old comparison of the two values reported +100.00 (+10%) as the week's
	// gain; the week the series measures earned nothing.
	t.Run("a deposit is not the week's gain", func(t *testing.T) {
		data := send(t, onePortfolio, &fakePortfolioReader{
			getPortfolioTrailingReturns: byPortfolio(map[uuid.UUID][]portfolio.TrailingReturn{stocksID: week("0", "0")}),
			getTrailingReturns:          account(week("0", "0")),
		})

		if data.WeekChangeValue != "+0.00" || data.WeekChangePct != "+0.00" {
			t.Errorf("change = %q/%q, want +0.00/+0.00", data.WeekChangeValue, data.WeekChangePct)
		}
		if data.TotalValue != "1100.00" {
			t.Errorf("TotalValue = %q, want the value as it stands, deposit included", data.TotalValue)
		}
	})

	t.Run("the rows add up to the account total", func(t *testing.T) {
		// +100 on stocks and -100 on crypto net out: the headline has to say
		// so, or the reader can add up the rows and catch the digest lying.
		data := send(t, twoPortfolios, &fakePortfolioReader{
			getPortfolioTrailingReturns: byPortfolio(map[uuid.UUID][]portfolio.TrailingReturn{
				stocksID: week("100", "10"),
				cryptoID: week("-100", "-20"),
			}),
			getTrailingReturns: account(week("0", "0")),
		})

		if data.WeekChangeValue != "+0.00" {
			t.Errorf("total change = %q, want +0.00 (the rows cancel out)", data.WeekChangeValue)
		}
		if data.WeekChangePct != "+0.00" {
			t.Errorf("total pct = %q, want +0.00", data.WeekChangePct)
		}
		if data.WeekChangeSince != "29 jul" {
			t.Errorf("WeekChangeSince = %q, want '29 jul'", data.WeekChangeSince)
		}
	})

	// Returns do not add up: +10% on a big portfolio and -20% on a small one is
	// not -10% for the account. The account's own series weighs them.
	t.Run("the account percentage comes from the account's series", func(t *testing.T) {
		var asked money.Currency
		data := send(t, twoPortfolios, &fakePortfolioReader{
			getPortfolioTrailingReturns: byPortfolio(map[uuid.UUID][]portfolio.TrailingReturn{
				stocksID: week("100", "10"),
				cryptoID: week("-100", "-20"),
			}),
			getTrailingReturns: func(_ context.Context, uid uuid.UUID, cur money.Currency) ([]portfolio.TrailingReturn, error) {
				if uid != userID {
					t.Errorf("userID = %s, want %s", uid, userID)
				}
				asked = cur
				return week("0", "3.25"), nil
			},
		})

		if data.WeekChangePct != "+3.25" {
			t.Errorf("total pct = %q, want the account's +3.25", data.WeekChangePct)
		}
		if asked != money.XXX {
			t.Errorf("currency = %q, want the preferred one (XXX)", asked)
		}
	})

	t.Run("a portfolio opened this week has no comparison and stays out of the total", func(t *testing.T) {
		data := send(t, twoPortfolios, &fakePortfolioReader{
			getPortfolioTrailingReturns: byPortfolio(map[uuid.UUID][]portfolio.TrailingReturn{stocksID: week("100", "10")}),
		})

		if !data.Portfolios[0].HasWeekChange {
			t.Error("the portfolio with history should carry a comparison")
		}
		if data.Portfolios[1].HasWeekChange {
			t.Error("a portfolio with no history should show no comparison")
		}
		if data.WeekChangeValue != "+100.00" {
			t.Errorf("total change = %q, want the stocks' +100.00 alone", data.WeekChangeValue)
		}
	})

	t.Run("a week that lost money reports a negative amount in the loss color", func(t *testing.T) {
		data := send(t, onePortfolio, &fakePortfolioReader{
			getPortfolioTrailingReturns: byPortfolio(map[uuid.UUID][]portfolio.TrailingReturn{stocksID: week("-150", "-12")}),
			getTrailingReturns:          account(week("-150", "-12")),
		})

		if data.WeekChangeValue != "-150.00" || data.WeekChangePct != "-12.00" {
			t.Errorf("change = %q/%q, want -150.00/-12.00", data.WeekChangeValue, data.WeekChangePct)
		}
		if data.WeekChangeColor != lossColor {
			t.Errorf("WeekChangeColor = %q, want the loss color", data.WeekChangeColor)
		}
	})

	t.Run("a week with no capital at work shows the amount but no percentage", func(t *testing.T) {
		data := send(t, onePortfolio, &fakePortfolioReader{
			getPortfolioTrailingReturns: byPortfolio(map[uuid.UUID][]portfolio.TrailingReturn{stocksID: week("0", "")}),
			getTrailingReturns:          account(week("0", "")),
		})

		if !data.HasWeekChange || data.WeekChangeValue != "+0.00" {
			t.Errorf("WeekChangeValue = %q, want +0.00", data.WeekChangeValue)
		}
		if data.WeekChangePct != "" || data.Portfolios[0].WeekChangePct != "" {
			t.Errorf("pct = %q / row %q, want both empty", data.WeekChangePct, data.Portfolios[0].WeekChangePct)
		}
	})

	t.Run("an account with no history hides the comparison", func(t *testing.T) {
		data := send(t, onePortfolio, &fakePortfolioReader{
			getPortfolioTrailingReturns: byPortfolio(map[uuid.UUID][]portfolio.TrailingReturn{
				stocksID: {{Period: portfolio.TrailingWeek, Available: false}},
			}),
		})

		if data.HasWeekChange {
			t.Error("HasWeekChange = true, want the block hidden with no baseline")
		}
		if data.Portfolios[0].HasWeekChange {
			t.Error("the portfolio row should show no comparison either")
		}
		if data.TotalValue != "1100.00" {
			t.Errorf("the rest of the digest should still be built, got TotalValue %q", data.TotalValue)
		}
	})

	t.Run("a failed lookup does not stop the digest", func(t *testing.T) {
		data := send(t, onePortfolio, &fakePortfolioReader{
			getPortfolioTrailingReturns: func(context.Context, uuid.UUID, uuid.UUID) ([]portfolio.TrailingReturn, error) {
				return nil, errors.New("snapshots table on fire")
			},
		})

		if data.HasWeekChange {
			t.Error("a failed lookup should leave the comparison out")
		}
		if data.TotalValue != "1100.00" {
			t.Errorf("TotalValue = %q, want the digest sent anyway", data.TotalValue)
		}
	})

	t.Run("a failed account lookup leaves only the percentage out", func(t *testing.T) {
		data := send(t, onePortfolio, &fakePortfolioReader{
			getPortfolioTrailingReturns: byPortfolio(map[uuid.UUID][]portfolio.TrailingReturn{stocksID: week("100", "10")}),
			getTrailingReturns: func(context.Context, uuid.UUID, money.Currency) ([]portfolio.TrailingReturn, error) {
				return nil, errors.New("no rate")
			},
		})

		if !data.HasWeekChange || data.WeekChangeValue != "+100.00" || data.WeekChangePct != "" {
			t.Errorf("change = %v %q/%q, want +100.00 with no percentage", data.HasWeekChange, data.WeekChangeValue, data.WeekChangePct)
		}
	})

	t.Run("the digest is dated by the most recent week start", func(t *testing.T) {
		// One portfolio missed a day. The digest should say the newer date
		// rather than backdating everything to the stale one.
		data := send(t, twoPortfolios, &fakePortfolioReader{
			getPortfolioTrailingReturns: byPortfolio(map[uuid.UUID][]portfolio.TrailingReturn{
				stocksID: weekFrom(lastMonday.AddDate(0, 0, -3), "100", "10"),
				cryptoID: week("-100", "-20"),
			}),
		})

		if data.WeekChangeSince != "29 jul" {
			t.Errorf("WeekChangeSince = %q, want the most recent week start", data.WeekChangeSince)
		}
	})

	t.Run("the total adds exact decimals, not float64", func(t *testing.T) {
		// 0.07 a hundred times is 7 exactly; in float64 it is 7.000000000000005.
		summaries := make([]portfolio.SummaryView, 0, 100)
		weeks := map[uuid.UUID][]portfolio.TrailingReturn{}
		for range 100 {
			id := uuid.New()
			summaries = append(summaries, portfolio.SummaryView{
				ID: id, Name: "P", BaseCurrency: money.USD,
				TotalMarketValue: "0.07", TotalGainLoss: "0.00", TotalGainLossPct: "0.00",
			})
			weeks[id] = week("0.07", "")
		}

		data := send(t, summaries, &fakePortfolioReader{getPortfolioTrailingReturns: byPortfolio(weeks)})

		if data.WeekChangeValue != "+7.00" {
			t.Errorf("WeekChangeValue = %q, want +7.00", data.WeekChangeValue)
		}
	})
}

func TestFormatDay(t *testing.T) {
	cases := []struct {
		date time.Time
		want string
	}{
		{time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC), "1 ene"},
		{time.Date(2026, time.July, 29, 0, 0, 0, 0, time.UTC), "29 jul"},
		{time.Date(2026, time.December, 31, 0, 0, 0, 0, time.UTC), "31 dic"},
	}

	for _, tc := range cases {
		if got := formatDay(tc.date); got != tc.want {
			t.Errorf("formatDay(%v) = %q, want %q", tc.date, got, tc.want)
		}
	}
}

// The cash block totals what the account keeps in cash, what that earned this
// month, and the rate the earning balances are on — weighted by what each of
// them holds, with the idle ones counted rather than averaged in.
func TestCashBlock(t *testing.T) {
	now := time.Date(2026, time.September, 15, 0, 0, 0, 0, time.UTC)
	nu, davivienda := uuid.New(), uuid.New()

	balances := []portfolio.CashBalance{
		{
			SourceID: nu, Currency: money.COP, DisplayCurrency: money.COP,
			Value: "30000000", InterestThisMonthValue: "34000",
		},
		{
			SourceID: davivienda, Currency: money.COP, DisplayCurrency: money.COP,
			Value: "10000000", InterestThisMonthValue: "0",
		},
	}

	rates := []portfolio.CashRate{
		{SourceID: nu, Currency: money.COP, AnnualRatePct: "9", EffectiveFrom: now.AddDate(0, -1, 0), Latest: true},
	}

	block := cashBlock(balances, rates, now)
	if block == nil {
		t.Fatal("no cash block for an account that holds cash")
	}

	if block.Value != "40000000.00" || block.Currency != "COP" {
		t.Errorf("value = %s %s, want 40000000.00 COP", block.Value, block.Currency)
	}
	if block.Interest != "+34000.00" || block.InterestColor != gainColor {
		t.Errorf("interest = %s in %s, want +34000.00 in the gain colour", block.Interest, block.InterestColor)
	}
	// Only the balance that earns is averaged: 9 %, not 6.75 %.
	if block.AverageRatePct != "9.00" {
		t.Errorf("average rate = %s, want 9.00", block.AverageRatePct)
	}
	if block.Accounts != 2 || block.Idle != 1 {
		t.Errorf("accounts = %d with %d idle, want 2 and 1", block.Accounts, block.Idle)
	}
}

// An account with no cash gets no block, and one whose balances all sit idle
// gets a block with no average to report.
func TestCashBlockWithoutRates(t *testing.T) {
	now := time.Date(2026, time.September, 15, 0, 0, 0, 0, time.UTC)

	if block := cashBlock(nil, nil, now); block != nil {
		t.Errorf("block = %+v, want none without cash", block)
	}

	ended := now.AddDate(0, 0, -1)
	idle := []portfolio.CashBalance{{
		SourceID: uuid.New(), Currency: money.USD, DisplayCurrency: money.USD,
		Value: "500", InterestThisMonthValue: "0",
	}}
	// The account's only version stopped yesterday, so nothing earns today.
	paused := []portfolio.CashRate{{
		SourceID: idle[0].SourceID, Currency: money.USD, AnnualRatePct: "4",
		EffectiveFrom: now.AddDate(0, -2, 0), EndedOn: &ended, Latest: true,
	}}

	block := cashBlock(idle, paused, now)
	if block == nil {
		t.Fatal("no cash block for an account that holds cash")
	}
	if block.AverageRatePct != "" || block.Idle != 1 {
		t.Errorf("block = %+v, want no average and one idle balance", block)
	}
}

// A rate with tiers averages in at what its account earns on all it holds, not
// at the rate it pays from zero, whichever portfolio holds each part.
func TestCashBlockWeighsATieredRateAtWhatItComesTo(t *testing.T) {
	now := time.Date(2026, time.September, 15, 0, 0, 0, 0, time.UTC)
	nu := uuid.New()

	balances := []portfolio.CashBalance{
		{SourceID: nu, Currency: money.COP, DisplayCurrency: money.COP, Balance: "6000000", Value: "6000000", InterestThisMonthValue: "0"},
		{SourceID: nu, Currency: money.COP, DisplayCurrency: money.COP, Balance: "2000000", Value: "2000000", InterestThisMonthValue: "0"},
	}

	rates := []portfolio.CashRate{{
		SourceID: nu, Currency: money.COP, AnnualRatePct: "12", EffectiveFrom: now.AddDate(0, -1, 0), Latest: true,
		Tiers: []portfolio.CashRateTier{{FromBalance: "5000000", AnnualRatePct: "8"}},
	}}

	block := cashBlock(balances, rates, now)
	if block == nil {
		t.Fatal("no cash block for an account that holds cash")
	}

	// 12 % on the first five million and 8 % on the other three come to 10.48 %
	// on all eight.
	if block.AverageRatePct != "10.48" {
		t.Errorf("average rate = %s, want 10.48", block.AverageRatePct)
	}
}

// A fund in the digest says how the last 30 days went, or how it went since it
// opened when it is younger, and asks for the statement when its value is old.
func TestFundRowOfTheDigest(t *testing.T) {
	now := time.Date(2026, time.October, 5, 9, 0, 0, 0, time.UTC)
	valued := time.Date(2026, time.September, 30, 0, 0, 0, 0, time.UTC)
	thirty, inception := "-0.4577", "0.5383"

	fund := portfolio.Fund{Name: "Bolsillo", Currency: money.COP, Value: "13049999.99951645", ValuedOn: &valued}
	perf := portfolio.FundPerformance{Periods: []portfolio.FundPeriod{
		{Key: "30d", Pct: &thirty},
		{Key: "inception", Pct: &inception},
	}}

	row := fundRow(fund, perf, now)

	if row.Value != "13050000.00" || row.ValuedOn != "30 sep" || row.Stale {
		t.Errorf("row = %+v", row)
	}

	if row.ReturnPct != "-0.46" || row.ReturnLabel != "30 días" || row.ReturnColor != lossColor {
		t.Errorf("return = %s %s %s, want the 30 days, as a loss", row.ReturnPct, row.ReturnLabel, row.ReturnColor)
	}

	// Younger than a month: since it opened.
	perf.Periods[0].Pct = nil

	if row := fundRow(fund, perf, now); row.ReturnPct != "+0.54" || row.ReturnLabel != "desde el inicio" {
		t.Errorf("young fund = %s %s", row.ReturnPct, row.ReturnLabel)
	}

	// Forty days without a statement, and a fund never valued, are both stale.
	if row := fundRow(fund, perf, valued.AddDate(0, 0, 40)); !row.Stale {
		t.Error("a 40-day-old value is not flagged")
	}

	fund.ValuedOn = nil

	if row := fundRow(fund, perf, now); !row.Stale || row.ValuedOn != "" {
		t.Errorf("a fund with no value = %+v", row)
	}
}
