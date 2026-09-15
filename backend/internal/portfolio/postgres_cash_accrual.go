package portfolio

import (
	"context"
	"errors"
	"time"

	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/platform/database"
)

// cashBalanceAtClose is what a balance held at the close of a day: every row
// dated on or before it, and every interest the ledger credited for a day
// before it, whatever day its transaction carries. A catch-up dates those
// later (see AccrueCashInterestDay), and reading them by that date would leave
// each caught-up day compounding on less than the balance held.
const cashBalanceAtClose = `
	SELECT COALESCE(SUM(CASE
		WHEN t.type IN ('buy', 'transfer_in', 'cash_interest') THEN t.quantity
		WHEN t.type IN ('sell', 'transfer_out')                THEN -t.quantity
		ELSE 0 END), 0)
	FROM transactions t
	LEFT JOIN LATERAL (
		SELECT MIN(ac.accrual_date) AS accrual_date
		FROM cash_interest_accruals ac
		WHERE ac.transaction_id = t.id
	) credited ON TRUE
	WHERE t.entry_id = $1
	  AND COALESCE(credited.accrual_date + 1, t.transaction_date) <= $2::date`

// GetCashAccrualTargets lists every cash balance whose account has a rate that
// started by through, with where its ledger stands and those versions of the
// rate.
//
// Only the balances the cash writers keep earn: in their own currency, at one
// unit per unit. A dollar position bought with pesos is not a savings account.
func (r *PostgresRepository) GetCashAccrualTargets(ctx context.Context, through time.Time) ([]CashAccrualTarget, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			pe.id,
			(pe.created_at AT TIME ZONE 'UTC')::date,
			(SELECT MAX(ac.accrual_date) FROM cash_interest_accruals ac WHERE ac.entry_id = pe.id),
			r.id,
			r.annual_rate,
			r.withholding_rate,
			r.effective_from,
			r.ended_on
		FROM portfolio_entries pe
		JOIN assets a           ON a.id = pe.asset_id
		JOIN cash_yield_rates r ON r.source_id = pe.source_id AND r.currency = pe.cost_currency
		WHERE a.asset_type = 'cash'
		  AND a.currency = pe.cost_currency
		  AND r.effective_from <= $1::date
		  AND cash_entry_at_par(pe.id)
		ORDER BY pe.id, r.effective_from
	`, cashRateDay(through).Format(time.DateOnly))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	targets := make([]CashAccrualTarget, 0)
	for rows.Next() {
		var (
			entryID     uuid.UUID
			openedOn    time.Time
			lastAccrued *time.Time
			version     CashRateVersion
		)

		if err := rows.Scan(
			&entryID,
			&openedOn,
			&lastAccrued,
			&version.ID,
			&version.AnnualRate,
			&version.WithholdingRate,
			&version.EffectiveFrom,
			&version.EndedOn,
		); err != nil {
			return nil, err
		}

		if n := len(targets); n == 0 || targets[n-1].EntryID != entryID {
			targets = append(targets, CashAccrualTarget{EntryID: entryID, OpenedOn: openedOn, LastAccrued: lastAccrued})
		}

		target := &targets[len(targets)-1]
		target.Versions = append(target.Versions, version)
	}

	return targets, rows.Err()
}

// AccrueCashInterestDay computes one day of interest on one balance, at one
// version of its rate, and credits it. It reports whether anything was
// credited: a day whose rounded interest is zero is computed and carried.
//
// It all happens in one transaction, in the order that keeps a day from being
// credited twice. The ledger row is written first, and a day already taken — by
// an earlier run or a concurrent one — writes nothing at all.
//
// The credit is dated on the day it was earned unless a snapshot of the
// portfolio has already closed that day; then it is dated on the last snapshot.
// The growth series reads a transaction dated before the stretch it lands in as
// loaded history — money that walked in, not return — and interest caught up
// after the job was down is return all the same.
func (r *PostgresRepository) AccrueCashInterestDay(ctx context.Context, entryID, rateID uuid.UUID, day time.Time) (bool, error) {
	credited := false
	date := cashRateDay(day).Format(time.DateOnly)

	err := database.WithinTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		var (
			portfolioID uuid.UUID
			cur         money.Currency
		)

		// The lock a withdrawal takes: the basis read below is what the balance
		// holds while this runs.
		err := tx.QueryRow(ctx, `
			SELECT portfolio_id, cost_currency FROM portfolio_entries WHERE id = $1 FOR UPDATE
		`, entryID).Scan(&portfolioID, &cur)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}

		// The version is read here rather than trusted from the listing, and it
		// must still be the account's rate for the day. The shared lock holds a
		// correction back until the day is written, and the correction then finds
		// the day computed.
		version := CashRateVersion{ID: rateID}
		err = tx.QueryRow(ctx, `
			SELECT r.annual_rate, r.withholding_rate, r.effective_from, r.ended_on
			FROM cash_yield_rates r
			JOIN portfolio_entries pe ON pe.source_id = r.source_id AND pe.cost_currency = r.currency
			WHERE r.id = $1
			  AND pe.id = $2
			  AND NOT EXISTS (
			    SELECT 1 FROM cash_yield_rates later
			    WHERE later.source_id = r.source_id
			      AND later.currency  = r.currency
			      AND later.effective_from > r.effective_from
			      AND later.effective_from <= $3::date
			  )
			FOR SHARE OF r
		`, rateID, entryID, date).Scan(&version.AnnualRate, &version.WithholdingRate, &version.EffectiveFrom, &version.EndedOn)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}

		if !version.coversDay(cashRateDay(day)) {
			return nil
		}

		carry := decimal.Zero
		err = tx.QueryRow(ctx, `
			SELECT rounding_carry FROM cash_interest_accruals
			WHERE entry_id = $1 AND accrual_date < $2::date
			ORDER BY accrual_date DESC
			LIMIT 1
		`, entryID, date).Scan(&carry)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return err
		}

		var basis decimal.Decimal
		if err := tx.QueryRow(ctx, cashBalanceAtClose, entryID, date).Scan(&basis); err != nil {
			return err
		}

		accrual, err := accrueDay(basis, version, carry, cur)
		if err != nil {
			return err
		}

		status := "carried"
		if accrual.Credited.IsPos() {
			status = "posted"
		}

		var accrualID uuid.UUID
		err = tx.QueryRow(ctx, `
			INSERT INTO cash_interest_accruals
				(entry_id, rate_id, accrual_date, annual_rate, balance_basis, gross_amount, withholding, net_amount, rounding_carry, status)
			VALUES ($1, $2, $3::date, $4::numeric, $5::numeric, $6::numeric, $7::numeric, $8::numeric, $9::numeric, $10)
			ON CONFLICT (entry_id, accrual_date) DO NOTHING
			RETURNING id
		`, entryID, rateID, date, version.AnnualRate.String(), accrual.Basis.String(), accrual.Gross.String(),
			accrual.Withholding.String(), accrual.Net.String(), accrual.Carry.String(), status).Scan(&accrualID)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}

		if !accrual.Credited.IsPos() {
			return nil
		}

		var txnID uuid.UUID
		if err := tx.QueryRow(ctx, `
			INSERT INTO transactions (entry_id, type, quantity, price, currency, fx_rate, fees, fees_currency, transaction_date, notes)
			SELECT $1::uuid, 'cash_interest', $2::numeric, 1, $3::char(3), 1, 0, $3::char(3),
			       GREATEST($4::date, COALESCE(MAX(ps.snapshot_date), $4::date)), ''
			FROM portfolio_snapshots ps
			WHERE ps.portfolio_id = $5
			RETURNING id
		`, entryID, accrual.Credited.String(), cur, date, portfolioID).Scan(&txnID); err != nil {
			return err
		}

		if _, err := tx.Exec(ctx, `
			UPDATE cash_interest_accruals SET transaction_id = $2 WHERE id = $1
		`, accrualID, txnID); err != nil {
			return err
		}

		credited = true

		return nil
	})

	return credited, err
}

// addCashInterest fills in what each balance has earned: the interest credited
// to it — by the ledger or by hand — ever and in the current UTC month, that
// month's part in the display currency, and the last day the ledger computed.
func (r *PostgresRepository) addCashInterest(ctx context.Context, userID uuid.UUID, displayCurrency money.Currency, balances []CashBalance) error {
	if len(balances) == 0 {
		return nil
	}

	ids := make([]uuid.UUID, len(balances))
	index := make(map[uuid.UUID]int, len(balances))
	for i := range balances {
		ids[i] = balances[i].EntryID
		index[balances[i].EntryID] = i
		balances[i].InterestEarned = "0"
		balances[i].InterestThisMonth = "0"
		balances[i].InterestThisMonthValue = "0"
	}

	rows, err := r.db.Query(ctx, `
		SELECT
			pe.id,
			ROUND(COALESCE(SUM(t.quantity) FILTER (WHERE t.type = 'cash_interest'), 0), 8)::text,
			ROUND(COALESCE(SUM(t.quantity) FILTER (
				WHERE t.type = 'cash_interest' AND t.transaction_date >= period.starts_on
			), 0), 8)::text,
			ROUND(COALESCE(SUM(t.quantity) FILTER (
				WHERE t.type = 'cash_interest' AND t.transaction_date >= period.starts_on
			), 0) * COALESCE(fx.rate, 1), 8)::text,
			(SELECT MAX(ac.accrual_date) FROM cash_interest_accruals ac WHERE ac.entry_id = pe.id)
		FROM portfolio_entries pe
		JOIN portfolios p ON p.id = pe.portfolio_id
		JOIN users u      ON u.id = p.user_id
		JOIN assets a     ON a.id = pe.asset_id
		LEFT JOIN transactions t ON t.entry_id = pe.id
		CROSS JOIN LATERAL (
			SELECT date_trunc('month', NOW() AT TIME ZONE 'UTC')::date AS starts_on
		) period
		CROSS JOIN LATERAL (
			SELECT COALESCE(NULLIF($3::text, ''), u.preferred_currency, 'USD')::char(3) AS code
		) target
		CROSS JOIN LATERAL (
			SELECT fx_rate(p.user_id, COALESCE(a.currency, pe.cost_currency), target.code) AS rate
		) fx
		WHERE p.user_id = $1
		  AND pe.id = ANY($2)
		GROUP BY pe.id, fx.rate
	`, userID, ids, currencyParam(displayCurrency))
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id                        uuid.UUID
			earned, month, monthValue string
			lastAccrual               *time.Time
		)

		if err := rows.Scan(&id, &earned, &month, &monthValue, &lastAccrual); err != nil {
			return err
		}

		if i, ok := index[id]; ok {
			balances[i].InterestEarned = earned
			balances[i].InterestThisMonth = month
			balances[i].InterestThisMonthValue = monthValue
			balances[i].LastAccrualDate = lastAccrual
		}
	}

	return rows.Err()
}
