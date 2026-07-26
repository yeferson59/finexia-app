package market

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/platform/logger"
	"github.com/yeferson59/finexia-app/internal/platform/marketdata"
	"github.com/yeferson59/finexia-app/internal/platform/secretbox"
)

// fakeRepository embeds the Repository interface so tests only override the
// methods a scenario needs; calling anything else panics loudly. The asset
// hooks default to a no-op success when unset, so seeding loops don't have to
// stub every collaborator call along the way.
type fakeRepository struct {
	Repository
	// creds backs the CredentialStore half. Nil in scenarios that do not touch
	// BYO-key, in which case those methods panic like any other unstubbed one.
	creds *credentialStore

	upsertExchangeRate func(ctx context.Context, from, to string, rate money.Decimal, rateDate time.Time) (ExchangeRate, error)

	updateAssetPrice func(ctx context.Context, assetID uuid.UUID, price money.Money) (Asset, error)
	upsertAsset      func(ctx context.Context, ticker, name string, assetType AssetType, exchange, currency string) (Asset, error)
	getAssets        func(ctx context.Context, offset, limit uint) ([]Asset, error)
	getAssetByID     func(ctx context.Context, assetID uuid.UUID) (Asset, error)
	searchAssets     func(ctx context.Context, search string, offset, limit uint) ([]Asset, error)
}

func (f *fakeRepository) UpsertExchangeRate(ctx context.Context, from, to string, rate money.Decimal, rateDate time.Time) (ExchangeRate, error) {
	return f.upsertExchangeRate(ctx, from, to, rate, rateDate)
}

func (f *fakeRepository) UpdateAssetPrice(ctx context.Context, assetID uuid.UUID, price money.Money) (Asset, error) {
	if f.updateAssetPrice == nil {
		return Asset{}, nil
	}
	return f.updateAssetPrice(ctx, assetID, price)
}

func (f *fakeRepository) UpsertAsset(ctx context.Context, ticker, name string, assetType AssetType, exchange, currency string) (Asset, error) {
	if f.upsertAsset == nil {
		return Asset{}, nil
	}
	return f.upsertAsset(ctx, ticker, name, assetType, exchange, currency)
}

func (f *fakeRepository) GetAssets(ctx context.Context, offset, limit uint) ([]Asset, error) {
	if f.getAssets == nil {
		return nil, nil
	}
	return f.getAssets(ctx, offset, limit)
}

func (f *fakeRepository) GetAssetByID(ctx context.Context, assetID uuid.UUID) (Asset, error) {
	if f.getAssetByID == nil {
		return Asset{}, nil
	}
	return f.getAssetByID(ctx, assetID)
}

func (f *fakeRepository) SearchAssets(ctx context.Context, search string, offset, limit uint) ([]Asset, error) {
	if f.searchAssets == nil {
		return nil, nil
	}
	return f.searchAssets(ctx, search, offset, limit)
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
	return newService(repo, storage, nil, testKeyring(), logger.Noop())
}

// fakeFactory stands in for marketdata/providers. It ignores the keys and
// returns a canned provider, so tests exercise the sync logic rather than the
// chain assembly (which providers has its own tests for).
type fakeFactory struct {
	provider marketdata.Provider
	// gotCreds records what the service handed over, so a test can assert the
	// right user's keys were opened.
	gotCreds []marketdata.Credential
	err      error
}

func (f *fakeFactory) For(creds []marketdata.Credential) (marketdata.Provider, error) {
	f.gotCreds = creds

	if f.err != nil {
		return nil, f.err
	}
	if len(creds) == 0 {
		return nil, marketdata.ErrNoCredentials
	}

	return f.provider, nil
}

// testKeyring builds a real keyring over a throwaway key: the sealing path is
// cheap and exercising it for real is worth more than stubbing it.
func testKeyring() *secretbox.Keyring {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		panic("rand: " + err.Error())
	}

	ring, err := secretbox.NewKeyring("1:"+base64.StdEncoding.EncodeToString(key), "1")
	if err != nil {
		panic("keyring: " + err.Error())
	}

	return ring
}

// credentialStore is an in-memory CredentialStore, enough to drive the BYO-key
// sync without Postgres.
type credentialStore struct {
	mu     sync.Mutex
	sealed map[uuid.UUID][]sealedCredential
	status map[string]CredentialStatus
	prices map[string]money.Money
	rates  map[string]money.Decimal
}

func newCredentialStore() *credentialStore {
	return &credentialStore{
		sealed: map[uuid.UUID][]sealedCredential{},
		status: map[string]CredentialStatus{},
		prices: map[string]money.Money{},
		rates:  map[string]money.Decimal{},
	}
}

