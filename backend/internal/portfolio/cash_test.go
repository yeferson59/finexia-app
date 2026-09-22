package portfolio

import (
	"errors"
	"strings"
	"testing"
	"time"

	"uuid"

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
		{name: "a dividend, which only its dividend writes", mutate: func(in *CashMovementInput) { in.Kind = CashKindDividend }, wantErr: "kind must be"},
		{name: "a sale, which only its sale writes", mutate: func(in *CashMovementInput) { in.Kind = CashKindSale }, wantErr: "kind must be"},
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
	for _, kind := range []CashMovementKind{CashKindDeposit, CashKindWithdrawal, CashKindInterest, CashKindDividend, CashKindSale} {
		if got := cashKindOf(kind.TransactionType()); got != kind {
			t.Errorf("cashKindOf(%s) = %s, want %s back", kind.TransactionType(), got, kind)
		}
	}

	// A balance recorded by hand before these screens existed was a buy and a
	// sell of its cash asset; those read as what they did to it. A paid-out
	// interest did nothing to it and is none of the kinds; neither is a dividend
	// recorded on the balance itself, which is not the credit of one.
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
		CashDividend: "40",
		CashSale:     "40",
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

// A cash row is written only with the transaction whose money it is, so the
// generic writers refuse the types outright; requireTypeAllowed checks it
// before the asset.
func TestCashLinksAreValidTypesNoOneWritesByHand(t *testing.T) {
	for settled, want := range map[TransactionType]struct {
		link TransactionType
		kind CashMovementKind
		// effect is which way the row moves the balance: a dividend and a sale
		// pay in, a purchase pays out.
		effect string
	}{
		Dividend: {CashDividend, CashKindDividend, "1"},
		Sell:     {CashSale, CashKindSale, "1"},
		Buy:      {CashPurchase, CashKindPurchase, "-1"},
	} {
		link, ok := settled.cashLink()
		if !ok || link != want.link {
			t.Errorf("%s.cashLink() = %s, %v; want %s", settled, link, ok, want.link)
		}
		if !link.IsValid() || !link.isCashLinked() {
			t.Errorf("%s is not a valid cash row, so it could not be read back", link)
		}
		if want.kind.IsValid() {
			t.Errorf("the cash screens can write a %s", want.kind)
		}
		if got := want.kind.TransactionType(); got != link {
			t.Errorf("%s.TransactionType() = %s, want %s", want.kind, got, link)
		}
		if got := cashKindOf(link); got != want.kind {
			t.Errorf("cashKindOf(%s) = %s, want %s", link, got, want.kind)
		}
		if got := balanceEffect(link, decimal.One).String(); got != want.effect {
			t.Errorf("balanceEffect(%s, 1) = %s, want %s", link, got, want.effect)
		}
	}

	for _, txnType := range []TransactionType{TransferIn, TransferOut, Fee, Interest, Split, CashInterest} {
		if _, ok := txnType.cashLink(); ok {
			t.Errorf("%s settles against cash", txnType)
		}
		if txnType.isCashLinked() {
			t.Errorf("%s reads as the cash side of another transaction", txnType)
		}
	}
}

// The two answers about cash each belong to one type, and Validate is where
// that is said once for every writer. A buy does not pay into cash and a
// dividend does not come out of it.
func TestCashSideBelongsToItsTransactionType(t *testing.T) {
	for _, tc := range []struct {
		name    string
		in      TransactionInput
		wantErr error
	}{
		{"a purchase paid from cash", TransactionInput{Type: Buy, PayFromCash: true}, nil},
		{"a transfer paid from cash", TransactionInput{Type: TransferIn, PayFromCash: true}, ErrNotPayableFromCash},
		{"a dividend paid from cash", TransactionInput{Type: Dividend, PayFromCash: true}, ErrNotPayableFromCash},
		{"a purchase credited to cash", TransactionInput{Type: Buy, CreditCash: true}, ErrNotCreditable},
		{"a dividend credited to cash", TransactionInput{Type: Dividend, CreditCash: true}, nil},
		{"a sale credited to cash", TransactionInput{Type: Sell, CreditCash: true}, nil},
		{"a fee credited to cash", TransactionInput{Type: Fee, CreditCash: true}, ErrNotCreditable},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.in.Validate(money.USD)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("Validate() error = %v, want %v", err, tc.wantErr)
			}
		})
	}
}

