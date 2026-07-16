package auth

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/yeferson59/finexia-app/internal/identity"
	"github.com/yeferson59/finexia-app/internal/platform/config"
	"github.com/yeferson59/finexia-app/internal/platform/logger"
	"github.com/yeferson59/finexia-app/internal/platform/mail"
)

// fakeRepository implements every store interface of the module through
// per-method hooks, so tests only set the methods a scenario needs; calling
// anything else panics loudly (nil hook). Unlike the old services fake, it no
// longer embeds a 90-method god interface: only auth's methods exist here.
type fakeRepository struct {
	getAccountByEmail                func(ctx context.Context, email string) (identity.User, error)
	getAccountByUserID               func(ctx context.Context, userID uuid.UUID) (identity.Account, error)
	createSession                    func(ctx context.Context, userID uuid.UUID, token string, ip, ua *string, expiresAt time.Time) (uuid.UUID, error)
	updateSessionToken               func(ctx context.Context, sessionID uuid.UUID, newToken string, expiresAt time.Time) (string, error)
	updateSessionLocation            func(ctx context.Context, sessionID uuid.UUID, location string) error
	listSessionsByUserID             func(ctx context.Context, userID uuid.UUID) ([]identity.Session, error)
	getRefreshTokensBySessionIDs     func(ctx context.Context, userID uuid.UUID, sessionIDs []uuid.UUID) ([]string, []uuid.UUID, error)
	deleteSessionsByIDs              func(ctx context.Context, userID uuid.UUID, sessionIDs []uuid.UUID) (int64, error)
	hasKnownLoginIP                  func(ctx context.Context, userID uuid.UUID, ip string) (bool, error)
	recordKnownLoginIP               func(ctx context.Context, userID uuid.UUID, ip string) error
	createRefreshToken               func(ctx context.Context, userID uuid.UUID, tokenHash string, familyID, sessionID uuid.UUID, ip, ua *string, expiresAt time.Time) (uuid.UUID, error)
	getRefreshTokenByHash            func(ctx context.Context, tokenHash string) (identity.RefreshToken, error)
	markRefreshTokenUsed             func(ctx context.Context, id uuid.UUID) error
	revokeRefreshTokenFamily         func(ctx context.Context, familyID uuid.UUID) ([]string, error)
	getRefreshTokenFamiliesBySession func(ctx context.Context, userID uuid.UUID, sessionToken string) ([]string, []uuid.UUID, error)
	register                         func(ctx context.Context, name, email, password string) (identity.User, error)
	getSessionByUserIDToken          func(ctx context.Context, userID uuid.UUID, token string) (identity.User, error)
	getSessionByToken                func(ctx context.Context, token string) (identity.User, error)
	deleteSessionByUserIDToken       func(ctx context.Context, userID uuid.UUID, token string) error
	deleteExpiredRefreshTokens       func(ctx context.Context) (int64, error)
	deleteExpiredSessions            func(ctx context.Context) (int64, error)
	getUserByEmail                   func(ctx context.Context, email string) (identity.User, error)
	getUserByID                      func(ctx context.Context, id uuid.UUID) (identity.User, error)

	getTwoFactor                      func(ctx context.Context, userID uuid.UUID) (TwoFactor, error)
	upsertTwoFactorSecret             func(ctx context.Context, userID uuid.UUID, secret string) error
	enableTwoFactor                   func(ctx context.Context, userID uuid.UUID) error
	deleteTwoFactor                   func(ctx context.Context, userID uuid.UUID) error
	replaceTwoFactorRecoveryCodes     func(ctx context.Context, userID uuid.UUID, codeHashes []string) error
	consumeTwoFactorRecoveryCode      func(ctx context.Context, userID uuid.UUID, codeHash string) error
	countUnusedTwoFactorRecoveryCodes func(ctx context.Context, userID uuid.UUID) (int, error)

	createEmailVerification    func(ctx context.Context, email, tokenHash string, expiresAt time.Time) (Verification, error)
	getEmailVerificationByHash func(ctx context.Context, tokenHash string) (Verification, error)
	consumeEmailVerification   func(ctx context.Context, id uuid.UUID, email string) error
}

var (
	_ AccountStore      = (*fakeRepository)(nil)
	_ SessionStore      = (*fakeRepository)(nil)
	_ RefreshTokenStore = (*fakeRepository)(nil)
	_ TwoFactorStore    = (*fakeRepository)(nil)
	_ VerificationStore = (*fakeRepository)(nil)
)

