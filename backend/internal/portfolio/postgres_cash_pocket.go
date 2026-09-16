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

// The pockets of a cash account (000047). Every write locks the platform first,
// the way a rate's does, so the pockets of one account are written one at a
// time and two requests cannot both find a name free.

// cashPocketColumns is a pocket and what says whether it can still be deleted:
// what its balances hold, how many there are, and how many movements they ever
// took.
const cashPocketColumns = `
	p.id, p.source_id, s.name, p.currency, p.name, p.kind::text,
	p.opened_on, p.matures_on, p.closed_on,
	ROUND(held.balance, 8)::text, held.balances, held.movements,
	p.created_at, p.updated_at`

const cashPocketFrom = `
	FROM cash_pockets p
	JOIN investment_sources s ON s.id = p.source_id
	CROSS JOIN LATERAL (
		SELECT
			COALESCE(SUM(pe.quantity), 0)::numeric AS balance,
			COUNT(pe.id)                           AS balances,
			COALESCE(SUM((SELECT COUNT(*) FROM transactions t WHERE t.entry_id = pe.id)), 0) AS movements
		FROM portfolio_entries pe
		WHERE pe.pocket_id = p.id
	) held`

func scanCashPocket(row pgx.Row) (CashPocket, error) {
	var pocket CashPocket

	err := row.Scan(
		&pocket.ID,
		&pocket.SourceID,
		&pocket.SourceName,
		&pocket.Currency,
		&pocket.Name,
		&pocket.Kind,
		&pocket.OpenedOn,
		&pocket.MaturesOn,
		&pocket.ClosedOn,
		&pocket.Balance,
		&pocket.Balances,
		&pocket.Movements,
		&pocket.CreatedAt,
		&pocket.UpdatedAt,
	)

	return pocket, err
}

