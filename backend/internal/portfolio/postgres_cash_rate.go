package portfolio

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/platform/database"
)

// Every date these writes send travels as its calendar day in text. A
// time.Time is encoded as a timestamp, and casting that to a date happens in
// the session's time zone, which is not guaranteed to be UTC.

// cashRateLatest is the SQL test behind CashRate.Latest: no version of the same
// account starts later. The pocket is part of the account (000047), so a rate
// given to a pocket does not make the main account's rate stop being the latest
// of its own.
const cashRateLatest = `NOT EXISTS (
	SELECT 1 FROM cash_yield_rates later
	WHERE later.source_id = r.source_id
	  AND later.currency  = r.currency
	  AND later.pocket_id IS NOT DISTINCT FROM r.pocket_id
	  AND later.effective_from > r.effective_from
)`

// cashRateAccruedThrough is the SQL behind CashRate.AccruedThrough.
const cashRateAccruedThrough = `(
	SELECT MAX(ac.accrual_date) FROM cash_interest_accruals ac WHERE ac.rate_id = r.id
)`

// The percentages come back trimmed: "9.25", not "9.250000".
const cashRateColumns = `
	r.id, r.source_id, s.name, r.currency,
	r.pocket_id, COALESCE(pk.name, ''),
	trim_scale(r.annual_rate * 100)::text,
	trim_scale(r.withholding_rate * 100)::text,
	r.posting::text,
	COALESCE((
		SELECT json_agg(json_build_object(
			'fromBalance',   trim_scale(t.from_balance)::text,
			'annualRatePct', trim_scale(t.annual_rate * 100)::text
		) ORDER BY t.from_balance)
		FROM cash_yield_rate_tiers t
		WHERE t.rate_id = r.id
	), '[]')::text,
	r.effective_from, r.ended_on,
	` + cashRateLatest + `,
	` + cashRateAccruedThrough + `,
	r.created_at, r.updated_at`

const cashRateFrom = `
	FROM cash_yield_rates r
	JOIN investment_sources s      ON s.id = r.source_id
	LEFT JOIN cash_pockets pk      ON pk.id = r.pocket_id`

// writeCashRateTiers states a version's tiers whole: the ones it had go, and the
// ones given take their place. The caller holds the lock every write to a rate
// takes.
func writeCashRateTiers(ctx context.Context, tx pgx.Tx, rateID uuid.UUID, tiers []CashRateTierInput) error {
	if _, err := tx.Exec(ctx, `DELETE FROM cash_yield_rate_tiers WHERE rate_id = $1`, rateID); err != nil {
		return err
	}

	if len(tiers) == 0 {
		return nil
	}

	from := make([]string, len(tiers))
	pct := make([]string, len(tiers))
	for i, tier := range tiers {
		from[i] = tier.FromBalance.String()
		pct[i] = tier.AnnualRatePct.String()
	}

	_, err := tx.Exec(ctx, `
		INSERT INTO cash_yield_rate_tiers (rate_id, from_balance, annual_rate)
		SELECT $1::uuid, step.from_balance::numeric, step.pct::numeric / 100
		FROM unnest($2::text[], $3::text[]) AS step(from_balance, pct)
	`, rateID, from, pct)

	return err
}

func scanCashRate(row pgx.Row) (CashRate, error) {
	var (
		rate CashRate
		// Never NULL: a version without tiers comes back as an empty array.
		tiers string
	)

	if err := row.Scan(
		&rate.ID,
		&rate.SourceID,
		&rate.SourceName,
		&rate.Currency,
		&rate.PocketID,
		&rate.PocketName,
		&rate.AnnualRatePct,
		&rate.WithholdingPct,
		&rate.Posting,
		&tiers,
		&rate.EffectiveFrom,
		&rate.EndedOn,
		&rate.Latest,
		&rate.AccruedThrough,
		&rate.CreatedAt,
		&rate.UpdatedAt,
	); err != nil {
		return rate, err
	}

	err := json.Unmarshal([]byte(tiers), &rate.Tiers)

	return rate, err
}

