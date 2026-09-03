package portfolio

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"uuid"

	"github.com/gofiber/fiber/v3"
	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/market"
	"github.com/yeferson59/finexia-app/internal/user"
)

// This file covers the write and listing handlers of the module — the ones the
// service-level tests never reach because they exercise binding, path-param
// parsing and DTO validation rather than domain logic.

// fakeAssets stands in for the market module behind portfolio's AssetReader:
// the /portfolios/assets catalog and the admin price update are served through
// it, not through portfolio's own repository.
type fakeAssets struct {
	getAssets        func(ctx context.Context, view market.CatalogView, offset, limit uint) ([]market.Asset, error)
	searchAssets     func(ctx context.Context, view market.CatalogView, search string, offset, limit uint) ([]market.Asset, error)
	updateAssetPrice func(ctx context.Context, assetID uuid.UUID, price money.Money) (market.Asset, error)
}

func (f fakeAssets) GetAssets(ctx context.Context, view market.CatalogView, offset, limit uint) ([]market.Asset, error) {
	return f.getAssets(ctx, view, offset, limit)
}

func (f fakeAssets) SearchAssets(ctx context.Context, view market.CatalogView, search string, offset, limit uint) ([]market.Asset, error) {
	return f.searchAssets(ctx, view, search, offset, limit)
}

func (f fakeAssets) UpdateAssetPrice(ctx context.Context, assetID uuid.UUID, price money.Money) (market.Asset, error) {
	return f.updateAssetPrice(ctx, assetID, price)
}

var _ AssetReader = fakeAssets{}

// newTestModuleWithAssets is newTestModule plus an AssetReader, for the routes
// that read the market catalog.
func newTestModuleWithAssets(t *testing.T, repo *fakeRepository, assets AssetReader, userID uuid.UUID, role string) *fiber.App {
	t.Helper()
	noopLimiter := func(c fiber.Ctx) error { return c.Next() }
	mod := newModule(Deps{
		AuthMiddl: stubAuth{userID: userID, role: role, authenticated: true},
		Limiter:   noopLimiter,
		Assets:    assets,
	}, newTestServices(repo, newMemStorage()))

	app := fiber.New()
	mod.Routes(app)
	return app
}

// doJSON issues a request with a JSON body, for the handlers that bind one.
func doJSON(t *testing.T, app *fiber.App, method, target, body string) *http.Response {
	t.Helper()
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, target, err)
	}
	return resp
}

func TestHandlerCreatePortfolio(t *testing.T) {
	userID := uuid.New()
	riskID := uuid.New()

	newApp := func(t *testing.T, created *Portfolio) *fiber.App {
		repo := new(fakeRepository{
			createPortfolio: func(_ context.Context, uid uuid.UUID, name, description string, baseCurrency money.Currency, rid uuid.UUID, typePortfolio Type, _ money.Money, isDefault bool) (Portfolio, error) {
				*created = Portfolio{ID: uuid.New(), UserID: uid, Name: name, Description: description, BaseCurrency: baseCurrency, RiskID: rid, Type: typePortfolio, IsDefault: isDefault}
				return *created, nil
			},
		})
		return newTestModule(t, repo, userID, "user")
	}

	t.Run("creates and echoes the portfolio", func(t *testing.T) {
		var created Portfolio
		body := `{"name":"Growth","description":"long term","currency":"USD","type":"stocks","riskId":"` + riskID.String() + `","isDefault":true}`

		resp := doJSON(t, newApp(t, &created), http.MethodPost, "/portfolios", body)
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
		if created.UserID != userID {
			t.Errorf("userID = %s, want %s (the handler must take it from the token, not the body)", created.UserID, userID)
		}
		if created.Name != "Growth" || created.BaseCurrency != money.USD || created.RiskID != riskID || !created.IsDefault {
			t.Errorf("created = %+v", created)
		}
	})

	t.Run("rejects an unsupported type before touching the service", func(t *testing.T) {
		// A nil createPortfolio hook panics if the handler calls through, so
		// reaching a 400 proves validation short-circuits.
		app := newTestModule(t, new(fakeRepository{}), userID, "user")

		resp := doJSON(t, app, http.MethodPost, "/portfolios",
			`{"name":"Growth","currency":"USD","type":"nonsense","riskId":"`+riskID.String()+`"}`)
		if resp.StatusCode != fiber.StatusBadRequest {
			t.Fatalf("status = %d, want 400", resp.StatusCode)
		}
	})

	t.Run("requires a risk level", func(t *testing.T) {
		app := newTestModule(t, new(fakeRepository{}), userID, "user")

		resp := doJSON(t, app, http.MethodPost, "/portfolios", `{"name":"Growth","currency":"USD","type":"stocks"}`)
		if resp.StatusCode != fiber.StatusBadRequest {
			t.Fatalf("status = %d, want 400 without riskId", resp.StatusCode)
		}
	})

	t.Run("rejects a malformed body", func(t *testing.T) {
		app := newTestModule(t, new(fakeRepository{}), userID, "user")

		resp := doJSON(t, app, http.MethodPost, "/portfolios", `{`)
		if resp.StatusCode != fiber.StatusBadRequest {
			t.Fatalf("status = %d, want 400", resp.StatusCode)
		}
	})
}

