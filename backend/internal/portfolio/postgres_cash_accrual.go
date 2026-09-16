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

// What a balance held at the close of a day is cash_entry_balance_at_close
// (000045): every row dated on or before it, every interest the ledger credited
// for a day before it whatever day its transaction carries, and every day it
// computed and has not credited yet.

// cashAccountHeld is what every balance of an account — a platform, a currency
// and a pocket of it — held at the close of a day. Only a rate with tiers reads
// it: the tiers belong to the account, so its balances earn their share of what
// it earns.
//
// The pocket is part of the key. An account at 8 % and its pocket at 10 % each
// count what their own balances hold, so neither pushes the other up a tier.
const cashAccountHeld = `
	SELECT COALESCE(SUM(cash_entry_balance_at_close(pe.id, $2::date)), 0)
	FROM portfolio_entries pe
	JOIN assets a ON a.id = pe.asset_id
	WHERE pe.source_id = $1
	  AND pe.cost_currency = $3::char(3)
	  AND pe.pocket_id IS NOT DISTINCT FROM $4::uuid
	  AND a.asset_type = 'cash'
	  AND a.currency = pe.cost_currency
	  AND cash_entry_at_par(pe.id)`

// cashAccountScope narrows a statement to the balances of one account. Every
// part of the filter is optional, and an empty one leaves every balance in:
// that is what the nightly job runs over.
//
// The pocket takes two parameters rather than one, because "no pocket" is an
// answer and not an omission: $5 says whether a pocket was named at all, and
// $6 which one, NULL being the main account.
const cashAccountScope = `
	AND ($2::uuid IS NULL     OR p.user_id = $2::uuid)
	AND ($3::uuid IS NULL     OR pe.source_id = $3::uuid)
	AND ($4::char(3) IS NULL  OR pe.cost_currency = $4::char(3))
	AND (NOT $5::boolean      OR pe.pocket_id IS NOT DISTINCT FROM $6::uuid)`

// scopeArgs is the filter as the five parameters cashAccountScope reads.
func scopeArgs(filter CashAccrualFilter) (userID, sourceID *uuid.UUID, cur *string, scoped bool, pocketID *uuid.UUID) {
	if filter.UserID != (uuid.UUID{}) {
		userID = &filter.UserID
	}

	if filter.SourceID != (uuid.UUID{}) {
		sourceID = &filter.SourceID
	}

	if code := currencyParam(filter.Currency); code != "" {
		cur = &code
	}

	if filter.Pocket != nil {
		scoped = true

		if *filter.Pocket != (uuid.UUID{}) {
			pocketID = filter.Pocket
		}
	}

	return userID, sourceID, cur, scoped, pocketID
}

