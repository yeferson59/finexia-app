package handlers

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/yeferson59/finexia-app/internal/config"
	"github.com/yeferson59/finexia-app/internal/entities"
	"github.com/yeferson59/finexia-app/internal/logger"
	"github.com/yeferson59/finexia-app/internal/repositories"
	"github.com/yeferson59/finexia-app/internal/services"
)

// authStubRepository embeds services.Repository so each test only wires the
// methods its endpoint touches; anything else panics loudly.
type authStubRepository struct {
	services.Repository

	getAccountByEmail     func(ctx context.Context, email string) (entities.User, error)
	getTwoFactor          func(ctx context.Context, userID uuid.UUID) (entities.TwoFactor, error)
	hasKnownLoginIP       func(ctx context.Context, userID uuid.UUID, ip string) (bool, error)
	recordKnownLoginIP    func(ctx context.Context, userID uuid.UUID, ip string) error
	createSession         func(ctx context.Context, userID uuid.UUID, token string, ip, ua *string, expiresAt time.Time) (uuid.UUID, error)
	createRefreshToken    func(ctx context.Context, userID uuid.UUID, tokenHash string, familyID, sessionID uuid.UUID, ip, ua *string, expiresAt time.Time) (uuid.UUID, error)
	getRefreshTokenByHash func(ctx context.Context, tokenHash string) (entities.RefreshToken, error)
	markRefreshTokenUsed  func(ctx context.Context, id uuid.UUID) error
	updateSessionToken    func(ctx context.Context, sessionID uuid.UUID, newToken string, expiresAt time.Time) (string, error)
}

func (s *authStubRepository) GetAccountByEmail(ctx context.Context, email string) (entities.User, error) {
	return s.getAccountByEmail(ctx, email)
}

func (s *authStubRepository) GetTwoFactor(ctx context.Context, userID uuid.UUID) (entities.TwoFactor, error) {
	return s.getTwoFactor(ctx, userID)
}

func (s *authStubRepository) HasKnownLoginIP(ctx context.Context, userID uuid.UUID, ip string) (bool, error) {
	return s.hasKnownLoginIP(ctx, userID, ip)
}

func (s *authStubRepository) RecordKnownLoginIP(ctx context.Context, userID uuid.UUID, ip string) error {
	return s.recordKnownLoginIP(ctx, userID, ip)
}

func (s *authStubRepository) CreateSession(ctx context.Context, userID uuid.UUID, token string, ip, ua *string, expiresAt time.Time) (uuid.UUID, error) {
	return s.createSession(ctx, userID, token, ip, ua, expiresAt)
}

func (s *authStubRepository) CreateRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, familyID, sessionID uuid.UUID, ip, ua *string, expiresAt time.Time) (uuid.UUID, error) {
	return s.createRefreshToken(ctx, userID, tokenHash, familyID, sessionID, ip, ua, expiresAt)
}

func (s *authStubRepository) GetRefreshTokenByHash(ctx context.Context, tokenHash string) (entities.RefreshToken, error) {
	return s.getRefreshTokenByHash(ctx, tokenHash)
}

func (s *authStubRepository) MarkRefreshTokenUsed(ctx context.Context, id uuid.UUID) error {
	return s.markRefreshTokenUsed(ctx, id)
}

func (s *authStubRepository) UpdateSessionToken(ctx context.Context, sessionID uuid.UUID, newToken string, expiresAt time.Time) (string, error) {
	return s.updateSessionToken(ctx, sessionID, newToken, expiresAt)
}

// memStorage is a minimal in-memory fiber.Storage so the auth handlers can
// exercise their caching/lockout logic without Redis.
type memStorage struct {
	mu    sync.Mutex
	items map[string][]byte
}

func newMemStorage() *memStorage {
	return &memStorage{items: map[string][]byte{}}
}

func (s *memStorage) GetWithContext(_ context.Context, key string) ([]byte, error) {
	return s.Get(key)
}

func (s *memStorage) Get(key string) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.items[key], nil
}

func (s *memStorage) SetWithContext(_ context.Context, key string, val []byte, exp time.Duration) error {
	return s.Set(key, val, exp)
}

func (s *memStorage) Set(key string, val []byte, _ time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[key] = append([]byte(nil), val...)
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
	s.items = map[string][]byte{}
	return nil
}

func (s *memStorage) Close() error { return nil }

func authTestConfig() *config.Env {
	return &config.Env{
		JWTSecret:          "test-secret",
		JWTAccessDuration:  15 * time.Minute,
		JWTRefreshDuration: 30 * 24 * time.Hour,
		RefreshGracePeriod: 30 * time.Second,
		MaxLoginAttempts:   5,
		LoginLockout:       time.Minute,
		PublicURL:          "http://localhost:8080",
	}
}

func newAuthTestHandlers(repo services.Repository, cfg *config.Env) *Handlers {
	svc := services.New(repo, cfg, nil, newMemStorage(), nil, nil, logger.Noop(), nil)
	h := New(svc, cfg)
	return &h
}