func TestHandlerUpdatePortfolio(t *testing.T) {
	userID := uuid.New()
	portfolioID := uuid.New()
	riskID := uuid.New()

	t.Run("updates the addressed portfolio", func(t *testing.T) {
		var gotPortfolioID, gotRiskID uuid.UUID
		repo := new(fakeRepository{
			updatePortfolio: func(_ context.Context, uid, pid uuid.UUID, name, _ string, _ Type, rid uuid.UUID, _ bool) (Portfolio, error) {
				gotPortfolioID, gotRiskID = pid, rid
				return Portfolio{ID: pid, UserID: uid, Name: name}, nil
			},
		})
		app := newTestModule(t, repo, userID, "user")

		resp := doJSON(t, app, http.MethodPatch, "/portfolios/"+portfolioID.String(),
			`{"name":"Renamed","type":"bonds","riskId":"`+riskID.String()+`"}`)
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
		if gotPortfolioID != portfolioID || gotRiskID != riskID {
			t.Errorf("ids = %s/%s, want %s/%s", gotPortfolioID, gotRiskID, portfolioID, riskID)
		}
	})

	t.Run("rejects a non-uuid portfolio id", func(t *testing.T) {
		app := newTestModule(t, new(fakeRepository{}), userID, "user")

		resp := doJSON(t, app, http.MethodPatch, "/portfolios/not-a-uuid", `{"riskId":"`+riskID.String()+`"}`)
		if resp.StatusCode != fiber.StatusBadRequest {
			t.Fatalf("status = %d, want 400", resp.StatusCode)
		}
	})

	t.Run("rejects a non-uuid risk id", func(t *testing.T) {
		app := newTestModule(t, new(fakeRepository{}), userID, "user")

		resp := doJSON(t, app, http.MethodPatch, "/portfolios/"+portfolioID.String(), `{"riskId":"nope"}`)
		if resp.StatusCode != fiber.StatusBadRequest {
			t.Fatalf("status = %d, want 400", resp.StatusCode)
		}
	})
}

func TestHandlerGetPortfolioTopTransaction(t *testing.T) {
	userID := uuid.New()
	portfolioID := uuid.New()

	repo := new(fakeRepository{
		getTopTransactionByPortfolio: func(_ context.Context, uid, pid uuid.UUID) (TopTransactionDTO, error) {
			if uid != userID || pid != portfolioID {
				t.Errorf("ids = %s/%s, want %s/%s", uid, pid, userID, portfolioID)
			}
			return TopTransactionDTO{AssetTicker: "AAPL"}, nil
		},
	})
	app := newTestModule(t, repo, userID, "user")

	resp := do(t, app, http.MethodGet, "/portfolios/"+portfolioID.String()+"/top-transaction")
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	ok, data := decodeEnvelope(t, resp)
	if !ok {
		t.Error("success = false, want true")
	}
	var dto TopTransactionDTO
	if err := json.Unmarshal(data, &dto); err != nil || dto.AssetTicker != "AAPL" {
		t.Errorf("data = %s (err %v)", data, err)
	}
}

