package portfolio

import (
	"context"
	"errors"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/market"
)

func (r *PostgresRepository) GetTopTransactionByPortfolioID(ctx context.Context, userID, portfolioID uuid.UUID) (TopTransactionDTO, error) {
	var dto TopTransactionDTO
	err := r.db.QueryRow(ctx, `
		SELECT
			(t.quantity::numeric * t.price::numeric)::text,
			t.type,
			t.currency,
			t.transaction_date,
			a.ticker,
			a.name
		FROM transactions t
		JOIN portfolio_entries pe ON pe.id = t.entry_id
		JOIN assets a ON a.id = pe.asset_id
		JOIN portfolios p ON p.id = pe.portfolio_id
		WHERE pe.portfolio_id = $1 AND p.user_id = $2
		ORDER BY t.quantity::numeric * t.price::numeric DESC
		LIMIT 1
	`, portfolioID, userID).Scan(
		&dto.Value,
		&dto.Type,
		&dto.Currency,
		&dto.TransactionDate,
		&dto.AssetTicker,
		&dto.AssetName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return TopTransactionDTO{}, nil
		}
		return TopTransactionDTO{}, err
	}
	return dto, nil
}

func (r *PostgresRepository) GetRecentTransactionsByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]Transaction, error) {
	rows, err := r.db.Query(ctx, `
		SELECT t.id, t.entry_id, t.type, t.quantity, t.price, t.currency, t.fees,
		       t.transaction_date, COALESCE(t.notes, ''), t.created_at, t.updated_at,
		       a.ticker, a.name
		FROM transactions t
		JOIN portfolio_entries pe ON pe.id = t.entry_id
		JOIN portfolios p ON p.id = pe.portfolio_id
		JOIN assets a ON a.id = pe.asset_id
		WHERE p.user_id = $1
		ORDER BY t.transaction_date DESC, t.created_at DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	txns := make([]Transaction, 0)
	for rows.Next() {
		var txn Transaction
		var quantity, price, fees string
		if err := rows.Scan(
			&txn.ID, &txn.EntryID, &txn.Type, &quantity, &price,
			&txn.Currency, &fees, &txn.TransactionDate, &txn.Notes,
			&txn.CreatedAt, &txn.UpdatedAt,
			&txn.Entry.Asset.Ticker, &txn.Entry.Asset.Name,
		); err != nil {
			return nil, err
		}
		txn.Quantity = decimal.MustFromString(quantity)
		txn.Price = moneyOf(price, txn.Currency)
		txn.Fees = moneyOf(fees, txn.Currency)
		txns = append(txns, txn)
	}
	return txns, nil
}

// GetAssetAllocationByUserID totals what the user holds per category, in one
// currency. The allocation spans every portfolio they own, and those can be
// denominated differently, so without conversion the shares were computed from
// a sum of euros and dollars — a percentage of nothing.
//
// targetCurrency is the currency to report in; empty means the user's own
// preference, which is the account-level answer to the same question and what
// the dashboard falls back to. Conversion goes through fx_rate (migration
// 000024), the same resolution the rest of the app uses: the user's own rate,
// then the shared one, inverted or hopped through USD as needed.
//
// A position with no rate is still counted, at face value, and reported in
// PositionsUnconverted — the same choice the summary and the holdings make.
// Dropping it would understate the category; hiding that it was not converted
// is what this is fixing.
//
// The category comes from the asset's own type, not from
// portfolio_entries.category. That column is a copy of the type taken when the
// entry was created — the create handler and the importer both derive it
// through entryCategoryFor — and nothing writes it again. Correcting an asset
// afterwards (a bond ETF first filed as a plain ETF) therefore moved it in the
// per-portfolio donut, which groups the holdings by their asset type, and left
// it in the old slice here: two charts over the same positions, disagreeing.
// Reading assets.asset_type is what makes both answer with one rule.
//
// The rows come back keyed by asset type and are folded into the entry
// categories the response has always spoken, so the vocabulary clients map to
// labels does not change.
func (r *PostgresRepository) GetAssetAllocationByUserID(ctx context.Context, userID uuid.UUID, targetCurrency string) ([]AllocationItem, error) {
	// The sum is rounded to the scale every money column here keeps. Postgres
	// adds the scales of what it multiplies, so quantity × price × rate reaches
	// thirty-odd decimals, and the decimal engine that parses this text back in
	// NewAllocationResponse caps at nineteen: past it the value read as zero and
	// the category's share silently collapsed.
	rows, err := r.db.Query(ctx, `
		SELECT
			a.asset_type,
			ROUND(COALESCE(SUM(pe.quantity::numeric * v.price * COALESCE(fx.rate, 1)), 0), 8)::text AS market_value,
			target.code,
			COUNT(DISTINCT pe.asset_id) FILTER (WHERE fx.rate IS NULL)::bigint AS positions_unconverted
		FROM portfolio_entries pe
		JOIN portfolios p ON p.id = pe.portfolio_id
		JOIN users u      ON u.id = p.user_id
		JOIN assets a     ON a.id = pe.asset_id
		LEFT JOIN user_asset_prices uap ON uap.asset_id = a.id AND uap.user_id = p.user_id
		CROSS JOIN LATERAL (
			SELECT COALESCE(NULLIF($2::text, ''), u.preferred_currency) AS code
		) target
		-- The price and the currency it is quoted in have to be chosen together:
		-- a position with no market price is carried at cost, and that amount is
		-- in the cost currency, not in the asset's.
		CROSS JOIN LATERAL (
			SELECT
				COALESCE(uap.price::numeric, a.current_price::numeric, pe.price::numeric) AS price,
				CASE
					WHEN COALESCE(uap.price, a.current_price) IS NOT NULL
						THEN COALESCE(a.currency, pe.cost_currency)
					ELSE pe.cost_currency
				END AS currency
		) v
		CROSS JOIN LATERAL (
			SELECT fx_rate(p.user_id, v.currency, target.code) AS rate
		) fx
		WHERE p.user_id = $1
		  AND pe.quantity::numeric > 0
		GROUP BY a.asset_type, target.code
		ORDER BY ROUND(COALESCE(SUM(pe.quantity::numeric * v.price * COALESCE(fx.rate, 1)), 0), 8) DESC
	`, userID, targetCurrency)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rowsByType := make([]AllocationItem, 0)
	for rows.Next() {
		var assetType string
		var item AllocationItem
		if err := rows.Scan(&assetType, &item.MarketValue, &item.Currency, &item.PositionsUnconverted); err != nil {
			return nil, err
		}
		item.Category = entryCategoryFor(market.AssetType(assetType))
		rowsByType = append(rowsByType, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return foldAllocationByCategory(rowsByType), nil
}

// foldAllocationByCategory merges the rows that share a category and restores
// the by-value ordering the query asked for.
//
// entryCategoryFor sends any asset type it does not recognise to Others, so two
// of them can land on one category. Adding them is the only reading that keeps
// the shares summing to the whole — letting the second row win would drop a
// slice's worth of money off the chart — and a merged row can outgrow the one
// above it, which is why the order is rebuilt rather than left almost-sorted.
func foldAllocationByCategory(items []AllocationItem) []AllocationItem {
	folded := make([]AllocationItem, 0, len(items))
	index := make(map[EntryCategory]int, len(items))

	for _, item := range items {
		if at, seen := index[item.Category]; seen {
			folded[at].MarketValue = sumAmounts(folded[at].MarketValue, item.MarketValue)
			folded[at].PositionsUnconverted += item.PositionsUnconverted
			continue
		}
		index[item.Category] = len(folded)
		folded = append(folded, item)
	}

	slices.SortStableFunc(folded, func(a, b AllocationItem) int {
		return amountOf(b.MarketValue).Cmp(amountOf(a.MarketValue))
	})
	return folded
}

// amountOf parses one of the text amounts these queries return, treating a
// malformed one as zero — the same reading NewAllocationResponse takes, so a
// row cannot sort by one value and be shown as another.
func amountOf(raw string) decimal.Decimal {
	value, err := decimal.NewFromString(raw)
	if err != nil {
		return decimal.Zero
	}
	return value
}

// sumAmounts adds two of those amounts back into the same text form.
func sumAmounts(a, b string) string {
	return amountOf(a).Add(amountOf(b)).String()
}

func (r *PostgresRepository) GetTransactionsByEntryID(ctx context.Context, userID, entryID uuid.UUID) ([]Transaction, error) {
	var owned bool
	if err := r.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM portfolio_entries pe
			JOIN portfolios p ON p.id = pe.portfolio_id
			WHERE pe.id = $1 AND p.user_id = $2
		)
	`, entryID, userID).Scan(&owned); err != nil {
		return nil, err
	}
	if !owned {
		return nil, ErrEntryNotFound
	}

	rows, err := r.db.Query(ctx, `
		SELECT id, entry_id, type, quantity, price, currency, fees, transaction_date, COALESCE(notes, ''), created_at, updated_at
		FROM transactions
		WHERE entry_id = $1
		ORDER BY transaction_date DESC, created_at DESC
	`, entryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	txns := make([]Transaction, 0)
	for rows.Next() {
		var txn Transaction
		var quantity, price, fees string
		if err := rows.Scan(
			&txn.ID,
			&txn.EntryID,
			&txn.Type,
			&quantity,
			&price,
			&txn.Currency,
			&fees,
			&txn.TransactionDate,
			&txn.Notes,
			&txn.CreatedAt,
			&txn.UpdatedAt,
		); err != nil {
			return nil, err
		}
		txn.Quantity = decimal.MustFromString(quantity)
		txn.Price = moneyOf(price, txn.Currency)
		txn.Fees = moneyOf(fees, txn.Currency)
		txns = append(txns, txn)
	}
	return txns, nil
}

func (r *PostgresRepository) CountAssetTransactions(ctx context.Context, userID, portfolioID uuid.UUID, ticker string) (int, error) {
	var total int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM transactions t
		JOIN portfolio_entries pe ON pe.id = t.entry_id
		JOIN assets a ON a.id = pe.asset_id
		JOIN portfolios p ON p.id = pe.portfolio_id
		WHERE p.id = $1 AND a.ticker = $2 AND p.user_id = $3
	`, portfolioID, ticker, userID).Scan(&total)
	return total, err
}

func (r *PostgresRepository) GetAssetTransactionsPaginated(ctx context.Context, userID, portfolioID uuid.UUID, ticker string, limit, offset int) ([]Transaction, error) {
	rows, err := r.db.Query(ctx, `
		SELECT t.id, t.entry_id, t.type, t.quantity, t.price, t.currency, t.fees,
		       t.transaction_date, COALESCE(t.notes, ''), t.created_at, t.updated_at
		FROM transactions t
		JOIN portfolio_entries pe ON pe.id = t.entry_id
		JOIN assets a ON a.id = pe.asset_id
		JOIN portfolios p ON p.id = pe.portfolio_id
		WHERE p.id = $1 AND a.ticker = $2 AND p.user_id = $3
		ORDER BY t.transaction_date DESC, t.created_at DESC
		LIMIT $4 OFFSET $5
	`, portfolioID, ticker, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	txns := make([]Transaction, 0)
	for rows.Next() {
		var txn Transaction
		var quantity, price, fees string
		if err := rows.Scan(
			&txn.ID,
			&txn.EntryID,
			&txn.Type,
			&quantity,
			&price,
			&txn.Currency,
			&fees,
			&txn.TransactionDate,
			&txn.Notes,
			&txn.CreatedAt,
			&txn.UpdatedAt,
		); err != nil {
			return nil, err
		}
		txn.Quantity = decimal.MustFromString(quantity)
		txn.Price = moneyOf(price, txn.Currency)
		txn.Fees = moneyOf(fees, txn.Currency)
		txns = append(txns, txn)
	}
	return txns, nil
}

func (r *PostgresRepository) CreateTransaction(ctx context.Context, userID, entryID uuid.UUID, txnType TransactionType, quantity decimal.Decimal, price money.Money, currency string, fees money.Money, transactionDate time.Time, notes string) (Transaction, error) {
	var owned bool
	if err := r.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM portfolio_entries pe
			JOIN portfolios p ON p.id = pe.portfolio_id
			WHERE pe.id = $1 AND p.user_id = $2
		)
	`, entryID, userID).Scan(&owned); err != nil {
		return Transaction{}, err
	}
	if !owned {
		return Transaction{}, ErrEntryNotFound
	}

	var txn Transaction
	var quantityValue, priceValue, feesValue string
	if err := r.db.QueryRow(ctx, `
		INSERT INTO transactions (entry_id, type, quantity, price, currency, fees, transaction_date, notes)
		VALUES ($1::uuid, $2::transaction_type, $3::numeric, $4::numeric, $5::char(3), $6::numeric, $7::date, $8)
		RETURNING id, entry_id, type, quantity, price, currency, fees, transaction_date, COALESCE(notes, ''), created_at, updated_at
	`, entryID, txnType, quantity.String(), price.String(), currency, fees.String(), transactionDate, notes).Scan(
		&txn.ID,
		&txn.EntryID,
		&txn.Type,
		&quantityValue,
		&priceValue,
		&txn.Currency,
		&feesValue,
		&txn.TransactionDate,
		&txn.Notes,
		&txn.CreatedAt,
		&txn.UpdatedAt,
	); err != nil {
		return Transaction{}, err
	}

	txn.Quantity = decimal.MustFromString(quantityValue)
	txn.Price = moneyOf(priceValue, txn.Currency)
	txn.Fees = moneyOf(feesValue, txn.Currency)
	return txn, nil
}

