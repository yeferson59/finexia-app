package alphavantage

import (
	"context"
	"net/url"
	"time"

	"github.com/yeferson59/finexia-app/internal/platform/marketdata"
)

type envelopeQuote struct {
	Data map[string]string `json:"Global Quote"`
}

func (c *Client) FetchQuote(ctx context.Context, symbol string) (marketdata.QuoteResult, error) {
	var envelope envelopeQuote

	params := url.Values{}
	params.Set("function", "GLOBAL_QUOTE")
	params.Set("symbol", symbol)

	if err := c.get(ctx, params, symbol, &envelope); err != nil {
		return marketdata.QuoteResult{}, err
	}

	price, ok := envelope.Data["05. price"]
	if !ok || price == "" {
		return marketdata.QuoteResult{}, marketdata.Errorf(marketdata.AlphaVantage, c.apiKey, marketdata.ErrUnsupported, "alphavantage: no price for %s", symbol)
	}

	return marketdata.QuoteResult{Price: price, Source: marketdata.AlphaVantage, FetchedAt: time.Now().UTC()}, nil
}
