package portfolio

import (
	"context"
	"errors"
	"fmt"
	"time"

	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/platform/database"
)

// Every date these writes send travels as its calendar day in text. A
// time.Time is encoded as a timestamp, and casting that to a date happens in
// the session's time zone, which is not guaranteed to be UTC.

// cashRateLatest is the SQL test behind CashRate.Latest: no version of the same
// account starts later.
const cashRateLatest = `NOT EXISTS (
	SELECT 1 FROM cash_yield_rates later
	WHERE later.source_id = r.source_id
	  AND later.currency  = r.currency
	  AND later.effective_from > r.effective_from
)`

// cashRateAccruedThrough is the SQL behind CashRate.AccruedThrough.
const cashRateAccruedThrough = `(
	SELECT MAX(ac.accrual_date) FROM cash_interest_accruals ac WHERE ac.rate_id = r.id
)`

// The percentages come back trimmed: "9.25", not "9.250000".
const cashRateColumns = `
	r.id, r.source_id, s.name, r.currency,
	trim_scale(r.annual_rate * 100)::text,
	trim_scale(r.withholding_rate * 100)::text,
	r.posting::text, trim_scale(r.max_balance)::text, r.effective_from, r.ended_on,
	` + cashRateLatest + `,
	` + cashRateAccruedThrough + `,
	r.created_at, r.updated_at`

const cashRateFrom = `
	FROM cash_yield_rates r
	JOIN investment_sources s ON s.id = r.source_id`

// maxBalanceParam is the cap as a numeric parameter: its digits, or NULL when
// the rate has none.
func maxBalanceParam(limit *decimal.Decimal) *string {
	if limit == nil {
		return nil
	}

	digits := limit.String()

	return &digits
}

func scanCashRate(row pgx.Row) (CashRate, error) {
	var rate CashRate

	err := row.Scan(
		&rate.ID,
		&rate.SourceID,
		&rate.SourceName,
		&rate.Currency,
		&rate.AnnualRatePct,
		&rate.WithholdingPct,
		&rate.Posting,
		&rate.MaxBalance,
		&rate.EffectiveFrom,
		&rate.EndedOn,
		&rate.Latest,
		&rate.AccruedThrough,
		&rate.CreatedAt,
		&rate.UpdatedAt,
	)

	return rate, err
}

