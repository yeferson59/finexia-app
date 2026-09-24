package portfolio

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"uuid"

	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

var fundToday = time.Date(2026, time.September, 30, 15, 0, 0, 0, time.UTC)

func validFundInput(t *testing.T) NewFundInput {
	t.Helper()

	return NewFundInput{
		PortfolioID: uuid.New(),
		SourceID:    uuid.New(),
		Name:        " FIC Renta Fija ",
		Currency:    money.COP,
		Tracking:    FundUnits,
		Date:        time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC),
		Units:       mustDecimal(t, "1000"),
		UnitValue:   mustDecimal(t, "12345.678901"),
	}
}

func TestNewFundInputValidate(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*NewFundInput)
		want   string
	}{
		{"valid", func(*NewFundInput) {}, ""},
		{"with a current value", func(in *NewFundInput) {
			in.CurrentUnitValue = mustDecimal(t, "12431.22")
			in.CurrentDate = fundToday
		}, ""},
		{"no portfolio", func(in *NewFundInput) { in.PortfolioID = uuid.UUID{} }, "portfolioId and sourceId are required"},
		{"no platform", func(in *NewFundInput) { in.SourceID = uuid.UUID{} }, "portfolioId and sourceId are required"},
		{"blank name", func(in *NewFundInput) { in.Name = "   " }, "name is required"},
		{"long name", func(in *NewFundInput) { in.Name = strings.Repeat("a", maxFundNameLen+1) }, "cannot exceed"},
		{"unsupported currency", func(in *NewFundInput) { in.Currency = money.XXX }, "currency must be one of"},
		{"by balance with units", func(in *NewFundInput) { in.Tracking = FundBalance }, "states amounts and balances"},
		{"by units with an amount", func(in *NewFundInput) { in.Amount = mustDecimal(t, "100") }, "states units and unit values"},
		{"unknown tracking", func(in *NewFundInput) { in.Tracking = "shares" }, "tracking must be"},
		{"no date", func(in *NewFundInput) { in.Date = time.Time{} }, "date is required"},
		{"future date", func(in *NewFundInput) { in.Date = fundToday.AddDate(0, 0, 1) }, "future"},
		{"ancient date", func(in *NewFundInput) { in.Date = fundToday.AddDate(-maxFundMarkYears-1, 0, 0) }, "years ago"},
		{"no units", func(in *NewFundInput) { in.Units = mustDecimal(t, "0") }, "units must be greater than zero"},
		{"negative units", func(in *NewFundInput) { in.Units = mustDecimal(t, "-1") }, "units must be greater than zero"},
		{"no unit value", func(in *NewFundInput) { in.UnitValue = mustDecimal(t, "0") }, "unit value must be"},
		{"huge unit value", func(in *NewFundInput) { in.UnitValue = maxFundUnitValue }, "unit value must be"},
		{"negative current value", func(in *NewFundInput) { in.CurrentUnitValue = mustDecimal(t, "-1") }, "cannot be negative"},
		{"current value before the purchase", func(in *NewFundInput) {
			in.CurrentUnitValue = mustDecimal(t, "12000")
			in.CurrentDate = in.Date.AddDate(0, 0, -1)
		}, "before the purchase"},
		{"current value in the future", func(in *NewFundInput) {
			in.CurrentUnitValue = mustDecimal(t, "12000")
			in.CurrentDate = fundToday.AddDate(0, 0, 2)
		}, "future"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := validFundInput(t)
			tc.mutate(&in)

			err := in.Validate(fundToday)
			if tc.want == "" {
				if err != nil {
					t.Fatalf("Validate() = %v, want nil", err)
				}

				return
			}

			if !errors.Is(err, ErrInvalidFund) || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("Validate() = %v, want ErrInvalidFund mentioning %q", err, tc.want)
			}
		})
	}
}

// A current value given without a day is today's, and the purchase it is
// checked against is the one the owner stated.
func TestNewFundInputDefaultsTheCurrentDate(t *testing.T) {
	in := validFundInput(t)
	in.CurrentUnitValue = mustDecimal(t, "12431.22")

	in = in.withDefaults(fundToday)

	if want := cashRateDay(fundToday); !in.CurrentDate.Equal(want) {
		t.Fatalf("CurrentDate = %v, want %v", in.CurrentDate, want)
	}

	if err := in.Validate(fundToday); err != nil {
		t.Fatalf("Validate() = %v", err)
	}

	// Without a current value there is nothing to date.
	bare := validFundInput(t).withDefaults(fundToday)
	if !bare.CurrentDate.IsZero() {
		t.Fatalf("CurrentDate = %v, want zero without a current value", bare.CurrentDate)
	}
}

