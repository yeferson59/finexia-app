package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/yeferson59/gofinance/money"

	"github.com/yeferson59/finexia-app/internal/entities"
	"github.com/yeferson59/finexia-app/internal/middlewares"
	"github.com/yeferson59/finexia-app/internal/platform/config"
	"github.com/yeferson59/finexia-app/internal/platform/logger"
	"github.com/yeferson59/finexia-app/internal/services"
)

// stubRepository embeds services.Repository so each test only wires the
// methods its endpoint touches; anything else panics loudly.
type stubRepository struct {
	services.Repository

	getPortfoliosByUserID           func(ctx context.Context, userID uuid.UUID) ([]entities.Portfolio, error)
	getPortfolioByID                func(ctx context.Context, portfolioID, userID uuid.UUID) (entities.Portfolio, error)
	getEntriesByPortfolioID         func(ctx context.Context, portfolioID uuid.UUID) ([]entities.PortfolioEntry, error)
	createPortfolio                 func(ctx context.Context, userID uuid.UUID, name, description, baseCurrency string, riskID uuid.UUID, typePortfolio entities.PortfolioType, priceValue money.Money, isDefault bool) (entities.Portfolio, error)
	updatePortfolio                 func(ctx context.Context, userID, portfolioID uuid.UUID, name, description string, portfolioType entities.PortfolioType, riskID uuid.UUID, isDefault bool) (entities.Portfolio, error)
	createPlatform                  func(ctx context.Context, userID uuid.UUID, sourceType entities.SourceType, name, description string) (entities.InvestmentSource, error)
	createPortfolioEntry            func(ctx context.Context, userID, portfolioID, assetID, sourceID uuid.UUID, txnType entities.TransactionType, quantity money.Decimal, price money.Money, costCurrency string, category entities.PortfolioEntryCategory, entryDate time.Time, notes string) (entities.PortfolioEntry, error)
	createTransaction               func(ctx context.Context, userID, entryID uuid.UUID, txnType entities.TransactionType, quantity money.Decimal, price money.Money, currency string, fees money.Money, transactionDate time.Time, notes string) (entities.Transaction, error)
	updateTransaction               func(ctx context.Context, userID, txnID uuid.UUID, txnType entities.TransactionType, quantity money.Decimal, price money.Money, currency string, fees money.Money, transactionDate time.Time, notes string) (entities.Transaction, error)
	getUserPreferences              func(ctx context.Context, userID uuid.UUID) (entities.UserPreferences, error)
	getPortfolioGrowthByUserID      func(ctx context.Context, userID uuid.UUID, hasSince bool, since time.Time) ([]entities.PortfolioGrowthPoint, error)
	getPortfolioGrowthByPortfolioID func(ctx context.Context, userID, portfolioID uuid.UUID, hasSince bool, since time.Time) ([]entities.PortfolioGrowthPoint, error)
	getRecentTransactionsByUser     func(ctx context.Context, userID uuid.UUID, limit int) ([]entities.Transaction, error)
	getAssetAllocationByUserID      func(ctx context.Context, userID uuid.UUID) ([]entities.AllocationItem, error)
	getPortfoliosSummaryByUserID    func(ctx context.Context, userID uuid.UUID) ([]entities.PortfolioSummaryView, error)
	getExchangeRateByPair           func(ctx context.Context, from, to string) (entities.ExchangeRate, error)
}

func (s *stubRepository) GetPortfoliosByUserID(ctx context.Context, userID uuid.UUID) ([]entities.Portfolio, error) {
	return s.getPortfoliosByUserID(ctx, userID)
}

func (s *stubRepository) GetPortfolioByID(ctx context.Context, portfolioID, userID uuid.UUID) (entities.Portfolio, error) {
	return s.getPortfolioByID(ctx, portfolioID, userID)
}

func (s *stubRepository) GetEntriesByPortfolioID(ctx context.Context, portfolioID uuid.UUID) ([]entities.PortfolioEntry, error) {
	return s.getEntriesByPortfolioID(ctx, portfolioID)
}

func (s *stubRepository) CreatePortfolio(ctx context.Context, userID uuid.UUID, name, description, baseCurrency string, riskID uuid.UUID, typePortfolio entities.PortfolioType, priceValue money.Money, isDefault bool) (entities.Portfolio, error) {
	return s.createPortfolio(ctx, userID, name, description, baseCurrency, riskID, typePortfolio, priceValue, isDefault)
}

