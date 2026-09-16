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

func TestHandlerGetCashBalances(t *testing.T) {
	userID := uuid.New()
	var gotCurrency money.Currency

	repo := new(fakeRepository{
		getCashBalancesByUserID: func(_ context.Context, id uuid.UUID, cur money.Currency) ([]CashBalance, error) {
			if id != userID {
				t.Errorf("userID = %v, want %v", id, userID)
			}
			gotCurrency = cur
			return []CashBalance{{
				EntryID:         uuid.New(),
				Ticker:          "CASH-USD",
				Balance:         "1200.50000000",
				Currency:        money.USD,
				Value:           "4802000.00000000",
				DisplayCurrency: money.COP,
				FXConverted:     true,
				Movements:       3,
			}}, nil
		},
	})
	app := newTestModule(t, repo, userID, "user")

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/portfolios/cash?currency=COP", nil))
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	ok, data := decodeEnvelope(t, resp)
	if !ok {
		t.Fatal("success = false")
	}
	if gotCurrency != money.COP {
		t.Errorf("currency passed to the repository = %v, want COP", gotCurrency)
	}

	var balances []CashBalance
	if err := json.Unmarshal(data, &balances); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if len(balances) != 1 || balances[0].Ticker != "CASH-USD" || balances[0].DisplayCurrency != money.COP {
		t.Errorf("balances = %+v", balances)
	}
}

// ?currency= takes the same list as the holdings; a currency with no rate
// behind it is refused before anything is read.
func TestHandlerGetCashBalancesRejectsUnsupportedCurrency(t *testing.T) {
	app := newTestModule(t, new(fakeRepository{}), uuid.New(), "user")

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/portfolios/cash?currency=ARS", nil))
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}

func TestHandlerGetCashMovementsPaginates(t *testing.T) {
	var gotLimit, gotOffset int

	repo := new(fakeRepository{
		countCashMovements: func(context.Context, uuid.UUID) (int, error) { return 12, nil },
		getCashMovementsPaginated: func(_ context.Context, _ uuid.UUID, limit, offset int) ([]CashMovement, error) {
			gotLimit, gotOffset = limit, offset
			return []CashMovement{{ID: uuid.New(), Type: TransferIn, Kind: CashKindDeposit}}, nil
		},
	})
	app := newTestModule(t, repo, uuid.New(), "user")

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/portfolios/cash/movements?page=2&limit=5", nil))
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	_, data := decodeEnvelope(t, resp)
	var page PaginatedCashMovementsDTO
	if err := json.Unmarshal(data, &page); err != nil {
		t.Fatalf("decode data: %v", err)
	}

	if gotLimit != 5 || gotOffset != 5 {
		t.Errorf("limit, offset = %d, %d; want 5, 5", gotLimit, gotOffset)
	}
	if page.Total != 12 || page.TotalPages != 3 || len(page.Data) != 1 {
		t.Errorf("page = %+v, want 12 movements over 3 pages", page)
	}
	if page.Data[0].Kind != CashKindDeposit {
		t.Errorf("kind = %q, want deposit", page.Data[0].Kind)
	}
}

// The body the cash form sends: amounts as JSON numbers and the date as the ISO
// string a JavaScript Date serialises to.
func TestHandlerCreateCashMovement(t *testing.T) {
	userID, portfolioID, sourceID := uuid.New(), uuid.New(), uuid.New()
	var (
		got       CashMovementInput
		gotPocket uuid.UUID
	)

	repo := new(fakeRepository{
		createCashMovement: func(_ context.Context, uid, pid, sid, pocketID uuid.UUID, in CashMovementInput) (CashMovement, error) {
			if uid != userID || pid != portfolioID || sid != sourceID {
				t.Errorf("ids = %v %v %v, want %v %v %v", uid, pid, sid, userID, portfolioID, sourceID)
			}
			got, gotPocket = in, pocketID
			return CashMovement{ID: uuid.New(), Type: TransferIn, Kind: CashKindDeposit, Amount: "150000.5", Currency: money.COP}, nil
		},
	})
	app := newTestModule(t, repo, userID, "user")

	body := `{"portfolioId":"` + portfolioID.String() + `","sourceId":"` + sourceID.String() +
		`","currency":"COP","kind":"deposit","amount":150000.5,"fees":0,"date":"2026-09-10T00:00:00.000Z","notes":"nómina"}`
	resp := doJSON(t, app, http.MethodPost, "/portfolios/cash/movements", body)
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d, want 200: %s", resp.StatusCode, raw)
	}

	// No pocketId is the main account, which is where every movement went
	// before pockets existed.
	if gotPocket != (uuid.UUID{}) {
		t.Errorf("pocket = %v, want the main account", gotPocket)
	}

	if got.Kind != CashKindDeposit || got.Currency != money.COP || got.Notes != "nómina" {
		t.Errorf("input = %+v", got)
	}
	if !got.Amount.Equal(mustDecimal(t, "150000.5")) {
		t.Errorf("amount = %s, want 150000.5", got.Amount)
	}
	if want := time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC); !got.Date.Equal(want) {
		t.Errorf("date = %v, want %v", got.Date, want)
	}
}