// The first purchase is an ordinary buy, priced in the fund's currency, on the
// day it was made.
func TestNewFundInputPurchase(t *testing.T) {
	in := validFundInput(t)
	in.Date = time.Date(2026, time.September, 1, 18, 30, 0, 0, time.UTC)
	in.PayFromCash = true

	p := in.purchase()

	if p.Type != Buy || !p.Quantity.Equal(in.Units) || p.Currency != money.COP || !p.PayFromCash {
		t.Fatalf("purchase = %+v", p)
	}

	if got := p.Price.String(); got != "12345.678901" {
		t.Errorf("price = %s, want 12345.678901", got)
	}

	if want := time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC); !p.TransactionDate.Equal(want) {
		t.Errorf("date = %v, want %v", p.TransactionDate, want)
	}

	if _, err := p.Validate(money.COP); err != nil {
		t.Errorf("the purchase does not pass TransactionInput.Validate: %v", err)
	}
}

func TestFundMarkInputValidate(t *testing.T) {
	valid := func() FundMarkInput {
		return FundMarkInput{Date: fundToday, UnitValue: mustDecimal(t, "12431.22")}
	}

	cases := []struct {
		name     string
		mutate   func(*FundMarkInput)
		tracking FundTracking
		want     string
	}{
		{"valid", func(*FundMarkInput) {}, FundUnits, ""},
		{"no date", func(in *FundMarkInput) { in.Date = time.Time{} }, FundUnits, "date is required"},
		{"future", func(in *FundMarkInput) { in.Date = fundToday.AddDate(0, 0, 1) }, FundUnits, "future"},
		{"no unit value", func(in *FundMarkInput) { in.UnitValue = mustDecimal(t, "0") }, FundUnits, "unit value must be"},
		{"a balance on a fund by units", func(in *FundMarkInput) { in.Balance = mustDecimal(t, "100") }, FundUnits, "not a balance"},
		{"a unit value on a fund by balance", func(*FundMarkInput) {}, FundBalance, "not a unit value"},
		{"a fund by balance", func(in *FundMarkInput) {
			in.UnitValue = mustDecimal(t, "0")
			in.Balance = mustDecimal(t, "13050000")
		}, FundBalance, ""},
		{"a fund by balance with no balance", func(in *FundMarkInput) { in.UnitValue = mustDecimal(t, "0") }, FundBalance, "balance must be"},
		{"long notes", func(in *FundMarkInput) { in.Notes = strings.Repeat("n", maxCashNotesLen+1) }, FundUnits, "notes cannot exceed"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := valid()
			tc.mutate(&in)

			err := in.Validate(fundToday, tc.tracking)
			if tc.want == "" {
				if err != nil {
					t.Fatalf("Validate() = %v, want nil", err)
				}

				return
			}

			if !errors.Is(err, ErrInvalidFundMark) || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("Validate() = %v, want ErrInvalidFundMark mentioning %q", err, tc.want)
			}
		})
	}
}

func TestNewFundTicker(t *testing.T) {
	a, b := newFundTicker(), newFundTicker()

	if !strings.HasPrefix(a, fundTickerPrefix) || len(a) != len(fundTickerPrefix)+8 || len(a) > 20 {
		t.Fatalf("ticker %q, want %s and eight hex digits within 20 characters", a, fundTickerPrefix)
	}

	if a == b {
		t.Fatalf("two tickers came out the same: %q", a)
	}
}

// A fund adds its positions up and values them at the latest mark; with no
// mark it is worth its cost, and says so.
func TestFundTotal(t *testing.T) {
	positions := []FundPosition{
		{Units: "600", Cost: "7407407.3406"},
		{Units: "400", Cost: "4938271.5604"},
	}

	marked := Fund{Positions: positions, UnitValue: new("12431.22")}
	if err := marked.total(); err != nil {
		t.Fatalf("total() = %v", err)
	}

	sameAmount(t, "units", marked.Units, "1000")
	sameAmount(t, "cost", marked.Cost, "12345678.901")
	sameAmount(t, "value", marked.Value, "12431220")

	if marked.PricedAtCost {
		t.Error("a marked fund reads as priced at cost")
	}

	unmarked := Fund{Positions: positions}
	if err := unmarked.total(); err != nil {
		t.Fatalf("total() = %v", err)
	}

	sameAmount(t, "value at cost", unmarked.Value, "12345678.901")

	if !unmarked.PricedAtCost {
		t.Error("a fund with no mark does not say it is priced at cost")
	}
}

