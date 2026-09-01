package market

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"uuid"

	"github.com/gofiber/fiber/v3"

	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
	"github.com/yeferson59/finexia-app/internal/platform/marketdata"
)

// The API key used throughout. Every test in this file eventually checks that
// this string does not appear in a response body: the guarantee the BYO-key
// model rests on is that a stored key never comes back, not even to its owner.
const testAPIKey = "sk-live-supersecret-9f3a"

// stubAuth injects the locals the JWT middleware would set.
type stubAuth struct {
	userID uuid.UUID
	role   string
}

func (a stubAuth) RequireAuth() fiber.Handler {
	return func(c fiber.Ctx) error {
		c.Locals(httpx.LocalUserID, a.userID.String())
		c.Locals(httpx.LocalToken, "test-token")
		c.Locals(httpx.LocalRole, a.role)

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

// stubHoldings stands in for the portfolio module.
type stubHoldings struct {
	assetIDs []uuid.UUID
	pairs    []CurrencyPair
	err      error
}

func (h stubHoldings) HeldAssetIDs(context.Context, uuid.UUID) ([]uuid.UUID, error) {
	return h.assetIDs, h.err
}

func (h stubHoldings) RequiredCurrencyPairs(context.Context, uuid.UUID) ([]CurrencyPair, error) {
	return h.pairs, h.err
}

// newCredentialApp mounts the market routes over the BYO-key fixture, acting as
// userID — the same id a scenario seeds its credential under.
func newCredentialApp(t *testing.T, f *byoFixture, holdings Holdings, userID uuid.UUID) *fiber.App {
	t.Helper()

	noopLimiter := func(c fiber.Ctx) error { return c.Next() }
	mod := New(Deps{
		Service:           f.svc,
		AuthMiddleware:    stubAuth{userID: userID, role: "user"},
		Limiter:           noopLimiter,
		CredentialLimiter: noopLimiter,
		Holdings:          holdings,
	})

	app := fiber.New()
	mod.Routes(app)

	return app
}

func request(t *testing.T, app *fiber.App, method, target, body string) *http.Response {
	t.Helper()

	req := httptest.NewRequest(method, target, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, target, err)
	}

	return resp
}

// bodyOf reads the whole response body as a string, which is what the leak
// assertions search. Reading it raw rather than through a struct is the point:
// a key could leak through a field nobody modelled.
func bodyOf(t *testing.T, resp *http.Response) string {
	t.Helper()

	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	return string(raw)
}

func assertNoKeyLeak(t *testing.T, where, body string) {
	t.Helper()

	if strings.Contains(body, testAPIKey) {
		t.Fatalf("%s leaked the API key in its response body: %s", where, body)
	}
}

func TestCredentialRoutesNeverReturnTheKey(t *testing.T) {
	f := newBYOFixture(t, new(fakeRepository{}), quoteOK())
	app := newCredentialApp(t, f, stubHoldings{}, uuid.New())

	save := request(t, app, http.MethodPut, "/market/credentials/finnhub", `{"apiKey":"`+testAPIKey+`"}`)
	if save.StatusCode != http.StatusOK {
		t.Fatalf("PUT status = %d, want 200", save.StatusCode)
	}

	saveBody := bodyOf(t, save)
	assertNoKeyLeak(t, "PUT /market/credentials/:provider", saveBody)

	// What it does return is the fragment the UI shows, and nothing else usable.
	var saved struct {
		Data Credential `json:"data"`
	}
	if err := json.Unmarshal([]byte(saveBody), &saved); err != nil {
		t.Fatalf("decode save body: %v", err)
	}
	if saved.Data.Last4 != "9f3a" {
		t.Errorf("last4 = %q, want %q", saved.Data.Last4, "9f3a")
	}

	assertNoKeyLeak(t, "GET /market/credentials", bodyOf(t, request(t, app, http.MethodGet, "/market/credentials", "")))
	assertNoKeyLeak(t, "POST /market/credentials/:provider/verify", bodyOf(t, request(t, app, http.MethodPost, "/market/credentials/finnhub/verify", "")))
}

func TestSaveCredentialHandlerErrors(t *testing.T) {
	t.Run("a rejected key is a 400 that says so", func(t *testing.T) {
		f := newBYOFixture(t, new(fakeRepository{}), quoteFailing(providerErr(Finnhub, marketdata.ErrUnauthorized, "finnhub: status 401")))
		app := newCredentialApp(t, f, stubHoldings{}, uuid.New())

		resp := request(t, app, http.MethodPut, "/market/credentials/finnhub", `{"apiKey":"`+testAPIKey+`"}`)
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", resp.StatusCode)
		}

		assertNoKeyLeak(t, "PUT with a rejected key", bodyOf(t, resp))
	})

	// The provider's own text is not echoed back: with Alpha Vantage the key
	// rides in the request URL, so quoting a transport error would quote the key.
	t.Run("provider text does not reach the response body", func(t *testing.T) {
		leaky := "alphavantage: http get https://www.alphavantage.co/query?apikey=" + testAPIKey
		f := newBYOFixture(t, new(fakeRepository{}), new(fakePriceProvider{
			fetchQuote: func(context.Context, string) (marketdata.QuoteResult, error) {
				// Deliberately unscrubbed, standing in for anything that might
				// slip past the client's own scrubbing.
				return marketdata.QuoteResult{}, new(rawError{msg: leaky})
			},
		}))
		app := newCredentialApp(t, f, stubHoldings{}, uuid.New())

		body := bodyOf(t, request(t, app, http.MethodPut, "/market/credentials/alphavantage", `{"apiKey":"`+testAPIKey+`"}`))
		assertNoKeyLeak(t, "PUT with a leaky provider error", body)
	})

	t.Run("an unknown provider is a 400", func(t *testing.T) {
		f := newBYOFixture(t, new(fakeRepository{}), quoteOK())
		app := newCredentialApp(t, f, stubHoldings{}, uuid.New())

		resp := request(t, app, http.MethodPut, "/market/credentials/nasdaq", `{"apiKey":"`+testAPIKey+`"}`)
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", resp.StatusCode)
		}
	})
}

