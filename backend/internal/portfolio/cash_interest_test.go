package portfolio

import (
	"testing"
	"time"

	"uuid"

	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"
)

func TestDailyRateCompoundsToTheAnnualRate(t *testing.T) {
	year := decimal.MustFromString("365")

	for _, annual := range []string{"0.005", "0.09", "0.135"} {
		daily, err := dailyRate(mustDecimal(t, annual))
		if err != nil {
			t.Fatalf("dailyRate(%s): %v", annual, err)
		}

		compounded := daily.Add(decimal.One).MustPow(year).Sub(decimal.One)
		aboutRate(t, "compounded "+annual, compounded, mustDecimal(t, annual))
	}
}

// aboutRate compares two fractions to twelve decimals.
func aboutRate(t *testing.T, label string, got, want decimal.Decimal) {
	t.Helper()

	if got.Sub(want).Abs().GreaterThan(decimal.MustFromString("0.000000000001")) {
		t.Errorf("%s = %s, want %s", label, got, want)
	}
}

func nineAPercent(t *testing.T, withheld string) CashRateVersion {
	t.Helper()

	return CashRateVersion{
		ID:              uuid.New(),
		AnnualRate:      mustDecimal(t, "0.09"),
		WithholdingRate: mustDecimal(t, withheld),
	}
}

// aDay is a day that credits what it earns, on a balance that is its whole
// account: the shape of every day of a rate posted daily.
func aDay(t *testing.T, basis, carry string) CashDay {
	t.Helper()

	held := mustDecimal(t, basis)

	return CashDay{Basis: held, AccountHeld: held, Carry: mustDecimal(t, carry), Credits: true}
}

func TestAccrueDay(t *testing.T) {
	cases := []struct {
		name     string
		basis    string
		withheld string
		carry    string
		cur      money.Currency
		credited string
		carried  string
	}{
		// 10 000 × ((1.09)^(1/365) − 1) = 2.3613115…
		{"a day on ten thousand dollars", "10000", "0", "0", money.USD, "2.36", "0.0013"},
		{"with the carry of the day before", "10000", "0", "0.004", money.USD, "2.37", "-0.0047"},
		// 7 % withheld leaves 2.1960197…
		{"net of withholding", "10000", "0.07", "0", money.USD, "2.20", "-0.0040"},
		{"a currency without cents", "1000000", "0", "0", money.JPY, "236", "0.1312"},
		{"an empty balance keeps the carry", "0", "0", "0.004", money.USD, "0", "0.004"},
		{"a balance below zero earns nothing", "-50", "0", "0", money.USD, "0", "0"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := accrueDay(aDay(t, tc.basis, tc.carry), nineAPercent(t, tc.withheld), tc.cur)
			if err != nil {
				t.Fatalf("accrueDay: %v", err)
			}

			sameAmount(t, "credited", got.Credited.String(), tc.credited)
			aboutAmount(t, "carry", got.Carry.String(), tc.carried)

			// Whatever rounding does, nothing is created or lost on the day.
			sameAmount(t, "credited + carry", got.Credited.Add(got.Carry).String(), got.Net.Add(mustDecimal(t, tc.carry)).String())
		})
	}
}

// A rate posted monthly computes every day exactly as a daily one does. The day
// that closes the month credits what every day of it earned, in one amount; the
// days before it credit nothing and leave the carry alone.
func TestAccrueDayHoldsWhatItDoesNotCredit(t *testing.T) {
	version := nineAPercent(t, "0")
	version.Posting = PostingMonthly

	day := aDay(t, "10000", "0.004")
	day.Credits = false

	held, err := accrueDay(day, version, money.USD)
	if err != nil {
		t.Fatalf("accrueDay: %v", err)
	}

	aboutAmount(t, "net", held.Net.String(), "2.3613")
	sameAmount(t, "credited", held.Credited.String(), "0")
	sameAmount(t, "carry", held.Carry.String(), "0.004")

	// The month closes on a day that earns the same and pays them both.
	day.Credits = true
	day.Held = held.Net

	closing, err := accrueDay(day, version, money.USD)
	if err != nil {
		t.Fatalf("accrueDay: %v", err)
	}

	sameAmount(t, "credited", closing.Credited.String(), "4.73")
	sameAmount(t, "credited + carry", closing.Credited.Add(closing.Carry).String(),
		closing.Net.Add(day.Held).Add(day.Carry).String())
}

func TestCreditsOn(t *testing.T) {
	daily := CashRateVersion{Posting: PostingDaily}
	monthly := CashRateVersion{Posting: PostingMonthly}

	days := []struct {
		day   time.Time
		close bool
	}{
		{time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC), false},
		{time.Date(2026, time.September, 29, 0, 0, 0, 0, time.UTC), false},
		{time.Date(2026, time.September, 30, 0, 0, 0, 0, time.UTC), true},
		{time.Date(2026, time.December, 31, 0, 0, 0, 0, time.UTC), true},
		// 2028 is a leap year, so February closes a day later.
		{time.Date(2028, time.February, 28, 0, 0, 0, 0, time.UTC), false},
		{time.Date(2028, time.February, 29, 0, 0, 0, 0, time.UTC), true},
	}

	for _, d := range days {
		if !daily.creditsOn(d.day) {
			t.Errorf("a daily rate does not credit on %s", d.day.Format(time.DateOnly))
		}

		if got := monthly.creditsOn(d.day); got != d.close {
			t.Errorf("a monthly rate credits on %s = %v, want %v", d.day.Format(time.DateOnly), got, d.close)
		}
	}
}

