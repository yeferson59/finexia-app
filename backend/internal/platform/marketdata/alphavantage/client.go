// Package alphavantage talks to the Alpha Vantage API with a key the user
// brought themselves.
//
// Alpha Vantage only accepts the key as a URL query parameter, so the key is
// present in every request URL. That makes error handling security-relevant:
// Go's transport errors quote the URL they failed on, so no error from this
// package may be returned verbatim. Everything goes through
// marketdata.Errorf, which scrubs the key out of the message.
package alphavantage

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	json "github.com/bytedance/sonic"

	"github.com/yeferson59/finexia-app/internal/platform/marketdata"
)

const baseURL = "https://www.alphavantage.co/query"

var _ marketdata.Provider = (*Client)(nil)

type Client struct {
	apiKey     string
	httpClient *http.Client
}

// New builds a client for one user's key. Callers should pass the shared
// marketdata.DefaultHTTPClient; a nil client falls back to it rather than
// minting a new one per call.
func New(apiKey string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = marketdata.DefaultHTTPClient
	}

	return new(Client{apiKey: apiKey, httpClient: httpClient})
}

// get issues a query and decodes it into out. The response envelope is
// inspected first: Alpha Vantage reports a bad key and an exhausted quota with
// HTTP 200 and a JSON field, not a status code.
func (c *Client) get(ctx context.Context, params url.Values, what string, out any) error {
	params.Set("apikey", c.apiKey)

	endpoint := baseURL + "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return marketdata.Errorf(marketdata.AlphaVantage, c.apiKey, nil, "alphavantage: build request %s: %v", what, err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return marketdata.Errorf(marketdata.AlphaVantage, c.apiKey, nil, "alphavantage: http get %s: %v", what, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return marketdata.Errorf(marketdata.AlphaVantage, c.apiKey, marketdata.ErrUnauthorized, "alphavantage: %s: status %d", what, resp.StatusCode)
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return marketdata.Errorf(marketdata.AlphaVantage, c.apiKey, marketdata.ErrRateLimited, "alphavantage: %s: status %d", what, resp.StatusCode)
	}

	if resp.StatusCode != http.StatusOK {
		return marketdata.Errorf(marketdata.AlphaVantage, c.apiKey, nil, "alphavantage: %s: status %d", what, resp.StatusCode)
	}

	// Decode into a buffer first so the envelope can be classified before the
	// caller's own decoding.
	var raw json.NoCopyRawMessage
	if err := json.ConfigFastest.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return marketdata.Errorf(marketdata.AlphaVantage, c.apiKey, nil, "alphavantage: decode %s: %v", what, err)
	}

	if err := c.classify(raw, what); err != nil {
		return err
	}

	if err := json.Unmarshal(raw, out); err != nil {
		return marketdata.Errorf(marketdata.AlphaVantage, c.apiKey, nil, "alphavantage: decode %s: %v", what, err)
	}

	return nil
}

// classify turns Alpha Vantage's 200-with-a-message replies into the sentinels
// the credential store uses to decide whether a key is dead or just throttled.
func (c *Client) classify(raw json.NoCopyRawMessage, what string) error {
	var envelope struct {
		ErrorMessage string `json:"Error Message"`
		Note         string `json:"Note"`
		Information  string `json:"Information"`
	}

	if err := json.Unmarshal(raw, &envelope); err != nil {
		return nil
	}

	switch {
	case envelope.ErrorMessage != "":
		return marketdata.Errorf(marketdata.AlphaVantage, c.apiKey, marketdata.ErrUnauthorized, "alphavantage: %s: %s", what, envelope.ErrorMessage)
	case envelope.Note != "":
		return marketdata.Errorf(marketdata.AlphaVantage, c.apiKey, marketdata.ErrRateLimited, "alphavantage: %s: %s", what, envelope.Note)
	case envelope.Information != "":
		// Used both for quota exhaustion on the free tier and for premium-only
		// endpoints; treating it as throttling keeps the key marked usable.
		return marketdata.Errorf(marketdata.AlphaVantage, c.apiKey, marketdata.ErrRateLimited, "alphavantage: %s: %s", what, envelope.Information)
	default:
		return nil
	}
}

func (c *Client) FetchExchangeRate(ctx context.Context, from, to string) (marketdata.ExchangeRateResult, error) {
	what := fmt.Sprintf("%s/%s", from, to)

	var envelope struct {
		Data map[string]string `json:"Realtime Currency Exchange Rate"`
	}

	params := url.Values{}

	params.Set("function", "CURRENCY_EXCHANGE_RATE")
	params.Set("from_currency", from)
	params.Set("to_currency", to)

	if err := c.get(ctx, params, what, &envelope); err != nil {
		return marketdata.ExchangeRateResult{}, err
	}

	rate, ok := envelope.Data["5. Exchange Rate"]
	if !ok || rate == "" {
		return marketdata.ExchangeRateResult{}, marketdata.Errorf(marketdata.AlphaVantage, c.apiKey, marketdata.ErrUnsupported, "alphavantage: no rate for %s", what)
	}

	return marketdata.ExchangeRateResult{Rate: rate, Source: marketdata.AlphaVantage, FetchedAt: time.Now().UTC()}, nil
}
