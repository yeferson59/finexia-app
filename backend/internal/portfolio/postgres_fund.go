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

// The investment funds (000056, 000057). A fund is a position like any other;
// what is written here is what a share does not need: the row that says the
// owner follows it, the marks that price it, and the snapshots a late mark
// revalues. Every write to a fund's marks locks its user_funds row first, so
// the marks of one fund are written one at a time and the price copied to
// user_asset_prices is always the latest of them.
//
// A fund can also reach a portfolio without going through CreateFund: a
// transaction file whose category says "fondo", or an asset created from the
// position form. The owner follows it all the same — it is in their portfolio —
// so a fund asset they hold counts as one they follow by units, and the first
// write to it records that (adoptHeldFund).

// fundQuerier is what the fund reads need, satisfied by the pool and by a
// transaction alike: a write answers with the fund as its own transaction
// left it.
type fundQuerier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

// GetFundsByUserID lists every fund the user follows, by name.
func (r *PostgresRepository) GetFundsByUserID(ctx context.Context, userID uuid.UUID) ([]Fund, error) {
	return readFunds(ctx, r.db, userID, nil)
}

// GetFund reads one fund the user follows.
func (r *PostgresRepository) GetFund(ctx context.Context, userID, assetID uuid.UUID) (Fund, error) {
	return readFund(ctx, r.db, userID, assetID)
}

func readFund(ctx context.Context, q fundQuerier, userID, assetID uuid.UUID) (Fund, error) {
	funds, err := readFunds(ctx, q, userID, &assetID)
	if err != nil {
		return Fund{}, err
	}

	if len(funds) == 0 {
		return Fund{}, ErrFundNotFound
	}

	return funds[0], nil
}

// readFunds reads the funds and then their positions, and adds the positions
// up in Go. Two statements rather than one aggregate because a fund is shown
// with the portfolios that hold it, and a fund nobody holds any more is still
// one the owner follows — with its marks — until they delete it.
func readFunds(ctx context.Context, q fundQuerier, userID uuid.UUID, assetID *uuid.UUID) ([]Fund, error) {
	rows, err := q.Query(ctx, `
		WITH followed AS (
			SELECT uf.asset_id, uf.tracking, uf.created_at
			FROM user_funds uf
			WHERE uf.user_id = $1
			UNION ALL
			-- A fund held without having been created here: followed by units.
			SELECT pe.asset_id, 'units'::fund_tracking, MIN(pe.created_at)
			FROM portfolio_entries pe
			JOIN portfolios p ON p.id = pe.portfolio_id
			JOIN assets a     ON a.id = pe.asset_id
			WHERE p.user_id = $1
			  AND a.asset_type = 'fund'
			  AND NOT EXISTS (SELECT 1 FROM user_funds uf WHERE uf.user_id = $1 AND uf.asset_id = pe.asset_id)
			GROUP BY pe.asset_id
		)
		SELECT a.id, a.ticker, a.name, a.currency, f.tracking, f.created_at,
		       lm.unit_value::text, lm.mark_date,
		       (SELECT COUNT(*) FROM fund_marks m WHERE m.user_id = $1 AND m.asset_id = f.asset_id)
		FROM followed f
		JOIN assets a ON a.id = f.asset_id
		LEFT JOIN LATERAL (
			SELECT m.unit_value, m.mark_date
			FROM fund_marks m
			WHERE m.user_id = $1 AND m.asset_id = f.asset_id
			ORDER BY m.mark_date DESC
			LIMIT 1
		) lm ON TRUE
		WHERE ($2::uuid IS NULL OR f.asset_id = $2)
		ORDER BY a.name, a.id
	`, userID, assetID)
	if err != nil {
		return nil, err
	}

	funds := make([]Fund, 0)
	index := make(map[uuid.UUID]int)

	for rows.Next() {
		var f Fund
		if err := rows.Scan(&f.AssetID, &f.Ticker, &f.Name, &f.Currency, &f.Tracking, &f.CreatedAt,
			&f.UnitValue, &f.ValuedOn, &f.Marks); err != nil {
			rows.Close()

			return nil, err
		}

		f.Positions = make([]FundPosition, 0)
		index[f.AssetID] = len(funds)
		funds = append(funds, f)
	}

	rows.Close()

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(funds) == 0 {
		return funds, nil
	}

	// What each position cost, in the fund's currency. A position opened here
	// costs in it already; one whose settlement was changed later is converted
	// with the same rate every valuation uses.
	posRows, err := q.Query(ctx, `
		SELECT pe.asset_id, pe.id, p.id, p.name, COALESCE(s.id, '00000000-0000-0000-0000-000000000000'::uuid), COALESCE(s.name, ''),
		       pe.quantity::text,
		       ROUND(pe.quantity * pe.price * COALESCE(fx_rate(p.user_id, pe.cost_currency, a.currency), 1), 8)::text
		FROM portfolio_entries pe
		JOIN portfolios p ON p.id = pe.portfolio_id
		JOIN assets a     ON a.id = pe.asset_id
		LEFT JOIN investment_sources s ON s.id = pe.source_id
		WHERE p.user_id = $1
		  AND a.asset_type = 'fund'
		  AND ($2::uuid IS NULL OR pe.asset_id = $2)
		  AND pe.quantity > 0
		ORDER BY p.name, s.name, pe.id
	`, userID, assetID)
	if err != nil {
		return nil, err
	}
	defer posRows.Close()

	for posRows.Next() {
		var (
			fundID uuid.UUID
			pos    FundPosition
		)

		if err := posRows.Scan(&fundID, &pos.EntryID, &pos.PortfolioID, &pos.PortfolioName,
			&pos.SourceID, &pos.SourceName, &pos.Units, &pos.Cost); err != nil {
			return nil, err
		}

		if i, ok := index[fundID]; ok {
			funds[i].Positions = append(funds[i].Positions, pos)
		}
	}

	if err := posRows.Err(); err != nil {
		return nil, err
	}

	for i := range funds {
		if err := funds[i].total(); err != nil {
			return nil, err
		}
	}

	return funds, nil
}