func (r *PostgresRepository) UpdateTransaction(ctx context.Context, userID, txnID uuid.UUID, txnType TransactionType, quantity decimal.Decimal, price money.Money, currency string, fees money.Money, transactionDate time.Time, notes string) (Transaction, error) {
	var txn Transaction
	var quantityValue, priceValue, feesValue string
	err := r.db.QueryRow(ctx, `
		UPDATE transactions SET
			type             = $1::transaction_type,
			quantity         = $2::numeric,
			price            = $3::numeric,
			currency         = $4::char(3),
			fees             = $5::numeric,
			transaction_date = $6::date,
			notes            = $7,
			updated_at       = NOW()
		WHERE id = $8
		  AND entry_id IN (
			SELECT pe.id FROM portfolio_entries pe
			JOIN portfolios p ON p.id = pe.portfolio_id
			WHERE p.user_id = $9
		  )
		RETURNING id, entry_id, type, quantity, price, currency, fees, transaction_date, COALESCE(notes, ''), created_at, updated_at
	`, txnType, quantity.String(), price.String(), currency, fees.String(), transactionDate, notes, txnID, userID).Scan(
		&txn.ID,
		&txn.EntryID,
		&txn.Type,
		&quantityValue,
		&priceValue,
		&txn.Currency,
		&feesValue,
		&txn.TransactionDate,
		&txn.Notes,
		&txn.CreatedAt,
		&txn.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Transaction{}, ErrTransactionNotFound
		}
		return Transaction{}, err
	}
	txn.Quantity = decimal.MustFromString(quantityValue)
	txn.Price = moneyOf(priceValue, txn.Currency)
	txn.Fees = moneyOf(feesValue, txn.Currency)
	return txn, nil
}

