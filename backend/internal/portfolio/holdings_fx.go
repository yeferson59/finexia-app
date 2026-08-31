package portfolio

// Expressing a portfolio's positions in one currency.
//
// A position carries two currencies that need not agree with each other or with
// the portfolio: the cost was settled in cost_currency, and the market price is
// quoted in the asset's own currency. Adding those raw numbers — a EUR holding
// on top of a USD one — produces a total that means nothing, and subtracting a
// EUR market value from a USD cost basis produces a return that means less. So
// the totals are converted here, once, before they leave the service.
//
// What is converted is deliberate: amounts, never unit prices. money.Convert
// rounds to the target currency's precision (two decimals for USD), which is
// right for a total and destructive for the price of a fractional share or a
// crypto lot. Per-unit figures stay native and the client formats them with
// their own currency.

import (
	"context"

	"uuid"

	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

// valueEntriesInBase fills CostBasisBase, MarketValueBase and FXConverted on
// every entry, converting each position's totals into baseCurrency.
//
// It never fails. Rates are BYO-key data: a user whose key has not fetched
// EUR→USD yet simply has no rate, and that must not take down the portfolio
// page. Such an entry keeps its native totals and is marked FXConverted=false,
// which is the client's cue to show the amounts unconverted and say so.
func (s *Service) valueEntriesInBase(ctx context.Context, userID uuid.UUID, baseCurrency money.Currency, entries []Entry) []Entry {
	// One rate lookup per distinct source currency, not per position: a
	// portfolio with twenty EUR holdings resolves EUR→USD once.
	rates := make(map[money.Currency]decimal.Decimal, 2)
	missing := make(map[money.Currency]struct{}, 2)

	rateFor := func(from money.Currency) (decimal.Decimal, bool) {
		if rate, seen := rates[from]; seen {
			return rate, true
		}
		if _, seen := missing[from]; seen {
			return decimal.Decimal{}, false
		}

		rate, err := s.GetConversionRate(ctx, userID, from, baseCurrency)
		if err != nil {
			missing[from] = struct{}{}
			return decimal.Decimal{}, false
		}

		rates[from] = rate

		return rate, true
	}

	for i := range entries {
		cost, marketValue := entryTotals(entries[i])

		costBase, costOK := convertTotal(cost, baseCurrency, rateFor)
		marketBase, marketOK := convertTotal(marketValue, baseCurrency, rateFor)

		entries[i].CostBasisBase = costBase
		entries[i].MarketValueBase = marketBase
		entries[i].FXConverted = costOK && marketOK
	}

	if len(missing) > 0 {
		s.log.Warn(ctx, "no exchange rate for some holdings; totals left unconverted",
			logger.Str("baseCurrency", baseCurrency.String()),
			logger.Int("currencies", len(missing)),
		)
	}

	return entries
}

// entryTotals returns the position's cost basis and market value, each still in
// its own currency: cost in the currency the purchase settled in, market value
// in the asset's quote currency. Without a market price the position is valued
// at cost, which is the same convention PriceSourceCost describes.
func entryTotals(entry Entry) (cost, marketValue money.Money) {
	cost = entry.Price.MulDecimal(entry.Quantity)

	if entry.Asset.CurrentPrice == nil {
		return cost, cost
	}

	return cost, entry.Asset.CurrentPrice.MulDecimal(entry.Quantity)
}

// convertTotal moves one amount into the base currency, reporting whether it
// got there. On a missing or unusable rate the amount is returned untouched, in
// its own currency.
func convertTotal(amount money.Money, base money.Currency, rateFor func(money.Currency) (decimal.Decimal, bool)) (money.Money, bool) {
	if amount.GetCurrency() == base {
		return amount, true
	}

	from := amount.GetCurrency()

	rate, ok := rateFor(from)
	if !ok {
		return amount, false
	}

	converted, err := amount.Convert(base, rate)
	if err != nil {
		return amount, false
	}

	return converted, true
}
