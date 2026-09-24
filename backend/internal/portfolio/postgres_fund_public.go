package portfolio

import (
	"context"
	"errors"
	"time"

	"uuid"

	"github.com/jackc/pgx/v5"

	"github.com/yeferson59/finexia-app/internal/platform/database"
)

// The SFC's catalog of funds and the links to it (000059). A published value
// is written as a mark of every fund linked to it, under the lock every write
// to a fund's marks takes (lockFund), so it lands the way a mark the owner
// typed does: the price follows the latest mark and the snapshots from its day
// on are revalued.

// UpsertPublicFunds writes the catalog of the latest day the SFC published. A
// fund missing from it keeps its row, with the last value it had: one that
// stopped publishing still names what some owner linked to.
func (r *PostgresRepository) UpsertPublicFunds(ctx context.Context, funds []PublicFund) (int, error) {
	if len(funds) == 0 {
		return 0, nil
	}

	n := len(funds)
	ids, entities, names, kinds, search, values, dates := make([]string, n), make([]string, n), make([]string, n),
		make([]string, n), make([]string, n), make([]string, n), make([]string, n)
	entityTypes, entityCodes, fundCodes, compartments, participations, investors := make([]int32, n), make([]int32, n),
		make([]int32, n), make([]int32, n), make([]int32, n), make([]int32, n)

	for i, f := range funds {
		ids[i], entities[i], names[i], kinds[i] = f.ID, f.EntityName, f.FundName, f.FundKind
		search[i], values[i], dates[i] = f.searchText(), f.UnitValue, f.ValueDate.Format(time.DateOnly)
		entityTypes[i], entityCodes[i], fundCodes[i] = int32(f.key.EntityType), int32(f.key.Entity), int32(f.key.Fund)
		compartments[i], participations[i], investors[i] = int32(f.key.Compartment), int32(f.key.Participation), int32(f.Investors)
	}

	tag, err := r.db.Exec(ctx, `
		INSERT INTO public_funds (id, entity_type, entity_code, fund_code, compartment, participation,
		                          entity_name, fund_name, fund_kind, search_text, unit_value, value_date, investors)
		SELECT u.id, u.entity_type, u.entity_code, u.fund_code, u.compartment, u.participation,
		       LEFT(u.entity_name, 255), LEFT(u.fund_name, 255), LEFT(u.fund_kind, 255), u.search_text,
		       u.unit_value::numeric, u.value_date::date, u.investors
		FROM unnest($1::text[], $2::int[], $3::int[], $4::int[], $5::int[], $6::int[],
		            $7::text[], $8::text[], $9::text[], $10::text[], $11::text[], $12::text[], $13::int[])
		     AS u(id, entity_type, entity_code, fund_code, compartment, participation,
		          entity_name, fund_name, fund_kind, search_text, unit_value, value_date, investors)
		ON CONFLICT (id) DO UPDATE SET
			entity_name = EXCLUDED.entity_name,
			fund_name   = EXCLUDED.fund_name,
			fund_kind   = EXCLUDED.fund_kind,
			search_text = EXCLUDED.search_text,
			unit_value  = EXCLUDED.unit_value,
			value_date  = EXCLUDED.value_date,
			investors   = EXCLUDED.investors,
			updated_at  = NOW()
		WHERE public_funds.value_date <= EXCLUDED.value_date
	`, ids, entityTypes, entityCodes, fundCodes, compartments, participations,
		entities, names, kinds, search, values, dates, investors)
	if err != nil {
		return 0, err
	}

	return int(tag.RowsAffected()), nil
}

