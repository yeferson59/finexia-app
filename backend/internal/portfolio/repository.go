package portfolio

import (
	"context"
	"time"

	"uuid"

	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"
)

// The persistence surface is split into cohesive, consumer-defined stores
// (mirroring auth.Stores) so each stays small and fakes only implement what a
// scenario needs. Repository is their union (31 methods, around the ~30
// criterion), kept as a single alias because the portfolio Service
// orchestrates across all of them. The asset catalog belongs to the market
// module; portfolio reads assets via its AssetReader interface instead.

// PortfolioStore persists portfolios themselves plus their risk catalog.
type PortfolioStore interface {
	GetPortfoliosRisks(ctx context.Context) ([]Risk, error)
	GetPortfoliosByUserID(ctx context.Context, userID uuid.UUID) ([]Portfolio, error)
	GetPortfoliosSummaryByUserID(ctx context.Context, userID uuid.UUID) ([]SummaryView, error)
	GetPortfolioByID(ctx context.Context, portfolioID, userID uuid.UUID) (Portfolio, error)
	CreatePortfolio(ctx context.Context, userID uuid.UUID, name, description string, baseCurrency money.Currency, riskID uuid.UUID, typePortfolio Type, priceValue money.Money, isDefault bool) (Portfolio, error)
	UpdatePortfolio(ctx context.Context, userID, portfolioID uuid.UUID, name, description string, portfolioType Type, riskID uuid.UUID, isDefault bool) (Portfolio, error)
	GetEntriesByPortfolioID(ctx context.Context, portfolioID uuid.UUID) ([]Entry, error)
	GetTopTransactionByPortfolioID(ctx context.Context, userID, portfolioID uuid.UUID) (TopTransactionDTO, error)
}

// PlatformStore persists investment sources (platforms).
type PlatformStore interface {
	CreatePlatform(ctx context.Context, userID uuid.UUID, sourceType SourceType, name, description string) (InvestmentSource, error)
	// GetPlatformsWithStats reports each platform's totals in displayCurrency,
	// or in the account's preferred currency when it is empty.
	GetPlatformsWithStats(ctx context.Context, userID uuid.UUID, displayCurrency money.Currency) ([]PlatformStats, error)
	UpdatePlatform(ctx context.Context, userID, sourceID uuid.UUID, name, description string, sourceType SourceType, isActive bool) (PlatformStats, error)
	DeletePlatform(ctx context.Context, userID, sourceID uuid.UUID) error
}

// RateStore reads stored exchange rates for display-currency conversion
// (writing/syncing rates is owned by the market module). The asset catalog
// also moved to market; portfolio reads assets through its AssetReader
// interface, not this repository.
type RateStore interface {
	GetExchangeRateByPair(ctx context.Context, from, to money.Currency) (decimal.Decimal, error)
	// GetUserExchangeRateByPair reads the rate the user's own key fetched.
	// Under BYO-key it is consulted before the shared table, which now only
	// holds admin-entered rows.
	GetUserExchangeRateByPair(ctx context.Context, userID uuid.UUID, from, to money.Currency) (decimal.Decimal, error)
}

// HoldingsStore answers what the per-user market sync needs to know: which
// assets this user holds, and which currency conversions their portfolios
// require. Reading it is what keeps a personal API quota spent only on the
// assets the user actually owns.
type HoldingsStore interface {
	GetHeldAssetIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
	GetRequiredCurrencyPairs(ctx context.Context, userID uuid.UUID) ([]CurrencyPair, error)
}

// CurrencyPair is a conversion a user's portfolios need. It mirrors
// market.CurrencyPair; the two are kept separate so this module's repository
// surface does not leak the market module's types.
type CurrencyPair struct{ From, To money.Currency }

