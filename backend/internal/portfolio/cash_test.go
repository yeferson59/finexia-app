package portfolio

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/market"
	"github.com/yeferson59/finexia-app/internal/platform/currency"
)

func cashDeposit(t *testing.T) CashMovementInput {
	t.Helper()

	return CashMovementInput{
		Kind:     CashKindDeposit,
		Amount:   mustDecimal(t, "250.5"),
		Currency: money.USD,
		Date:     time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
	}
}

func TestCashMovementInputValidate(t *testing.T) {
	cases := []struct {
		name    string
		mutate  func(in *CashMovementInput)
		wantErr string
	}{
		{name: "a deposit", mutate: func(*CashMovementInput) {}},
		{name: "a deposit with fees", mutate: func(in *CashMovementInput) { in.Fees = mustDecimal(t, "2") }},
		{name: "a withdrawal whose fees take all of it", mutate: func(in *CashMovementInput) {
			in.Kind, in.Fees = CashKindWithdrawal, in.Amount
		}},
		{name: "interest", mutate: func(in *CashMovementInput) { in.Kind = CashKindInterest }},
		{name: "notes at the column limit, in multibyte runes", mutate: func(in *CashMovementInput) {
			in.Notes = strings.Repeat("ñ", maxCashNotesLen)
		}},
		{name: "an unknown kind", mutate: func(in *CashMovementInput) { in.Kind = "transfer" }, wantErr: "kind must be"},
		{name: "other, which is only ever read back", mutate: func(in *CashMovementInput) { in.Kind = CashKindOther }, wantErr: "kind must be"},
		{name: "no amount", mutate: func(in *CashMovementInput) { in.Amount = decimal.Zero }, wantErr: "greater than zero"},
		{name: "a negative amount", mutate: func(in *CashMovementInput) { in.Amount = mustDecimal(t, "-5") }, wantErr: "greater than zero"},
		{name: "negative fees", mutate: func(in *CashMovementInput) { in.Fees = mustDecimal(t, "-1") }, wantErr: "cannot be negative"},
		// cash_interest has no flow, so a fee on it would be subtracted from
		// nothing and disappear.
		{name: "interest with fees", mutate: func(in *CashMovementInput) {
			in.Kind, in.Fees = CashKindInterest, mustDecimal(t, "1")
		}, wantErr: "net of fees"},
		// Past the amount, the flow of a transfer_out turns into money put in.
		{name: "a withdrawal whose fees exceed it", mutate: func(in *CashMovementInput) {
			in.Kind, in.Fees = CashKindWithdrawal, mustDecimal(t, "250.51")
		}, wantErr: "cannot exceed its amount"},
		{name: "no date", mutate: func(in *CashMovementInput) { in.Date = time.Time{} }, wantErr: "date is required"},
		{name: "notes past the column limit", mutate: func(in *CashMovementInput) {
			in.Notes = strings.Repeat("ñ", maxCashNotesLen+1)
		}, wantErr: "notes cannot exceed"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := cashDeposit(t)
			tc.mutate(&in)

			err := in.Validate()
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate() = %v, want nil", err)
				}
				return
			}

			if !errors.Is(err, ErrInvalidCashMovement) {
				t.Fatalf("Validate() = %v, want ErrInvalidCashMovement", err)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("Validate() = %q, want it to mention %q", err, tc.wantErr)
			}
		})
	}
}

// A balance in a currency with no rate would sit in every total at face value,
// so a new one has to be in a currency the app can convert.
func TestCashMovementInputValidateNewNeedsAConvertibleCurrency(t *testing.T) {
	ars, err := money.GetCurrencyFromISOCode("ARS")
	if err != nil {
		t.Fatalf("GetCurrencyFromISOCode(ARS): %v", err)
	}

	for _, cur := range []money.Currency{money.XXX, ars} {
		in := cashDeposit(t)
		in.Currency = cur

		if err := in.ValidateNew(); !errors.Is(err, ErrInvalidCashMovement) {
			t.Errorf("ValidateNew() with %v = %v, want ErrInvalidCashMovement", cur, err)
		}
	}

	for _, cur := range currency.Supported {
		in := cashDeposit(t)
		in.Currency = cur

		if err := in.ValidateNew(); err != nil {
			t.Errorf("ValidateNew() with %v = %v, want nil", cur, err)
		}
	}
}

