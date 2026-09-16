package portfolio

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"uuid"

	"github.com/yeferson59/gofinance/v2/money"
)

// todayJSON is today's UTC day as the rate form sends it. The service refuses a
// start before yesterday, so these bodies cannot carry a fixed date.
func todayJSON() string {
	return time.Now().UTC().Format(time.DateOnly) + "T00:00:00Z"
}

func TestHandlerGetCashRates(t *testing.T) {
	userID := uuid.New()

	repo := new(fakeRepository{
		getCashRatesByUserID: func(_ context.Context, id uuid.UUID) ([]CashRate, error) {
			if id != userID {
				t.Errorf("userID = %v, want %v", id, userID)
			}
			return []CashRate{{
				ID:             uuid.New(),
				Currency:       money.COP,
				AnnualRatePct:  "9.25",
				WithholdingPct: "0",
				Posting:        PostingDaily,
				Latest:         true,
			}}, nil
		},
	})
	app := newTestModule(t, repo, userID, "user")

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/portfolios/cash/rates", nil))
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	_, data := decodeEnvelope(t, resp)
	var rates []CashRate
	if err := json.Unmarshal(data, &rates); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if len(rates) != 1 || rates[0].AnnualRatePct != "9.25" || !rates[0].Latest || rates[0].Currency != money.COP {
		t.Errorf("rates = %+v", rates)
	}
}

// The body the rate form sends: percentages as JSON numbers, no posting, and the
// day as midnight UTC.
func TestHandlerCreateCashRate(t *testing.T) {
	userID, sourceID := uuid.New(), uuid.New()
	var (
		gotUser uuid.UUID
		got     NewCashRateInput
	)

	repo := new(fakeRepository{
		createCashRate: func(_ context.Context, uid uuid.UUID, in NewCashRateInput) (CashRate, error) {
			gotUser, got = uid, in
			return CashRate{ID: uuid.New(), SourceID: in.SourceID, Currency: in.Currency, AnnualRatePct: "9.25", Posting: PostingDaily, Latest: true}, nil
		},
	})
	app := newTestModule(t, repo, userID, "user")

	today := todayJSON()
	body := `{"sourceId":"` + sourceID.String() + `","currency":"COP","annualRatePct":9.25,"withholdingPct":7,"effectiveFrom":"` + today + `"}`
	resp := doJSON(t, app, http.MethodPost, "/portfolios/cash/rates", body)
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d, want 200: %s", resp.StatusCode, raw)
	}

	if gotUser != userID || got.SourceID != sourceID || got.Currency != money.COP {
		t.Errorf("user %v, input %+v", gotUser, got)
	}
	if !got.AnnualRatePct.Equal(mustDecimal(t, "9.25")) || !got.WithholdingPct.Equal(mustDecimal(t, "7")) {
		t.Errorf("rate, withholding = %s, %s; want 9.25, 7", got.AnnualRatePct, got.WithholdingPct)
	}
	if got.Posting != PostingDaily {
		t.Errorf("posting = %q, want daily for an omitted one", got.Posting)
	}
	if got.EffectiveFrom.UTC().Format(time.DateOnly)+"T00:00:00Z" != today {
		t.Errorf("effectiveFrom = %v, want %s", got.EffectiveFrom, today)
	}
}

// Tiers travel as a list of steps, and a cap sent the way it was before tiers
// arrives as the step at zero it now is, after them.
func TestHandlerCreateCashRateReadsTiers(t *testing.T) {
	var got NewCashRateInput

	repo := new(fakeRepository{
		createCashRate: func(_ context.Context, _ uuid.UUID, in NewCashRateInput) (CashRate, error) {
			got = in
			return CashRate{ID: uuid.New(), Currency: in.Currency, AnnualRatePct: "12", Posting: PostingDaily, Latest: true, Tiers: []CashRateTier{}}, nil
		},
	})
	app := newTestModule(t, repo, uuid.New(), "user")

	body := `{"sourceId":"` + uuid.New().String() + `","currency":"COP","annualRatePct":12,` +
		`"tiers":[{"fromBalance":5000000,"annualRatePct":8}],"maxBalance":20000000,"effectiveFrom":"` + todayJSON() + `"}`
	resp := doJSON(t, app, http.MethodPost, "/portfolios/cash/rates", body)
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d, want 200: %s", resp.StatusCode, raw)
	}

	if len(got.Tiers) != 2 {
		t.Fatalf("tiers = %+v, want the step and the cap", got.Tiers)
	}
	if !got.Tiers[0].FromBalance.Equal(mustDecimal(t, "5000000")) || !got.Tiers[0].AnnualRatePct.Equal(mustDecimal(t, "8")) {
		t.Errorf("first step = %+v, want 8 %% from 5 000 000", got.Tiers[0])
	}
	if !got.Tiers[1].FromBalance.Equal(mustDecimal(t, "20000000")) || !got.Tiers[1].AnnualRatePct.IsZero() {
		t.Errorf("second step = %+v, want the cap at 0 %% from 20 000 000", got.Tiers[1])
	}
}

