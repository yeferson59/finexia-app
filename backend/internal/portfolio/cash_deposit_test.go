package portfolio

import (
	"errors"
	"strings"
	"testing"
	"time"

	"uuid"

	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"
)

// A fixed deposit (000048): what it earns from the day it opened, when a rate
// posted at maturity credits, and the rules a deposit and its cancellation are
// held to.

func sept(day int) time.Time {
	return time.Date(2026, time.September, day, 0, 0, 0, 0, time.UTC)
}

// A rate posted at maturity holds every day until the last one it earns, and
// pays them there. Without an end there is no day to pay on, and it holds them
// all — which is why only a deposit, whose end is written with it, is offered
// it.
func TestCreditsOnAtMaturity(t *testing.T) {
	last := sept(29)
	term := CashRateVersion{Posting: PostingAtMaturity, EffectiveFrom: sept(1), EndedOn: &last}

	for _, day := range []int{1, 15, 28} {
		if term.creditsOn(sept(day)) {
			t.Errorf("a deposit posted at maturity credits on %s, want it held", sept(day).Format(time.DateOnly))
		}
	}

	if !term.creditsOn(sept(29)) {
		t.Error("a deposit posted at maturity does not credit on its last day")
	}

	open := CashRateVersion{Posting: PostingAtMaturity, EffectiveFrom: sept(1)}
	if open.creditsOn(sept(29)) {
		t.Error("a version with no end credits at maturity; there is no maturity to credit on")
	}
}

// A deposit opened on the first and registered on the fifteenth owes fourteen
// days the moment it is recorded. Its balance is its own account, and its rate
// starts the day the money went in — which is what GetCashAccrualTargets reads
// from the pocket rather than from the day the row was written.
func TestPendingDaysOfADepositOpenedInThePast(t *testing.T) {
	matures := sept(29)
	term := CashRateVersion{ID: uuid.New(), EffectiveFrom: sept(1), EndedOn: &matures}

	deposit := CashAccrualTarget{OpenedOn: sept(1), Versions: []CashRateVersion{term}}

	days := deposit.PendingDays(sept(14))
	if len(days) != 14 {
		t.Fatalf("days = %d, want 14 (1 Sep through 14 Sep)", len(days))
	}

	if !days[0].Day.Equal(sept(1)) || !days[13].Day.Equal(sept(14)) {
		t.Errorf("days run %s..%s, want 2026-09-01..2026-09-14",
			days[0].Day.Format(time.DateOnly), days[13].Day.Format(time.DateOnly))
	}

	// Past its term there is nothing left to compute, whatever day is asked for.
	if got := len(deposit.PendingDays(sept(30))); got != 29 {
		t.Errorf("days through 30 Sep = %d, want 29: the rate ends on the 29th", got)
	}
}

// The worked example: $10.000.000 at 10 % E.A. opened on 1 September and
// registered on the 15th is worth 36.624,23 more than what went into it.
//
// The credit of each day is rounded to the currency's minor unit and the next
// day earns on it, so what the balance shows differs from the exact figure by
// the rounding; the carry holds that difference, which is why the two together
// land on it.
func TestFourteenDaysOfADepositAtTenPercent(t *testing.T) {
	version := CashRateVersion{
		ID:              uuid.New(),
		AnnualRate:      mustDecimal(t, "0.10"),
		WithholdingRate: decimal.Zero,
		Posting:         PostingDaily,
	}

	held, carry, credited := mustDecimal(t, "10000000"), decimal.Zero, decimal.Zero

	for range 14 {
		day, err := accrueDay(CashDay{Basis: held, AccountHeld: held, Carry: carry, Credits: true}, version, money.COP)
		if err != nil {
			t.Fatalf("accrueDay: %v", err)
		}

		credited = credited.Add(day.Credited)
		held = held.Add(day.Credited)
		carry = day.Carry
	}

	near(t, "fourteen days of interest", credited.Add(carry), "36624.23", "0.01")
}

// checkCashPocketError wants no error for an empty want, and otherwise an
// ErrInvalidCashPocket that says want.
func checkCashPocketError(t *testing.T, err error, want string) {
	t.Helper()

	if want == "" {
		if err != nil {
			t.Errorf("err = %v, want none", err)
		}

		return
	}

	if !errors.Is(err, ErrInvalidCashPocket) || !strings.Contains(err.Error(), want) {
		t.Errorf("err = %v, want ErrInvalidCashPocket saying %q", err, want)
	}
}

