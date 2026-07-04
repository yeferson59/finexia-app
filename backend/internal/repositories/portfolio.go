package repositories

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/yeferson59/gofinance/money"

	"github.com/yeferson59/finexia-app/internal/dtos/portfolio"
	"github.com/yeferson59/finexia-app/internal/entities"
)

func (r *Repository) GetPortfoliosSummaryByUserID(ctx context.Context, userID uuid.UUID) ([]entities.PortfolioSummaryView, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			p.id,
			p.name,
			COALESCE(p.description, ''),
			p.type,
			p.base_currency,
			p.is_default,
			ri.id,
			ri.name,
			COALESCE(ps.total_positions, 0)::bigint,
			COALESCE(ps.total_cost_base,    0)::text,
			COALESCE(ps.total_market_value, 0)::text,
			COALESCE(ps.total_gain_loss,    0)::text,
			COALESCE(ps.total_gain_loss_pct,0)::text,
			p.created_at
		FROM portfolios p
		JOIN  risks ri          ON ri.id = p.risk_id
		LEFT JOIN portfolio_summary ps ON ps.portfolio_id = p.id
		WHERE p.user_id = $1
		ORDER BY p.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]entities.PortfolioSummaryView, 0)
	for rows.Next() {
		var item entities.PortfolioSummaryView
		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Description,
			&item.Type,
			&item.BaseCurrency,
			&item.IsDefault,
			&item.RiskID,
			&item.RiskName,
			&item.TotalPositions,
			&item.TotalCostBase,
			&item.TotalMarketValue,
			&item.TotalGainLoss,
			&item.TotalGainLossPct,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, nil
}

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

// scanAssetCurrentPrice populates asset.CurrentPrice from a nullable numeric string
// using the asset's own currency. money.Money.Scan only stores the value and leaves
// the currency at the zero value (XXX), which serializes to "¤" in the browser.
func scanAssetCurrentPrice(asset *entities.Asset, priceStr *string) {
	if priceStr == nil {
		return
	}
	cur, err := money.CurrencyFromISOCode(asset.Currency)
	if err != nil {
		return
	}
	m, err := money.NewMoneyFromString(*priceStr, cur)
	if err != nil {
		return
	}
	asset.CurrentPrice = &m
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
		var assetPriceStr *string

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
			&assetPriceStr,
			&entry.Asset.PriceUpdatedAt,
			&entry.Asset.CreatedAt,
			&entry.Asset.UpdatedAt,
		); err != nil {
			return nil, err
		}

		scanAssetCurrentPrice(&entry.Asset, assetPriceStr)

		if sourceID.Valid {
			entry.SourceID = uuid.UUID(sourceID.Bytes)
		}

		entry.Quantity = money.MustFromString(quantity)
		entry.Price = money.MustMoneyFromString(price, money.USD)

		entries = append(entries, entry)
	}

	return entries, nil
}

func (r *Repository) GetTopTransactionByPortfolioID(ctx context.Context, userID, portfolioID uuid.UUID) (portfolio.PortfolioTopTransactionDTO, error) {
	var dto portfolio.PortfolioTopTransactionDTO
	err := r.db.QueryRow(ctx, `
		SELECT
			(t.quantity::numeric * t.price::numeric)::text,
			t.type,
			t.currency,
			t.transaction_date,
			a.ticker,
			a.name
		FROM transactions t
		JOIN portfolio_entries pe ON pe.id = t.entry_id
		JOIN assets a ON a.id = pe.asset_id
		JOIN portfolios p ON p.id = pe.portfolio_id
		WHERE pe.portfolio_id = $1 AND p.user_id = $2
		ORDER BY t.quantity::numeric * t.price::numeric DESC
		LIMIT 1
	`, portfolioID, userID).Scan(
		&dto.Value,
		&dto.Type,
		&dto.Currency,
		&dto.TransactionDate,
		&dto.AssetTicker,
		&dto.AssetName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return portfolio.PortfolioTopTransactionDTO{}, nil
		}
		return portfolio.PortfolioTopTransactionDTO{}, err
	}
	return dto, nil
}