// A move states where the money lands as well as where it leaves, and the two
// currencies are what decides whether a rate belongs in it.
func TestCashMoveInputValidate(t *testing.T) {
	bank := uuid.New()
	broker := uuid.New()
	pocket := uuid.New()

	base := CashMoveInput{
		Currency: money.COP,
		To:       pocket,
		Amount:   mustDecimal(t, "400000"),
		Date:     time.Date(2026, 9, 21, 0, 0, 0, 0, time.UTC),
	}

	for _, tc := range []struct {
		name    string
		mutate  func(in *CashMoveInput)
		wantErr string
	}{
		{name: "between two drawers of one account"},
		{name: "to another platform", mutate: func(in *CashMoveInput) {
			in.ToSource, in.To = broker, uuid.UUID{}
		}},
		{name: "to another platform and another currency", mutate: func(in *CashMoveInput) {
			in.ToSource, in.To = broker, uuid.UUID{}
			in.ToCurrency, in.ToAmount = money.USD, mustDecimal(t, "98.50")
		}},
		// The main account of another platform is not the main account of this
		// one, so two zero drawers are two different places.
		{name: "to the main account of another platform", mutate: func(in *CashMoveInput) {
			in.ToSource, in.From, in.To = broker, uuid.UUID{}, uuid.UUID{}
		}},
		{name: "to where it already is", mutate: func(in *CashMoveInput) {
			in.From, in.To = pocket, pocket
		}, wantErr: "moving money to where it already is"},
		{name: "arriving at something else in the same currency", mutate: func(in *CashMoveInput) {
			in.ToAmount = mustDecimal(t, "399000")
		}, wantErr: "arrives at what it left"},
		{name: "crossing currencies without saying what arrived", mutate: func(in *CashMoveInput) {
			in.ToSource, in.To = broker, uuid.UUID{}
			in.ToCurrency, in.ToAmount = money.USD, decimal.Decimal{}
		}, wantErr: "how much USD arrived"},
		{name: "arriving in a currency nobody keeps", mutate: func(in *CashMoveInput) {
			in.ToCurrency = money.DKK
		}, wantErr: "the currency it arrives in"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in := base
			if tc.mutate != nil {
				tc.mutate(&in)
			}

			// withDefaults is what every caller applies first: a move that says
			// nothing about its destination means the account it starts in.
			err := in.withDefaults(bank).Validate(bank)

			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate() = %v, want nil", err)
				}
				return
			}

			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("Validate() = %v, want an error about %q", err, tc.wantErr)
			}
			if !errors.Is(err, ErrInvalidCashMove) {
				t.Errorf("Validate() = %v, want ErrInvalidCashMove", err)
			}
		})
	}
}

// The two legs are the two amounts the move states: what left the savings app
// and what reached the broker, each in its own currency and neither derived
// from the other.
func TestCashMoveLegsConvert(t *testing.T) {
	bank := uuid.New()

	in := CashMoveInput{
		Currency: money.COP,
		Amount:   mustDecimal(t, "400000"),
		Date:     time.Date(2026, 9, 21, 0, 0, 0, 0, time.UTC),
		Notes:    "para comprar AAPL",
	}
	in.ToSource, in.ToCurrency, in.ToAmount = uuid.New(), money.USD, mustDecimal(t, "98.50")
	in = in.withDefaults(bank)

	out, into := in.legs()

	if out.Kind != CashKindWithdrawal || into.Kind != CashKindDeposit {
		t.Fatalf("legs = %s and %s, want a withdrawal and a deposit", out.Kind, into.Kind)
	}
	if out.Currency != money.COP || out.Amount.String() != "400000" {
		t.Errorf("what left = %s %s, want 400000 COP", out.Amount, out.Currency)
	}
	// To the cent the statement says, not to what a rate would have computed.
	// The trailing zero goes in the rounding to eight decimals; the value does
	// not change.
	if into.Currency != money.USD || into.Amount.String() != "98.5" {
		t.Errorf("what arrived = %s %s, want 98.50 USD", into.Amount, into.Currency)
	}
	// Neither leg carries a fee: one would stop the two cancelling out and the
	// money would read as a loss.
	if out.Fees.IsPos() || into.Fees.IsPos() {
		t.Errorf("legs carry fees %s and %s, want none", out.Fees, into.Fees)
	}
	if into.Notes != in.Notes || !into.Date.Equal(out.Date) {
		t.Errorf("the legs disagree on the note or the day")
	}
}