// Reading how a fund is followed comes first: the rules of a mark depend on it,
// and a fund the user does not follow answers not found without writing.
func TestSaveFundMarkChecksTheFundFirst(t *testing.T) {
	assetID := uuid.New()
	wrote := false

	repo := new(fakeRepository{
		getFund: func(context.Context, uuid.UUID, uuid.UUID) (Fund, error) {
			return Fund{}, ErrFundNotFound
		},
		upsertFundMark: func(context.Context, uuid.UUID, uuid.UUID, FundMarkInput) (FundMark, error) {
			wrote = true
			return FundMark{}, nil
		},
	})
	svc := newService(repo, testConfig(), nil, nil, nil, logger.Noop())

	_, err := svc.SaveFundMark(context.Background(), uuid.New(), assetID, FundMarkInput{
		Date: time.Now(), UnitValue: mustDecimal(t, "10"),
	})
	if !errors.Is(err, ErrFundNotFound) || wrote {
		t.Fatalf("SaveFundMark() = %v (wrote %v), want ErrFundNotFound and no write", err, wrote)
	}

	repo.getFund = func(context.Context, uuid.UUID, uuid.UUID) (Fund, error) {
		return Fund{AssetID: assetID, Tracking: FundUnits}, nil
	}

	if _, err := svc.SaveFundMark(context.Background(), uuid.New(), assetID, FundMarkInput{
		Date: time.Now(), UnitValue: mustDecimal(t, "0"),
	}); !errors.Is(err, ErrInvalidFundMark) || wrote {
		t.Fatalf("SaveFundMark(0) = %v (wrote %v), want ErrInvalidFundMark and no write", err, wrote)
	}

	if _, err := svc.SaveFundMark(context.Background(), uuid.New(), assetID, FundMarkInput{
		Date: time.Now(), UnitValue: mustDecimal(t, "10"),
	}); err != nil || !wrote {
		t.Fatalf("SaveFundMark(10) = %v (wrote %v), want the mark written", err, wrote)
	}
}

func validBalanceFund(t *testing.T) NewFundInput {
	t.Helper()

	return NewFundInput{
		PortfolioID: uuid.New(),
		SourceID:    uuid.New(),
		Name:        "Bolsillo de inversión",
		Currency:    money.COP,
		Tracking:    FundBalance,
		Date:        time.Date(2026, time.January, 10, 0, 0, 0, 0, time.UTC),
		Amount:      mustDecimal(t, "12000000"),
	}
}

func TestNewBalanceFundInputValidate(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*NewFundInput)
		want   string
	}{
		{"valid", func(*NewFundInput) {}, ""},
		{"with a current balance", func(in *NewFundInput) {
			in.CurrentBalance = mustDecimal(t, "12640000")
			in.CurrentDate = fundToday
		}, ""},
		{"no amount", func(in *NewFundInput) { in.Amount = mustDecimal(t, "0") }, "amount must be"},
		{"a unit value", func(in *NewFundInput) { in.CurrentUnitValue = mustDecimal(t, "1") }, "not units"},
		{"a negative balance", func(in *NewFundInput) { in.CurrentBalance = mustDecimal(t, "-1") }, "cannot be negative"},
		{"a balance before the contribution", func(in *NewFundInput) {
			in.CurrentBalance = mustDecimal(t, "1")
			in.CurrentDate = in.Date.AddDate(0, 0, -1)
		}, "before the contribution"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := validBalanceFund(t)
			tc.mutate(&in)

			err := in.Validate(fundToday)
			if tc.want == "" {
				if err != nil {
					t.Fatalf("Validate() = %v, want nil", err)
				}

				return
			}

			if !errors.Is(err, ErrInvalidFund) || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("Validate() = %v, want ErrInvalidFund mentioning %q", err, tc.want)
			}
		})
	}
}

