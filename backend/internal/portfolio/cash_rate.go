package portfolio

import (
	"errors"
	"fmt"
	"time"

	"uuid"

	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/platform/currency"
	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

// The rate a cash account earns: what the owner states, and the rules that keep
// it usable for computing interest. The model — one rate per platform and
// currency, stored as an effective annual rate, versioned by the date it takes
// effect — is described in migration 000043, and the tiers a version can pay
// in steps in 000046.
//
// Only the latest version of an account can change, and only for the days its
// interest has not been computed yet (000044). The versions before it, and the
// days already computed, are the rates the interest was earned at: a new rate is
// a new version, not an edit.

var (
	// ErrInvalidCashRate rejects a rate that cannot be recorded as stated. It is
	// wrapped with the rule that was broken.
	ErrInvalidCashRate = httpx.AsBadRequest(errors.New("invalid cash rate"))
	// ErrCashRateNotFound answers for a rate that does not exist or is on
	// someone else's platform. The two are the same thing to these endpoints.
	ErrCashRateNotFound = httpx.AsNotFound(errors.New("cash rate not found"))
	// ErrCashRateNotLatest refuses to rewrite, end or delete a version a later
	// one already follows.
	ErrCashRateNotLatest = httpx.AsConflict(errors.New("only the latest version of a cash rate can change"))
	// ErrCashRateOverlaps refuses a version that does not start after the latest
	// one: two versions from the same day, or one slotted before a version
	// already recorded, would leave some day with two rates.
	ErrCashRateOverlaps = httpx.AsConflict(errors.New("a version of this cash rate already starts on or after that date"))
	// ErrCashRateInUse refuses a change that reaches a day whose interest was
	// already computed at the rate. It is wrapped with that day and with what
	// can be done instead.
	ErrCashRateInUse = httpx.AsConflict(errors.New("the cash rate already earned interest"))
)

// InterestPosting is how often the interest a rate earns is credited to the
// balance.
type InterestPosting string

const (
	PostingDaily   InterestPosting = "daily"
	PostingMonthly InterestPosting = "monthly"
	// PostingAtMaturity computes every day as the other two do and credits them
	// all on the last day the rate earns (000048). Only a fixed deposit takes
	// it: it is the version's end that says when the credit lands, and only a
	// deposit has one it cannot move.
	PostingAtMaturity InterestPosting = "at_maturity"
)

var (
	maxAnnualRatePct  = decimal.MustFromString("100")
	maxWithholdingPct = decimal.MustFromString("100")
	// maxCashRateBalance is what NUMERIC(20, 8) holds, exclusive.
	maxCashRateBalance = decimal.MustFromString("1000000000000")
	// The owner states rates as percentages, and the ledger reads fractions.
	pctToFraction = decimal.MustFromString("0.01")
	fractionToPct = decimal.MustFromString("100")
)

const (
	// annual_rate is a NUMERIC(9, 6) fraction, which holds four decimals of a
	// percentage; withholding_rate is a NUMERIC(5, 4) one, which holds two.
	annualRatePctDecimals  = 4
	withholdingPctDecimals = 2
	// A tier's from_balance is a NUMERIC(20, 8): twelve digits before the point
	// and eight after.
	tierBalanceDecimals = 8
	// maxCashRateTiers is how many steps a version takes above the rate it pays
	// from zero. Platforms quote two or three.
	maxCashRateTiers = 10

	// cashRateDateGrace is how many days before the server's a rate may start
	// or end. Days are UTC, as they are for snapshots, and an owner west of
	// Greenwich is still on the previous one for part of the evening: at 7 p.m.
	// in Bogotá it is already tomorrow in UTC.
	cashRateDateGrace = 1
)

// CashRate is one version of the rate a cash account earns.
type CashRate struct {
	ID         uuid.UUID      `json:"id"`
	SourceID   uuid.UUID      `json:"sourceId"`
	SourceName string         `json:"sourceName"`
	Currency   money.Currency `json:"currency"`
	// PocketID is the pocket of the account the rate belongs to, nil for the
	// main account (000047). A platform can pay one rate on its account and
	// another on its pocket, and each earns on its own balances.
	PocketID   *uuid.UUID `json:"pocketId"`
	PocketName string     `json:"pocketName"`
	// AnnualRatePct is the effective annual rate as a percentage: "9.25" is
	// 9.25 % E.A. WithholdingPct is the share of the interest withheld as tax,
	// also as a percentage.
	AnnualRatePct  string          `json:"annualRatePct"`
	WithholdingPct string          `json:"withholdingPct"`
	Posting        InterestPosting `json:"posting"`
	// Tiers are the steps above AnnualRatePct, lowest first: the part of the
	// account above a step's FromBalance earns its rate, up to the next step. A
	// step at 0 is a cap. They belong to the account, so the balances of one
	// account share them. Empty, never nil, when the account earns AnnualRatePct
	// on all of it.
	Tiers         []CashRateTier `json:"tiers"`
	EffectiveFrom time.Time      `json:"effectiveFrom"`
	// EndedOn is the last day the version earns, nil while it has no end.
	EndedOn *time.Time `json:"endedOn"`
	// Latest is whether this is the newest version of its account: the only one
	// that can still be ended, and corrected or deleted while AccruedThrough is
	// nil.
	Latest bool `json:"latest"`
	// AccruedThrough is the last day whose interest was computed at this
	// version, nil if none has been.
	AccruedThrough *time.Time `json:"accruedThrough"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

// InEffectOn reports whether this version is the one its account earns at on
// the given day. Only the newest version of an account can still be running, so
// a version a later one follows never is, whatever its dates say.
func (r CashRate) InEffectOn(day time.Time) bool {
	day = cashRateDay(day)

	switch {
	case !r.Latest, cashRateDay(r.EffectiveFrom).After(day):
		return false
	case r.EndedOn != nil && cashRateDay(*r.EndedOn).Before(day):
		return false
	default:
		return true
	}
}

// CashRateTier is a step of a rate, both figures as text like the rate's own:
// what the account holds from which it applies, and the effective annual rate
// on the part of the account in it, as a percentage.
type CashRateTier struct {
	FromBalance   string `json:"fromBalance"`
	AnnualRatePct string `json:"annualRatePct"`
}

// EffectiveAnnualPct is the rate an account earns on everything it holds when
// it holds held, in its currency, as a percentage. Without tiers it is the
// version's own rate. With them it is what the steps come to together,
// compounded the way the ledger compounds a day: the rate a day is kept at.
func (r CashRate) EffectiveAnnualPct(held decimal.Decimal) (decimal.Decimal, error) {
	version, err := r.ledgerVersion()
	if err != nil {
		return decimal.Zero, err
	}

	rate, err := version.effectiveRate(held)
	if err != nil {
		return decimal.Zero, err
	}

	return rate.Mul(fractionToPct), nil
}

// ledgerVersion is the version as the ledger computes with it: the rates as
// fractions.
func (r CashRate) ledgerVersion() (CashRateVersion, error) {
	annual, err := decimal.NewFromString(r.AnnualRatePct)
	if err != nil {
		return CashRateVersion{}, err
	}

	version := CashRateVersion{ID: r.ID, AnnualRate: annual.Mul(pctToFraction), Posting: r.Posting}

	for _, tier := range r.Tiers {
		from, err := decimal.NewFromString(tier.FromBalance)
		if err != nil {
			return CashRateVersion{}, err
		}

		rate, err := decimal.NewFromString(tier.AnnualRatePct)
		if err != nil {
			return CashRateVersion{}, err
		}

		version.Tiers = append(version.Tiers, CashRateStep{From: from, AnnualRate: rate.Mul(pctToFraction)})
	}

	return version, nil
}

// CashRateInput is what a version of a rate states, and what a correction can
// rewrite.
type CashRateInput struct {
	AnnualRatePct  decimal.Decimal
	WithholdingPct decimal.Decimal
	Posting        InterestPosting
	// Tiers are the steps above AnnualRatePct, lowest first. A correction states
	// the version whole, so leaving them out removes them.
	Tiers []CashRateTierInput
}

// CashRateTierInput is a step as the owner states it: from what the account
// holds, the effective annual rate on the part above it, as a percentage.
type CashRateTierInput struct {
	FromBalance   decimal.Decimal
	AnnualRatePct decimal.Decimal
}

// NewCashRateInput is a new version: the values, and the account and day they
// apply from.
type NewCashRateInput struct {
	SourceID uuid.UUID
	Currency money.Currency
	// PocketID is which drawer of the account earns it, the zero UUID being the
	// main one. Versions of one pocket close each other; the main account's and
	// a pocket's never meet.
	PocketID      uuid.UUID
	EffectiveFrom time.Time
	CashRateInput
}

func invalidCashRate(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalidCashRate, fmt.Sprintf(format, args...))
}

// cashRateInUse refuses a change that reaches a day already computed at the
// rate. instead says what can be done, and from which day when there is one.
func cashRateInUse(accruedThrough time.Time, instead string) error {
	return fmt.Errorf("%w: interest through %s was computed at it; %s", ErrCashRateInUse, accruedThrough.Format(time.DateOnly), instead)
}

// cashRateDay is t's calendar day at midnight UTC, the form every rate date is
// stored and compared in.
func cashRateDay(t time.Time) time.Time {
	u := t.UTC()
	return time.Date(u.Year(), u.Month(), u.Day(), 0, 0, 0, 0, time.UTC)
}

// notBeforeGrace refuses a date earlier than the grace allows. Past days are
// not recomputed, so a rate cannot start or stop in them.
func notBeforeGrace(field string, day, today time.Time) error {
	earliest := cashRateDay(today).AddDate(0, 0, -cashRateDateGrace)
	if cashRateDay(day).Before(earliest) {
		return invalidCashRate("%s cannot be before %s: past days are not recomputed", field, earliest.Format(time.DateOnly))
	}

	return nil
}

// Validate checks the values of a version, which a correction rewrites.
func (in CashRateInput) Validate() error {
	return in.validateValues(false)
}

// validateValues is Validate with the one rule that depends on where the
// version lands: at_maturity is only offered where there is a maturity to
// credit on, which is a fixed deposit (000048). An account's own rate is
// refused it, because nothing would ever make the credit land.
func (in CashRateInput) validateValues(atMaturity bool) error {
	if !in.AnnualRatePct.IsPos() || in.AnnualRatePct.GreaterThan(maxAnnualRatePct) {
		return invalidCashRate("annualRatePct must be greater than 0 and at most 100")
	}

	if !in.AnnualRatePct.Equal(in.AnnualRatePct.Trunc(annualRatePctDecimals)) {
		return invalidCashRate("annualRatePct takes at most %d decimals", annualRatePctDecimals)
	}

	if in.WithholdingPct.IsNeg() || !in.WithholdingPct.LessThan(maxWithholdingPct) {
		return invalidCashRate("withholdingPct must be at least 0 and less than 100")
	}

	if !in.WithholdingPct.Equal(in.WithholdingPct.Trunc(withholdingPctDecimals)) {
		return invalidCashRate("withholdingPct takes at most %d decimals", withholdingPctDecimals)
	}

	switch in.Posting {
	case PostingDaily, PostingMonthly:
	case PostingAtMaturity:
		if !atMaturity {
			return invalidCashRate("posting must be one of: daily, monthly")
		}
	default:
		if atMaturity {
			return invalidCashRate("posting must be one of: daily, monthly, at_maturity")
		}

		return invalidCashRate("posting must be one of: daily, monthly")
	}

	if len(in.Tiers) > maxCashRateTiers {
		return invalidCashRate("tiers take at most %d steps", maxCashRateTiers)
	}

	for i, tier := range in.Tiers {
		if !tier.FromBalance.IsPos() || !tier.FromBalance.LessThan(maxCashRateBalance) {
			return invalidCashRate("tier fromBalance must be greater than 0 and less than %s", maxCashRateBalance)
		}

		if !tier.FromBalance.Equal(tier.FromBalance.Trunc(tierBalanceDecimals)) {
			return invalidCashRate("tier fromBalance takes at most %d decimals", tierBalanceDecimals)
		}

		// In order, so the ledger reads each step's end as the next one's start.
		if i > 0 && !tier.FromBalance.GreaterThan(in.Tiers[i-1].FromBalance) {
			return invalidCashRate("tiers must go up: each fromBalance above the one before")
		}

		if tier.AnnualRatePct.IsNeg() || tier.AnnualRatePct.GreaterThan(maxAnnualRatePct) {
			return invalidCashRate("tier annualRatePct must be at least 0 and at most 100")
		}

		if !tier.AnnualRatePct.Equal(tier.AnnualRatePct.Trunc(annualRatePctDecimals)) {
			return invalidCashRate("tier annualRatePct takes at most %d decimals", annualRatePctDecimals)
		}
	}

	return nil
}

// ValidateNew is Validate plus what only a new version states: the account, a
// currency the app can convert, and a first day that is not in the past.
func (in NewCashRateInput) ValidateNew(today time.Time) error {
	if in.SourceID == (uuid.UUID{}) {
		return invalidCashRate("sourceId is required")
	}

	if !currency.IsSupported(in.Currency) {
		return invalidCashRate("currency must be one of: %s", currency.List())
	}

	if in.EffectiveFrom.IsZero() {
		return invalidCashRate("effectiveFrom is required")
	}

	if err := notBeforeGrace("effectiveFrom", in.EffectiveFrom, today); err != nil {
		return err
	}

	return in.CashRateInput.Validate()
}

// ValidateCashRateEnd checks the first day a rate stops earning.
func ValidateCashRateEnd(endsOn, today time.Time) error {
	if endsOn.IsZero() {
		return invalidCashRate("endsOn is required")
	}

	return notBeforeGrace("endsOn", endsOn, today)
}
