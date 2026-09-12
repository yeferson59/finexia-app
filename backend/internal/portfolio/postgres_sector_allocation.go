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
// What it adds is the axis itself, and the axis has two complications the other
// two do not.
//
// The first is that not every asset has a sector, and "has none" comes in two
// kinds that must not be added together.
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
// The second is the one that makes this query unlike its two siblings: an asset
// can belong to several sectors at once. A whole-market ETF is every industry
// there is, in the proportions its fact sheet publishes, so a position in VOO is
// not one slice of this chart but eleven — and the question the chart exists to
// answer ("how much of my money rides on semiconductors?") has the wrong answer
// without it, since the fund is where most people's technology exposure
// actually is. Those assets carry a market.SectorBreakdown and the position is
// split across it, each sector taking its share of the same market value the
// other two views total whole. This is the one place in the app where a
// position is divided rather than filed.
//
// The weights need not add up to a hundred — a fact sheet's own leave a little
// in cash — so each share is the weight over the total of that asset's weights,
// not over 100. Normalising is what keeps this chart adding up to the same
// number the other two do: a breakdown covering 97.3 % of a fund still accounts
// for 100 % of the money in it, spread in the proportions that are known. The
// alternative, sending the missing 2.7 % to a bucket, would put a rounding
// artefact on screen beside real industries.
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
	// Assets is a count per row, so a fund spread over eleven sectors is one
	// asset in each of them and the counts add up to more than the portfolio
	// holds. That is the honest reading of the question the column answers —
	// "how many different things put me in this industry" — and the alternative,
	// a fraction of an asset, is not a number anybody can act on.
	//
	// Rounded to eight decimals for the reason the sibling queries are: Postgres
	// adds the scales of what it multiplies, and the decimal engine that parses
	// this text back caps at nineteen — past it a row's value read as zero and
	// its share silently collapsed.
	rows, err := r.db.Query(ctx, `
		SELECT
			sw.sector,
			ROUND(COALESCE(SUM(pe.quantity::numeric * v.price * COALESCE(fx.rate, 1) * sw.share), 0), 8)::text AS market_value,
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
		-- One row per sector this position counts towards, and the fraction of
		-- it that goes there. Exactly one of the two branches produces rows for
		-- any asset, so every position is still accounted for once in total —
		-- just not necessarily in one place.
		CROSS JOIN LATERAL (
			-- The asset is made of several industries: one row each, the weight
			-- over the total of the weights. A LATERAL and not a join, so a fund
			-- with no breakdown produces nothing here and falls through to the
			-- branch below.
			SELECT w.sector::text AS sector, w.weight / whole.total AS share
			FROM asset_sector_weights w
			CROSS JOIN LATERAL (
				SELECT SUM(w2.weight) AS total
				FROM asset_sector_weights w2
				WHERE w2.asset_id = a.id
			) whole
			WHERE w.asset_id = a.id
			  -- Unreachable given the column's CHECK (weight > 0), so this
			  -- guards the division and nothing else: a breakdown that somehow
			  -- summed to zero would take the position out of the chart
			  -- entirely, and a division error is the louder failure.
			  AND whole.total > 0

			UNION ALL

			-- The ordinary asset: one sector, the whole position. The CASE is
			-- the same one that has always bucketed this chart, and the NOT
			-- EXISTS is what keeps an asset from being counted by both branches.
			SELECT
				CASE
					WHEN a.sector IS NOT NULL AND a.sector <> '' THEN a.sector
					WHEN a.asset_type IN ('stock', 'etf', 'bond', 'other') THEN $3::text
					ELSE $4::text
				END,
				1
			WHERE NOT EXISTS (
				SELECT 1 FROM asset_sector_weights w3 WHERE w3.asset_id = a.id
			)
		) sw
		WHERE p.user_id = $1
		  AND pe.quantity::numeric > 0
		GROUP BY sw.sector, target.code
		ORDER BY ROUND(COALESCE(SUM(pe.quantity::numeric * v.price * COALESCE(fx.rate, 1) * sw.share), 0), 8) DESC, sw.sector
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
