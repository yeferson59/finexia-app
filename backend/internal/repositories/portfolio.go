package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/yeferson59/finexia-app/internal/entities"
	"github.com/yeferson59/gofinance/money"
)

func (r *Repository) GetPortfoliosByUserID(ctx context.Context, userID uuid.UUID) ([]entities.Portfolio, error) {
	rows, err := r.db.Query(ctx, "SELECT p.*, r.* FROM portfolios p JOIN risks r ON p.risk_id = r.id WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	portfolios := make([]entities.Portfolio, 0)
	for rows.Next() {
		var portfolio entities.Portfolio

		if err := rows.Scan(&portfolio.ID, &portfolio.UserID, &portfolio.Name, &portfolio.Description, &portfolio.Type, &portfolio.RiskID, &portfolio.BaseCurrency, &portfolio.IsDefault, &portfolio.PriceValue, &portfolio.CreatedAt, &portfolio.UpdatedAt, &portfolio.Risk.ID, &portfolio.Risk.Name, &portfolio.Risk.Description, &portfolio.Risk.CreatedAt, &portfolio.Risk.UpdatedAt); err != nil {
			return nil, err
		}

		portfolios = append(portfolios, portfolio)
	}

	return portfolios, nil
}

func (r *Repository) CreatePortfolio(ctx context.Context, userID uuid.UUID, name, description, baseCurrency string, riskID uuid.UUID, typePortfolio entities.PortfolioType, price_value money.Money, isDefault bool) (entities.Portfolio, error) {
	var portfolio entities.Portfolio
	if err := r.db.QueryRow(ctx, "INSERT INTO portfolios(user_id, name, description, base_currency, type, risk_id, price_value, is_default) VALUES($1, $2, $3, $4, $5, $6, $7, $8) RETURNING *", userID, name, description, baseCurrency, typePortfolio, riskID, price_value, isDefault).Scan(
		&portfolio.ID,
		&portfolio.UserID,
		&portfolio.Name,
		&portfolio.Description,
		&portfolio.Type,
		&portfolio.RiskID,
		&portfolio.BaseCurrency,
		&portfolio.IsDefault,
		&portfolio.PriceValue,
		&portfolio.CreatedAt,
		&portfolio.UpdatedAt,
	); err != nil {
		return entities.Portfolio{}, err
	}

	return portfolio, nil
}

func (r *Repository) GetPortfolioByID(ctx context.Context, portfolioID, userID uuid.UUID) (entities.Portfolio, error) {
	var portfolio entities.Portfolio
	err := r.db.QueryRow(ctx, `
		SELECT p.id, p.user_id, p.name, COALESCE(p.description, ''), p.type, p.risk_id, p.base_currency, p.is_default, p.price_value, p.created_at, p.updated_at,
		       r.id, r.name, COALESCE(r.description, ''), r.created_at, r.updated_at
		FROM portfolios p
		JOIN risks r ON p.risk_id = r.id
		WHERE p.id = $1 AND p.user_id = $2
	`, portfolioID, userID).Scan(
		&portfolio.ID,
		&portfolio.UserID,
		&portfolio.Name,
		&portfolio.Description,
		&portfolio.Type,
		&portfolio.RiskID,
		&portfolio.BaseCurrency,
		&portfolio.IsDefault,
		&portfolio.PriceValue,
		&portfolio.CreatedAt,
		&portfolio.UpdatedAt,
		&portfolio.Risk.ID,
		&portfolio.Risk.Name,
		&portfolio.Risk.Description,
		&portfolio.Risk.CreatedAt,
		&portfolio.Risk.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entities.Portfolio{}, errors.New("portfolio not found")
		}
		return entities.Portfolio{}, err
	}

	return portfolio, nil
}

