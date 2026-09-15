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
	SourceID       uuid.UUID       `json:"sourceId"`
	Currency       money.Currency  `json:"currency"`
	AnnualRatePct  decimal.Decimal `json:"annualRatePct"`
	WithholdingPct decimal.Decimal `json:"withholdingPct"`
	Posting        string          `json:"posting"`
	EffectiveFrom  time.Time       `json:"effectiveFrom"`
}

func (d CreateCashRateRequestDTO) Input() NewCashRateInput {
	return NewCashRateInput{
		SourceID:      d.SourceID,
		Currency:      d.Currency,
		EffectiveFrom: d.EffectiveFrom,
		CashRateInput: cashRateValues(d.AnnualRatePct, d.WithholdingPct, d.Posting),
	}
}

// UpdateCashRateRequestDTO has no account and no dates: a correction rewrites
// what a version says, not when it applies.
type UpdateCashRateRequestDTO struct {
	AnnualRatePct  decimal.Decimal `json:"annualRatePct"`
	WithholdingPct decimal.Decimal `json:"withholdingPct"`
	Posting        string          `json:"posting"`
}

func (d UpdateCashRateRequestDTO) Input() CashRateInput {
	return cashRateValues(d.AnnualRatePct, d.WithholdingPct, d.Posting)
}

// EndCashRateRequestDTO names the first day the rate stops earning.
type EndCashRateRequestDTO struct {
	EndsOn time.Time `json:"endsOn"`
}

// cashRateValues reads the values of a version. An omitted posting is daily,
// the only one there is so far.
func cashRateValues(rate, withholding decimal.Decimal, posting string) CashRateInput {
	p := InterestPosting(posting)
	if p == "" {
		p = PostingDaily
	}

	return CashRateInput{AnnualRatePct: rate, WithholdingPct: withholding, Posting: p}
}
