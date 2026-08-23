package auth

import (
	"context"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"uuid"

	"github.com/gofiber/fiber/v3"

	"github.com/yeferson59/finexia-app/internal/identity"
	"github.com/yeferson59/finexia-app/internal/platform/httpx"
	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

// TestRequireAuth exercises the JWT gate end to end: with the session check
// running in the success handler against the request context, a valid token
// must reach the route with the identity in locals, and a token the session
// store rejects must yield 401.
func TestRequireAuth(t *testing.T) {
	userID := uuid.New()
	cfg := testConfig()

	// Sign one token; any service built with the same secret accepts it.
	signer := newService(testStores(new(fakeRepository{})), cfg, newMemStorage(), nil, nil, logger.Noop())
	token, err := signer.CreateJWToken(userID, "user", time.Now().UTC().Add(time.Hour))
	if err != nil {
		t.Fatalf("CreateJWToken: %v", err)
	}

	// probeApp mounts RequireAuth on a route that echoes the authenticated user
	// id from locals. A fresh storage per app keeps ValidateToken's positive
	// cache from leaking between subtests.
	probeApp := func(sessionUser identity.User, sessionErr error) *fiber.App {
		repo := new(fakeRepository{
			getSessionByToken: func(context.Context, string) (identity.User, error) {
				return sessionUser, sessionErr
			},
		})
		service := newService(testStores(repo), cfg, newMemStorage(), nil, nil, logger.Noop())
		m := newModule(Deps{Cfg: cfg, Storage: newMemStorage(), Log: logger.Noop()}, service)

		app := fiber.New()
		app.Get("/probe", m.RequireAuth(), func(c fiber.Ctx) error {
			return c.SendString(c.Locals(httpx.LocalUserID).(string))
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
