package portfolio

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/yeferson59/gofinance/v2/money"
)

// The persistence surface is split into cohesive, consumer-defined stores
// (mirroring auth.Stores) so each stays small and fakes only implement what a
// scenario needs. Repository is their union (27 methods, under the ~30
// criterion), kept as a single alias because the portfolio Service
// orchestrates across all of them. The asset catalog moved to the market
// module (TECH_DEBT #12); portfolio reads assets via its AssetReader interface.

// PortfolioStore persists portfolios themselves plus their risk catalog.
type PortfolioStore interface {
	GetPortfoliosRisks(ctx context.Context) ([]Risk, error)
	GetPortfoliosByUserID(ctx context.Context, userID uuid.UUID) ([]Portfolio, error)
	GetPortfoliosSummaryByUserID(ctx context.Context, userID uuid.UUID) ([]SummaryView, error)
	GetPortfolioByID(ctx context.Context, portfolioID, userID uuid.UUID) (Portfolio, error)
	CreatePortfolio(ctx context.Context, userID uuid.UUID, name, description, baseCurrency string, riskID uuid.UUID, typePortfolio Type, priceValue money.Money, isDefault bool) (Portfolio, error)
	UpdatePortfolio(ctx context.Context, userID, portfolioID uuid.UUID, name, description string, portfolioType Type, riskID uuid.UUID, isDefault bool) (Portfolio, error)
	GetEntriesByPortfolioID(ctx context.Context, portfolioID uuid.UUID) ([]Entry, error)
	GetTopTransactionByPortfolioID(ctx context.Context, userID, portfolioID uuid.UUID) (TopTransactionDTO, error)
}

// PlatformStore persists investment sources (platforms).
type PlatformStore interface {
	CreatePlatform(ctx context.Context, userID uuid.UUID, sourceType SourceType, name, description string) (InvestmentSource, error)
	GetPlatformsWithStats(ctx context.Context, userID uuid.UUID) ([]PlatformStats, error)
	UpdatePlatform(ctx context.Context, userID, sourceID uuid.UUID, name, description string, sourceType SourceType, isActive bool) (PlatformStats, error)
	DeletePlatform(ctx context.Context, userID, sourceID uuid.UUID) error
}

// RateStore reads stored exchange rates for display-currency conversion
// (writing/syncing rates is owned by the market module). The asset catalog
// also moved to market; portfolio reads assets through its AssetReader
// interface, not this repository.
type RateStore interface {
	GetExchangeRateByPair(ctx context.Context, from, to string) (money.Decimal, error)
}

// TransactionStore persists portfolio entries and their transactions.
type TransactionStore interface {
	CreatePortfolioEntry(ctx context.Context, userID, portfolioID, assetID uuid.UUID, sourceID uuid.UUID, txnType TransactionType, quantity money.Decimal, price money.Money, costCurrency string, category EntryCategory, entryDate time.Time, notes string) (Entry, error)
	GetEntryWithAsset(ctx context.Context, entryID uuid.UUID) (Entry, error)
	GetTransactionsByEntryID(ctx context.Context, userID, entryID uuid.UUID) ([]Transaction, error)
	CountAssetTransactions(ctx context.Context, userID, portfolioID uuid.UUID, ticker string) (int, error)
	GetAssetTransactionsPaginated(ctx context.Context, userID, portfolioID uuid.UUID, ticker string, limit, offset int) ([]Transaction, error)
	GetRecentTransactionsByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]Transaction, error)
	GetAssetAllocationByUserID(ctx context.Context, userID uuid.UUID) ([]AllocationItem, error)
	CreateTransaction(ctx context.Context, userID, entryID uuid.UUID, txnType TransactionType, quantity money.Decimal, price money.Money, currency string, fees money.Money, transactionDate time.Time, notes string) (Transaction, error)
	UpdateTransaction(ctx context.Context, userID, txnID uuid.UUID, txnType TransactionType, quantity money.Decimal, price money.Money, currency string, fees money.Money, transactionDate time.Time, notes string) (Transaction, error)
	ImportEntryTransactions(ctx context.Context, userID, portfolioID, sourceID uuid.UUID, rows []ImportTransactionRow) (int, error)
}

// SnapshotStore persists daily portfolio snapshots and reads growth series.
type SnapshotStore interface {
	GetAllPortfolioSummaryRows(ctx context.Context) ([]SnapshotRow, error)
	UpsertPortfolioSnapshot(ctx context.Context, portfolioID uuid.UUID, snapshotDate time.Time, totalValue, currency, totalGainLoss, totalGainLossPct string) error
	GetPortfolioGrowthByUserID(ctx context.Context, userID uuid.UUID, hasSince bool, since time.Time) ([]GrowthPoint, error)
	GetPortfolioGrowthByPortfolioID(ctx context.Context, userID, portfolioID uuid.UUID, hasSince bool, since time.Time) ([]GrowthPoint, error)
}

// Repository is the union of the module's stores, satisfied by
// *PostgresRepository. The Service orchestrates across all of them.
type Repository interface {
	PortfolioStore
	PlatformStore
	RateStore
	TransactionStore
	SnapshotStore
}

// Ensure the concrete repository keeps satisfying the interface.
var _ Repository = (*PostgresRepository)(nil)
