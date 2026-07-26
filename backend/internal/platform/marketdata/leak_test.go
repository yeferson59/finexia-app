package marketdata

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"testing"
)

// The API key belongs to the user, and the application is the only thing
// standing between it and a log aggregator. These tests pin the guarantee that
// no ordinary formatting path can print it.

const secretKey = "SUPERSECRETAPIKEY1234"

func TestCredentialNeverPrintsTheKey(t *testing.T) {
	cred := Credential{Provider: Finnhub, APIKey: secretKey}

	t.Run("fmt verbs", func(t *testing.T) {
		// %v on a struct would print every field; String() is what stops it.
		for _, format := range []string{"%v", "%s", "%+v", "%#v"} {
			if got := fmt.Sprintf(format, cred); strings.Contains(got, secretKey) {
				t.Errorf("fmt %s printed the key: %s", format, got)
			}
		}
	})

	t.Run("inside a wrapping struct", func(t *testing.T) {
		wrapper := struct {
			Cred Credential
			Note string
		}{cred, "context"}

		if got := fmt.Sprintf("%v", wrapper); strings.Contains(got, secretKey) {
			t.Errorf("printing a struct that holds a Credential leaked the key: %s", got)
		}
	})

	t.Run("json", func(t *testing.T) {
		encoded, err := json.Marshal(cred)
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		if strings.Contains(string(encoded), secretKey) {
			t.Errorf("JSON encoding leaked the key: %s", encoded)
		}
	})

	t.Run("slog", func(t *testing.T) {
		var buf strings.Builder
		slog.New(slog.NewTextHandler(&buf, nil)).Info("using credential", "cred", cred)

		if strings.Contains(buf.String(), secretKey) {
			t.Errorf("structured logging leaked the key: %s", buf.String())
		}
	})

	t.Run("a slice of credentials", func(t *testing.T) {
		if got := fmt.Sprintf("%v", []Credential{cred}); strings.Contains(got, secretKey) {
			t.Errorf("printing a slice leaked the key: %s", got)
		}
	})
}

// Providers take the key as a URL query parameter, and Go's transport errors
// quote the URL they failed on. Errorf is the chokepoint that scrubs it.
func TestErrorfScrubsTheKeyFromTheMessage(t *testing.T) {
	// The shape a real transport failure takes: the whole URL, key included.
	transport := fmt.Errorf(`Get "https://www.alphavantage.co/query?apikey=%s&symbol=AAPL": dial tcp: i/o timeout`, secretKey)

	err := Errorf(AlphaVantage, secretKey, nil, "alphavantage: http get %s: %v", "AAPL", transport)

	if strings.Contains(err.Error(), secretKey) {
		t.Fatalf("the error carries the key: %s", err)
	}
	if !strings.Contains(err.Error(), redacted) {
		t.Errorf("the key was removed but not marked as redacted: %s", err)
	}
	// The rest of the message must survive, or the error stops being useful.
	if !strings.Contains(err.Error(), "i/o timeout") {
		t.Errorf("scrubbing destroyed the diagnostic: %s", err)
	}
}

// Wrapping must not reopen the leak: errors.Is has to keep working while the
// original key-bearing text stays out of the chain.
func TestScrubbedErrorSurvivesWrappingWithoutLeaking(t *testing.T) {
	transport := fmt.Errorf(`Get "https://www.alphavantage.co/query?apikey=%s": refused`, secretKey)
	err := Errorf(AlphaVantage, secretKey, ErrUnauthorized, "alphavantage: %v", transport)

	wrapped := fmt.Errorf("fetch quote %q: %w", "AAPL", err)
	joined := errors.Join(wrapped, errors.New("and another failure"))

	for name, candidate := range map[string]error{"wrapped": wrapped, "joined": joined} {
		if strings.Contains(candidate.Error(), secretKey) {
			t.Errorf("%s error leaked the key: %s", name, candidate)
		}
	}

	if !errors.Is(joined, ErrUnauthorized) {
		t.Error("classification was lost through wrapping")
	}
}

func TestVerdictsAttributeEachFailureToItsProvider(t *testing.T) {
	// What a fallback chain produces when both of a user's keys fail
	// differently. Attribution is what stops a working key being demoted.
	joined := errors.Join(
		Errorf(Finnhub, "key-one", ErrUnauthorized, "finnhub: status 401"),
		Errorf(AlphaVantage, "key-two", ErrRateLimited, "alphavantage: quota exhausted"),
	)

	verdicts := Verdicts(fmt.Errorf("fetch quote %q: %w", "AAPL", joined))

	if len(verdicts) != 2 {
		t.Fatalf("got %d verdicts, want 2: %+v", len(verdicts), verdicts)
	}

	got := map[ProviderName]error{}
	for _, v := range verdicts {
		got[v.Provider] = v.Err
	}

	if !errors.Is(got[Finnhub], ErrUnauthorized) {
		t.Errorf("finnhub verdict = %v, want ErrUnauthorized", got[Finnhub])
	}
	if !errors.Is(got[AlphaVantage], ErrRateLimited) {
		t.Errorf("alphavantage verdict = %v, want ErrRateLimited", got[AlphaVantage])
	}
}

func TestVerdictsIgnoresUnclassifiedErrors(t *testing.T) {
	if v := Verdicts(errors.New("plain failure")); len(v) != 0 {
		t.Fatalf("got %+v, want no verdicts for an unattributed error", v)
	}
	if v := Verdicts(nil); len(v) != 0 {
		t.Fatalf("got %+v, want no verdicts for nil", v)
	}
}