func getCashPocket(ctx context.Context, tx pgx.Tx, userID, pocketID uuid.UUID) (CashPocket, error) {
	pocket, err := scanCashPocket(tx.QueryRow(ctx, `SELECT `+cashPocketColumns+cashPocketFrom+`
		WHERE p.id = $1 AND s.user_id = $2
	`, pocketID, userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return CashPocket{}, ErrCashPocketNotFound
	}

	return pocket, err
}

// GetCashPocketsByUserID lists every pocket the user has, open and closed, in
// the order the cash screen shows them: by platform, by currency, and by name
// inside each account.
func (r *PostgresRepository) GetCashPocketsByUserID(ctx context.Context, userID uuid.UUID) ([]CashPocket, error) {
	rows, err := r.db.Query(ctx, `SELECT `+cashPocketColumns+cashPocketFrom+`
		WHERE s.user_id = $1
		ORDER BY s.name, p.currency, p.name
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pockets := make([]CashPocket, 0)
	for rows.Next() {
		pocket, err := scanCashPocket(rows)
		if err != nil {
			return nil, err
		}

		pockets = append(pockets, pocket)
	}

	return pockets, rows.Err()
}

// lockedCashPocket is what a write that names a pocket has to know about it
// first: the account it belongs to, and what kind it is.
type lockedCashPocket struct {
	id       uuid.UUID
	sourceID uuid.UUID
	currency money.Currency
	kind     CashPocketKind
	closedOn *time.Time
}

// lockCashPocket finds a pocket the user owns and takes the lock every write to
// its account takes. The zero pocket — the main account — is not a row, so it
// is never passed here.
func lockCashPocket(ctx context.Context, tx pgx.Tx, userID, pocketID uuid.UUID) (lockedCashPocket, error) {
	var locked lockedCashPocket

	err := tx.QueryRow(ctx, `
		SELECT p.id, p.source_id, p.currency, p.kind::text, p.closed_on
		FROM cash_pockets p
		JOIN investment_sources s ON s.id = p.source_id
		WHERE p.id = $1 AND s.user_id = $2
		FOR UPDATE OF s, p
	`, pocketID, userID).Scan(&locked.id, &locked.sourceID, &locked.currency, &locked.kind, &locked.closedOn)
	if errors.Is(err, pgx.ErrNoRows) {
		return locked, ErrCashPocketNotFound
	}

	return locked, err
}

// isPocketNameTaken reports whether a Postgres error is the unique key over an
// account's pocket names.
func isPocketNameTaken(err error) bool {
	var pgErr *pgconn.PgError

	return errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "uk_cash_pockets_name"
}

// CreateCashPocket opens a flexible pocket on an account.
//
// It opens today: a pocket is a drawer of an account that already exists, and
// the money that goes into it arrives with a movement of its own, dated
// whenever the owner says. A fixed deposit is the one that opens in the past,
// and it comes with its own writer.
func (r *PostgresRepository) CreateCashPocket(ctx context.Context, userID uuid.UUID, in NewCashPocketInput) (CashPocket, error) {
	var pocket CashPocket

	if err := database.WithinTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		var active bool
		if err := tx.QueryRow(ctx, `
			SELECT is_active FROM investment_sources WHERE id = $1 AND user_id = $2 FOR UPDATE
		`, in.SourceID, userID).Scan(&active); errors.Is(err, pgx.ErrNoRows) {
			return ErrPlatformNotFound
		} else if err != nil {
			return err
		}

		// An inactive platform takes no new money, so it opens no new drawer to
		// put it in — the same rule a new rate follows.
		if !active {
			return invalidCashPocket("the platform is inactive; activate it before giving it a pocket")
		}

		var pocketID uuid.UUID
		if err := tx.QueryRow(ctx, `
			INSERT INTO cash_pockets (source_id, currency, name, kind, opened_on)
			VALUES ($1::uuid, $2::char(3), $3, 'flexible', $4::date)
			RETURNING id
		`, in.SourceID, in.Currency, in.CleanName(), cashRateDay(time.Now()).Format(time.DateOnly)).Scan(&pocketID); err != nil {
			if isPocketNameTaken(err) {
				return ErrCashPocketNameTaken
			}

			return err
		}

		var err error
		pocket, err = getCashPocket(ctx, tx, userID, pocketID)

		return err
	}); err != nil {
		return CashPocket{}, err
	}

	return pocket, nil
}

// RenameCashPocket gives a pocket another name. Nothing else about it can be
// rewritten: what it holds moves with movements, and when it earns moves with
// its rate.
func (r *PostgresRepository) RenameCashPocket(ctx context.Context, userID, pocketID uuid.UUID, in RenameCashPocketInput) (CashPocket, error) {
	var pocket CashPocket

	if err := database.WithinTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		locked, err := lockCashPocket(ctx, tx, userID, pocketID)
		if err != nil {
			return err
		}

		if _, err := tx.Exec(ctx, `
			UPDATE cash_pockets SET name = $2 WHERE id = $1
		`, locked.id, in.CleanName()); err != nil {
			if isPocketNameTaken(err) {
				return ErrCashPocketNameTaken
			}

			return err
		}

		pocket, err = getCashPocket(ctx, tx, userID, pocketID)

		return err
	}); err != nil {
		return CashPocket{}, err
	}

	return pocket, nil
}

// DeleteCashPocket removes a pocket that never held anything, with the rate
// versions it earned at.
//
// A pocket that still holds money, or that ever took a movement, is refused:
// deleting it would either lose what it holds or rewrite a history that
// happened. Emptying it is a move back to the main account, and that is what
// the owner is asked to do first.
//
// A fixed deposit is the exception, and deliberately so. It cannot be emptied —
// nothing is written in it by hand — and what it holds is one deposit, its
// interest and its rate, all recorded by the app in one go. So deleting one
// takes the whole thing, which is what "I recorded this wrong" means: the money
// never was on the platform, and a cancellation, which does leave a history,
// is what to use when it was.
//
// The balances it opened go with it, and their transactions and ledger days
// with them, by the cascade on portfolio_entries. For a flexible pocket they
// are empty shells — a position at zero with no transaction on it — left behind
// when a movement opened one and was then deleted; portfolio_entries.pocket_id
// does not cascade, so they are deleted here rather than by the database.
func (r *PostgresRepository) DeleteCashPocket(ctx context.Context, userID, pocketID uuid.UUID) error {
	return database.WithinTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		locked, err := lockCashPocket(ctx, tx, userID, pocketID)
		if err != nil {
			return err
		}

		if locked.kind == PocketFixed {
			return deleteCashPocketRows(ctx, tx, locked.id)
		}

		var inUse bool
		if err := tx.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM portfolio_entries pe
				WHERE pe.pocket_id = $1
				  AND (pe.quantity <> 0 OR EXISTS (SELECT 1 FROM transactions t WHERE t.entry_id = pe.id))
			)
		`, locked.id).Scan(&inUse); err != nil {
			return err
		}

		if inUse {
			return ErrCashPocketNotEmpty
		}

		return deleteCashPocketRows(ctx, tx, locked.id)
	})
}