func (r *Repository) UpdatePortfolio(ctx context.Context, userID, portfolioID uuid.UUID, name, description string, portfolioType entities.PortfolioType, riskID uuid.UUID, isDefault bool) (entities.Portfolio, error) {
	var portfolio entities.Portfolio
	err := r.db.QueryRow(ctx, `
		UPDATE portfolios
		SET name = $1, description = $2, type = $3, risk_id = $4, is_default = $5, updated_at = NOW()
		WHERE id = $6 AND user_id = $7
		RETURNING id, user_id, name, COALESCE(description, ''), type, risk_id, base_currency, is_default, price_value, created_at, updated_at
	`, name, description, portfolioType, riskID, isDefault, portfolioID, userID).Scan(
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
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entities.Portfolio{}, errors.New("portfolio not found")
		}
		return entities.Portfolio{}, err
	}

	if err := r.db.QueryRow(ctx, "SELECT id, name, COALESCE(description, ''), created_at, updated_at FROM risks WHERE id = $1", riskID).Scan(
		&portfolio.Risk.ID,
		&portfolio.Risk.Name,
		&portfolio.Risk.Description,
		&portfolio.Risk.CreatedAt,
		&portfolio.Risk.UpdatedAt,
	); err != nil {
		return entities.Portfolio{}, err
	}

	return portfolio, nil
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

func (r *Repository) GetPlatformsWithStats(ctx context.Context, userID uuid.UUID) ([]entities.PlatformStats, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			is_.id,
			is_.name,
			COALESCE(is_.description, ''),
			is_.source_type,
			is_.is_active,
			is_.created_at,
			is_.updated_at,
			COUNT(pe.id)::bigint AS investments,
			COALESCE(SUM(pe.quantity::numeric * pe.price::numeric), 0)::text AS total_value
		FROM investment_sources is_
		LEFT JOIN portfolio_entries pe ON pe.source_id = is_.id
		WHERE is_.user_id = $1
		GROUP BY is_.id
		ORDER BY is_.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]entities.PlatformStats, 0)
	for rows.Next() {
		var p entities.PlatformStats
		if err := rows.Scan(
			&p.ID, &p.Name, &p.Description, &p.SourceType,
			&p.IsActive, &p.CreatedAt, &p.UpdatedAt,
			&p.Investments, &p.TotalValue,
		); err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, nil
}

func (r *Repository) UpdatePlatform(ctx context.Context, userID, sourceID uuid.UUID, name, description string, sourceType entities.SourceType, isActive bool) (entities.PlatformStats, error) {
	tag, err := r.db.Exec(ctx, `
		UPDATE investment_sources
		SET name = $3, description = $4, source_type = $5, is_active = $6, updated_at = NOW()
		WHERE id = $1 AND user_id = $2
	`, sourceID, userID, name, description, sourceType, isActive)
	if err != nil {
		return entities.PlatformStats{}, err
	}
	if tag.RowsAffected() == 0 {
		return entities.PlatformStats{}, errors.New("platform not found")
	}

	var p entities.PlatformStats
	if err := r.db.QueryRow(ctx, `
		SELECT
			is_.id, is_.name, COALESCE(is_.description, ''),
			is_.source_type, is_.is_active, is_.created_at, is_.updated_at,
			COUNT(pe.id)::bigint,
			COALESCE(SUM(pe.quantity::numeric * pe.price::numeric), 0)::text
		FROM investment_sources is_
		LEFT JOIN portfolio_entries pe ON pe.source_id = is_.id
		WHERE is_.id = $1 AND is_.user_id = $2
		GROUP BY is_.id
	`, sourceID, userID).Scan(
		&p.ID, &p.Name, &p.Description, &p.SourceType,
		&p.IsActive, &p.CreatedAt, &p.UpdatedAt,
		&p.Investments, &p.TotalValue,
	); err != nil {
		return entities.PlatformStats{}, err
	}
	return p, nil
}

func (r *Repository) DeletePlatform(ctx context.Context, userID, sourceID uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `
		DELETE FROM investment_sources WHERE id = $1 AND user_id = $2
	`, sourceID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("platform not found")
	}
	return nil
}

func (r *Repository) CreatePlatform(ctx context.Context, userID uuid.UUID, sourceType entities.SourceType, name, desciption string) (entities.InvestmentSource, error) {
	var platform entities.InvestmentSource
	err := r.db.QueryRow(ctx, "INSERT INTO investment_sources(user_id, source_type, name, description) VALUES ($1, $2, $3, $4) RETURNING id, name, description, created_at, updated_at", userID, sourceType, name, desciption).Scan(&platform.ID, &platform.Name, &platform.Description, &platform.CreatedAt, &platform.UpdatedAt)
	if err != nil {
		return entities.InvestmentSource{}, err
	}

	return platform, nil
}

