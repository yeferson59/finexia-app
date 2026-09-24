package portfolio

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/market"
	"github.com/yeferson59/finexia-app/internal/platform/database"
)

func (r *PostgresRepository) GetTopTransactionByPortfolioID(ctx context.Context, userID, portfolioID uuid.UUID) (TopTransactionDTO, error) {
	var dto TopTransactionDTO
	// The amount is the position's, not the trade's: ranking by the quoted price
	// puts a 606 EUR fill above an 800 USD one, and reporting the winner under
	// the currency it was quoted in makes the dashboard's "largest transaction"
	// a figure that cannot be compared with the totals beside it. fx_rate turns
	// both into what the account paid, which is the only basis on which two
	// trades in different currencies can be ordered at all.
	err := r.db.QueryRow(ctx, `
		SELECT
			(t.quantity::numeric * t.price::numeric * t.fx_rate::numeric)::text,
			t.type,
			pe.cost_currency,
			t.transaction_date,
			a.ticker,
			a.name
		FROM transactions t
		JOIN portfolio_entries pe ON pe.id = t.entry_id
		JOIN assets a ON a.id = pe.asset_id
		JOIN portfolios p ON p.id = pe.portfolio_id
		WHERE pe.portfolio_id = $1 AND p.user_id = $2
		ORDER BY t.quantity::numeric * t.price::numeric * t.fx_rate::numeric DESC
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

// GetRecentTransactionsByUserID lists the user's latest transactions: the
// activity feed, the transactions page, the export and the MCP tool.
//
// The cash side of another transaction is left out: the credit of a dividend or
// a sale, the debit of a purchase. That transaction is already on the list, and
// listing the money it moved through cash beside it would read as two
// operations. The cash screens are where those rows show.
func (r *PostgresRepository) GetRecentTransactionsByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]Transaction, error) {
	rows, err := r.db.Query(ctx, `
		SELECT t.id, t.entry_id, t.type, t.quantity, t.price, t.currency, t.fx_rate,
		       pe.cost_currency, t.fees, t.fees_currency,
		       t.transaction_date, COALESCE(t.notes, ''), t.created_at, t.updated_at,
		       a.ticker, a.name
		FROM transactions t
		JOIN portfolio_entries pe ON pe.id = t.entry_id
		JOIN portfolios p ON p.id = pe.portfolio_id
		JOIN assets a ON a.id = pe.asset_id
		WHERE p.user_id = $1
		  AND t.type NOT IN ('cash_dividend', 'cash_sale', 'cash_purchase')
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

		if err := rows.Scan(
			&txn.ID,
			&txn.EntryID,
			&txn.Type,
			&txn.Quantity,
			&txn.Price,
			&txn.Currency,
			&txn.FXRate,
			&txn.CostCurrency,
			&txn.Fees,
			&txn.FeesCurrency,
			&txn.TransactionDate,
			&txn.Notes,
			&txn.CreatedAt,
			&txn.UpdatedAt,
			&txn.Entry.Asset.Ticker,
			&txn.Entry.Asset.Name,
		); err != nil {
			return nil, err
		}

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
// portfolio_entries.category. That column was a copy of the type taken when the
// entry was created, which nothing wrote again, and 000026 dropped it.
// Correcting an asset afterwards (a bond ETF first filed as a plain ETF) moved
// it in the per-portfolio donut, which groups the holdings by their type, and left
// it in the old slice here: two charts over the same positions, disagreeing.
// Reading assets.asset_type is what makes both answer with one rule.
//
// The rows come back keyed by asset type, and that is the vocabulary the
// response speaks: AllocationItem.Category is a market.AssetType. It is the
// singular one — "stock", not "stocks" — which is *not* the plural vocabulary
// of portfolio.Type; clients map labels and colours from the former.
func (r *PostgresRepository) GetAssetAllocationByUserID(ctx context.Context, userID uuid.UUID, targetCurrency money.Currency) ([]AllocationItem, error) {
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
	`, userID, currencyParam(targetCurrency))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rowsByType := make([]AllocationItem, 0)
	for rows.Next() {
		var item AllocationItem

		if err := rows.Scan(&item.Category, &item.MarketValue, &item.Currency, &item.PositionsUnconverted); err != nil {
			return nil, err
		}

		rowsByType = append(rowsByType, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return foldAllocationByCategory(rowsByType), nil
}

// foldAllocationByCategory merges the rows that share an asset type and
// restores the by-value ordering the query asked for. It does not translate the
// vocabulary: the type the query grouped by is what the response carries.
//
// Two rows can land on one category — the importer files every label it cannot
// place as Other. Adding them is the only reading that keeps the shares summing
// to the whole — letting the second row win would drop a slice's worth of money
// off the chart — and a merged row can outgrow the one above it, which is why
// the order is rebuilt rather than left almost-sorted.
func foldAllocationByCategory(items []AllocationItem) []AllocationItem {
	folded := make([]AllocationItem, 0, len(items))
	index := make(map[market.AssetType]int, len(items))

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
		SELECT t.id, t.entry_id, t.type, t.quantity, t.price, t.currency, t.fx_rate,
		       pe.cost_currency, t.fees, t.fees_currency,
		       t.transaction_date, COALESCE(t.notes, ''), t.created_at, t.updated_at,
		       EXISTS (SELECT 1 FROM transactions c WHERE c.credited_from = t.id AND c.type <> 'cash_purchase'),
		       EXISTS (SELECT 1 FROM transactions c WHERE c.credited_from = t.id AND c.type =  'cash_purchase'),
		       -- Which drawer the purchase was paid from, so the edit form can send
		       -- the same answer back instead of moving the money to the main one.
		       (SELECT ce.pocket_id
		          FROM transactions c
		          JOIN portfolio_entries ce ON ce.id = c.entry_id
		         WHERE c.credited_from = t.id AND c.type = 'cash_purchase')
		FROM transactions t
		JOIN portfolio_entries pe ON pe.id = t.entry_id
		WHERE t.entry_id = $1
		ORDER BY t.transaction_date DESC, t.created_at DESC
	`, entryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	txns := make([]Transaction, 0)
	for rows.Next() {
		var txn Transaction

		if err := rows.Scan(
			&txn.ID,
			&txn.EntryID,
			&txn.Type,
			&txn.Quantity,
			&txn.Price,
			&txn.Currency,
			&txn.FXRate,
			&txn.CostCurrency,
			&txn.Fees,
			&txn.FeesCurrency,
			&txn.TransactionDate,
			&txn.Notes,
			&txn.CreatedAt,
			&txn.UpdatedAt,
			&txn.CashCredited,
			&txn.CashPaid,
			&txn.CashPocketID,
		); err != nil {
			return nil, err
		}

		txns = append(txns, txn)
	}

	if err := rows.Err(); err != nil {
		return nil, err
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
		SELECT t.id, t.entry_id, t.type, t.quantity, t.price, t.currency, t.fx_rate,
		       pe.cost_currency, t.fees, t.fees_currency,
		       t.transaction_date, COALESCE(t.notes, ''), t.created_at, t.updated_at,
		       EXISTS (SELECT 1 FROM transactions c WHERE c.credited_from = t.id AND c.type <> 'cash_purchase'),
		       EXISTS (SELECT 1 FROM transactions c WHERE c.credited_from = t.id AND c.type =  'cash_purchase'),
		       -- Which drawer the purchase was paid from, so the edit form can send
		       -- the same answer back instead of moving the money to the main one.
		       (SELECT ce.pocket_id
		          FROM transactions c
		          JOIN portfolio_entries ce ON ce.id = c.entry_id
		         WHERE c.credited_from = t.id AND c.type = 'cash_purchase')
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

		if err := rows.Scan(
			&txn.ID,
			&txn.EntryID,
			&txn.Type,
			&txn.Quantity,
			&txn.Price,
			&txn.Currency,
			&txn.FXRate,
			&txn.CostCurrency,
			&txn.Fees,
			&txn.FeesCurrency,
			&txn.TransactionDate,
			&txn.Notes,
			&txn.CreatedAt,
			&txn.UpdatedAt,
			&txn.CashCredited,
			&txn.CashPaid,
			&txn.CashPocketID,
		); err != nil {
			return nil, err
		}

		txns = append(txns, txn)
	}

	return txns, nil
}

// CreateTransaction records one trade on an existing position.
//
// The position's cost currency is read first, and not only to be handed to
// Validate: that read is also the ownership check, since an entry the caller
// does not own returns no row. Both it and the insert run on tx — the previous
// version issued them on r.db, so the surrounding transaction was opened,
// committed and never actually used.
//
// A dividend or a sale stated with CreditCash is paid into the platform's cash
// in the same transaction, and a purchase stated with PayFromCash is paid out
// of it, so the two rows exist together or not at all. The position's other
// cash rows are re-synced after it: a purchase moves the average cost its
// credited sales carried out.
func (r *PostgresRepository) CreateTransaction(ctx context.Context, userID, entryID uuid.UUID, in TransactionInput) (Transaction, error) {
	var txn Transaction

	if err := database.WithinTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		txn, err = createTransactionTx(ctx, tx, userID, entryID, in, false)

		return err
	}); err != nil {
		return txn, err
	}

	return txn, nil
}