// seed stores a key for a user, sealed exactly as production would.
func (c *credentialStore) seed(t *testing.T, ring *secretbox.Keyring, userID uuid.UUID, provider ProviderID, apiKey string) {
	t.Helper()

	sealed, err := ring.Seal([]byte(apiKey), credentialAAD(userID.String(), string(provider)))
	if err != nil {
		t.Fatalf("seal: %v", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.sealed[userID] = append(c.sealed[userID], sealedCredential{Provider: provider, Sealed: sealed})
}

func (c *credentialStore) GetSealedCredentials(_ context.Context, userID uuid.UUID) ([]sealedCredential, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.sealed[userID], nil
}

func (c *credentialStore) GetSealedCredential(_ context.Context, userID uuid.UUID, provider ProviderID) (sealedCredential, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, sc := range c.sealed[userID] {
		if sc.Provider == provider {
			return sc, nil
		}
	}

	return sealedCredential{}, ErrCredentialNotFound
}

func (c *credentialStore) SetCredentialStatus(_ context.Context, userID uuid.UUID, provider ProviderID, status CredentialStatus, _ string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.status[userID.String()+"/"+string(provider)] = status

	return nil
}

func (c *credentialStore) statusOf(userID uuid.UUID, provider ProviderID) CredentialStatus {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.status[userID.String()+"/"+string(provider)]
}

func (c *credentialStore) UpsertUserAssetPrice(_ context.Context, userID, assetID uuid.UUID, price money.Money, _ string, _ ProviderID, _ time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.prices[userID.String()+"/"+assetID.String()] = price

	return nil
}

func (c *credentialStore) priceOf(userID, assetID uuid.UUID) (money.Money, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	p, ok := c.prices[userID.String()+"/"+assetID.String()]

	return p, ok
}

func (c *credentialStore) UpsertUserExchangeRate(_ context.Context, userID uuid.UUID, from, to string, rate money.Decimal, _ ProviderID, _ time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.rates[userID.String()+"/"+from+to] = rate

	return nil
}

func (c *credentialStore) UsersWithCredentials(context.Context) ([]uuid.UUID, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	ids := make([]uuid.UUID, 0, len(c.sealed))
	for id := range c.sealed {
		ids = append(ids, id)
	}

	return ids, nil
}

// The remaining CredentialStore methods are not exercised by the sync tests.
func (c *credentialStore) UpsertCredential(context.Context, uuid.UUID, sealedCredential, string, *time.Time) (Credential, error) {
	panic("UpsertCredential not stubbed")
}

func (c *credentialStore) ListCredentials(context.Context, uuid.UUID) ([]Credential, error) {
	panic("ListCredentials not stubbed")
}

func (c *credentialStore) DeleteCredential(context.Context, uuid.UUID, ProviderID) error {
	panic("DeleteCredential not stubbed")
}

// The CredentialStore half of Repository, forwarded to the in-memory store.
// Explicit forwarding rather than embedding: embedding both Repository and
// CredentialStore would make every one of these selectors ambiguous.

func (f *fakeRepository) UpsertCredential(ctx context.Context, userID uuid.UUID, cred sealedCredential, keyLast4 string, verifiedAt *time.Time) (Credential, error) {
	return f.creds.UpsertCredential(ctx, userID, cred, keyLast4, verifiedAt)
}

func (f *fakeRepository) ListCredentials(ctx context.Context, userID uuid.UUID) ([]Credential, error) {
	return f.creds.ListCredentials(ctx, userID)
}

func (f *fakeRepository) GetSealedCredentials(ctx context.Context, userID uuid.UUID) ([]sealedCredential, error) {
	return f.creds.GetSealedCredentials(ctx, userID)
}

func (f *fakeRepository) GetSealedCredential(ctx context.Context, userID uuid.UUID, provider ProviderID) (sealedCredential, error) {
	return f.creds.GetSealedCredential(ctx, userID, provider)
}

func (f *fakeRepository) DeleteCredential(ctx context.Context, userID uuid.UUID, provider ProviderID) error {
	return f.creds.DeleteCredential(ctx, userID, provider)
}

func (f *fakeRepository) SetCredentialStatus(ctx context.Context, userID uuid.UUID, provider ProviderID, status CredentialStatus, lastErr string) error {
	return f.creds.SetCredentialStatus(ctx, userID, provider, status, lastErr)
}

func (f *fakeRepository) UsersWithCredentials(ctx context.Context) ([]uuid.UUID, error) {
	return f.creds.UsersWithCredentials(ctx)
}

func (f *fakeRepository) UpsertUserAssetPrice(ctx context.Context, userID, assetID uuid.UUID, price money.Money, currency string, source ProviderID, fetchedAt time.Time) error {
	return f.creds.UpsertUserAssetPrice(ctx, userID, assetID, price, currency, source, fetchedAt)
}

func (f *fakeRepository) UpsertUserExchangeRate(ctx context.Context, userID uuid.UUID, from, to string, rate money.Decimal, source ProviderID, fetchedAt time.Time) error {
	return f.creds.UpsertUserExchangeRate(ctx, userID, from, to, rate, source, fetchedAt)
}
