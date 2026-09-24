package portfolio

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"uuid"

	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/platform/currency"
	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

// An investment fund is money that earns whatever it turns out to earn: a FIC,
// a voluntary pension fund, a neobank's money-market pocket. Nobody promises it
// a rate. The manager publishes the value of a unit every day, and the return
// is how that value moved — which can be down.
//
// So it is not a cash balance. It is a position of units, like a share, whose
// price the owner writes down from the statement because no market-data
// provider lists it: a mark. The latest mark is copied to user_asset_prices,
// where every valuation already looks first, so the summary, the holdings and
// the snapshot job see a position priced by its owner and need nothing new.
// Buying units is a purchase and selling them a sale, through the transaction
// endpoints every position uses.
//
// What a fund adds is the history of its marks — the return it publishes is a
// ratio of two of them — and the snapshots that a late mark revalues. See
// migrations 000056 and 000057, and docs/PLAN_FONDOS_INVERSION.md.

var (
	// ErrFundNotFound answers for a fund that does not exist or that the user
	// does not follow. The two are the same thing to these endpoints.
	ErrFundNotFound = httpx.AsNotFound(errors.New("fund not found"))
	// ErrFundMarkNotFound answers for a date the fund has no mark on.
	ErrFundMarkNotFound = httpx.AsNotFound(errors.New("fund mark not found"))
	// ErrInvalidFund rejects a fund that cannot be recorded as stated. It is
	// wrapped with the rule that was broken.
	ErrInvalidFund = httpx.AsBadRequest(errors.New("invalid fund"))
	// ErrInvalidFundMark rejects a mark that cannot be recorded as stated. It is
	// wrapped with the rule that was broken.
	ErrInvalidFundMark = httpx.AsBadRequest(errors.New("invalid fund mark"))
	// ErrFundHasPositions refuses to stop following a fund some portfolio still
	// holds. It is a conflict: the same deletion is fine once its positions are
	// deleted where every position is.
	ErrFundHasPositions = httpx.AsConflict(errors.New("the fund still has positions"))
)

// FundTracking is how the owner follows a fund, chosen when it is created.
type FundTracking string

const (
	// FundUnits is a fund whose statement shows units and the unit value, both
	// written as they appear there.
	FundUnits FundTracking = "units"
	// FundBalance is a fund whose app only shows a balance. Its units are
	// synthetic; it is not open yet (docs/PLAN_FONDOS_INVERSION.md, phase 2).
	FundBalance FundTracking = "balance"
)

// IsValid reports whether the tracking is one the schema knows.
func (t FundTracking) IsValid() bool {
	return t == FundUnits || t == FundBalance
}

const (
	// maxFundNameLen keeps a fund's name inside what a screen can show. The
	// column (assets.name) takes 255.
	maxFundNameLen = 100
	// maxFundMarkYears is how far back a mark can be dated. A voluntary pension
	// fund can be decades old, but a statement nobody keeps is a mistyped year.
	maxFundMarkYears = 50
	// fundTickerPrefix starts the ticker a fund is created under. The ticker is
	// generated, not typed: a FIC has no symbol, and one the owner made up could
	// collide with a listed share the price sync would then quote.
	fundTickerPrefix = "FND-"
	// fundPriceSource is what user_asset_prices.source says of a price copied
	// from a mark: the owner wrote it, no provider fetched it.
	fundPriceSource = "user"
)

// maxFundUnitValue keeps a unit value inside NUMERIC(20, 8).
var maxFundUnitValue = decimal.MustFromString("1000000000000")

// Fund is one fund the user follows, with what every portfolio holds of it.
type Fund struct {
	AssetID  uuid.UUID      `json:"assetId"`
	Ticker   string         `json:"ticker"`
	Name     string         `json:"name"`
	Currency money.Currency `json:"currency"`
	Tracking FundTracking   `json:"tracking"`
	// Units, Cost and Value add up every position, in Currency. Value is the
	// units at the latest mark, or at their cost while there is none — which is
	// what PricedAtCost says, so a gain of zero is not read as a flat fund.
	Units        string `json:"units"`
	Cost         string `json:"cost"`
	Value        string `json:"value"`
	PricedAtCost bool   `json:"pricedAtCost"`
	// UnitValue and ValuedOn are the latest mark, nil until there is one.
	UnitValue *string    `json:"unitValue"`
	ValuedOn  *time.Time `json:"valuedOn"`
	// Marks is how many marks it has.
	Marks     int64          `json:"marks"`
	Positions []FundPosition `json:"positions"`
	CreatedAt time.Time      `json:"createdAt"`
}

// FundPosition is what one portfolio holds of a fund, on one platform.
type FundPosition struct {
	EntryID       uuid.UUID `json:"entryId"`
	PortfolioID   uuid.UUID `json:"portfolioId"`
	PortfolioName string    `json:"portfolioName"`
	SourceID      uuid.UUID `json:"sourceId"`
	SourceName    string    `json:"sourceName"`
	Units         string    `json:"units"`
	// Cost is what the units cost, in the fund's currency.
	Cost string `json:"cost"`
}

