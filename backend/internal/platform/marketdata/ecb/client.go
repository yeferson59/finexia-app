// Package ecb reads the euro foreign exchange reference rates the European
// Central Bank publishes every business day.
//
// Like marketdata/dolarapi, and unlike every other client here, it takes no API
// key — and that is what makes its data shareable. What 000018 forbade is
// serving one user's key-fetched data to another; the ECB reference rates are
// published to anyone who asks, free to reuse, so one fetch serves everybody
// and the result belongs in the shared exchange_rates table.
//
// It exists because until now nothing filled a rate for anything but USD/COP.
// A user holding a euro-quoted stock in a dollar portfolio had no EUR→USD
// anywhere unless they had brought their own provider key or an operator had
// typed the rate in by hand, so their position could not be converted at all.
//
// What is published here is USD↔X for the currencies in majors below, derived
// from the euro-based table the ECB publishes. USD is the hub the conversion
// hops through (portfolio.GetConversionRate tries direct, inverse, then a
// two-leg hop via USD), so USD↔X puts every pair among these currencies at most
// two legs away, while keeping the shared table at a couple of dozen rows a
// person can read. Both directions are stored rather than left to inversion
// because the portfolio_summary view looks up the pair as written and does not
// invert.
//
// COP is deliberately absent: the peso's official rate is the TRM, which
// dolarapi already publishes, and the ECB does not quote it at all.
package ecb

import (
	"context"
	"encoding/xml"
	"net/http"
	"time"

	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/platform/marketdata"
)

const dailyURL = "https://www.ecb.europa.eu/stats/eurofxref/eurofxref-daily.xml"

// base is the currency the ECB quotes everything in: each published rate says
// how many units of a currency one euro buys.
const base = money.EUR

// hub is the currency the stored pairs are anchored to. See the package comment.
const hub = money.USD

// majors are the currencies worth publishing a pair for. The ECB quotes about
// thirty; storing them all would fill the shared table with rows nothing in the
// app converts and push the dashboard's rate list past being readable. These
// are the ones an asset in this catalog is plausibly quoted in.
var majors = []money.Currency{money.EUR, money.GBP, money.CHF, money.JPY, money.CAD, money.AUD, money.CNY, money.MXN, money.BRL}

var _ marketdata.PublicRateSource = (*Client)(nil)

// DefaultHTTPClient mirrors dolarapi's: a background refresh pays a cold
// connect to a host nothing else in the process talks to, which the 10s shared
// timeout is too tight for, and 25s keeps the deadline that fires this one —
// which names the feed — rather than the scheduler's, which names only the job.
var DefaultHTTPClient = new(http.Client{Timeout: 25 * time.Second})

// Client reads the public feed. There is no credential to hold.
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

// envelope is the eurofxref-daily.xml document:
//
//	<gesmes:Envelope>
//	  <Cube><Cube time="2026-08-13">
//	    <Cube currency="USD" rate="1.1642"/>
//	    …
//
// The three nested elements share a name, so they are decoded as three levels
// rather than by matching on attributes.
type envelope struct {
	Days []struct {
		Time  string `xml:"time,attr"`
		Rates []struct {
			Currency string `xml:"currency,attr"`
			Rate     string `xml:"rate,attr"`
		} `xml:"Cube"`
	} `xml:"Cube>Cube"`
}