// deleteCashPocketRows takes a pocket and its balances away. The rate versions
// it earned at go with it by the cascade on cash_yield_rates, and each balance
// takes its transactions and its ledger days with it by the cascade on
// portfolio_entries.
func deleteCashPocketRows(ctx context.Context, tx pgx.Tx, pocketID uuid.UUID) error {
	if _, err := tx.Exec(ctx, `DELETE FROM portfolio_entries WHERE pocket_id = $1`, pocketID); err != nil {
		return err
	}

	_, err := tx.Exec(ctx, `DELETE FROM cash_pockets WHERE id = $1`, pocketID)

	return err
}

// GetCashPocketByID reads one pocket the user owns, with what it holds now. The
// writes that change a deposit answer with it, so the screen sees the money and
// the interest the same call moved.
func (r *PostgresRepository) GetCashPocketByID(ctx context.Context, userID, pocketID uuid.UUID) (CashPocket, error) {
	var pocket CashPocket

	err := database.WithinTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		pocket, err = getCashPocket(ctx, tx, userID, pocketID)

		return err
	})

	return pocket, err
}

// The fixed deposits (000048). A deposit is a pocket of kind fixed, and the
// three writes below are its whole life: it is opened with its money, its rate
// and its term in one transaction, it earns until the day before it comes due,
// and on that day its balance goes back to the main account.
//
// Nothing else writes to it. requireWritablePocket, requireRatePocket and the
// generic transaction writers all refuse a fixed pocket, so the deposit that
// opened it stays the only one and its rate stays the rate of the day it was
// opened. What these three do, they do through writeCashMovement with the
// pocket in hand, which is the path those guards sit in front of rather than
// inside.

// OpenFixedDeposit records a deposit whole: the pocket, its balance, the deposit
// that opened it, and the one version of the rate it keeps.
//
// It is the one cash write that may be dated in the past. A deposit takes no
// movements by hand, so nothing the platform already paid can be recorded twice:
// the days between the day it opened and yesterday are computed straight away by
// the caller, and credited as the return they were.
//
// The rate ends the day before the deposit comes due, which is what makes the
// last day it earns the day before maturity and, for a rate posted at maturity,
// the day the whole term is credited on. Without a term it has no end.
func (r *PostgresRepository) OpenFixedDeposit(ctx context.Context, userID uuid.UUID, in NewFixedDepositInput) (CashPocket, error) {
	var pocket CashPocket

	opened := cashRateDay(in.OpenedOn).Format(time.DateOnly)

	var matures *string
	if in.MaturesOn != nil {
		day := cashRateDay(*in.MaturesOn).Format(time.DateOnly)
		matures = &day
	}

	if err := database.WithinTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		if err := requireCashAccount(ctx, tx, userID, in.PortfolioID, in.SourceID); err != nil {
			return err
		}

		var active bool
		if err := tx.QueryRow(ctx, `
			SELECT is_active FROM investment_sources WHERE id = $1 AND user_id = $2 FOR UPDATE
		`, in.SourceID, userID).Scan(&active); errors.Is(err, pgx.ErrNoRows) {
			return ErrPlatformNotFound
		} else if err != nil {
			return err
		}

		// An inactive platform takes no new money, so no new deposit either —
		// the same rule a pocket and a rate follow.
		if !active {
			return invalidCashPocket("the platform is inactive; activate it before opening a deposit on it")
		}

		var pocketID uuid.UUID
		if err := tx.QueryRow(ctx, `
			INSERT INTO cash_pockets (source_id, currency, name, kind, opened_on, matures_on)
			VALUES ($1::uuid, $2::char(3), $3, 'fixed', $4::date, $5::date)
			RETURNING id
		`, in.SourceID, in.Currency, in.CleanName(), opened, matures).Scan(&pocketID); err != nil {
			if isPocketNameTaken(err) {
				return ErrCashPocketNameTaken
			}

			return err
		}

		// The balance opens with the deposit, dated the day the money went in.
		if _, err := writeCashMovement(ctx, tx, userID, in.PortfolioID, in.SourceID, &pocketID, CashMovementInput{
			Kind:     CashKindDeposit,
			Amount:   in.Amount,
			Currency: in.Currency,
			Date:     cashRateDay(in.OpenedOn),
			Notes:    "Apertura del depósito",
		}); err != nil {
			return err
		}

		var rateID uuid.UUID
		if err := tx.QueryRow(ctx, `
			INSERT INTO cash_yield_rates (source_id, currency, pocket_id, annual_rate, withholding_rate, posting, effective_from, ended_on)
			VALUES ($1::uuid, $2::char(3), $3::uuid, $4::numeric / 100, $5::numeric / 100, $6::cash_interest_posting, $7::date, $8::date - 1)
			RETURNING id
		`, in.SourceID, in.Currency, pocketID, in.AnnualRatePct.String(), in.WithholdingPct.String(),
			string(in.Posting), opened, matures).Scan(&rateID); err != nil {
			return err
		}

		if err := writeCashRateTiers(ctx, tx, rateID, in.Tiers); err != nil {
			return err
		}

		var err error
		pocket, err = getCashPocket(ctx, tx, userID, pocketID)

		return err
	}); err != nil {
		return CashPocket{}, err
	}

	return pocket, nil
}