func (s *stubRepository) UpdatePortfolio(ctx context.Context, userID, portfolioID uuid.UUID, name, description string, portfolioType entities.PortfolioType, riskID uuid.UUID, isDefault bool) (entities.Portfolio, error) {
	return s.updatePortfolio(ctx, userID, portfolioID, name, description, portfolioType, riskID, isDefault)
}

func (s *stubRepository) CreatePlatform(ctx context.Context, userID uuid.UUID, sourceType entities.SourceType, name, description string) (entities.InvestmentSource, error) {
	return s.createPlatform(ctx, userID, sourceType, name, description)
}

func (s *stubRepository) CreatePortfolioEntry(ctx context.Context, userID, portfolioID, assetID uuid.UUID, sourceID uuid.UUID, txnType entities.TransactionType, quantity money.Decimal, price money.Money, costCurrency string, category entities.PortfolioEntryCategory, entryDate time.Time, notes string) (entities.PortfolioEntry, error) {
	return s.createPortfolioEntry(ctx, userID, portfolioID, assetID, sourceID, txnType, quantity, price, costCurrency, category, entryDate, notes)
}

func (s *stubRepository) CreateTransaction(ctx context.Context, userID, entryID uuid.UUID, txnType entities.TransactionType, quantity money.Decimal, price money.Money, currency string, fees money.Money, transactionDate time.Time, notes string) (entities.Transaction, error) {
	return s.createTransaction(ctx, userID, entryID, txnType, quantity, price, currency, fees, transactionDate, notes)
}

func (s *stubRepository) UpdateTransaction(ctx context.Context, userID, txnID uuid.UUID, txnType entities.TransactionType, quantity money.Decimal, price money.Money, currency string, fees money.Money, transactionDate time.Time, notes string) (entities.Transaction, error) {
	return s.updateTransaction(ctx, userID, txnID, txnType, quantity, price, currency, fees, transactionDate, notes)
}

func (s *stubRepository) GetUserPreferences(ctx context.Context, userID uuid.UUID) (entities.UserPreferences, error) {
	return s.getUserPreferences(ctx, userID)
}

func (s *stubRepository) GetPortfolioGrowthByUserID(ctx context.Context, userID uuid.UUID, hasSince bool, since time.Time) ([]entities.PortfolioGrowthPoint, error) {
	return s.getPortfolioGrowthByUserID(ctx, userID, hasSince, since)
}

func (s *stubRepository) GetPortfolioGrowthByPortfolioID(ctx context.Context, userID, portfolioID uuid.UUID, hasSince bool, since time.Time) ([]entities.PortfolioGrowthPoint, error) {
	return s.getPortfolioGrowthByPortfolioID(ctx, userID, portfolioID, hasSince, since)
}

func (s *stubRepository) GetRecentTransactionsByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]entities.Transaction, error) {
	return s.getRecentTransactionsByUser(ctx, userID, limit)
}

func (s *stubRepository) GetAssetAllocationByUserID(ctx context.Context, userID uuid.UUID) ([]entities.AllocationItem, error) {
	return s.getAssetAllocationByUserID(ctx, userID)
}

func (s *stubRepository) GetPortfoliosSummaryByUserID(ctx context.Context, userID uuid.UUID) ([]entities.PortfolioSummaryView, error) {
	return s.getPortfoliosSummaryByUserID(ctx, userID)
}

func (s *stubRepository) GetExchangeRateByPair(ctx context.Context, from, to string) (entities.ExchangeRate, error) {
	return s.getExchangeRateByPair(ctx, from, to)
}

func newTestHandlers(repo services.Repository) *Handlers {
	cfg := &config.Env{PublicURL: "http://localhost:8080"}
	svc := services.New(repo, cfg, nil, nil, nil, nil, logger.Noop(), nil)
	h := New(svc, cfg)
	return &h
}