// CountPublicFunds is how many funds the catalog has.
func (r *PostgresRepository) CountPublicFunds(ctx context.Context) (int, error) {
	var n int

	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM public_funds`).Scan(&n)

	return n, err
}

const publicFundColumns = `id, entity_type, entity_code, fund_code, compartment, participation,
	entity_name, fund_name, fund_kind, unit_value::text, value_date, investors`

func scanPublicFund(row pgx.Row) (PublicFund, error) {
	var f PublicFund

	err := row.Scan(&f.ID, &f.key.EntityType, &f.key.Entity, &f.key.Fund, &f.key.Compartment, &f.key.Participation,
		&f.EntityName, &f.FundName, &f.FundKind, &f.UnitValue, &f.ValueDate, &f.Investors)
	f.FundCode, f.Participation = f.key.Fund, f.key.Participation

	return f, err
}

// SearchPublicFunds finds the funds whose names or codes hold every word,
// published on or after since, the most held first: the fund an owner is
// looking for is far more often a large one than a closed fund of eleven
// investors.
func (r *PostgresRepository) SearchPublicFunds(ctx context.Context, words []string, since time.Time, limit int) ([]PublicFund, error) {
	patterns := make([]string, len(words))
	for i, w := range words {
		patterns[i] = "%" + w + "%"
	}

	rows, err := r.db.Query(ctx, `
		SELECT `+publicFundColumns+`
		FROM public_funds
		WHERE search_text LIKE ALL($1::text[])
		  AND value_date >= $2::date
		ORDER BY investors DESC, fund_name, participation
		LIMIT $3
	`, patterns, since.Format(time.DateOnly), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	funds := make([]PublicFund, 0)

	for rows.Next() {
		f, err := scanPublicFund(rows)
		if err != nil {
			return nil, err
		}

		funds = append(funds, f)
	}

	return funds, rows.Err()
}

// GetPublicFund reads one fund of the catalog.
func (r *PostgresRepository) GetPublicFund(ctx context.Context, id string) (PublicFund, error) {
	f, err := scanPublicFund(r.db.QueryRow(ctx, `SELECT `+publicFundColumns+` FROM public_funds WHERE id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return PublicFund{}, ErrPublicFundNotFound
	}

	return f, err
}

// LinkFund links a fund followed by units to a published one and writes the
// values given as its marks. A link to another published fund is replaced, and
// the marks it brought go with it: they were the unit values of some other
// fund.
func (r *PostgresRepository) LinkFund(ctx context.Context, userID, assetID uuid.UUID, publicID string, values []PublicFundValue) (Fund, error) {
	var fund Fund

	if err := database.WithinTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		tracking, err := lockFund(ctx, tx, userID, assetID)
		if err != nil {
			return err
		}

		if tracking != FundUnits {
			return ErrFundNotLinkable
		}

		var current *string
		if err := tx.QueryRow(ctx, `
			SELECT public_fund_id FROM user_funds WHERE user_id = $1 AND asset_id = $2
		`, userID, assetID).Scan(&current); err != nil {
			return err
		}

		var from time.Time

		if current != nil && *current != publicID {
			if from, err = deletePublicMarks(ctx, tx, userID, assetID); err != nil {
				return err
			}
		}

		if _, err := tx.Exec(ctx, `
			UPDATE user_funds SET public_fund_id = $3, updated_at = NOW()
			WHERE user_id = $1 AND asset_id = $2
		`, userID, assetID, publicID); err != nil {
			return err
		}

		written, err := writePublicMarks(ctx, tx, userID, assetID, values)
		if err != nil {
			return err
		}

		if err := settleFundMarks(ctx, tx, userID, assetID, earliest(from, written)); err != nil {
			return err
		}

		fund, err = readFund(ctx, tx, userID, assetID)

		return err
	}); err != nil {
		return Fund{}, err
	}

	return fund, nil
}

// UnlinkFund ends a fund's link and takes back every mark it brought. The
// owner's own marks stay; the price and the snapshots fall back to them, or to
// the cost when there are none.
func (r *PostgresRepository) UnlinkFund(ctx context.Context, userID, assetID uuid.UUID) (Fund, error) {
	var fund Fund

	if err := database.WithinTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		if _, err := lockFund(ctx, tx, userID, assetID); err != nil {
			return err
		}

		if _, err := tx.Exec(ctx, `
			UPDATE user_funds SET public_fund_id = NULL, updated_at = NOW()
			WHERE user_id = $1 AND asset_id = $2 AND public_fund_id IS NOT NULL
		`, userID, assetID); err != nil {
			return err
		}

		from, err := deletePublicMarks(ctx, tx, userID, assetID)
		if err != nil {
			return err
		}

		if err := settleFundMarks(ctx, tx, userID, assetID, from); err != nil {
			return err
		}

		fund, err = readFund(ctx, tx, userID, assetID)

		return err
	}); err != nil {
		return Fund{}, err
	}

	return fund, nil
}

// GetLinkedFunds lists every fund linked to a published one, with the day its
// next import starts: the day after its latest published mark, or its first
// purchase before it has any, or a year ago for a fund it holds no purchase of.
func (r *PostgresRepository) GetLinkedFunds(ctx context.Context) ([]LinkedFund, error) {
	rows, err := r.db.Query(ctx, `
		SELECT uf.user_id, uf.asset_id, uf.public_fund_id,
		       COALESCE(
		           (SELECT MAX(m.mark_date) + 1 FROM fund_marks m
		            WHERE m.user_id = uf.user_id AND m.asset_id = uf.asset_id AND m.source = 'public'),
		           (SELECT MIN(t.transaction_date)::date FROM transactions t
		            JOIN portfolio_entries pe ON pe.id = t.entry_id
		            JOIN portfolios p         ON p.id = pe.portfolio_id
		            WHERE p.user_id = uf.user_id AND pe.asset_id = uf.asset_id),
		           CURRENT_DATE - $1::int
		       )
		FROM user_funds uf
		WHERE uf.public_fund_id IS NOT NULL
		ORDER BY uf.public_fund_id, uf.user_id, uf.asset_id
	`, publicFundBackfillDays)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	linked := make([]LinkedFund, 0)

	for rows.Next() {
		var l LinkedFund
		if err := rows.Scan(&l.UserID, &l.AssetID, &l.PublicFundID, &l.Since); err != nil {
			return nil, err
		}

		linked = append(linked, l)
	}

	return linked, rows.Err()
}