func TestHandlerPlatformCRUD(t *testing.T) {
	userID := uuid.New()
	sourceID := uuid.New()

	t.Run("creates a platform", func(t *testing.T) {
		var gotType SourceType
		var gotName string
		repo := new(fakeRepository{
			createPlatform: func(_ context.Context, _ uuid.UUID, sourceType SourceType, name, description string) (InvestmentSource, error) {
				gotType, gotName = sourceType, name
				return InvestmentSource{ID: sourceID, Name: name, Description: description}, nil
			},
		})
		app := newTestModule(t, repo, userID, "user")

		resp := doJSON(t, app, http.MethodPost, "/portfolios/sources", `{"name":"Interactive Brokers","type":"broker"}`)
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
		if gotType != SourceType("broker") {
			t.Errorf("sourceType = %q, want broker", gotType)
		}
		// The service lowercases the name before persisting.
		if gotName != "interactive brokers" {
			t.Errorf("name = %q, want %q", gotName, "interactive brokers")
		}
	})

	t.Run("rejects an unsupported source type", func(t *testing.T) {
		app := newTestModule(t, new(fakeRepository{}), userID, "user")

		resp := doJSON(t, app, http.MethodPost, "/portfolios/sources", `{"name":"Mystery","type":"nonsense"}`)
		if resp.StatusCode != fiber.StatusBadRequest {
			t.Fatalf("status = %d, want 400", resp.StatusCode)
		}
	})

	t.Run("lists platforms", func(t *testing.T) {
		repo := new(fakeRepository{
			getPlatformsWithStats: func(_ context.Context, uid uuid.UUID, displayCurrency money.Currency) ([]PlatformStats, error) {
				if uid != userID {
					t.Errorf("userID = %s, want %s", uid, userID)
				}
				return []PlatformStats{{
					ID: sourceID, Name: "ibkr", SourceType: SourceType("broker"), Investments: 3,
					TotalValue: "1200.50", DisplayCurrency: money.USD, PositionsUnconverted: 1,
				}}, nil
			},
		})
		app := newTestModule(t, repo, userID, "user")

		resp := do(t, app, http.MethodGet, "/portfolios/sources")
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}

		_, data := decodeEnvelope(t, resp)
		var dtos []PlatformResponseDTO
		if err := json.Unmarshal(data, &dtos); err != nil {
			t.Fatalf("decode data: %v (%s)", err, data)
		}
		if len(dtos) != 1 || dtos[0].Name != "ibkr" || dtos[0].Investments != 3 {
			t.Errorf("dtos = %+v", dtos)
		}
		// The currency of the total and the count of what could not be
		// converted into it have to survive the DTO: without them the amount
		// is a bare number every client reads as dollars.
		if dtos[0].DisplayCurrency != "USD" || dtos[0].PositionsUnconverted != 1 {
			t.Errorf("displayCurrency = %q, positionsUnconverted = %d, want USD / 1", dtos[0].DisplayCurrency, dtos[0].PositionsUnconverted)
		}
	})

	t.Run("lists platforms in a requested currency", func(t *testing.T) {
		var gotCurrency money.Currency
		repo := new(fakeRepository{
			getPlatformsWithStats: func(_ context.Context, _ uuid.UUID, displayCurrency money.Currency) ([]PlatformStats, error) {
				gotCurrency = displayCurrency

				return nil, nil
			},
		})
		app := newTestModule(t, repo, userID, "user")

		resp := do(t, app, http.MethodGet, "/portfolios/sources?currency=cop")
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
		if gotCurrency != money.COP {
			t.Errorf("displayCurrency = %q, want COP", gotCurrency)
		}
	})

	t.Run("rejects an unsupported currency", func(t *testing.T) {
		repo := new(fakeRepository{
			getPlatformsWithStats: func(_ context.Context, _ uuid.UUID, _ money.Currency) ([]PlatformStats, error) {
				t.Error("repository reached with an unsupported currency")
				return nil, nil
			},
		})
		app := newTestModule(t, repo, userID, "user")

		resp := do(t, app, http.MethodGet, "/portfolios/sources?currency=XYZ")
		if resp.StatusCode != fiber.StatusBadRequest {
			t.Fatalf("status = %d, want 400", resp.StatusCode)
		}
	})

	t.Run("updates a platform", func(t *testing.T) {
		var gotSourceID uuid.UUID
		repo := new(fakeRepository{
			updatePlatform: func(_ context.Context, _, sid uuid.UUID, name, _ string, _ SourceType, _ bool) (PlatformStats, error) {
				gotSourceID = sid
				return PlatformStats{ID: sid, Name: name}, nil
			},
		})
		app := newTestModule(t, repo, userID, "user")

		resp := doJSON(t, app, http.MethodPatch, "/portfolios/sources/"+sourceID.String(), `{"name":"ibkr pro","isActive":true}`)
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
		if gotSourceID != sourceID {
			t.Errorf("sourceID = %s, want %s", gotSourceID, sourceID)
		}
	})

	t.Run("deletes a platform", func(t *testing.T) {
		var deleted uuid.UUID
		repo := new(fakeRepository{
			deletePlatform: func(_ context.Context, _, sid uuid.UUID) error {
				deleted = sid
				return nil
			},
		})
		app := newTestModule(t, repo, userID, "user")

		resp := do(t, app, http.MethodDelete, "/portfolios/sources/"+sourceID.String())
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
		if deleted != sourceID {
			t.Errorf("deleted = %s, want %s", deleted, sourceID)
		}
	})

	t.Run("rejects a non-uuid platform id", func(t *testing.T) {
		app := newTestModule(t, new(fakeRepository{}), userID, "user")

		if resp := do(t, app, http.MethodDelete, "/portfolios/sources/not-a-uuid"); resp.StatusCode != fiber.StatusBadRequest {
			t.Errorf("DELETE status = %d, want 400", resp.StatusCode)
		}
		if resp := doJSON(t, app, http.MethodPatch, "/portfolios/sources/not-a-uuid", `{}`); resp.StatusCode != fiber.StatusBadRequest {
			t.Errorf("PATCH status = %d, want 400", resp.StatusCode)
		}
	})
}