// authed injects the locals the JWT middleware would normally set.
func authed(userID uuid.UUID) fiber.Handler {
	return func(c fiber.Ctx) error {
		c.Locals(middlewares.LocalUserID, userID.String())
		c.Locals(middlewares.LocalToken, "test-token")
		c.Locals(middlewares.LocalRole, "user")
		return c.Next()
	}
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

func TestGetPortfoliosHandler(t *testing.T) {
	userID := uuid.New()

	t.Run("rejects unauthenticated requests", func(t *testing.T) {
		h := newTestHandlers(&stubRepository{})
		app := fiber.New()
		app.Get("/portfolios", h.GetPortfolios)

		_, status, payload := doJSON(t, app, "GET", "/portfolios", "")
		if status != fiber.StatusBadRequest {
			t.Errorf("status = %d, want 400", status)
		}
		if success, _ := payload["success"].(bool); success {
			t.Error("success should be false")
		}
	})

	t.Run("returns the user's portfolios", func(t *testing.T) {
		repo := &stubRepository{
			getPortfoliosByUserID: func(_ context.Context, uid uuid.UUID) ([]entities.Portfolio, error) {
				if uid != userID {
					t.Errorf("userID = %s, want %s", uid, userID)
				}
				return []entities.Portfolio{{ID: uuid.New(), Name: "Main"}}, nil
			},
		}
		h := newTestHandlers(repo)
		app := fiber.New()
		app.Use(authed(userID))
		app.Get("/portfolios", h.GetPortfolios)

		_, status, payload := doJSON(t, app, "GET", "/portfolios", "")
		if status != fiber.StatusOK {
			t.Fatalf("status = %d, want 200", status)
		}
		data, _ := payload["data"].([]any)
		if len(data) != 1 {
			t.Errorf("data = %v, want one portfolio", payload["data"])
		}
	})

	t.Run("maps a not-found service error to 404", func(t *testing.T) {
		repo := &stubRepository{
			getPortfoliosByUserID: func(context.Context, uuid.UUID) ([]entities.Portfolio, error) {
				return nil, errors.New("portfolios not found")
			},
		}
		h := newTestHandlers(repo)
		app := fiber.New()
		app.Use(authed(userID))
		app.Get("/portfolios", h.GetPortfolios)

		_, status, _ := doJSON(t, app, "GET", "/portfolios", "")
		if status != fiber.StatusNotFound {
			t.Errorf("status = %d, want 404", status)
		}
	})
}

func TestGetPortfolioHandler(t *testing.T) {
	userID := uuid.New()

	t.Run("rejects a malformed portfolio ID", func(t *testing.T) {
		h := newTestHandlers(&stubRepository{})
		app := fiber.New()
		app.Use(authed(userID))
		app.Get("/portfolios/:id", h.GetPortfolio)

		_, status, _ := doJSON(t, app, "GET", "/portfolios/not-a-uuid", "")
		if status != fiber.StatusBadRequest {
			t.Errorf("status = %d, want 400", status)
		}
	})

	t.Run("returns the detail response with holdings", func(t *testing.T) {
		portfolioID := uuid.New()
		repo := &stubRepository{
			getPortfolioByID: func(_ context.Context, pid, uid uuid.UUID) (entities.Portfolio, error) {
				return entities.Portfolio{ID: pid, Name: "Main", Risk: entities.Risk{Name: "moderate"}}, nil
			},
			getEntriesByPortfolioID: func(context.Context, uuid.UUID) ([]entities.PortfolioEntry, error) {
				return []entities.PortfolioEntry{{ID: uuid.New(), Asset: entities.Asset{Ticker: "AAPL"}}}, nil
			},
		}
		h := newTestHandlers(repo)
		app := fiber.New()
		app.Use(authed(userID))
		app.Get("/portfolios/:id", h.GetPortfolio)

		_, status, payload := doJSON(t, app, "GET", "/portfolios/"+portfolioID.String(), "")
		if status != fiber.StatusOK {
			t.Fatalf("status = %d, want 200", status)
		}
		data, _ := payload["data"].(map[string]any)
		if data["riskName"] != "moderate" {
			t.Errorf("riskName = %v", data["riskName"])
		}
		holdings, _ := data["holdings"].([]any)
		if len(holdings) != 1 {
			t.Fatalf("holdings = %v, want 1", data["holdings"])
		}
		if h0, _ := holdings[0].(map[string]any); h0["ticker"] != "AAPL" {
			t.Errorf("ticker = %v, want AAPL", holdings[0])
		}
	})
}

func TestCreatePortfolioHandler(t *testing.T) {
	userID, riskID := uuid.New(), uuid.New()

	newApp := func(repo services.Repository) *fiber.App {
		h := newTestHandlers(repo)
		app := fiber.New()
		app.Use(authed(userID))
		app.Post("/portfolios", h.CreatePortfolio)
		return app
	}

	t.Run("rejects an unsupported portfolio type", func(t *testing.T) {
		app := newApp(&stubRepository{})
		body := `{"name":"X","currency":"USD","type":"lottery","riskId":"` + riskID.String() + `"}`
		_, status, payload := doJSON(t, app, "POST", "/portfolios", body)
		if status != fiber.StatusBadRequest {
			t.Errorf("status = %d, want 400", status)
		}
		if msg, _ := payload["message"].(string); msg != "Invalid portfolio type" {
			t.Errorf("message = %q", msg)
		}
	})

	t.Run("rejects a missing risk", func(t *testing.T) {
		app := newApp(&stubRepository{})
		body := `{"name":"X","currency":"USD","type":"stocks"}`
		_, status, payload := doJSON(t, app, "POST", "/portfolios", body)
		if status != fiber.StatusBadRequest {
			t.Errorf("status = %d, want 400", status)
		}
		if msg, _ := payload["message"].(string); msg != "Invalid risk" {
			t.Errorf("message = %q", msg)
		}
	})

	t.Run("creates the portfolio", func(t *testing.T) {
		var gotType entities.PortfolioType
		repo := &stubRepository{
			createPortfolio: func(_ context.Context, uid uuid.UUID, name, description, baseCurrency string, rid uuid.UUID, typePortfolio entities.PortfolioType, priceValue money.Money, isDefault bool) (entities.Portfolio, error) {
				if uid != userID || rid != riskID {
					t.Error("IDs not forwarded correctly")
				}
				gotType = typePortfolio
				return entities.Portfolio{ID: uuid.New(), Name: name, Type: typePortfolio}, nil
			},
		}
		app := newApp(repo)
		body := `{"name":"Retirement","description":"long term","currency":"USD","type":"diversified","riskId":"` + riskID.String() + `","isDefault":true}`
		_, status, payload := doJSON(t, app, "POST", "/portfolios", body)
		if status != fiber.StatusOK {
			t.Fatalf("status = %d, want 200; payload = %v", status, payload)
		}
		if gotType != entities.PortfolioTypeDiversified {
			t.Errorf("type = %q, want diversified", gotType)
		}
	})
}

func TestUpdatePortfolioHandler(t *testing.T) {
	userID, portfolioID := uuid.New(), uuid.New()

	t.Run("rejects a malformed risk ID", func(t *testing.T) {
		h := newTestHandlers(&stubRepository{})
		app := fiber.New()
		app.Use(authed(userID))
		app.Put("/portfolios/:id", h.UpdatePortfolio)

		body := `{"name":"X","type":"stocks","riskId":"nope"}`
		_, status, _ := doJSON(t, app, "PUT", "/portfolios/"+portfolioID.String(), body)
		if status != fiber.StatusBadRequest {
			t.Errorf("status = %d, want 400", status)
		}
	})

	t.Run("updates the portfolio", func(t *testing.T) {
		riskID := uuid.New()
		repo := &stubRepository{
			updatePortfolio: func(_ context.Context, uid, pid uuid.UUID, name, description string, portfolioType entities.PortfolioType, rid uuid.UUID, isDefault bool) (entities.Portfolio, error) {
				if pid != portfolioID || rid != riskID {
					t.Error("IDs not forwarded correctly")
				}
				return entities.Portfolio{ID: pid, Name: name}, nil
			},
		}
		h := newTestHandlers(repo)
		app := fiber.New()
		app.Use(authed(userID))
		app.Put("/portfolios/:id", h.UpdatePortfolio)

		body := `{"name":"Renamed","type":"stocks","riskId":"` + riskID.String() + `"}`
		_, status, _ := doJSON(t, app, "PUT", "/portfolios/"+portfolioID.String(), body)
		if status != fiber.StatusOK {
			t.Errorf("status = %d, want 200", status)
		}
	})
}

func TestCreatePlatformHandler(t *testing.T) {
	userID := uuid.New()

	newApp := func(repo services.Repository) *fiber.App {
		h := newTestHandlers(repo)
		app := fiber.New()
		app.Use(authed(userID))
		app.Post("/platforms", h.CreatePlatform)
		return app
	}

	t.Run("rejects an unsupported source type", func(t *testing.T) {
		app := newApp(&stubRepository{})
		_, status, _ := doJSON(t, app, "POST", "/platforms", `{"name":"X","type":"casino"}`)
		if status != fiber.StatusBadRequest {
			t.Errorf("status = %d, want 400", status)
		}
	})

	t.Run("creates the platform", func(t *testing.T) {
		repo := &stubRepository{
			createPlatform: func(_ context.Context, uid uuid.UUID, sourceType entities.SourceType, name, description string) (entities.InvestmentSource, error) {
				if sourceType != entities.Broker {
					t.Errorf("sourceType = %q, want broker", sourceType)
				}
				return entities.InvestmentSource{ID: uuid.New(), Name: name, SourceType: sourceType}, nil
			},
		}
		app := newApp(repo)
		_, status, _ := doJSON(t, app, "POST", "/platforms", `{"name":"IBKR","type":"broker"}`)
		if status != fiber.StatusOK {
			t.Errorf("status = %d, want 200", status)
		}
	})
}

func TestCreatePortfolioEntryHandler(t *testing.T) {
	userID := uuid.New()

	newApp := func(repo services.Repository) *fiber.App {
		h := newTestHandlers(repo)
		app := fiber.New()
		app.Use(authed(userID))
		app.Post("/entries", h.CreatePortfolioEntry)
		return app
	}

	baseBody := func(category, txnType string) string {
		return `{"portfolioId":"` + uuid.NewString() + `","assetId":"` + uuid.NewString() + `","sourceId":"` + uuid.NewString() +
			`","costCurrency":"USD","category":"` + category + `","transactionType":"` + txnType + `","entryDate":"2026-06-01T00:00:00Z"}`
	}

	t.Run("rejects an invalid transaction type", func(t *testing.T) {
		app := newApp(&stubRepository{})
		_, status, payload := doJSON(t, app, "POST", "/entries", baseBody("stock", "swap"))
		if status != fiber.StatusBadRequest {
			t.Errorf("status = %d, want 400", status)
		}
		if msg, _ := payload["message"].(string); msg != "Invalid transaction type" {
			t.Errorf("message = %q", msg)
		}
	})

	t.Run("defaults the transaction type to buy and maps the asset category", func(t *testing.T) {
		var gotType entities.TransactionType
		var gotCategory entities.PortfolioEntryCategory
		repo := &stubRepository{
			createPortfolioEntry: func(_ context.Context, uid, pid, aid, sid uuid.UUID, txnType entities.TransactionType, quantity money.Decimal, price money.Money, costCurrency string, category entities.PortfolioEntryCategory, entryDate time.Time, notes string) (entities.PortfolioEntry, error) {
				gotType, gotCategory = txnType, category
				return entities.PortfolioEntry{ID: uuid.New()}, nil
			},
		}
		app := newApp(repo)
		_, status, _ := doJSON(t, app, "POST", "/entries", baseBody("crypto", ""))
		if status != fiber.StatusOK {
			t.Fatalf("status = %d, want 200", status)
		}
		if gotType != entities.Buy {
			t.Errorf("txnType = %q, want the buy default", gotType)
		}
		if gotCategory != entities.Cryptos {
			t.Errorf("category = %q, want cryptos (transformed from the crypto asset type)", gotCategory)
		}
	})

	t.Run("unknown categories fall back to others", func(t *testing.T) {
		var gotCategory entities.PortfolioEntryCategory
		repo := &stubRepository{
			createPortfolioEntry: func(_ context.Context, _, _, _, _ uuid.UUID, _ entities.TransactionType, _ money.Decimal, _ money.Money, _ string, category entities.PortfolioEntryCategory, _ time.Time, _ string) (entities.PortfolioEntry, error) {
				gotCategory = category
				return entities.PortfolioEntry{ID: uuid.New()}, nil
			},
		}
		app := newApp(repo)
		_, status, _ := doJSON(t, app, "POST", "/entries", baseBody("mystery", "buy"))
		if status != fiber.StatusOK {
			t.Fatalf("status = %d, want 200", status)
		}
		if gotCategory != entities.Others {
			t.Errorf("category = %q, want others", gotCategory)
		}
	})
}

func TestCreateTransactionHandler(t *testing.T) {
	userID, entryID := uuid.New(), uuid.New()

	newApp := func(repo services.Repository) *fiber.App {
		h := newTestHandlers(repo)
		app := fiber.New()
		app.Use(authed(userID))
		app.Post("/entries/:entryId/transactions", h.CreateTransaction)
		return app
	}

	t.Run("rejects a malformed entry ID", func(t *testing.T) {
		app := newApp(&stubRepository{})
		_, status, _ := doJSON(t, app, "POST", "/entries/oops/transactions", `{"type":"buy","currency":"USD","transactionDate":"2026-06-20T00:00:00Z"}`)
		if status != fiber.StatusBadRequest {
			t.Errorf("status = %d, want 400", status)
		}
	})

	t.Run("rejects an invalid transaction type", func(t *testing.T) {
		app := newApp(&stubRepository{})
		_, status, _ := doJSON(t, app, "POST", "/entries/"+entryID.String()+"/transactions", `{"type":"swap","currency":"USD","transactionDate":"2026-06-20T00:00:00Z"}`)
		if status != fiber.StatusBadRequest {
			t.Errorf("status = %d, want 400", status)
		}
	})

	t.Run("creates the transaction", func(t *testing.T) {
		prefsChecked := make(chan struct{})
		repo := &stubRepository{
			createTransaction: func(_ context.Context, uid, eid uuid.UUID, txnType entities.TransactionType, quantity money.Decimal, price money.Money, currency string, fees money.Money, transactionDate time.Time, notes string) (entities.Transaction, error) {
				if uid != userID || eid != entryID {
					t.Error("IDs not forwarded correctly")
				}
				return entities.Transaction{ID: uuid.New(), EntryID: eid, Type: txnType, Currency: currency}, nil
			},
			// The alert goroutine checks preferences; opt out so the test
			// never reaches the (nil) mailer.
			getUserPreferences: func(context.Context, uuid.UUID) (entities.UserPreferences, error) {
				defer close(prefsChecked)
				return entities.UserPreferences{EmailAlerts: false}, nil
			},
		}
		app := newApp(repo)
		_, status, payload := doJSON(t, app, "POST", "/entries/"+entryID.String()+"/transactions", `{"type":"sell","currency":"USD","transactionDate":"2026-06-20T00:00:00Z","notes":"partial exit"}`)
		if status != fiber.StatusOK {
			t.Fatalf("status = %d, want 200; payload = %v", status, payload)
		}
		data, _ := payload["data"].(map[string]any)
		if data["type"] != "sell" {
			t.Errorf("type = %v, want sell", data["type"])
		}

		select {
		case <-prefsChecked:
		case <-time.After(2 * time.Second):
			t.Fatal("the alert goroutine never checked preferences")
		}
	})
}

func TestUpdateTransactionHandler(t *testing.T) {
	userID, txnID := uuid.New(), uuid.New()

	t.Run("rejects an invalid transaction type", func(t *testing.T) {
		h := newTestHandlers(&stubRepository{})
		app := fiber.New()
		app.Use(authed(userID))
		app.Put("/transactions/:txnId", h.UpdateTransaction)

		_, status, _ := doJSON(t, app, "PUT", "/transactions/"+txnID.String(), `{"type":"swap"}`)
		if status != fiber.StatusBadRequest {
			t.Errorf("status = %d, want 400", status)
		}
	})

	t.Run("updates the transaction", func(t *testing.T) {
		repo := &stubRepository{
			updateTransaction: func(_ context.Context, uid, tid uuid.UUID, txnType entities.TransactionType, quantity money.Decimal, price money.Money, currency string, fees money.Money, transactionDate time.Time, notes string) (entities.Transaction, error) {
				if tid != txnID {
					t.Errorf("txnID = %s, want %s", tid, txnID)
				}
				return entities.Transaction{ID: tid, Type: txnType}, nil
			},
		}
		h := newTestHandlers(repo)
		app := fiber.New()
		app.Use(authed(userID))
		app.Put("/transactions/:txnId", h.UpdateTransaction)

		_, status, _ := doJSON(t, app, "PUT", "/transactions/"+txnID.String(), `{"type":"dividend","currency":"USD"}`)
		if status != fiber.StatusOK {
			t.Errorf("status = %d, want 200", status)
		}
	})
}

func TestGetPortfolioGrowthHandler(t *testing.T) {
	userID := uuid.New()

	t.Run("forwards the period filter and shapes the response", func(t *testing.T) {
		var gotHasSince bool
		repo := &stubRepository{
			getPortfolioGrowthByUserID: func(_ context.Context, uid uuid.UUID, hasSince bool, since time.Time) ([]entities.PortfolioGrowthPoint, error) {
				gotHasSince = hasSince
				return []entities.PortfolioGrowthPoint{
					{Date: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), TotalValue: "100.00"},
					{Date: time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC), TotalValue: "125.00"},
				}, nil
			},
		}
		h := newTestHandlers(repo)
		app := fiber.New()
		app.Use(authed(userID))
		app.Get("/growth", h.GetPortfolioGrowth)

		_, status, payload := doJSON(t, app, "GET", "/growth?period=1M", "")
		if status != fiber.StatusOK {
			t.Fatalf("status = %d, want 200", status)
		}
		if !gotHasSince {
			t.Error("period=1M should enable the since filter")
		}
		data, _ := payload["data"].(map[string]any)
		summary, _ := data["summary"].(map[string]any)
		if summary["totalGrowthPct"] != "25.00" {
			t.Errorf("totalGrowthPct = %v, want 25.00", summary["totalGrowthPct"])
		}
	})

	t.Run("defaults to the ALL period", func(t *testing.T) {
		var gotHasSince bool
		repo := &stubRepository{
			getPortfolioGrowthByUserID: func(_ context.Context, _ uuid.UUID, hasSince bool, _ time.Time) ([]entities.PortfolioGrowthPoint, error) {
				gotHasSince = hasSince
				return nil, nil
			},
		}
		h := newTestHandlers(repo)
		app := fiber.New()
		app.Use(authed(userID))
		app.Get("/growth", h.GetPortfolioGrowth)

		_, status, _ := doJSON(t, app, "GET", "/growth", "")
		if status != fiber.StatusOK {
			t.Fatalf("status = %d, want 200", status)
		}
		if gotHasSince {
			t.Error("the default ALL period must not filter by date")
		}
	})
}

