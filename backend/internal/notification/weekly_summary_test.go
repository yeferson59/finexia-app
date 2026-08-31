package notification

import (
	"context"
	"errors"
	"testing"
	"time"

	"uuid"

	"github.com/yeferson59/finexia-app/internal/identity"
	"github.com/yeferson59/finexia-app/internal/platform/mail"
	"github.com/yeferson59/finexia-app/internal/portfolio"
	"github.com/yeferson59/gofinance/v2/money"
)

// fakeUserReader stubs the users the weekly summary iterates over.
type fakeUserReader struct {
	getUsers func(ctx context.Context) ([]identity.User, error)
}

func (f *fakeUserReader) GetUsersWithWeeklySummary(ctx context.Context) ([]identity.User, error) {
	return f.getUsers(ctx)
}

// fakePortfolioReader stubs the per-user portfolio summaries and the past
// total the weekly change is measured against.
type fakePortfolioReader struct {
	getSummary func(ctx context.Context, userID uuid.UUID) ([]portfolio.SummaryView, error)
	// getPortfolioValuesAsOf is optional: left unset, the account reads as
	// having no history, which is what an unstubbed test wants.
	getPortfolioValuesAsOf func(ctx context.Context, userID uuid.UUID, asOf time.Time) ([]portfolio.PortfolioValuePoint, error)
}

func (f *fakePortfolioReader) GetPortfoliosSummary(ctx context.Context, userID uuid.UUID) ([]portfolio.SummaryView, error) {
	return f.getSummary(ctx, userID)
}