func TestHandlerTransactions(t *testing.T) {
	userID := uuid.New()
	entryID := uuid.New()
	txnID := uuid.New()
	txnDate := time.Date(2026, time.March, 2, 0, 0, 0, 0, time.UTC)

	// A transaction alert fires from CreateTransaction; the preference lookup
	// returning alerts-off keeps the goroutine from reaching the mailer.
	alertsOff := func(_ context.Context, _ uuid.UUID) (user.UserPreferences, error) {
		return user.UserPreferences{EmailAlerts: false}, nil
	}

	t.Run("creates a transaction on an entry", func(t *testing.T) {
		var gotEntryID uuid.UUID
		var gotType TransactionType
		repo := new(fakeRepository{
			getUserPreferences: alertsOff,
			createTransaction: func(_ context.Context, _, eid uuid.UUID, txnType TransactionType, _ decimal.Decimal, _ money.Money, currency money.Currency, _ money.Money, _ time.Time, _ string) (Transaction, error) {
				gotEntryID, gotType = eid, txnType
				return Transaction{ID: txnID, EntryID: eid, Type: txnType, Currency: currency}, nil
			},
		})
		app := newTestModule(t, repo, userID, "user")

		body := `{"type":"buy","quantity":"10","price":"150.00","currency":"USD","transactionDate":"` + txnDate.Format(time.RFC3339) + `"}`
		resp := doJSON(t, app, http.MethodPost, "/portfolios/entries/"+entryID.String()+"/transactions", body)
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
		if gotEntryID != entryID || gotType != Buy {
			t.Errorf("entryID/type = %s/%q, want %s/buy", gotEntryID, gotType, entryID)
		}
	})

	t.Run("updates a transaction", func(t *testing.T) {
		var gotTxnID uuid.UUID
		repo := new(fakeRepository{
			updateTransaction: func(_ context.Context, _, tid uuid.UUID, txnType TransactionType, _ decimal.Decimal, _ money.Money, _ string, _ money.Money, _ time.Time, _ string) (Transaction, error) {
				gotTxnID = tid
				return Transaction{ID: tid, Type: txnType}, nil
			},
		})
		app := newTestModule(t, repo, userID, "user")

		body := `{"type":"sell","quantity":"5","price":"160.00","currency":"USD","transactionDate":"` + txnDate.Format(time.RFC3339) + `"}`
		resp := doJSON(t, app, http.MethodPut, "/portfolios/transactions/"+txnID.String(), body)
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
		if gotTxnID != txnID {
			t.Errorf("txnID = %s, want %s", gotTxnID, txnID)
		}
	})

	t.Run("deletes a transaction", func(t *testing.T) {
		var gotUserID, gotTxnID uuid.UUID
		repo := new(fakeRepository{
			deleteTransaction: func(_ context.Context, uid, tid uuid.UUID) error {
				gotUserID, gotTxnID = uid, tid
				return nil
			},
		})
		app := newTestModule(t, repo, userID, "user")

		resp := do(t, app, http.MethodDelete, "/portfolios/transactions/"+txnID.String())
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
		// The caller's id reaches the repository: ownership is enforced inside
		// the DELETE, so dropping it there would delete anybody's row.
		if gotUserID != userID || gotTxnID != txnID {
			t.Errorf("ids = %s/%s, want %s/%s", gotUserID, gotTxnID, userID, txnID)
		}
	})

	// Somebody else's transaction and one that never existed are the same
	// answer: the DELETE matches no row and the repository reports not found.
	t.Run("deleting a transaction the user does not own answers 404", func(t *testing.T) {
		repo := new(fakeRepository{
			deleteTransaction: func(context.Context, uuid.UUID, uuid.UUID) error {
				return ErrTransactionNotFound
			},
		})
		app := newTestModule(t, repo, userID, "user")

		resp := do(t, app, http.MethodDelete, "/portfolios/transactions/"+txnID.String())
		if resp.StatusCode != fiber.StatusNotFound {
			t.Fatalf("status = %d, want 404", resp.StatusCode)
		}
	})

	t.Run("rejects a non-uuid transaction id on delete", func(t *testing.T) {
		app := newTestModule(t, new(fakeRepository{}), userID, "user")

		resp := do(t, app, http.MethodDelete, "/portfolios/transactions/not-a-uuid")
		if resp.StatusCode != fiber.StatusBadRequest {
			t.Fatalf("status = %d, want 400", resp.StatusCode)
		}
	})

	t.Run("rejects an unsupported transaction type", func(t *testing.T) {
		app := newTestModule(t, new(fakeRepository{}), userID, "user")

		body := `{"type":"teleport","quantity":"1","price":"1.00","currency":"USD","transactionDate":"` + txnDate.Format(time.RFC3339) + `"}`
		resp := doJSON(t, app, http.MethodPost, "/portfolios/entries/"+entryID.String()+"/transactions", body)
		if resp.StatusCode != fiber.StatusBadRequest {
			t.Fatalf("status = %d, want 400", resp.StatusCode)
		}
	})

	t.Run("lists the transactions of an entry", func(t *testing.T) {
		repo := new(fakeRepository{
			getTransactionsByEntryID: func(_ context.Context, uid, eid uuid.UUID) ([]Transaction, error) {
				if uid != userID || eid != entryID {
					t.Errorf("ids = %s/%s, want %s/%s", uid, eid, userID, entryID)
				}
				return []Transaction{{ID: txnID, EntryID: eid, Type: Buy}}, nil
			},
		})
		app := newTestModule(t, repo, userID, "user")

		resp := do(t, app, http.MethodGet, "/portfolios/entries/"+entryID.String()+"/transactions")
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
	})

	t.Run("lists the user's recent transactions", func(t *testing.T) {
		var gotLimit int
		repo := new(fakeRepository{
			getRecentTransactionsByUserID: func(_ context.Context, uid uuid.UUID, limit int) ([]Transaction, error) {
				gotLimit = limit
				if uid != userID {
					t.Errorf("userID = %s, want %s", uid, userID)
				}
				return []Transaction{{ID: txnID, Type: Buy}}, nil
			},
		})
		app := newTestModule(t, repo, userID, "user")

		resp := do(t, app, http.MethodGet, "/portfolios/transactions")
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
		if gotLimit != 50 {
			t.Errorf("limit = %d, want the handler's fixed 50", gotLimit)
		}
	})

	t.Run("rejects a non-uuid entry id", func(t *testing.T) {
		app := newTestModule(t, new(fakeRepository{}), userID, "user")

		resp := do(t, app, http.MethodGet, "/portfolios/entries/not-a-uuid/transactions")
		if resp.StatusCode != fiber.StatusBadRequest {
			t.Fatalf("status = %d, want 400", resp.StatusCode)
		}
	})
}