func (r *Repository) GetAssetByID(ctx context.Context, assetID uuid.UUID) (entities.Asset, error) {
	var asset entities.Asset
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
			return entities.Asset{}, errors.New("asset not found")
		}
		return entities.Asset{}, err
	}
	scanAssetCurrentPrice(&asset, priceStr)
	return asset, nil
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

func (r *Repository) UpdateAssetPrice(ctx context.Context, assetID uuid.UUID, price money.Money) (entities.Asset, error) {
	var asset entities.Asset
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
			return entities.Asset{}, errors.New("asset not found")
		}
		return entities.Asset{}, err
	}

	scanAssetCurrentPrice(&asset, priceStr)
	return asset, nil
}

func (r *Repository) SearchAssets(ctx context.Context, search string, offset, limit uint) ([]entities.Asset, error) {
	pattern := "%" + strings.ToUpper(strings.TrimSpace(search)) + "%"
	rows, err := r.db.Query(ctx, `
		SELECT id, ticker, name, asset_type, COALESCE(exchange, ''), currency, current_price, price_updated_at, created_at, updated_at
		FROM assets
		WHERE UPPER(ticker) LIKE $1 OR UPPER(name) LIKE $1
		ORDER BY ticker
		LIMIT $2 OFFSET $3
	`, pattern, limit, offset)
	if err != nil {
		return []entities.Asset{}, err
	}
	defer rows.Close()

	assets := make([]entities.Asset, 0)
	for rows.Next() {
		var asset entities.Asset
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

func (r *Repository) UpsertAsset(ctx context.Context, ticker, name string, assetType entities.AssetType, exchange, currency string) (entities.Asset, error) {
	var asset entities.Asset
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

func (r *Repository) GetRecentTransactionsByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]entities.Transaction, error) {
	rows, err := r.db.Query(ctx, `
		SELECT t.id, t.entry_id, t.type, t.quantity, t.price, t.currency, t.fees,
		       t.transaction_date, COALESCE(t.notes, ''), t.created_at, t.updated_at,
		       a.ticker, a.name
		FROM transactions t
		JOIN portfolio_entries pe ON pe.id = t.entry_id
		JOIN portfolios p ON p.id = pe.portfolio_id
		JOIN assets a ON a.id = pe.asset_id
		WHERE p.user_id = $1
		ORDER BY t.transaction_date DESC, t.created_at DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	txns := make([]entities.Transaction, 0)
	for rows.Next() {
		var txn entities.Transaction
		var quantity, price, fees string
		if err := rows.Scan(
			&txn.ID, &txn.EntryID, &txn.Type, &quantity, &price,
			&txn.Currency, &fees, &txn.TransactionDate, &txn.Notes,
			&txn.CreatedAt, &txn.UpdatedAt,
			&txn.Entry.Asset.Ticker, &txn.Entry.Asset.Name,
		); err != nil {
			return nil, err
		}
		txn.Quantity = money.MustFromString(quantity)
		txn.Price = money.MustMoneyFromString(price, money.USD)
		txn.Fees = money.MustMoneyFromString(fees, money.USD)
		txns = append(txns, txn)
	}
	return txns, nil
}

func (r *Repository) GetAssetAllocationByUserID(ctx context.Context, userID uuid.UUID) ([]entities.AllocationItem, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			pe.category,
			COALESCE(SUM(pe.quantity::numeric * COALESCE(a.current_price::numeric, pe.price::numeric)), 0)::text AS market_value
		FROM portfolio_entries pe
		JOIN portfolios p ON p.id = pe.portfolio_id
		JOIN assets a ON a.id = pe.asset_id
		WHERE p.user_id = $1
		  AND pe.quantity::numeric > 0
		GROUP BY pe.category
		ORDER BY COALESCE(SUM(pe.quantity::numeric * COALESCE(a.current_price::numeric, pe.price::numeric)), 0) DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]entities.AllocationItem, 0)
	for rows.Next() {
		var item entities.AllocationItem
		if err := rows.Scan(&item.Category, &item.MarketValue); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *Repository) GetTransactionsByEntryID(ctx context.Context, userID, entryID uuid.UUID) ([]entities.Transaction, error) {
	var owned bool
	if err := r.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM portfolio_entries pe
			JOIN portfolios p ON p.id = pe.portfolio_id
			WHERE pe.id = $1 AND p.user_id = $2
		)
	`, entryID, userID).Scan(&owned); err != nil {
		return nil, err
	}
	if !owned {
		return nil, errors.New("portfolio entry not found")
	}

	rows, err := r.db.Query(ctx, `
		SELECT id, entry_id, type, quantity, price, currency, fees, transaction_date, COALESCE(notes, ''), created_at, updated_at
		FROM transactions
		WHERE entry_id = $1
		ORDER BY transaction_date DESC, created_at DESC
	`, entryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	txns := make([]entities.Transaction, 0)
	for rows.Next() {
		var txn entities.Transaction
		var quantity, price, fees string
		if err := rows.Scan(
			&txn.ID,
			&txn.EntryID,
			&txn.Type,
			&quantity,
			&price,
			&txn.Currency,
			&fees,
			&txn.TransactionDate,
			&txn.Notes,
			&txn.CreatedAt,
			&txn.UpdatedAt,
		); err != nil {
			return nil, err
		}
		txn.Quantity = money.MustFromString(quantity)
		txn.Price = money.MustMoneyFromString(price, money.USD)
		txn.Fees = money.MustMoneyFromString(fees, money.USD)
		txns = append(txns, txn)
	}
	return txns, nil
}

func (r *Repository) CountAssetTransactions(ctx context.Context, userID, portfolioID uuid.UUID, ticker string) (int, error) {
	var total int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM transactions t
		JOIN portfolio_entries pe ON pe.id = t.entry_id
		JOIN assets a ON a.id = pe.asset_id
		JOIN portfolios p ON p.id = pe.portfolio_id
		WHERE p.id = $1 AND a.ticker = $2 AND p.user_id = $3
	`, portfolioID, ticker, userID).Scan(&total)
	return total, err
}

func (r *Repository) GetAssetTransactionsPaginated(ctx context.Context, userID, portfolioID uuid.UUID, ticker string, limit, offset int) ([]entities.Transaction, error) {
	rows, err := r.db.Query(ctx, `
		SELECT t.id, t.entry_id, t.type, t.quantity, t.price, t.currency, t.fees,
		       t.transaction_date, COALESCE(t.notes, ''), t.created_at, t.updated_at
		FROM transactions t
		JOIN portfolio_entries pe ON pe.id = t.entry_id
		JOIN assets a ON a.id = pe.asset_id
		JOIN portfolios p ON p.id = pe.portfolio_id
		WHERE p.id = $1 AND a.ticker = $2 AND p.user_id = $3
		ORDER BY t.transaction_date DESC, t.created_at DESC
		LIMIT $4 OFFSET $5
	`, portfolioID, ticker, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	txns := make([]entities.Transaction, 0)
	for rows.Next() {
		var txn entities.Transaction
		var quantity, price, fees string
		if err := rows.Scan(
			&txn.ID,
			&txn.EntryID,
			&txn.Type,
			&quantity,
			&price,
			&txn.Currency,
			&fees,
			&txn.TransactionDate,
			&txn.Notes,
			&txn.CreatedAt,
			&txn.UpdatedAt,
		); err != nil {
			return nil, err
		}
		txn.Quantity = money.MustFromString(quantity)
		txn.Price = money.MustMoneyFromString(price, money.USD)
		txn.Fees = money.MustMoneyFromString(fees, money.USD)
		txns = append(txns, txn)
	}
	return txns, nil
}

func (r *Repository) CreateTransaction(ctx context.Context, userID, entryID uuid.UUID, txnType entities.TransactionType, quantity money.Decimal, price money.Money, currency string, fees money.Money, transactionDate time.Time, notes string) (entities.Transaction, error) {
	var owned bool
	if err := r.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM portfolio_entries pe
			JOIN portfolios p ON p.id = pe.portfolio_id
			WHERE pe.id = $1 AND p.user_id = $2
		)
	`, entryID, userID).Scan(&owned); err != nil {
		return entities.Transaction{}, err
	}
	if !owned {
		return entities.Transaction{}, errors.New("portfolio entry not found")
	}

	var txn entities.Transaction
	var quantityValue, priceValue, feesValue string
	if err := r.db.QueryRow(ctx, `
		INSERT INTO transactions (entry_id, type, quantity, price, currency, fees, transaction_date, notes)
		VALUES ($1::uuid, $2::transaction_type, $3::numeric, $4::numeric, $5::char(3), $6::numeric, $7::date, $8)
		RETURNING id, entry_id, type, quantity, price, currency, fees, transaction_date, COALESCE(notes, ''), created_at, updated_at
	`, entryID, txnType, quantity.String(), price.String(), currency, fees.String(), transactionDate, notes).Scan(
		&txn.ID,
		&txn.EntryID,
		&txn.Type,
		&quantityValue,
		&priceValue,
		&txn.Currency,
		&feesValue,
		&txn.TransactionDate,
		&txn.Notes,
		&txn.CreatedAt,
		&txn.UpdatedAt,
	); err != nil {
		return entities.Transaction{}, err
	}

	txn.Quantity = money.MustFromString(quantityValue)
	txn.Price = money.MustMoneyFromString(priceValue, money.USD)
	txn.Fees = money.MustMoneyFromString(feesValue, money.USD)
	return txn, nil
}

func (r *Repository) UpdateTransaction(ctx context.Context, userID, txnID uuid.UUID, txnType entities.TransactionType, quantity money.Decimal, price money.Money, currency string, fees money.Money, transactionDate time.Time, notes string) (entities.Transaction, error) {
	var txn entities.Transaction
	var quantityValue, priceValue, feesValue string
	err := r.db.QueryRow(ctx, `
		UPDATE transactions SET
			type             = $1::transaction_type,
			quantity         = $2::numeric,
			price            = $3::numeric,
			currency         = $4::char(3),
			fees             = $5::numeric,
			transaction_date = $6::date,
			notes            = $7,
			updated_at       = NOW()
		WHERE id = $8
		  AND entry_id IN (
			SELECT pe.id FROM portfolio_entries pe
			JOIN portfolios p ON p.id = pe.portfolio_id
			WHERE p.user_id = $9
		  )
		RETURNING id, entry_id, type, quantity, price, currency, fees, transaction_date, COALESCE(notes, ''), created_at, updated_at
	`, txnType, quantity.String(), price.String(), currency, fees.String(), transactionDate, notes, txnID, userID).Scan(
		&txn.ID,
		&txn.EntryID,
		&txn.Type,
		&quantityValue,
		&priceValue,
		&txn.Currency,
		&feesValue,
		&txn.TransactionDate,
		&txn.Notes,
		&txn.CreatedAt,
		&txn.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entities.Transaction{}, errors.New("transaction not found")
		}
		return entities.Transaction{}, err
	}
	txn.Quantity = money.MustFromString(quantityValue)
	txn.Price = money.MustMoneyFromString(priceValue, money.USD)
	txn.Fees = money.MustMoneyFromString(feesValue, money.USD)
	return txn, nil
}

