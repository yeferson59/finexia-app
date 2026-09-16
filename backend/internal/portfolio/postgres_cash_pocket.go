package portfolio

import (
	"context"
	"errors"
	"time"

	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
// The balances it opened go with it. They are empty shells — a position at zero
// with no transaction on it — and they only exist because a movement opened one
// and was then deleted; portfolio_entries.pocket_id does not cascade, so they
// are deleted here rather than by the database.
func (r *PostgresRepository) DeleteCashPocket(ctx context.Context, userID, pocketID uuid.UUID) error {
	return database.WithinTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		locked, err := lockCashPocket(ctx, tx, userID, pocketID)
		if err != nil {
			return err
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

		if _, err := tx.Exec(ctx, `DELETE FROM portfolio_entries WHERE pocket_id = $1`, locked.id); err != nil {
			return err
		}

		_, err = tx.Exec(ctx, `DELETE FROM cash_pockets WHERE id = $1`, locked.id)

		return err
	})
}
