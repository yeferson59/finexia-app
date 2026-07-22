package market

import (
	"time"

	"github.com/google/uuid"
	"github.com/yeferson59/gofinance/v2/money"
)

type ExchangeRate struct {
	ID           uuid.UUID     `json:"id"`
	FromCurrency string        `json:"fromCurrency"`
	ToCurrency   string        `json:"toCurrency"`
	Rate         money.Decimal `json:"rate"`
	RateDate     time.Time     `json:"rateDate"`
	CreatedAt    time.Time     `json:"createdAt"`
	UpdatedAt    time.Time     `json:"updatedAt"`
}
