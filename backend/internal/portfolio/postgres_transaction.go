package portfolio

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"
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
		txn.Price = money.MustMoneyFromString(price, money.USD)
		txn.Fees = money.MustMoneyFromString(fees, money.USD)
		txns = append(txns, txn)
	}
	return txns, nil
}

func (r *PostgresRepository) GetAssetAllocationByUserID(ctx context.Context, userID uuid.UUID) ([]AllocationItem, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			pe.category,
			COALESCE(SUM(pe.quantity::numeric * COALESCE(a.current_price::numeric, pe.price::numeric)), 0)::text AS market_value
		FROM portfolio_entries pe
		JOIN portfolios p ON p.id = pe.portfolio_id
		JOIN assets a ON a.id = pe.asset_id
		WHERE p.user_id = $1
		  AND pe.quantity::numeric > 0
		GROUP BY pe.category
		ORDER BY COALESCE(SUM(pe.quantity::numeric * COALESCE(a.current_price::numeric, pe.price::numeric)), 0) DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]AllocationItem, 0)
	for rows.Next() {
		var item AllocationItem
		if err := rows.Scan(&item.Category, &item.MarketValue); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
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
		txn.Price = money.MustMoneyFromString(price, money.USD)
		txn.Fees = money.MustMoneyFromString(fees, money.USD)
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
		txn.Price = money.MustMoneyFromString(price, money.USD)
		txn.Fees = money.MustMoneyFromString(fees, money.USD)
		txns = append(txns, txn)
	}
	return txns, nil
}

func (r *PostgresRepository) CreateTransaction(ctx context.Context, userID, entryID uuid.UUID, txnType TransactionType, quantity money.Decimal, price money.Money, currency string, fees money.Money, transactionDate time.Time, notes string) (Transaction, error) {
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
	txn.Price = money.MustMoneyFromString(priceValue, money.USD)
	txn.Fees = money.MustMoneyFromString(feesValue, money.USD)
	return txn, nil
}

func (r *PostgresRepository) UpdateTransaction(ctx context.Context, userID, txnID uuid.UUID, txnType TransactionType, quantity money.Decimal, price money.Money, currency string, fees money.Money, transactionDate time.Time, notes string) (Transaction, error) {
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
	txn.Price = money.MustMoneyFromString(priceValue, money.USD)
	txn.Fees = money.MustMoneyFromString(feesValue, money.USD)
	return txn, nil
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
				err = tx.QueryRow(ctx, `
					INSERT INTO assets (ticker, name, asset_type, exchange, currency, created_at, updated_at)
					VALUES ($1, $2, $3::asset_type, NULL, $4, NOW(), NOW())
					ON CONFLICT (ticker, COALESCE(exchange, '')) DO UPDATE SET updated_at = NOW()
					RETURNING id
				`, row.Ticker, row.AssetName, row.AssetType, row.Currency).Scan(&assetID)
			}
			if err != nil {
				return 0, err
			}
			assetIDs[row.Ticker] = assetID
		}

		var entryID uuid.UUID
		if err := tx.QueryRow(ctx, `
			INSERT INTO portfolio_entries (portfolio_id, asset_id, source_id, quantity, price, cost_currency, category, entry_date, notes)
			VALUES ($1::uuid, $2::uuid, $3::uuid, 0, $4::numeric, $5::char(3), $6::portfolio_entry_category, $7::date, $8)
			ON CONFLICT (portfolio_id, asset_id, COALESCE(source_id::TEXT, ''))
			DO UPDATE SET updated_at = NOW()
			RETURNING id
		`, portfolioID, assetID, sourceID, row.Price.String(), row.Currency, row.Category, row.Date, row.Notes).Scan(&entryID); err != nil {
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