// total adds the positions up and values them at the latest mark, or at their
// cost while there is none.
func (f *Fund) total() error {
	var units, cost decimal.Decimal

	for _, pos := range f.Positions {
		u, err := decimal.NewFromString(pos.Units)
		if err != nil {
			return err
		}

		c, err := decimal.NewFromString(pos.Cost)
		if err != nil {
			return err
		}

		units, cost = units.Add(u), cost.Add(c)
	}

	f.Units, f.Cost = units.String(), cost.String()

	if f.UnitValue == nil {
		f.Value, f.PricedAtCost = cost.String(), true

		return nil
	}

	uv, err := decimal.NewFromString(*f.UnitValue)
	if err != nil {
		return err
	}

	f.Value = units.Mul(uv).RoundHAZ(8).String()

	return nil
}

// CreateFund records a fund whole: its asset, the row that says the owner
// follows it, the first purchase, and — when the owner knows it — what a unit
// is worth now.
//
// The asset is contributed (000021): created by this user and visible only to
// them, under a generated ticker. The current value is written before the
// purchase, and that order matters. The trigger of 000036 records what a
// purchase was worth when the app first saw it, reading the price the
// valuation uses at that moment; with the mark already in user_asset_prices, a
// fund bought months ago walks in at what it is worth, and the gain it made
// before anyone here saw it is not booked as the return of the day it was typed
// in.
func (r *PostgresRepository) CreateFund(ctx context.Context, userID uuid.UUID, in NewFundInput) (Fund, error) {
	var fund Fund

	if err := database.WithinTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		if err := requireCashAccount(ctx, tx, userID, in.PortfolioID, in.SourceID); err != nil {
			return err
		}

		var assetID uuid.UUID
		if err := tx.QueryRow(ctx, `
			INSERT INTO assets (ticker, name, asset_type, exchange, currency, created_by, is_curated, created_at, updated_at)
			VALUES ($1, $2, 'fund', NULL, $3::char(3), $4, FALSE, NOW(), NOW())
			RETURNING id
		`, newFundTicker(), in.CleanName(), in.Currency, userID).Scan(&assetID); err != nil {
			return err
		}

		if _, err := tx.Exec(ctx, `
			INSERT INTO user_catalog_assets (user_id, asset_id) VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, userID, assetID); err != nil {
			return err
		}

		if _, err := tx.Exec(ctx, `
			INSERT INTO user_funds (user_id, asset_id, tracking) VALUES ($1, $2, $3::fund_tracking)
		`, userID, assetID, in.Tracking); err != nil {
			return err
		}

		mark, err := in.openingMark()
		if err != nil {
			return err
		}

		if mark != nil {
			if _, err := writeFundMark(ctx, tx, userID, assetID, *mark); err != nil {
				return err
			}

			if err := syncFundPrice(ctx, tx, userID, assetID); err != nil {
				return err
			}
		}

		_, txnID, err := createPortfolioEntryTx(ctx, tx, userID, in.PortfolioID, assetID, in.SourceID, in.Currency, in.purchase())
		if err != nil {
			return err
		}

		// Followed by balance, the purchase is a contribution whose fact is its
		// money. The replay gives it the units it already has — its amount, at
		// the opening unit value of 1 — and checks the balance against them.
		if in.Tracking == FundBalance {
			if err := recordFundMovement(ctx, tx, txnID, in.Amount, false); err != nil {
				return err
			}

			if _, err := replayAndRestate(ctx, tx, userID, assetID, cashRateDay(in.Date)); err != nil {
				return err
			}
		}

		fund, err = readFund(ctx, tx, userID, assetID)

		return err
	}); err != nil {
		return Fund{}, err
	}

	return fund, nil
}

// DeleteFund stops following a fund that no portfolio holds any more: its
// marks go, and so does the price they put in user_asset_prices.
//
// A fund still held is refused. Its positions carry transactions, maybe paid
// from cash, and they are deleted where every position is, one at a time and
// with what each one undoes; doing it here in bulk would skip that.
//
// The asset row goes too when nothing else needs it — nobody else holds or
// lists it — which is the normal case for a fund created here. Otherwise it
// stays in the catalog and only this user's membership goes.
func (r *PostgresRepository) DeleteFund(ctx context.Context, userID, assetID uuid.UUID) error {
	return database.WithinTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		if _, err := lockFund(ctx, tx, userID, assetID); err != nil {
			return err
		}

		var held bool
		if err := tx.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM portfolio_entries pe
				JOIN portfolios p ON p.id = pe.portfolio_id
				WHERE p.user_id = $1 AND pe.asset_id = $2
			)
		`, userID, assetID).Scan(&held); err != nil {
			return err
		}

		if held {
			return ErrFundHasPositions
		}

		if _, err := tx.Exec(ctx, `
			DELETE FROM user_asset_prices WHERE user_id = $1 AND asset_id = $2 AND source = $3
		`, userID, assetID, fundPriceSource); err != nil {
			return err
		}

		if _, err := tx.Exec(ctx, `
			DELETE FROM user_funds WHERE user_id = $1 AND asset_id = $2
		`, userID, assetID); err != nil {
			return err
		}

		if _, err := tx.Exec(ctx, `
			DELETE FROM user_catalog_assets WHERE user_id = $1 AND asset_id = $2
		`, userID, assetID); err != nil {
			return err
		}

		_, err := tx.Exec(ctx, `
			DELETE FROM assets a
			WHERE a.id = $1
			  AND a.asset_type = 'fund'
			  AND NOT a.is_curated
			  AND NOT EXISTS (SELECT 1 FROM portfolio_entries pe WHERE pe.asset_id = a.id)
			  AND NOT EXISTS (SELECT 1 FROM user_catalog_assets uca WHERE uca.asset_id = a.id)
			  AND NOT EXISTS (SELECT 1 FROM user_funds uf WHERE uf.asset_id = a.id)
		`, assetID)

		return err
	})
}

