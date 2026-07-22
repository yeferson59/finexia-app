package portfolio

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"
)

func (r *PostgresRepository) GetEntriesByPortfolioID(ctx context.Context, portfolioID uuid.UUID) ([]Entry, error) {
	rows, err := r.db.Query(ctx, `
		SELECT pe.id, pe.portfolio_id, pe.asset_id, pe.source_id, pe.quantity, pe.price, pe.cost_currency, pe.category, pe.entry_date, COALESCE(pe.notes, ''), pe.created_at, pe.updated_at,
		       a.id, a.ticker, a.name, a.asset_type, COALESCE(a.exchange, ''), a.currency, a.current_price, a.price_updated_at, a.created_at, a.updated_at
		FROM portfolio_entries pe
		JOIN assets a ON a.id = pe.asset_id
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
		var quantity string
		var price string
		var sourceID pgtype.UUID
		var assetPriceStr *string

		if err := rows.Scan(
			&entry.ID,
			&entry.PortfolioID,
			&entry.AssetID,
			&sourceID,
			&quantity,
			&price,
			&entry.CostCurrency,
			&entry.Category,
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
			&entry.Asset.CreatedAt,
			&entry.Asset.UpdatedAt,
		); err != nil {
			return nil, err
		}

		scanAssetCurrentPrice(&entry.Asset, assetPriceStr)

		if sourceID.Valid {
			entry.SourceID = uuid.UUID(sourceID.Bytes)
		}

		entry.Quantity = decimal.MustFromString(quantity)
		entry.Price = money.MustMoneyFromString(price, money.USD)

		entries = append(entries, entry)
	}

	return entries, nil
}

func (r *PostgresRepository) GetEntryWithAsset(ctx context.Context, entryID uuid.UUID) (Entry, error) {
	var entry Entry
	err := r.db.QueryRow(ctx, `
		SELECT pe.id, pe.portfolio_id, pe.asset_id,
		       a.ticker, a.name
		FROM portfolio_entries pe
		JOIN assets a ON a.id = pe.asset_id
		WHERE pe.id = $1
	`, entryID).Scan(
		&entry.ID,
		&entry.PortfolioID,
		&entry.AssetID,
		&entry.Asset.Ticker,
		&entry.Asset.Name,
	)
	if err != nil {
		return Entry{}, err
	}
	return entry, nil
}

func (r *PostgresRepository) CreatePortfolioEntry(ctx context.Context, userID, portfolioID, assetID uuid.UUID, sourceID uuid.UUID, txnType TransactionType, quantity money.Decimal, price money.Money, costCurrency string, category EntryCategory, entryDate time.Time, notes string) (Entry, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Entry{}, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var owned bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS (SELECT 1 FROM portfolios WHERE id = $1 AND user_id = $2)
		   AND ($3::uuid IS NULL OR EXISTS (SELECT 1 FROM investment_sources WHERE id = $3 AND user_id = $2))
	`, portfolioID, userID, sourceID).Scan(&owned); err != nil {
		return Entry{}, err
	}

	if !owned {
		return Entry{}, errors.New("portfolio or source not found")
	}

	var entryID uuid.UUID
	if err := tx.QueryRow(ctx, `
		INSERT INTO portfolio_entries (portfolio_id, asset_id, source_id, quantity, price, cost_currency, category, entry_date, notes)
		VALUES ($1::uuid, $2::uuid, $3::uuid, 0, $4::numeric, $5::char(3), $6::portfolio_entry_category, $7::date, $8)
		ON CONFLICT (portfolio_id, asset_id, COALESCE(source_id::TEXT, ''))
		DO UPDATE SET updated_at = NOW()
		RETURNING id
	`, portfolioID, assetID, sourceID, price.String(), costCurrency, category, entryDate, notes).Scan(&entryID); err != nil {
		return Entry{}, err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO transactions (entry_id, type, quantity, price, currency, fees, transaction_date, notes)
		VALUES ($1::uuid, $2::transaction_type, $3::numeric, $4::numeric, $5::char(3), 0, $6::date, $7)
	`, entryID, txnType, quantity.String(), price.String(), costCurrency, entryDate, notes); err != nil {
		return Entry{}, err
	}

	// Read the position back with the values the trigger just recomputed.
	var entry Entry
	var quantityValue, priceValue string
	var sourceIDResult pgtype.UUID
	if err := tx.QueryRow(ctx, `
		SELECT id, portfolio_id, asset_id, source_id, quantity, price, cost_currency, category, entry_date, COALESCE(notes, ''), created_at, updated_at
		FROM portfolio_entries
		WHERE id = $1
	`, entryID).Scan(
		&entry.ID,
		&entry.PortfolioID,
		&entry.AssetID,
		&sourceIDResult,
		&quantityValue,
		&priceValue,
		&entry.CostCurrency,
		&entry.Category,
		&entry.EntryDate,
		&entry.Notes,
		&entry.CreatedAt,
		&entry.UpdatedAt,
	); err != nil {
		return Entry{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return Entry{}, err
	}

	if sourceIDResult.Valid {
		entry.SourceID = uuid.UUID(sourceIDResult.Bytes)
	}

	entry.Quantity = decimal.MustFromString(quantityValue)
	entry.Price = money.MustMoneyFromString(priceValue, money.USD)

	return entry, nil
}
