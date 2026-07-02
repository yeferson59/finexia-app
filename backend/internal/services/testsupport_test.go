package services

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/yeferson59/finexia-app/internal/config"
	"github.com/yeferson59/finexia-app/internal/entities"
)

// fakeRepository embeds the Repository interface so tests only override the
// methods a scenario needs; calling anything else panics loudly.
type fakeRepository struct {
	Repository

	getAccountByEmail                func(ctx context.Context, email string) (entities.User, error)
	getAccountByUserID               func(ctx context.Context, userID uuid.UUID) (entities.Account, error)
	createSession                    func(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) (uuid.UUID, error)
	updateSessionToken               func(ctx context.Context, sessionID uuid.UUID, newToken string, expiresAt time.Time) (string, error)
	createRefreshToken               func(ctx context.Context, userID uuid.UUID, tokenHash string, familyID, sessionID uuid.UUID, ip, ua *string, expiresAt time.Time) (uuid.UUID, error)
	getRefreshTokenByHash            func(ctx context.Context, tokenHash string) (entities.RefreshToken, error)
	markRefreshTokenUsed             func(ctx context.Context, id uuid.UUID) error
	revokeRefreshTokenFamily         func(ctx context.Context, familyID uuid.UUID) ([]string, error)
	getRefreshTokenFamiliesBySession func(ctx context.Context, userID uuid.UUID, sessionToken string) ([]string, []uuid.UUID, error)
	register                         func(ctx context.Context, name, email, password string) (entities.User, error)
	getSessionByToken                func(ctx context.Context, token string) (entities.User, error)
	deleteSessionByUserIDToken       func(ctx context.Context, userID uuid.UUID, token string) error
	deleteExpiredRefreshTokens       func(ctx context.Context) (int64, error)
	deleteExpiredSessions            func(ctx context.Context) (int64, error)
	getUserByEmail                   func(ctx context.Context, email string) (entities.User, error)
	getUserByID                      func(ctx context.Context, id uuid.UUID) (entities.User, error)
	updateUser                       func(ctx context.Context, id uuid.UUID, name, email, image string) (entities.User, error)
	updateUserProfile                func(ctx context.Context, id uuid.UUID, name, preferredCurrency, image string) (entities.User, error)
	updateUserPassword               func(ctx context.Context, userID uuid.UUID, hashedPassword string) error
	countAssetTransactions           func(ctx context.Context, userID, portfolioID uuid.UUID, ticker string) (int, error)
	getAssetTransactionsPaginated    func(ctx context.Context, userID, portfolioID uuid.UUID, ticker string, limit, offset int) ([]entities.Transaction, error)
	getPortfolioGrowthByUserID       func(ctx context.Context, userID uuid.UUID, hasSince bool, since time.Time) ([]entities.PortfolioGrowthPoint, error)
}

func (f *fakeRepository) GetAccountByEmail(ctx context.Context, email string) (entities.User, error) {
	return f.getAccountByEmail(ctx, email)
}

func (f *fakeRepository) GetAccountByUserID(ctx context.Context, userID uuid.UUID) (entities.Account, error) {
	return f.getAccountByUserID(ctx, userID)
}

func (f *fakeRepository) CreateSession(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) (uuid.UUID, error) {
	return f.createSession(ctx, userID, token, expiresAt)
}

func (f *fakeRepository) UpdateSessionToken(ctx context.Context, sessionID uuid.UUID, newToken string, expiresAt time.Time) (string, error) {
	return f.updateSessionToken(ctx, sessionID, newToken, expiresAt)
}

func (f *fakeRepository) CreateRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, familyID, sessionID uuid.UUID, ip, ua *string, expiresAt time.Time) (uuid.UUID, error) {
	return f.createRefreshToken(ctx, userID, tokenHash, familyID, sessionID, ip, ua, expiresAt)
}

func (f *fakeRepository) GetRefreshTokenByHash(ctx context.Context, tokenHash string) (entities.RefreshToken, error) {
	return f.getRefreshTokenByHash(ctx, tokenHash)
}

