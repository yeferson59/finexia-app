package portfolio

import (
	"context"
	"fmt"
	"time"

	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/market"
	"github.com/yeferson59/finexia-app/internal/platform/database"
)

func (r *PostgresRepository) GetEntriesByPortfolioID(ctx context.Context, portfolioID uuid.UUID) ([]Entry, error) {
	rows, err := r.db.Query(ctx, `
		SELECT pe.id, pe.portfolio_id, pe.asset_id, pe.source_id, pe.quantity, pe.price, pe.cost_currency, pe.entry_date, COALESCE(pe.notes, ''), pe.created_at, pe.updated_at,
		       a.id, a.ticker, a.name, a.asset_type, COALESCE(a.exchange, ''), a.currency,
		       -- BYO-key: the price this user's own key fetched wins; the shared
		       -- column now only ever holds an admin-entered manual price. The
		       -- owner comes from the portfolio row, so no signature has to carry
		       -- a user id down here.
		       COALESCE(uap.price::text, a.current_price::text), COALESCE(uap.fetched_at, a.price_updated_at),
		       -- Which arm of that COALESCE won. Without it the caller gets a
		       -- number it cannot qualify: a market price and a cost basis look
		       -- identical, and only one of them makes a return meaningful.
		       CASE
		         WHEN uap.price       IS NOT NULL THEN 'own'
		         WHEN a.current_price IS NOT NULL THEN 'manual'
		         ELSE 'cost'
		       END,
		       a.created_at, a.updated_at
		FROM portfolio_entries pe
		JOIN portfolios p ON p.id = pe.portfolio_id
		JOIN assets a ON a.id = pe.asset_id
		LEFT JOIN user_asset_prices uap ON uap.asset_id = a.id AND uap.user_id = p.user_id
		WHERE pe.portfolio_id = $1
		ORDER BY pe.created_at DESC
	`, portfolioID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := make([]Entry, 0)
	for rows.Next() {
		var entry Entry
		var assetPriceStr *string

		if err := rows.Scan(
			&entry.ID,
			&entry.PortfolioID,
			&entry.AssetID,
			&entry.SourceID,
			&entry.Quantity,
			&entry.Price,
			&entry.CostCurrency,
			&entry.EntryDate,
			&entry.Notes,
			&entry.CreatedAt,
			&entry.UpdatedAt,
			&entry.Asset.ID,
			&entry.Asset.Ticker,
			&entry.Asset.Name,
			&entry.Asset.AssetType,
			&entry.Asset.Exchange,
			&entry.Asset.Currency,
			&assetPriceStr,
			&entry.Asset.PriceUpdatedAt,
			&entry.PriceSource,
			&entry.Asset.CreatedAt,
			&entry.Asset.UpdatedAt,
		); err != nil {
			return nil, err
		}

		scanAssetCurrentPrice(&entry.Asset, assetPriceStr)
		// The position's class follows the asset, so it is read off the row the
		// join already brought rather than off a column of its own.
		entry.Category = entryCategoryFor(entry.Asset.AssetType)

		fmt.Println("price es ", entry.Price)
		entry.Price.SetCurrency(entry.CostCurrency)

		entries = append(entries, entry)
	}

	return entries, nil
}

func (r *PostgresRepository) GetEntryWithAsset(ctx context.Context, entryID uuid.UUID) (Entry, error) {
	var entry Entry
	if err := r.db.QueryRow(ctx, `
		SELECT pe.id, pe.portfolio_id, pe.asset_id, a.ticker, a.name
		FROM portfolio_entries pe
		JOIN assets a ON a.id = pe.asset_id
		WHERE pe.id = $1
	`, entryID).Scan(
		&entry.ID,
		&entry.PortfolioID,
		&entry.AssetID,
		&entry.Asset.Ticker,
		&entry.Asset.Name,
	); err != nil {
		return entry, err
	}

	return entry, nil
}

func (r *PostgresRepository) CreatePortfolioEntry(ctx context.Context, userID, portfolioID, assetID uuid.UUID, sourceID uuid.UUID, txnType TransactionType, quantity decimal.Decimal, price money.Money, costCurrency string, entryDate time.Time, notes string) (Entry, error) {
	var entry Entry

	if err := database.WithinTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		var owned bool

		if err := tx.QueryRow(ctx, `
		SELECT EXISTS (SELECT 1 FROM portfolios WHERE id = $1 AND user_id = $2)
		   AND ($3::uuid IS NULL OR EXISTS (SELECT 1 FROM investment_sources WHERE id = $3 AND user_id = $2))
	`, portfolioID, userID, sourceID).Scan(&owned); err != nil {
			return err
		}

		if !owned {
			return ErrPortfolioOrSourceNotFound
		}

		var entryID uuid.UUID
		if err := tx.QueryRow(ctx, `
		INSERT INTO portfolio_entries (portfolio_id, asset_id, source_id, quantity, price, cost_currency, entry_date, notes)
		VALUES ($1::uuid, $2::uuid, $3::uuid, 0, $4::numeric, $5::char(3), $6::date, $7)
		ON CONFLICT (portfolio_id, asset_id, COALESCE(source_id::TEXT, ''))
		DO UPDATE SET updated_at = NOW()
		RETURNING id
	`, portfolioID, assetID, sourceID, price.String(), costCurrency, entryDate, notes).Scan(&entryID); err != nil {
			return err
		}

		if _, err := tx.Exec(ctx, `
		INSERT INTO transactions (entry_id, type, quantity, price, currency, fees, transaction_date, notes)
		VALUES ($1::uuid, $2::transaction_type, $3::numeric, $4::numeric, $5::char(3), 0, $6::date, $7)
	`, entryID, txnType, quantity.String(), price.String(), costCurrency, entryDate, notes); err != nil {
			return err
		}

		// Read the position back with the values the trigger just recomputed.
		var assetType market.AssetType
		if err := tx.QueryRow(ctx, `
		SELECT pe.id, pe.portfolio_id, pe.asset_id, pe.source_id, pe.quantity, pe.price, pe.cost_currency, a.asset_type, pe.entry_date, COALESCE(pe.notes, ''), pe.created_at, pe.updated_at
		FROM portfolio_entries pe
		JOIN assets a ON a.id = pe.asset_id
		WHERE pe.id = $1
	`, entryID).Scan(
			&entry.ID,
			&entry.PortfolioID,
			&entry.AssetID,
			&entry.SourceID,
			&entry.Quantity,
			&entry.Price,
			&entry.CostCurrency,
			&assetType,
			&entry.EntryDate,
			&entry.Notes,
			&entry.CreatedAt,
			&entry.UpdatedAt,
		); err != nil {
			return err
		}

		entry.Category = entryCategoryFor(assetType)
		price, err := entry.Price.SetCurrency(entry.CostCurrency)
		if err != nil {
			return err
		}

		entry.Price = price

		return nil
	}); err != nil {
		return entry, err
	}

	return entry, nil
}