// The quick start of a fund followed by balance: its amount buys at the
// opening unit value, and the balance it says it holds now fixes the unit value
// the replay will give it.
func TestBalanceFundOpening(t *testing.T) {
	in := validBalanceFund(t)
	in.CurrentBalance = mustDecimal(t, "12640000")
	in.CurrentDate = fundToday

	p := in.purchase()
	sameAmount(t, "units", p.Quantity.String(), "120000")
	sameAmount(t, "price", p.Price.String(), "100")

	mark, err := in.openingMark()
	if err != nil || mark == nil {
		t.Fatalf("openingMark() = %v, %v", mark, err)
	}

	sameAmount(t, "unit value", mark.UnitValue.String(), "105.33333333")
	sameAmount(t, "balance", mark.Balance.String(), "12640000")

	in.CurrentBalance = mustDecimal(t, "0")
	if mark, err := in.openingMark(); err != nil || mark != nil {
		t.Fatalf("openingMark() without a balance = %v, %v; want none", mark, err)
	}
}

func TestFundMovementInputsValidate(t *testing.T) {
	day := fundToday.AddDate(0, 0, -3)

	contribution := func() FundContributionInput {
		return FundContributionInput{PortfolioID: uuid.New(), SourceID: uuid.New(), Date: day, Amount: mustDecimal(t, "5000000")}
	}
	withdrawal := func() FundWithdrawalInput {
		return FundWithdrawalInput{EntryID: uuid.New(), Date: day, Amount: mustDecimal(t, "2000000")}
	}

	check := func(t *testing.T, err error, want string) {
		t.Helper()

		if want == "" {
			if err != nil {
				t.Fatalf("Validate() = %v, want nil", err)
			}

			return
		}

		if !errors.Is(err, ErrInvalidFund) || !strings.Contains(err.Error(), want) {
			t.Fatalf("Validate() = %v, want ErrInvalidFund mentioning %q", err, want)
		}
	}

	t.Run("contribution", func(t *testing.T) {
		check(t, contribution().Validate(fundToday), "")

		in := contribution()
		in.SourceID = uuid.UUID{}
		check(t, in.Validate(fundToday), "portfolioId and sourceId")

		in = contribution()
		in.Amount = mustDecimal(t, "0")
		check(t, in.Validate(fundToday), "amount must be")

		in = contribution()
		in.BalanceBefore = mustDecimal(t, "-1")
		check(t, in.Validate(fundToday), "balanceBefore cannot be negative")

		in = contribution()
		in.Date = fundToday.AddDate(0, 0, 1)
		check(t, in.Validate(fundToday), "future")
	})

	t.Run("withdrawal", func(t *testing.T) {
		check(t, withdrawal().Validate(fundToday), "")

		in := withdrawal()
		in.EntryID = uuid.UUID{}
		check(t, in.Validate(fundToday), "entryId is required")

		in = withdrawal()
		in.Fees = mustDecimal(t, "2000000")
		check(t, in.Validate(fundToday), "fees must be less")

		in = withdrawal()
		in.Fees = mustDecimal(t, "-1")
		check(t, in.Validate(fundToday), "fees cannot be negative")
	})

	t.Run("edit", func(t *testing.T) {
		edit := FundMovementEdit{Date: day, Amount: mustDecimal(t, "100"), All: true}
		check(t, edit.Validate(fundToday, FundWithdrawal), "")
		check(t, edit.Validate(fundToday, FundContribution), "only a withdrawal")

		edit.All = false
		edit.Fees = mustDecimal(t, "1")
		check(t, edit.Validate(fundToday, FundContribution), "carries no fees")
	})
}

func TestValidateFundMarks(t *testing.T) {
	mark := func(day int) FundMarkInput {
		return FundMarkInput{Date: time.Date(2026, time.September, day, 0, 0, 0, 0, time.UTC), UnitValue: mustDecimal(t, "10")}
	}

	if err := validateFundMarks([]FundMarkInput{mark(1), mark(2)}, fundToday, FundUnits); err != nil {
		t.Fatalf("two days: %v", err)
	}

	for name, in := range map[string][]FundMarkInput{
		"empty":         nil,
		"a day twice":   {mark(1), mark(1)},
		"one bad":       {mark(1), {Date: fundToday}},
		"too many rows": make([]FundMarkInput, maxFundMarksBatch+1),
	} {
		if err := validateFundMarks(in, fundToday, FundUnits); !errors.Is(err, ErrInvalidFundMark) {
			t.Errorf("%s = %v, want ErrInvalidFundMark", name, err)
		}
	}
}