// TransactionStore persists portfolio entries and their transactions.
type TransactionStore interface {
	CreatePortfolioEntry(ctx context.Context, userID, portfolioID, assetID, sourceID uuid.UUID, costCurrency money.Currency, in TransactionInput) (Entry, error)
	GetEntryWithAsset(ctx context.Context, entryID uuid.UUID) (Entry, error)
	DeletePortfolioEntry(ctx context.Context, userID, entryID uuid.UUID) (int, error)
	// ChangeEntrySettlement restates a position in another cost currency with the
	// rate of each of its transactions; see PlanSettlement.
	ChangeEntrySettlement(ctx context.Context, userID, entryID uuid.UUID, costCurrency money.Currency, rates map[uuid.UUID]decimal.Decimal) (int, error)
	GetTransactionsByEntryID(ctx context.Context, userID, entryID uuid.UUID) ([]Transaction, error)
	CountAssetTransactions(ctx context.Context, userID, portfolioID uuid.UUID, ticker string) (int, error)
	GetAssetTransactionsPaginated(ctx context.Context, userID, portfolioID uuid.UUID, ticker string, limit, offset int) ([]Transaction, error)
	GetRecentTransactionsByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]Transaction, error)
	GetAssetAllocationByUserID(ctx context.Context, userID uuid.UUID, targetCurrency money.Currency) ([]AllocationItem, error)
	// GetAssetHoldingsByUserID is the same aggregation one level finer: per
	// asset instead of per category, plus the units held.
	GetAssetHoldingsByUserID(ctx context.Context, userID uuid.UUID, targetCurrency money.Currency) ([]AssetHolding, error)
	// GetSectorAllocationByUserID is the same aggregation down the other axis:
	// per industry rather than per kind of instrument. Same positions, same
	// valuation and same currency contract as the two above.
	GetSectorAllocationByUserID(ctx context.Context, userID uuid.UUID, targetCurrency money.Currency) ([]SectorAllocationItem, error)
	CreateTransaction(ctx context.Context, userID, entryID uuid.UUID, in TransactionInput) (Transaction, error)
	UpdateTransaction(ctx context.Context, userID, txnID uuid.UUID, in TransactionInput) (Transaction, error)
	DeleteTransaction(ctx context.Context, userID, txnID uuid.UUID) error
	ImportEntryTransactions(ctx context.Context, userID, portfolioID, sourceID uuid.UUID, rows []ImportTransactionRow) (int, error)
}

// SnapshotStore persists daily portfolio snapshots and reads growth series.
//
// A growth series closes on asOf, read live instead of from a snapshot: see
// GetPortfolioGrowthByUserID in postgres_snapshot.go.
type SnapshotStore interface {
	GetAllPortfolioSummaryRows(ctx context.Context) ([]SnapshotRow, error)
	UpsertPortfolioSnapshot(ctx context.Context, row SnapshotRow, snapshotDate time.Time) error
	GetPortfolioGrowthByUserID(ctx context.Context, userID uuid.UUID, currency money.Currency, hasSince bool, since, asOf time.Time) ([]GrowthPoint, error)
	GetPortfolioGrowthByPortfolioID(ctx context.Context, userID, portfolioID uuid.UUID, hasSince bool, since, asOf time.Time) ([]GrowthPoint, error)
	GetPortfolioValuesAsOf(ctx context.Context, userID uuid.UUID, asOf time.Time) ([]PortfolioValuePoint, error)
}

// CashStore persists cash balances: positions of type cash and the movements
// on them. It is a separate store because every write checks the balance it
// would leave while holding a lock on it, which none of the generic transaction
// writers do.
type CashStore interface {
	GetCashBalancesByUserID(ctx context.Context, userID uuid.UUID, displayCurrency money.Currency) ([]CashBalance, error)
	CountCashMovements(ctx context.Context, userID uuid.UUID) (int, error)
	GetCashMovementsPaginated(ctx context.Context, userID uuid.UUID, limit, offset int) ([]CashMovement, error)
	CreateCashMovement(ctx context.Context, userID, portfolioID, sourceID, pocketID uuid.UUID, in CashMovementInput) (CashMovement, error)
	UpdateCashMovement(ctx context.Context, userID, txnID uuid.UUID, in CashMovementInput) (CashMovement, error)
	DeleteCashMovement(ctx context.Context, userID, txnID uuid.UUID) error
	// MoveCash records the two legs of a move between balances of one account
	// in one transaction, so the money is never in both places or in neither.
	MoveCash(ctx context.Context, userID, portfolioID, sourceID uuid.UUID, in CashMoveInput) (CashMove, error)
}