func TestHandlerGetAssetAllocation(t *testing.T) {
	userID := uuid.New()

	newApp := func(t *testing.T, gotCurrency *money.Currency) *fiber.App {
		t.Helper()
		repo := new(fakeRepository{
			getAssetAllocationByUserID: func(_ context.Context, uid uuid.UUID, currency money.Currency) ([]AllocationItem, error) {
				if uid != userID {
					t.Errorf("userID = %s, want %s", uid, userID)
				}
				*gotCurrency = currency

				return []AllocationItem{{Category: market.AssetType("stocks"), MarketValue: "1000", Currency: money.USD}}, nil
			},
		})

		return newTestModule(t, repo, userID, "user")
	}

	// Omitted means "the user's own preference", resolved further down. The
	// handler must pass the absence through rather than pick a currency here.
	t.Run("passes no currency through when none is asked for", func(t *testing.T) {
		var gotCurrency money.Currency

		resp := do(t, newApp(t, &gotCurrency), http.MethodGet, "/portfolios/allocation")
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
	})

	t.Run("normalises the requested currency", func(t *testing.T) {
		var gotCurrency money.Currency

		resp := do(t, newApp(t, &gotCurrency), http.MethodGet, "/portfolios/allocation?currency=cop")
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
		if gotCurrency != money.COP {
			t.Errorf("currency = %q, want COP", gotCurrency)
		}
	})

	// Same rule as the summary's ?currency=: an unsupported code is rejected
	// rather than silently ignored, which would answer in another currency.
	t.Run("rejects an unsupported currency", func(t *testing.T) {
		app := newTestModule(t, new(fakeRepository{}), userID, "user")

		resp := do(t, app, http.MethodGet, "/portfolios/allocation?currency=ARS")
		if resp.StatusCode != fiber.StatusBadRequest {
			t.Fatalf("status = %d, want 400", resp.StatusCode)
		}
	})
}

