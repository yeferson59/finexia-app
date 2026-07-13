package yahoo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/yeferson59/finexia-app/internal/platform/marketdata"
)

const baseURL = "https://query1.finance.yahoo.com/v8/finance/chart"

var _ marketdata.Provider = (*Client)(nil)

type Client struct {
	httpClient *http.Client
}

func New() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) fetch(ctx context.Context, symbol string) (float64, error) {
	url := fmt.Sprintf("%s/%s", baseURL, symbol)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, fmt.Errorf("yahoo: build request %s: %w", symbol, err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; finexia/1.0)")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("yahoo: http get %s: %w", symbol, err)
	}
	defer func() { _ = resp.Body.Close() }()

	var envelope struct {
		Chart struct {
			Result []struct {
				Meta struct {
					RegularMarketPrice float64 `json:"regularMarketPrice"`
				} `json:"meta"`
			} `json:"result"`
		} `json:"chart"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return 0, fmt.Errorf("yahoo: decode %s: %w", symbol, err)
	}

	if len(envelope.Chart.Result) == 0 {
		return 0, fmt.Errorf("yahoo: no result for %s (invalid symbol or market unavailable)", symbol)
	}

	price := envelope.Chart.Result[0].Meta.RegularMarketPrice
	if price == 0 {
		return 0, fmt.Errorf("yahoo: zero price for %s (market closed or invalid symbol)", symbol)
	}

	return price, nil
}

// FetchQuote retrieves the current price for a stock, ETF, bond, or crypto
// using the Yahoo Finance chart API.
func (c *Client) FetchQuote(ctx context.Context, symbol string) (marketdata.QuoteResult, error) {
	price, err := c.fetch(ctx, symbol)
	if err != nil {
		return marketdata.QuoteResult{}, err
	}
	return marketdata.QuoteResult{
		Price:     strconv.FormatFloat(price, 'f', -1, 64),
		FetchedAt: time.Now().UTC(),
	}, nil
}

// FetchExchangeRate retrieves the exchange rate between two currencies.
// Yahoo Finance uses the "{FROM}{TO}=X" symbol format which covers both fiat
// pairs (e.g. EURUSD=X) and crypto pairs (e.g. BTCUSD=X).
func (c *Client) FetchExchangeRate(ctx context.Context, from, to string) (marketdata.ExchangeRateResult, error) {
	symbol := from + to + "=X"
	price, err := c.fetch(ctx, symbol)
	if err != nil {
		return marketdata.ExchangeRateResult{}, err
	}
	return marketdata.ExchangeRateResult{
		Rate:      strconv.FormatFloat(price, 'f', -1, 64),
		FetchedAt: time.Now().UTC(),
	}, nil
}