// GetFundMarks lists a fund's marks, the most recent first.
func (r *PostgresRepository) GetFundMarks(ctx context.Context, userID, assetID uuid.UUID) ([]FundMark, error) {
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

	rows, err := r.db.Query(ctx, `
		SELECT `+fundMarkColumns+`
		FROM fund_marks
		WHERE user_id = $1 AND asset_id = $2
		ORDER BY mark_date DESC
	`, userID, assetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	marks := make([]FundMark, 0)

	for rows.Next() {
		mark, err := scanFundMark(rows)
		if err != nil {
			return nil, err
		}

		marks = append(marks, mark)
	}

	return marks, rows.Err()
}

// UpsertFundMark records what a fund was worth on a day, replacing the mark it
// had on that day. The price the valuation reads follows the latest mark, and
// every snapshot from the mark's day on is revalued with it (D9).
func (r *PostgresRepository) UpsertFundMark(ctx context.Context, userID, assetID uuid.UUID, in FundMarkInput) (FundMark, error) {
	var mark FundMark

	if err := database.WithinTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		tracking, err := lockFund(ctx, tx, userID, assetID)
		if err != nil {
			return err
		}

		day := cashRateDay(in.Date)

		if tracking == FundBalance {
			// The unit value is the replay's: written provisionally, and fixed
			// by it from the units held that day.
			in.UnitValue = fundOpeningUnitValue
		}

		if mark, err = writeFundMark(ctx, tx, userID, assetID, in); err != nil {
			return err
		}

		if tracking != FundBalance {
			if err := syncFundPrice(ctx, tx, userID, assetID); err != nil {
				return err
			}

			return restateFundSnapshots(ctx, tx, userID, assetID, day)
		}

		replay, err := replayAndRestate(ctx, tx, userID, assetID, day)
		if err != nil {
			return err
		}

		if replay.skipped(day) {
			return fmt.Errorf("%w: %s", ErrFundNoUnits, day.Format(time.DateOnly))
		}

		mark, err = scanFundMark(tx.QueryRow(ctx, `
			SELECT `+fundMarkColumns+` FROM fund_marks
			WHERE user_id = $1 AND asset_id = $2 AND mark_date = $3::date
		`, userID, assetID, day.Format(time.DateOnly)))

		return err
	}); err != nil {
		return FundMark{}, err
	}

	return mark, nil
}

