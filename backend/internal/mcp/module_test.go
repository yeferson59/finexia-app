package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
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
	"github.com/yeferson59/finexia-app/internal/platform/httpx"
	"github.com/yeferson59/finexia-app/internal/platform/logger"
	"github.com/yeferson59/finexia-app/internal/portfolio"
)

// stubAuth injects the request locals the JWT middleware would normally set.
// With authenticated false the request reaches bindCaller with no identity,
// which is the only way the guard can be bypassed in a test.
type stubAuth struct {
	userID        uuid.UUID
	role          string
	authenticated bool
}

func (a stubAuth) RequireAuth() fiber.Handler {
	return func(c fiber.Ctx) error {
		if a.authenticated {
			c.Locals(httpx.LocalUserID, a.userID.String())
			c.Locals(httpx.LocalToken, "test-token")
			c.Locals(httpx.LocalRole, a.role)
		}

		return c.Next()
	}
}

// fakePortfolios records the user id every read was made for, which is what the
// isolation tests assert on, and answers with one canned row each.
type fakePortfolios struct {
	sawUserID   uuid.UUID
	sawCurrency money.Currency
	sawLimit    int
	sawPeriod   string
	err         error
}

func (f *fakePortfolios) GetPortfoliosSummary(_ context.Context, userID uuid.UUID) ([]portfolio.SummaryView, error) {
	f.sawUserID = userID

	return []portfolio.SummaryView{{
		Name: "Long term", BaseCurrency: money.USD, DisplayCurrency: money.USD,
		TotalPositions: 3, TotalCostBase: "1000.00", TotalMarketValue: "1200.00",
		TotalGainLoss: "200.00", TotalGainLossPct: "20.00", PositionsAtCost: 1,
	}}, f.err
}

func (f *fakePortfolios) GetPortfoliosSummaryInCurrency(_ context.Context, userID uuid.UUID, target money.Currency) ([]portfolio.SummaryView, error) {
	f.sawUserID, f.sawCurrency = userID, target

	return []portfolio.SummaryView{{Name: "Long term", BaseCurrency: money.USD, DisplayCurrency: target}}, f.err
}

func (f *fakePortfolios) GetAssetHoldings(_ context.Context, userID uuid.UUID, target money.Currency) ([]portfolio.AssetHolding, error) {
	f.sawUserID, f.sawCurrency = userID, target

	return []portfolio.AssetHolding{{
		Ticker: "AAPL", Name: "Apple Inc.", AssetType: market.Stock,
		Currency: money.USD, DisplayCurrency: money.USD, Quantity: "10",
		MarketPrice: "150.00", MarketValue: "1500.00", Portfolios: 2,
		PriceSource: portfolio.PriceSourceOwn,
	}}, f.err
}

func (f *fakePortfolios) GetAssetAllocation(_ context.Context, userID uuid.UUID, target money.Currency) ([]portfolio.AllocationItem, error) {
	f.sawUserID, f.sawCurrency = userID, target

	return []portfolio.AllocationItem{{Category: market.Stock, MarketValue: "1500.00", Currency: money.USD}}, f.err
}

func (f *fakePortfolios) GetRecentUserTransactions(_ context.Context, userID uuid.UUID, limit int) ([]portfolio.Transaction, error) {
	f.sawUserID, f.sawLimit = userID, limit

	txn := portfolio.Transaction{
		Type: portfolio.Buy, Quantity: decimal.One, Price: money.MustMoneyFromString("150.00", money.USD),
		Currency: money.USD, FXRate: decimal.One, Fees: money.MustMoneyFromString("1.00", money.USD),
		TransactionDate: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
	}
	txn.Entry.Asset.Ticker = "AAPL"

	return []portfolio.Transaction{txn}, f.err
}

func (f *fakePortfolios) GetPortfolioGrowth(_ context.Context, userID uuid.UUID, cur money.Currency, period string) ([]portfolio.GrowthPoint, portfolio.GrowthSummary, error) {
	f.sawUserID, f.sawCurrency, f.sawPeriod = userID, cur, period

	return []portfolio.GrowthPoint{{Date: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC), TotalValue: "1200.00", Currency: money.USD}},
		portfolio.GrowthSummary{CurrentValue: "1200.00", Currency: money.USD}, f.err
}