// Every rule the input breaks is answered by the service, so the repository —
// here a hook left nil, which would panic — is never reached.
func TestHandlerCreateCashRateRefusesBeforeTheRepository(t *testing.T) {
	source := `"sourceId":"` + uuid.New().String() + `",`
	today := todayJSON()

	cases := map[string]string{
		"no rate":            `{` + source + `"currency":"COP","annualRatePct":0,"effectiveFrom":"` + today + `"}`,
		"over a hundred":     `{` + source + `"currency":"COP","annualRatePct":101,"effectiveFrom":"` + today + `"}`,
		"five decimals":      `{` + source + `"currency":"COP","annualRatePct":9.12345,"effectiveFrom":"` + today + `"}`,
		"an unknown posting": `{` + source + `"currency":"COP","annualRatePct":9,"posting":"weekly","effectiveFrom":"` + today + `"}`,
		"a cap of nothing":   `{` + source + `"currency":"COP","annualRatePct":9,"maxBalance":0,"effectiveFrom":"` + today + `"}`,
		"steps out of order": `{` + source + `"currency":"COP","annualRatePct":9,"tiers":[{"fromBalance":20000000,"annualRatePct":8},{"fromBalance":5000000,"annualRatePct":0}],"effectiveFrom":"` + today + `"}`,
		"a past start":       `{` + source + `"currency":"COP","annualRatePct":9,"effectiveFrom":"2020-01-01T00:00:00Z"}`,
		"unsupported":        `{` + source + `"currency":"ARS","annualRatePct":9,"effectiveFrom":"` + today + `"}`,
		"no platform":        `{"currency":"COP","annualRatePct":9,"effectiveFrom":"` + today + `"}`,
	}

	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			app := newTestModule(t, new(fakeRepository{}), uuid.New(), "user")

			resp := doJSON(t, app, http.MethodPost, "/portfolios/cash/rates", body)
			if resp.StatusCode != http.StatusBadRequest {
				t.Errorf("status = %d, want 400", resp.StatusCode)
			}
		})
	}
}