func TestHandlerGetAssetHoldings(t *testing.T) {
	userID := uuid.New()

	newApp := func(t *testing.T, gotCurrency *money.Currency) *fiber.App {
		t.Helper()
		repo := new(fakeRepository{
			getAssetHoldingsByUserID: func(_ context.Context, uid uuid.UUID, currency money.Currency) ([]AssetHolding, error) {
				if uid != userID {
					t.Errorf("userID = %s, want %s", uid, userID)
				}
				*gotCurrency = currency

				return []AssetHolding{{
					Ticker:          "AAPL",
					Name:            "Apple Inc.",
					AssetType:       market.Stock,
					Quantity:        "12.5",
					MarketValue:     "1000",
					DisplayCurrency: money.USD,
					Portfolios:      2,
				}}, nil
			},
		})

		return newTestModule(t, repo, userID, "user")
	}

	// Same contract as /allocation: omitted means "the account's own currency",
	// resolved in SQL, so the handler passes the absence through.
	t.Run("passes no currency through when none is asked for", func(t *testing.T) {
		var gotCurrency money.Currency

		resp := do(t, newApp(t, &gotCurrency), http.MethodGet, "/portfolios/holdings")
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
		if gotCurrency != money.XXX {
			t.Errorf("currency = %q, want the omitted marker", gotCurrency)
		}
	})

	t.Run("normalises the requested currency", func(t *testing.T) {
		var gotCurrency money.Currency

		resp := do(t, newApp(t, &gotCurrency), http.MethodGet, "/portfolios/holdings?currency=cop")
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
		if gotCurrency != money.COP {
			t.Errorf("currency = %q, want COP", gotCurrency)
		}
	})

	t.Run("rejects an unsupported currency", func(t *testing.T) {
		app := newTestModule(t, new(fakeRepository{}), userID, "user")

		resp := do(t, app, http.MethodGet, "/portfolios/holdings?currency=ARS")
		if resp.StatusCode != fiber.StatusBadRequest {
			t.Fatalf("status = %d, want 400", resp.StatusCode)
		}
	})

	// "/holdings" has to keep matching before the "/:id" family below it, the
	// same trap "/allocation" and "/summary" sit in: registered the other way
	// round it would be read as a portfolio id and answer a 400.
	t.Run("does not fall through to the parametric portfolio route", func(t *testing.T) {
		var gotCurrency money.Currency

		resp := do(t, newApp(t, &gotCurrency), http.MethodGet, "/portfolios/holdings")
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("status = %d, want 200 from the holdings handler", resp.StatusCode)
		}
	})
}

func TestHandlerGetPortfolioGrowth(t *testing.T) {
	userID := uuid.New()
	portfolioID := uuid.New()

	t.Run("aggregated growth defaults to the full history", func(t *testing.T) {
		var gotHasSince bool
		repo := new(fakeRepository{
			getPortfolioGrowthByUserID: func(_ context.Context, _ uuid.UUID, _ money.Currency, hasSince bool, _ time.Time) ([]GrowthPoint, error) {
				gotHasSince = hasSince
				return nil, nil
			},
		})
		app := newTestModule(t, repo, userID, "user")

		resp := do(t, app, http.MethodGet, "/portfolios/growth")
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
		if gotHasSince {
			t.Error("hasSince = true; the default period (ALL) must not bound the range")
		}
	})

	t.Run("a period query bounds the range", func(t *testing.T) {
		var gotHasSince bool
		repo := new(fakeRepository{
			getPortfolioGrowthByUserID: func(_ context.Context, _ uuid.UUID, _ money.Currency, hasSince bool, _ time.Time) ([]GrowthPoint, error) {
				gotHasSince = hasSince
				return nil, nil
			},
		})
		app := newTestModule(t, repo, userID, "user")

		if resp := do(t, app, http.MethodGet, "/portfolios/growth?period=1M"); resp.StatusCode != fiber.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
		if !gotHasSince {
			t.Error("hasSince = false, want true for period=1M")
		}
	})

	t.Run("per-portfolio growth", func(t *testing.T) {
		var gotPortfolioID uuid.UUID
		repo := new(fakeRepository{
			getPortfolioGrowthByPortfolioID: func(_ context.Context, _, pid uuid.UUID, _ bool, _ time.Time) ([]GrowthPoint, error) {
				gotPortfolioID = pid
				return nil, nil
			},
		})
		app := newTestModule(t, repo, userID, "user")

		resp := do(t, app, http.MethodGet, "/portfolios/"+portfolioID.String()+"/growth")
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
		if gotPortfolioID != portfolioID {
			t.Errorf("portfolioID = %s, want %s", gotPortfolioID, portfolioID)
		}
	})
}