// cashRateSteps reads a version's tiers, lowest first, as the ledger computes
// with them.
func cashRateSteps(ctx context.Context, tx pgx.Tx, rateID uuid.UUID) ([]CashRateStep, error) {
	rows, err := tx.Query(ctx, `
		SELECT from_balance, annual_rate FROM cash_yield_rate_tiers
		WHERE rate_id = $1
		ORDER BY from_balance
	`, rateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	steps := make([]CashRateStep, 0)
	for rows.Next() {
		var step CashRateStep
		if err := rows.Scan(&step.From, &step.AnnualRate); err != nil {
			return nil, err
		}

		steps = append(steps, step)
	}

	return steps, rows.Err()
}

// GetCashAccrualTargets lists every cash balance whose account has a rate that
// started by through, with where its ledger stands and those versions of the
// rate. The filter narrows it to one account; an empty one takes them all.
//
// Only the balances the cash writers keep earn: in their own currency, at one
// unit per unit. A dollar position bought with pesos is not a savings account.
func (r *PostgresRepository) GetCashAccrualTargets(ctx context.Context, through time.Time, filter CashAccrualFilter) ([]CashAccrualTarget, error) {
	userID, sourceID, cur, scoped, pocketID := scopeArgs(filter)

	rows, err := r.db.Query(ctx, `
		SELECT
			pe.id,
			(pe.created_at AT TIME ZONE 'UTC')::date,
			(SELECT MAX(ac.accrual_date) FROM cash_interest_accruals ac WHERE ac.entry_id = pe.id),
			r.id,
			r.annual_rate,
			r.withholding_rate,
			r.posting::text,
			r.effective_from,
			r.ended_on
		FROM portfolio_entries pe
		JOIN portfolios p       ON p.id = pe.portfolio_id
		JOIN assets a           ON a.id = pe.asset_id
		JOIN cash_yield_rates r ON r.source_id = pe.source_id
		                       AND r.currency  = pe.cost_currency
		                       AND r.pocket_id IS NOT DISTINCT FROM pe.pocket_id
		WHERE a.asset_type = 'cash'
		  AND a.currency = pe.cost_currency
		  AND r.effective_from <= $1::date
		  AND cash_entry_at_par(pe.id)`+cashAccountScope+`
		ORDER BY pe.id, r.effective_from
	`, cashRateDay(through).Format(time.DateOnly), userID, sourceID, cur, scoped, pocketID)
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
			&version.Posting,
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
			sourceID    uuid.UUID
			cur         money.Currency
			// The drawer the balance sits in, NULL for the main account. It is
			// part of the account's key, so the version of the rate and the sum
			// the tiers read are both looked up with it.
			pocketID *uuid.UUID
		)

		// The lock a withdrawal takes: the basis read below is what the balance
		// holds while this runs.
		err := tx.QueryRow(ctx, `
			SELECT portfolio_id, source_id, cost_currency, pocket_id FROM portfolio_entries WHERE id = $1 FOR UPDATE
		`, entryID).Scan(&portfolioID, &sourceID, &cur, &pocketID)
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
			SELECT r.annual_rate, r.withholding_rate, r.posting::text, r.effective_from, r.ended_on
			FROM cash_yield_rates r
			JOIN portfolio_entries pe ON pe.source_id = r.source_id
			                         AND pe.cost_currency = r.currency
			                         AND pe.pocket_id IS NOT DISTINCT FROM r.pocket_id
			WHERE r.id = $1
			  AND pe.id = $2
			  AND NOT EXISTS (
			    SELECT 1 FROM cash_yield_rates later
			    WHERE later.source_id = r.source_id
			      AND later.currency  = r.currency
			      AND later.pocket_id IS NOT DISTINCT FROM r.pocket_id
			      AND later.effective_from > r.effective_from
			      AND later.effective_from <= $3::date
			  )
			FOR SHARE OF r
		`, rateID, entryID, date).Scan(&version.AnnualRate, &version.WithholdingRate, &version.Posting,
			&version.EffectiveFrom, &version.EndedOn)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}

		if !version.coversDay(cashRateDay(day)) {
			return nil
		}

		if version.Tiers, err = cashRateSteps(ctx, tx, rateID); err != nil {
			return err
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

		// What the days the balance holds earned. A rate posted monthly leaves
		// them waiting in the ledger, and the day that closes the month pays
		// them all in one credit.
		held := decimal.Zero
		if err := tx.QueryRow(ctx, `
			SELECT COALESCE(SUM(net_amount), 0) FROM cash_interest_accruals
			WHERE entry_id = $1 AND accrual_date < $2::date AND status = 'pending'
		`, entryID, date).Scan(&held); err != nil {
			return err
		}

		atClose := CashDay{Carry: carry, Held: held, Credits: version.creditsOn(cashRateDay(day))}
		if err := tx.QueryRow(ctx, `
			SELECT cash_entry_balance_at_close($1, $2::date)
		`, entryID, date).Scan(&atClose.Basis); err != nil {
			return err
		}

		atClose.AccountHeld = atClose.Basis
		if len(version.Tiers) > 0 {
			if err := tx.QueryRow(ctx, cashAccountHeld, sourceID, date, cur, pocketID).Scan(&atClose.AccountHeld); err != nil {
				return err
			}
		}

		accrual, err := accrueDay(atClose, version, cur)
		if err != nil {
			return err
		}

		status := "pending"
		switch {
		case !atClose.Credits:
		case accrual.Credited.IsPos():
			status = "posted"
		default:
			status = "carried"
		}

		var accrualID uuid.UUID
		err = tx.QueryRow(ctx, `
			INSERT INTO cash_interest_accruals
				(entry_id, rate_id, accrual_date, annual_rate, balance_basis, gross_amount, withholding, net_amount, rounding_carry, status)
			VALUES ($1, $2, $3::date, $4::numeric, $5::numeric, $6::numeric, $7::numeric, $8::numeric, $9::numeric, $10)
			ON CONFLICT (entry_id, accrual_date) DO NOTHING
			RETURNING id
		`, entryID, rateID, date, accrual.AnnualRate.String(), accrual.Basis.String(), accrual.Gross.String(),
			accrual.Withholding.String(), accrual.Net.String(), accrual.Carry.String(), status).Scan(&accrualID)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}

		if !atClose.Credits {
			return nil
		}

		// The rounded credit came to nothing, so the days it would have paid
		// stop waiting: their interest is in the carry now, and the next credit
		// pays it.
		if !accrual.Credited.IsPos() {
			_, err := tx.Exec(ctx, `
				UPDATE cash_interest_accruals SET status = 'carried'
				WHERE entry_id = $1 AND accrual_date < $2::date AND status = 'pending'
			`, entryID, date)

			return err
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

		// The credit is linked to every day it pays, so what a balance held
		// reads the same however many days one transaction covers.
		if _, err := tx.Exec(ctx, `
			UPDATE cash_interest_accruals SET transaction_id = $2, status = 'posted'
			WHERE id = $1
			   OR (entry_id = $3 AND accrual_date < $4::date AND status = 'pending')
		`, accrualID, txnID, entryID, date); err != nil {
			return err
		}

		credited = true

		return nil
	})

	return credited, err
}

// GetHeldCashInterest lists the balances holding days no month end will ever
// close. The filter narrows it to one account, as it does the targets; an empty
// one takes them all.
//
// A rate posted monthly leaves each day waiting until the last day of its
// month. A balance whose rate is paused, ended or replaced partway through a
// month never reaches that day, and what it earned would wait in the ledger
// forever. These are those balances: some day is still held, no later day was
// computed, and no version of the account's rate covers a day between the last
// one held and the end of its month.
func (r *PostgresRepository) GetHeldCashInterest(ctx context.Context, through time.Time, filter CashAccrualFilter) ([]uuid.UUID, error) {
	userID, sourceID, cur, scoped, pocketID := scopeArgs(filter)

	rows, err := r.db.Query(ctx, `
		WITH held AS (
			SELECT ac.entry_id, pe.source_id, pe.cost_currency, pe.pocket_id, MAX(ac.accrual_date) AS last_day
			FROM cash_interest_accruals ac
			JOIN portfolio_entries pe ON pe.id = ac.entry_id
			JOIN portfolios p         ON p.id = pe.portfolio_id
			WHERE ac.status = 'pending'`+cashAccountScope+`
			GROUP BY ac.entry_id, pe.source_id, pe.cost_currency, pe.pocket_id
		)
		SELECT entry_id FROM held
		WHERE last_day < $1::date
		  AND NOT EXISTS (
		    SELECT 1 FROM cash_yield_rates r
		    WHERE r.source_id = held.source_id
		      AND r.currency  = held.cost_currency
		      AND r.pocket_id IS NOT DISTINCT FROM held.pocket_id
		      AND r.effective_from <= (date_trunc('month', held.last_day) + INTERVAL '1 month - 1 day')::date
		      AND (r.ended_on IS NULL OR r.ended_on > held.last_day)
		  )
		ORDER BY entry_id
	`, cashRateDay(through).Format(time.DateOnly), userID, sourceID, cur, scoped, pocketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := make([]uuid.UUID, 0)
	for rows.Next() {
		var entryID uuid.UUID
		if err := rows.Scan(&entryID); err != nil {
			return nil, err
		}

		entries = append(entries, entryID)
	}

	return entries, rows.Err()
}

// PostHeldCashInterest credits what a balance holds, on the last day it held.
// It reports whether anything was credited: what rounds to nothing is carried
// instead, as it is on any other credit.
//
// It takes the same lock and dates the credit the same way AccrueCashInterestDay
// does, and it is idempotent for the same reason: a balance whose days it
// resolves holds none afterwards, so a second run finds nothing to credit.
func (r *PostgresRepository) PostHeldCashInterest(ctx context.Context, entryID uuid.UUID) (bool, error) {
	credited := false

	err := database.WithinTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		var (
			portfolioID uuid.UUID
			cur         money.Currency
		)

		err := tx.QueryRow(ctx, `
			SELECT portfolio_id, cost_currency FROM portfolio_entries WHERE id = $1 FOR UPDATE
		`, entryID).Scan(&portfolioID, &cur)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}

		var (
			held    decimal.Decimal
			lastDay *time.Time
		)
		if err := tx.QueryRow(ctx, `
			SELECT COALESCE(SUM(net_amount), 0), MAX(accrual_date)
			FROM cash_interest_accruals
			WHERE entry_id = $1 AND status = 'pending'
		`, entryID).Scan(&held, &lastDay); err != nil {
			return err
		}

		// Another run credited them while this one waited for the lock.
		if lastDay == nil {
			return nil
		}

		date := cashRateDay(*lastDay).Format(time.DateOnly)

		// A held day carries what it inherited, so the last one carries what the
		// last credit left.
		var carry decimal.Decimal
		if err := tx.QueryRow(ctx, `
			SELECT rounding_carry FROM cash_interest_accruals
			WHERE entry_id = $1 AND accrual_date = $2::date
		`, entryID, date).Scan(&carry); err != nil {
			return err
		}

		amount, left, err := creditHeld(held, carry, cur)
		if err != nil {
			return err
		}

		status, txnID := "carried", (*uuid.UUID)(nil)
		if amount.IsPos() {
			var id uuid.UUID
			if err := tx.QueryRow(ctx, `
				INSERT INTO transactions (entry_id, type, quantity, price, currency, fx_rate, fees, fees_currency, transaction_date, notes)
				SELECT $1::uuid, 'cash_interest', $2::numeric, 1, $3::char(3), 1, 0, $3::char(3),
				       GREATEST($4::date, COALESCE(MAX(ps.snapshot_date), $4::date)), ''
				FROM portfolio_snapshots ps
				WHERE ps.portfolio_id = $5
				RETURNING id
			`, entryID, amount.String(), cur, date, portfolioID).Scan(&id); err != nil {
				return err
			}

			status, txnID = "posted", &id
		}

		if _, err := tx.Exec(ctx, `
			UPDATE cash_interest_accruals SET
				status         = $2,
				transaction_id = $3,
				rounding_carry = CASE WHEN accrual_date = $4::date THEN $5::numeric ELSE rounding_carry END
			WHERE entry_id = $1 AND status = 'pending'
		`, entryID, status, txnID, date, left.String()); err != nil {
			return err
		}

		credited = amount.IsPos()

		return nil
	})

	return credited, err
}

// ClearCashInterest throws away the days an account's ledger computed from a
// day, and the credits that paid them, so they can be computed again.
//
// A day is never revisited on its own: it earned on what the balance held at
// its close, and a deposit recorded afterwards with a past date changes what
// that was. Clearing is how the owner asks for the days after it to be redone.
//
// The window widens backwards to the first day of any credit that reaches
// across the day asked for: a month's credit pays every day of its month, and
// half of one cannot be undone. Deleting a cash_interest leaves the growth
// series alone — only what moves a quantity is retired (000038) — so what the
// recalculation writes instead lands as return, exactly as the first run did.
//
// A filter that names a platform and an owner is checked against
// investment_sources first, so someone else's account answers not found rather
// than reporting that it cleared nothing — the same answer every other write to
// a rate gives. The check is inside the transaction, like CreateCashRate's, so
// there is no window between asking and clearing.
func (r *PostgresRepository) ClearCashInterest(ctx context.Context, filter CashAccrualFilter, from time.Time) (CashInterestCleared, error) {
	cleared := CashInterestCleared{From: cashRateDay(from)}
	asked := cleared.From.Format(time.DateOnly)
	userID, sourceID, cur, scoped, pocketID := scopeArgs(filter)

	err := database.WithinTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		// Whether the platform is still active is not asked: one that stopped
		// taking money can still have its past corrected.
		if filter.UserID != (uuid.UUID{}) && filter.SourceID != (uuid.UUID{}) {
			var owned bool
			if err := tx.QueryRow(ctx, `
				SELECT EXISTS (SELECT 1 FROM investment_sources WHERE id = $1 AND user_id = $2)
			`, filter.SourceID, filter.UserID).Scan(&owned); err != nil {
				return err
			}

			if !owned {
				return ErrPlatformNotFound
			}
		}

		// Only the balances with a day to throw away, under the lock a credit
		// takes and in a fixed order, so two runs queue rather than deadlock.
		rows, err := tx.Query(ctx, `
			SELECT pe.id
			FROM portfolio_entries pe
			JOIN portfolios p ON p.id = pe.portfolio_id
			JOIN assets a     ON a.id = pe.asset_id
			WHERE a.asset_type = 'cash'
			  AND a.currency = pe.cost_currency
			  AND EXISTS (
			    SELECT 1 FROM cash_interest_accruals ac
			    WHERE ac.entry_id = pe.id AND ac.accrual_date >= $1::date
			  )`+cashAccountScope+`
			ORDER BY pe.id
			FOR UPDATE OF pe
		`, asked, userID, sourceID, cur, scoped, pocketID)
		if err != nil {
			return err
		}

		entries := make([]uuid.UUID, 0)
		for rows.Next() {
			var entryID uuid.UUID
			if err := rows.Scan(&entryID); err != nil {
				rows.Close()

				return err
			}

			entries = append(entries, entryID)
		}
		rows.Close()

		if err := rows.Err(); err != nil {
			return err
		}

		for _, entryID := range entries {
			var first time.Time
			if err := tx.QueryRow(ctx, `
				SELECT LEAST($2::date, COALESCE(MIN(ac.accrual_date), $2::date))
				FROM cash_interest_accruals ac
				WHERE ac.entry_id = $1
				  AND ac.transaction_id IN (
				    SELECT paid.transaction_id FROM cash_interest_accruals paid
				    WHERE paid.entry_id = $1
				      AND paid.accrual_date >= $2::date
				      AND paid.transaction_id IS NOT NULL
				  )
			`, entryID, asked).Scan(&first); err != nil {
				return err
			}

			day := first.Format(time.DateOnly)

			if _, err := tx.Exec(ctx, `
				DELETE FROM transactions WHERE id IN (
					SELECT DISTINCT ac.transaction_id FROM cash_interest_accruals ac
					WHERE ac.entry_id = $1 AND ac.accrual_date >= $2::date AND ac.transaction_id IS NOT NULL
				)
			`, entryID, day); err != nil {
				return err
			}

			tag, err := tx.Exec(ctx, `
				DELETE FROM cash_interest_accruals WHERE entry_id = $1 AND accrual_date >= $2::date
			`, entryID, day)
			if err != nil {
				return err
			}

			if days := int(tag.RowsAffected()); days > 0 {
				cleared.Balances++
				cleared.Days += days

				if first.Before(cleared.From) {
					cleared.From = first
				}
			}
		}

		return nil
	})
	if err != nil {
		return CashInterestCleared{}, err
	}

	return cleared, nil
}

// addCashInterest fills in what each balance has earned: the interest credited
// to it — by the ledger or by hand — ever and in the current UTC month, that
// month's part in the display currency, what the ledger computed and has not
// credited yet, and the last day it computed.
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
		balances[i].PendingInterest = "0"
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
			ROUND(COALESCE((
				SELECT SUM(ac.net_amount) FROM cash_interest_accruals ac
				WHERE ac.entry_id = pe.id AND ac.status = 'pending'
			), 0), 8)::text,
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
			id                                 uuid.UUID
			earned, month, monthValue, pending string
			lastAccrual                        *time.Time
		)

		if err := rows.Scan(&id, &earned, &month, &monthValue, &pending, &lastAccrual); err != nil {
			return err
		}

		if i, ok := index[id]; ok {
			balances[i].InterestEarned = earned
			balances[i].InterestThisMonth = month
			balances[i].InterestThisMonthValue = monthValue
			balances[i].PendingInterest = pending
			balances[i].LastAccrualDate = lastAccrual
		}
	}

	return rows.Err()
}
