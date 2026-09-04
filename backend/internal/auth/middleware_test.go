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

// TestRequireMCPAuth exercises the guard /mcp is mounted behind. It is the
// only guard in the app that accepts two credentials, so what it has to pin
// down is that each one is answered by its own path and that both leave the
// same identity behind — everything downstream reads the locals, not the
// header.
func TestRequireMCPAuth(t *testing.T) {
	cfg := testConfig()
	jwtUserID, mcpUserID := uuid.New(), uuid.New()

	accessToken, err := newService(testStores(new(fakeRepository{})), cfg, newMemStorage(), nil, nil, logger.Noop()).
		CreateJWToken(jwtUserID, "user", time.Now().UTC().Add(time.Hour))
	if err != nil {
		t.Fatalf("CreateJWToken: %v", err)
	}

	mcpToken, mcpHash, _, err := generateMCPToken()
	if err != nil {
		t.Fatalf("generateMCPToken: %v", err)
	}

	// probeApp echoes the authenticated identity from the locals, which is the
	// only thing the MCP transport behind this guard ever reads.
	probeApp := func() *fiber.App {
		repo := new(fakeRepository{
			getSessionByToken: func(context.Context, string) (identity.User, error) {
				return identity.User{
					ID:       jwtUserID,
					Role:     identity.Role{Name: "user"},
					Sessions: []identity.Session{{Token: accessToken}},
				}, nil
			},
			getMCPTokenByHash: func(_ context.Context, hash string) (mcpTokenIdentity, error) {
				if hash != mcpHash {
					return mcpTokenIdentity{}, ErrMCPTokenNotFound
				}

				return mcpTokenIdentity{ID: uuid.New(), UserID: mcpUserID, Role: "admin"}, nil
			},
			touchMCPToken: func(context.Context, uuid.UUID) error { return nil },
		})

		service := newService(testStores(repo), cfg, newMemStorage(), nil, nil, logger.Noop())
		m := newModule(Deps{Cfg: cfg, Storage: newMemStorage(), Log: logger.Noop()}, service)

		app := fiber.New()
		app.Get("/probe", m.MCPAuth().RequireAuth(), func(c fiber.Ctx) error {
			userID, _, role, err := httpx.Identity(c)
			if err != nil {
				return c.SendStatus(fiber.StatusInternalServerError)
			}

			return c.SendString(userID.String() + " " + role)
		})

		return app
	}

	for _, tc := range []struct {
		name       string
		bearer     string
		wantStatus int
		wantBody   string
	}{
		{
			name:   "a personal access token authenticates its owner",
			bearer: mcpToken, wantStatus: fiber.StatusOK,
			wantBody: mcpUserID.String() + " admin",
		},
		{
			name:   "a session access token still works",
			bearer: accessToken, wantStatus: fiber.StatusOK,
			wantBody: jwtUserID.String() + " user",
		},
		{
			// Answered by the MCP path, not the JWT one: it claimed the prefix.
			name:   "an unknown personal access token is a 401",
			bearer: mcpTokenPrefix + "not-a-real-token", wantStatus: fiber.StatusUnauthorized,
		},
		{
			name:   "a garbage bearer is still rejected",
			bearer: "nonsense", wantStatus: fiber.StatusUnauthorized,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/probe", nil)
			req.Header.Set("Authorization", "Bearer "+tc.bearer)

			resp, err := probeApp().Test(req)
			if err != nil {
				t.Fatalf("app.Test: %v", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != tc.wantStatus {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tc.wantStatus)
			}

			if tc.wantBody == "" {
				return
			}

			body, _ := io.ReadAll(resp.Body)
			if string(body) != tc.wantBody {
				t.Errorf("identity = %q, want %q", body, tc.wantBody)
			}
		})
	}
}