func TestGetPortfolioGrowthByIDHandler(t *testing.T) {
	userID := uuid.New()
	portfolioID := uuid.New()

	var gotPortfolioID uuid.UUID
	repo := &stubRepository{
		getPortfolioGrowthByPortfolioID: func(_ context.Context, _ uuid.UUID, pid uuid.UUID, hasSince bool, since time.Time) ([]entities.PortfolioGrowthPoint, error) {
			gotPortfolioID = pid
			return []entities.PortfolioGrowthPoint{
				{Date: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), TotalValue: "100.00"},
				{Date: time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC), TotalValue: "150.00"},
			}, nil
		},
	}
	h := newTestHandlers(repo)
	app := fiber.New()
	app.Use(authed(userID))
	app.Get("/:id/growth", h.GetPortfolioGrowthByID)

	_, status, payload := doJSON(t, app, "GET", "/"+portfolioID.String()+"/growth", "")
	if status != fiber.StatusOK {
		t.Fatalf("status = %d, want 200", status)
	}
	if gotPortfolioID != portfolioID {
		t.Errorf("portfolioID = %s, want %s", gotPortfolioID, portfolioID)
	}
	data, _ := payload["data"].(map[string]any)
	summary, _ := data["summary"].(map[string]any)
	if summary["totalGrowthPct"] != "50.00" {
		t.Errorf("totalGrowthPct = %v, want 50.00", summary["totalGrowthPct"])
	}
}

