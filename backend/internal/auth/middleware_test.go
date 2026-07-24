package auth

import (
	"context"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/yeferson59/finexia-app/internal/identity"
	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

// TestRequireAuth exercises the JWT gate end to end: with the session check now
// running in the success handler (using the request context, TECH_DEBT #6), a
// valid token must reach the route with the identity in locals, and a token the
// session store rejects must yield 401.
func TestRequireAuth(t *testing.T) {
	userID := uuid.New()
	cfg := testConfig()

	// Sign one token; any service built with the same secret accepts it.
	signer := NewService(testStores(&fakeRepository{}), cfg, newMemStorage(), nil, nil, logger.Noop())
	token, err := signer.CreateJWToken(userID, "user", time.Now().UTC().Add(time.Hour))
	if err != nil {
		t.Fatalf("CreateJWToken: %v", err)
	}

	// probeApp mounts RequireAuth on a route that echoes the authenticated user
	// id from locals. A fresh storage per app keeps ValidateToken's positive
	// cache from leaking between subtests.
	probeApp := func(sessionUser identity.User, sessionErr error) *fiber.App {
		repo := &fakeRepository{
			getSessionByToken: func(context.Context, string) (identity.User, error) {
				return sessionUser, sessionErr
			},
		}
		service := NewService(testStores(repo), cfg, newMemStorage(), nil, nil, logger.Noop())
		m := newModule(Deps{Cfg: cfg, Storage: newMemStorage(), Log: logger.Noop()}, service)

		app := fiber.New()
		app.Get("/probe", m.RequireAuth(), func(c fiber.Ctx) error {
			return c.SendString(c.Locals(LocalUserID).(string))
		})
		return app
	}

	t.Run("valid token passes and populates locals", func(t *testing.T) {
		app := probeApp(identity.User{
			ID:       userID,
			Role:     identity.Role{Name: "user"},
			Sessions: []identity.Session{{Token: token}},
		}, nil)

		req := httptest.NewRequest("GET", "/probe", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
		body, _ := io.ReadAll(resp.Body)
		if string(body) != userID.String() {
			t.Errorf("locals user id = %q, want %q", body, userID.String())
		}
	})

	t.Run("session store rejection yields 401", func(t *testing.T) {
		app := probeApp(identity.User{}, ErrSessionNotFound)

		req := httptest.NewRequest("GET", "/probe", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != fiber.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", resp.StatusCode)
		}
	})
}

// newRBACApp builds an app that fakes the JWT middleware by injecting the
// given role into locals before the RBAC handler runs.
func newRBACApp(role string, handler fiber.Handler) *fiber.App {
	app := fiber.New()
	app.Get("/protected", func(c fiber.Ctx) error {
		if role != "" {
			c.Locals(LocalRole, role)
		}
		return c.Next()
	}, handler, func(c fiber.Ctx) error {
		return c.SendString("ok")
	})
	return app
}

func TestRequireRole(t *testing.T) {
	m := &Module{}

	cases := []struct {
		name       string
		role       string
		allowed    []string
		wantStatus int
	}{
		{"matching role passes", "admin", []string{"admin"}, fiber.StatusOK},
		{"role in list passes", "editor", []string{"admin", "editor"}, fiber.StatusOK},
		{"other role is forbidden", "user", []string{"admin"}, fiber.StatusForbidden},
		{"missing role is forbidden", "", []string{"admin"}, fiber.StatusForbidden},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			app := newRBACApp(tc.role, m.RequireRole(tc.allowed...))

			resp, err := app.Test(httptest.NewRequest("GET", "/protected", nil))
			if err != nil {
				t.Fatalf("app.Test: %v", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != tc.wantStatus {
				t.Errorf("status = %d, want %d", resp.StatusCode, tc.wantStatus)
			}
		})
	}
}

func TestRequireAdmin(t *testing.T) {
	m := &Module{}

	t.Run("admin passes", func(t *testing.T) {
		app := newRBACApp("admin", m.RequireAdmin())
		resp, err := app.Test(httptest.NewRequest("GET", "/protected", nil))
		if err != nil {
			t.Fatalf("app.Test: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("status = %d, want 200", resp.StatusCode)
		}
	})

	t.Run("regular user is forbidden", func(t *testing.T) {
		app := newRBACApp("user", m.RequireAdmin())
		resp, err := app.Test(httptest.NewRequest("GET", "/protected", nil))
		if err != nil {
			t.Fatalf("app.Test: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode != fiber.StatusForbidden {
			t.Errorf("status = %d, want 403", resp.StatusCode)
		}
	})
}