// DeleteFundMark takes back what a fund was said to be worth on a day. The
// price falls back to the mark before it, or to the cost when none is left,
// and the snapshots from that day on are revalued the same way.
func (r *PostgresRepository) DeleteFundMark(ctx context.Context, userID, assetID uuid.UUID, date time.Time) error {
	day := cashRateDay(date)

	return database.WithinTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		tracking, err := lockFund(ctx, tx, userID, assetID)
		if err != nil {
			return err
		}

		tag, err := tx.Exec(ctx, `
			DELETE FROM fund_marks WHERE user_id = $1 AND asset_id = $2 AND mark_date = $3::date
		`, userID, assetID, day.Format(time.DateOnly))
		if err != nil {
			return err
		}

		if tag.RowsAffected() == 0 {
			return ErrFundMarkNotFound
		}

		if tracking == FundBalance {
			_, err := replayAndRestate(ctx, tx, userID, assetID, day)

			return err
		}

		if err := syncFundPrice(ctx, tx, userID, assetID); err != nil {
			return err
		}

		return restateFundSnapshots(ctx, tx, userID, assetID, day)
	})
}

// heldFundQuery selects a fund asset $2 that user $1 holds in some portfolio.
const heldFundQuery = `
	SELECT 1
	FROM portfolio_entries pe
	JOIN portfolios p ON p.id = pe.portfolio_id
	JOIN assets a     ON a.id = pe.asset_id
	WHERE p.user_id = $1 AND pe.asset_id = $2 AND a.asset_type = 'fund'`

