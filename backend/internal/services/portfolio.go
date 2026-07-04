package services

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/yeferson59/gofinance/money"
	"golang.org/x/sync/errgroup"

	portfoliodto "github.com/yeferson59/finexia-app/internal/dtos/portfolio"
	"github.com/yeferson59/finexia-app/internal/entities"
	"github.com/yeferson59/finexia-app/internal/mail"
)

// risksCache memoizes the risk catalog: it is seed data shared by every user
// and requested on each portfolio page, so a short TTL avoids one DB
// round-trip per page view without risking staleness for long.
type risksCache struct {
	mu        sync.RWMutex
	risks     []entities.Risk
	expiresAt time.Time
}

const risksCacheTTL = 10 * time.Minute

func (s *Services) GetPortfoliosRisks(ctx context.Context) ([]entities.Risk, error) {
	if c := s.risksCache; c != nil {
		c.mu.RLock()
		risks, fresh := c.risks, time.Now().Before(c.expiresAt)
		c.mu.RUnlock()
		if fresh {
			return risks, nil
		}
	}

	risks, err := s.repos.GetPortfoliosRisks(ctx)
	if err != nil {
		return []entities.Risk{}, err
	}

	if c := s.risksCache; c != nil {
		c.mu.Lock()
		c.risks, c.expiresAt = risks, time.Now().Add(risksCacheTTL)
		c.mu.Unlock()
	}

	return risks, nil
}

func (s *Services) GetPortfolios(ctx context.Context, userID uuid.UUID) ([]entities.Portfolio, error) {
	portfolios, err := s.repos.GetPortfoliosByUserID(ctx, userID)
	if err != nil {
		return []entities.Portfolio{}, err
	}

	return portfolios, nil
}

func (s *Services) GetPortfoliosSummary(ctx context.Context, userID uuid.UUID) ([]entities.PortfolioSummaryView, error) {
	return s.repos.GetPortfoliosSummaryByUserID(ctx, userID)
}

func (s *Services) GetPortfolio(ctx context.Context, userID, portfolioID uuid.UUID) (entities.Portfolio, error) {
	// The portfolio header and its entries are independent queries; running
	// them concurrently halves the latency of the portfolio detail endpoint.
	// The ownership check in GetPortfolioByID still gates the response: if it
	// fails, the fetched entries are discarded with the error.
	var (
		portfolio entities.Portfolio
		entries   []entities.PortfolioEntry
	)

	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		var err error
		portfolio, err = s.repos.GetPortfolioByID(gctx, portfolioID, userID)
		return err
	})
	g.Go(func() error {
		var err error
		entries, err = s.repos.GetEntriesByPortfolioID(gctx, portfolioID)
		return err
	})
	if err := g.Wait(); err != nil {
		return entities.Portfolio{}, err
	}

	portfolio.PortfolioEntries = entries

	return portfolio, nil
}

func (s *Services) CreatePortfolio(ctx context.Context, userID uuid.UUID, name string, description string, baseCurrency string, riskID uuid.UUID, typePortfolio entities.PortfolioType, price_value money.Money, isDefault bool) (entities.Portfolio, error) {
	portfolio, err := s.repos.CreatePortfolio(ctx, userID, name, description, baseCurrency, riskID, typePortfolio, price_value, isDefault)
	if err != nil {
		return entities.Portfolio{}, err
	}

	return portfolio, nil
}

func (s *Services) GetPortfolioTopTransaction(ctx context.Context, userID, portfolioID uuid.UUID) (portfoliodto.PortfolioTopTransactionDTO, error) {
	return s.repos.GetTopTransactionByPortfolioID(ctx, userID, portfolioID)
}

func (s *Services) UpdatePortfolio(ctx context.Context, userID, portfolioID uuid.UUID, name, description string, portfolioType entities.PortfolioType, riskID uuid.UUID, isDefault bool) (entities.Portfolio, error) {
	return s.repos.UpdatePortfolio(ctx, userID, portfolioID, name, description, portfolioType, riskID, isDefault)
}

func (s *Services) CreatePlatform(ctx context.Context, userID uuid.UUID, sourceType entities.SourceType, name, description string) (entities.InvestmentSource, error) {
	platform, err := s.repos.CreatePlatform(ctx, userID, sourceType, strings.ToLower(name), description)
	if err != nil {
		return entities.InvestmentSource{}, err
	}

	return platform, nil
}

func (s *Services) GetPlatforms(ctx context.Context, userID uuid.UUID) ([]entities.PlatformStats, error) {
	return s.repos.GetPlatformsWithStats(ctx, userID)
}

func (s *Services) UpdatePlatform(ctx context.Context, userID, sourceID uuid.UUID, name, description string, sourceType entities.SourceType, isActive bool) (entities.PlatformStats, error) {
	return s.repos.UpdatePlatform(ctx, userID, sourceID, name, description, sourceType, isActive)
}

