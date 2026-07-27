package portfolio

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"
)

// rate is a shorthand for the decimal values these tables hold.
func rate(t *testing.T, v string) money.Decimal {
	t.Helper()

	d, err := decimal.NewFromString(v)
	if err != nil {
		t.Fatalf("decimal %q: %v", v, err)
	}

	return d
}

func TestGetConversionRate(t *testing.T) {
	userID := uuid.New()

	t.Run("the user's own rate wins over the shared one", func(t *testing.T) {
		repo := new(fakeRepository{
			getUserExchangeRateByPair: func(_ context.Context, _ uuid.UUID, from, to string) (money.Decimal, error) {
				if from == "USD" && to == "COP" {
					return rate(t, "4100"), nil
				}

				return money.Decimal{}, ErrExchangeRateNotFound
			},
			getExchangeRateByPair: func(context.Context, string, string) (money.Decimal, error) {
				return rate(t, "3900"), nil
			},
		})

		got, err := newTestServices(repo, newMemStorage()).GetConversionRate(context.Background(), userID, "USD", "COP")
		if err != nil {
			t.Fatalf("GetConversionRate: %v", err)
		}

		// Not 3900: serving another user's fetched rate is the redistribution
		// the BYO-key model exists to avoid.
		if got.String() != "4100" {
			t.Errorf("rate = %s, want 4100 (the user's own)", got)
		}
	})

	t.Run("it falls back to the shared, admin-entered rate", func(t *testing.T) {
		repo := new(fakeRepository{
			getExchangeRateByPair: func(_ context.Context, from, to string) (money.Decimal, error) {
				if from == "USD" && to == "COP" {
					return rate(t, "3900"), nil
				}

				return money.Decimal{}, ErrExchangeRateNotFound
			},
		})

		got, err := newTestServices(repo, newMemStorage()).GetConversionRate(context.Background(), userID, "USD", "COP")
		if err != nil {
			t.Fatalf("GetConversionRate: %v", err)
		}
		if got.String() != "3900" {
			t.Errorf("rate = %s, want 3900", got)
		}
	})

	t.Run("the opposite direction is inverted", func(t *testing.T) {
		repo := new(fakeRepository{
			getExchangeRateByPair: func(_ context.Context, from, to string) (money.Decimal, error) {
				if from == "USD" && to == "COP" {
					return rate(t, "4000"), nil
				}

				return money.Decimal{}, ErrExchangeRateNotFound
			},
		})

		got, err := newTestServices(repo, newMemStorage()).GetConversionRate(context.Background(), userID, "COP", "USD")
		if err != nil {
			t.Fatalf("GetConversionRate: %v", err)
		}

		// Only one direction is ever stored, which is why the sync asks for one
		// direction per pair.
		want, err := decimal.One.Div(rate(t, "4000"))
		if err != nil {
			t.Fatalf("invert: %v", err)
		}
		if got.String() != want.String() {
			t.Errorf("rate = %s, want %s (1/4000)", got, want)
		}
	})

	t.Run("an unrelated pair is resolved through USD", func(t *testing.T) {
		repo := new(fakeRepository{
			getExchangeRateByPair: func(_ context.Context, from, to string) (money.Decimal, error) {
				switch {
				case from == "EUR" && to == "USD":
					return rate(t, "1.1"), nil
				case from == "USD" && to == "COP":
					return rate(t, "4000"), nil
				}

				return money.Decimal{}, ErrExchangeRateNotFound
			},
		})

		got, err := newTestServices(repo, newMemStorage()).GetConversionRate(context.Background(), userID, "EUR", "COP")
		if err != nil {
			t.Fatalf("GetConversionRate: %v", err)
		}
		if want := rate(t, "1.1").Mul(rate(t, "4000")); got.String() != want.String() {
			t.Errorf("rate = %s, want %s (1.1 * 4000)", got, want)
		}
	})

	t.Run("the same currency needs no rate at all", func(t *testing.T) {
		got, err := newTestServices(new(fakeRepository{}), newMemStorage()).GetConversionRate(context.Background(), userID, "USD", "USD")
		if err != nil {
			t.Fatalf("GetConversionRate: %v", err)
		}
		if got.String() != "1" {
			t.Errorf("rate = %s, want 1", got)
		}
	})

	// This is the state a user is in right after the BYO-key migration cleared
	// the shared table and before their first sync. It has to be a clean 404,
	// not a wrong number.
	t.Run("no rate anywhere is reported, not guessed", func(t *testing.T) {
		repo := new(fakeRepository{
			getExchangeRateByPair: func(context.Context, string, string) (money.Decimal, error) {
				return money.Decimal{}, ErrExchangeRateNotFound
			},
		})

		_, err := newTestServices(repo, newMemStorage()).GetConversionRate(context.Background(), userID, "USD", "COP")
		if err == nil {
			t.Fatal("GetConversionRate = nil error, want ErrExchangeRateUnavailable")
		}
	})
}