func (f *fakePortfolioReader) GetPortfolioValuesAsOf(ctx context.Context, userID uuid.UUID, asOf time.Time) ([]portfolio.PortfolioValuePoint, error) {
	if f.getPortfolioValuesAsOf == nil {
		return nil, nil
	}
	return f.getPortfolioValuesAsOf(ctx, userID, asOf)
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
	return NewService(users, ports, mailer, Config{PublicURL: "https://finexia.test"})
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

	// send runs the digest with a stubbed baseline and returns the email data.
	send := func(t *testing.T, summaries []portfolio.SummaryView, baseline func(context.Context, uuid.UUID, time.Time) ([]portfolio.PortfolioValuePoint, error)) mail.WeeklySummaryData {
		t.Helper()
		users := new(fakeUserReader{getUsers: func(context.Context) ([]identity.User, error) {
			return []identity.User{{ID: userID, Name: "Ada", Email: "ada@example.com"}}, nil
		}})
		ports := new(fakePortfolioReader{
			getSummary: func(context.Context, uuid.UUID) ([]portfolio.SummaryView, error) {
				return summaries, nil
			},
			getPortfolioValuesAsOf: baseline,
		})
		mailer := new(fakeMailer{})
		sent, errs := newTestService(users, ports, mailer).SendWeeklySummaryEmails(context.Background())
		if sent != 1 || len(errs) != 0 {
			t.Fatalf("sent/errs = %d/%v, want 1/none", sent, errs)
		}
		return mailer.weekly[0].Data
	}

	// at builds a baseline stub from portfolio/value pairs, all dated the same
	// day, which is how SyncPortfolioSnapshots writes them.
	at := func(values map[uuid.UUID]string) func(context.Context, uuid.UUID, time.Time) ([]portfolio.PortfolioValuePoint, error) {
		return func(context.Context, uuid.UUID, time.Time) ([]portfolio.PortfolioValuePoint, error) {
			points := make([]portfolio.PortfolioValuePoint, 0, len(values))
			for id, v := range values {
				points = append(points, portfolio.PortfolioValuePoint{PortfolioID: id, Date: lastMonday, TotalValue: v})
			}
			return points, nil
		}
	}

	twoPortfolios := []portfolio.SummaryView{
		{ID: stocksID, Name: "Acciones", BaseCurrency: money.USD, TotalMarketValue: "1100.00", TotalGainLoss: "50.00", TotalGainLossPct: "4.76"},
		{ID: cryptoID, Name: "Cripto", BaseCurrency: money.USD, TotalMarketValue: "400.00", TotalGainLoss: "-20.00", TotalGainLossPct: "-4.76"},
	}

	onePortfolio := []portfolio.SummaryView{twoPortfolios[0]}

	t.Run("each portfolio reports its own movement", func(t *testing.T) {
		data := send(t, twoPortfolios, at(map[uuid.UUID]string{
			stocksID: "1000.00",
			cryptoID: "500.00",
		}))

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

	t.Run("the rows add up to the account total", func(t *testing.T) {
		// +100 on stocks and -100 on crypto net out: the headline has to say
		// so, or the reader can add up the rows and catch the digest lying.
		data := send(t, twoPortfolios, at(map[uuid.UUID]string{
			stocksID: "1000.00",
			cryptoID: "500.00",
		}))

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

	t.Run("a portfolio opened this week has no comparison but still counts in the total", func(t *testing.T) {
		// Only the stocks portfolio existed a week ago. Crypto's row shows no
		// change, and the account baseline is the stocks figure alone.
		data := send(t, twoPortfolios, at(map[uuid.UUID]string{stocksID: "1000.00"}))

		if !data.Portfolios[0].HasWeekChange {
			t.Error("the portfolio with history should carry a comparison")
		}
		if data.Portfolios[1].HasWeekChange {
			t.Error("a portfolio with no history should show no comparison")
		}
		// 1500 now against a 1000 baseline.
		if data.WeekChangeValue != "+500.00" || data.WeekChangePct != "+50.00" {
			t.Errorf("total change = %q/%q, want +500.00/+50.00", data.WeekChangeValue, data.WeekChangePct)
		}
	})

	t.Run("a week that gained reports the amount, the percentage and the date", func(t *testing.T) {
		data := send(t, onePortfolio, at(map[uuid.UUID]string{stocksID: "1000.00"}))

		if !data.HasWeekChange {
			t.Fatal("HasWeekChange = false, want the comparison to be shown")
		}
		if data.WeekChangeValue != "+100.00" || data.WeekChangePct != "+10.00" {
			t.Errorf("change = %q/%q, want +100.00/+10.00", data.WeekChangeValue, data.WeekChangePct)
		}
		if data.WeekChangeColor != gainColor {
			t.Errorf("WeekChangeColor = %q, want the gain color", data.WeekChangeColor)
		}
		if data.WeekChangeSince != "29 jul" {
			t.Errorf("WeekChangeSince = %q, want '29 jul'", data.WeekChangeSince)
		}
	})

	t.Run("a week that lost value reports a negative amount in the loss color", func(t *testing.T) {
		data := send(t, onePortfolio, at(map[uuid.UUID]string{stocksID: "1250.00"}))

		if data.WeekChangeValue != "-150.00" || data.WeekChangePct != "-12.00" {
			t.Errorf("change = %q/%q, want -150.00/-12.00", data.WeekChangeValue, data.WeekChangePct)
		}
		if data.WeekChangeColor != lossColor {
			t.Errorf("WeekChangeColor = %q, want the loss color", data.WeekChangeColor)
		}
	})

	t.Run("a flat week reports zero rather than hiding the comparison", func(t *testing.T) {
		data := send(t, onePortfolio, at(map[uuid.UUID]string{stocksID: "1100.00"}))

		if !data.HasWeekChange {
			t.Fatal("a week with no movement is still a real comparison")
		}
		if data.WeekChangeValue != "+0.00" || data.WeekChangePct != "+0.00" {
			t.Errorf("change = %q/%q, want +0.00/+0.00", data.WeekChangeValue, data.WeekChangePct)
		}
	})

	t.Run("an account with no history hides the comparison", func(t *testing.T) {
		data := send(t, onePortfolio, func(context.Context, uuid.UUID, time.Time) ([]portfolio.PortfolioValuePoint, error) {
			return nil, nil
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

	t.Run("a baseline lookup failure does not stop the digest", func(t *testing.T) {
		data := send(t, onePortfolio, func(context.Context, uuid.UUID, time.Time) ([]portfolio.PortfolioValuePoint, error) {
			return nil, errors.New("snapshots table on fire")
		})

		if data.HasWeekChange {
			t.Error("a failed lookup should leave the comparison out")
		}
		if data.TotalValue != "1100.00" {
			t.Errorf("TotalValue = %q, want the digest sent anyway", data.TotalValue)
		}
	})

	t.Run("starting from nothing shows the amount but no percentage", func(t *testing.T) {
		// The portfolio was worth nothing a week ago, so the move is not a
		// percentage of anything — returns.ROI refuses the division and the
		// amount stands on its own.
		data := send(t, onePortfolio, at(map[uuid.UUID]string{stocksID: "0"}))

		if !data.HasWeekChange || data.WeekChangeValue != "+1100.00" {
			t.Errorf("WeekChangeValue = %q, want +1100.00", data.WeekChangeValue)
		}
		if data.WeekChangePct != "" {
			t.Errorf("WeekChangePct = %q, want it empty when there is no base", data.WeekChangePct)
		}
		if data.Portfolios[0].WeekChangePct != "" {
			t.Errorf("row pct = %q, want it empty too", data.Portfolios[0].WeekChangePct)
		}
	})

	t.Run("the digest is dated by the most recent baseline snapshot", func(t *testing.T) {
		// One portfolio missed a day. The digest should say the newer date
		// rather than backdating everything to the stale one.
		older := lastMonday.AddDate(0, 0, -3)
		data := send(t, twoPortfolios, func(context.Context, uuid.UUID, time.Time) ([]portfolio.PortfolioValuePoint, error) {
			return []portfolio.PortfolioValuePoint{
				{PortfolioID: stocksID, Date: older, TotalValue: "1000.00"},
				{PortfolioID: cryptoID, Date: lastMonday, TotalValue: "500.00"},
			}, nil
		})

		if data.WeekChangeSince != "29 jul" {
			t.Errorf("WeekChangeSince = %q, want the most recent baseline date", data.WeekChangeSince)
		}
	})

	t.Run("the baseline is looked up a week back", func(t *testing.T) {
		var asked time.Time
		before := time.Now()
		send(t, onePortfolio, func(_ context.Context, uid uuid.UUID, asOf time.Time) ([]portfolio.PortfolioValuePoint, error) {
			if uid != userID {
				t.Errorf("userID = %s, want %s", uid, userID)
			}
			asked = asOf
			return []portfolio.PortfolioValuePoint{{PortfolioID: stocksID, Date: lastMonday, TotalValue: "1000.00"}}, nil
		})

		want := before.Add(-digestPeriod)
		if diff := asked.Sub(want); diff < -time.Minute || diff > time.Minute {
			t.Errorf("asOf = %v, want ~%v (a week back)", asked, want)
		}
	})

	t.Run("the comparison uses exact decimals, not float64", func(t *testing.T) {
		// 0.07 a hundred times is 7 exactly; in float64 it is 7.000000000000005,
		// which would report a change of -0.00 against a 7.00 baseline.
		summaries := make([]portfolio.SummaryView, 0, 100)
		baseline := map[uuid.UUID]string{}
		for range 100 {
			id := uuid.New()
			summaries = append(summaries, portfolio.SummaryView{
				ID: id, Name: "P", BaseCurrency: money.USD,
				TotalMarketValue: "0.07", TotalGainLoss: "0.00", TotalGainLossPct: "0.00",
			})
			baseline[id] = "0.07"
		}

		data := send(t, summaries, at(baseline))

		if data.WeekChangeValue != "+0.00" {
			t.Errorf("WeekChangeValue = %q, want +0.00", data.WeekChangeValue)
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
