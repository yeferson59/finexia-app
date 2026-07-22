package market

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/platform/logger"
	"github.com/yeferson59/finexia-app/internal/platform/marketdata"
	"github.com/yeferson59/finexia-app/internal/portfolio"
)

// fakeRepository embeds the Repository interface so tests only override the
// methods a scenario needs; calling anything else panics loudly.
type fakeRepository struct {
	Repository

	upsertExchangeRate func(ctx context.Context, from, to string, rate money.Decimal, rateDate time.Time) (ExchangeRate, error)
}

func (f *fakeRepository) UpsertExchangeRate(ctx context.Context, from, to string, rate money.Decimal, rateDate time.Time) (ExchangeRate, error) {
	return f.upsertExchangeRate(ctx, from, to, rate, rateDate)
}

// fakePortfolioService stubs the portfolioService slice market depends on for
// asset lookups/updates.
type fakePortfolioService struct {
	updateAssetPrice     func(ctx context.Context, assetID uuid.UUID, price money.Money) (portfolio.Asset, error)
	createAsset          func(ctx context.Context, ticker, name string, assetType portfolio.AssetType, exchange, currency string) (portfolio.Asset, error)
	getAssets            func(ctx context.Context, offset, limit uint) ([]portfolio.Asset, error)
	getAssetByID         func(ctx context.Context, assetID uuid.UUID) (portfolio.Asset, error)
	importAssetsFromFile func(ctx context.Context, data []byte, filename, sheet string) (portfolio.ImportResultResponseDTO, error)
}

// Every method defaults to a no-op success when its hook is unset: scenarios
// that only care about one code path (e.g. the default-asset seeding loop)
// should not have to stub every collaborator call along the way.
func (p *fakePortfolioService) UpdateAssetPrice(ctx context.Context, assetID uuid.UUID, price money.Money) (portfolio.Asset, error) {
	if p.updateAssetPrice == nil {
		return portfolio.Asset{}, nil
	}
	return p.updateAssetPrice(ctx, assetID, price)
}

func (p *fakePortfolioService) CreateAsset(ctx context.Context, ticker, name string, assetType portfolio.AssetType, exchange, currency string) (portfolio.Asset, error) {
	if p.createAsset == nil {
		return portfolio.Asset{}, nil
	}
	return p.createAsset(ctx, ticker, name, assetType, exchange, currency)
}

func (p *fakePortfolioService) GetAssets(ctx context.Context, offset, limit uint) ([]portfolio.Asset, error) {
	if p.getAssets == nil {
		return nil, nil
	}
	return p.getAssets(ctx, offset, limit)
}

func (p *fakePortfolioService) GetAssetByID(ctx context.Context, assetID uuid.UUID) (portfolio.Asset, error) {
	if p.getAssetByID == nil {
		return portfolio.Asset{}, nil
	}
	return p.getAssetByID(ctx, assetID)
}

func (p *fakePortfolioService) ImportAssetsFromFile(ctx context.Context, data []byte, filename, sheet string) (portfolio.ImportResultResponseDTO, error) {
	if p.importAssetsFromFile == nil {
		return portfolio.ImportResultResponseDTO{}, nil
	}
	return p.importAssetsFromFile(ctx, data, filename, sheet)
}

// fakePriceProvider stubs the market data provider used by the sync jobs.
type fakePriceProvider struct {
	fetchQuote        func(ctx context.Context, symbol string) (marketdata.QuoteResult, error)
	fetchExchangeRate func(ctx context.Context, from, to string) (marketdata.ExchangeRateResult, error)
}

func (p *fakePriceProvider) FetchQuote(ctx context.Context, symbol string) (marketdata.QuoteResult, error) {
	return p.fetchQuote(ctx, symbol)
}

func (p *fakePriceProvider) FetchExchangeRate(ctx context.Context, from, to string) (marketdata.ExchangeRateResult, error) {
	return p.fetchExchangeRate(ctx, from, to)
}

// memStorage is an in-memory fiber.Storage that honours TTLs, good enough to
// exercise the sync-marker caching logic without Redis.
type memStorage struct {
	mu    sync.Mutex
	items map[string]memItem
}

type memItem struct {
	value     []byte
	expiresAt time.Time
}

func newMemStorage() *memStorage {
	return &memStorage{items: map[string]memItem{}}
}

func (s *memStorage) GetWithContext(_ context.Context, key string) ([]byte, error) {
	return s.Get(key)
}

func (s *memStorage) Get(key string) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.items[key]
	if !ok {
		return nil, nil
	}
	if !item.expiresAt.IsZero() && time.Now().After(item.expiresAt) {
		delete(s.items, key)
		return nil, nil
	}
	return item.value, nil
}

func (s *memStorage) SetWithContext(_ context.Context, key string, val []byte, exp time.Duration) error {
	return s.Set(key, val, exp)
}

func (s *memStorage) Set(key string, val []byte, exp time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	item := memItem{value: append([]byte(nil), val...)}
	if exp > 0 {
		item.expiresAt = time.Now().Add(exp)
	}
	s.items[key] = item
	return nil
}

func (s *memStorage) DeleteWithContext(_ context.Context, key string) error {
	return s.Delete(key)
}

func (s *memStorage) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.items, key)
	return nil
}

func (s *memStorage) ResetWithContext(_ context.Context) error {
	return s.Reset()
}

func (s *memStorage) Reset() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = map[string]memItem{}
	return nil
}

func (s *memStorage) Close() error { return nil }

// mustUSD parses amount as a USD money.Money, failing the test on error.
func mustUSD(t *testing.T, amount string) money.Money {
	t.Helper()

	m, err := money.NewMoneyFromString(amount, money.USD)
	if err != nil {
		t.Fatalf("mustUSD(%q): %v", amount, err)
	}

	return m
}

func newTestServices(repo Repository, storage *memStorage) *Service {
	return NewService(repo, nil, storage, nil, logger.Noop())
}

// newTestServicesFull wires a price provider and portfolio service in
// addition to the repository, for flows that hit market data or asset
// lookups. The unused third parameter mirrors the other modules' test
// helper shape; market has nothing to plug in there today.
func newTestServicesFull(repo Repository, storage *memStorage, _ any, provider marketdata.Provider, port portfolioService) *Service {
	return NewService(repo, port, storage, provider, logger.Noop())
}
