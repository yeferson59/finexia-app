package alphavantage

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/yeferson59/finexia-app/internal/prices"
)

const baseURL = "https://www.alphavantage.co/query"

var _ prices.Provider = (*Client)(nil)

type Client struct {
	apiKey     string
	httpClient *http.Client
}

func New(apiKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) FetchExchangeRate(ctx context.Context, from, to string) (prices.ExchangeRateResult, error) {
	url := fmt.Sprintf(
		"%s?function=CURRENCY_EXCHANGE_RATE&from_currency=%s&to_currency=%s&apikey=%s",
		baseURL, from, to, c.apiKey,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return prices.ExchangeRateResult{}, fmt.Errorf("alphavantage: build request %s/%s: %w", from, to, err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return prices.ExchangeRateResult{}, fmt.Errorf("alphavantage: http get %s/%s: %w", from, to, err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var envelope struct {
		Data map[string]string `json:"Realtime Currency Exchange Rate"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return prices.ExchangeRateResult{}, fmt.Errorf("alphavantage: decode %s/%s: %w", from, to, err)
	}

	rate, ok := envelope.Data["5. Exchange Rate"]
	if !ok || rate == "" {
		return prices.ExchangeRateResult{}, fmt.Errorf("alphavantage: missing rate for %s/%s (API limit reached or invalid key)", from, to)
	}

	return prices.ExchangeRateResult{Rate: rate, FetchedAt: time.Now().UTC()}, nil
}