func (s *Services) DeletePlatform(ctx context.Context, userID, sourceID uuid.UUID) error {
	return s.repos.DeletePlatform(ctx, userID, sourceID)
}

func (s *Services) GetAssets(ctx context.Context, offset, limit uint) ([]entities.Asset, error) {
	return s.repos.GetAssets(ctx, offset, limit)
}

func (s *Services) SearchAssets(ctx context.Context, search string, offset, limit uint) ([]entities.Asset, error) {
	return s.repos.SearchAssets(ctx, search, offset, limit)
}

func (s *Services) CreateAsset(ctx context.Context, ticker, name string, assetType entities.AssetType, exchange, currency string) (entities.Asset, error) {
	return s.repos.UpsertAsset(ctx, ticker, name, assetType, exchange, currency)
}

func (s *Services) UpdateAssetPrice(ctx context.Context, assetID uuid.UUID, price money.Money) (entities.Asset, error) {
	return s.repos.UpdateAssetPrice(ctx, assetID, price)
}

func (s *Services) CreatePortfolioEntry(ctx context.Context, userID, portfolioID, assetID uuid.UUID, sourceID uuid.UUID, txnType entities.TransactionType, quantity money.Decimal, price money.Money, costCurrency string, category entities.PortfolioEntryCategory, entryDate time.Time, notes string) (entities.PortfolioEntry, error) {
	entry, err := s.repos.CreatePortfolioEntry(ctx, userID, portfolioID, assetID, sourceID, txnType, quantity, price, costCurrency, category, entryDate, notes)
	if err != nil {
		return entities.PortfolioEntry{}, err
	}

	return entry, nil
}

func (s *Services) GetTransactionsByEntry(ctx context.Context, userID, entryID uuid.UUID) ([]entities.Transaction, error) {
	return s.repos.GetTransactionsByEntryID(ctx, userID, entryID)
}

func (s *Services) GetAssetTransactionsPaginated(ctx context.Context, userID, portfolioID uuid.UUID, ticker string, page, limit int) ([]entities.Transaction, int, error) {
	offset := (page - 1) * limit

	// The count and the page are independent reads; overlap them instead of
	// paying two sequential DB round-trips.
	var (
		total int
		txns  []entities.Transaction
	)

	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		var err error
		total, err = s.repos.CountAssetTransactions(gctx, userID, portfolioID, ticker)
		return err
	})
	g.Go(func() error {
		var err error
		txns, err = s.repos.GetAssetTransactionsPaginated(gctx, userID, portfolioID, ticker, limit, offset)
		return err
	})
	if err := g.Wait(); err != nil {
		return nil, 0, err
	}

	return txns, total, nil
}

func (s *Services) GetRecentUserTransactions(ctx context.Context, userID uuid.UUID, limit int) ([]entities.Transaction, error) {
	return s.repos.GetRecentTransactionsByUserID(ctx, userID, limit)
}

func (s *Services) GetAssetAllocation(ctx context.Context, userID uuid.UUID) ([]entities.AllocationItem, error) {
	return s.repos.GetAssetAllocationByUserID(ctx, userID)
}

func (s *Services) UpdateTransaction(ctx context.Context, userID, txnID uuid.UUID, txnType entities.TransactionType, quantity money.Decimal, price money.Money, currency string, fees money.Money, transactionDate time.Time, notes string) (entities.Transaction, error) {
	return s.repos.UpdateTransaction(ctx, userID, txnID, txnType, quantity, price, currency, fees, transactionDate, notes)
}

func (s *Services) CreateTransaction(ctx context.Context, userID, entryID uuid.UUID, txnType entities.TransactionType, quantity money.Decimal, price money.Money, currency string, fees money.Money, transactionDate time.Time, notes string) (entities.Transaction, error) {
	txn, err := s.repos.CreateTransaction(ctx, userID, entryID, txnType, quantity, price, currency, fees, transactionDate, notes)
	if err != nil {
		return entities.Transaction{}, err
	}

	go s.sendTransactionAlert(userID, entryID, txn)

	return txn, nil
}

func (s *Services) sendTransactionAlert(userID, entryID uuid.UUID, txn entities.Transaction) {
	ctx := context.Background()

	prefs, err := s.repos.GetUserPreferences(ctx, userID)
	if err != nil || !prefs.EmailAlerts {
		return
	}

	user, err := s.repos.GetUserByID(ctx, userID)
	if err != nil {
		return
	}

	entry, err := s.repos.GetEntryWithAsset(ctx, entryID)
	if err != nil {
		return
	}

	qty := txn.Quantity.String()
	priceStr := txn.Price.String()
	totalStr := fmt.Sprintf("%.2f", txn.Quantity.InexactFloat64()*txn.Price.InexactFloat64())

	data := mail.ActivityAlertData{
		UserName:        user.Name,
		AssetTicker:     entry.Asset.Ticker,
		AssetName:       entry.Asset.Name,
		TransactionType: string(txn.Type),
		Quantity:        qty,
		Price:           priceStr,
		Total:           totalStr,
		Currency:        txn.Currency,
		TransactionDate: txn.TransactionDate.Format("02 Jan 2006"),
		DashboardURL:    s.cfg.PublicURL + "/dashboard/portfolios",
	}

	_ = s.mail.SendActivityAlert(user.Email, data)
}