func TestLoginHandlerSmoke(t *testing.T) {
	userID := uuid.New()
	hash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("bcrypt: %v", err)
	}

	newRepo := func() *authStubRepository {
		return &authStubRepository{
			getAccountByEmail: func(_ context.Context, email string) (entities.User, error) {
				if email != "user@example.com" {
					return entities.User{}, errors.New("account not found")
				}
				return entities.User{
					ID:            userID,
					Name:          "Test User",
					EmailVerified: true,
					Role:          entities.Role{Name: "user"},
					Accounts:      []entities.Account{{Password: string(hash)}},
				}, nil
			},
			getTwoFactor: func(context.Context, uuid.UUID) (entities.TwoFactor, error) {
				return entities.TwoFactor{}, repositories.ErrTwoFactorNotFound
			},
			hasKnownLoginIP: func(context.Context, uuid.UUID, string) (bool, error) {
				return true, nil
			},
			recordKnownLoginIP: func(context.Context, uuid.UUID, string) error { return nil },
			createSession: func(context.Context, uuid.UUID, string, *string, *string, time.Time) (uuid.UUID, error) {
				return uuid.New(), nil
			},
			createRefreshToken: func(context.Context, uuid.UUID, string, uuid.UUID, uuid.UUID, *string, *string, time.Time) (uuid.UUID, error) {
				return uuid.New(), nil
			},
		}
	}

	t.Run("logs in with valid credentials and sets the refresh cookie", func(t *testing.T) {
		h := newAuthTestHandlers(newRepo(), authTestConfig())
		app := fiber.New()
		app.Post("/auth/login", h.Login)

		req := httptest.NewRequest("POST", "/auth/login",
			strings.NewReader(`{"email":"user@example.com","password":"correct-password"}`))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}

		var refreshCookie *http.Cookie
		for _, c := range resp.Cookies() {
			if c.Name == "refresh_token" {
				refreshCookie = c
			}
		}
		if refreshCookie == nil || refreshCookie.Value == "" {
			t.Error("refresh_token cookie not set")
		}
	})

	t.Run("rejects invalid credentials with 400", func(t *testing.T) {
		h := newAuthTestHandlers(newRepo(), authTestConfig())
		app := fiber.New()
		app.Post("/auth/login", h.Login)

		_, status, payload := doJSON(t, app, "POST", "/auth/login",
			`{"email":"user@example.com","password":"wrong-password"}`)
		if status != fiber.StatusBadRequest {
			t.Errorf("status = %d, want 400", status)
		}
		if success, _ := payload["success"].(bool); success {
			t.Error("success should be false")
		}
	})
}

func TestRefreshHandlerSmoke(t *testing.T) {
	t.Run("rejects a request without the refresh cookie", func(t *testing.T) {
		h := newAuthTestHandlers(&authStubRepository{}, authTestConfig())
		app := fiber.New()
		app.Post("/auth/refresh", h.Refresh)

		_, status, payload := doJSON(t, app, "POST", "/auth/refresh", "")
		if status != fiber.StatusUnauthorized {
			t.Errorf("status = %d, want 401", status)
		}
		if success, _ := payload["success"].(bool); success {
			t.Error("success should be false")
		}
	})

	t.Run("rotates a valid refresh token", func(t *testing.T) {
		userID := uuid.New()
		// The raw cookie value must be URL-safe base64: the service decodes it
		// before hashing and rejects anything else as an invalid token.
		rawToken := base64.URLEncoding.EncodeToString(bytes.Repeat([]byte{1}, 32))
		repo := &authStubRepository{
			getRefreshTokenByHash: func(context.Context, string) (entities.RefreshToken, error) {
				return entities.RefreshToken{
					ID:        uuid.New(),
					UserID:    userID,
					FamilyID:  uuid.New(),
					SessionID: uuid.New(),
					ExpiresAt: time.Now().UTC().Add(time.Hour),
					Role:      "user",
				}, nil
			},
			markRefreshTokenUsed: func(context.Context, uuid.UUID) error { return nil },
			updateSessionToken: func(context.Context, uuid.UUID, string, time.Time) (string, error) {
				return "", nil
			},
			createRefreshToken: func(context.Context, uuid.UUID, string, uuid.UUID, uuid.UUID, *string, *string, time.Time) (uuid.UUID, error) {
				return uuid.New(), nil
			},
		}
		h := newAuthTestHandlers(repo, authTestConfig())
		app := fiber.New()
		app.Post("/auth/refresh", h.Refresh)

		req := httptest.NewRequest("POST", "/auth/refresh", nil)
		req.AddCookie(&http.Cookie{Name: "refresh_token", Value: rawToken})
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}

		var rotated *http.Cookie
		for _, c := range resp.Cookies() {
			if c.Name == "refresh_token" {
				rotated = c
			}
		}
		if rotated == nil || rotated.Value == "" || rotated.Value == rawToken {
			t.Error("refresh_token cookie should be rotated to a new value")
		}
	})
}

func TestRegisterHandlerSmoke(t *testing.T) {
	t.Run("returns 403 while self-registration is disabled", func(t *testing.T) {
		cfg := authTestConfig()
		cfg.SelfRegistrationEnabled = false
		h := newAuthTestHandlers(&authStubRepository{}, cfg)
		app := fiber.New()
		app.Post("/auth/register", h.Register)

		_, status, payload := doJSON(t, app, "POST", "/auth/register",
			`{"name":"New User","email":"new@example.com","password":"secret-password"}`)
		if status != fiber.StatusForbidden {
			t.Errorf("status = %d, want 403", status)
		}
		if action, _ := payload["action"].(string); action != "auth:register:disabled" {
			t.Errorf("action = %q, want auth:register:disabled", action)
		}
	})
}