// FundMark is what a fund was worth on a day.
type FundMark struct {
	Date      time.Time `json:"date"`
	UnitValue string    `json:"unitValue"`
	// Balance is the balance a FundBalance fund was marked with; nil in units.
	Balance   *string   `json:"balance"`
	Notes     string    `json:"notes"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func invalidFund(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalidFund, fmt.Sprintf(format, args...))
}

func invalidFundMark(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalidFundMark, fmt.Sprintf(format, args...))
}

// NewFundInput is a fund as the owner states it the first time: what it is,
// where it is held, and the first purchase of units — plus, optionally, what a
// unit is worth now, so a fund bought months ago does not open at cost.
type NewFundInput struct {
	PortfolioID uuid.UUID
	SourceID    uuid.UUID
	Name        string
	Currency    money.Currency
	Tracking    FundTracking
	// Date, Units and UnitValue are the first purchase.
	Date      time.Time
	Units     decimal.Decimal
	UnitValue decimal.Decimal
	// CurrentUnitValue, when positive, is a mark on CurrentDate — today when
	// that is left out.
	CurrentUnitValue decimal.Decimal
	CurrentDate      time.Time
	// PayFromCash and CashPocketID pay the purchase out of the platform's cash,
	// as any purchase can be (TransactionInput).
	PayFromCash  bool
	CashPocketID uuid.UUID
	Notes        string
}

// CleanName is the name as it is stored.
func (in NewFundInput) CleanName() string {
	return strings.TrimSpace(in.Name)
}

// withDefaults dates the current mark today when one was given without a day.
func (in NewFundInput) withDefaults(today time.Time) NewFundInput {
	if in.CurrentUnitValue.IsPos() && in.CurrentDate.IsZero() {
		in.CurrentDate = cashRateDay(today)
	}

	return in
}

// Validate checks what the fund states. today is the server's clock. The
// portfolio and the platform are checked where they can be locked, and the
// purchase by TransactionInput, like every other purchase.
func (in NewFundInput) Validate(today time.Time) error {
	if in.PortfolioID == (uuid.UUID{}) || in.SourceID == (uuid.UUID{}) {
		return invalidFund("portfolioId and sourceId are required")
	}

	name := in.CleanName()
	if name == "" {
		return invalidFund("name is required")
	}

	if utf8.RuneCountInString(name) > maxFundNameLen {
		return invalidFund("name cannot exceed %d characters", maxFundNameLen)
	}

	if !currency.IsSupported(in.Currency) {
		return invalidFund("currency must be one of: %s", currency.List())
	}

	switch in.Tracking {
	case FundUnits:
	case FundBalance:
		return invalidFund("tracking by balance is not available yet; follow the fund by units")
	default:
		return invalidFund("tracking must be units or balance")
	}

	if err := validateFundDate(in.Date, today, invalidFund); err != nil {
		return err
	}

	if !in.Units.IsPos() {
		return invalidFund("units must be greater than zero")
	}

	if err := validateUnitValue(in.UnitValue, invalidFund); err != nil {
		return err
	}

	if in.CurrentUnitValue.IsNeg() {
		return invalidFund("currentUnitValue cannot be negative")
	}

	if in.CurrentUnitValue.IsPos() {
		if err := validateUnitValue(in.CurrentUnitValue, invalidFund); err != nil {
			return err
		}

		if err := validateFundDate(in.CurrentDate, today, invalidFund); err != nil {
			return err
		}

		if cashRateDay(in.CurrentDate).Before(cashRateDay(in.Date)) {
			return invalidFund("the current unit value cannot be dated before the purchase")
		}
	}

	return nil
}

// purchase is the first purchase as the transaction every position opens with.
func (in NewFundInput) purchase() TransactionInput {
	return TransactionInput{
		Type:            Buy,
		Quantity:        in.Units,
		Price:           money.NewFromDecimal(in.UnitValue, in.Currency),
		Currency:        in.Currency,
		TransactionDate: cashRateDay(in.Date),
		Notes:           in.Notes,
		PayFromCash:     in.PayFromCash,
		CashPocketID:    in.CashPocketID,
	}
}

// FundMarkInput is what a fund was worth on a day, as the owner read it.
type FundMarkInput struct {
	Date      time.Time
	UnitValue decimal.Decimal
	// Balance is what a FundBalance fund is marked with. It is refused on a
	// fund followed by units, whose statement already says the unit value.
	Balance decimal.Decimal
	Notes   string
}

// Validate checks the mark against how the fund is followed. today is the
// server's clock.
func (in FundMarkInput) Validate(today time.Time, tracking FundTracking) error {
	if err := validateFundDate(in.Date, today, invalidFundMark); err != nil {
		return err
	}

	if tracking != FundUnits {
		return invalidFundMark("only funds followed by units take marks for now")
	}

	if !in.Balance.IsZero() {
		return invalidFundMark("a fund followed by units is marked with its unit value, not a balance")
	}

	if err := validateUnitValue(in.UnitValue, invalidFundMark); err != nil {
		return err
	}

	if utf8.RuneCountInString(in.Notes) > maxCashNotesLen {
		return invalidFundMark("notes cannot exceed %d characters", maxCashNotesLen)
	}

	return nil
}

func validateFundDate(date, today time.Time, invalid func(string, ...any) error) error {
	day := cashRateDay(today)

	switch {
	case date.IsZero():
		return invalid("date is required")
	case cashRateDay(date).After(day):
		return invalid("date cannot be in the future")
	case cashRateDay(date).Before(day.AddDate(-maxFundMarkYears, 0, 0)):
		return invalid("date cannot be more than %d years ago", maxFundMarkYears)
	}

	return nil
}

func validateUnitValue(v decimal.Decimal, invalid func(string, ...any) error) error {
	if !v.IsPos() || !v.LessThan(maxFundUnitValue) {
		return invalid("unit value must be greater than 0 and less than %s", maxFundUnitValue)
	}

	return nil
}

// newFundTicker is the ticker a fund is created under: the prefix and eight
// hex digits, which fits assets.ticker VARCHAR(20) and no listed symbol.
func newFundTicker() string {
	id := uuid.New()

	return fundTickerPrefix + strings.ToUpper(strings.ReplaceAll(id.String(), "-", "")[:8])
}