// testStores fills every store slot with the same fake, mirroring how the
// composition root wires the single Postgres implementation.
func testStores(f *fakeRepository) Stores {
	return Stores{
		Accounts:      f,
		Sessions:      f,
		RefreshTokens: f,
		TwoFactor:     f,
		Verifications: f,
	}
}

func (f *fakeRepository) GetAccountByEmail(ctx context.Context, email string) (identity.User, error) {
	return f.getAccountByEmail(ctx, email)
}

func (f *fakeRepository) GetAccountByUserID(ctx context.Context, userID uuid.UUID) (identity.Account, error) {
	return f.getAccountByUserID(ctx, userID)
}

func (f *fakeRepository) CreateSession(ctx context.Context, userID uuid.UUID, token string, ip, ua *string, expiresAt time.Time) (uuid.UUID, error) {
	return f.createSession(ctx, userID, token, ip, ua, expiresAt)
}

func (f *fakeRepository) UpdateSessionToken(ctx context.Context, sessionID uuid.UUID, newToken string, expiresAt time.Time) (string, error) {
	return f.updateSessionToken(ctx, sessionID, newToken, expiresAt)
}

// UpdateSessionLocation runs in a fire-and-forget goroutine, so a nil hook
// must be a no-op instead of a panic that would kill the test process.
func (f *fakeRepository) UpdateSessionLocation(ctx context.Context, sessionID uuid.UUID, location string) error {
	if f.updateSessionLocation == nil {
		return nil
	}
	return f.updateSessionLocation(ctx, sessionID, location)
}

func (f *fakeRepository) ListSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]identity.Session, error) {
	return f.listSessionsByUserID(ctx, userID)
}

func (f *fakeRepository) GetRefreshTokensBySessionIDs(ctx context.Context, userID uuid.UUID, sessionIDs []uuid.UUID) ([]string, []uuid.UUID, error) {
	return f.getRefreshTokensBySessionIDs(ctx, userID, sessionIDs)
}

func (f *fakeRepository) DeleteSessionsByIDs(ctx context.Context, userID uuid.UUID, sessionIDs []uuid.UUID) (int64, error) {
	return f.deleteSessionsByIDs(ctx, userID, sessionIDs)
}

func (f *fakeRepository) HasKnownLoginIP(ctx context.Context, userID uuid.UUID, ip string) (bool, error) {
	return f.hasKnownLoginIP(ctx, userID, ip)
}

// RecordKnownLoginIP runs in a fire-and-forget goroutine, so a nil hook must
// be a no-op instead of a panic that would kill the test process.
func (f *fakeRepository) RecordKnownLoginIP(ctx context.Context, userID uuid.UUID, ip string) error {
	if f.recordKnownLoginIP == nil {
		return nil
	}
	return f.recordKnownLoginIP(ctx, userID, ip)
}

func (f *fakeRepository) CreateRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, familyID, sessionID uuid.UUID, ip, ua *string, expiresAt time.Time) (uuid.UUID, error) {
	return f.createRefreshToken(ctx, userID, tokenHash, familyID, sessionID, ip, ua, expiresAt)
}