// CashPocketStore persists the pockets of a cash account: the subaccounts its
// money can sit in, each earning a rate of its own (000047). Every write locks
// the platform first, so the pockets of one account are written one at a time.
type CashPocketStore interface {
	GetCashPocketsByUserID(ctx context.Context, userID uuid.UUID) ([]CashPocket, error)
	GetCashPocketByID(ctx context.Context, userID, pocketID uuid.UUID) (CashPocket, error)
	CreateCashPocket(ctx context.Context, userID uuid.UUID, in NewCashPocketInput) (CashPocket, error)
	RenameCashPocket(ctx context.Context, userID, pocketID uuid.UUID, in RenameCashPocketInput) (CashPocket, error)
	DeleteCashPocket(ctx context.Context, userID, pocketID uuid.UUID) error
	// The three writes of a fixed deposit (000048): opening one whole, stopping
	// its rate, and moving what it holds back to the main account. Cancelling one
	// is the last two with the days it still owes computed in between, which is
	// why they are separate; the maturity sweep is the last one over every
	// deposit that has come due.
	OpenFixedDeposit(ctx context.Context, userID uuid.UUID, in NewFixedDepositInput) (CashPocket, error)
	EndFixedDeposit(ctx context.Context, userID, pocketID uuid.UUID, closesOn time.Time) (CashPocket, error)
	SettleFixedDeposit(ctx context.Context, userID, pocketID uuid.UUID, on time.Time, penalty decimal.Decimal, notes string) (CashPocket, error)
	MatureCashPockets(ctx context.Context, on time.Time) (int, error)
}

// CashRateStore persists the rates cash accounts earn, as versions by the day
// each takes effect. Every write locks the platform first, so the versions of
// one account are written one at a time.
type CashRateStore interface {
	GetCashRatesByUserID(ctx context.Context, userID uuid.UUID) ([]CashRate, error)
	CreateCashRate(ctx context.Context, userID uuid.UUID, in NewCashRateInput) (CashRate, error)
	UpdateCashRate(ctx context.Context, userID, rateID uuid.UUID, in CashRateInput) (CashRate, error)
	EndCashRate(ctx context.Context, userID, rateID uuid.UUID, endsOn time.Time) (CashRate, error)
	RescheduleCashRate(ctx context.Context, userID, rateID uuid.UUID, in RescheduleCashRateInput) (CashRate, error)
	DeleteCashRate(ctx context.Context, userID, rateID uuid.UUID) error
}

// CashAccrualStore keeps the ledger of the interest cash balances earn: which
// balances earn, and one day of interest at a time.
//
// The last two are for the days a rate posted monthly holds: which balances
// hold days no month end will ever close, and the credit that pays them.
type CashAccrualStore interface {
	GetCashAccrualTargets(ctx context.Context, through time.Time, filter CashAccrualFilter) ([]CashAccrualTarget, error)
	AccrueCashInterestDay(ctx context.Context, entryID, rateID uuid.UUID, day time.Time) (bool, error)
	GetHeldCashInterest(ctx context.Context, through time.Time, filter CashAccrualFilter) ([]uuid.UUID, error)
	PostHeldCashInterest(ctx context.Context, entryID uuid.UUID) (bool, error)
	// ClearCashInterest throws away the days computed from a day, so a
	// recalculation can compute them again.
	ClearCashInterest(ctx context.Context, filter CashAccrualFilter, from time.Time) (CashInterestCleared, error)
	// SumRecalculatedCashInterest reads what a recalculation wrote: the days it
	// cleared, as they are now, and the days no balance had computed yet.
	SumRecalculatedCashInterest(ctx context.Context, cleared CashInterestCleared, through time.Time) (redone, fresh CashInterestDays, err error)
}

