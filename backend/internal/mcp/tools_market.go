package mcp

import (
	"context"
	"strings"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/yeferson59/finexia-app/internal/market"
)

const (
	defaultAssetLimit = 25
	maxAssetLimit     = 100
)

// addMarketTools registers the two reads that are about the market rather than
// about the caller: the asset catalog and the shared exchange rates.
//
// The catalog is still scoped to the caller — users can contribute assets, and
// a contributed row is visible only to whoever contributed it — which is what
// the CatalogView carries. Admins moderate those contributions, so they see
// everything, exactly as on the REST route.
func (m *Module) addMarketTools(s *mcpsdk.Server, c caller) {
	view := market.CatalogView{ViewerID: c.userID, All: c.isAdmin()}

	readTool(s, "search_assets", "Search assets",
		"Search the asset catalog by ticker or name. This is the catalog, not the user's positions: a match here says the asset exists and can be held, not that the user holds it.",
		func(ctx context.Context, in AssetSearchInput) (AssetsOutput, error) {
			limit := clampLimit(in.Limit, defaultAssetLimit, maxAssetLimit)
			query := strings.TrimSpace(in.Query)

			var (
				assets []market.Asset
				err    error
			)

			// An empty query is a listing, not a search for the empty string:
			// the search path matches against the text and would return
			// everything anyway, but through a LIKE the catalog has no reason
			// to run.
			if query == "" {
				assets, err = m.assets.GetAssets(ctx, view, 0, uint(limit))
			} else {
				assets, err = m.assets.SearchAssets(ctx, view, query, 0, uint(limit))
			}

			if err != nil {
				return AssetsOutput{}, m.logToolError(ctx, "search_assets", c, err)
			}

			return AssetsOutput{Assets: assetRows(assets)}, nil
		})

	readTool(s, "list_exchange_rates", "List exchange rates",
		"List the latest shared exchange rates the application converts with. These are the keyless public feeds (the ECB reference rates and Colombia's TRM), not a per-user market-data subscription.",
		func(ctx context.Context, _ EmptyInput) (RatesOutput, error) {
			rates, err := m.assets.GetLatestExchangeRates(ctx)
			if err != nil {
				return RatesOutput{}, m.logToolError(ctx, "list_exchange_rates", c, err)
			}

			return RatesOutput{Rates: rateRows(rates)}, nil
		})
}

func assetRows(assets []market.Asset) []Asset {
	out := make([]Asset, 0, len(assets))

	for _, a := range assets {
		row := Asset{
			ID:        a.ID.String(),
			Ticker:    a.Ticker,
			Name:      a.Name,
			AssetType: string(a.AssetType),
			Exchange:  a.Exchange,
			Currency:  a.Currency.String(),
			IsCurated: a.IsCurated,
		}

		// Both are nil for an asset nothing has priced yet, and a zero price
		// under a real currency would read as "worth nothing" rather than
		// "unknown".
		if a.CurrentPrice != nil {
			row.Price = a.CurrentPrice.String()
		}

		if a.PriceUpdatedAt != nil {
			row.PriceUpdatedAt = timeText(*a.PriceUpdatedAt)
		}

		out = append(out, row)
	}

	return out
}

func rateRows(rates []market.ExchangeRate) []ExchangeRate {
	out := make([]ExchangeRate, 0, len(rates))

	for _, r := range rates {
		out = append(out, ExchangeRate{
			From:   r.FromCurrency.String(),
			To:     r.ToCurrency.String(),
			Rate:   r.Rate.String(),
			Date:   timeText(r.RateDate),
			Source: string(r.Source),
		})
	}

	return out
}
