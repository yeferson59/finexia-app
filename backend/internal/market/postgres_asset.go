package market

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/yeferson59/gofinance/v2/money"
)

func (r *PostgresRepository) GetAssetByID(ctx context.Context, assetID uuid.UUID) (Asset, error) {
	var asset Asset
	var priceStr *string
	err := r.db.QueryRow(ctx, `
		SELECT id, ticker, name, asset_type, COALESCE(exchange, ''), currency, current_price, price_updated_at, created_at, updated_at
		FROM assets WHERE id = $1
	`, assetID).Scan(
		&asset.ID, &asset.Ticker, &asset.Name, &asset.AssetType, &asset.Exchange,
		&asset.Currency, &priceStr, &asset.PriceUpdatedAt, &asset.CreatedAt, &asset.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Asset{}, errors.New("asset not found")
		}
		return Asset{}, err
	}
	scanAssetCurrentPrice(&asset, priceStr)
	return asset, nil
}

func (r *PostgresRepository) GetAssets(ctx context.Context, offset, limit uint) ([]Asset, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, ticker, name, asset_type, COALESCE(exchange, ''), currency, current_price, price_updated_at, created_at, updated_at
		FROM assets
		ORDER BY ticker
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return []Asset{}, err
	}
	defer rows.Close()

	assets := make([]Asset, 0)

	for rows.Next() {
		var asset Asset
		var priceStr *string

		if err := rows.Scan(
			&asset.ID,
			&asset.Ticker,
			&asset.Name,
			&asset.AssetType,
			&asset.Exchange,
			&asset.Currency,
			&priceStr,
			&asset.PriceUpdatedAt,
			&asset.CreatedAt,
			&asset.UpdatedAt,
		); err != nil {
			return nil, err
		}

		scanAssetCurrentPrice(&asset, priceStr)
		assets = append(assets, asset)
	}

	return assets, nil
}

func (r *PostgresRepository) UpdateAssetPrice(ctx context.Context, assetID uuid.UUID, price money.Money) (Asset, error) {
	var asset Asset
	var priceStr *string

	err := r.db.QueryRow(ctx, `
		UPDATE assets
		SET current_price = $2::numeric, price_updated_at = NOW()
		WHERE id = $1
		RETURNING id, ticker, name, asset_type, COALESCE(exchange, ''), currency, current_price, price_updated_at, created_at, updated_at
	`, assetID, price.String()).Scan(
		&asset.ID,
		&asset.Ticker,
		&asset.Name,
		&asset.AssetType,
		&asset.Exchange,
		&asset.Currency,
		&priceStr,
		&asset.PriceUpdatedAt,
		&asset.CreatedAt,
		&asset.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Asset{}, errors.New("asset not found")
		}
		return Asset{}, err
	}

	scanAssetCurrentPrice(&asset, priceStr)
	return asset, nil
}

func (r *PostgresRepository) SearchAssets(ctx context.Context, search string, offset, limit uint) ([]Asset, error) {
	pattern := "%" + strings.ToUpper(strings.TrimSpace(search)) + "%"
	rows, err := r.db.Query(ctx, `
		SELECT id, ticker, name, asset_type, COALESCE(exchange, ''), currency, current_price, price_updated_at, created_at, updated_at
		FROM assets
		WHERE UPPER(ticker) LIKE $1 OR UPPER(name) LIKE $1
		ORDER BY ticker
		LIMIT $2 OFFSET $3
	`, pattern, limit, offset)
	if err != nil {
		return []Asset{}, err
	}
	defer rows.Close()

	assets := make([]Asset, 0)
	for rows.Next() {
		var asset Asset
		var priceStr *string
		if err := rows.Scan(
			&asset.ID, &asset.Ticker, &asset.Name, &asset.AssetType, &asset.Exchange,
			&asset.Currency, &priceStr, &asset.PriceUpdatedAt, &asset.CreatedAt, &asset.UpdatedAt,
		); err != nil {
			return nil, err
		}
		scanAssetCurrentPrice(&asset, priceStr)
		assets = append(assets, asset)
	}
	return assets, nil
}

func (r *PostgresRepository) UpsertAsset(ctx context.Context, ticker, name string, assetType AssetType, exchange, currency string) (Asset, error) {
	var asset Asset
	var priceStr *string
	err := r.db.QueryRow(ctx, `
		INSERT INTO assets (ticker, name, asset_type, exchange, currency, created_at, updated_at)
		VALUES ($1, $2, $3::asset_type, NULLIF($4, ''), $5, NOW(), NOW())
		ON CONFLICT (ticker, COALESCE(exchange, ''))
		DO UPDATE SET name = EXCLUDED.name, asset_type = EXCLUDED.asset_type, currency = EXCLUDED.currency, updated_at = NOW()
		RETURNING id, ticker, name, asset_type, COALESCE(exchange, ''), currency, current_price, price_updated_at, created_at, updated_at
	`, ticker, name, assetType, exchange, currency).Scan(
		&asset.ID,
		&asset.Ticker,
		&asset.Name,
		&asset.AssetType,
		&asset.Exchange,
		&asset.Currency,
		&priceStr,
		&asset.PriceUpdatedAt,
		&asset.CreatedAt,
		&asset.UpdatedAt,
	)
	if err != nil {
		return asset, err
	}
	scanAssetCurrentPrice(&asset, priceStr)
	return asset, nil
}