// createTransactionTx is CreateTransaction inside a transaction the caller
// holds. fundWrite is a write of a fund followed by balance, the one writer
// requireWritableEntry lets through to such a fund.
func createTransactionTx(ctx context.Context, tx pgx.Tx, userID, entryID uuid.UUID, in TransactionInput, fundWrite bool) (Transaction, error) {
	var txn Transaction

	var costCurrency money.Currency
	if err := tx.QueryRow(ctx, `
	SELECT pe.cost_currency
	FROM portfolio_entries pe
	JOIN portfolios p ON p.id = pe.portfolio_id
	WHERE pe.id = $1 AND p.user_id = $2
`, entryID, userID).Scan(&costCurrency); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return txn, ErrEntryNotFound
		}

		return txn, err
	}

	settled, err := in.Validate(costCurrency)
	if err != nil {
		return txn, err
	}

	if err := requireTypeAllowed(ctx, tx, entryID, settled.Type); err != nil {
		return txn, err
	}

	// A fund followed by balance writes its own purchases and sales, whose
	// quantities its replay derives; only the generic writers are refused.
	if !fundWrite {
		if err := requireWritableEntry(ctx, tx, entryID); err != nil {
			return txn, err
		}
	}

	if err := tx.QueryRow(ctx, `
	INSERT INTO transactions (entry_id, type, quantity, price, currency, fx_rate, fees, fees_currency, transaction_date, notes)
	VALUES ($1::uuid, $2::transaction_type, $3::numeric, $4::numeric, $5::char(3), $6::numeric, $7::numeric, $8::char(3), $9::date, $10)
	RETURNING id, entry_id, type, quantity, price, currency, fx_rate, fees, fees_currency, transaction_date, COALESCE(notes, ''), created_at, updated_at
`, entryID, settled.Type, settled.Quantity.String(), settled.Price.String(), settled.Currency,
		settled.FXRate.String(), settled.Fees.String(), settled.FeesCurrency, settled.TransactionDate, settled.Notes).Scan(
		&txn.ID,
		&txn.EntryID,
		&txn.Type,
		&txn.Quantity,
		&txn.Price,
		&txn.Currency,
		&txn.FXRate,
		&txn.Fees,
		&txn.FeesCurrency,
		&txn.TransactionDate,
		&txn.Notes,
		&txn.CreatedAt,
		&txn.UpdatedAt,
	); err != nil {
		return txn, err
	}

	txn.CostCurrency = costCurrency
	txn.Price.SetCurrency(txn.Currency)
	txn.Fees.SetCurrency(txn.FeesCurrency)

	if settled.CreditCash || settled.PayFromCash {
		if err := syncCashLink(ctx, tx, userID, txn.ID, true, statedPocket(settled.CashPocketID)); err != nil {
			return txn, err
		}

		txn.CashCredited = settled.CreditCash
		txn.CashPaid = settled.PayFromCash
		txn.CashPocketID = paidFromPocket(settled)
	}

	return txn, syncEntryCashLinks(ctx, tx, userID, entryID, true)
}

