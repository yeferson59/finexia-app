package app

import (
	"crypto/rand"
	"encoding/base64"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yeferson59/finexia-app/internal/platform/cache"
	"github.com/yeferson59/finexia-app/internal/platform/config"
	"github.com/yeferson59/finexia-app/internal/platform/logger"
	"github.com/yeferson59/finexia-app/internal/platform/mail"
	"github.com/yeferson59/finexia-app/internal/platform/secretbox"
)

// TestNewValidatesRequiredDeps checks that New fails fast with a clear error
// when a required dependency is missing, instead of panicking later in wire().
func TestNewValidatesRequiredDeps(t *testing.T) {
	if _, err := New(Deps{}); err == nil {
		t.Fatal("expected New to reject an empty Deps, got nil error")
	}

	// Envs present but DB missing must still fail (Envs is dereferenced in
	// New; the rest are only needed later, so validation must catch them up
	// front).
	if _, err := New(Deps{Envs: new(config.EnvConfig{})}); err == nil {
		t.Fatal("expected New to reject Deps missing DB/Storage/Mail/Log, got nil error")
	}
}

// TestAppWiresAndRoutes is the boot smoke test of the composition root: it
// composes the real App (pgx pool is lazy, so no database is needed) and
// checks that public routes, module routes and the JWT gate all answer.
func TestAppWiresAndRoutes(t *testing.T) {
	ctx := t.Context() // cancelled on cleanup, which stops the schedulers started by wire

	pool, err := pgxpool.New(ctx, "postgres://user:pass@127.0.0.1:1/finexia_test")
	if err != nil {
		t.Fatalf("pgxpool.New: %v", err)
	}
	defer pool.Close()

	mailService, err := mail.New("", "test@example.com")
	if err != nil {
		t.Fatalf("mail.New: %v", err)
	}

	a, err := New(Deps{
		Envs: new(config.EnvConfig{
			Port:               "0",
			Environment:        "test",
			JWTSecret:          "test-secret",
			JWTAccessDuration:  15 * time.Minute,
			JWTRefreshDuration: 30 * 24 * time.Hour,
			PublicURL:          "http://localhost:8080",
			CORSOrigin:         []string{"http://localhost:5173"},
		}),
		DB:      pool,
		Cache:   cache.Conn("127.0.0.1", "1", "", 0),
		S3:      nil,
		Mail:    mailService,
		Keyring: testKeyring(t),
		Log:     logger.Noop(),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
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

	// The admin invitation/waitlist dashboard is gated by inline guards (401
	// with a bogus token, before RequireAdmin runs). /users/waitlist is served
	// by marketing and /users/invitations by user, so this also proves both
	// modules apply the shared guards themselves.
	if status := request("GET", "/users/waitlist", "Authorization", "Bearer bogus-token"); status != fiber.StatusUnauthorized {
		t.Errorf("GET /users/waitlist = %d, want 401 with an invalid token", status)
	}
	if status := request("GET", "/users/invitations", "Authorization", "Bearer bogus-token"); status != fiber.StatusUnauthorized {
		t.Errorf("GET /users/invitations = %d, want 401 with an invalid token", status)
	}

	// The 401s above stop at the JWT gate, so on their own they say nothing
	// about the role guard sitting behind it. That is how the waitlist listing
	// — every sign-up's email address — lost its admin guard in a refactor and
	// stayed open to any signed-in user without a test noticing.
	//
	// The composition root cannot mint a valid token here, so what is asserted
	// is the wiring rather than a live response: the admin routes must carry a
	// handler after the gate. Their own module tests cover the behaviour.
	assertGuarded := func(method, path string, want int) {
		t.Helper()
		for _, stack := range a.fiber.Stack() {
			for _, route := range stack {
				if route.Method == method && route.Path == path {
					if len(route.Handlers) < want {
						t.Errorf("%s %s has %d handlers, want at least %d (auth gate, limiter, role guard, handler)", method, path, len(route.Handlers), want)
					}

					return
				}
			}
		}

		t.Errorf("%s %s is not registered", method, path)
	}

	// RequireAuth + limiter + RequireAdmin + paginate + handler.
	assertGuarded(fiber.MethodGet, "/users/waitlist", 5)

	// The avatar route (docs/API.md §2.3) is public: without a token the
	// request must get past the JWT gate and reach the handler chain.
	if status := request("GET", "/users/0b7f9c7e-1111-4222-8333-444455556666/avatar"); status == fiber.StatusBadRequest || status == fiber.StatusUnauthorized {
		t.Errorf("GET /users/:id/avatar = %d; the route must stay public (no JWT gate)", status)
	}

	// The self-service password change is served by the auth module, which owns
	// the credentials; it must still answer at its /users path, behind the gate.
	if status := request("PATCH", "/users/me/password", "Authorization", "Bearer bogus-token"); status != fiber.StatusUnauthorized {
		t.Errorf("PATCH /users/me/password = %d, want 401 with an invalid token", status)
	}

	// Fiber matches routes in registration order: the static /users paths must
	// register before the user module's "/users/:id" or they get captured by
	// it (400 "invalid uuid" instead of the intended handler). Three modules
	// register under /users — auth (password, invitations), marketing
	// (waitlist) and user — so the ordering is a property of mountRoutes, not
	// of any single module.
	routesUnder := func(method, prefix string) []string {
		var paths []string
		for _, stack := range a.fiber.Stack() {
			for _, route := range stack {
				if route.Method == method && strings.HasPrefix(route.Path, prefix) {
					paths = append(paths, route.Path)
				}
			}
		}
		return paths
	}
	assertBefore := func(method, static, parametric string) {
		t.Helper()
		paths := routesUnder(method, "/users")
		index := func(path string) int {
			for i, p := range paths {
				if p == path {
					return i
				}
			}
			return -1
		}
		if index(static) == -1 {
			t.Errorf("%s %s is not registered (%s /users* routes: %v)", method, static, method, paths)
			return
		}
		if index(parametric) != -1 && index(static) > index(parametric) {
			t.Errorf("%s %s must register before %s %s (%s /users* routes: %v)", method, static, method, parametric, method, paths)
		}
	}

	if len(routesUnder(fiber.MethodGet, "/users/:id/avatar")) == 0 {
		t.Error("GET /users/:id/avatar is not registered")
	}
	assertBefore(fiber.MethodGet, "/users/invitations", "/users/:id")
	assertBefore(fiber.MethodGet, "/users/waitlist", "/users/:id")
	assertBefore(fiber.MethodPatch, "/users/me/password", "/users/:id")
}

// testKeyring builds the market-credential keyring the composition root now
// requires. New refuses to build without one, deliberately: a default would
// mean sealing users' provider keys under a guessable key.
func testKeyring(t *testing.T) *secretbox.Keyring {
	t.Helper()

	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("rand: %v", err)
	}

	ring, err := secretbox.NewKeyring([]string{"1:" + base64.StdEncoding.EncodeToString(key)}, 1)
	if err != nil {
		t.Fatalf("NewKeyring: %v", err)
	}

	return ring
}