func TestHandlerAssets(t *testing.T) {
	userID := uuid.New()
	assetID := uuid.New()

	t.Run("lists the catalog scoped to the caller", func(t *testing.T) {
		var gotView market.CatalogView
		assets := fakeAssets{
			getAssets: func(_ context.Context, view market.CatalogView, _, _ uint) ([]market.Asset, error) {
				gotView = view
				return []market.Asset{{ID: assetID, Ticker: "AAPL"}}, nil
			},
		}
		app := newTestModuleWithAssets(t, new(fakeRepository{}), assets, userID, "user")

		resp := do(t, app, http.MethodGet, "/portfolios/assets?page=1&limit=10")
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
		if gotView.ViewerID != userID {
			t.Errorf("viewerID = %s, want %s", gotView.ViewerID, userID)
		}
		if gotView.All {
			t.Error("a non-admin was served the whole catalog")
		}
	})

	t.Run("an admin sees the whole catalog", func(t *testing.T) {
		var gotView market.CatalogView
		assets := fakeAssets{
			getAssets: func(_ context.Context, view market.CatalogView, _, _ uint) ([]market.Asset, error) {
				gotView = view
				return nil, nil
			},
		}
		app := newTestModuleWithAssets(t, new(fakeRepository{}), assets, userID, "admin")

		if resp := do(t, app, http.MethodGet, "/portfolios/assets?page=1&limit=10"); resp.StatusCode != fiber.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
		if !gotView.All {
			t.Error("an admin was not served the whole catalog, so contributed assets are unmoderatable")
		}
	})

	t.Run("a search query goes to SearchAssets, scoped the same way", func(t *testing.T) {
		var gotSearch string
		var gotView market.CatalogView
		assets := fakeAssets{
			searchAssets: func(_ context.Context, view market.CatalogView, search string, _, _ uint) ([]market.Asset, error) {
				gotSearch, gotView = search, view
				return nil, nil
			},
		}
		app := newTestModuleWithAssets(t, new(fakeRepository{}), assets, userID, "user")

		if resp := do(t, app, http.MethodGet, "/portfolios/assets?page=1&limit=10&search=apple"); resp.StatusCode != fiber.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
		if gotSearch != "apple" {
			t.Errorf("search = %q, want %q", gotSearch, "apple")
		}
		if gotView.ViewerID != userID || gotView.All {
			t.Errorf("view = %+v, want the caller's own scope", gotView)
		}
	})

	t.Run("the manual price update is admin-only", func(t *testing.T) {
		called := false
		assets := fakeAssets{
			updateAssetPrice: func(_ context.Context, id uuid.UUID, _ money.Money) (market.Asset, error) {
				called = true
				return market.Asset{ID: id}, nil
			},
		}
		body := `{"price":"200.00"}`

		app := newTestModuleWithAssets(t, new(fakeRepository{}), assets, userID, "user")
		if resp := doJSON(t, app, http.MethodPatch, "/portfolios/assets/"+assetID.String()+"/price", body); resp.StatusCode != fiber.StatusForbidden {
			t.Errorf("status = %d, want 403 for a non-admin", resp.StatusCode)
		}
		if called {
			t.Error("the service was called despite the 403")
		}

		adminApp := newTestModuleWithAssets(t, new(fakeRepository{}), assets, userID, "admin")
		if resp := doJSON(t, adminApp, http.MethodPatch, "/portfolios/assets/"+assetID.String()+"/price", body); resp.StatusCode != fiber.StatusOK {
			t.Fatalf("status = %d, want 200 for an admin", resp.StatusCode)
		}
		if !called {
			t.Error("the admin request never reached the service")
		}
	})
}

func TestHandlerCreatePortfolioEntry(t *testing.T) {
	userID := uuid.New()
	portfolioID := uuid.New()
	assetID := uuid.New()
	sourceID := uuid.New()
	entryDate := time.Date(2026, time.February, 10, 0, 0, 0, 0, time.UTC)

	body := func(category, txnType string) string {
		return `{"portfolioId":"` + portfolioID.String() + `","assetId":"` + assetID.String() +
			`","sourceId":"` + sourceID.String() + `","category":"` + category +
			`","transactionType":"` + txnType + `","quantity":"10","price":"150.00",` +
			`"costCurrency":"USD","entryDate":"` + entryDate.Format(time.RFC3339) + `"}`
	}

	t.Run("forwards the position without a category of its own", func(t *testing.T) {
		var gotPortfolio, gotAsset, gotSource uuid.UUID
		var gotType TransactionType
		repo := new(fakeRepository{
			createPortfolioEntry: func(_ context.Context, _, pid, aid, sid uuid.UUID, txnType TransactionType, _ decimal.Decimal, _ money.Money, _ money.Currency, _ time.Time, _ string) (Entry, error) {
				gotPortfolio, gotAsset, gotSource, gotType = pid, aid, sid, txnType
				return Entry{ID: uuid.New(), PortfolioID: pid, AssetID: aid, SourceID: sid}, nil
			},
		})
		app := newTestModule(t, repo, userID, "user")

		resp := doJSON(t, app, http.MethodPost, "/portfolios/entries", body(string(market.Stock), "buy"))
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
		if gotPortfolio != portfolioID || gotAsset != assetID || gotSource != sourceID {
			t.Errorf("ids = %s/%s/%s, want %s/%s/%s", gotPortfolio, gotAsset, gotSource, portfolioID, assetID, sourceID)
		}
		if gotType != Buy {
			t.Errorf("transactionType = %q, want buy", gotType)
		}
	})

	// A category in the body is what an older client sends. It carried the
	// asset's type, which the catalogue owns, so it is now ignored rather than
	// validated — including a value that used to earn a 400 (migration 000026).
	t.Run("a category in the body is ignored rather than rejected", func(t *testing.T) {
		called := false
		repo := new(fakeRepository{
			createPortfolioEntry: func(_ context.Context, _, _, _, _ uuid.UUID, _ TransactionType, _ decimal.Decimal, _ money.Money, _ money.Currency, _ time.Time, _ string) (Entry, error) {
				called = true
				return Entry{ID: uuid.New()}, nil
			},
		})
		app := newTestModule(t, repo, userID, "user")

		resp := doJSON(t, app, http.MethodPost, "/portfolios/entries", body("nonsense", "buy"))
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
		if !called {
			t.Error("the entry never reached the repository")
		}
	})

	t.Run("an empty transaction type defaults to buy", func(t *testing.T) {
		var gotType TransactionType
		repo := new(fakeRepository{
			createPortfolioEntry: func(_ context.Context, _, _, _, _ uuid.UUID, txnType TransactionType, _ decimal.Decimal, _ money.Money, _ money.Currency, _ time.Time, _ string) (Entry, error) {
				gotType = txnType
				return Entry{ID: uuid.New()}, nil
			},
		})
		app := newTestModule(t, repo, userID, "user")

		if resp := doJSON(t, app, http.MethodPost, "/portfolios/entries", body(string(market.Stock), "")); resp.StatusCode != fiber.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
		if gotType != Buy {
			t.Errorf("transactionType = %q, want the buy default", gotType)
		}
	})

	t.Run("rejects an unsupported transaction type", func(t *testing.T) {
		app := newTestModule(t, new(fakeRepository{}), userID, "user")

		resp := doJSON(t, app, http.MethodPost, "/portfolios/entries", body(string(market.Stock), "teleport"))
		if resp.StatusCode != fiber.StatusBadRequest {
			t.Fatalf("status = %d, want 400", resp.StatusCode)
		}
	})
}

