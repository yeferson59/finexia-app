package portfolio

import (
	"time"

	"uuid"

	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"
)

// The fund DTOs carry no validate tags, for the reason the cash ones give:
// every rule is checked by the input, with one message per broken rule.

// CreateFundRequestDTO is a fund as the owner states it the first time. date,
// units and unitValue are the first purchase; currentUnitValue and currentDate,
// both optional, what a unit is worth now.
type CreateFundRequestDTO struct {
	PortfolioID      uuid.UUID       `json:"portfolioId"`
	SourceID         uuid.UUID       `json:"sourceId"`
	Name             string          `json:"name"`
	Currency         money.Currency  `json:"currency"`
	Tracking         string          `json:"tracking"`
	Date             time.Time       `json:"date"`
	Units            decimal.Decimal `json:"units"`
	UnitValue        decimal.Decimal `json:"unitValue"`
	CurrentUnitValue decimal.Decimal `json:"currentUnitValue"`
	CurrentDate      time.Time       `json:"currentDate"`
	PayFromCash      bool            `json:"payFromCash"`
	CashPocketID     uuid.UUID       `json:"cashPocketId"`
	Notes            string          `json:"notes"`
}

func (d CreateFundRequestDTO) Input() NewFundInput {
	tracking := FundTracking(d.Tracking)
	// Omitted is the only kind there is today.
	if tracking == "" {
		tracking = FundUnits
	}

	return NewFundInput{
		PortfolioID:      d.PortfolioID,
		SourceID:         d.SourceID,
		Name:             d.Name,
		Currency:         d.Currency,
		Tracking:         tracking,
		Date:             d.Date,
		Units:            d.Units,
		UnitValue:        d.UnitValue,
		CurrentUnitValue: d.CurrentUnitValue,
		CurrentDate:      d.CurrentDate,
		PayFromCash:      d.PayFromCash,
		CashPocketID:     d.CashPocketID,
		Notes:            d.Notes,
	}
}

// SaveFundMarkRequestDTO is what a fund was worth on a day: its unit value, or
// — for a fund followed by balance — its balance.
type SaveFundMarkRequestDTO struct {
	Date      time.Time       `json:"date"`
	UnitValue decimal.Decimal `json:"unitValue"`
	Balance   decimal.Decimal `json:"balance"`
	Notes     string          `json:"notes"`
}

func (d SaveFundMarkRequestDTO) Input() FundMarkInput {
	return FundMarkInput(d)
}