func (r *Repository) GetEntryWithAsset(ctx context.Context, entryID uuid.UUID) (entities.PortfolioEntry, error) {
	var entry entities.PortfolioEntry
	err := r.db.QueryRow(ctx, `
		SELECT pe.id, pe.portfolio_id, pe.asset_id,
		       a.ticker, a.name
		FROM portfolio_entries pe
		JOIN assets a ON a.id = pe.asset_id
		WHERE pe.id = $1
	`, entryID).Scan(
		&entry.ID,
		&entry.PortfolioID,
		&entry.AssetID,
		&entry.Asset.Ticker,
		&entry.Asset.Name,
	)
	if err != nil {
		return entities.PortfolioEntry{}, err
	}
	return entry, nil
}

func (r *Repository) CreatePortfolioEntry(ctx context.Context, userID, portfolioID, assetID uuid.UUID, sourceID uuid.UUID, txnType entities.TransactionType, quantity money.Decimal, price money.Money, costCurrency string, category entities.PortfolioEntryCategory, entryDate time.Time, notes string) (entities.PortfolioEntry, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return entities.PortfolioEntry{}, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var owned bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS (SELECT 1 FROM portfolios WHERE id = $1 AND user_id = $2)
		   AND ($3::uuid IS NULL OR EXISTS (SELECT 1 FROM investment_sources WHERE id = $3 AND user_id = $2))
	`, portfolioID, userID, sourceID).Scan(&owned); err != nil {
		return entities.PortfolioEntry{}, err
	}

	if !owned {
		return entities.PortfolioEntry{}, errors.New("portfolio or source not found")
	}

	var entryID uuid.UUID
	if err := tx.QueryRow(ctx, `
		INSERT INTO portfolio_entries (portfolio_id, asset_id, source_id, quantity, price, cost_currency, category, entry_date, notes)
		VALUES ($1::uuid, $2::uuid, $3::uuid, 0, $4::numeric, $5::char(3), $6::portfolio_entry_category, $7::date, $8)
		ON CONFLICT (portfolio_id, asset_id, COALESCE(source_id::TEXT, ''))
		DO UPDATE SET updated_at = NOW()
		RETURNING id
	`, portfolioID, assetID, sourceID, price.String(), costCurrency, category, entryDate, notes).Scan(&entryID); err != nil {
		return entities.PortfolioEntry{}, err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO transactions (entry_id, type, quantity, price, currency, fees, transaction_date, notes)
		VALUES ($1::uuid, $2::transaction_type, $3::numeric, $4::numeric, $5::char(3), 0, $6::date, $7)
	`, entryID, txnType, quantity.String(), price.String(), costCurrency, entryDate, notes); err != nil {
		return entities.PortfolioEntry{}, err
	}

	// Read the position back with the values the trigger just recomputed.
	var entry entities.PortfolioEntry
	var quantityValue, priceValue string
	var sourceIDResult pgtype.UUID
	if err := tx.QueryRow(ctx, `
		SELECT id, portfolio_id, asset_id, source_id, quantity, price, cost_currency, category, entry_date, COALESCE(notes, ''), created_at, updated_at
		FROM portfolio_entries
		WHERE id = $1
	`, entryID).Scan(
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
	); err != nil {
		return entities.PortfolioEntry{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return entities.PortfolioEntry{}, err
	}

	if sourceIDResult.Valid {
		entry.SourceID = uuid.UUID(sourceIDResult.Bytes)
	}

	entry.Quantity = money.MustFromString(quantityValue)
	entry.Price = money.MustMoneyFromString(priceValue, money.USD)

	return entry, nil
}