func TestHandlerGetAssetTransactions(t *testing.T) {
	userID := uuid.New()
	portfolioID := uuid.New()

	newApp := func(t *testing.T, total int) *fiber.App {
		repo := new(fakeRepository{
			countAssetTransactions: func(_ context.Context, _, _ uuid.UUID, _ string) (int, error) {
				return total, nil
			},
			getAssetTransactionsPaginated: func(_ context.Context, uid, pid uuid.UUID, ticker string, limit, offset int) ([]Transaction, error) {
				if uid != userID || pid != portfolioID {
					t.Errorf("ids = %s/%s, want %s/%s", uid, pid, userID, portfolioID)
				}
				if ticker != "AAPL" {
					t.Errorf("ticker = %q, want AAPL", ticker)
				}
				if limit != 10 || offset != 10 {
					t.Errorf("limit/offset = %d/%d, want 10/10 for page 2", limit, offset)
				}
				return []Transaction{{ID: uuid.New(), Type: Buy}}, nil
			},
		})
		return newTestModule(t, repo, userID, "user")
	}

	t.Run("paginates and reports the page count", func(t *testing.T) {
		resp := do(t, newApp(t, 25), http.MethodGet,
			"/portfolios/"+portfolioID.String()+"/assets/AAPL/transactions?page=2&limit=10")
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
		_, data := decodeEnvelope(t, resp)
		var page PaginatedTransactionsDTO
		if err := json.Unmarshal(data, &page); err != nil {
			t.Fatalf("decode data: %v (%s)", err, data)
		}
		// 25 rows at 10 per page rounds up to 3.
		if page.Total != 25 || page.Page != 2 || page.Limit != 10 || page.TotalPages != 3 {
			t.Errorf("page = %+v", page)
		}
	})

	t.Run("an empty result reports zero pages", func(t *testing.T) {
		repo := new(fakeRepository{
			countAssetTransactions: func(_ context.Context, _, _ uuid.UUID, _ string) (int, error) {
				return 0, nil
			},
			getAssetTransactionsPaginated: func(_ context.Context, _, _ uuid.UUID, _ string, _, _ int) ([]Transaction, error) {
				return nil, nil
			},
		})
		app := newTestModule(t, repo, userID, "user")

		resp := do(t, app, http.MethodGet,
			"/portfolios/"+portfolioID.String()+"/assets/AAPL/transactions?page=1&limit=10")
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
		_, data := decodeEnvelope(t, resp)
		var page PaginatedTransactionsDTO
		if err := json.Unmarshal(data, &page); err != nil {
			t.Fatalf("decode data: %v (%s)", err, data)
		}
		if page.TotalPages != 0 {
			t.Errorf("totalPages = %d, want 0 for an empty result", page.TotalPages)
		}
	})

	t.Run("rejects a non-uuid portfolio id", func(t *testing.T) {
		app := newTestModule(t, new(fakeRepository{}), userID, "user")

		resp := do(t, app, http.MethodGet, "/portfolios/not-a-uuid/assets/AAPL/transactions?page=1&limit=10")
		if resp.StatusCode != fiber.StatusBadRequest {
			t.Fatalf("status = %d, want 400", resp.StatusCode)
		}
	})
}