// paidFromPocket is the drawer a written transaction reads back as paid from:
// the one it stated, or none — the main account, or no cash at all.
func paidFromPocket(in TransactionInput) *uuid.UUID {
	if !in.PayFromCash || in.CashPocketID == (uuid.UUID{}) {
		return nil
	}

	pocket := in.CashPocketID

	return &pocket
}

// UpdateTransaction rewrites one transaction the caller owns.
//
// It grew a surrounding database transaction for the same reason CreateTransaction
// has one: the rate can only be judged against the position's cost currency, so
// that has to be read before the write, and reading it separately from a write
// that depends on it is only safe if the two are atomic. The UPDATE keeps its
// own ownership predicate anyway — the read cannot be trusted to still hold by
// the time the write lands, and the cost of restating it is one join.
//
// The row's cash side is rewritten after it, from the row as it now stands:
// kept and resized while CreditCash or PayFromCash still asks for it, taken out
// when neither does or the row stopped being a kind that settles against cash.
// That is also how a purchase recorded before any of this was offered comes to
// be paid from cash, and how one paid by mistake stops being: the edit states
// the answer and the debit appears or goes. Then the position's other cash
// rows, whose sales' cost the edit may have moved. A cash row cannot be edited
// on its own.
func (r *PostgresRepository) UpdateTransaction(ctx context.Context, userID, txnID uuid.UUID, in TransactionInput) (Transaction, error) {
	var txn Transaction

	if err := database.WithinTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		var (
			costCurrency money.Currency
			entryID      uuid.UUID
			currentType  TransactionType
		)
		if err := tx.QueryRow(ctx, `
		SELECT pe.cost_currency, pe.id, t.type
		FROM transactions t
		JOIN portfolio_entries pe ON pe.id = t.entry_id
		JOIN portfolios p        ON p.id = pe.portfolio_id
		WHERE t.id = $1 AND p.user_id = $2
	`, txnID, userID).Scan(&costCurrency, &entryID, &currentType); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrTransactionNotFound
			}

			return err
		}

		if currentType.isCashLinked() {
			return ErrCashCreditLinked
		}

		settled, err := in.Validate(costCurrency)
		if err != nil {
			return err
		}

		if err := requireTypeAllowed(ctx, tx, entryID, settled.Type); err != nil {
			return err
		}

		if err := requireWritableEntry(ctx, tx, entryID); err != nil {
			return err
		}

		if err := tx.QueryRow(ctx, `
		UPDATE transactions SET
			type             = $1::transaction_type,
			quantity         = $2::numeric,
			price            = $3::numeric,
			currency         = $4::char(3),
			fx_rate          = $5::numeric,
			fees             = $6::numeric,
			fees_currency    = $7::char(3),
			transaction_date = $8::date,
			notes            = $9,
			updated_at       = NOW()
		WHERE id = $10
		  AND entry_id IN (
			SELECT pe.id FROM portfolio_entries pe
			JOIN portfolios p ON p.id = pe.portfolio_id
			WHERE p.user_id = $11
		  )
		RETURNING id, entry_id, type, quantity, price, currency, fx_rate, fees, fees_currency, transaction_date, COALESCE(notes, ''), created_at, updated_at
	`, settled.Type, settled.Quantity.String(), settled.Price.String(), settled.Currency, settled.FXRate.String(),
			settled.Fees.String(), settled.FeesCurrency, settled.TransactionDate, settled.Notes, txnID, userID).Scan(
			&txn.ID,
			&txn.EntryID,
			&txn.Type,
			&txn.Quantity,
			&txn.Price,
			&txn.Currency,
			&txn.FXRate,
			&txn.Fees,
			&txn.FeesCurrency,
			&txn.TransactionDate,
			&txn.Notes,
			&txn.CreatedAt,
			&txn.UpdatedAt,
		); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrTransactionNotFound
			}

			return err
		}

		txn.CostCurrency = costCurrency
		txn.Price.SetCurrency(txn.Currency)
		txn.Fees.SetCurrency(txn.FeesCurrency)

		if err := syncCashLink(ctx, tx, userID, txnID, settled.CreditCash || settled.PayFromCash, statedPocket(settled.CashPocketID)); err != nil {
			return err
		}

		txn.CashCredited = settled.CreditCash
		txn.CashPaid = settled.PayFromCash
		txn.CashPocketID = paidFromPocket(settled)

		return syncEntryCashLinks(ctx, tx, userID, entryID, true)
	}); err != nil {
		return Transaction{}, err
	}

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
//
// The cash row of a dividend, a sale or a purchase goes with it — the credit
// once its balance is known to spare the money, the debit by giving it back —
// and that row on its own cannot be deleted here. Deleting a purchase moves the
// cost the position's credited sales carried out, so those are re-synced
// after.
func (r *PostgresRepository) DeleteTransaction(ctx context.Context, userID, txnID uuid.UUID) error {
	return database.WithinTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		// A row of a fixed deposit is refused before the delete rather than left
		// out of it, so it answers "that is a deposit" and not "there is no such
		// transaction". The read and the delete share a transaction so the
		// position cannot become one in between.
		var (
			entryID uuid.UUID
			txnType TransactionType
		)
		if err := tx.QueryRow(ctx, `
			SELECT t.entry_id, t.type
			FROM transactions t
			JOIN portfolio_entries pe ON pe.id = t.entry_id
			JOIN portfolios p         ON p.id = pe.portfolio_id
			WHERE t.id = $1 AND p.user_id = $2
		`, txnID, userID).Scan(&entryID, &txnType); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrTransactionNotFound
			}

			return err
		}

		if txnType.isCashLinked() {
			return ErrCashCreditLinked
		}

		if err := requireWritableEntry(ctx, tx, entryID); err != nil {
			return err
		}

		if err := syncCashLink(ctx, tx, userID, txnID, false, cashPlacement{}); err != nil {
			return err
		}

		tag, err := tx.Exec(ctx, `DELETE FROM transactions WHERE id = $1`, txnID)
		if err != nil {
			return err
		}

		if tag.RowsAffected() == 0 {
			return ErrTransactionNotFound
		}

		return syncEntryCashLinks(ctx, tx, userID, entryID, true)
	})
}

