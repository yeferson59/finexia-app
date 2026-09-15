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
// effect — is described in migration 000043.
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
)

var (
	maxAnnualRatePct  = decimal.MustFromString("100")
	maxWithholdingPct = decimal.MustFromString("100")
	// maxCashRateBalance is what NUMERIC(20, 8) holds, exclusive.
	maxCashRateBalance = decimal.MustFromString("1000000000000")
)

const (
	// annual_rate is a NUMERIC(9, 6) fraction, which holds four decimals of a
	// percentage; withholding_rate is a NUMERIC(5, 4) one, which holds two.
	annualRatePctDecimals  = 4
	withholdingPctDecimals = 2
	// max_balance is a NUMERIC(20, 8): twelve digits before the point and
	// eight after.
	maxBalanceDecimals = 8

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
	// AnnualRatePct is the effective annual rate as a percentage: "9.25" is
	// 9.25 % E.A. WithholdingPct is the share of the interest withheld as tax,
	// also as a percentage.
	AnnualRatePct  string          `json:"annualRatePct"`
	WithholdingPct string          `json:"withholdingPct"`
	Posting        InterestPosting `json:"posting"`
	// MaxBalance is the most the account earns on, nil when it earns on all of
	// it. It belongs to the account, so the balances of one account share it.
	MaxBalance    *string   `json:"maxBalance"`
	EffectiveFrom time.Time `json:"effectiveFrom"`
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

// CashRateInput is what a version of a rate states, and what a correction can
// rewrite.
type CashRateInput struct {
	AnnualRatePct  decimal.Decimal
	WithholdingPct decimal.Decimal
	Posting        InterestPosting
	// MaxBalance is the cap, nil for none. A correction states the version
	// whole, so leaving it out removes the cap.
	MaxBalance *decimal.Decimal
}

// NewCashRateInput is a new version: the values, and the account and day they
// apply from.
type NewCashRateInput struct {
	SourceID      uuid.UUID
	Currency      money.Currency
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
	default:
		return invalidCashRate("posting must be one of: daily, monthly")
	}

	if limit := in.MaxBalance; limit != nil {
		if !limit.IsPos() || !limit.LessThan(maxCashRateBalance) {
			return invalidCashRate("maxBalance must be greater than 0 and less than %s", maxCashRateBalance)
		}

		if !limit.Equal(limit.Trunc(maxBalanceDecimals)) {
			return invalidCashRate("maxBalance takes at most %d decimals", maxBalanceDecimals)
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
