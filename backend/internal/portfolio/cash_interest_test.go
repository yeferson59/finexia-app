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

// twelveThenEight pays 12 % up to from and 8 % on what is above it.
func twelveThenEight(t *testing.T, from string) CashRateVersion {
	t.Helper()

	return CashRateVersion{
		ID:              uuid.New(),
		AnnualRate:      mustDecimal(t, "0.12"),
		WithholdingRate: decimal.Zero,
		Tiers:           []CashRateStep{{From: mustDecimal(t, from), AnnualRate: mustDecimal(t, "0.08")}},
	}
}

// near fails unless got is within tolerance of want.
func near(t *testing.T, label string, got decimal.Decimal, want, tolerance string) {
	t.Helper()

	if got.Sub(mustDecimal(t, want)).Abs().GreaterThan(mustDecimal(t, tolerance)) {
		t.Errorf("%s = %s, want %s ± %s", label, got, want, tolerance)
	}
}

// A version pays in steps of what its account holds: its own rate from zero,
// and each tier's from its own point up.
func TestAccountGrossPaysInSteps(t *testing.T) {
	tiered := twelveThenEight(t, "5000000")

	cases := []struct {
		name string
		held string
		want string
	}{
		// 5 000 000 × ((1.12)^(1/365) − 1) + 3 000 000 × ((1.08)^(1/365) − 1).
		{"across two steps", "8000000", "2185.311973"},
		{"inside the first", "4000000", "1242.151023"},
		{"exactly where the second starts", "5000000", "1552.688778"},
		{"nothing held", "0", "0"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tiered.accountGross(mustDecimal(t, tc.held))
			if err != nil {
				t.Fatalf("accountGross: %v", err)
			}

			near(t, "gross", got, tc.want, "0.000001")
		})
	}
}

// Three steps, so a middle one is bounded on both sides: 12 % up to 5 000 000,
// 8 % up to 10 000 000, 5 % above that. On 12 000 000 the middle step earns on
// the 5 000 000 between its neighbours, and the last on the 2 000 000 over it.
func TestAccountGrossPaysInThreeSteps(t *testing.T) {
	version := CashRateVersion{
		AnnualRate: mustDecimal(t, "0.12"),
		Tiers: []CashRateStep{
			{From: mustDecimal(t, "5000000"), AnnualRate: mustDecimal(t, "0.08")},
			{From: mustDecimal(t, "10000000"), AnnualRate: mustDecimal(t, "0.05")},
		},
	}

	got, err := version.accountGross(mustDecimal(t, "12000000"))
	if err != nil {
		t.Fatalf("accountGross: %v", err)
	}

	// 5 000 000 × i(12 %) + 5 000 000 × i(8 %) + 2 000 000 × i(5 %).
	near(t, "gross", got, "2874.422004", "0.000001")

	// A step above what the account holds earns nothing: on 7 000 000 only the
	// first two steps pay, and the third never starts.
	inTwo, err := version.accountGross(mustDecimal(t, "7000000"))
	if err != nil {
		t.Fatalf("accountGross: %v", err)
	}

	near(t, "gross inside the middle step", inTwo, "1974.437575", "0.000001")
}

// A cap is a step at zero. The account earns on the cap and no more, and its
// balances share that in proportion to what each holds: what the cap column
// computed before 000046, on the same figures.
func TestDayGrossSharesACap(t *testing.T) {
	nine := mustDecimal(t, "0.09")
	capped := CashRateVersion{
		AnnualRate: nine,
		Tiers:      []CashRateStep{{From: mustDecimal(t, "5000"), AnnualRate: decimal.Zero}},
	}

	daily, err := dailyRate(nine)
	if err != nil {
		t.Fatalf("dailyRate: %v", err)
	}

	cases := []struct {
		name        string
		basis       string
		accountHeld string
		earnsOn     string
	}{
		{"under the cap", "4000", "4000", "4000"},
		{"exactly the cap", "5000", "5000", "5000"},
		{"the whole account over it", "8000", "8000", "5000"},
		{"a quarter of an account over it", "2500", "10000", "1250"},
		{"the rest of that account", "7500", "10000", "3750"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := capped.dayGross(mustDecimal(t, tc.basis), mustDecimal(t, tc.accountHeld))
			if err != nil {
				t.Fatalf("dayGross: %v", err)
			}

			aboutRate(t, "gross", got, mustDecimal(t, tc.earnsOn).Mul(daily))
		})
	}

	t.Run("no tiers earns on all of it", func(t *testing.T) {
		got, err := CashRateVersion{AnnualRate: nine}.dayGross(mustDecimal(t, "8000"), mustDecimal(t, "10000"))
		if err != nil {
			t.Fatalf("dayGross: %v", err)
		}

		aboutRate(t, "gross", got, mustDecimal(t, "8000").Mul(daily))
	})
}

