package market

import (
	"time"

	"uuid"

	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"
)

type ExchangeRate struct {
	ID           uuid.UUID       `json:"id"`
	FromCurrency money.Currency  `json:"fromCurrency"`
	ToCurrency   money.Currency  `json:"toCurrency"`
	Rate         decimal.Decimal `json:"rate"`
	RateDate     time.Time       `json:"rateDate"`
	// Source says who put this number here: ManualRateSource for a rate an
	// operator typed or imported, or the feed that published it. Two writers
	// now share the table and a stale hand-entered rate looks exactly like a
	// fresh automatic one without it.
	Source    ProviderID `json:"source"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

// ManualRateSource marks a shared rate an operator entered by hand, through
// POST /exchange-rates, the spreadsheet import or PATCH /exchange-rates/:id.
// It is the column default, so every row that predates the public feed carries
// it.
//
// It is not a ProviderID: no provider produced it. The column holds both
// because provenance is one question with two kinds of answer.
const ManualRateSource = "manual"
