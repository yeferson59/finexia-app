package app

import (
	"context"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yeferson59/finexia-app/internal/platform/config"
	"github.com/yeferson59/finexia-app/internal/platform/logger"
	"github.com/yeferson59/finexia-app/internal/platform/mail"
)

// memStorage is a minimal in-memory fiber.Storage so wiring needs no Redis.
type memStorage struct {
	mu    sync.Mutex
	items map[string][]byte
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
	if s.items == nil {
		s.items = map[string][]byte{}
	}
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

func (s *memStorage) ResetWithContext(_ context.Context) error { return s.Reset() }

func (s *memStorage) Reset() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = map[string][]byte{}
	return nil
}

func (s *memStorage) Close() error { return nil }

// TestAppWiresAndRoutes is the boot smoke test of the composition root: it
// composes the real App (pgx pool is lazy, so no database is needed) and
// checks that public routes, module routes and the JWT gate all answer.
func TestAppWiresAndRoutes(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // stops the schedulers started by wire

	pool, err := pgxpool.New(ctx, "postgres://user:pass@127.0.0.1:1/finexia_test")
	if err != nil {
		t.Fatalf("pgxpool.New: %v", err)
	}
	defer pool.Close()

	mailService, err := mail.New("", "test@example.com")
	if err != nil {
		t.Fatalf("mail.New: %v", err)
	}

	a := New(Deps{
		Envs: &config.Env{
			Port:               "0",
			Environment:        "test",
			JWTSecret:          "test-secret",
			JWTAccessDuration:  15 * time.Minute,
			JWTRefreshDuration: 30 * 24 * time.Hour,
			PublicURL:          "http://localhost:8080",
			CORSOrigin:         []string{"http://localhost:5173"},
		},
		DB:      pool,
		Storage: &memStorage{},
		S3:      nil,
		Mail:    mailService,
		Log:     logger.Noop(),
	})
	a.wire(ctx)

	request := func(method, target string, header ...string) int {
		t.Helper()
		req := httptest.NewRequest(method, target, nil)
		for i := 0; i+1 < len(header); i += 2 {
			req.Header.Set(header[i], header[i+1])
		}
		resp, err := a.fiber.Test(req)
		if err != nil {
			t.Fatalf("%s %s: %v", method, target, err)
		}
		defer func() { _ = resp.Body.Close() }()
		return resp.StatusCode
	}

	if status := request("GET", "/health/livez"); status != fiber.StatusOK {
		t.Errorf("GET /health/livez = %d, want 200", status)
	}

	// The marketing module route must exist (anything but 404/401 proves the
	// module is wired into the public zone; the empty body yields a 400).
	if status := request("POST", "/marketing/waitlists"); status != fiber.StatusBadRequest {
		t.Errorf("POST /marketing/waitlists = %d, want 400 for an empty body", status)
	}

	// A protected route with a bogus token must be stopped by the JWT gate
	// (401 comes from the middleware; the handler itself would answer 400).
	if status := request("GET", "/users/me", "Authorization", "Bearer bogus-token"); status != fiber.StatusUnauthorized {
		t.Errorf("GET /users/me = %d, want 401 with an invalid token", status)
	}

	// The auth module's public routes answer in the public zone (an empty
	// login body yields a 400, not a 401/404).
	if status := request("POST", "/auth/login"); status != fiber.StatusBadRequest {
		t.Errorf("POST /auth/login = %d, want 400 for an empty body", status)
	}

	// The auth module's own protected group is gated by its RequireAuth.
	if status := request("GET", "/auth/session", "Authorization", "Bearer bogus-token"); status != fiber.StatusUnauthorized {
		t.Errorf("GET /auth/session = %d, want 401 with an invalid token", status)
	}

	// The password-reset flow now answers from the module's public zone.
	if status := request("POST", "/auth/password-reset"); status != fiber.StatusBadRequest {
		t.Errorf("POST /auth/password-reset = %d, want 400 for an empty body", status)
	}

	// The public invitation flow answers from the module's public zone.
	if status := request("GET", "/auth/invitations"); status != fiber.StatusBadRequest {
		t.Errorf("GET /auth/invitations = %d, want 400 without a token", status)
	}

	// The admin invitation/waitlist dashboard is gated by the module's own
	// inline RequireAuth (401 with a bogus token, before RequireAdmin runs).
	if status := request("GET", "/users/waitlist", "Authorization", "Bearer bogus-token"); status != fiber.StatusUnauthorized {
		t.Errorf("GET /users/waitlist = %d, want 401 with an invalid token", status)
	}
	if status := request("GET", "/users/invitations", "Authorization", "Bearer bogus-token"); status != fiber.StatusUnauthorized {
		t.Errorf("GET /users/invitations = %d, want 401 with an invalid token", status)
	}
}