// adoptHeldFund records that the owner follows, by units, a fund they hold but
// never created here. It does nothing for a fund already followed, or for an
// asset that is not a fund they hold — lockFund then answers not found.
func adoptHeldFund(ctx context.Context, tx pgx.Tx, userID, assetID uuid.UUID) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO user_funds (user_id, asset_id, tracking)
		SELECT $1, $2, 'units'
		WHERE EXISTS (`+heldFundQuery+`)
		ON CONFLICT (user_id, asset_id) DO NOTHING
	`, userID, assetID)

	return err
}

// lockFund takes the fund's user_funds row, which is what serialises the
// writes to its marks, and answers with how it is followed.
func lockFund(ctx context.Context, tx pgx.Tx, userID, assetID uuid.UUID) (FundTracking, error) {
	if err := adoptHeldFund(ctx, tx, userID, assetID); err != nil {
		return "", err
	}

	var tracking FundTracking

	err := tx.QueryRow(ctx, `
		SELECT tracking FROM user_funds WHERE user_id = $1 AND asset_id = $2 FOR UPDATE
	`, userID, assetID).Scan(&tracking)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrFundNotFound
	}

	return tracking, err
}

const fundMarkColumns = `mark_date, unit_value::text, balance::text, COALESCE(notes, ''), created_at, updated_at`

func scanFundMark(row pgx.Row) (FundMark, error) {
	var m FundMark

	err := row.Scan(&m.Date, &m.UnitValue, &m.Balance, &m.Notes, &m.CreatedAt, &m.UpdatedAt)

	return m, err
}

// writeFundMark writes one mark, replacing the one on its day.
func writeFundMark(ctx context.Context, tx pgx.Tx, userID, assetID uuid.UUID, in FundMarkInput) (FundMark, error) {
	var balance *string
	if in.Balance.IsPos() {
		b := in.Balance.String()
		balance = &b
	}

	return scanFundMark(tx.QueryRow(ctx, `
		INSERT INTO fund_marks (user_id, asset_id, mark_date, unit_value, balance, notes)
		VALUES ($1, $2, $3::date, $4::numeric, $5::numeric, NULLIF($6, ''))
		ON CONFLICT (user_id, asset_id, mark_date) DO UPDATE SET
			unit_value = EXCLUDED.unit_value,
			balance    = EXCLUDED.balance,
			notes      = EXCLUDED.notes,
			updated_at = NOW()
		RETURNING `+fundMarkColumns,
		userID, assetID, cashRateDay(in.Date).Format(time.DateOnly), in.UnitValue.String(), balance, in.Notes))
}

// syncFundPrice copies the latest mark to user_asset_prices, where the summary,
// the holdings and the snapshot job look first, or removes the copy when no
// mark is left. fetched_at is the mark's day, not the write's: it is when the
// price was true.
//
// Only a price this code wrote is removed. The market sync never prices a
// fund, but a row some other writer left is not this code's to throw away.
func syncFundPrice(ctx context.Context, tx pgx.Tx, userID, assetID uuid.UUID) error {
	var (
		unitValue *string
		markDate  *time.Time
		cur       money.Currency
	)

	if err := tx.QueryRow(ctx, `
		SELECT a.currency, lm.unit_value::text, lm.mark_date
		FROM assets a
		LEFT JOIN LATERAL (
			SELECT m.unit_value, m.mark_date
			FROM fund_marks m
			WHERE m.user_id = $1 AND m.asset_id = a.id
			ORDER BY m.mark_date DESC
			LIMIT 1
		) lm ON TRUE
		WHERE a.id = $2
	`, userID, assetID).Scan(&cur, &unitValue, &markDate); err != nil {
		return err
	}

	if unitValue == nil {
		_, err := tx.Exec(ctx, `
			DELETE FROM user_asset_prices WHERE user_id = $1 AND asset_id = $2 AND source = $3
		`, userID, assetID, fundPriceSource)

		return err
	}

	_, err := tx.Exec(ctx, `
		INSERT INTO user_asset_prices (user_id, asset_id, price, currency, source, fetched_at)
		VALUES ($1, $2, $3::numeric, $4, $5, $6::date)
		ON CONFLICT (user_id, asset_id) DO UPDATE SET
			price      = EXCLUDED.price,
			currency   = EXCLUDED.currency,
			source     = EXCLUDED.source,
			fetched_at = EXCLUDED.fetched_at
	`, userID, assetID, *unitValue, cur, fundPriceSource, markDate.Format(time.DateOnly))

	return err
}

// restateFundSnapshots revalues the snapshots that priced a fund on or after
// from, now that its marks changed (D9 of docs/PLAN_FONDOS_INVERSION.md).
//
// Each snapshot recorded the units, the unit value and the rate it used for
// every fund position (fund_snapshot_values), so the correction is exact: the
// unit value it should have used — the latest mark on or before its day, or the
// position's cost that day when there is none — against the one it did, times
// the units and the rate it did use. The difference moves the snapshot's total,
// its gain and the fund slice of its allocation, and the percentage is
// recomputed over the cost, which a mark does not move. The flows are not
// touched: a mark is not money coming in or going out.
//
// Rows whose unit value already agrees are left alone, so a mark that changes
// nothing writes nothing.
func restateFundSnapshots(ctx context.Context, tx pgx.Tx, userID, assetID uuid.UUID, from time.Time) error {
	_, err := tx.Exec(ctx, `
		WITH due AS (
			SELECT f.entry_id, f.snapshot_date, f.portfolio_id, f.units, f.fx_rate,
			       f.unit_value AS old_value,
			       COALESCE((
			           SELECT m.unit_value
			           FROM fund_marks m
			           WHERE m.user_id = $1 AND m.asset_id = f.asset_id AND m.mark_date <= f.snapshot_date
			           ORDER BY m.mark_date DESC
			           LIMIT 1
			       ), f.unit_cost) AS new_value
			FROM fund_snapshot_values f
			JOIN portfolios p ON p.id = f.portfolio_id AND p.user_id = $1
			WHERE f.asset_id = $2 AND f.snapshot_date >= $3::date
		), changed AS (
			UPDATE fund_snapshot_values f
			   SET unit_value = due.new_value
			  FROM due
			 WHERE f.entry_id = due.entry_id
			   AND f.snapshot_date = due.snapshot_date
			   AND due.new_value <> due.old_value
			RETURNING due.portfolio_id, due.snapshot_date,
			          ROUND(due.units * (due.new_value - due.old_value) * due.fx_rate, 8) AS delta
		), per_snapshot AS (
			SELECT portfolio_id, snapshot_date, SUM(delta) AS delta
			FROM changed
			GROUP BY portfolio_id, snapshot_date
		)
		UPDATE portfolio_snapshots ps
		   SET total_value         = ps.total_value + d.delta,
		       total_gain_loss     = ps.total_gain_loss + d.delta,
		       total_gain_loss_pct = CASE
		           WHEN ps.total_value - ps.total_gain_loss > 0
		           THEN (ps.total_gain_loss + d.delta) / (ps.total_value - ps.total_gain_loss) * 100
		           ELSE 0
		       END,
		       allocation          = jsonb_set(
		           ps.allocation,
		           '{fund}',
		           to_jsonb((COALESCE((ps.allocation ->> 'fund')::numeric, 0) + d.delta)::text)
		       )
		  FROM per_snapshot d
		 WHERE ps.portfolio_id = d.portfolio_id
		   AND ps.snapshot_date = d.snapshot_date
	`, userID, assetID, from.Format(time.DateOnly))

	return err
}