// ImportEntryTransactions persists a batch of validated spreadsheet rows in a
// single database transaction: each row resolves (or creates) its asset,
// upserts the portfolio position and inserts the transaction, so a mid-batch
// failure never leaves a half-imported file behind.
func (r *Repository) ImportEntryTransactions(ctx context.Context, userID, portfolioID, sourceID uuid.UUID, rows []entities.ImportTransactionRow) (int, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var owned bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS (SELECT 1 FROM portfolios WHERE id = $1 AND user_id = $2)
		   AND EXISTS (SELECT 1 FROM investment_sources WHERE id = $3 AND user_id = $2)
	`, portfolioID, userID, sourceID).Scan(&owned); err != nil {
		return 0, err
	}
	if !owned {
		return 0, errors.New("portfolio or source not found")
	}

	// Cache asset lookups: classic spreadsheets repeat the same ticker on
	// many rows.
	assetIDs := make(map[string]uuid.UUID)
	imported := 0
	for _, row := range rows {
		assetID, ok := assetIDs[row.Ticker]
		if !ok {
			// Reuse an existing asset for the ticker (regardless of exchange)
			// before creating a new one, so imports never clobber curated
			// asset data the way an upsert would.
			err := tx.QueryRow(ctx, `
				SELECT id FROM assets WHERE UPPER(ticker) = $1 ORDER BY created_at LIMIT 1
			`, row.Ticker).Scan(&assetID)
			if errors.Is(err, pgx.ErrNoRows) {
				err = tx.QueryRow(ctx, `
					INSERT INTO assets (ticker, name, asset_type, exchange, currency, created_at, updated_at)
					VALUES ($1, $2, $3::asset_type, NULL, $4, NOW(), NOW())
					ON CONFLICT (ticker, COALESCE(exchange, '')) DO UPDATE SET updated_at = NOW()
					RETURNING id
				`, row.Ticker, row.AssetName, row.AssetType, row.Currency).Scan(&assetID)
			}
			if err != nil {
				return 0, err
			}
			assetIDs[row.Ticker] = assetID
		}

		var entryID uuid.UUID
		if err := tx.QueryRow(ctx, `
			INSERT INTO portfolio_entries (portfolio_id, asset_id, source_id, quantity, price, cost_currency, category, entry_date, notes)
			VALUES ($1::uuid, $2::uuid, $3::uuid, 0, $4::numeric, $5::char(3), $6::portfolio_entry_category, $7::date, $8)
			ON CONFLICT (portfolio_id, asset_id, COALESCE(source_id::TEXT, ''))
			DO UPDATE SET updated_at = NOW()
			RETURNING id
		`, portfolioID, assetID, sourceID, row.Price.String(), row.Currency, row.Category, row.Date, row.Notes).Scan(&entryID); err != nil {
			return 0, err
		}

		if _, err := tx.Exec(ctx, `
			INSERT INTO transactions (entry_id, type, quantity, price, currency, fees, transaction_date, notes)
			VALUES ($1::uuid, $2::transaction_type, $3::numeric, $4::numeric, $5::char(3), $6::numeric, $7::date, $8)
		`, entryID, row.Type, row.Quantity.String(), row.Price.String(), row.Currency, row.Fees.String(), row.Date, row.Notes); err != nil {
			return 0, err
		}
		imported++
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return imported, nil
}

func (r *Repository) GetAllPortfolioSummaryRows(ctx context.Context) ([]entities.PortfolioSnapshotRow, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			p.id,
			p.base_currency,
			COALESCE(ps.total_market_value, 0)::text,
			COALESCE(ps.total_cost_base,    0)::text,
			COALESCE(ps.total_gain_loss,    0)::text,
			COALESCE(ps.total_gain_loss_pct,0)::text
		FROM portfolios p
		LEFT JOIN portfolio_summary ps ON ps.portfolio_id = p.id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]entities.PortfolioSnapshotRow, 0)
	for rows.Next() {
		var row entities.PortfolioSnapshotRow
		if err := rows.Scan(
			&row.PortfolioID,
			&row.BaseCurrency,
			&row.TotalMarketValue,
			&row.TotalCostBase,
			&row.TotalGainLoss,
			&row.TotalGainLossPct,
		); err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *Repository) UpsertPortfolioSnapshot(
	ctx context.Context,
	portfolioID uuid.UUID,
	snapshotDate time.Time,
	totalValue, currency, totalGainLoss, totalGainLossPct string,
) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO portfolio_snapshots
			(portfolio_id, snapshot_date, total_value, currency, allocation, total_gain_loss, total_gain_loss_pct)
		VALUES ($1, $2::date, $3::numeric, $4, '{}', $5::numeric, $6::numeric)
		ON CONFLICT (portfolio_id, snapshot_date)
		DO UPDATE SET
			total_value         = EXCLUDED.total_value,
			total_gain_loss     = EXCLUDED.total_gain_loss,
			total_gain_loss_pct = EXCLUDED.total_gain_loss_pct
	`, portfolioID, snapshotDate, totalValue, currency, totalGainLoss, totalGainLossPct)
	return err
}