// A cap belongs to the account, so the balances of one account earn on their
// share of it and never on more than it together.
func TestEarningBasisSharesTheCap(t *testing.T) {
	limit := mustDecimal(t, "5000")
	capped := CashRateVersion{MaxBalance: &limit}

	cases := []struct {
		name        string
		basis       string
		accountHeld string
		want        string
	}{
		{"under the cap", "4000", "4000", "4000"},
		{"exactly the cap", "5000", "5000", "5000"},
		{"the whole account over it", "8000", "8000", "5000"},
		{"a quarter of an account over it", "2500", "10000", "1250"},
		{"the rest of that account", "7500", "10000", "3750"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := capped.earningBasis(mustDecimal(t, tc.basis), mustDecimal(t, tc.accountHeld))
			if err != nil {
				t.Fatalf("earningBasis: %v", err)
			}

			sameAmount(t, "earning basis", got.String(), tc.want)
		})
	}

	t.Run("no cap earns on all of it", func(t *testing.T) {
		got, err := CashRateVersion{}.earningBasis(mustDecimal(t, "8000"), mustDecimal(t, "10000"))
		if err != nil {
			t.Fatalf("earningBasis: %v", err)
		}

		sameAmount(t, "earning basis", got.String(), "8000")
	})
}

func TestCreditHeld(t *testing.T) {
	credited, left, err := creditHeld(mustDecimal(t, "70.8135"), mustDecimal(t, "0.0042"), money.USD)
	if err != nil {
		t.Fatalf("creditHeld: %v", err)
	}

	sameAmount(t, "credited", credited.String(), "70.82")
	sameAmount(t, "credited + left", credited.Add(left).String(), "70.8177")

	// Less than a cent of interest is carried, not credited.
	credited, left, err = creditHeld(mustDecimal(t, "0.003"), decimal.Zero, money.USD)
	if err != nil {
		t.Fatalf("creditHeld: %v", err)
	}

	sameAmount(t, "credited", credited.String(), "0")
	sameAmount(t, "left", left.String(), "0.003")
}

// A small balance earns less than a cent a day. The carry keeps adding it up
// until there is a cent to credit, so a year of rounded credits is the year's
// interest.
func TestAccrueDayLosesNothingToRoundingOverAYear(t *testing.T) {
	version := nineAPercent(t, "0")
	basis := mustDecimal(t, "10")

	carry, credited, net := decimal.Zero, decimal.Zero, decimal.Zero
	for range 365 {
		day, err := accrueDay(CashDay{Basis: basis, AccountHeld: basis, Carry: carry, Credits: true}, version, money.USD)
		if err != nil {
			t.Fatalf("accrueDay: %v", err)
		}

		credited = credited.Add(day.Credited)
		net = net.Add(day.Net)
		carry = day.Carry
	}

	aboutAmount(t, "credited over the year", credited.String(), "0.86")
	aboutRate(t, "credited + final carry", credited.Add(carry), net)
}

func TestCashAccrualTargetPendingDays(t *testing.T) {
	sep := func(day int) time.Time { return time.Date(2026, time.September, day, 0, 0, 0, 0, time.UTC) }
	endedOn := sep(2)

	paused := CashRateVersion{ID: uuid.New(), EffectiveFrom: sep(1), EndedOn: &endedOn}
	resumed := CashRateVersion{ID: uuid.New(), EffectiveFrom: sep(5)}

	target := CashAccrualTarget{OpenedOn: sep(1).AddDate(0, 0, -20), Versions: []CashRateVersion{paused, resumed}}

	type want struct {
		day  time.Time
		rate uuid.UUID
	}

	check := func(t *testing.T, got []CashAccrualDay, wants ...want) {
		t.Helper()

		if len(got) != len(wants) {
			t.Fatalf("days = %+v, want %d of them", got, len(wants))
		}
		for i, w := range wants {
			if !got[i].Day.Equal(w.day) || got[i].RateID != w.rate {
				t.Errorf("day %d = %s at %v, want %s at %v", i, got[i].Day.Format(time.DateOnly), got[i].RateID, w.day.Format(time.DateOnly), w.rate)
			}
		}
	}

	t.Run("skips the pause", func(t *testing.T) {
		check(t, target.PendingDays(sep(6)),
			want{sep(1), paused.ID}, want{sep(2), paused.ID}, want{sep(5), resumed.ID}, want{sep(6), resumed.ID})
	})

	t.Run("starts after the last day computed", func(t *testing.T) {
		last := sep(5)
		computed := target
		computed.LastAccrued = &last

		check(t, computed.PendingDays(sep(6)), want{sep(6), resumed.ID})
	})

	t.Run("starts on the day the balance was opened", func(t *testing.T) {
		opened := target
		opened.OpenedOn = sep(6).Add(15 * time.Hour)

		check(t, opened.PendingDays(sep(6)), want{sep(6), resumed.ID})
	})

	t.Run("nothing before the first rate or without one", func(t *testing.T) {
		check(t, target.PendingDays(sep(1).AddDate(0, 0, -1)))
		check(t, CashAccrualTarget{OpenedOn: sep(1)}.PendingDays(sep(6)))
	})
}