func (s *Services) SendWeeklySummaryEmails(ctx context.Context) (int, []error) {
	users, err := s.repos.GetUsersWithWeeklySummary(ctx)
	if err != nil {
		return 0, []error{err}
	}

	now := time.Now()
	year, week := now.ISOWeek()
	weekLabel := fmt.Sprintf("Semana %d — %d", week, year)

	var errs []error
	sent := 0

	for _, u := range users {
		summaries, err := s.repos.GetPortfoliosSummaryByUserID(ctx, u.ID)
		if err != nil || len(summaries) == 0 {
			continue
		}

		var totalValue, totalGain, totalGainPct float64
		portfolios := make([]mail.WeeklySummaryPortfolio, 0, len(summaries))

		for _, p := range summaries {
			mv, _ := strconv.ParseFloat(p.TotalMarketValue, 64)
			gl, _ := strconv.ParseFloat(p.TotalGainLoss, 64)
			glp, _ := strconv.ParseFloat(p.TotalGainLossPct, 64)
			totalValue += mv
			totalGain += gl

			color := "#22c97e"
			if glp < 0 {
				color = "#e05a5a"
			}

			portfolios = append(portfolios, mail.WeeklySummaryPortfolio{
				Name:             p.Name,
				Type:             string(p.Type),
				TotalMarketValue: fmt.Sprintf("%.2f %s", mv, p.BaseCurrency),
				TotalGainLoss:    fmt.Sprintf("%.2f", gl),
				TotalGainLossPct: fmt.Sprintf("%.2f", glp),
				GainLossColor:    color,
			})
		}

		if totalValue > 0 {
			totalGainPct = (totalGain / (totalValue - totalGain)) * 100
		}

		gainColor := "#22c97e"
		if totalGain < 0 {
			gainColor = "#e05a5a"
		}

		data := mail.WeeklySummaryData{
			UserName:         u.Name,
			TotalValue:       fmt.Sprintf("%.2f", totalValue),
			TotalGainLoss:    fmt.Sprintf("%.2f", totalGain),
			TotalGainLossPct: fmt.Sprintf("%.2f", totalGainPct),
			GainLossColor:    gainColor,
			Portfolios:       portfolios,
			DashboardURL:     s.cfg.PublicURL + "/dashboard",
			WeekLabel:        weekLabel,
		}

		if err := s.mail.SendWeeklySummary(u.Email, data); err != nil {
			errs = append(errs, fmt.Errorf("user %s: %w", u.ID, err))
			continue
		}
		sent++
	}

	return sent, errs
}

func (s *Services) GetPortfolioGrowth(ctx context.Context, userID uuid.UUID, period string) ([]entities.PortfolioGrowthPoint, entities.PortfolioGrowthSummary, error) {
	hasSince, since := parsePeriod(period)
	points, err := s.repos.GetPortfolioGrowthByUserID(ctx, userID, hasSince, since)
	if err != nil {
		return nil, entities.PortfolioGrowthSummary{}, err
	}
	return points, buildGrowthSummary(points), nil
}

func parsePeriod(period string) (bool, time.Time) {
	now := time.Now().UTC()
	switch period {
	case "1M":
		return true, now.AddDate(0, -1, 0)
	case "3M":
		return true, now.AddDate(0, -3, 0)
	case "6M":
		return true, now.AddDate(0, -6, 0)
	case "1Y":
		return true, now.AddDate(-1, 0, 0)
	default:
		return false, time.Time{}
	}
}

func buildGrowthSummary(points []entities.PortfolioGrowthPoint) entities.PortfolioGrowthSummary {
	if len(points) == 0 {
		return entities.PortfolioGrowthSummary{}
	}
	first, last := points[0], points[len(points)-1]
	initialVal, _ := strconv.ParseFloat(first.TotalValue, 64)
	currentVal, _ := strconv.ParseFloat(last.TotalValue, 64)
	var growthPct float64
	if initialVal > 0 {
		growthPct = ((currentVal - initialVal) / initialVal) * 100
	}
	return entities.PortfolioGrowthSummary{
		FirstDate:      first.Date,
		InitialValue:   strconv.FormatFloat(initialVal, 'f', 2, 64),
		CurrentValue:   strconv.FormatFloat(currentVal, 'f', 2, 64),
		TotalGrowthPct: strconv.FormatFloat(growthPct, 'f', 2, 64),
	}
}
