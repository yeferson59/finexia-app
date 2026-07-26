package providers

import (
	"errors"
	"testing"

	"github.com/yeferson59/finexia-app/internal/platform/marketdata"
	"github.com/yeferson59/finexia-app/internal/platform/marketdata/alphavantage"
	"github.com/yeferson59/finexia-app/internal/platform/marketdata/finnhub"
)

func TestForBuildsTheChainInQuotaOrder(t *testing.T) {
	// Finnhub must lead: 60 calls/minute on its free tier against Alpha
	// Vantage's 5, so it drains a personal quota far more slowly. Declaration
	// order in the argument must not change that.
	chain, err := New(nil).For([]marketdata.Credential{
		{Provider: marketdata.AlphaVantage, APIKey: "av"},
		{Provider: marketdata.Finnhub, APIKey: "fh"},
	})
	if err != nil {
		t.Fatalf("For: %v", err)
	}

	fallback, ok := chain.(*marketdata.FallbackProvider)
	if !ok {
		t.Fatalf("chain is %T, want *marketdata.FallbackProvider", chain)
	}

	providers := fallback.Providers()
	if len(providers) != 2 {
		t.Fatalf("chain has %d providers, want 2", len(providers))
	}
	if _, ok := providers[0].(*finnhub.Client); !ok {
		t.Errorf("first provider is %T, want the finnhub client", providers[0])
	}
	if _, ok := providers[1].(*alphavantage.Client); !ok {
		t.Errorf("second provider is %T, want the alphavantage client", providers[1])
	}
}

func TestForRejectsAnEmptySet(t *testing.T) {
	tests := map[string][]marketdata.Credential{
		"no credentials":   nil,
		"blank key":        {{Provider: marketdata.Finnhub, APIKey: ""}},
		"unknown provider": {{Provider: "bloomberg", APIKey: "k"}},
	}

	for name, creds := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := New(nil).For(creds); !errors.Is(err, marketdata.ErrNoCredentials) {
				t.Fatalf("For = %v, want ErrNoCredentials", err)
			}
		})
	}
}

// A provider that has been retired from the code must not lock a user out of
// the keys that still work.
func TestForSkipsUnknownProvidersButKeepsTheRest(t *testing.T) {
	chain, err := New(nil).For([]marketdata.Credential{
		{Provider: "bloomberg", APIKey: "retired"},
		{Provider: marketdata.Finnhub, APIKey: "fh"},
	})
	if err != nil {
		t.Fatalf("For: %v", err)
	}

	fallback := chain.(*marketdata.FallbackProvider)
	if got := len(fallback.Providers()); got != 1 {
		t.Fatalf("chain has %d providers, want just the usable one", got)
	}
}
