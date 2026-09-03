package portfolio

import (
	"context"

	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/yeferson59/gofinance/v2/money"

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
			&entry.Asset.CurrentPrice,
			&entry.Asset.PriceUpdatedAt,
			&entry.PriceSource,
			&entry.Asset.CreatedAt,
			&entry.Asset.UpdatedAt,
		); err != nil {
			return nil, err
		}

		entry.Category = entry.Asset.AssetType
		entry.Price.SetCurrency(entry.CostCurrency)
		if entry.Asset.CurrentPrice != nil {
			entry.Asset.CurrentPrice.SetCurrency(entry.Asset.Currency)
		}

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

// CreatePortfolioEntry opens a position and records the trade that opened it.
//
// costCurrency is the position's — what the account was debited in — and
// in.Currency is the trade's, which is the asset's own whenever the fill
// happened on its home exchange. The two differ exactly when the broker had to
// convert, and in.FXRate is the rate it converted at; Validate refuses the
// combinations that cannot be true.
//
// costCurrency only takes effect when this opens a new position. The endpoint
// upserts, so calling it again for the same portfolio/asset/source adds a trade
// to the position that is already there, and that position keeps the currency it
// was opened with.
func (r *PostgresRepository) CreatePortfolioEntry(ctx context.Context, userID, portfolioID, assetID, sourceID uuid.UUID, costCurrency money.Currency, in TransactionInput) (Entry, error) {
	var entry Entry

	rate := in.Rate()
	// The seeded price is in the position's currency, like the column it goes
	// into. The trigger overwrites it from the transactions a moment later, so
	// this only ever shows through if that trigger is gone — but a row that is
	// right on its own is worth the multiplication. The rate is not validated
	// yet, and does not need to be: an unvalidated Rate() is either what the
	// caller stated or 1, and the upsert below is rolled back if it turns out to
	// be neither legal nor needed.
	costPrice := in.Price.MulDecimal(rate)

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

		// cost_currency comes back from the upsert rather than being assumed to
		// be the one that was asked for. On a conflict this row already existed
		// and DO UPDATE deliberately leaves its currency alone — a position
		// cannot change what it costs in halfway through its life — so the
		// requested currency may simply not be the position's. Validating the
		// rate against the requested one would then approve a conversion into a
		// currency the entry does not use, and the trigger would average euros
		// into a dollar column with a rate applied on top.
		var entryID uuid.UUID
		var entryCostCurrency money.Currency
		if err := tx.QueryRow(ctx, `
		INSERT INTO portfolio_entries (portfolio_id, asset_id, source_id, quantity, price, cost_currency, entry_date, notes)
		VALUES ($1::uuid, $2::uuid, $3::uuid, 0, $4::numeric, $5::char(3), $6::date, $7)
		ON CONFLICT (portfolio_id, asset_id, COALESCE(source_id::TEXT, ''))
		DO UPDATE SET updated_at = NOW()
		RETURNING id, cost_currency
	`, portfolioID, assetID, sourceID, costPrice.String(), costCurrency, in.TransactionDate, in.Notes).Scan(&entryID, &entryCostCurrency); err != nil {
			return err
		}

		// A refusal here rolls the upsert back with it, so a rejected request
		// leaves no half-opened position behind.
		tradeCurrency, err := in.Validate(entryCostCurrency)
		if err != nil {
			return err
		}

		if _, err := tx.Exec(ctx, `
		INSERT INTO transactions (entry_id, type, quantity, price, currency, fx_rate, fees, transaction_date, notes)
		VALUES ($1::uuid, $2::transaction_type, $3::numeric, $4::numeric, $5::char(3), $6::numeric, 0, $7::date, $8)
	`, entryID, in.Type, in.Quantity.String(), in.Price.String(), tradeCurrency, rate.String(), in.TransactionDate, in.Notes); err != nil {
			return err
		}

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
			&entry.Category,
			&entry.EntryDate,
			&entry.Notes,
			&entry.CreatedAt,
			&entry.UpdatedAt,
		); err != nil {
			return err
		}

		entry.Price.SetCurrency(entry.CostCurrency)

		return nil
	}); err != nil {
		return entry, err
	}

	return entry, nil
}
