package mcp

import (
	"context"

	"uuid"

	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/market"
	"github.com/yeferson59/finexia-app/internal/portfolio"
)

// The two interfaces below are this module's whole view of the rest of the
// application, declared here rather than imported from the modules that
// implement them: the consumer owns the interface, the composition root does
// the matching. They are read-only by construction — no method on either can
// change anything — which is what makes the package-level promise in the doc
// comment enforceable rather than a convention.

// PortfolioReader is the slice of the portfolio module the tools serve.
// Satisfied by *portfolio.service.
//
// Every method takes the user id explicitly: the identity comes off the bearer
// token at the edge and is threaded down, so a tool cannot read a portfolio
// that is not the caller's even by mistake.
type PortfolioReader interface {
	GetPortfoliosSummary(ctx context.Context, userID uuid.UUID) ([]portfolio.SummaryView, error)
	GetPortfoliosSummaryInCurrency(ctx context.Context, userID uuid.UUID, targetCurrency money.Currency) ([]portfolio.SummaryView, error)
	GetAssetHoldings(ctx context.Context, userID uuid.UUID, targetCurrency money.Currency) ([]portfolio.AssetHolding, error)
	GetAssetAllocation(ctx context.Context, userID uuid.UUID, targetCurrency money.Currency) ([]portfolio.AllocationItem, error)
	GetRecentUserTransactions(ctx context.Context, userID uuid.UUID, limit int) ([]portfolio.Transaction, error)
	GetPortfolioGrowth(ctx context.Context, userID uuid.UUID, currency money.Currency, period string) ([]portfolio.GrowthPoint, portfolio.GrowthSummary, error)
	GetPlatforms(ctx context.Context, userID uuid.UUID, displayCurrency money.Currency) ([]portfolio.PlatformStats, error)
}

// MarketReader is the slice of the market module the tools serve: the asset
// catalog and the shared exchange rates. Satisfied by *market.service.
//
// The catalog reads take a market.CatalogView because assets can be
// contributed by users, so "the catalog" is not one list — each caller sees the
// curated rows plus their own, and an admin sees everything.
type MarketReader interface {
	GetAssets(ctx context.Context, view market.CatalogView, offset, limit uint) ([]market.Asset, error)
	SearchAssets(ctx context.Context, view market.CatalogView, search string, offset, limit uint) ([]market.Asset, error)
	GetLatestExchangeRates(ctx context.Context) ([]market.ExchangeRate, error)
}
