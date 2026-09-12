package portfolio

import (
	"context"

	"uuid"

	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/market"
)

// GetSectorAllocationByUserID totals what the user holds per industry, in one
// currency, across every portfolio they own.
//
// It is the third reading of the same positions, beside
// GetAssetAllocationByUserID (by asset type) and GetAssetHoldingsByUserID (by
// asset), and it deliberately shares their valuation rules to the letter: the
// user's own price, then the operator's manual one, then the entry's cost; the
// currency chosen with the price; conversion through fx_rate, with a position
// that has no rate still counted at face value and reported in
// PositionsUnconverted. Three charts over one set of positions have to be able
// to disagree about nothing but the axis they group on.
//
// What it adds is the axis itself, and the axis has a complication the other
// two do not: not every asset has a sector, and "has none" comes in two kinds
// that must not be added together.
//
//   - An asset whose type cannot have one — a coin, a cash balance, a flat, a
//     bar of gold — is market.SectorNotApplicable. Nothing is missing; there is
//     nothing to fill in.
//   - An asset whose type can have one and whose catalog row does not is
//     market.SectorUnclassified. That is work to do, and the size of it is
//     exactly what the user needs to see: a portfolio that is 60 % unclassified
//     has no sector answer yet, and a chart that hid those rows would claim it
//     did.
//
// Neither bucket is ever stored. The CASE below derives them at read time from
// assets.sector and assets.asset_type, which is what keeps one asset's answer
// the same everywhere — reclassifying a row moves it in this chart with no
// second column to keep in step.
//
// Rows the user holds are counted whole, including the unclassified ones, so
// the shares always add up to the portfolio. targetCurrency empty means the
// user's stored preference, same contract as the other two.
func (r *PostgresRepository) GetSectorAllocationByUserID(ctx context.Context, userID uuid.UUID, targetCurrency money.Currency) ([]SectorAllocationItem, error) {
	// The classifiable types are spelled out here rather than derived from
	// market.AssetType.HasSector because this runs in Postgres. The Go method is
	// the one that decides what may be written; this list decides how what is
	// there is read, and asset_service.go's validation is what keeps a value
	// from existing that only one of the two would accept.
	//
	// Rounded to eight decimals for the reason the sibling queries are: Postgres
	// adds the scales of what it multiplies, and the decimal engine that parses
	// this text back caps at nineteen — past it a row's value read as zero and
	// its share silently collapsed.
	rows, err := r.db.Query(ctx, `
		SELECT
			CASE
				WHEN a.sector IS NOT NULL AND a.sector <> '' THEN a.sector
				WHEN a.asset_type IN ('stock', 'etf', 'bond', 'other') THEN $3::text
				ELSE $4::text
			END AS sector,
			ROUND(COALESCE(SUM(pe.quantity::numeric * v.price * COALESCE(fx.rate, 1)), 0), 8)::text AS market_value,
			target.code,
			COUNT(DISTINCT pe.asset_id)::bigint AS assets,
			COUNT(DISTINCT pe.asset_id) FILTER (WHERE fx.rate IS NULL)::bigint AS positions_unconverted
		FROM portfolio_entries pe
		JOIN portfolios p ON p.id = pe.portfolio_id
		JOIN users u      ON u.id = p.user_id
		JOIN assets a     ON a.id = pe.asset_id
		LEFT JOIN user_asset_prices uap ON uap.asset_id = a.id AND uap.user_id = p.user_id
		CROSS JOIN LATERAL (
			SELECT COALESCE(NULLIF($2::text, ''), u.preferred_currency) AS code
		) target
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
		GROUP BY 1, target.code
		ORDER BY ROUND(COALESCE(SUM(pe.quantity::numeric * v.price * COALESCE(fx.rate, 1)), 0), 8) DESC, 1
	`, userID, currencyParam(targetCurrency), market.SectorUnclassified, market.SectorNotApplicable)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]SectorAllocationItem, 0)
	for rows.Next() {
		var item SectorAllocationItem

		if err := rows.Scan(
			&item.Sector,
			&item.MarketValue,
			&item.Currency,
			&item.Assets,
			&item.PositionsUnconverted,
		); err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, rows.Err()
}