// ImportEntryTransactions persists a batch of validated spreadsheet rows in a
// single database transaction: each row resolves (or creates) its asset,
// upserts the portfolio position and inserts the transaction, so a mid-batch
// failure never leaves a half-imported file behind.
func (r *PostgresRepository) ImportEntryTransactions(ctx context.Context, userID, portfolioID, sourceID uuid.UUID, rows []ImportTransactionRow) (int, error) {
	var imported int

	if err := database.WithinTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		var owned bool

		if err := tx.QueryRow(ctx, `
		SELECT EXISTS (SELECT 1 FROM portfolios WHERE id = $1 AND user_id = $2)
		   AND EXISTS (SELECT 1 FROM investment_sources WHERE id = $3 AND user_id = $2)
	`, portfolioID, userID, sourceID).Scan(&owned); err != nil {
			return err
		}

		if !owned {
			return ErrPortfolioOrSourceNotFound
		}

		// Cache asset lookups: classic spreadsheets repeat the same ticker on
		// many rows.
		assetIDs := make(map[string]uuid.UUID)
		// The positions the file wrote to, in the order it first reached them.
		// A purchase moves the cost their credited sales carried out.
		var touched []uuid.UUID
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
					return err
				}

				if _, err := tx.Exec(ctx, `
				INSERT INTO user_catalog_assets (user_id, asset_id) VALUES ($1, $2)
				ON CONFLICT DO NOTHING
			`, userID, assetID); err != nil {
					return err
				}

				assetIDs[row.Ticker] = assetID
			}

			// The position's currency comes back from the upsert for the same
			// reason it does in CreatePortfolioEntry: a ticker already held in
			// this portfolio keeps the currency it was opened with, and a file
			// declaring a different account currency must not be allowed to
			// reinterpret it.
			var entryID uuid.UUID
			var entryCostCurrency money.Currency
			if err := tx.QueryRow(ctx, `
			INSERT INTO portfolio_entries (portfolio_id, asset_id, source_id, quantity, price, cost_currency, entry_date, notes)
			VALUES ($1::uuid, $2::uuid, $3::uuid, 0, $4::numeric, $5::char(3), $6::date, $7)
			ON CONFLICT (portfolio_id, asset_id, COALESCE(source_id::TEXT, '')) WHERE pocket_id IS NULL
			DO UPDATE SET updated_at = NOW()
			RETURNING id, cost_currency
		`, portfolioID, assetID, sourceID, row.Price.MulDecimal(row.FXRate).String(), row.CostCurrency, row.Date, row.Notes).Scan(&entryID, &entryCostCurrency); err != nil {
				return err
			}

			// The importer builds its own per-row messages in Spanish, but this
			// is the check that actually gates the write, so the file path and
			// the HTTP path cannot disagree about what a legal row is. Reaching
			// it means the preview and the commit saw different data, which
			// should fail the batch rather than write half of it.
			settled, err := row.Input().Validate(entryCostCurrency)
			if err != nil {
				return fmt.Errorf("row %d: %w", row.RowNumber, err)
			}

			if _, err := tx.Exec(ctx, `
			INSERT INTO transactions (entry_id, type, quantity, price, currency, fx_rate, fees, fees_currency, transaction_date, notes)
			VALUES ($1::uuid, $2::transaction_type, $3::numeric, $4::numeric, $5::char(3), $6::numeric, $7::numeric, $8::char(3), $9::date, $10)
		`, entryID, settled.Type, settled.Quantity.String(), settled.Price.String(), settled.Currency, settled.FXRate.String(),
				settled.Fees.String(), settled.FeesCurrency, settled.TransactionDate, settled.Notes); err != nil {
				return err
			}

			if !slices.Contains(touched, entryID) {
				touched = append(touched, entryID)
			}

			imported++
		}

		for _, entryID := range touched {
			if err := syncEntryCashLinks(ctx, tx, userID, entryID, true); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return 0, err
	}

	return imported, nil
}