func (f *fakePortfolios) GetPlatforms(_ context.Context, userID uuid.UUID, display money.Currency) ([]portfolio.PlatformStats, error) {
	f.sawUserID, f.sawCurrency = userID, display

	return []portfolio.PlatformStats{{Name: "IBKR", SourceType: portfolio.Broker, IsActive: true, Investments: 3}}, f.err
}

// fakeMarket records the CatalogView it was asked for: that view is the whole
// of the catalog's per-user scoping, so it is what the audience test reads.
type fakeMarket struct {
	sawView   market.CatalogView
	sawSearch string
	sawLimit  uint
}

func (f *fakeMarket) GetAssets(_ context.Context, view market.CatalogView, _, limit uint) ([]market.Asset, error) {
	f.sawView, f.sawLimit = view, limit

	return []market.Asset{{Ticker: "AAPL", Name: "Apple Inc.", AssetType: market.Stock, Currency: money.USD, IsCurated: true}}, nil
}

func (f *fakeMarket) SearchAssets(_ context.Context, view market.CatalogView, search string, _, limit uint) ([]market.Asset, error) {
	f.sawView, f.sawSearch, f.sawLimit = view, search, limit

	return []market.Asset{{Ticker: "AAPL", Name: "Apple Inc.", AssetType: market.Stock, Currency: money.USD, IsCurated: true}}, nil
}

func (f *fakeMarket) GetLatestExchangeRates(context.Context) ([]market.ExchangeRate, error) {
	return []market.ExchangeRate{{
		FromCurrency: money.USD, ToCurrency: money.COP, Rate: decimal.MustFromString("4000"),
		RateDate: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC), Source: "dolarapi",
	}}, nil
}

// testApp is one wired module plus the fakes behind it, returned as a struct
// so a test that only needs the router does not have to spell out the rest.
type testApp struct {
	app        *fiber.App
	userID     uuid.UUID
	portfolios *fakePortfolios
	assets     *fakeMarket
}

func newTestApp(t *testing.T, role string, authenticated bool) testApp {
	t.Helper()

	userID := uuid.New()
	portfolios, assets := &fakePortfolios{}, &fakeMarket{}

	mod := New(Deps{
		Portfolios: portfolios,
		Assets:     assets,
		AuthMiddl:  stubAuth{userID: userID, role: role, authenticated: authenticated},
		Limiter:    func(c fiber.Ctx) error { return c.Next() },
		Log:        logger.Noop(),
	})

	app := fiber.New()
	mod.Routes(app)

	return testApp{app: app, userID: userID, portfolios: portfolios, assets: assets}
}

// rpc posts one JSON-RPC request to the MCP endpoint and returns the decoded
// envelope. Both media types are on Accept because the streamable transport
// rejects a POST that does not offer them, even when it answers with JSON.
func rpc(t *testing.T, app *fiber.App, method string, params any) (int, map[string]any) {
	t.Helper()

	body, err := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  method,
		"params":  params,
	})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatalf("POST /mcp %s: %v", method, err)
	}
	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}

	var envelope map[string]any
	if err := json.Unmarshal(raw, &envelope); err != nil {
		t.Fatalf("decode %s response %q: %v", method, raw, err)
	}

	return resp.StatusCode, envelope
}

// callTool runs one tools/call and returns its result object.
func callTool(t *testing.T, app *fiber.App, name string, args map[string]any) map[string]any {
	t.Helper()

	status, envelope := rpc(t, app, "tools/call", map[string]any{"name": name, "arguments": args})
	if status != http.StatusOK {
		t.Fatalf("tools/call %s: status %d", name, status)
	}

	if rpcErr, ok := envelope["error"]; ok {
		t.Fatalf("tools/call %s: protocol error %v", name, rpcErr)
	}

	result, ok := envelope["result"].(map[string]any)
	if !ok {
		t.Fatalf("tools/call %s: no result in %v", name, envelope)
	}

	return result
}