func (f *fakeRepository) MarkRefreshTokenUsed(ctx context.Context, id uuid.UUID) error {
	return f.markRefreshTokenUsed(ctx, id)
}

func (f *fakeRepository) RevokeRefreshTokenFamily(ctx context.Context, familyID uuid.UUID) ([]string, error) {
	return f.revokeRefreshTokenFamily(ctx, familyID)
}

func (f *fakeRepository) GetRefreshTokenFamiliesBySession(ctx context.Context, userID uuid.UUID, sessionToken string) ([]string, []uuid.UUID, error) {
	return f.getRefreshTokenFamiliesBySession(ctx, userID, sessionToken)
}

func (f *fakeRepository) Register(ctx context.Context, name, email, password string) (entities.User, error) {
	return f.register(ctx, name, email, password)
}

func (f *fakeRepository) GetSessionByToken(ctx context.Context, token string) (entities.User, error) {
	return f.getSessionByToken(ctx, token)
}

func (f *fakeRepository) DeleteSessionByUserIDToken(ctx context.Context, userID uuid.UUID, token string) error {
	return f.deleteSessionByUserIDToken(ctx, userID, token)
}

func (f *fakeRepository) DeleteExpiredRefreshTokens(ctx context.Context) (int64, error) {
	return f.deleteExpiredRefreshTokens(ctx)
}

func (f *fakeRepository) DeleteExpiredSessions(ctx context.Context) (int64, error) {
	return f.deleteExpiredSessions(ctx)
}

func (f *fakeRepository) GetUserByEmail(ctx context.Context, email string) (entities.User, error) {
	return f.getUserByEmail(ctx, email)
}

func (f *fakeRepository) GetUserByID(ctx context.Context, id uuid.UUID) (entities.User, error) {
	return f.getUserByID(ctx, id)
}

func (f *fakeRepository) UpdateUser(ctx context.Context, id uuid.UUID, name, email, image string) (entities.User, error) {
	return f.updateUser(ctx, id, name, email, image)
}

func (f *fakeRepository) UpdateUserProfile(ctx context.Context, id uuid.UUID, name, preferredCurrency, image string) (entities.User, error) {
	return f.updateUserProfile(ctx, id, name, preferredCurrency, image)
}

func (f *fakeRepository) UpdateUserPassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error {
	return f.updateUserPassword(ctx, userID, hashedPassword)
}

func (f *fakeRepository) CountAssetTransactions(ctx context.Context, userID, portfolioID uuid.UUID, ticker string) (int, error) {
	return f.countAssetTransactions(ctx, userID, portfolioID, ticker)
}

func (f *fakeRepository) GetAssetTransactionsPaginated(ctx context.Context, userID, portfolioID uuid.UUID, ticker string, limit, offset int) ([]entities.Transaction, error) {
	return f.getAssetTransactionsPaginated(ctx, userID, portfolioID, ticker, limit, offset)
}

func (f *fakeRepository) GetPortfolioGrowthByUserID(ctx context.Context, userID uuid.UUID, hasSince bool, since time.Time) ([]entities.PortfolioGrowthPoint, error) {
	return f.getPortfolioGrowthByUserID(ctx, userID, hasSince, since)
}

// memStorage is an in-memory fiber.Storage that honours TTLs, good enough to
// exercise the auth caching logic without Redis.
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

func (s *memStorage) has(key string) bool {
	v, _ := s.Get(key)
	return v != nil
}

func testConfig() *config.Env {
	return &config.Env{
		JWTSecret:          "test-secret",
		JWTAccessDuration:  15 * time.Minute,
		JWTRefreshDuration: 30 * 24 * time.Hour,
		RefreshGracePeriod: 30 * time.Second,
		PublicURL:          "http://localhost:8080",
	}
}

func newTestServices(repo Repository, storage *memStorage) *Services {
	svc := New(repo, testConfig(), nil, storage, nil, nil, nil)
	return &svc
}
