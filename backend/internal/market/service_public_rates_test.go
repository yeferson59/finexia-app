package market

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/platform/logger"
	"github.com/yeferson59/finexia-app/internal/platform/marketdata"
)

// storedRate is one row the fake repository accepted.
type storedRate struct {
	from, to money.Currency
	rate     decimal.Decimal
	rateDate time.Time
	source   ProviderID
}

// newPublicRatesFixture wires the service against a recording repository and a
// canned feed, and returns both so a test can assert on what was written.
func newPublicRatesFixture(source marketdata.PublicRateSource) (*Service, *[]storedRate) {
	var written []storedRate

	repo := new(fakeRepository{
		upsertPublicExchangeRate: func(_ context.Context, from, to money.Currency, rate decimal.Decimal, rateDate time.Time, src ProviderID) (ExchangeRate, error) {
			written = append(written, storedRate{from, to, rate, rateDate, src})

			return ExchangeRate{FromCurrency: from, ToCurrency: to, Rate: rate, RateDate: rateDate, Source: src}, nil
		},
	})

	return newService(repo, newMemStorage(), nil, source, testKeyring(), logger.Noop()), &written
}

func TestRefreshPublicRates(t *testing.T) {
	asOf := time.Date(2026, 8, 9, 18, 1, 11, 0, time.UTC)

	t.Run("stores what the feed published", func(t *testing.T) {
		feed := new(fakePublicRateSource{rates: []marketdata.PublicRate{{
			From: money.USD, To: money.COP, Rate: "3157.43", Source: marketdata.DolarAPI, AsOf: asOf,
		}}})
		svc, written := newPublicRatesFixture(feed)

		stored, err := svc.RefreshPublicRates(context.Background())
		if err != nil {
			t.Fatalf("RefreshPublicRates: %v", err)
		}
		if len(stored) != 1 || len(*written) != 1 {
			t.Fatalf("stored %d rates and wrote %d rows, want 1 of each", len(stored), len(*written))
		}

		row := (*written)[0]
		if row.from != money.USD || row.to != money.COP {
			t.Errorf("pair = %s/%s, want USD/COP", row.from, row.to)
		}
		if row.rate.String() != "3157.43" {
			t.Errorf("rate = %s, want 3157.43", row.rate)
		}
		// The feed's own timestamp dates the row, so a feed that stopped
		// updating is visibly stale instead of being restamped as today's.
		if !row.rateDate.Equal(asOf) {
			t.Errorf("rateDate = %s, want the feed's timestamp %s", row.rateDate, asOf)
		}
		if row.source != marketdata.DolarAPI {
			t.Errorf("source = %q, want dolarapi", row.source)
		}
	})

	t.Run("dates a rate with no timestamp as today", func(t *testing.T) {
		feed := new(fakePublicRateSource{rates: []marketdata.PublicRate{{
			From: money.USD, To: money.COP, Rate: "3157.43", Source: marketdata.DolarAPI,
		}}})
		svc, written := newPublicRatesFixture(feed)

		if _, err := svc.RefreshPublicRates(context.Background()); err != nil {
			t.Fatalf("RefreshPublicRates: %v", err)
		}
		if (*written)[0].rateDate.IsZero() {
			t.Error("rateDate is zero; a feed with no timestamp must date the row now")
		}
	})

	// A feed is not more trusted than an operator: everything it publishes goes
	// through the same ISO 4217 table and positive-rate rule the admin
	// endpoints apply, because this is what portfolio then converts with.
	t.Run("rejects rates no conversion could apply", func(t *testing.T) {
		cases := []struct {
			name string
			rate marketdata.PublicRate
		}{
			{"unknown currency", marketdata.PublicRate{From: money.USD, To: money.Currency(255), Rate: "1", Source: marketdata.DolarAPI}},
			{"not a number", marketdata.PublicRate{From: money.USD, To: money.COP, Rate: "n/a", Source: marketdata.DolarAPI}},
			{"zero", marketdata.PublicRate{From: money.USD, To: money.COP, Rate: "0", Source: marketdata.DolarAPI}},
			{"negative", marketdata.PublicRate{From: money.USD, To: money.COP, Rate: "-3157.43", Source: marketdata.DolarAPI}},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				svc, written := newPublicRatesFixture(new(fakePublicRateSource{rates: []marketdata.PublicRate{tc.rate}}))

				if _, err := svc.RefreshPublicRates(context.Background()); err == nil {
					t.Error("expected an error")
				}
				if len(*written) != 0 {
					t.Errorf("wrote %+v; nothing should reach the table", *written)
				}
			})
		}
	})

	// One unusable pair must not cost the others their refresh: the rates are
	// independent, and dropping a good USD/COP because some other pair was
	// malformed would leave the dashboard without the number it needs.
	t.Run("stores the good pairs and reports the bad ones", func(t *testing.T) {
		feed := new(fakePublicRateSource{rates: []marketdata.PublicRate{
			{From: money.USD, To: money.Currency(255), Rate: "1", Source: marketdata.DolarAPI},
			{From: money.USD, To: money.COP, Rate: "3157.43", Source: marketdata.DolarAPI, AsOf: asOf},
		}})
		svc, written := newPublicRatesFixture(feed)

		stored, err := svc.RefreshPublicRates(context.Background())
		if err == nil {
			t.Error("expected the bad pair to be reported")
		}
		if len(stored) != 1 || len(*written) != 1 {
			t.Fatalf("stored %d rates and wrote %d rows, want the one good pair", len(stored), len(*written))
		}
		if row := (*written)[0]; row.to != money.COP {
			t.Errorf("wrote %s/%s, want the valid USD/COP", row.from, row.to)
		}
	})

	// A feed that is down leaves the stored rates alone. Stale beats absent for
	// a rate: a portfolio valued at yesterday's TRM is still valued.
	t.Run("writes nothing when the feed fails", func(t *testing.T) {
		feed := new(fakePublicRateSource{err: errors.New("dolarapi: TRM: status 503")})
		svc, written := newPublicRatesFixture(feed)

		if _, err := svc.RefreshPublicRates(context.Background()); err == nil {
			t.Fatal("expected the feed's error to surface")
		}
		if len(*written) != 0 {
			t.Errorf("wrote %+v after a failed fetch", *written)
		}
	})

	// The source is several feeds read as one, so a fetch can come back as
	// rates *and* an error. The pairs one feed owns must not be dropped because
	// another feed is down — that is the whole reason they are read together.
	t.Run("stores what a partial fetch published", func(t *testing.T) {
		feed := new(fakePublicRateSource{
			rates: []marketdata.PublicRate{{From: money.USD, To: money.COP, Rate: "3157.43", Source: marketdata.DolarAPI, AsOf: asOf}},
			err:   errors.New("ecb: reference rates: status 503"),
		})
		svc, written := newPublicRatesFixture(feed)

		stored, err := svc.RefreshPublicRates(context.Background())
		if err == nil {
			t.Error("expected the failed feed to be reported")
		}
		if len(stored) != 1 || len(*written) != 1 {
			t.Fatalf("stored %d rates and wrote %d rows, want the pair the working feed published", len(stored), len(*written))
		}
	})

	t.Run("reports a deployment with no feed wired", func(t *testing.T) {
		svc, _ := newPublicRatesFixture(nil)

		if _, err := svc.RefreshPublicRates(context.Background()); !errors.Is(err, ErrPublicRatesUnavailable) {
			t.Errorf("error = %v, want ErrPublicRatesUnavailable", err)
		}
	})
}

