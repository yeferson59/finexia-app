package portfolio

import (
	"time"

	"uuid"

	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"
)

// The rate DTOs carry no validate tags, for the reason the cash DTOs give:
// CashRateInput answers every rule with a message of its own.

type CreateCashRateRequestDTO struct {
	SourceID uuid.UUID      `json:"sourceId"`
	Currency money.Currency `json:"currency"`
	// PocketID is which drawer of the account earns it; omitted or null is the
	// main one (000047).
	PocketID       uuid.UUID         `json:"pocketId"`
	AnnualRatePct  decimal.Decimal   `json:"annualRatePct"`
	WithholdingPct decimal.Decimal   `json:"withholdingPct"`
	Posting        string            `json:"posting"`
	Tiers          []CashRateTierDTO `json:"tiers"`
	MaxBalance     *decimal.Decimal  `json:"maxBalance"`
	EffectiveFrom  time.Time         `json:"effectiveFrom"`
}

func (d CreateCashRateRequestDTO) Input() NewCashRateInput {
	return NewCashRateInput{
		SourceID:      d.SourceID,
		Currency:      d.Currency,
		PocketID:      d.PocketID,
		EffectiveFrom: d.EffectiveFrom,
		CashRateInput: cashRateValues(d.AnnualRatePct, d.WithholdingPct, d.Posting, d.Tiers, d.MaxBalance),
	}
}

// UpdateCashRateRequestDTO has no account and no dates: a correction rewrites
// what a version says, not when it applies.
type UpdateCashRateRequestDTO struct {
	AnnualRatePct  decimal.Decimal   `json:"annualRatePct"`
	WithholdingPct decimal.Decimal   `json:"withholdingPct"`
	Posting        string            `json:"posting"`
	Tiers          []CashRateTierDTO `json:"tiers"`
	MaxBalance     *decimal.Decimal  `json:"maxBalance"`
}

func (d UpdateCashRateRequestDTO) Input() CashRateInput {
	return cashRateValues(d.AnnualRatePct, d.WithholdingPct, d.Posting, d.Tiers, d.MaxBalance)
}

// CashRateTierDTO is a step of a rate: from what the account holds, the rate
// on the part above it.
type CashRateTierDTO struct {
	FromBalance   decimal.Decimal `json:"fromBalance"`
	AnnualRatePct decimal.Decimal `json:"annualRatePct"`
}

// RecalculateCashRequestDTO asks for an account's interest to be computed again
// from a day, on what its balances hold now. pocketId names which drawer of the
// account; omitted or null is the main one.
type RecalculateCashRequestDTO struct {
	SourceID uuid.UUID      `json:"sourceId"`
	Currency money.Currency `json:"currency"`
	PocketID uuid.UUID      `json:"pocketId"`
	From     time.Time      `json:"from"`
}

func (d RecalculateCashRequestDTO) Input() RecalculateCashInterestInput {
	return RecalculateCashInterestInput(d)
}

// EndCashRateRequestDTO names the first day the rate stops earning.
type EndCashRateRequestDTO struct {
	EndsOn time.Time `json:"endsOn"`
}

// cashRateValues reads the values of a version. An omitted posting is daily,
// and omitted tiers are none: a correction states the version whole, so one
// that leaves them out removes them.
//
// maxBalance is how a cap was sent before tiers (000046). It is still read, so
// a caller that sends one keeps its cap rather than losing it in silence: it
// becomes the step at 0 % a cap now is, after the tiers it comes with.
func cashRateValues(rate, withholding decimal.Decimal, posting string, tiers []CashRateTierDTO, maxBalance *decimal.Decimal) CashRateInput {
	p := InterestPosting(posting)
	if p == "" {
		p = PostingDaily
	}

	steps := make([]CashRateTierInput, 0, len(tiers)+1)
	for _, tier := range tiers {
		steps = append(steps, CashRateTierInput{FromBalance: tier.FromBalance, AnnualRatePct: tier.AnnualRatePct})
	}

	if maxBalance != nil {
		steps = append(steps, CashRateTierInput{FromBalance: *maxBalance, AnnualRatePct: decimal.Zero})
	}

	return CashRateInput{AnnualRatePct: rate, WithholdingPct: withholding, Posting: p, Tiers: steps}
}