// rawError is an error whose message is not scrubbed and carries no provider
// attribution, i.e. the worst case the handler has to contain.
type rawError struct{ msg string }

func (e *rawError) Error() string { return e.msg }

func TestDeleteCredentialHandler(t *testing.T) {
	f := newBYOFixture(t, new(fakeRepository{}), quoteOK())
	app := newCredentialApp(t, f, stubHoldings{}, uuid.New())

	if resp := request(t, app, http.MethodDelete, "/market/credentials/finnhub", ""); resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404 when nothing is configured", resp.StatusCode)
	}

	if resp := request(t, app, http.MethodPut, "/market/credentials/finnhub", `{"apiKey":"`+testAPIKey+`"}`); resp.StatusCode != http.StatusOK {
		t.Fatalf("PUT status = %d, want 200", resp.StatusCode)
	}

	if resp := request(t, app, http.MethodDelete, "/market/credentials/finnhub", ""); resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200 once the key exists", resp.StatusCode)
	}
}

func TestSyncMarketDataHandler(t *testing.T) {
	assetID := uuid.New()

	repoFor := func() *fakeRepository {
		return new(fakeRepository{
			getAssetByID: func(context.Context, uuid.UUID) (Asset, error) {
				return Asset{ID: assetID, Ticker: "AAPL", AssetType: Stock, Currency: money.USD}, nil
			},
		})
	}

	// The regression: the on-demand sync used to fetch prices only, leaving a
	// multi-currency portfolio with no rate to convert through — and under
	// BYO-key nobody else's rate may be used.
	t.Run("it syncs rates as well as prices", func(t *testing.T) {
		provider := new(fakePriceProvider{
			fetchQuote: func(context.Context, string) (marketdata.QuoteResult, error) {
				return marketdata.QuoteResult{Price: "190.55", Source: Finnhub}, nil
			},
			fetchExchangeRate: func(context.Context, money.Currency, money.Currency) (marketdata.ExchangeRateResult, error) {
				return marketdata.ExchangeRateResult{Rate: "4100.5", Source: Finnhub}, nil
			},
		})

		f := newBYOFixture(t, repoFor(), provider)
		holdings := stubHoldings{
			assetIDs: []uuid.UUID{assetID},
			pairs:    []CurrencyPair{{From: money.USD, To: money.COP}},
		}
		userID := uuid.New()
		app := newCredentialApp(t, f, holdings, userID)
		f.creds.seed(t, f.ring, userID, Finnhub, "user-finnhub-key")

		resp := request(t, app, http.MethodPost, "/market/sync", "")
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}

		var env struct {
			Data SyncResultDTO `json:"data"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
			t.Fatalf("decode: %v", err)
		}

		if len(env.Data.Prices) != 1 {
			t.Errorf("prices = %d, want 1", len(env.Data.Prices))
		}
		if len(env.Data.Rates) != 1 {
			t.Errorf("rates = %d, want 1 — the rate leg of the sync did not run", len(env.Data.Rates))
		}
	})

	t.Run("a user with no key gets a 400, not a 500", func(t *testing.T) {
		f := newBYOFixture(t, repoFor(), quoteOK())
		app := newCredentialApp(t, f, stubHoldings{assetIDs: []uuid.UUID{assetID}}, uuid.New())

		resp := request(t, app, http.MethodPost, "/market/sync", "")
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400 so the UI can point at settings", resp.StatusCode)
		}
	})
}