// EndFixedDeposit stops a deposit earning from closesOn, the first day it no
// longer does, and leaves the money where it is. It is the first half of
// cancelling one: the caller computes the days it still owes at the rate that
// just ended, and then settles it.
//
// A day already computed cannot be taken back, so a cancellation cannot reach
// behind one — the same answer ending an account's rate gives. Past the day it
// comes due there is nothing to shorten: the rate already ends the day before,
// and a cancellation dated later leaves it there.
func (r *PostgresRepository) EndFixedDeposit(ctx context.Context, userID, pocketID uuid.UUID, closesOn time.Time) (CashPocket, error) {
	var pocket CashPocket
	lastDay := cashRateDay(closesOn).AddDate(0, 0, -1)

	if err := database.WithinTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		locked, err := lockFixedDeposit(ctx, tx, userID, pocketID)
		if err != nil {
			return err
		}

		var (
			rateID         uuid.UUID
			effectiveFrom  time.Time
			accruedThrough *time.Time
		)
		err = tx.QueryRow(ctx, `
			SELECT r.id, r.effective_from, (
				SELECT MAX(ac.accrual_date) FROM cash_interest_accruals ac WHERE ac.rate_id = r.id
			)
			FROM cash_yield_rates r
			WHERE r.pocket_id = $1
			ORDER BY r.effective_from DESC
			LIMIT 1
			FOR UPDATE OF r
		`, locked.id).Scan(&rateID, &effectiveFrom, &accruedThrough)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return err
		}

		if !errors.Is(err, pgx.ErrNoRows) {
			if accruedThrough != nil && lastDay.Before(*accruedThrough) {
				return cashRateInUse(*accruedThrough, "the deposit can be cancelled from "+accruedThrough.AddDate(0, 0, 1).Format(time.DateOnly))
			}

			// Cancelled before it earned a single day, the version never applied:
			// it goes, rather than ending before it started, which the ledger
			// would not accept.
			if lastDay.Before(effectiveFrom) {
				if _, err := tx.Exec(ctx, `DELETE FROM cash_yield_rates WHERE id = $1`, rateID); err != nil {
					return err
				}
			} else if _, err := tx.Exec(ctx, `
				UPDATE cash_yield_rates SET ended_on = LEAST($2::date, ended_on)
				WHERE id = $1
			`, rateID, lastDay.Format(time.DateOnly)); err != nil {
				return err
			}
		}

		pocket, err = getCashPocket(ctx, tx, userID, locked.id)

		return err
	}); err != nil {
		return CashPocket{}, err
	}

	return pocket, nil
}