func (f *fakeRepository) GetRefreshTokenByHash(ctx context.Context, tokenHash string) (identity.RefreshToken, error) {
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

func (f *fakeRepository) Register(ctx context.Context, name, email, password string) (identity.User, error) {
	return f.register(ctx, name, email, password)
}

func (f *fakeRepository) GetSessionByUserIDToken(ctx context.Context, userID uuid.UUID, token string) (identity.User, error) {
	return f.getSessionByUserIDToken(ctx, userID, token)
}

func (f *fakeRepository) GetSessionByToken(ctx context.Context, token string) (identity.User, error) {
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

func (f *fakeRepository) GetUserByEmail(ctx context.Context, email string) (identity.User, error) {
	return f.getUserByEmail(ctx, email)
}

func (f *fakeRepository) GetUserByID(ctx context.Context, id uuid.UUID) (identity.User, error) {
	return f.getUserByID(ctx, id)
}

func (f *fakeRepository) GetTwoFactor(ctx context.Context, userID uuid.UUID) (TwoFactor, error) {
	if f.getTwoFactor == nil {
		return TwoFactor{}, ErrTwoFactorNotFound
	}
	return f.getTwoFactor(ctx, userID)
}

func (f *fakeRepository) UpsertTwoFactorSecret(ctx context.Context, userID uuid.UUID, secret string) error {
	return f.upsertTwoFactorSecret(ctx, userID, secret)
}

func (f *fakeRepository) EnableTwoFactor(ctx context.Context, userID uuid.UUID) error {
	return f.enableTwoFactor(ctx, userID)
}

func (f *fakeRepository) DeleteTwoFactor(ctx context.Context, userID uuid.UUID) error {
	return f.deleteTwoFactor(ctx, userID)
}

func (f *fakeRepository) ReplaceTwoFactorRecoveryCodes(ctx context.Context, userID uuid.UUID, codeHashes []string) error {
	return f.replaceTwoFactorRecoveryCodes(ctx, userID, codeHashes)
}

func (f *fakeRepository) ConsumeTwoFactorRecoveryCode(ctx context.Context, userID uuid.UUID, codeHash string) error {
	return f.consumeTwoFactorRecoveryCode(ctx, userID, codeHash)
}

func (f *fakeRepository) CountUnusedTwoFactorRecoveryCodes(ctx context.Context, userID uuid.UUID) (int, error) {
	return f.countUnusedTwoFactorRecoveryCodes(ctx, userID)
}

func (f *fakeRepository) CreateEmailVerification(ctx context.Context, email, tokenHash string, expiresAt time.Time) (Verification, error) {
	return f.createEmailVerification(ctx, email, tokenHash, expiresAt)
}

func (f *fakeRepository) GetEmailVerificationByHash(ctx context.Context, tokenHash string) (Verification, error) {
	return f.getEmailVerificationByHash(ctx, tokenHash)
}

func (f *fakeRepository) ConsumeEmailVerification(ctx context.Context, id uuid.UUID, email string) error {
	return f.consumeEmailVerification(ctx, id, email)
}

// fakeMailer records outbound emails so tests can assert on the alert flows
// without a Resend client. It grows with the module: password reset and
// invitation capture arrive with their sub-areas.
type fakeMailer struct {
	mu sync.Mutex

	securityErr          error
	emailVerificationErr error

	security []struct {
		To   string
		Data mail.SecurityAlertData
	}
	emailVerificationTo []struct {
		To   string
		Data mail.EmailVerificationData
	}
}

func (m *fakeMailer) SendSecurityAlert(email string, data mail.SecurityAlertData) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.securityErr != nil {
		return m.securityErr
	}
	m.security = append(m.security, struct {
		To   string
		Data mail.SecurityAlertData
	}{email, data})
	return nil
}

func (m *fakeMailer) SendEmailVerification(email string, data mail.EmailVerificationData) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.emailVerificationErr != nil {
		return m.emailVerificationErr
	}
	m.emailVerificationTo = append(m.emailVerificationTo, struct {
		To   string
		Data mail.EmailVerificationData
	}{email, data})
	return nil
}

func (m *fakeMailer) emailVerificationCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.emailVerificationTo)
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
		JWTSecret:              "test-secret",
		JWTAccessDuration:      15 * time.Minute,
		JWTRefreshDuration:     30 * 24 * time.Hour,
		RefreshGracePeriod:     30 * time.Second,
		PublicURL:              "http://localhost:8080",
		TwoFactorPendingExpiry: 5 * time.Minute,
	}
}

func newTestService(repo *fakeRepository, storage *memStorage) *Service {
	return NewService(testStores(repo), testConfig(), storage, nil, nil, logger.Noop())
}

// newTestServiceFull wires a fake mailer in addition to the repository, for
// flows that send email.
func newTestServiceFull(repo *fakeRepository, storage *memStorage, mailer Mailer) *Service {
	return NewService(testStores(repo), testConfig(), storage, mailer, nil, logger.Noop())
}

// notFound is a GetUserByEmail hook for scenarios where no account exists.
func notFound(context.Context, string) (identity.User, error) {
	return identity.User{}, errors.New("no rows")
}

// waitFor polls cond until it returns true or the deadline expires. Used to
// synchronise with the fire-and-forget alert goroutines.
func waitFor(t *testing.T, timeout time.Duration, cond func() bool) bool {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return true
		}
		time.Sleep(5 * time.Millisecond)
	}
	return cond()
}
