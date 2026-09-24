// Package sfc reads the unit values of Colombia's collective investment funds
// (FIC) from the open data the Superintendencia Financiera publishes.
//
// The dataset is "Rentabilidades de los Fondos de Inversión Colectiva (FIC)",
// qhpu-8ixx on datos.gov.co, served by the Socrata API. It has one row per fund,
// type of participation and day since 2016, with the value of a unit that day:
// the figure a fund's statement prints. It takes no key, and like the TRM of
// marketdata/dolarapi it is public official data, so one fetch serves every
// user.
//
// What the dataset looks like, measured when this was written (Sep 2026):
//
//   - About 1,040 rows a day, one per fund and type of participation; each type
//     of participation has its own unit value.
//   - Published with two days of delay: on the 24th the latest day is the 22nd.
//     Weekends and holidays have rows too.
//   - Some funds appear two or three times on the same day, with identical
//     figures. Rows are therefore folded by key and day here, and callers never
//     see a duplicate.
//   - Numbers come as JSON strings, so a unit value reaches the numeric column
//     as it was published, never through a float.
package sfc

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	json "github.com/bytedance/sonic"

	"github.com/yeferson59/finexia-app/internal/platform/marketdata"
)

// Provider names this feed in errors, on the terms of marketdata.DolarAPI: a
// source, never a credential a user can store.
const Provider marketdata.ProviderName = "sfc"

const datasetURL = "https://www.datos.gov.co/resource/qhpu-8ixx.json"

// pageSize is how many rows one request asks for. It is the most Socrata
// serves at once; a fund's whole history since 2016 is about 3,900 days, so a
// history takes one page and the catalog of a day takes one too.
const pageSize = 50000

// maxPages stops a paging loop that the server keeps feeding.
const maxPages = 20

// DefaultHTTPClient is this feed's own client, for the reason dolarapi gives:
// a background job pays a cold connect to a host nothing else talks to.
var DefaultHTTPClient = new(http.Client{Timeout: 45 * time.Second})

// ErrNotFound means the feed has no row for the fund asked about.
var ErrNotFound = errors.New("sfc: fund not published")

// FundKey identifies one type of participation of one fund, which is what has
// a unit value. An entity's code is only unique within its type (a trust
// company and a broker can share one), and a fund's only within its entity, so
// the key takes all five.
type FundKey struct {
	EntityType    int
	Entity        int
	Fund          int
	Compartment   int
	Participation int
}

// String is the key as it is stored and sent: the five codes joined by dashes,
// "5-31-3644-1-501".
func (k FundKey) String() string {
	return fmt.Sprintf("%d-%d-%d-%d-%d", k.EntityType, k.Entity, k.Fund, k.Compartment, k.Participation)
}

// ParseFundKey reads a key written by String.
func ParseFundKey(s string) (FundKey, error) {
	parts := strings.Split(s, "-")
	if len(parts) != 5 {
		return FundKey{}, fmt.Errorf("sfc: fund key %q: want five codes", s)
	}

	var codes [5]int

	for i, part := range parts {
		n, err := strconv.Atoi(part)
		if err != nil || n < 0 {
			return FundKey{}, fmt.Errorf("sfc: fund key %q: code %q is not a number", s, part)
		}

		codes[i] = n
	}

	return FundKey{codes[0], codes[1], codes[2], codes[3], codes[4]}, nil
}

// Fund is one type of participation of one fund, as the latest day published
// it.
type Fund struct {
	Key FundKey
	// EntityName is the manager: "Fiduciaria Bancolombia S.A. Sociedad Fiduciaria".
	EntityName string
	// Name is the fund's: "FONDO DE INVERSIÓN COLECTIVA ABIERTO RENTA ACCIONES LATAM".
	Name string
	// Kind is its class: "FIC DE MERCADO MONETARIO", "FIC DE TIPO GENERAL"…
	Kind string
	// UnitValue is a decimal string, the value of a unit on Date.
	UnitValue string
	Date      time.Time
	// Investors is how many investors that type of participation had.
	Investors int
}

// UnitValue is what a unit of a fund was worth on a day.
type UnitValue struct {
	Date  time.Time
	Value string
}

// Client reads the feed. Like dolarapi's, it holds no credential.
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// New builds the client. A nil httpClient uses DefaultHTTPClient.
func New(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = DefaultHTTPClient
	}

	return new(Client{httpClient: httpClient, baseURL: datasetURL})
}

// row is one record of the dataset, only the columns read here.
type row struct {
	Date          string `json:"fecha_corte"`
	EntityType    string `json:"tipo_entidad"`
	Entity        string `json:"codigo_entidad"`
	EntityName    string `json:"nombre_entidad"`
	Fund          string `json:"codigo_negocio"`
	Name          string `json:"nombre_patrimonio"`
	Kind          string `json:"nombre_subtipo_patrimonio"`
	Compartment   string `json:"principal_compartimento"`
	Participation string `json:"tipo_participacion"`
	UnitValue     string `json:"valor_unidad_operaciones"`
	Investors     string `json:"numero_inversionistas"`
}

const catalogColumns = "fecha_corte,tipo_entidad,codigo_entidad,nombre_entidad,codigo_negocio,nombre_patrimonio," +
	"nombre_subtipo_patrimonio,principal_compartimento,tipo_participacion,valor_unidad_operaciones,numero_inversionistas"