// FetchRates returns USD↔X for the majors the feed quoted today.
//
// A currency the ECB stopped publishing simply yields no pair: its previous
// rows stay as they were, which is the right failure mode for a rate — stale
// beats absent — and the same one RefreshPublicRates applies per pair.
func (c *Client) FetchRates(ctx context.Context) ([]marketdata.PublicRate, error) {
	var doc envelope

	if err := c.get(ctx, "reference rates", &doc); err != nil {
		return nil, err
	}

	if len(doc.Days) == 0 {
		return nil, marketdata.Errorf(marketdata.ECB, "", marketdata.ErrUnsupported, "ecb: reference rates: no day published")
	}

	day := doc.Days[0]

	// Euro per unit of currency, the euro included so the arithmetic below has
	// no special case for it.
	perEuro := map[money.Currency]decimal.Decimal{base: decimal.One}

	for _, quote := range day.Rates {
		// A code gofinance does not know is skipped rather than defaulted: the
		// majors loop below only reads codes it asked for, so an unknown one has
		// no pair to fill, and guessing a currency here would file a rate under
		// the wrong one.
		code, err := money.GetCurrencyFromISOCode(quote.Currency)
		if err != nil {
			continue
		}

		rate, err := decimal.NewFromString(quote.Rate)
		if err != nil || !rate.IsPos() {
			// One malformed quote is not a failed fetch: skip it and keep the
			// currencies that parsed.
			continue
		}

		perEuro[code] = rate
	}

	usd, ok := perEuro[hub]
	if !ok {
		// Without the dollar leg nothing here can be anchored, and a euro-only
		// table is not what this feed is being read for.
		return nil, marketdata.Errorf(marketdata.ECB, "", marketdata.ErrUnsupported, "ecb: reference rates: no %s quote", hub)
	}

	asOf := parseDay(day.Time)

	rates := make([]marketdata.PublicRate, 0, 2*len(majors))

	for _, code := range majors {
		perEuroCode, ok := perEuro[code]
		if !ok {
			continue
		}

		// One dollar buys perEuroCode/usd of the currency, and one unit of the
		// currency buys usd/perEuroCode dollars. Both are divisions on the
		// decimal engine: at float64 the two directions stopped being each
		// other's inverse in the last digits the numeric column keeps.
		forward, err := perEuroCode.Div(usd)
		if err != nil {
			continue
		}
		backward, err := usd.Div(perEuroCode)
		if err != nil {
			continue
		}

		rates = append(rates,
			marketdata.PublicRate{From: money.USD, To: code, Rate: forward.String(), Source: marketdata.ECB, AsOf: asOf},
			marketdata.PublicRate{From: code, To: money.USD, Rate: backward.String(), Source: marketdata.ECB, AsOf: asOf},
		)
	}

	if len(rates) == 0 {
		return nil, marketdata.Errorf(marketdata.ECB, "", marketdata.ErrUnsupported, "ecb: reference rates: no usable pair among the majors")
	}

	return rates, nil
}

// parseDay reads the day the ECB dated its table. A document without a usable
// date falls back to now, the same rule storePublicRate applies to a rate that
// arrives undated.
func parseDay(value string) time.Time {
	day, err := time.Parse(time.DateOnly, value)
	if err != nil {
		return time.Now().UTC()
	}

	return day.UTC()
}

func (c *Client) get(ctx context.Context, what string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, dailyURL, nil)
	if err != nil {
		return marketdata.Errorf(marketdata.ECB, "", nil, "ecb: build request %s: %v", what, err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return marketdata.Errorf(marketdata.ECB, "", nil, "ecb: http get %s: %v", what, err)
	}
	defer func() { _ = resp.Body.Close() }()

	// A public feed has no key to reject, so the only classification worth
	// making is "slow down": the refresh job backs off on nothing else.
	switch {
	case resp.StatusCode == http.StatusTooManyRequests:
		return marketdata.Errorf(marketdata.ECB, "", marketdata.ErrRateLimited, "ecb: %s: status %d", what, resp.StatusCode)
	case resp.StatusCode != http.StatusOK:
		return marketdata.Errorf(marketdata.ECB, "", nil, "ecb: %s: status %d", what, resp.StatusCode)
	}

	if err := xml.NewDecoder(resp.Body).Decode(out); err != nil {
		return marketdata.Errorf(marketdata.ECB, "", nil, "ecb: decode %s: %v", what, err)
	}

	return nil
}