func TestGetUserTransactionsHandler(t *testing.T) {
	userID := uuid.New()
	repo := &stubRepository{
		getRecentTransactionsByUser: func(_ context.Context, uid uuid.UUID, limit int) ([]entities.Transaction, error) {
			if limit != 50 {
				t.Errorf("limit = %d, want 50", limit)
			}
			return []entities.Transaction{{ID: uuid.New(), Type: entities.Buy, Entry: entities.PortfolioEntry{Asset: entities.Asset{Ticker: "AAPL"}}}}, nil
		},
	}
	h := newTestHandlers(repo)
	app := fiber.New()
	app.Use(authed(userID))
	app.Get("/transactions", h.GetUserTransactions)

	_, status, payload := doJSON(t, app, "GET", "/transactions", "")
	if status != fiber.StatusOK {
		t.Fatalf("status = %d, want 200", status)
	}
	data, _ := payload["data"].([]any)
	if len(data) != 1 {
		t.Fatalf("data = %v, want one transaction", payload["data"])
	}
	if txn, _ := data[0].(map[string]any); txn["assetTicker"] != "AAPL" {
		t.Errorf("assetTicker = %v, want AAPL", data[0])
	}
}

func TestGetAssetAllocationHandler(t *testing.T) {
	userID := uuid.New()
	repo := &stubRepository{
		getAssetAllocationByUserID: func(context.Context, uuid.UUID) ([]entities.AllocationItem, error) {
			return []entities.AllocationItem{
				{Category: entities.Stocks, MarketValue: "750.00"},
				{Category: entities.Cryptos, MarketValue: "250.00"},
			}, nil
		},
	}
	h := newTestHandlers(repo)
	app := fiber.New()
	app.Use(authed(userID))
	app.Get("/allocation", h.GetAssetAllocation)

	_, status, payload := doJSON(t, app, "GET", "/allocation", "")
	if status != fiber.StatusOK {
		t.Fatalf("status = %d, want 200", status)
	}
	data, _ := payload["data"].([]any)
	if len(data) != 2 {
		t.Fatalf("data = %v, want two categories", payload["data"])
	}
	first, _ := data[0].(map[string]any)
	if first["percent"] != float64(75) {
		t.Errorf("percent = %v, want 75", first["percent"])
	}
}

