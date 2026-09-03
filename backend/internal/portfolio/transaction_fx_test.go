package portfolio

import (
	"errors"
	"testing"

	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"
)

func mustEUR(t *testing.T, s string) money.Money {
	t.Helper()
	cur, err := money.GetCurrencyFromISOCode("EUR")
	if err != nil {
		t.Fatalf("GetCurrencyFromISOCode: %v", err)
	}
	m, err := money.NewMoneyFromString(s, cur)
	if err != nil {
		t.Fatalf("NewMoneyFromString(%q): %v", s, err)
	}

	return m
}

func TestTransactionInputRate(t *testing.T) {
	t.Run("an omitted rate is one, not zero", func(t *testing.T) {
		var in TransactionInput
		if got := in.Rate(); !got.Equal(decimal.One) {
			t.Errorf("Rate() = %s, want 1", got)
		}
	})

	t.Run("a stated rate is returned as stated", func(t *testing.T) {
		in := TransactionInput{FXRate: mustDecimal(t, "1.0638")}
		if got := in.Rate(); got.String() != "1.0638" {
			t.Errorf("Rate() = %s, want 1.0638", got)
		}
	})
}

func TestTransactionInputValidate(t *testing.T) {
	t.Run("same currency with no rate is the ordinary case", func(t *testing.T) {
		in := TransactionInput{Currency: money.USD, Price: mustUSD(t, "150.00")}

		currency, err := in.Validate(money.USD)
		if err != nil {
			t.Fatalf("Validate: %v", err)
		}
		if currency != money.USD {
			t.Errorf("currency = %q, want USD", currency)
		}
	})

	// A client that predates the field sends no currency at all. Falling back to
	// the position's is what keeps those requests meaning what they used to.
	t.Run("an unset currency falls back to the position's", func(t *testing.T) {
		in := TransactionInput{Price: mustUSD(t, "150.00")}

		currency, err := in.Validate(money.EUR)
		if err != nil {
			t.Fatalf("Validate: %v", err)
		}
		if currency != money.EUR {
			t.Errorf("currency = %q, want EUR", currency)
		}
	})

	t.Run("a cross-currency trade needs its rate", func(t *testing.T) {
		in := TransactionInput{Currency: money.EUR, Price: mustEUR(t, "606.60")}

		if _, err := in.Validate(money.USD); !errors.Is(err, ErrTransactionFXRate) {
			t.Fatalf("err = %v, want ErrTransactionFXRate", err)
		}
	})

	t.Run("a cross-currency trade with a rate is accepted", func(t *testing.T) {
		in := TransactionInput{
			Currency: money.EUR,
			Price:    mustEUR(t, "606.60"),
			FXRate:   mustDecimal(t, "1.0638"),
		}

		currency, err := in.Validate(money.USD)
		if err != nil {
			t.Fatalf("Validate: %v", err)
		}
		if currency != money.EUR {
			t.Errorf("currency = %q, want EUR", currency)
		}
	})

	// The refusal that matters most: nothing downstream could detect this, and
	// it scales the position's whole cost basis by whatever was typed.
	t.Run("a currency does not convert into itself at anything but one", func(t *testing.T) {
		in := TransactionInput{
			Currency: money.USD,
			Price:    mustUSD(t, "150.00"),
			FXRate:   mustDecimal(t, "1.0638"),
		}

		if _, err := in.Validate(money.USD); !errors.Is(err, ErrTransactionFXRate) {
			t.Fatalf("err = %v, want ErrTransactionFXRate", err)
		}
	})

	t.Run("a non-positive rate is refused", func(t *testing.T) {
		in := TransactionInput{
			Currency: money.EUR,
			Price:    mustEUR(t, "606.60"),
			FXRate:   mustDecimal(t, "-1.0638"),
		}

		if _, err := in.Validate(money.USD); !errors.Is(err, ErrTransactionFXRate) {
			t.Fatalf("err = %v, want ErrTransactionFXRate", err)
		}
	})
}

// The arithmetic the whole column exists for, on the position that prompted it:
// 0.0241 shares of LVMH filled at 606.60 EUR while the account was funded in
// USD and the broker converted at 1.0638.
//
// The broker's own confirmation says the open value was 15.55 USD. What is
// being checked is that price × rate reproduces the per-unit cost the average
// cost trigger will store, and that quantity × that is the amount the account
// was actually debited — because if this multiplication is off, every figure
// derived from the position is off by the same factor and none of them look
// wrong.
func TestCrossCurrencyCostBasis(t *testing.T) {
	quantity := mustDecimal(t, "0.0241")
	price := mustEUR(t, "606.60")
	rate := mustDecimal(t, "1.0638")

	costPerUnit := price.MulDecimal(rate)
	if got := costPerUnit.RoundBank(5).StringFixed(5); got != "645.30108" {
		t.Errorf("cost per unit = %s, want 645.30108", got)
	}

	total := costPerUnit.MulDecimal(quantity)
	if got := total.RoundBank(2).StringFixed(2); got != "15.55" {
		t.Errorf("open value = %s, want 15.55 (the broker's figure)", got)
	}

	// Recording the same trade in EUR and letting the reader translate it at
	// today's rate (1.1565) is what the app did before: the cost basis comes out
	// 1.34 USD higher, and the loss on screen is overstated by exactly that.
	today := mustDecimal(t, "1.1565")
	retranslated := price.MulDecimal(quantity).MulDecimal(today)
	if got := retranslated.RoundBank(2).StringFixed(2); got != "16.91" {
		t.Errorf("re-translated cost = %s, want 16.91", got)
	}
}

// A transaction read from a query that does not select fx_rate must not reach a
// client as a rate of zero: multiplying by it erases the price.
func TestTransactionResponseDefaultsRateToOne(t *testing.T) {
	dto := NewTransactionResponse(Transaction{Price: mustUSD(t, "150.00"), Currency: money.USD})
	if dto.FXRate != "1" {
		t.Errorf("fxRate = %q, want \"1\"", dto.FXRate)
	}
	if dto.CostCurrency != "" {
		t.Errorf("costCurrency = %q, want empty when the entry was not joined", dto.CostCurrency)
	}

	dto = NewTransactionResponse(Transaction{
		Price:        mustEUR(t, "606.60"),
		Currency:     money.EUR,
		FXRate:       mustDecimal(t, "1.0638"),
		CostCurrency: money.USD,
	})
	if dto.FXRate != "1.0638" || dto.CostCurrency != "USD" {
		t.Errorf("fxRate/costCurrency = %q/%q, want 1.0638/USD", dto.FXRate, dto.CostCurrency)
	}
}
