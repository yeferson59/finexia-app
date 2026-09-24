package sfc

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/yeferson59/finexia-app/internal/platform/marketdata"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func response(status int, body string) *http.Response {
	return new(http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	})
}

func newTestClient(fn roundTripFunc) *Client {
	return New(new(http.Client{Transport: fn}))
}

func TestFundKeyRoundTrip(t *testing.T) {
	key := FundKey{EntityType: 5, Entity: 31, Fund: 3644, Compartment: 1, Participation: 501}

	if got := key.String(); got != "5-31-3644-1-501" {
		t.Fatalf("String = %q", got)
	}

	back, err := ParseFundKey(key.String())
	if err != nil || back != key {
		t.Fatalf("ParseFundKey = %+v, %v", back, err)
	}

	for _, bad := range []string{"", "5-31-3644-1", "5-31-x-1-501", "5-31-3644-1-501-2", "5--31-3644-1"} {
		if _, err := ParseFundKey(bad); err == nil {
			t.Errorf("ParseFundKey(%q) accepted", bad)
		}
	}
}

func TestLatestFunds(t *testing.T) {
	var queries []string

	c := newTestClient(func(r *http.Request) (*http.Response, error) {
		q := r.URL.Query()
		queries = append(queries, q.Get("$where"))

		if strings.HasPrefix(q.Get("$select"), "max(") {
			return response(http.StatusOK, `[{"d":"2026-09-22T00:00:00.000"}]`), nil
		}

		return response(http.StatusOK, `[
			{"fecha_corte":"2026-09-22T00:00:00.000","tipo_entidad":"5","codigo_entidad":"31","nombre_entidad":"Fiduciaria Bancolombia","codigo_negocio":"3644","nombre_patrimonio":"FIC RENTA ACCIONES LATAM","nombre_subtipo_patrimonio":"FIC DE TIPO GENERAL","principal_compartimento":"1","tipo_participacion":"501","valor_unidad_operaciones":"77723.164524","numero_inversionistas":"6383"},
			{"fecha_corte":"2026-09-22T00:00:00.000","tipo_entidad":"5","codigo_entidad":"31","nombre_entidad":"Fiduciaria Bancolombia","codigo_negocio":"3644","nombre_patrimonio":"FIC RENTA ACCIONES LATAM","nombre_subtipo_patrimonio":"FIC DE TIPO GENERAL","principal_compartimento":"1","tipo_participacion":"501","valor_unidad_operaciones":"77723.164524","numero_inversionistas":"6383"},
			{"fecha_corte":"2026-09-22T00:00:00.000","tipo_entidad":"85","codigo_entidad":"27","nombre_entidad":"Bbva Valores","codigo_negocio":"98186","nombre_patrimonio":"FIC BBVA VALORES MONEY MARKET","nombre_subtipo_patrimonio":"FIC DE MERCADO MONETARIO","principal_compartimento":"1","tipo_participacion":"502","valor_unidad_operaciones":"15909.022094","numero_inversionistas":"409"},
			{"fecha_corte":"2026-09-22T00:00:00.000","tipo_entidad":"85","codigo_entidad":"27","nombre_entidad":"Bbva Valores","codigo_negocio":"98186","nombre_patrimonio":"FIC BBVA VALORES MONEY MARKET","nombre_subtipo_patrimonio":"FIC DE MERCADO MONETARIO","principal_compartimento":"1","tipo_participacion":"503","valor_unidad_operaciones":"0","numero_inversionistas":"0"}
		]`), nil
	})

	funds, err := c.LatestFunds(context.Background())
	if err != nil {
		t.Fatalf("LatestFunds: %v", err)
	}

	if len(queries) != 2 || queries[1] != "fecha_corte='2026-09-22T00:00:00.000'" {
		t.Fatalf("queries = %q", queries)
	}

	// The duplicate folds into one and the fund without a unit value is left out.
	if len(funds) != 2 {
		t.Fatalf("got %d funds, want 2: %+v", len(funds), funds)
	}

	got := funds[0]
	if got.Key.String() != "5-31-3644-1-501" || got.UnitValue != "77723.164524" || got.Investors != 6383 ||
		got.Kind != "FIC DE TIPO GENERAL" || !got.Date.Equal(time.Date(2026, 9, 22, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("funds[0] = %+v", got)
	}

	if funds[1].Key.String() != "85-27-98186-1-502" {
		t.Errorf("funds[1] = %+v", funds[1])
	}
}

func TestUnitValues(t *testing.T) {
	var where, order string

	c := newTestClient(func(r *http.Request) (*http.Response, error) {
		where, order = r.URL.Query().Get("$where"), r.URL.Query().Get("$order")

		return response(http.StatusOK, `[
			{"fecha_corte":"2026-09-20T00:00:00.000","valor_unidad_operaciones":"77700.1"},
			{"fecha_corte":"2026-09-21T00:00:00.000","valor_unidad_operaciones":"77710.2"},
			{"fecha_corte":"2026-09-21T00:00:00.000","valor_unidad_operaciones":"77710.2"},
			{"fecha_corte":"2026-09-22T00:00:00.000","valor_unidad_operaciones":"-1"}
		]`), nil
	})

	key := FundKey{EntityType: 5, Entity: 31, Fund: 3644, Compartment: 1, Participation: 501}

	values, err := c.UnitValues(context.Background(), key, time.Date(2026, 9, 20, 15, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("UnitValues: %v", err)
	}

	wantWhere := "tipo_entidad=5 AND codigo_entidad=31 AND codigo_negocio=3644 AND principal_compartimento=1 AND tipo_participacion=501 AND fecha_corte>='2026-09-20T00:00:00'"
	if where != wantWhere || order != "fecha_corte" {
		t.Errorf("where = %q, order = %q", where, order)
	}

	if len(values) != 2 || values[0].Value != "77700.1" || values[1].Value != "77710.2" ||
		!values[1].Date.Equal(time.Date(2026, 9, 21, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("values = %+v", values)
	}
}

func TestErrors(t *testing.T) {
	t.Run("rate limited", func(t *testing.T) {
		c := newTestClient(func(*http.Request) (*http.Response, error) {
			return response(http.StatusTooManyRequests, ``), nil
		})

		_, err := c.UnitValues(context.Background(), FundKey{}, time.Now())
		if !errors.Is(err, marketdata.ErrRateLimited) {
			t.Errorf("err = %v, want ErrRateLimited", err)
		}
	})

	t.Run("empty dataset", func(t *testing.T) {
		c := newTestClient(func(*http.Request) (*http.Response, error) {
			return response(http.StatusOK, `[]`), nil
		})

		_, err := c.LatestFunds(context.Background())
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("err = %v, want ErrNotFound", err)
		}
	})

	t.Run("server error", func(t *testing.T) {
		c := newTestClient(func(*http.Request) (*http.Response, error) {
			return response(http.StatusInternalServerError, `oops`), nil
		})

		if _, err := c.LatestFunds(context.Background()); err == nil {
			t.Error("a 500 was accepted")
		}
	})
}