// DeleteTransaction removes one transaction the caller owns. The position it
// belonged to is left in place with quantity 0 if it was the last one:
// trg_recalculate_avg_cost (migration 000023) recomputes the entry from what
// remains, so nothing here has to touch portfolio_entries.
//
// Ownership is enforced in the WHERE clause rather than by a prior read, so a
// transaction id belonging to somebody else is indistinguishable from one that
// does not exist — both delete no rows and answer 404.
func (r *PostgresRepository) DeleteTransaction(ctx context.Context, userID, txnID uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `
		DELETE FROM transactions
		WHERE id = $1
		  AND entry_id IN (
			SELECT pe.id FROM portfolio_entries pe
			JOIN portfolios p ON p.id = pe.portfolio_id
			WHERE p.user_id = $2
		  )
	`, txnID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrTransactionNotFound
	}

	return nil
}

// ImportEntryTransactions persists a batch of validated spreadsheet rows in a
// single database transaction: each row resolves (or creates) its asset,
// upserts the portfolio position and inserts the transaction, so a mid-batch
// failure never leaves a half-imported file behind.
func (r *PostgresRepository) ImportEntryTransactions(ctx context.Context, userID, portfolioID, sourceID uuid.UUID, rows []ImportTransactionRow) (int, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var owned bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS (SELECT 1 FROM portfolios WHERE id = $1 AND user_id = $2)
		   AND EXISTS (SELECT 1 FROM investment_sources WHERE id = $3 AND user_id = $2)
	`, portfolioID, userID, sourceID).Scan(&owned); err != nil {
		return 0, err
	}
	if !owned {
		return 0, ErrPortfolioOrSourceNotFound
	}

	// Cache asset lookups: classic spreadsheets repeat the same ticker on
	// many rows.
	assetIDs := make(map[string]uuid.UUID)
	imported := 0
	for _, row := range rows {
		assetID, ok := assetIDs[row.Ticker]
		if !ok {
			// Reuse an existing asset for the ticker (regardless of exchange)
			// before creating a new one, so imports never clobber curated
			// asset data the way an upsert would.
			err := tx.QueryRow(ctx, `
				SELECT id FROM assets WHERE UPPER(ticker) = $1 ORDER BY created_at LIMIT 1
			`, row.Ticker).Scan(&assetID)
			if errors.Is(err, pgx.ErrNoRows) {
				// created_by, and the membership row below, are what a
				// contributed asset needs to be visible to its contributor:
				// the catalog only serves curated rows to everybody. Without
				// them an imported ticker would be in the user's portfolio and
				// missing from the picker they add the next one with.
				err = tx.QueryRow(ctx, `
					INSERT INTO assets (ticker, name, asset_type, exchange, currency, created_by, is_curated, created_at, updated_at)
					VALUES ($1, $2, $3::asset_type, NULL, $4, $5, FALSE, NOW(), NOW())
					ON CONFLICT (ticker, COALESCE(exchange, '')) DO UPDATE SET updated_at = NOW()
					RETURNING id
				`, row.Ticker, row.AssetName, row.AssetType, row.Currency, userID).Scan(&assetID)
			}
			if err != nil {
				return 0, err
			}

			if _, err := tx.Exec(ctx, `
				INSERT INTO user_catalog_assets (user_id, asset_id) VALUES ($1, $2)
				ON CONFLICT DO NOTHING
			`, userID, assetID); err != nil {
				return 0, err
			}

			assetIDs[row.Ticker] = assetID
		}

		var entryID uuid.UUID
		if err := tx.QueryRow(ctx, `
			INSERT INTO portfolio_entries (portfolio_id, asset_id, source_id, quantity, price, cost_currency, entry_date, notes)
			VALUES ($1::uuid, $2::uuid, $3::uuid, 0, $4::numeric, $5::char(3), $6::date, $7)
			ON CONFLICT (portfolio_id, asset_id, COALESCE(source_id::TEXT, ''))
			DO UPDATE SET updated_at = NOW()
			RETURNING id
		`, portfolioID, assetID, sourceID, row.Price.String(), row.Currency, row.Date, row.Notes).Scan(&entryID); err != nil {
			return 0, err
		}

		if _, err := tx.Exec(ctx, `
			INSERT INTO transactions (entry_id, type, quantity, price, currency, fees, transaction_date, notes)
			VALUES ($1::uuid, $2::transaction_type, $3::numeric, $4::numeric, $5::char(3), $6::numeric, $7::date, $8)
		`, entryID, row.Type, row.Quantity.String(), row.Price.String(), row.Currency, row.Fees.String(), row.Date, row.Notes); err != nil {
			return 0, err
		}
		imported++
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return imported, nil
}