func TestGetPortfoliosSummaryHandler(t *testing.T) {
	userID := uuid.New()
	repo := &stubRepository{
		getPortfoliosSummaryByUserID: func(context.Context, uuid.UUID) ([]entities.PortfolioSummaryView, error) {
			return []entities.PortfolioSummaryView{{Name: "Growth", TotalMarketValue: "1000.00"}}, nil
		},
	}
	h := newTestHandlers(repo)
	app := fiber.New()
	app.Use(authed(userID))
	app.Get("/summary", h.GetPortfoliosSummary)

	_, status, payload := doJSON(t, app, "GET", "/summary", "")
	if status != fiber.StatusOK {
		t.Fatalf("status = %d, want 200", status)
	}
	data, _ := payload["data"].([]any)
	if len(data) != 1 {
		t.Errorf("data = %v, want one summary", payload["data"])
	}
}

func TestGetPortfoliosSummaryHandlerCurrencyConversion(t *testing.T) {
	userID := uuid.New()
	repo := &stubRepository{
		getPortfoliosSummaryByUserID: func(context.Context, uuid.UUID) ([]entities.PortfolioSummaryView, error) {
			return []entities.PortfolioSummaryView{{
				Name: "Growth", BaseCurrency: "USD",
				TotalMarketValue: "1000", TotalCostBase: "900", TotalGainLoss: "100", TotalGainLossPct: "11.11",
			}}, nil
		},
		getExchangeRateByPair: func(_ context.Context, from, to string) (entities.ExchangeRate, error) {
			if from == "USD" && to == "COP" {
				return entities.ExchangeRate{FromCurrency: "USD", ToCurrency: "COP", Rate: money.MustFromString("4000")}, nil
			}
			return entities.ExchangeRate{}, errors.New("exchange rate not found")
		},
	}
	h := newTestHandlers(repo)
	app := fiber.New()
	app.Use(authed(userID))
	app.Get("/summary", h.GetPortfoliosSummary)

	_, status, payload := doJSON(t, app, "GET", "/summary?currency=cop", "")
	if status != fiber.StatusOK {
		t.Fatalf("status = %d, want 200", status)
	}
	data, _ := payload["data"].([]any)
	if len(data) != 1 {
		t.Fatalf("data = %v, want one summary", payload["data"])
	}
	item, _ := data[0].(map[string]any)
	if item["totalMarketValue"] != "4000000" || item["displayCurrency"] != "COP" {
		t.Errorf("item = %+v, want totalMarketValue=4000000 displayCurrency=COP", item)
	}
}

func TestGetPortfoliosSummaryHandlerRejectsUnsupportedCurrency(t *testing.T) {
	userID := uuid.New()
	h := newTestHandlers(&stubRepository{})
	app := fiber.New()
	app.Use(authed(userID))
	app.Get("/summary", h.GetPortfoliosSummary)

	_, status, _ := doJSON(t, app, "GET", "/summary?currency=xyz", "")
	if status != fiber.StatusBadRequest {
		t.Fatalf("status = %d, want 400", status)
	}
}