// GetCashRatesByUserID lists every version of every rate on the user's
// platforms: by platform and currency, the newest version first.
func (r *PostgresRepository) GetCashRatesByUserID(ctx context.Context, userID uuid.UUID) ([]CashRate, error) {
	rows, err := r.db.Query(ctx, `SELECT `+cashRateColumns+cashRateFrom+`
		WHERE s.user_id = $1
		ORDER BY s.name, r.currency, r.effective_from DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rates := make([]CashRate, 0)
	for rows.Next() {
		rate, err := scanCashRate(rows)
		if err != nil {
			return nil, err
		}

		rates = append(rates, rate)
	}

	return rates, rows.Err()
}

func getCashRate(ctx context.Context, tx pgx.Tx, userID, rateID uuid.UUID) (CashRate, error) {
	rate, err := scanCashRate(tx.QueryRow(ctx, `SELECT `+cashRateColumns+cashRateFrom+`
		WHERE r.id = $1 AND s.user_id = $2
	`, rateID, userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return CashRate{}, ErrCashRateNotFound
	}

	return rate, err
}

// CreateCashRate records a rate, or a new version of one, from its first day.
//
// The version running on that day, if any, ends the day before, so no day has
// two rates. One that starts on that day or later is refused instead: slotting
// a version in before it would rewrite what it said. So is a first day whose
// interest was already computed at the rate before: that day keeps the rate it
// was earned at.
//
// The platform row is locked first. Every write to a rate takes that lock, so
// versions of one account are written one at a time — including the first,
// which has no version yet to lock.
func (r *PostgresRepository) CreateCashRate(ctx context.Context, userID uuid.UUID, in NewCashRateInput) (CashRate, error) {
	var rate CashRate
	start := cashRateDay(in.EffectiveFrom)

	if err := database.WithinTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		var active bool
		err := tx.QueryRow(ctx, `
			SELECT is_active FROM investment_sources WHERE id = $1 AND user_id = $2 FOR UPDATE
		`, in.SourceID, userID).Scan(&active)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrPlatformNotFound
		}
		if err != nil {
			return err
		}

		// An inactive platform takes no new money, so it is given no new rate.
		if !active {
			return invalidCashRate("the platform is inactive; activate it before giving it a rate")
		}

		var (
			latestID   uuid.UUID
			latestFrom time.Time
			latestEnd  *time.Time
		)
		err = tx.QueryRow(ctx, `
			SELECT id, effective_from, ended_on
			FROM cash_yield_rates
			WHERE source_id = $1 AND currency = $2::char(3)
			ORDER BY effective_from DESC
			LIMIT 1
		`, in.SourceID, in.Currency).Scan(&latestID, &latestFrom, &latestEnd)

		switch {
		case errors.Is(err, pgx.ErrNoRows):
		case err != nil:
			return err
		case !latestFrom.Before(start):
			return fmt.Errorf("%w: the latest starts on %s", ErrCashRateOverlaps, latestFrom.Format(time.DateOnly))
		default:
			var accruedThrough *time.Time
			if err := tx.QueryRow(ctx, `
				SELECT MAX(ac.accrual_date)
				FROM cash_interest_accruals ac
				JOIN cash_yield_rates v ON v.id = ac.rate_id
				WHERE v.source_id = $1 AND v.currency = $2::char(3)
			`, in.SourceID, in.Currency).Scan(&accruedThrough); err != nil {
				return err
			}

			if accruedThrough != nil && !accruedThrough.Before(start) {
				return cashRateInUse(*accruedThrough, "the new version can start from "+accruedThrough.AddDate(0, 0, 1).Format(time.DateOnly))
			}

			if latestEnd == nil || !latestEnd.Before(start) {
				if _, err := tx.Exec(ctx, `
					UPDATE cash_yield_rates SET ended_on = $2::date - 1 WHERE id = $1
				`, latestID, start.Format(time.DateOnly)); err != nil {
					return err
				}
			}
		}

		var rateID uuid.UUID
		if err := tx.QueryRow(ctx, `
			INSERT INTO cash_yield_rates (source_id, currency, annual_rate, withholding_rate, posting, max_balance, effective_from)
			VALUES ($1::uuid, $2::char(3), $3::numeric / 100, $4::numeric / 100, $5::cash_interest_posting, $6::numeric, $7::date)
			RETURNING id
		`, in.SourceID, in.Currency, in.AnnualRatePct.String(), in.WithholdingPct.String(), string(in.Posting),
			maxBalanceParam(in.MaxBalance), start.Format(time.DateOnly)).Scan(&rateID); err != nil {
			// The lock makes this unreachable from here; a writer that skipped it
			// would still get the answer the check above gives.
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return ErrCashRateOverlaps
			}

			return err
		}

		rate, err = getCashRate(ctx, tx, userID, rateID)

		return err
	}); err != nil {
		return CashRate{}, err
	}

	return rate, nil
}

// lockedCashRate is what a write to an existing version has to know first.
type lockedCashRate struct {
	sourceID       uuid.UUID
	currency       money.Currency
	effectiveFrom  time.Time
	accruedThrough *time.Time
}

// lockLatestCashRate finds a version the user owns, takes the lock
// CreateCashRate takes, and refuses one a later version follows.
//
// Whether it is the latest, and how far its interest was computed, are read
// after the lock, in a statement of their own: a version created or a day
// computed while this one waited is only visible to a statement that starts
// afterwards.
func lockLatestCashRate(ctx context.Context, tx pgx.Tx, userID, rateID uuid.UUID) (lockedCashRate, error) {
	var locked lockedCashRate

	err := tx.QueryRow(ctx, `
		SELECT r.source_id
		FROM cash_yield_rates r
		JOIN investment_sources s ON s.id = r.source_id
		WHERE r.id = $1 AND s.user_id = $2
		FOR UPDATE OF s
	`, rateID, userID).Scan(&locked.sourceID)
	if errors.Is(err, pgx.ErrNoRows) {
		return locked, ErrCashRateNotFound
	}
	if err != nil {
		return locked, err
	}

	var latest bool
	err = tx.QueryRow(ctx, `
		SELECT r.currency, r.effective_from, `+cashRateLatest+`, `+cashRateAccruedThrough+`
		FROM cash_yield_rates r
		WHERE r.id = $1
		FOR UPDATE OF r
	`, rateID).Scan(&locked.currency, &locked.effectiveFrom, &latest, &locked.accruedThrough)
	if errors.Is(err, pgx.ErrNoRows) {
		return locked, ErrCashRateNotFound
	}
	if err != nil {
		return locked, err
	}

	if !latest {
		return locked, ErrCashRateNotLatest
	}

	return locked, nil
}

// UpdateCashRate corrects what the latest version says, while no day has been
// computed at it. Its dates stay: when a rate applies is changed with a new
// version, not by moving an old one.
func (r *PostgresRepository) UpdateCashRate(ctx context.Context, userID, rateID uuid.UUID, in CashRateInput) (CashRate, error) {
	var rate CashRate

	if err := database.WithinTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		locked, err := lockLatestCashRate(ctx, tx, userID, rateID)
		if err != nil {
			return err
		}

		if locked.accruedThrough != nil {
			return cashRateInUse(*locked.accruedThrough, "record a new version instead")
		}

		if _, err := tx.Exec(ctx, `
			UPDATE cash_yield_rates SET
				annual_rate      = $2::numeric / 100,
				withholding_rate = $3::numeric / 100,
				posting          = $4::cash_interest_posting,
				max_balance      = $5::numeric
			WHERE id = $1
		`, rateID, in.AnnualRatePct.String(), in.WithholdingPct.String(), string(in.Posting),
			maxBalanceParam(in.MaxBalance)); err != nil {
			return err
		}

		rate, err = getCashRate(ctx, tx, userID, rateID)

		return err
	}); err != nil {
		return CashRate{}, err
	}

	return rate, nil
}

// EndCashRate stops the latest version from endsOn, the first day it no longer
// earns. It cannot stop before a day already computed at it, and a version that
// has not started by then is not paused but deleted.
func (r *PostgresRepository) EndCashRate(ctx context.Context, userID, rateID uuid.UUID, endsOn time.Time) (CashRate, error) {
	var rate CashRate
	lastDay := cashRateDay(endsOn).AddDate(0, 0, -1)

	if err := database.WithinTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		locked, err := lockLatestCashRate(ctx, tx, userID, rateID)
		if err != nil {
			return err
		}

		if through := locked.accruedThrough; through != nil && lastDay.Before(*through) {
			return cashRateInUse(*through, "it can stop from "+through.AddDate(0, 0, 1).Format(time.DateOnly))
		}

		if lastDay.Before(locked.effectiveFrom) {
			return invalidCashRate("the rate does not start until %s; delete it instead", locked.effectiveFrom.Format(time.DateOnly))
		}

		if _, err := tx.Exec(ctx, `
			UPDATE cash_yield_rates SET ended_on = $2::date WHERE id = $1
		`, rateID, lastDay.Format(time.DateOnly)); err != nil {
			return err
		}

		rate, err = getCashRate(ctx, tx, userID, rateID)

		return err
	}); err != nil {
		return CashRate{}, err
	}

	return rate, nil
}

// DeleteCashRate removes the latest version of an account's rate, while no day
// has been computed at it.
//
// If the version before it ends the day before this one starts, it was ended
// to make room for this one, and it runs on again: deleting a mistaken change
// of rate leaves the account at the rate it had. A rate paused and resumed on
// the very next day reads the same way, and deleting the resumption reopens it
// too.
func (r *PostgresRepository) DeleteCashRate(ctx context.Context, userID, rateID uuid.UUID) error {
	return database.WithinTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		locked, err := lockLatestCashRate(ctx, tx, userID, rateID)
		if err != nil {
			return err
		}

		if locked.accruedThrough != nil {
			return cashRateInUse(*locked.accruedThrough, "end it instead")
		}

		if _, err := tx.Exec(ctx, `DELETE FROM cash_yield_rates WHERE id = $1`, rateID); err != nil {
			return err
		}

		_, err = tx.Exec(ctx, `
			UPDATE cash_yield_rates SET ended_on = NULL
			WHERE source_id = $1
			  AND currency  = $2::char(3)
			  AND ended_on  = $3::date - 1
			  AND effective_from = (
			    SELECT MAX(effective_from) FROM cash_yield_rates
			    WHERE source_id = $1 AND currency = $2::char(3)
			  )
		`, locked.sourceID, locked.currency, locked.effectiveFrom.Format(time.DateOnly))

		return err
	})
}