// Tiers belong to the account: two balances of it take the account's day in
// proportion to what each holds, not each a day of its own on its own figure.
func TestDayGrossSharesTiersAcrossTheAccount(t *testing.T) {
	tiered := twelveThenEight(t, "5000000")
	held := mustDecimal(t, "8000000")

	whole, err := tiered.accountGross(held)
	if err != nil {
		t.Fatalf("accountGross: %v", err)
	}

	larger, err := tiered.dayGross(mustDecimal(t, "6000000"), held)
	if err != nil {
		t.Fatalf("dayGross: %v", err)
	}

	smaller, err := tiered.dayGross(mustDecimal(t, "2000000"), held)
	if err != nil {
		t.Fatalf("dayGross: %v", err)
	}

	near(t, "the two shares", larger.Add(smaller), whole.String(), "0.000000001")
	near(t, "the smaller share", smaller, "546.327993", "0.000001")

	// On its own the smaller balance would sit wholly in the first step.
	alone, err := tiered.dayGross(mustDecimal(t, "2000000"), mustDecimal(t, "2000000"))
	if err != nil {
		t.Fatalf("dayGross: %v", err)
	}
	if !alone.GreaterThan(smaller) {
		t.Errorf("alone = %s, want more than its share of the account, %s", alone, smaller)
	}
}

func TestEffectiveRate(t *testing.T) {
	tiered := twelveThenEight(t, "5000000")

	blended, err := tiered.effectiveRate(mustDecimal(t, "8000000"))
	if err != nil {
		t.Fatalf("effectiveRate: %v", err)
	}
	near(t, "across two steps", blended, "0.104829742", "0.000000001")

	inside, err := tiered.effectiveRate(mustDecimal(t, "4000000"))
	if err != nil {
		t.Fatalf("effectiveRate: %v", err)
	}
	near(t, "inside the first step", inside, "0.12", "0.000000001")

	// With nothing held, or without tiers, it is the version's own rate as it is.
	for label, version := range map[string]CashRateVersion{
		"nothing held": tiered,
		"no tiers":     {AnnualRate: mustDecimal(t, "0.12")},
	} {
		got, err := version.effectiveRate(decimal.Zero)
		if err != nil {
			t.Fatalf("effectiveRate: %v", err)
		}
		if !got.Equal(mustDecimal(t, "0.12")) {
			t.Errorf("%s = %s, want 0.12", label, got)
		}
	}
}

// A day on a tiered account credits the balance's share of the account's day,
// keeps what the balance held as its basis, and keeps the rate the steps came
// to as the day's rate.
func TestAccrueDayOnTiers(t *testing.T) {
	day := CashDay{Basis: mustDecimal(t, "6000"), AccountHeld: mustDecimal(t, "8000"), Carry: decimal.Zero, Credits: true}

	got, err := accrueDay(day, twelveThenEight(t, "5000"), money.USD)
	if err != nil {
		t.Fatalf("accrueDay: %v", err)
	}

	// Three quarters of 2.185311973.
	sameAmount(t, "basis", got.Basis.String(), "6000")
	near(t, "gross", got.Gross, "1.638983980", "0.000000001")
	sameAmount(t, "credited", got.Credited.String(), "1.64")
	near(t, "the day's rate", got.AnnualRate, "0.104829742", "0.000000001")

	// A reading of the account below the balance is taken as the balance alone.
	day.AccountHeld = decimal.Zero
	alone, err := accrueDay(day, twelveThenEight(t, "5000"), money.USD)
	if err != nil {
		t.Fatalf("accrueDay: %v", err)
	}
	near(t, "gross on the balance alone", alone.Gross, "1.763563177", "0.000000001")
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