func (r *Repository) GetEntriesByPortfolioID(ctx context.Context, portfolioID uuid.UUID) ([]entities.PortfolioEntry, error) {
	rows, err := r.db.Query(ctx, `
		SELECT pe.id, pe.portfolio_id, pe.asset_id, pe.source_id, pe.quantity, pe.price, pe.cost_currency, pe.category, pe.entry_date, COALESCE(pe.notes, ''), pe.created_at, pe.updated_at,
		       a.id, a.ticker, a.name, a.asset_type, COALESCE(a.exchange, ''), a.currency, a.current_price, a.price_updated_at, a.created_at, a.updated_at
		FROM portfolio_entries pe
		JOIN assets a ON a.id = pe.asset_id
		WHERE pe.portfolio_id = $1
		ORDER BY pe.created_at DESC
	`, portfolioID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := make([]entities.PortfolioEntry, 0)
	for rows.Next() {
		var entry entities.PortfolioEntry
		var quantity string
		var price string
		var sourceID pgtype.UUID

		if err := rows.Scan(
			&entry.ID,
			&entry.PortfolioID,
			&entry.AssetID,
			&sourceID,
			&quantity,
			&price,
			&entry.CostCurrency,
			&entry.Category,
			&entry.EntryDate,
			&entry.Notes,
			&entry.CreatedAt,
			&entry.UpdatedAt,
			&entry.Asset.ID,
			&entry.Asset.Ticker,
			&entry.Asset.Name,
			&entry.Asset.AssetType,
			&entry.Asset.Exchange,
			&entry.Asset.Currency,
			&entry.Asset.CurrentPrice,
			&entry.Asset.PriceUpdatedAt,
			&entry.Asset.CreatedAt,
			&entry.Asset.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if sourceID.Valid {
			entry.SourceID = uuid.UUID(sourceID.Bytes)
		}

		entry.Quantity = money.MustFromString(quantity)
		entry.Price = money.MustMoneyFromString(price, money.USD)

		entries = append(entries, entry)
	}

	return entries, nil
}

func (r *Repository) GetPortfoliosRisks(ctx context.Context) ([]entities.Risk, error) {
	rows, err := r.db.Query(ctx, "SELECT * FROM risks")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	risks := make([]entities.Risk, 0)
	for rows.Next() {
		var risk entities.Risk

		if err := rows.Scan(&risk.ID, &risk.Name, &risk.Description, &risk.CreatedAt, &risk.UpdatedAt); err != nil {
			return nil, err
		}

		risks = append(risks, risk)
	}

	return risks, nil
}

func (r *Repository) CreatePlatform(ctx context.Context, userID uuid.UUID, sourceType entities.SourceType, name, desciption string) (entities.InvestmentSource, error) {
	var platform entities.InvestmentSource
	err := r.db.QueryRow(ctx, "INSERT INTO investment_sources(user_id, source_type, name, description) VALUES ($1, $2, $3, $4) RETURNING id, name, description, created_at, updated_at", userID, sourceType, name, desciption).Scan(&platform.ID, &platform.Name, &platform.Description, &platform.CreatedAt, &platform.UpdatedAt)
	if err != nil {
		return entities.InvestmentSource{}, err
	}

	return platform, nil
}

func (r *Repository) GetPlatforms(ctx context.Context, userID uuid.UUID) ([]entities.InvestmentSource, error) {
	rows, err := r.db.Query(ctx, "SELECT * FROM investment_sources WHERE user_id = $1", userID.String())
	if err != nil {
		return []entities.InvestmentSource{}, err
	}

	platforms := make([]entities.InvestmentSource, 0)

	for rows.Next() {
		var platform entities.InvestmentSource

		if err := rows.Scan(
			&platform.ID,
			&platform.UserID,
			&platform.Name,
			&platform.SourceType,
			&platform.Description,
			&platform.IsActive,
			&platform.CreatedAt,
			&platform.UpdatedAt,
		); err != nil {
			return nil, err
		}

		portfolioEntries, err := r.GetPortfolioEntries(ctx, platform.ID)
		if err != nil {
			return []entities.InvestmentSource{}, err
		}

		platform.PortfolioEntries = append(platform.PortfolioEntries, portfolioEntries...)

		platforms = append(platforms, platform)
	}

	return platforms, nil
}

func (r *Repository) GetPortfolioEntries(ctx context.Context, sourceID uuid.UUID) ([]entities.PortfolioEntry, error) {
	rows, err := r.db.Query(ctx, "SELECT * FROM portfolio_entries WHERE source_id = $1", sourceID.String())
	if err != nil {
		return []entities.PortfolioEntry{}, err
	}

	portfolioEntries := make([]entities.PortfolioEntry, 0)

	for rows.Next() {
		var portfolioEntry entities.PortfolioEntry
		var quantity string
		var price string

		if err := rows.Scan(
			&portfolioEntry.ID,
			&portfolioEntry.PortfolioID,
			&portfolioEntry.AssetID,
			&portfolioEntry.SourceID,
			&quantity,
			&price,
			&portfolioEntry.CostCurrency,
			&portfolioEntry.Category,
			&portfolioEntry.EntryDate,
			&portfolioEntry.Notes,
			&portfolioEntry.CreatedAt,
			&portfolioEntry.UpdatedAt,
		); err != nil {
			return nil, err
		}

		portfolioEntry.Quantity = money.MustFromString(quantity)
		portfolioEntry.Price = money.MustMoneyFromString(price, money.USD)

		portfolioEntries = append(portfolioEntries, portfolioEntry)
	}

	return portfolioEntries, nil
}

func (r *Repository) GetAssets(ctx context.Context, offset, limit uint) ([]entities.Asset, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, ticker, name, asset_type, COALESCE(exchange, ''), currency, current_price, price_updated_at, created_at, updated_at
		FROM assets
		ORDER BY ticker
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return []entities.Asset{}, err
	}
	defer rows.Close()

	assets := make([]entities.Asset, 0)

	for rows.Next() {
		var asset entities.Asset

		if err := rows.Scan(
			&asset.ID,
			&asset.Ticker,
			&asset.Name,
			&asset.AssetType,
			&asset.Exchange,
			&asset.Currency,
			&asset.CurrentPrice,
			&asset.PriceUpdatedAt,
			&asset.CreatedAt,
			&asset.UpdatedAt,
		); err != nil {
			return nil, err
		}

		assets = append(assets, asset)
	}

	return assets, nil
}

func (r *Repository) UpdateAssetPrice(ctx context.Context, assetID uuid.UUID, price money.Money) (entities.Asset, error) {
	var asset entities.Asset

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
		&asset.CurrentPrice,
		&asset.PriceUpdatedAt,
		&asset.CreatedAt,
		&asset.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entities.Asset{}, errors.New("asset not found")
		}
		return entities.Asset{}, err
	}

	return asset, nil
}

