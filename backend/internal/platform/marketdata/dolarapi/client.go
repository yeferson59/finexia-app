// Package dolarapi reads the exchange rates dolarapi.com publishes for
// Colombia.
//
// It is the odd one out among the clients here, and deliberately so: it takes
// no API key. That is what makes its data shareable. Every other client in
// marketdata/* is built from a credential the user brought, so what it returns
// belongs to that user alone; this one asks a public endpoint for public
// official data, so one fetch serves everybody and the result goes to the
// shared exchange_rates table.
//
// Only the TRM is read. The TRM (Tasa Representativa del Mercado) is Colombia's
// official USD/COP rate, published by the Superintendencia Financiera and
// republished here; it is the number Colombian statements, banks and tax
// filings use, which makes it the right rate to value a portfolio at. The same
// host also serves an intraday market quote and rates for other currencies
// against the peso (GET /v1/cotizaciones) — adding them is a second request and
// a second decode, not a redesign, but nothing in the app converts those pairs
// today.
package dolarapi

import (
	"context"
	"net/http"
	"strconv"
	"time"

	json "github.com/bytedance/sonic"

	"github.com/yeferson59/finexia-app/internal/platform/marketdata"
)

const baseURL = "https://co.dolarapi.com/v1"

// trmBase is the currency the TRM prices. The endpoint states the unit it is
// quoted in (COP) but not what it quotes, because for the TRM there is only one
// answer: it is COP per US dollar, by definition.
const trmBase = "USD"

var _ marketdata.PublicRateSource = (*Client)(nil)

// DefaultHTTPClient is this feed's own client, not marketdata's.
//
// The 10s shared timeout is sized for a request a user is waiting on. This one
// is a background refresh every hour, which means every run pays a cold
// connect — TLS and DNS to a host nothing else in the process talks to — and
// that has been measured here at over 15s. Under the shared client the job
// failed on its first attempt and only succeeded on the retry, once the
// resolver was warm.
//
// 25s rather than more because the scheduler gives each attempt 30s: the
// deadline that fires should be this one, which reports which feed hung, not
// the runner's, which reports only that the job did.
var DefaultHTTPClient = new(http.Client{Timeout: 25 * time.Second})

// Client reads the public feed. It holds no credential — there is none to
// hold — only the HTTP client.
type Client struct {
	httpClient *http.Client
}

// New builds the client. A nil httpClient uses DefaultHTTPClient above.
func New(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = DefaultHTTPClient
	}

	return new(Client{httpClient: httpClient})
}

// trmResponse is GET /v1/trm:
//
//	{"unidad":"COP","nombre":"TRM","valor":3157.43,"fechaActualizacion":"..."}
type trmResponse struct {
	// Unidad is the currency valor is expressed in, read rather than assumed so
	// a feed that ever changed it would fail validation instead of silently
	// storing pesos under another code.
	Unidad             string    `json:"unidad"`
	Valor              float64   `json:"valor"`
	FechaActualizacion time.Time `json:"fechaActualizacion"`
}

// FetchRates returns the pairs this feed publishes: today, USD→COP at the TRM.
//
// Errors go through marketdata.Errorf like every other client's. There is no
// key to scrub here, but the constructor is also what attributes a failure to
// its provider, and going around it would leave this feed the one source whose
// failures arrive unlabelled.
func (c *Client) FetchRates(ctx context.Context) ([]marketdata.PublicRate, error) {
	var body trmResponse

	if err := c.get(ctx, "/trm", "TRM", &body); err != nil {
		return nil, err
	}

	if body.Valor <= 0 {
		return nil, marketdata.Errorf(marketdata.DolarAPI, "", marketdata.ErrUnsupported, "dolarapi: TRM: non-positive value %v", body.Valor)
	}
	if body.Unidad == "" {
		return nil, marketdata.Errorf(marketdata.DolarAPI, "", marketdata.ErrUnsupported, "dolarapi: TRM: response carries no currency unit")
	}

	asOf := body.FechaActualizacion
	if asOf.IsZero() {
		asOf = time.Now()
	}

	return []marketdata.PublicRate{{
		From: trmBase,
		To:   body.Unidad,
		// -1 precision prints the shortest decimal that reads back as the same
		// float64, so the two decimals the feed published arrive at the numeric
		// column as they were written.
		Rate:   strconv.FormatFloat(body.Valor, 'f', -1, 64),
		Source: marketdata.DolarAPI,
		AsOf:   asOf.UTC(),
	}}, nil
}

func (c *Client) get(ctx context.Context, path, what string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+path, nil)
	if err != nil {
		return marketdata.Errorf(marketdata.DolarAPI, "", nil, "dolarapi: build request %s: %v", what, err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return marketdata.Errorf(marketdata.DolarAPI, "", nil, "dolarapi: http get %s: %v", what, err)
	}
	defer func() { _ = resp.Body.Close() }()

	// A public feed has no key to reject, so the only classification worth
	// making is "slow down": the refresh job backs off on nothing else.
	switch {
	case resp.StatusCode == http.StatusTooManyRequests:
		return marketdata.Errorf(marketdata.DolarAPI, "", marketdata.ErrRateLimited, "dolarapi: %s: status %d", what, resp.StatusCode)
	case resp.StatusCode != http.StatusOK:
		return marketdata.Errorf(marketdata.DolarAPI, "", nil, "dolarapi: %s: status %d", what, resp.StatusCode)
	}

	if err := json.ConfigFastest.NewDecoder(resp.Body).Decode(out); err != nil {
		return marketdata.Errorf(marketdata.DolarAPI, "", nil, "dolarapi: decode %s: %v", what, err)
	}

	return nil
}