func TestNewFixedDepositInputValidate(t *testing.T) {
	// Mid-afternoon: the day is what is compared, not the hour.
	now := time.Date(2026, time.September, 15, 15, 30, 0, 0, time.UTC)

	valid := func() NewFixedDepositInput {
		matures := time.Date(2026, time.November, 30, 0, 0, 0, 0, time.UTC)

		return NewFixedDepositInput{
			PortfolioID:    uuid.New(),
			SourceID:       uuid.New(),
			Currency:       money.COP,
			Name:           "CDT 90 días",
			Amount:         mustDecimal(t, "10000000"),
			OpenedOn:       sept(1),
			MaturesOn:      &matures,
			AnnualRatePct:  mustDecimal(t, "10"),
			WithholdingPct: mustDecimal(t, "4"),
			Posting:        PostingDaily,
		}
	}

	day := func(y int, m time.Month, d int) *time.Time {
		on := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)

		return &on
	}

	cases := []struct {
		name string
		with func(in *NewFixedDepositInput)
		want string
	}{
		{"as stated", func(*NewFixedDepositInput) {}, ""},
		{"no term", func(in *NewFixedDepositInput) { in.MaturesOn = nil }, ""},
		{"opened today", func(in *NewFixedDepositInput) { in.OpenedOn = now }, ""},
		{"credited at maturity", func(in *NewFixedDepositInput) { in.Posting = PostingAtMaturity }, ""},

		{"no portfolio", func(in *NewFixedDepositInput) { in.PortfolioID = uuid.UUID{} }, "portfolioId and sourceId are required"},
		{"no platform", func(in *NewFixedDepositInput) { in.SourceID = uuid.UUID{} }, "portfolioId and sourceId are required"},
		{"a currency with no rate", func(in *NewFixedDepositInput) { in.Currency = money.ARS }, "currency must be one of"},
		{"no name", func(in *NewFixedDepositInput) { in.Name = "   " }, "name is required"},
		{"nothing in it", func(in *NewFixedDepositInput) { in.Amount = decimal.Zero }, "amount must be greater than 0"},
		{"more than the column holds", func(in *NewFixedDepositInput) { in.Amount = mustDecimal(t, "1000000000000") }, "amount must be greater than 0"},

		{"no opening day", func(in *NewFixedDepositInput) { in.OpenedOn = time.Time{} }, "openedOn is required"},
		{"opened tomorrow", func(in *NewFixedDepositInput) { in.OpenedOn = *day(2026, time.September, 16) }, "openedOn cannot be in the future"},
		{"opened six years ago", func(in *NewFixedDepositInput) { in.OpenedOn = *day(2020, time.September, 1) }, "more than 5 years ago"},

		{"matures the day it opened", func(in *NewFixedDepositInput) { in.MaturesOn = day(2026, time.September, 1) }, "maturesOn must be after openedOn"},
		{"already came due", func(in *NewFixedDepositInput) { in.MaturesOn = day(2026, time.September, 10) }, "maturesOn cannot be before"},
		{"came due yesterday", func(in *NewFixedDepositInput) { in.MaturesOn = day(2026, time.September, 14) }, ""},

		{"at maturity with no term", func(in *NewFixedDepositInput) {
			in.Posting = PostingAtMaturity
			in.MaturesOn = nil
		}, "needs a maturesOn"},

		{"a rate of nothing", func(in *NewFixedDepositInput) { in.AnnualRatePct = decimal.Zero }, ""},
		{"a posting nobody offers", func(in *NewFixedDepositInput) { in.Posting = InterestPosting("weekly") }, ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := valid()
			tc.with(&in)

			err := in.Validate(now)

			// The values of the rate answer as a rate, with their own wording.
			switch tc.name {
			case "a rate of nothing":
				checkCashRateError(t, err, "annualRatePct must be greater than 0")
			case "a posting nobody offers":
				checkCashRateError(t, err, "posting must be one of: daily, monthly, at_maturity")
			default:
				checkCashPocketError(t, err, tc.want)
			}
		})
	}
}

func TestCloseFixedDepositInputValidate(t *testing.T) {
	now := time.Date(2026, time.September, 15, 15, 30, 0, 0, time.UTC)

	cases := []struct {
		name string
		in   CloseFixedDepositInput
		want string
	}{
		{"today, with a penalty", CloseFixedDepositInput{ClosesOn: sept(15), Penalty: mustDecimal(t, "50000")}, ""},
		{"tomorrow", CloseFixedDepositInput{ClosesOn: sept(16)}, "closesOn cannot be in the future"},
		{"today, with none", CloseFixedDepositInput{ClosesOn: sept(15)}, ""},
		// The grace every rate date gets: an owner west of Greenwich is still on
		// yesterday for part of the evening.
		{"yesterday", CloseFixedDepositInput{ClosesOn: sept(14)}, ""},
		{"no day", CloseFixedDepositInput{}, "closesOn is required"},
		{"the day before yesterday", CloseFixedDepositInput{ClosesOn: sept(13)}, "closesOn cannot be before"},
		{"a penalty below zero", CloseFixedDepositInput{ClosesOn: sept(15), Penalty: mustDecimal(t, "-1")}, "penalty cannot be negative"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			checkCashPocketError(t, tc.in.Validate(now), tc.want)
		})
	}
}
