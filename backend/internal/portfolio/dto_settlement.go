package portfolio

import (
	"uuid"

	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"
)

// ChangeSettlementRequestDTO restates a position in the currency its account
// really settled in. Rates carries the rate of each transaction quoted in a
// currency other than CostCurrency; one quoted in CostCurrency settles at 1 and
// needs no entry.
type ChangeSettlementRequestDTO struct {
	CostCurrency money.Currency      `json:"costCurrency" validate:"required"`
	Rates        []SettlementRateDTO `json:"rates"`
}

// SettlementRateDTO is the rate one transaction converted at, as its broker
// confirmation shows it: how much of CostCurrency one unit of the transaction's
// currency bought that day.
type SettlementRateDTO struct {
	TransactionID uuid.UUID       `json:"transactionId"`
	FXRate        decimal.Decimal `json:"fxRate"`
}

// RatesByTransaction indexes Rates for the service. A transaction listed twice
// keeps the last rate given.
func (d ChangeSettlementRequestDTO) RatesByTransaction() map[uuid.UUID]decimal.Decimal {
	rates := make(map[uuid.UUID]decimal.Decimal, len(d.Rates))
	for _, rate := range d.Rates {
		rates[rate.TransactionID] = rate.FXRate
	}

	return rates
}

// ChangeSettlementResponseDTO says what the position now costs in, and how many
// of its transactions were restated.
type ChangeSettlementResponseDTO struct {
	CostCurrency string `json:"costCurrency"`
	Transactions int    `json:"transactions"`
}