// wantDomainDetails checks that a 4xx carries the reason, which is what the
// screen turns into the sentence it shows.
func wantDomainDetails(t *testing.T, resp *http.Response, err error) {
	t.Helper()

	var env struct {
		Details string `json:"details"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&env)
	if !strings.Contains(env.Details, strings.SplitN(err.Error(), ":", 2)[0]) {
		t.Errorf("details = %q, want the domain error", env.Details)
	}
}

func TestHandlerCreateCashRateConflicts(t *testing.T) {
	body := `{"sourceId":"` + uuid.New().String() + `","currency":"COP","annualRatePct":9,"effectiveFrom":"` + todayJSON() + `"}`

	cases := []struct {
		name string
		err  error
		want int
	}{
		{"a version already starts that day or later", ErrCashRateOverlaps, http.StatusConflict},
		{"someone else's platform", ErrPlatformNotFound, http.StatusNotFound},
		{"an inactive platform", invalidCashRate("the platform is inactive; activate it before giving it a rate"), http.StatusBadRequest},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := new(fakeRepository{
				createCashRate: func(context.Context, uuid.UUID, NewCashRateInput) (CashRate, error) {
					return CashRate{}, tc.err
				},
			})
			app := newTestModule(t, repo, uuid.New(), "user")

			resp := doJSON(t, app, http.MethodPost, "/portfolios/cash/rates", body)
			if resp.StatusCode != tc.want {
				t.Errorf("status = %d, want %d", resp.StatusCode, tc.want)
			}
			wantDomainDetails(t, resp, tc.err)
		})
	}
}

func TestHandlerCashRateErrorsMapToStatuses(t *testing.T) {
	rateID := uuid.New()
	values := `{"annualRatePct":8.5,"withholdingPct":0}`
	end := `{"endsOn":"` + todayJSON() + `"}`

	cases := []struct {
		name   string
		err    error
		method string
		path   string
		body   string
		want   int
	}{
		{"correcting a version a later one follows", ErrCashRateNotLatest, http.MethodPut, "", values, http.StatusConflict},
		{"correcting someone else's rate", ErrCashRateNotFound, http.MethodPut, "", values, http.StatusNotFound},
		{"correcting a rate that earned interest", cashRateInUse(time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC), "record a new version instead"), http.MethodPut, "", values, http.StatusConflict},
		{"pausing before a day already computed", cashRateInUse(time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC), "it can stop from 2026-09-14"), http.MethodPost, "/end", end, http.StatusConflict},
		{"deleting a rate that earned interest", cashRateInUse(time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC), "end it instead"), http.MethodDelete, "", "", http.StatusConflict},
		{"pausing before the rate starts", invalidCashRate("the rate does not start until 2026-10-01; delete it instead"), http.MethodPost, "/end", end, http.StatusBadRequest},
		{"pausing a version a later one follows", ErrCashRateNotLatest, http.MethodPost, "/end", end, http.StatusConflict},
		{"deleting a version a later one follows", ErrCashRateNotLatest, http.MethodDelete, "", "", http.StatusConflict},
		{"deleting a missing rate", ErrCashRateNotFound, http.MethodDelete, "", "", http.StatusNotFound},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			checkID := func(id uuid.UUID) {
				if id != rateID {
					t.Errorf("rateID = %v, want %v", id, rateID)
				}
			}

			repo := new(fakeRepository{
				updateCashRate: func(_ context.Context, _, id uuid.UUID, _ CashRateInput) (CashRate, error) {
					checkID(id)
					return CashRate{}, tc.err
				},
				endCashRate: func(_ context.Context, _, id uuid.UUID, _ time.Time) (CashRate, error) {
					checkID(id)
					return CashRate{}, tc.err
				},
				deleteCashRate: func(_ context.Context, _, id uuid.UUID) error {
					checkID(id)
					return tc.err
				},
			})
			app := newTestModule(t, repo, uuid.New(), "user")

			resp := doJSON(t, app, tc.method, "/portfolios/cash/rates/"+rateID.String()+tc.path, tc.body)
			if resp.StatusCode != tc.want {
				t.Errorf("status = %d, want %d", resp.StatusCode, tc.want)
			}
			wantDomainDetails(t, resp, tc.err)
		})
	}
}

func TestHandlerRecalculateCashInterest(t *testing.T) {
	userID, sourceID := uuid.New(), uuid.New()
	from := time.Now().UTC().AddDate(0, 0, -10).Format(time.DateOnly) + "T00:00:00Z"

	var (
		gotFilter CashAccrualFilter
		gotFrom   time.Time
	)

	repo := new(fakeRepository{
		clearCashInterest: func(_ context.Context, filter CashAccrualFilter, day time.Time) (CashInterestCleared, error) {
			gotFilter, gotFrom = filter, day

			return CashInterestCleared{From: day, Balances: 1, Days: 10}, nil
		},
		getCashAccrualTargets: func(context.Context, time.Time, CashAccrualFilter) ([]CashAccrualTarget, error) {
			return nil, nil
		},
	})
	app := newTestModule(t, repo, userID, "user")

	body := `{"sourceId":"` + sourceID.String() + `","currency":"COP","from":"` + from + `"}`
	resp := doJSON(t, app, http.MethodPost, "/portfolios/cash/interest/recalculate", body)
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d, want 200: %s", resp.StatusCode, raw)
	}

	if gotFilter.UserID != userID || gotFilter.SourceID != sourceID || gotFilter.Currency != money.COP {
		t.Errorf("filter = %+v, want the owner's account", gotFilter)
	}
	if gotFrom.UTC().Format(time.DateOnly)+"T00:00:00Z" != from {
		t.Errorf("from = %v, want %s", gotFrom, from)
	}
}

// Every rule is answered by the service, so the repository — here left nil,
// which would panic — is never reached.
func TestHandlerRecalculateCashInterestRefusesBeforeTheRepository(t *testing.T) {
	source := `"sourceId":"` + uuid.New().String() + `",`
	tomorrow := time.Now().UTC().AddDate(0, 0, 1).Format(time.DateOnly) + "T00:00:00Z"

	cases := map[string]string{
		"no platform":   `{"currency":"COP","from":"2026-01-01T00:00:00Z"}`,
		"unsupported":   `{` + source + `"currency":"ARS","from":"2026-01-01T00:00:00Z"}`,
		"no day":        `{` + source + `"currency":"COP"}`,
		"a day to come": `{` + source + `"currency":"COP","from":"` + tomorrow + `"}`,
	}

	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			app := newTestModule(t, new(fakeRepository{}), uuid.New(), "user")

			resp := doJSON(t, app, http.MethodPost, "/portfolios/cash/interest/recalculate", body)
			if resp.StatusCode != http.StatusBadRequest {
				t.Errorf("status = %d, want 400", resp.StatusCode)
			}
		})
	}
}

// The recalculation answers the same way the rate writes do.
func TestHandlerRecalculateCashInterestMapsTheDomainError(t *testing.T) {
	body := `{"sourceId":"` + uuid.New().String() + `","currency":"COP","from":"` +
		time.Now().UTC().AddDate(0, 0, -5).Format(time.DateOnly) + `T00:00:00Z"}`

	repo := new(fakeRepository{
		clearCashInterest: func(context.Context, CashAccrualFilter, time.Time) (CashInterestCleared, error) {
			return CashInterestCleared{}, ErrPlatformNotFound
		},
	})
	app := newTestModule(t, repo, uuid.New(), "user")

	resp := doJSON(t, app, http.MethodPost, "/portfolios/cash/interest/recalculate", body)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
	wantDomainDetails(t, resp, ErrPlatformNotFound)
}