// SettleFixedDeposit moves what a deposit holds back to the main account of its
// portfolio and closes it. It is what happens on the day it comes due, and the
// second half of cancelling one early.
//
// penalty is what the platform keeps for breaking the term. It rides on the
// withdrawal as its fee, so the two legs stop cancelling out by exactly that
// much: the money that arrives is what arrived, and what the platform kept
// reads as a loss rather than as money the owner took out. At maturity there is
// none, and the legs offset each other exactly — the return does not move.
//
// It is idempotent. A deposit it settles is closed, and closed_on is read under
// the lock, so a second run — the job catching up after a day down — finds
// nothing to move.
func (r *PostgresRepository) SettleFixedDeposit(ctx context.Context, userID, pocketID uuid.UUID, on time.Time, penalty decimal.Decimal, notes string) (CashPocket, error) {
	var pocket CashPocket

	if err := database.WithinTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		locked, err := lockFixedDeposit(ctx, tx, userID, pocketID)
		if err != nil {
			return err
		}

		if err := settleFixedDeposit(ctx, tx, userID, locked, cashRateDay(on), penalty, notes); err != nil {
			return err
		}

		pocket, err = getCashPocket(ctx, tx, userID, locked.id)

		return err
	}); err != nil {
		return CashPocket{}, err
	}

	return pocket, nil
}

// MatureCashPockets settles every deposit that has come due by `on` and is
// still open, and reports how many it moved. The nightly job runs it after the
// day's interest is computed, so the last day a deposit earns is credited before
// its balance leaves.
//
// A day the job was down is caught up by the next run, and the legs are dated on
// the day the deposit came due rather than on the day they were written: the two
// offset each other, so the history reads as the money never having moved.
func (r *PostgresRepository) MatureCashPockets(ctx context.Context, on time.Time) (int, error) {
	day := cashRateDay(on).Format(time.DateOnly)

	rows, err := r.db.Query(ctx, `
		SELECT p.id, s.user_id
		FROM cash_pockets p
		JOIN investment_sources s ON s.id = p.source_id
		WHERE p.kind = 'fixed'
		  AND p.closed_on IS NULL
		  AND p.matures_on IS NOT NULL
		  AND p.matures_on <= $1::date
		ORDER BY p.matures_on, p.id
	`, day)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	type due struct{ pocketID, userID uuid.UUID }

	var pockets []due
	for rows.Next() {
		var d due
		if err := rows.Scan(&d.pocketID, &d.userID); err != nil {
			return 0, err
		}

		pockets = append(pockets, d)
	}

	if err := rows.Err(); err != nil {
		return 0, err
	}

	// One deposit that cannot be settled — its main account holds that currency
	// as a position bought at another price, say — does not hold up the rest:
	// the first failure is reported once every other deposit has had its turn.
	matured, failed := 0, error(nil)

	for _, d := range pockets {
		settled := false

		err := database.WithinTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
			locked, err := lockFixedDeposit(ctx, tx, d.userID, d.pocketID)
			// Deleted or settled while this run was reading the list.
			if errors.Is(err, ErrCashPocketNotFound) || errors.Is(err, ErrCashPocketClosed) {
				return nil
			}
			if err != nil {
				return err
			}

			if locked.maturesOn == nil {
				return nil
			}

			settled = true

			return settleFixedDeposit(ctx, tx, d.userID, locked, cashRateDay(*locked.maturesOn), decimal.Zero, "Vencimiento del depósito")
		})

		switch {
		case err != nil && failed == nil:
			failed = err
		case err == nil && settled:
			matured++
		}
	}

	return matured, failed
}

// lockedFixedDeposit is what settling one has to know: where it sits, when it
// came due, and that it is still open.
type lockedFixedDeposit struct {
	id        uuid.UUID
	sourceID  uuid.UUID
	currency  money.Currency
	maturesOn *time.Time
}

