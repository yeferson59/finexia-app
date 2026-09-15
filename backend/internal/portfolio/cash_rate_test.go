package portfolio

import (
	"errors"
	"strings"
	"testing"
	"time"

	"uuid"

	"github.com/yeferson59/gofinance/v2/money"
)

// checkCashRateError wants no error for an empty want, and otherwise an
// ErrInvalidCashRate that says want.
func checkCashRateError(t *testing.T, err error, want string) {
	t.Helper()

	if want == "" {
		if err != nil {
			t.Errorf("err = %v, want none", err)
		}
		return
	}

	if !errors.Is(err, ErrInvalidCashRate) || !strings.Contains(err.Error(), want) {
		t.Errorf("err = %v, want ErrInvalidCashRate saying %q", err, want)
	}
}

func TestNewCashRateInputValidate(t *testing.T) {
	// Mid-afternoon: the day is what is compared, not the hour.
	now := time.Date(2026, time.September, 14, 15, 30, 0, 0, time.UTC)
	day := time.Date(2026, time.September, 14, 0, 0, 0, 0, time.UTC)

	cases := []struct {
		name string
		edit func(in *NewCashRateInput)
		want string
	}{
		{"a rate from today", func(*NewCashRateInput) {}, ""},
		{"from yesterday, still today west of UTC", func(in *NewCashRateInput) { in.EffectiveFrom = day.AddDate(0, 0, -1) }, ""},
		{"from next month", func(in *NewCashRateInput) { in.EffectiveFrom = day.AddDate(0, 1, 0) }, ""},
		{"four decimals", func(in *NewCashRateInput) { in.AnnualRatePct = mustDecimal(t, "9.2525") }, ""},
		{"a hundred percent", func(in *NewCashRateInput) { in.AnnualRatePct = mustDecimal(t, "100") }, ""},
		{"withheld interest", func(in *NewCashRateInput) { in.WithholdingPct = mustDecimal(t, "7.5") }, ""},
		{"no rate", func(in *NewCashRateInput) { in.AnnualRatePct = mustDecimal(t, "0") }, "greater than 0"},
		{"a negative rate", func(in *NewCashRateInput) { in.AnnualRatePct = mustDecimal(t, "-1") }, "greater than 0"},
		{"over a hundred", func(in *NewCashRateInput) { in.AnnualRatePct = mustDecimal(t, "100.0001") }, "at most 100"},
		{"five decimals", func(in *NewCashRateInput) { in.AnnualRatePct = mustDecimal(t, "9.25251") }, "annualRatePct takes at most 4 decimals"},
		{"negative withholding", func(in *NewCashRateInput) { in.WithholdingPct = mustDecimal(t, "-0.5") }, "at least 0"},
		{"all of it withheld", func(in *NewCashRateInput) { in.WithholdingPct = mustDecimal(t, "100") }, "less than 100"},
		{"withholding with three decimals", func(in *NewCashRateInput) { in.WithholdingPct = mustDecimal(t, "7.125") }, "withholdingPct takes at most 2 decimals"},
		{"monthly posting", func(in *NewCashRateInput) { in.Posting = PostingMonthly }, "not available yet"},
		{"an unknown posting", func(in *NewCashRateInput) { in.Posting = "weekly" }, "posting must be daily"},
		{"no platform", func(in *NewCashRateInput) { in.SourceID = uuid.UUID{} }, "sourceId is required"},
		{"a currency with no rate", func(in *NewCashRateInput) { in.Currency = money.XXX }, "currency must be one of"},
		{"no first day", func(in *NewCashRateInput) { in.EffectiveFrom = time.Time{} }, "effectiveFrom is required"},
		{"two days ago", func(in *NewCashRateInput) { in.EffectiveFrom = day.AddDate(0, 0, -2) }, "effectiveFrom cannot be before 2026-09-13"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := NewCashRateInput{
				SourceID:      uuid.New(),
				Currency:      money.COP,
				EffectiveFrom: day,
				CashRateInput: CashRateInput{
					AnnualRatePct:  mustDecimal(t, "9.25"),
					WithholdingPct: mustDecimal(t, "0"),
					Posting:        PostingDaily,
				},
			}
			tc.edit(&in)

			checkCashRateError(t, in.ValidateNew(now), tc.want)
		})
	}
}

func TestValidateCashRateEnd(t *testing.T) {
	now := time.Date(2026, time.September, 14, 15, 30, 0, 0, time.UTC)
	day := time.Date(2026, time.September, 14, 0, 0, 0, 0, time.UTC)

	checkCashRateError(t, ValidateCashRateEnd(day, now), "")
	checkCashRateError(t, ValidateCashRateEnd(day.AddDate(0, 0, -1), now), "")
	checkCashRateError(t, ValidateCashRateEnd(day.AddDate(0, 0, 10), now), "")
	checkCashRateError(t, ValidateCashRateEnd(time.Time{}, now), "endsOn is required")
	checkCashRateError(t, ValidateCashRateEnd(day.AddDate(0, 0, -2), now), "endsOn cannot be before 2026-09-13")
}