// GetCashRatesByUserID lists every version of every rate on the user's
// platforms: by platform and currency, the newest version first.
func (r *PostgresRepository) GetCashRatesByUserID(ctx context.Context, userID uuid.UUID) ([]CashRate, error) {
	rows, err := r.db.Query(ctx, `SELECT `+cashRateColumns+cashRateFrom+`
		WHERE s.user_id = $1
		ORDER BY s.name, r.currency, pk.name NULLS FIRST, r.effective_from DESC
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
// a version in before it would rewrite what it said.
//
// The first day can be in the past: the rate the account earned before it was
// recorded. A first day whose interest was already computed at the rate before
// is refused unless in.Recompute says otherwise; with it, the account's days
// from then are thrown away (clearCashInterestFrom), and the service computes
// them again at this version once it is written.
//
// The platform row is locked first. Every write to a rate takes that lock, so
// versions of one account are written one at a time — including the first,
// which has no version yet to lock. The account's balances follow, before any
// version is touched (lockCashAccountBalances).
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

		pocket, err := requireRatePocket(ctx, tx, userID, in.PocketID, in.SourceID, in.Currency)
		if err != nil {
			return err
		}

		balances, err := lockCashAccountBalances(ctx, tx, in.SourceID, in.Currency, pocket)
		if err != nil {
			return err
		}

		var (
			latestID   uuid.UUID
			latestFrom time.Time
			latestEnd  *time.Time
		)
		err = tx.QueryRow(ctx, `
			SELECT id, effective_from, ended_on
			FROM cash_yield_rates
			WHERE source_id = $1 AND currency = $2::char(3) AND pocket_id IS NOT DISTINCT FROM $3::uuid
			ORDER BY effective_from DESC
			LIMIT 1
		`, in.SourceID, in.Currency, pocket).Scan(&latestID, &latestFrom, &latestEnd)

		switch {
		case errors.Is(err, pgx.ErrNoRows):
		case err != nil:
			return err
		case !latestFrom.Before(start):
			return fmt.Errorf("%w: the latest starts on %s", ErrCashRateOverlaps, latestFrom.Format(time.DateOnly))
		default:
			accruedThrough, err := cashAccountAccruedThrough(ctx, tx, in.SourceID, in.Currency, pocket)
			if err != nil {
				return err
			}

			if accruedThrough != nil && !accruedThrough.Before(start) {
				if !in.Recompute {
					return cashRateInUse(*accruedThrough, "the new version can start from "+accruedThrough.AddDate(0, 0, 1).Format(time.DateOnly))
				}

				if _, err := clearCashInterestFrom(ctx, tx, balances, start); err != nil {
					return err
				}
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
			INSERT INTO cash_yield_rates (source_id, currency, pocket_id, annual_rate, withholding_rate, posting, effective_from)
			VALUES ($1::uuid, $2::char(3), $7::uuid, $3::numeric / 100, $4::numeric / 100, $5::cash_interest_posting, $6::date)
			RETURNING id
		`, in.SourceID, in.Currency, in.AnnualRatePct.String(), in.WithholdingPct.String(), string(in.Posting),
			start.Format(time.DateOnly), pocket).Scan(&rateID); err != nil {
			// The lock makes this unreachable from here; a writer that skipped it
			// would still get the answer the check above gives.
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return ErrCashRateOverlaps
			}

			return err
		}

		if err := writeCashRateTiers(ctx, tx, rateID, in.Tiers); err != nil {
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
	sourceID uuid.UUID
	currency money.Currency
	// pocketID is the drawer of the account the version belongs to, nil for the
	// main account: part of the key every other version is compared against.
	pocketID       *uuid.UUID
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

	var latest, fixed bool
	err = tx.QueryRow(ctx, `
		SELECT r.currency, r.pocket_id, r.effective_from, `+cashRateLatest+`, `+cashRateAccruedThrough+`,
		       EXISTS (SELECT 1 FROM cash_pockets pk WHERE pk.id = r.pocket_id AND pk.kind = 'fixed')
		FROM cash_yield_rates r
		WHERE r.id = $1
		FOR UPDATE OF r
	`, rateID).Scan(&locked.currency, &locked.pocketID, &locked.effectiveFrom, &latest, &locked.accruedThrough, &fixed)
	if errors.Is(err, pgx.ErrNoRows) {
		return locked, ErrCashRateNotFound
	}
	if err != nil {
		return locked, err
	}

	// The rate of a fixed deposit is the deposit: correcting it, pausing it or
	// deleting it would change the terms of money already placed. Cancelling the
	// deposit is what ends it, and deleting the deposit is what removes it.
	if fixed {
		return locked, fmt.Errorf("%w: cancel the deposit to end it, or delete it whole", ErrCashPocketFixed)
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
				posting          = $4::cash_interest_posting
			WHERE id = $1
		`, rateID, in.AnnualRatePct.String(), in.WithholdingPct.String(), string(in.Posting)); err != nil {
			return err
		}

		if err := writeCashRateTiers(ctx, tx, rateID, in.Tiers); err != nil {
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
			  AND pocket_id IS NOT DISTINCT FROM $4::uuid
			  AND ended_on  = $3::date - 1
			  AND effective_from = (
			    SELECT MAX(effective_from) FROM cash_yield_rates
			    WHERE source_id = $1 AND currency = $2::char(3) AND pocket_id IS NOT DISTINCT FROM $4::uuid
			  )
		`, locked.sourceID, locked.currency, locked.effectiveFrom.Format(time.DateOnly), locked.pocketID)

		return err
	})
}