func (r *Repository) GetPortfolioGrowthByUserID(
	ctx context.Context,
	userID uuid.UUID,
	hasSince bool,
	since time.Time,
) ([]entities.PortfolioGrowthPoint, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			ps.snapshot_date,
			SUM(ps.total_value)::text,
			(SUM(ps.total_value) - SUM(ps.total_gain_loss))::text,
			SUM(ps.total_gain_loss)::text,
			CASE
				WHEN (SUM(ps.total_value) - SUM(ps.total_gain_loss)) > 0
				THEN ((SUM(ps.total_gain_loss) / (SUM(ps.total_value) - SUM(ps.total_gain_loss))) * 100)::text
				ELSE '0'
			END
		FROM portfolio_snapshots ps
		JOIN portfolios p ON p.id = ps.portfolio_id
		WHERE p.user_id = $1
		  AND ($2::boolean = FALSE OR ps.snapshot_date >= $3::date)
		GROUP BY ps.snapshot_date
		ORDER BY ps.snapshot_date ASC
	`, userID, hasSince, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]entities.PortfolioGrowthPoint, 0)
	for rows.Next() {
		var point entities.PortfolioGrowthPoint
		if err := rows.Scan(
			&point.Date,
			&point.TotalValue,
			&point.TotalCostBase,
			&point.GainLoss,
			&point.GainLossPct,
		); err != nil {
			return nil, err
		}
		result = append(result, point)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
