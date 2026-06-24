package alphavantage

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type GlobalQuoteResult struct {
	Price     string
	FetchedAt time.Time
}

func (c *Client) FetchGlobalQuote(ctx context.Context, symbol string) (GlobalQuoteResult, error) {
	url := fmt.Sprintf(
		"%s?function=GLOBAL_QUOTE&symbol=%s&apikey=%s",
		baseURL, symbol, c.apiKey,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return GlobalQuoteResult{}, fmt.Errorf("alphavantage: build request %s: %w", symbol, err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return GlobalQuoteResult{}, fmt.Errorf("alphavantage: http get %s: %w", symbol, err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var envelope struct {
		Data map[string]string `json:"Global Quote"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return GlobalQuoteResult{}, fmt.Errorf("alphavantage: decode %s: %w", symbol, err)
	}

	price, ok := envelope.Data["05. price"]
	if !ok || price == "" {
		return GlobalQuoteResult{}, fmt.Errorf("alphavantage: missing price for %s (API limit reached or invalid symbol)", symbol)
	}

	return GlobalQuoteResult{Price: price, FetchedAt: time.Now().UTC()}, nil
}