// Every rule the input breaks is answered by the service, so the repository —
// here a hook left nil, which would panic — is never reached.
func TestHandlerCreateCashMovementRefusesBeforeTheRepository(t *testing.T) {
	portfolioID, sourceID := uuid.New(), uuid.New()
	ids := `"portfolioId":"` + portfolioID.String() + `","sourceId":"` + sourceID.String() + `",`

	cases := map[string]string{
		"interest with fees":   `{` + ids + `"currency":"USD","kind":"interest","amount":10,"fees":1,"date":"2026-09-10T00:00:00Z"}`,
		"unsupported":          `{` + ids + `"currency":"ARS","kind":"deposit","amount":10,"date":"2026-09-10T00:00:00Z"}`,
		"no amount":            `{` + ids + `"currency":"USD","kind":"deposit","amount":0,"date":"2026-09-10T00:00:00Z"}`,
		"no platform":          `{"portfolioId":"` + portfolioID.String() + `","currency":"USD","kind":"deposit","amount":10,"date":"2026-09-10T00:00:00Z"}`,
		"a read-only kind":     `{` + ids + `"currency":"USD","kind":"other","amount":10,"date":"2026-09-10T00:00:00Z"}`,
		"the transaction type": `{` + ids + `"currency":"USD","kind":"transfer_in","amount":10,"date":"2026-09-10T00:00:00Z"}`,
	}

	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			app := newTestModule(t, new(fakeRepository{}), uuid.New(), "user")

			resp := doJSON(t, app, http.MethodPost, "/portfolios/cash/movements", body)
			if resp.StatusCode != http.StatusBadRequest {
				t.Errorf("status = %d, want 400", resp.StatusCode)
			}
		})
	}
}

func TestHandlerCashMovementErrorsMapToStatuses(t *testing.T) {
	txnID := uuid.New()
	update := `{"kind":"withdrawal","amount":500,"fees":0,"date":"2026-09-10T00:00:00Z"}`

	cases := []struct {
		name   string
		err    error
		method string
		body   string
		want   int
	}{
		{"an overdraft is a conflict", ErrInsufficientCash, http.MethodPut, update, http.StatusConflict},
		{"a row that is not a plain movement", ErrCashMovementNotEditable, http.MethodPut, update, http.StatusBadRequest},
		{"someone else's movement", ErrCashMovementNotFound, http.MethodPut, update, http.StatusNotFound},
		{"deleting what a withdrawal spent", ErrInsufficientCash, http.MethodDelete, "", http.StatusConflict},
		{"deleting a missing movement", ErrCashMovementNotFound, http.MethodDelete, "", http.StatusNotFound},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := new(fakeRepository{
				updateCashMovement: func(_ context.Context, _, id uuid.UUID, _ CashMovementInput) (CashMovement, error) {
					if id != txnID {
						t.Errorf("txnID = %v, want %v", id, txnID)
					}
					return CashMovement{}, tc.err
				},
				deleteCashMovement: func(_ context.Context, _, id uuid.UUID) error {
					if id != txnID {
						t.Errorf("txnID = %v, want %v", id, txnID)
					}
					return tc.err
				},
			})
			app := newTestModule(t, repo, uuid.New(), "user")

			resp := doJSON(t, app, tc.method, "/portfolios/cash/movements/"+txnID.String(), tc.body)
			if resp.StatusCode != tc.want {
				t.Errorf("status = %d, want %d", resp.StatusCode, tc.want)
			}

			// A 4xx carries the reason, which is what the screen turns into the
			// sentence it shows.
			var env struct {
				Details string `json:"details"`
			}
			_ = json.NewDecoder(resp.Body).Decode(&env)
			if !strings.Contains(env.Details, strings.SplitN(tc.err.Error(), ":", 2)[0]) {
				t.Errorf("details = %q, want the domain error", env.Details)
			}
		})
	}
}