func structured(t *testing.T, result map[string]any) map[string]any {
	t.Helper()

	if isErr, _ := result["isError"].(bool); isErr {
		t.Fatalf("tool reported an error: %v", result["content"])
	}

	out, ok := result["structuredContent"].(map[string]any)
	if !ok {
		t.Fatalf("no structuredContent in %v", result)
	}

	return out
}

// TestToolsListIsReadOnly pins the advertised surface. The names are a contract
// with every client already configured against this server, and the read-only
// annotation is what lets a client call them without asking the user first — so
// a tool that stops being read-only has to fail here first.
func TestToolsListIsReadOnly(t *testing.T) {
	ta := newTestApp(t, "user", true)

	status, envelope := rpc(t, ta.app, "tools/list", map[string]any{})
	if status != http.StatusOK {
		t.Fatalf("tools/list: status %d", status)
	}

	result, ok := envelope["result"].(map[string]any)
	if !ok {
		t.Fatalf("tools/list: no result in %v", envelope)
	}

	tools, ok := result["tools"].([]any)
	if !ok {
		t.Fatalf("tools/list: no tools in %v", result)
	}

	want := map[string]bool{
		"list_portfolios": false, "get_holdings": false, "get_allocation": false,
		"list_recent_transactions": false, "get_portfolio_growth": false,
		"list_platforms": false, "search_assets": false, "list_exchange_rates": false,
	}

	for _, entry := range tools {
		tool, _ := entry.(map[string]any)
		name, _ := tool["name"].(string)

		if _, known := want[name]; !known {
			t.Errorf("tools/list advertises unknown tool %q", name)

			continue
		}

		want[name] = true

		annotations, _ := tool["annotations"].(map[string]any)
		if readOnly, _ := annotations["readOnlyHint"].(bool); !readOnly {
			t.Errorf("tool %q is not annotated read-only", name)
		}

		if _, hasSchema := tool["inputSchema"]; !hasSchema {
			t.Errorf("tool %q has no input schema", name)
		}
	}

	for name, seen := range want {
		if !seen {
			t.Errorf("tools/list is missing %q", name)
		}
	}
}

// TestToolCallReadsTheCallersOwnData is the isolation test: the user id the
// service is queried with must be the one the bearer token authenticated, and
// there must be no argument by which a caller can name another.
func TestToolCallReadsTheCallersOwnData(t *testing.T) {
	ta := newTestApp(t, "user", true)

	out := structured(t, callTool(t, ta.app, "get_holdings", map[string]any{}))

	if ta.portfolios.sawUserID != ta.userID {
		t.Errorf("service queried for %s, want the authenticated %s", ta.portfolios.sawUserID, ta.userID)
	}

	if ta.portfolios.sawCurrency != money.XXX {
		t.Errorf("omitted currency reached the service as %s, want the account's preference (XXX)", ta.portfolios.sawCurrency)
	}

	holdings, ok := out["holdings"].([]any)
	if !ok || len(holdings) != 1 {
		t.Fatalf("holdings: %v", out)
	}

	row, _ := holdings[0].(map[string]any)
	if row["ticker"] != "AAPL" || row["marketValue"] != "1500.00" || row["priceSource"] != "own" {
		t.Errorf("holding row = %v", row)
	}
}

// TestCurrencyArgumentReachesTheService covers the other half of the same
// contract: a currency the caller does name is passed through rather than
// silently ignored.
func TestCurrencyArgumentReachesTheService(t *testing.T) {
	ta := newTestApp(t, "user", true)

	structured(t, callTool(t, ta.app, "list_portfolios", map[string]any{"currency": "cop"}))

	if ta.portfolios.sawCurrency != money.COP {
		t.Errorf("currency reached the service as %s, want COP", ta.portfolios.sawCurrency)
	}
}