// LatestFunds returns every fund the latest published day has, one per type of
// participation. A row without a positive unit value is left out: it prices
// nothing.
func (c *Client) LatestFunds(ctx context.Context) ([]Fund, error) {
	var latest []struct {
		Date string `json:"d"`
	}

	if err := c.get(ctx, url.Values{"$select": {"max(fecha_corte) AS d"}}, &latest); err != nil {
		return nil, err
	}

	if len(latest) == 0 || latest[0].Date == "" {
		return nil, marketdata.Errorf(Provider, "", ErrNotFound, "sfc: the dataset has no rows")
	}

	rows, err := c.rows(ctx, url.Values{
		"$select": {catalogColumns},
		"$where":  {"fecha_corte='" + latest[0].Date + "'"},
	})
	if err != nil {
		return nil, err
	}

	funds := make([]Fund, 0, len(rows))
	seen := make(map[FundKey]int, len(rows))

	for _, r := range rows {
		f, ok := r.fund()
		if !ok {
			continue
		}

		// The same fund twice on a day carries the same figures; the later row
		// wins so that, if they ever differ, the answer at least does not depend
		// on anything but the feed's order.
		if i, dup := seen[f.Key]; dup {
			funds[i] = f

			continue
		}

		seen[f.Key] = len(funds)
		funds = append(funds, f)
	}

	return funds, nil
}

// UnitValues returns the unit values of one fund from from on, oldest first and
// one per day. A fund the feed has nothing for since then answers an empty
// slice, not an error: the next day may bring its row.
func (c *Client) UnitValues(ctx context.Context, key FundKey, from time.Time) ([]UnitValue, error) {
	where := fmt.Sprintf(
		"tipo_entidad=%d AND codigo_entidad=%d AND codigo_negocio=%d AND principal_compartimento=%d AND tipo_participacion=%d AND fecha_corte>='%s'",
		key.EntityType, key.Entity, key.Fund, key.Compartment, key.Participation, from.UTC().Format("2006-01-02T00:00:00"),
	)

	rows, err := c.rows(ctx, url.Values{
		"$select": {"fecha_corte,valor_unidad_operaciones"},
		"$where":  {where},
		"$order":  {"fecha_corte"},
	})
	if err != nil {
		return nil, err
	}

	values := make([]UnitValue, 0, len(rows))

	for _, r := range rows {
		date, ok := parseDate(r.Date)
		if !ok || !positive(r.UnitValue) {
			continue
		}

		// Ordered by date, so a duplicate is the previous one.
		if n := len(values); n > 0 && values[n-1].Date.Equal(date) {
			values[n-1].Value = r.UnitValue

			continue
		}

		values = append(values, UnitValue{Date: date, Value: r.UnitValue})
	}

	return values, nil
}

// rows reads every page of a query.
func (c *Client) rows(ctx context.Context, q url.Values) ([]row, error) {
	var all []row

	for page := range maxPages {
		q.Set("$limit", strconv.Itoa(pageSize))
		q.Set("$offset", strconv.Itoa(page*pageSize))

		var batch []row
		if err := c.get(ctx, q, &batch); err != nil {
			return nil, err
		}

		all = append(all, batch...)

		if len(batch) < pageSize {
			return all, nil
		}
	}

	return nil, marketdata.Errorf(Provider, "", nil, "sfc: more than %d pages", maxPages)
}

func (c *Client) get(ctx context.Context, q url.Values, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"?"+q.Encode(), nil)
	if err != nil {
		return marketdata.Errorf(Provider, "", nil, "sfc: build request: %v", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return marketdata.Errorf(Provider, "", nil, "sfc: http get: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	switch {
	case resp.StatusCode == http.StatusTooManyRequests:
		return marketdata.Errorf(Provider, "", marketdata.ErrRateLimited, "sfc: status %d", resp.StatusCode)
	case resp.StatusCode != http.StatusOK:
		return marketdata.Errorf(Provider, "", nil, "sfc: status %d", resp.StatusCode)
	}

	if err := json.ConfigFastest.NewDecoder(resp.Body).Decode(out); err != nil {
		return marketdata.Errorf(Provider, "", nil, "sfc: decode: %v", err)
	}

	return nil
}

func (r row) fund() (Fund, bool) {
	date, ok := parseDate(r.Date)
	if !ok || !positive(r.UnitValue) {
		return Fund{}, false
	}

	var codes [5]int

	for i, s := range []string{r.EntityType, r.Entity, r.Fund, r.Compartment, r.Participation} {
		n, err := strconv.Atoi(strings.TrimSpace(s))
		if err != nil || n < 0 {
			return Fund{}, false
		}

		codes[i] = n
	}

	investors, _ := strconv.Atoi(strings.TrimSpace(r.Investors))

	return Fund{
		Key:        FundKey{codes[0], codes[1], codes[2], codes[3], codes[4]},
		EntityName: strings.TrimSpace(r.EntityName),
		Name:       strings.TrimSpace(r.Name),
		Kind:       strings.TrimSpace(r.Kind),
		UnitValue:  r.UnitValue,
		Date:       date,
		Investors:  investors,
	}, true
}

// parseDate reads a floating timestamp, "2026-09-22T00:00:00.000", as its day.
func parseDate(s string) (time.Time, bool) {
	if len(s) < len(time.DateOnly) {
		return time.Time{}, false
	}

	d, err := time.Parse(time.DateOnly, s[:len(time.DateOnly)])

	return d, err == nil
}

// positive reports whether s is a number greater than zero. The value is kept
// as the string the feed sent; this only screens it.
func positive(s string) bool {
	f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)

	return err == nil && f > 0
}
