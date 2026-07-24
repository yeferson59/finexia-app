package auth

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/yeferson59/finexia-app/internal/identity"
	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

func authTestConfig() Config {
	cfg := testConfig()
	cfg.MaxLoginAttempts = 5
	cfg.LoginLockout = time.Minute
	return cfg
}

// newTestApp mounts the whole module (public routes, rate limiter and the
// group-local RequireAuth gate) on a fresh Fiber app, exercising the same
// chain a real request traverses.
func newTestApp(repo *fakeRepository, cfg Config) *fiber.App {
	service := NewService(testStores(repo), cfg, newMemStorage(), nil, nil, logger.Noop())
	m := newModule(Deps{
		Cfg:     cfg,
		Storage: newMemStorage(),
		Log:     logger.Noop(),
	}, service)

	app := fiber.New()
	m.Routes(app)
	return app
}

func doJSON(t *testing.T, app *fiber.App, method, target, body string) (*fiber.App, int, map[string]any) {
	t.Helper()
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, target, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	raw, _ := io.ReadAll(resp.Body)
	var payload map[string]any
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &payload); err != nil {
			t.Fatalf("invalid JSON response %q: %v", raw, err)
		}
	}
	return app, resp.StatusCode, payload
}

func TestLoginHandlerSmoke(t *testing.T) {
	userID := uuid.New()
	hash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("bcrypt: %v", err)
	}

	newRepo := func() *fakeRepository {
		return &fakeRepository{
			getAccountByEmail: func(_ context.Context, email string) (identity.User, error) {
				if email != "user@example.com" {
					return identity.User{}, errors.New("account not found")
				}
				return identity.User{
					ID:            userID,
					Name:          "Test User",
					EmailVerified: true,
					Role:          identity.Role{Name: "user"},
					Accounts:      []identity.Account{{Password: string(hash)}},
				}, nil
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
		app := newTestApp(newRepo(), authTestConfig())

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
		app := newTestApp(newRepo(), authTestConfig())

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
		app := newTestApp(&fakeRepository{}, authTestConfig())

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
		repo := &fakeRepository{
			getRefreshTokenByHash: func(context.Context, string) (identity.RefreshToken, error) {
				return identity.RefreshToken{
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
		app := newTestApp(repo, authTestConfig())

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
		app := newTestApp(&fakeRepository{}, cfg)

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

func TestProtectedAuthRoutesRequireToken(t *testing.T) {
	// getSessionByToken hook: RequireAuth validates the bearer token against
	// the session store; an unknown token must yield 401.
	repo := &fakeRepository{
		getSessionByToken: func(context.Context, string) (identity.User, error) {
			return identity.User{}, errors.New("no rows")
		},
	}
	app := newTestApp(repo, authTestConfig())

	for _, target := range []string{"/auth/session", "/auth/sessions", "/auth/2fa"} {
		req := httptest.NewRequest("GET", target, nil)
		req.Header.Set("Authorization", "Bearer bogus-token")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test(%s): %v", target, err)
		}
		_ = resp.Body.Close()
		if resp.StatusCode != fiber.StatusUnauthorized {
			t.Errorf("GET %s status = %d, want 401", target, resp.StatusCode)
		}
	}
}