// The transaction a movement becomes has to pass the rule every transaction
// writer applies, or the repository would refuse what the service accepted.
func TestCashMovementBecomesAParTransactionInTheBalanceCurrency(t *testing.T) {
	wantType := map[CashMovementKind]TransactionType{
		CashKindDeposit:    TransferIn,
		CashKindWithdrawal: TransferOut,
		CashKindInterest:   CashInterest,
	}

	for kind, want := range wantType {
		in := cashDeposit(t)
		in.Kind = kind
		in.Notes = "  nómina  "
		if kind != CashKindInterest {
			in.Fees = mustDecimal(t, "1.25")
		}

		settled, err := in.transactionInput(money.COP).Validate(money.COP)
		if err != nil {
			t.Fatalf("%s: Validate(COP) = %v", kind, err)
		}

		if settled.Type != want {
			t.Errorf("%s: type = %s, want %s", kind, settled.Type, want)
		}
		if !settled.Quantity.Equal(in.Amount) {
			t.Errorf("%s: quantity = %s, want the amount %s", kind, settled.Quantity, in.Amount)
		}
		if !settled.Price.GetDecimal().Equal(decimal.One) || settled.Currency != money.COP {
			t.Errorf("%s: price = %s %v, want 1 COP", kind, settled.Price.GetDecimal(), settled.Currency)
		}
		if !settled.FXRate.Equal(decimal.One) || settled.FeesCurrency != money.COP {
			t.Errorf("%s: rate %s, fees in %v; want 1 and COP", kind, settled.FXRate, settled.FeesCurrency)
		}
		if !settled.Fees.GetDecimal().Equal(in.Fees) {
			t.Errorf("%s: fees = %s, want %s", kind, settled.Fees.GetDecimal(), in.Fees)
		}
		if settled.Notes != "nómina" {
			t.Errorf("%s: notes = %q, want them trimmed", kind, settled.Notes)
		}
	}
}

func TestCashKindOf(t *testing.T) {
	for _, kind := range []CashMovementKind{CashKindDeposit, CashKindWithdrawal, CashKindInterest} {
		if got := cashKindOf(kind.TransactionType()); got != kind {
			t.Errorf("cashKindOf(%s) = %s, want %s back", kind.TransactionType(), got, kind)
		}
	}

	// A balance recorded by hand before these screens existed was a buy and a
	// sell of its cash asset; those read as what they did to it. A paid-out
	// interest did nothing to it and is not one of the three.
	for txnType, want := range map[TransactionType]CashMovementKind{
		Buy:      CashKindDeposit,
		Sell:     CashKindWithdrawal,
		Interest: CashKindOther,
		Dividend: CashKindOther,
		Fee:      CashKindOther,
		Split:    CashKindOther,
	} {
		if got := cashKindOf(txnType); got != want {
			t.Errorf("cashKindOf(%s) = %s, want %s", txnType, got, want)
		}
	}
}

// balanceEffect is the Go copy of the quantity arms in recalculate_avg_cost; a
// disagreement would let a write approve a balance the trigger then clamps.
func TestBalanceEffectMirrorsTheTrigger(t *testing.T) {
	q := mustDecimal(t, "40")

	for txnType, want := range map[TransactionType]string{
		Buy:          "40",
		TransferIn:   "40",
		CashInterest: "40",
		Sell:         "-40",
		TransferOut:  "-40",
		Interest:     "0",
		Dividend:     "0",
		Fee:          "0",
		Split:        "0",
	} {
		if got := balanceEffect(txnType, q); !got.Equal(mustDecimal(t, want)) {
			t.Errorf("balanceEffect(%s, 40) = %s, want %s", txnType, got, want)
		}
	}
}

func TestCashInterestIsOnlyAllowedOnCash(t *testing.T) {
	if !CashInterest.IsValid() {
		t.Fatal("cash_interest is not a valid transaction type")
	}

	if !CashInterest.AllowedOn(market.Cash) {
		t.Error("cash_interest refused on a cash position")
	}

	for _, at := range []market.AssetType{market.Stock, market.ETF, market.Bond, market.Crypto, market.Other} {
		if CashInterest.AllowedOn(at) {
			t.Errorf("cash_interest allowed on a %s position", at)
		}
	}

	// Nothing that was accepted everywhere before is narrowed.
	for _, txnType := range []TransactionType{Buy, Sell, Dividend, Split, TransferIn, TransferOut, Fee, Interest} {
		if !txnType.AllowedOn(market.Stock) || !txnType.AllowedOn(market.Cash) {
			t.Errorf("%s is no longer allowed on every asset", txnType)
		}
	}
}

func TestCashAssetNaming(t *testing.T) {
	if got := cashTicker(money.USD); got != "CASH-USD" {
		t.Errorf("cashTicker(USD) = %q, want CASH-USD", got)
	}

	if got := cashAssetName(money.COP); got != "Efectivo en pesos colombianos (COP)" {
		t.Errorf("cashAssetName(COP) = %q", got)
	}

	// Every currency a balance can be opened in gets a name of its own, and a
	// ticker the catalog column can hold.
	for _, cur := range currency.Supported {
		if _, ok := cashCurrencyNames[cur]; !ok {
			t.Errorf("no name for %v", cur)
		}
		if n := len(cashTicker(cur)); n > maxTickerLenForCash {
			t.Errorf("cashTicker(%v) is %d characters, the column holds %d", cur, n, maxTickerLenForCash)
		}
	}
}

// maxTickerLenForCash mirrors assets.ticker VARCHAR(20).
const maxTickerLenForCash = 20