// FundStore persists the investment funds an owner follows and the marks that
// price them (000057). Every write to a fund's marks locks its user_funds row
// first, so the price copied to user_asset_prices is always the latest mark.
type FundStore interface {
	GetFundsByUserID(ctx context.Context, userID uuid.UUID) ([]Fund, error)
	GetFund(ctx context.Context, userID, assetID uuid.UUID) (Fund, error)
	CreateFund(ctx context.Context, userID uuid.UUID, in NewFundInput) (Fund, error)
	DeleteFund(ctx context.Context, userID, assetID uuid.UUID) error
	GetFundMarks(ctx context.Context, userID, assetID uuid.UUID) ([]FundMark, error)
	// UpsertFundMark and DeleteFundMark also revalue every snapshot from the
	// mark's day on, in the same transaction.
	UpsertFundMark(ctx context.Context, userID, assetID uuid.UUID, in FundMarkInput) (FundMark, error)
	// UpsertFundMarks writes several at once — the table of a statement — with
	// one lock, one replay and one revaluation.
	UpsertFundMarks(ctx context.Context, userID, assetID uuid.UUID, in []FundMarkInput) (int, error)
	DeleteFundMark(ctx context.Context, userID, assetID uuid.UUID, date time.Time) error
	// The movements of a fund followed by balance (000058). Every write replays
	// the whole fund and revalues its snapshots in the same transaction.
	GetFundMovements(ctx context.Context, userID, assetID uuid.UUID) ([]FundMovement, error)
	GetFundMovement(ctx context.Context, userID, txnID uuid.UUID) (FundMovement, error)
	ContributeToFund(ctx context.Context, userID, assetID uuid.UUID, in FundContributionInput) (FundMovement, error)
	WithdrawFromFund(ctx context.Context, userID, assetID uuid.UUID, in FundWithdrawalInput) (FundMovement, error)
	UpdateFundMovement(ctx context.Context, userID, txnID uuid.UUID, in FundMovementEdit) (FundMovement, error)
	DeleteFundMovement(ctx context.Context, userID, txnID uuid.UUID) error
}

// PublicFundStore persists the SFC's catalog of funds and the links to it
// (000059). A link writes the published values as marks of the fund, under the
// same lock and with the same price and snapshot upkeep as FundStore's writes.
type PublicFundStore interface {
	UpsertPublicFunds(ctx context.Context, funds []PublicFund) (int, error)
	CountPublicFunds(ctx context.Context) (int, error)
	SearchPublicFunds(ctx context.Context, words []string, since time.Time, limit int) ([]PublicFund, error)
	GetPublicFund(ctx context.Context, id string) (PublicFund, error)
	LinkFund(ctx context.Context, userID, assetID uuid.UUID, publicID string, values []PublicFundValue) (Fund, error)
	UnlinkFund(ctx context.Context, userID, assetID uuid.UUID) (Fund, error)
	GetLinkedFunds(ctx context.Context) ([]LinkedFund, error)
	ImportPublicMarks(ctx context.Context, userID, assetID uuid.UUID, publicID string, values []PublicFundValue) (int, error)
}

// Repository is the union of the module's stores, satisfied by
// *PostgresRepository. The Service orchestrates across all of them.
type Repository interface {
	PortfolioStore
	PlatformStore
	RateStore
	TransactionStore
	SnapshotStore
	HoldingsStore
	CashStore
	CashPocketStore
	CashRateStore
	CashAccrualStore
	FundStore
	PublicFundStore
}

// Ensure the concrete repository keeps satisfying the interface.
var _ Repository = (*PostgresRepository)(nil)
