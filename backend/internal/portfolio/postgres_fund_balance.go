package portfolio

import (
	"context"
	"errors"
	"fmt"
	"time"

	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/platform/database"
)

// The funds followed by balance (000058). Their contributions and withdrawals
// are purchases and sales whose quantities are derived: every write stores the
// fact the owner stated — money, or a balance — and then replays the whole fund
// (fund_units.go) and writes back what the replay derived. It is the whole fund
// and not only what follows the write because a balance values the units of
// every portfolio at once (D7), and a fund has hundreds of rows, not thousands.
//
// Every write locks the fund first (lockFund), so two writes to one fund
// replay one after the other and neither overwrites the other's figures.

// skipped reports whether the replay left the balance of day out because the
// fund held nothing that day.
func (r fundReplay) skipped(day time.Time) bool {
	for _, m := range r.Marks {
		if m.Date.Equal(day) {
			return m.Skipped
		}
	}

	return false
}

// recordFundMovement stores the money a contribution or withdrawal was.
func recordFundMovement(ctx context.Context, tx pgx.Tx, txnID uuid.UUID, amount decimal.Decimal, all bool) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO fund_movements (txn_id, amount, withdraws_all) VALUES ($1, $2::numeric, $3)
		ON CONFLICT (txn_id) DO UPDATE SET amount = EXCLUDED.amount, withdraws_all = EXCLUDED.withdraws_all
	`, txnID, amount.String(), all)

	return err
}

// replayAndRestate replays a fund followed by balance and writes back what it
// derived: the quantity and price of every movement, the unit value of every
// balance, the cash rows of the movements paid from or into cash, and the
// price the valuation reads. The snapshots are then revalued from the earlier
// of from — the day the write touched — and the first balance whose unit value
// moved.
func replayAndRestate(ctx context.Context, tx pgx.Tx, userID, assetID uuid.UUID, from time.Time) (fundReplay, error) {
	movements, entries, err := readFundMovementFacts(ctx, tx, userID, assetID)
	if err != nil {
		return fundReplay{}, err
	}

	balances, err := readFundBalanceFacts(ctx, tx, userID, assetID)
	if err != nil {
		return fundReplay{}, err
	}

	replay, err := replayBalanceFund(movements, balances)
	if err != nil {
		return fundReplay{}, err
	}

	for _, m := range replay.Movements {
		if _, err := tx.Exec(ctx, `
			UPDATE transactions
			   SET quantity = $2::numeric, price = $3::numeric, updated_at = NOW()
			 WHERE id = $1 AND (quantity <> $2::numeric OR price <> $3::numeric)
		`, m.TxnID, m.Quantity.String(), m.Price.String()); err != nil {
			return fundReplay{}, err
		}
	}

	for _, m := range replay.Marks {
		if m.Skipped {
			continue
		}

		tag, err := tx.Exec(ctx, `
			UPDATE fund_marks
			   SET unit_value = $4::numeric, updated_at = NOW()
			 WHERE user_id = $1 AND asset_id = $2 AND mark_date = $3::date AND unit_value <> $4::numeric
		`, userID, assetID, m.Date.Format(time.DateOnly), m.UnitValue.String())
		if err != nil {
			return fundReplay{}, err
		}

		if tag.RowsAffected() > 0 && m.Date.Before(from) {
			from = m.Date
		}
	}

	// A purchase paid from cash, or a withdrawal paid into it, moved as much
	// money as its quantity × price said; the replay may have moved both.
	for _, entryID := range entries {
		if err := syncEntryCashLinks(ctx, tx, userID, entryID, true); err != nil {
			return fundReplay{}, err
		}
	}

	if err := syncFundPrice(ctx, tx, userID, assetID); err != nil {
		return fundReplay{}, err
	}

	return replay, restateFundSnapshots(ctx, tx, userID, assetID, from)
}

// readFundMovementFacts reads every purchase and sale of the fund in the
// user's portfolios, with the money it was, and the positions they are on. A
// transaction with no stored fact — written before the fund was followed by
// balance — counts as what its quantity and price say.
func readFundMovementFacts(ctx context.Context, tx pgx.Tx, userID, assetID uuid.UUID) ([]fundMovementFact, []uuid.UUID, error) {
	rows, err := tx.Query(ctx, `
		SELECT t.id, t.entry_id, t.type, t.transaction_date, t.created_at,
		       COALESCE(fm.amount, t.quantity * t.price)::text, COALESCE(fm.withdraws_all, FALSE)
		FROM transactions t
		JOIN portfolio_entries pe ON pe.id = t.entry_id
		JOIN portfolios p         ON p.id = pe.portfolio_id
		LEFT JOIN fund_movements fm ON fm.txn_id = t.id
		WHERE p.user_id = $1 AND pe.asset_id = $2 AND t.type IN ('buy', 'sell')
	`, userID, assetID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var (
		facts   []fundMovementFact
		entries []uuid.UUID
		seen    = make(map[uuid.UUID]bool)
	)

	for rows.Next() {
		var (
			f       fundMovementFact
			txnType TransactionType
			amount  string
		)

		if err := rows.Scan(&f.TxnID, &f.EntryID, &txnType, &f.Date, &f.CreatedAt, &amount, &f.All); err != nil {
			return nil, nil, err
		}

		if f.Amount, err = decimal.NewFromString(amount); err != nil {
			return nil, nil, err
		}

		f.Date = cashRateDay(f.Date)
		f.Withdrawal = txnType == Sell
		facts = append(facts, f)

		if !seen[f.EntryID] {
			seen[f.EntryID] = true
			entries = append(entries, f.EntryID)
		}
	}

	return facts, entries, rows.Err()
}

func readFundBalanceFacts(ctx context.Context, tx pgx.Tx, userID, assetID uuid.UUID) ([]fundBalanceFact, error) {
	rows, err := tx.Query(ctx, `
		SELECT mark_date, balance::text
		FROM fund_marks
		WHERE user_id = $1 AND asset_id = $2 AND balance IS NOT NULL
	`, userID, assetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var facts []fundBalanceFact

	for rows.Next() {
		var (
			f       fundBalanceFact
			balance string
		)

		if err := rows.Scan(&f.Date, &balance); err != nil {
			return nil, err
		}

		if f.Balance, err = decimal.NewFromString(balance); err != nil {
			return nil, err
		}

		f.Date = cashRateDay(f.Date)
		facts = append(facts, f)
	}

	return facts, rows.Err()
}

// lockFundWithCurrency locks a fund and answers how it is followed and the
// currency it is in.
func lockFundWithCurrency(ctx context.Context, tx pgx.Tx, userID, assetID uuid.UUID) (FundTracking, money.Currency, error) {
	tracking, err := lockFund(ctx, tx, userID, assetID)
	if err != nil {
		return "", money.XXX, err
	}

	var cur money.Currency
	err = tx.QueryRow(ctx, `SELECT currency FROM assets WHERE id = $1`, assetID).Scan(&cur)

	return tracking, cur, err
}

// ContributeToFund puts money into a fund, on the position its portfolio holds
// on the platform — opened if there is none.
//
// In a fund followed by balance the purchase is written provisionally as its
// amount at a price of one, so it moves exactly its amount through cash when
// it is paid from there; the replay then gives it its real units at the unit
// value of its day. In one followed by units it is the units at the unit value
// the owner stated (contributeUnits).
func (r *PostgresRepository) ContributeToFund(ctx context.Context, userID, assetID uuid.UUID, in FundContributionInput) (FundMovement, error) {
	var movement FundMovement

	if err := database.WithinTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		tracking, cur, err := lockFundWithCurrency(ctx, tx, userID, assetID)
		if err != nil {
			return err
		}

		if tracking == FundUnits {
			movement, err = contributeUnits(ctx, tx, userID, assetID, cur, in)

			return err
		}

		day := cashRateDay(in.Date)
		from := day

		var before time.Time
		if in.BalanceBefore.IsPos() {
			before = day.AddDate(0, 0, -1)
			from = before

			if _, err := writeFundMark(ctx, tx, userID, assetID, FundMarkInput{
				Date: before, UnitValue: fundOpeningUnitValue, Balance: in.BalanceBefore,
			}); err != nil {
				return err
			}
		}

		_, txnID, err := createPortfolioEntryTx(ctx, tx, userID, in.PortfolioID, assetID, in.SourceID, cur, TransactionInput{
			Type:            Buy,
			Quantity:        in.Amount,
			Price:           money.NewFromDecimal(decimal.One, cur),
			Currency:        cur,
			TransactionDate: day,
			Notes:           in.Notes,
			PayFromCash:     in.PayFromCash,
			CashPocketID:    in.CashPocketID,
		})
		if err != nil {
			return err
		}

		if err := recordFundMovement(ctx, tx, txnID, in.Amount, false); err != nil {
			return err
		}

		replay, err := replayAndRestate(ctx, tx, userID, assetID, from)
		if err != nil {
			return err
		}

		if !before.IsZero() && replay.skipped(before) {
			return fmt.Errorf("%w: %w", ErrFundNoUnits, errBalanceBeforeFirst)
		}

		movement, err = readFundMovement(ctx, tx, userID, txnID)

		return err
	}); err != nil {
		return FundMovement{}, err
	}

	return movement, nil
}

// errBalanceBeforeFirst says why a balance before a contribution was refused:
// the fund held nothing the day before, so there was nothing to value.
var errBalanceBeforeFirst = errors.New("leave out the balance before the first contribution")

// WithdrawFromFund takes money out of one position of a fund.
//
// In a fund followed by balance the sale is written provisionally for every
// unit the position holds, at the price that makes it the amount, so a
// withdrawal paid into cash credits exactly that; the replay then gives it its
// real units, and refuses it if the position does not hold that much on its
// day. In one followed by units it is the units at the unit value the owner
// stated (withdrawUnits).
func (r *PostgresRepository) WithdrawFromFund(ctx context.Context, userID, assetID uuid.UUID, in FundWithdrawalInput) (FundMovement, error) {
	var movement FundMovement

	if err := database.WithinTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		tracking, cur, err := lockFundWithCurrency(ctx, tx, userID, assetID)
		if err != nil {
			return err
		}

		var held string
		if err := tx.QueryRow(ctx, `
			SELECT pe.quantity::text
			FROM portfolio_entries pe
			JOIN portfolios p ON p.id = pe.portfolio_id
			WHERE pe.id = $1 AND p.user_id = $2 AND pe.asset_id = $3
		`, in.EntryID, userID, assetID).Scan(&held); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrEntryNotFound
			}

			return err
		}

		units, err := decimal.NewFromString(held)
		if err != nil {
			return err
		}

		if tracking == FundUnits {
			movement, err = withdrawUnits(ctx, tx, userID, cur, units, in)

			return err
		}

		if !units.IsPos() {
			return ErrFundNotEnoughUnits
		}

		price, err := in.Amount.Div(units)
		if err != nil {
			return err
		}

		txn, err := createTransactionTx(ctx, tx, userID, in.EntryID, TransactionInput{
			Type:            Sell,
			Quantity:        units,
			Price:           money.NewFromDecimal(price.RoundHAZ(8), cur),
			Currency:        cur,
			Fees:            money.NewFromDecimal(in.Fees, cur),
			FeesCurrency:    cur,
			TransactionDate: cashRateDay(in.Date),
			Notes:           in.Notes,
			CreditCash:      in.CreditCash,
			CashPocketID:    in.CashPocketID,
		}, true)
		if err != nil {
			return err
		}

		if err := recordFundMovement(ctx, tx, txn.ID, in.Amount, in.All); err != nil {
			return err
		}

		if _, err := replayAndRestate(ctx, tx, userID, assetID, cashRateDay(in.Date)); err != nil {
			return err
		}

		movement, err = readFundMovement(ctx, tx, userID, txn.ID)

		return err
	}); err != nil {
		return FundMovement{}, err
	}

	return movement, nil
}

// lockedFundMovement is what an edit or a deletion of a movement has to know.
type lockedFundMovement struct {
	assetID  uuid.UUID
	entryID  uuid.UUID
	date     time.Time
	tracking FundTracking
}

// lockFundMovement finds a purchase or sale of a fund the user follows and
// locks the fund. Anything else — another user's, another asset's, a
// dividend — is not found.
func lockFundMovement(ctx context.Context, tx pgx.Tx, userID, txnID uuid.UUID) (lockedFundMovement, error) {
	var m lockedFundMovement

	err := tx.QueryRow(ctx, `
		SELECT pe.asset_id, pe.id, t.transaction_date
		FROM transactions t
		JOIN portfolio_entries pe ON pe.id = t.entry_id
		JOIN portfolios p         ON p.id = pe.portfolio_id
		JOIN assets a             ON a.id = pe.asset_id
		WHERE t.id = $1 AND p.user_id = $2 AND a.asset_type = 'fund' AND t.type IN ('buy', 'sell')
	`, txnID, userID).Scan(&m.assetID, &m.entryID, &m.date)
	if errors.Is(err, pgx.ErrNoRows) {
		return m, ErrFundMovementNotFound
	}

	if err != nil {
		return m, err
	}

	m.date = cashRateDay(m.date)
	m.tracking, err = lockFund(ctx, tx, userID, m.assetID)

	return m, err
}

// GetFundMovement reads one purchase or sale of a fund.
func (r *PostgresRepository) GetFundMovement(ctx context.Context, userID, txnID uuid.UUID) (FundMovement, error) {
	var movement FundMovement

	err := database.WithinTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		movement, err = readFundMovement(ctx, tx, userID, txnID)

		return err
	})

	return movement, err
}

// UpdateFundMovement restates a contribution or withdrawal — its day, its
// money, its fees, its note — and replays the fund from the earlier of its old
// and new days. In a fund followed by units it restates the units and unit
// value instead, and there is nothing to replay (updateUnitsMovement).
func (r *PostgresRepository) UpdateFundMovement(ctx context.Context, userID, txnID uuid.UUID, in FundMovementEdit) (FundMovement, error) {
	var movement FundMovement

	if err := database.WithinTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		locked, err := lockFundMovement(ctx, tx, userID, txnID)
		if err != nil {
			return err
		}

		if locked.tracking == FundUnits {
			movement, err = updateUnitsMovement(ctx, tx, userID, txnID, locked, in)

			return err
		}

		day := cashRateDay(in.Date)

		if _, err := tx.Exec(ctx, `
			UPDATE transactions
			   SET transaction_date = $2::date, fees = $3::numeric, notes = NULLIF($4, ''), updated_at = NOW()
			 WHERE id = $1
		`, txnID, day.Format(time.DateOnly), in.Fees.String(), in.Notes); err != nil {
			return err
		}

		if err := recordFundMovement(ctx, tx, txnID, in.Amount, in.All); err != nil {
			return err
		}

		from := day
		if locked.date.Before(from) {
			from = locked.date
		}

		if _, err := replayAndRestate(ctx, tx, userID, locked.assetID, from); err != nil {
			return err
		}

		movement, err = readFundMovement(ctx, tx, userID, txnID)

		return err
	}); err != nil {
		return FundMovement{}, err
	}

	return movement, nil
}

// DeleteFundMovement takes a contribution or withdrawal back, with the cash
// row it moved, and replays a fund followed by balance from its day. One
// followed by units has nothing to replay: its position is recalculated from
// what remains, as after any deleted transaction.
func (r *PostgresRepository) DeleteFundMovement(ctx context.Context, userID, txnID uuid.UUID) error {
	return database.WithinTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		locked, err := lockFundMovement(ctx, tx, userID, txnID)
		if err != nil {
			return err
		}

		if err := syncCashLink(ctx, tx, userID, txnID, false, cashPlacement{}); err != nil {
			return err
		}

		if _, err := tx.Exec(ctx, `DELETE FROM transactions WHERE id = $1`, txnID); err != nil {
			return err
		}

		if locked.tracking == FundUnits {
			return syncEntryCashLinks(ctx, tx, userID, locked.entryID, true)
		}

		_, err = replayAndRestate(ctx, tx, userID, locked.assetID, locked.date)

		return err
	})
}

// GetFundMovements lists the purchases and sales of a fund the user follows,
// in any of their portfolios, the most recent first. A fund followed by units
// has no stored money, so its amount is its units at their price.
func (r *PostgresRepository) GetFundMovements(ctx context.Context, userID, assetID uuid.UUID) ([]FundMovement, error) {
	var follows bool
	if err := r.db.QueryRow(ctx, `
		SELECT EXISTS (SELECT 1 FROM user_funds WHERE user_id = $1 AND asset_id = $2)
		    OR EXISTS (`+heldFundQuery+`)
	`, userID, assetID).Scan(&follows); err != nil {
		return nil, err
	}

	if !follows {
		return nil, ErrFundNotFound
	}

	rows, err := r.db.Query(ctx, fundMovementSelect+`
		WHERE p.user_id = $1 AND pe.asset_id = $2 AND t.type IN ('buy', 'sell')
		ORDER BY t.transaction_date DESC, t.created_at DESC, t.id
	`, userID, assetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	movements := make([]FundMovement, 0)

	for rows.Next() {
		m, err := scanFundMovement(rows)
		if err != nil {
			return nil, err
		}

		movements = append(movements, m)
	}

	return movements, rows.Err()
}

const fundMovementSelect = `
	SELECT t.id, pe.asset_id, pe.id, p.id, p.name, COALESCE(s.name, ''), t.type, t.transaction_date,
	       ROUND(COALESCE(fm.amount, t.quantity * t.price), 8)::text, t.fees::text,
	       COALESCE(fm.withdraws_all, FALSE), t.quantity::text, t.price::text,
	       COALESCE(t.notes, ''), t.created_at
	FROM transactions t
	JOIN portfolio_entries pe ON pe.id = t.entry_id
	JOIN portfolios p         ON p.id = pe.portfolio_id
	LEFT JOIN investment_sources s ON s.id = pe.source_id
	LEFT JOIN fund_movements fm    ON fm.txn_id = t.id`

func scanFundMovement(row pgx.Row) (FundMovement, error) {
	var (
		m       FundMovement
		txnType TransactionType
	)

	if err := row.Scan(&m.TxnID, &m.AssetID, &m.EntryID, &m.PortfolioID, &m.PortfolioName, &m.SourceName, &txnType, &m.Date,
		&m.Amount, &m.Fees, &m.All, &m.Units, &m.UnitValue, &m.Notes, &m.CreatedAt); err != nil {
		return m, err
	}

	m.Kind = FundContribution
	if txnType == Sell {
		m.Kind = FundWithdrawal
	}

	return m, nil
}

// readFundMovement reads one purchase or sale of a fund the user follows.
func readFundMovement(ctx context.Context, tx pgx.Tx, userID, txnID uuid.UUID) (FundMovement, error) {
	m, err := scanFundMovement(tx.QueryRow(ctx, fundMovementSelect+`
		JOIN assets a ON a.id = pe.asset_id
		WHERE t.id = $1 AND p.user_id = $2 AND a.asset_type = 'fund' AND t.type IN ('buy', 'sell')
	`, txnID, userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return m, ErrFundMovementNotFound
	}

	return m, err
}

// replayAfterPositionDeleted replays a fund followed by balance once one of its
// positions is gone: the balances valued the units of every portfolio, and one
// fewer changes their unit values. Every snapshot of the fund is revalued.
func replayAfterPositionDeleted(ctx context.Context, tx pgx.Tx, userID, assetID uuid.UUID) error {
	var balance bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS (SELECT 1 FROM user_funds WHERE user_id = $1 AND asset_id = $2 AND tracking = 'balance')
	`, userID, assetID).Scan(&balance); err != nil || !balance {
		return err
	}

	if _, err := lockFund(ctx, tx, userID, assetID); err != nil {
		return err
	}

	_, err := replayAndRestate(ctx, tx, userID, assetID, time.Time{})

	return err
}
