package portfolio

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// stubAuth injects the request locals the JWT middleware would normally set.
// When authenticated is false, RequireAuth passes the request through without
// locals so handlers hit the "missing authenticated identity" branch.
type stubAuth struct {
	userID        uuid.UUID
	role          string
	authenticated bool
}

func (a stubAuth) RequireAuth() fiber.Handler {
	return func(c fiber.Ctx) error {
		if a.authenticated {
			c.Locals(LocalUserID, a.userID.String())
			c.Locals(LocalToken, "test-token")
			c.Locals(LocalRole, a.role)
		}
		return c.Next()
	}
}

func (a stubAuth) RequireAdmin() fiber.Handler {
	return func(c fiber.Ctx) error {
		if a.role != "admin" {
			return c.SendStatus(fiber.StatusForbidden)
		}
		return c.Next()
	}
}

// newTestModule mounts the portfolio routes on a fresh Fiber app, backed by
// the given repository and an authenticated user.
func newTestModule(t *testing.T, repo *fakeRepository, userID uuid.UUID, role string) *fiber.App {
	t.Helper()
	svc := newTestServices(repo, newMemStorage())
	noopLimiter := func(c fiber.Ctx) error { return c.Next() }
	mod := newModule(Deps{
		AuthMiddl: stubAuth{userID: userID, role: role, authenticated: true},
		Limiter:   noopLimiter,
	}, svc)

	app := fiber.New()
	mod.Routes(app)
	return app
}

func do(t *testing.T, app *fiber.App, method, target string) *http.Response {
	t.Helper()
	resp, err := app.Test(httptest.NewRequest(method, target, nil))
	if err != nil {
		t.Fatalf("%s %s: %v", method, target, err)
	}
	return resp
}

func decodeEnvelope(t *testing.T, resp *http.Response) (bool, json.RawMessage) {
	t.Helper()
	var env struct {
		Success bool            `json:"success"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	return env.Success, env.Data
}

func TestHandlerGetPortfoliosRisks(t *testing.T) {
	repo := &fakeRepository{
		getPortfoliosRisks: func(context.Context) ([]Risk, error) {
			return []Risk{{ID: uuid.New(), Name: "conservative"}}, nil
		},
	}
	app := newTestModule(t, repo, uuid.New(), "user")

	resp := do(t, app, http.MethodGet, "/portfolios/risks")
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	ok, data := decodeEnvelope(t, resp)
	if !ok {
		t.Errorf("success = false, want true")
	}
	var risks []Risk
	if err := json.Unmarshal(data, &risks); err != nil || len(risks) != 1 || risks[0].Name != "conservative" {
		t.Errorf("data = %s (err %v)", data, err)
	}
}

func TestHandlerGetPortfolios(t *testing.T) {
	userID := uuid.New()
	repo := &fakeRepository{
		getPortfoliosByUserID: func(_ context.Context, uid uuid.UUID) ([]Portfolio, error) {
			if uid != userID {
				t.Errorf("userID = %s, want %s", uid, userID)
			}
			return []Portfolio{{ID: uuid.New(), Name: "Growth"}}, nil
		},
	}
	app := newTestModule(t, repo, userID, "user")

	// The list route lives at the atypical "/portfolios/id" path (TECH_DEBT #3).
	resp := do(t, app, http.MethodGet, "/portfolios/id")
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestHandlerGetPortfolioByID(t *testing.T) {
	userID := uuid.New()
	portfolioID := uuid.New()

	t.Run("found", func(t *testing.T) {
		repo := &fakeRepository{
			getPortfolioByID: func(_ context.Context, pid, uid uuid.UUID) (Portfolio, error) {
				if pid != portfolioID || uid != userID {
					t.Errorf("ids = %s/%s, want %s/%s", pid, uid, portfolioID, userID)
				}
				return Portfolio{ID: portfolioID, Name: "Growth", UserID: userID}, nil
			},
			getEntriesByPortfolioID: func(context.Context, uuid.UUID) ([]Entry, error) {
				return nil, nil
			},
		}
		app := newTestModule(t, repo, userID, "user")

		resp := do(t, app, http.MethodGet, "/portfolios/"+portfolioID.String())
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
	})

	t.Run("invalid uuid returns 400", func(t *testing.T) {
		app := newTestModule(t, &fakeRepository{}, userID, "user")
		resp := do(t, app, http.MethodGet, "/portfolios/not-a-uuid")
		if resp.StatusCode != fiber.StatusBadRequest {
			t.Fatalf("status = %d, want 400", resp.StatusCode)
		}
	})

	t.Run("not found maps to 404", func(t *testing.T) {
		repo := &fakeRepository{
			getPortfolioByID: func(context.Context, uuid.UUID, uuid.UUID) (Portfolio, error) {
				return Portfolio{}, errors.New("portfolio not found")
			},
			// GetPortfolio fetches the header and entries concurrently, so this
			// hook runs even when the ownership check fails.
			getEntriesByPortfolioID: func(context.Context, uuid.UUID) ([]Entry, error) {
				return nil, nil
			},
		}
		app := newTestModule(t, repo, userID, "user")

		resp := do(t, app, http.MethodGet, "/portfolios/"+portfolioID.String())
		if resp.StatusCode != fiber.StatusNotFound {
			t.Fatalf("status = %d, want 404", resp.StatusCode)
		}
	})
}

func TestHandlerGetPortfoliosSummary(t *testing.T) {
	userID := uuid.New()

	t.Run("unsupported currency returns 400", func(t *testing.T) {
		app := newTestModule(t, &fakeRepository{}, userID, "user")
		resp := do(t, app, http.MethodGet, "/portfolios/summary?currency=EUR")
		if resp.StatusCode != fiber.StatusBadRequest {
			t.Fatalf("status = %d, want 400", resp.StatusCode)
		}
	})

	t.Run("supported currency converts and returns 200", func(t *testing.T) {
		repo := &fakeRepository{
			getPortfoliosSummaryByUserID: func(context.Context, uuid.UUID) ([]SummaryView, error) {
				return []SummaryView{{
					BaseCurrency: "USD", TotalMarketValue: "1000", TotalCostBase: "900",
					TotalGainLoss: "100", TotalGainLossPct: "11.11",
				}}, nil
			},
		}
		app := newTestModule(t, repo, userID, "user")

		// USD == base: conversion is skipped, no exchange-rate lookup needed.
		resp := do(t, app, http.MethodGet, "/portfolios/summary?currency=USD")
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
	})
}

func TestHandlerRejectsMissingIdentity(t *testing.T) {
	svc := newTestServices(&fakeRepository{}, newMemStorage())
	noopLimiter := func(c fiber.Ctx) error { return c.Next() }
	mod := newModule(Deps{
		AuthMiddl: stubAuth{authenticated: false},
		Limiter:   noopLimiter,
	}, svc)
	app := fiber.New()
	mod.Routes(app)

	// GetPortfolios needs the authenticated user; without locals it must 400.
	resp := do(t, app, http.MethodGet, "/portfolios/id")
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("status = %d, want 400 when identity is missing", resp.StatusCode)
	}
}