func TestGetLatestExchangeRates(t *testing.T) {
	var gotOffset, gotLimit uint

	repo := new(fakeRepository{
		getExchangeRates: func(_ context.Context, offset, limit uint) ([]ExchangeRate, error) {
			gotOffset, gotLimit = offset, limit

			return []ExchangeRate{{FromCurrency: money.USD, ToCurrency: money.COP, Source: marketdata.DolarAPI}}, nil
		},
	})
	svc := newService(repo, newMemStorage(), nil, nil, testKeyring(), logger.Noop())

	rates, err := svc.GetLatestExchangeRates(context.Background())
	if err != nil {
		t.Fatalf("GetLatestExchangeRates: %v", err)
	}
	if len(rates) != 1 {
		t.Fatalf("got %d rates, want 1", len(rates))
	}
	// Unpaginated for the caller, still bounded at the repository: the read is
	// "every shared rate", and the limit is only there to cap a table that grew
	// unexpectedly.
	if gotOffset != 0 || gotLimit != maxSharedRates {
		t.Errorf("read offset=%d limit=%d, want 0/%d", gotOffset, gotLimit, maxSharedRates)
	}
}

// The refresh job is the scheduler's view of the same call: it must surface a
// failure so the runner retries (a keyless fetch costs nobody's quota), and
// stay quiet when the refresh worked.
func TestPublicRatesJob(t *testing.T) {
	t.Run("succeeds when the feed answers", func(t *testing.T) {
		feed := new(fakePublicRateSource{rates: []marketdata.PublicRate{{
			From: money.USD, To: money.COP, Rate: "3157.43", Source: marketdata.DolarAPI,
		}}})
		svc, _ := newPublicRatesFixture(feed)

		if err := NewPublicRatesJob(svc, logger.Noop()).Run(context.Background()); err != nil {
			t.Errorf("Run: %v", err)
		}
		if feed.calls != 1 {
			t.Errorf("fetched %d times, want 1", feed.calls)
		}
	})

	t.Run("reports a failed refresh", func(t *testing.T) {
		svc, _ := newPublicRatesFixture(new(fakePublicRateSource{err: errors.New("boom")}))

		if err := NewPublicRatesJob(svc, logger.Noop()).Run(context.Background()); err == nil {
			t.Error("expected the job to report the failure so the runner retries")
		}
	})
}
