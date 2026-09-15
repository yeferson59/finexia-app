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
			got, err := accrueDay(mustDecimal(t, tc.basis), nineAPercent(t, tc.withheld), mustDecimal(t, tc.carry), tc.cur)
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

// A small balance earns less than a cent a day. The carry keeps adding it up
// until there is a cent to credit, so a year of rounded credits is the year's
// interest.
func TestAccrueDayLosesNothingToRoundingOverAYear(t *testing.T) {
	version := nineAPercent(t, "0")
	basis := mustDecimal(t, "10")

	carry, credited, net := decimal.Zero, decimal.Zero, decimal.Zero
	for range 365 {
		day, err := accrueDay(basis, version, carry, money.USD)
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