// TestUnsupportedCurrencyIsAToolError asserts a bad argument comes back as a
// tool error the model can correct, not as a protocol error that kills the call.
func TestUnsupportedCurrencyIsAToolError(t *testing.T) {
	ta := newTestApp(t, "user", true)

	result := callTool(t, ta.app, "list_portfolios", map[string]any{"currency": "ZZZ"})

	if isErr, _ := result["isError"].(bool); !isErr {
		t.Fatalf("unsupported currency did not produce a tool error: %v", result)
	}

	if ta.portfolios.sawUserID != uuid.Nil() {
		t.Error("the service was queried despite the currency being rejected")
	}

	content, _ := json.Marshal(result["content"])
	if !strings.Contains(string(content), "USD") {
		t.Errorf("the error does not say which currencies are accepted: %s", content)
	}
}

// TestCatalogAudienceFollowsTheRole pins the one place a caller's role changes
// what a tool returns: an admin moderates contributed assets and therefore sees
// the whole table, everyone else sees the curated rows plus their own.
func TestCatalogAudienceFollowsTheRole(t *testing.T) {
	for _, tc := range []struct {
		role    string
		wantAll bool
	}{
		{role: "user", wantAll: false},
		{role: httpx.RoleAdmin, wantAll: true},
	} {
		t.Run(tc.role, func(t *testing.T) {
			ta := newTestApp(t, tc.role, true)

			structured(t, callTool(t, ta.app, "search_assets", map[string]any{"query": "app"}))

			if ta.assets.sawView.ViewerID != ta.userID {
				t.Errorf("catalog read for %s, want %s", ta.assets.sawView.ViewerID, ta.userID)
			}

			if ta.assets.sawView.All != tc.wantAll {
				t.Errorf("CatalogView.All = %v, want %v", ta.assets.sawView.All, tc.wantAll)
			}

			if ta.assets.sawSearch != "app" {
				t.Errorf("search term = %q, want %q", ta.assets.sawSearch, "app")
			}
		})
	}
}

// TestBlankSearchListsTheCatalog covers the branch: no query is a listing, not
// a search for the empty string.
func TestBlankSearchListsTheCatalog(t *testing.T) {
	ta := newTestApp(t, "user", true)

	structured(t, callTool(t, ta.app, "search_assets", map[string]any{"query": "   "}))

	if ta.assets.sawSearch != "" {
		t.Errorf("blank query reached SearchAssets as %q; it should have listed instead", ta.assets.sawSearch)
	}

	if ta.assets.sawLimit != defaultAssetLimit {
		t.Errorf("limit = %d, want the default %d", ta.assets.sawLimit, defaultAssetLimit)
	}
}

// TestLimitIsClampedNotRefused: an over-large page size is a guess, and capping
// it answers the question the caller was asking.
func TestLimitIsClampedNotRefused(t *testing.T) {
	ta := newTestApp(t, "user", true)

	structured(t, callTool(t, ta.app, "list_recent_transactions", map[string]any{"limit": 10_000}))

	if ta.portfolios.sawLimit != maxTransactionLimit {
		t.Errorf("limit = %d, want it capped at %d", ta.portfolios.sawLimit, maxTransactionLimit)
	}
}

// TestUnauthenticatedRequestIsRejected: the endpoint is guarded like every
// other route that answers with one user's private holdings.
func TestUnauthenticatedRequestIsRejected(t *testing.T) {
	ta := newTestApp(t, "user", false)

	status, _ := rpc(t, ta.app, "tools/list", map[string]any{})
	if status != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", status)
	}

	if ta.portfolios.sawUserID != uuid.Nil() {
		t.Error("an unauthenticated request reached the service")
	}
}

// TestGetIsRejectedInStatelessMode documents the transport's shape: a stateless
// server holds no channel to push on, so the GET stream is a 405 by design and
// a client learns that immediately instead of waiting on a stream that will
// never carry anything.
func TestGetIsRejectedInStatelessMode(t *testing.T) {
	ta := newTestApp(t, "user", true)

	req := httptest.NewRequest(http.MethodGet, "/mcp", nil)
	req.Header.Set("Accept", "text/event-stream")

	resp, err := ta.app.Test(req, fiber.TestConfig{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatalf("GET /mcp: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", resp.StatusCode)
	}
}
