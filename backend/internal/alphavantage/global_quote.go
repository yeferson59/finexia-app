package alphavantage

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/yeferson59/finexia-app/internal/prices"
)

func (c *Client) FetchQuote(ctx context.Context, symbol string) (prices.QuoteResult, error) {
	url := fmt.Sprintf(
		"%s?function=GLOBAL_QUOTE&symbol=%s&apikey=%s",
		baseURL, symbol, c.apiKey,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return prices.QuoteResult{}, fmt.Errorf("alphavantage: build request %s: %w", symbol, err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return prices.QuoteResult{}, fmt.Errorf("alphavantage: http get %s: %w", symbol, err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var envelope struct {
		Data map[string]string `json:"Global Quote"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return prices.QuoteResult{}, fmt.Errorf("alphavantage: decode %s: %w", symbol, err)
	}

	price, ok := envelope.Data["05. price"]
	if !ok || price == "" {
		return prices.QuoteResult{}, fmt.Errorf("alphavantage: missing price for %s (API limit reached or invalid symbol)", symbol)
	}

	return prices.QuoteResult{Price: price, FetchedAt: time.Now().UTC()}, nil
}