// lockFixedDeposit is lockCashPocket for the writes only a deposit takes. A
// flexible pocket is refused — it is emptied with a move, not settled — and so
// is one already closed, which is what makes settling one idempotent.
func lockFixedDeposit(ctx context.Context, tx pgx.Tx, owner, pocketID uuid.UUID) (lockedFixedDeposit, error) {
	locked, err := lockCashPocket(ctx, tx, owner, pocketID)
	if err != nil {
		return lockedFixedDeposit{}, err
	}

	if locked.kind != PocketFixed {
		return lockedFixedDeposit{}, fmt.Errorf("%w: only a fixed deposit is cancelled; empty a flexible pocket with a move", ErrInvalidCashPocket)
	}

	if locked.closedOn != nil {
		return lockedFixedDeposit{}, ErrCashPocketClosed
	}

	var maturesOn *time.Time
	if err := tx.QueryRow(ctx, `SELECT matures_on FROM cash_pockets WHERE id = $1`, locked.id).Scan(&maturesOn); err != nil {
		return lockedFixedDeposit{}, err
	}

	return lockedFixedDeposit{id: locked.id, sourceID: locked.sourceID, currency: locked.currency, maturesOn: maturesOn}, nil
}

// settleFixedDeposit moves what the deposit holds to the main account and marks
// it closed, under the caller's lock.
//
// Both balances are locked before either is written, in the order of their
// position, so this and a move on the same account queue instead of deadlocking
// — the order MoveCash takes.
func settleFixedDeposit(ctx context.Context, tx pgx.Tx, owner uuid.UUID, locked lockedFixedDeposit, on time.Time, penalty decimal.Decimal, notes string) error {
	// Which portfolio holds it. A deposit is one lot of money bought once, so it
	// has one balance; the read takes no lock of its own, because the lock that
	// matters is the one below and taking one here first would be out of order.
	var (
		portfolioID uuid.UUID
		entryID     uuid.UUID
	)

	err := tx.QueryRow(ctx, `
		SELECT pe.portfolio_id, pe.id
		FROM portfolio_entries pe
		WHERE pe.pocket_id = $1
		ORDER BY pe.quantity DESC, pe.id
		LIMIT 1
	`, locked.id).Scan(&portfolioID, &entryID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	// A deposit that never took its money has nothing to move: it is closed and
	// left alone.
	if !errors.Is(err, pgx.ErrNoRows) {
		if _, err := tx.Exec(ctx, `
			SELECT pe.id
			FROM portfolio_entries pe
			WHERE pe.portfolio_id = $1
			  AND pe.source_id    = $2
			  AND (pe.pocket_id IS NOT DISTINCT FROM $3::uuid OR pe.pocket_id IS NULL)
			ORDER BY pe.id
			FOR UPDATE OF pe
		`, portfolioID, locked.sourceID, locked.id); err != nil {
			return err
		}

		// What it holds is read under that lock, so a credit landing at the same
		// time is either in it or waiting behind it.
		var balance decimal.Decimal
		if err := tx.QueryRow(ctx, `SELECT quantity FROM portfolio_entries WHERE id = $1`, entryID).Scan(&balance); err != nil {
			return err
		}

		// What the platform keeps cannot exceed what it holds: past that the
		// withdrawal would read as money the owner put in.
		if penalty.GreaterThan(balance) {
			penalty = balance
		}

		if balance.IsPos() {
			if _, err := writeCashMovement(ctx, tx, owner, portfolioID, locked.sourceID, &locked.id, CashMovementInput{
				Kind:     CashKindWithdrawal,
				Amount:   balance,
				Fees:     penalty,
				Currency: locked.currency,
				Date:     on,
				Notes:    notes,
			}); err != nil {
				return err
			}
		}

		if arriving := balance.Sub(penalty); arriving.IsPos() {
			if _, err := writeCashMovement(ctx, tx, owner, portfolioID, locked.sourceID, nil, CashMovementInput{
				Kind:     CashKindDeposit,
				Amount:   arriving,
				Currency: locked.currency,
				Date:     on,
				Notes:    notes,
			}); err != nil {
				return err
			}
		}
	}

	_, err = tx.Exec(ctx, `
		UPDATE cash_pockets SET closed_on = $2::date WHERE id = $1
	`, locked.id, on.Format(time.DateOnly))

	return err
}