func (r *Repository) CreatePortfolioEntry(ctx context.Context, userID, portfolioID, assetID uuid.UUID, sourceID uuid.UUID, quantity money.Decimal, price money.Money, costCurrency string, category entities.PortfolioEntryCategory, entryDate time.Time, notes string) (entities.PortfolioEntry, error) {
	var entry entities.PortfolioEntry
	var quantityValue string
	var priceValue string
	var sourceIDResult pgtype.UUID

	err := r.db.QueryRow(ctx, `
		INSERT INTO portfolio_entries (portfolio_id, asset_id, source_id, quantity, price, cost_currency, category, entry_date, notes)
		SELECT $1::uuid, $2::uuid, $3::uuid, $4::numeric, $5::numeric, $6::char(3), $7::portfolio_entry_category, $8::date, $9
		WHERE EXISTS (
			SELECT 1 FROM portfolios p WHERE p.id = $1 AND p.user_id = $10
		)
		AND ($3::uuid IS NULL OR EXISTS (
			SELECT 1 FROM investment_sources s WHERE s.id = $3 AND s.user_id = $10
		))
		RETURNING id, portfolio_id, asset_id, source_id, quantity, price, cost_currency, category, entry_date, notes, created_at, updated_at
	`, portfolioID, assetID, sourceID, quantity.String(), price.String(), costCurrency, category, entryDate, notes, userID).Scan(
		&entry.ID,
		&entry.PortfolioID,
		&entry.AssetID,
		&sourceIDResult,
		&quantityValue,
		&priceValue,
		&entry.CostCurrency,
		&entry.Category,
		&entry.EntryDate,
		&entry.Notes,
		&entry.CreatedAt,
		&entry.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entities.PortfolioEntry{}, errors.New("portfolio or source not found")
		}

		return entities.PortfolioEntry{}, err
	}

	if sourceIDResult.Valid {
		entry.SourceID = uuid.UUID(sourceIDResult.Bytes)
	}

	entry.Quantity = money.MustFromString(quantityValue)
	entry.Price = money.MustMoneyFromString(priceValue, money.USD)

	return entry, nil
}