// cashAccountAccruedThrough is the last day the ledger computed at any version
// of an account's rate, nil if none has been.
func cashAccountAccruedThrough(ctx context.Context, tx pgx.Tx, sourceID uuid.UUID, cur money.Currency, pocketID *uuid.UUID) (*time.Time, error) {
	var accruedThrough *time.Time
	err := tx.QueryRow(ctx, `
		SELECT MAX(ac.accrual_date)
		FROM cash_interest_accruals ac
		JOIN cash_yield_rates v ON v.id = ac.rate_id
		WHERE v.source_id = $1 AND v.currency = $2::char(3) AND v.pocket_id IS NOT DISTINCT FROM $3::uuid
	`, sourceID, cur, pocketID).Scan(&accruedThrough)

	return accruedThrough, err
}

// RescheduleCashRate moves the first day of the latest version, while no day
// has been computed at it: a change of rate announced for one day and made on
// another.
//
// It moves within the room the versions around it leave. It starts after the
// version before it does, and not after the day it stops earning if it was
// paused. The version before it follows it: if it was ended to make room for
// this one — it ended the day before this one started — it now ends the day
// before the new first day, whether that is earlier or later. One paused with
// days of no rate before this one keeps its pause, unless the new first day
// falls inside it.
//
// Moving the first day back into days already computed at the version before
// is CreateCashRate's past first day: refused, unless in.Recompute throws them
// away to compute them again.
func (r *PostgresRepository) RescheduleCashRate(ctx context.Context, userID, rateID uuid.UUID, in RescheduleCashRateInput) (CashRate, error) {
	var rate CashRate
	start := cashRateDay(in.EffectiveFrom)

	if err := database.WithinTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		// The platform, then the account's balances, then the versions: the
		// order CreateCashRate takes them in. lockLatestCashRate takes the
		// platform again, which the transaction already holds.
		var (
			sourceID uuid.UUID
			cur      money.Currency
			pocketID *uuid.UUID
		)
		err := tx.QueryRow(ctx, `
			SELECT r.source_id, r.currency, r.pocket_id
			FROM cash_yield_rates r
			JOIN investment_sources s ON s.id = r.source_id
			WHERE r.id = $1 AND s.user_id = $2
			FOR UPDATE OF s
		`, rateID, userID).Scan(&sourceID, &cur, &pocketID)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrCashRateNotFound
		}
		if err != nil {
			return err
		}

		balances, err := lockCashAccountBalances(ctx, tx, sourceID, cur, pocketID)
		if err != nil {
			return err
		}

		locked, err := lockLatestCashRate(ctx, tx, userID, rateID)
		if err != nil {
			return err
		}

		if locked.accruedThrough != nil {
			return cashRateInUse(*locked.accruedThrough, "record a new version instead")
		}

		var endedOn *time.Time
		if err := tx.QueryRow(ctx, `SELECT ended_on FROM cash_yield_rates WHERE id = $1`, rateID).Scan(&endedOn); err != nil {
			return err
		}

		if endedOn != nil && start.After(*endedOn) {
			return invalidCashRate("the rate stops earning after %s, so it cannot start later", endedOn.Format(time.DateOnly))
		}

		var (
			previousID   uuid.UUID
			previousFrom time.Time
			previousEnd  *time.Time
		)
		err = tx.QueryRow(ctx, `
			SELECT id, effective_from, ended_on
			FROM cash_yield_rates
			WHERE source_id = $1 AND currency = $2::char(3) AND pocket_id IS NOT DISTINCT FROM $3::uuid
			  AND effective_from < $4::date
			ORDER BY effective_from DESC
			LIMIT 1
			FOR UPDATE
		`, sourceID, cur, pocketID, locked.effectiveFrom.Format(time.DateOnly)).Scan(&previousID, &previousFrom, &previousEnd)

		previous := !errors.Is(err, pgx.ErrNoRows)
		if err != nil && previous {
			return err
		}

		if previous {
			if !previousFrom.Before(start) {
				return invalidCashRate("effectiveFrom must be after %s, when the version before it starts", previousFrom.Format(time.DateOnly))
			}

			accruedThrough, err := cashAccountAccruedThrough(ctx, tx, sourceID, cur, pocketID)
			if err != nil {
				return err
			}

			if accruedThrough != nil && !accruedThrough.Before(start) {
				if !in.Recompute {
					return cashRateInUse(*accruedThrough, "it can start from "+accruedThrough.AddDate(0, 0, 1).Format(time.DateOnly))
				}

				if _, err := clearCashInterestFrom(ctx, tx, balances, start); err != nil {
					return err
				}
			}

			madeRoom := previousEnd != nil && previousEnd.Equal(locked.effectiveFrom.AddDate(0, 0, -1))
			if previousEnd == nil || madeRoom || !previousEnd.Before(start) {
				if _, err := tx.Exec(ctx, `
					UPDATE cash_yield_rates SET ended_on = $2::date - 1 WHERE id = $1
				`, previousID, start.Format(time.DateOnly)); err != nil {
					return err
				}
			}
		}

		if _, err := tx.Exec(ctx, `
			UPDATE cash_yield_rates SET effective_from = $2::date WHERE id = $1
		`, rateID, start.Format(time.DateOnly)); err != nil {
			return err
		}

		rate, err = getCashRate(ctx, tx, userID, rateID)

		return err
	}); err != nil {
		return CashRate{}, err
	}

	return rate, nil
}

// requireRatePocket turns the pocket a rate names into the value
// cash_yield_rates.pocket_id takes: nil for the main account, or the pocket
// itself once it is known to be the owner's and part of the account stated.
//
// It is requireWritablePocket's answer for a rate: the pocket is locked with
// the platform, so a version and a rename cannot cross.
func requireRatePocket(ctx context.Context, tx pgx.Tx, userID, pocketID, sourceID uuid.UUID, cur money.Currency) (*uuid.UUID, error) {
	if pocketID == (uuid.UUID{}) {
		return nil, nil
	}

	locked, err := lockCashPocket(ctx, tx, userID, pocketID)
	if err != nil {
		return nil, err
	}

	if locked.sourceID != sourceID || locked.currency != cur {
		return nil, invalidCashRate("that pocket belongs to another account, so it cannot earn this one's rate")
	}

	// A fixed deposit keeps the rate of the day it was opened, so it never gets
	// another version (000048). Its one version is written with it, by
	// OpenFixedDeposit.
	if locked.kind == PocketFixed {
		return nil, fmt.Errorf("%w: it keeps the rate of the day it was opened", ErrCashPocketFixed)
	}

	return &locked.id, nil
}