// ImportPublicMarks writes published values as marks of one linked fund. It
// writes nothing when the fund is no longer linked to publicID — the owner
// unlinked it, or linked it elsewhere, while the values were being read — and
// answers how many marks it wrote or changed.
func (r *PostgresRepository) ImportPublicMarks(ctx context.Context, userID, assetID uuid.UUID, publicID string, values []PublicFundValue) (int, error) {
	if len(values) == 0 {
		return 0, nil
	}

	var n int

	err := database.WithinTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		if _, err := lockFund(ctx, tx, userID, assetID); err != nil {
			return err
		}

		var linked bool
		if err := tx.QueryRow(ctx, `
			SELECT EXISTS (SELECT 1 FROM user_funds WHERE user_id = $1 AND asset_id = $2 AND public_fund_id = $3)
		`, userID, assetID, publicID).Scan(&linked); err != nil || !linked {
			return err
		}

		written, err := writePublicMarks(ctx, tx, userID, assetID, values)
		if err != nil {
			return err
		}

		n = len(written)

		return settleFundMarks(ctx, tx, userID, assetID, earliest(time.Time{}, written))
	})

	return n, err
}

// writePublicMarks writes published values as marks, and answers the days it
// wrote or changed. A day the owner marked is left alone, and so is one whose
// published value is already there: a run that brings nothing new writes
// nothing.
func writePublicMarks(ctx context.Context, tx pgx.Tx, userID, assetID uuid.UUID, values []PublicFundValue) ([]time.Time, error) {
	if len(values) == 0 {
		return nil, nil
	}

	dates, unitValues := make([]string, len(values)), make([]string, len(values))
	for i, v := range values {
		dates[i], unitValues[i] = cashRateDay(v.Date).Format(time.DateOnly), v.UnitValue.String()
	}

	rows, err := tx.Query(ctx, `
		INSERT INTO fund_marks (user_id, asset_id, mark_date, unit_value, source)
		SELECT $1, $2, u.day::date, u.value::numeric, 'public'
		FROM unnest($3::text[], $4::text[]) AS u(day, value)
		ON CONFLICT (user_id, asset_id, mark_date) DO UPDATE SET
			unit_value = EXCLUDED.unit_value,
			updated_at = NOW()
		WHERE fund_marks.source = 'public'
		  AND fund_marks.unit_value <> EXCLUDED.unit_value
		RETURNING mark_date
	`, userID, assetID, dates, unitValues)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, pgx.RowTo[time.Time])
}

// deletePublicMarks takes back every published mark of a fund, and answers the
// earliest day it took, zero when there was none.
func deletePublicMarks(ctx context.Context, tx pgx.Tx, userID, assetID uuid.UUID) (time.Time, error) {
	var from *time.Time

	err := tx.QueryRow(ctx, `
		WITH gone AS (
			DELETE FROM fund_marks
			WHERE user_id = $1 AND asset_id = $2 AND source = 'public'
			RETURNING mark_date
		)
		SELECT MIN(mark_date) FROM gone
	`, userID, assetID).Scan(&from)
	if err != nil || from == nil {
		return time.Time{}, err
	}

	return *from, nil
}

// settleFundMarks brings the price and the snapshots of a fund followed by
// units in line with its marks after they changed from a day on. A zero day
// means nothing changed.
func settleFundMarks(ctx context.Context, tx pgx.Tx, userID, assetID uuid.UUID, from time.Time) error {
	if from.IsZero() {
		return nil
	}

	if err := syncFundPrice(ctx, tx, userID, assetID); err != nil {
		return err
	}

	return restateFundSnapshots(ctx, tx, userID, assetID, from)
}

// earliest is the first of a day and some more, ignoring a zero day.
func earliest(day time.Time, more []time.Time) time.Time {
	for _, d := range more {
		if day.IsZero() || d.Before(day) {
			day = d
		}
	}

	return day
}
