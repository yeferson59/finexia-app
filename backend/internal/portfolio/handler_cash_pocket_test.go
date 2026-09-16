package portfolio

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"uuid"

	"github.com/yeferson59/gofinance/v2/money"
)

// The pocket endpoints: what the form sends, what the service refuses before
// the repository is reached, and which status each domain error answers with.

func TestHandlerCreateCashPocket(t *testing.T) {
	userID, sourceID := uuid.New(), uuid.New()
	var got NewCashPocketInput

	repo := new(fakeRepository{
		createCashPocket: func(_ context.Context, uid uuid.UUID, in NewCashPocketInput) (CashPocket, error) {
			if uid != userID {
				t.Errorf("user = %v, want %v", uid, userID)
			}
			got = in
			return CashPocket{ID: uuid.New(), SourceID: sourceID, Currency: money.COP, Name: "Viajes", Kind: PocketFlexible}, nil
		},
	})
	app := newTestModule(t, repo, userID, "user")

	body := `{"sourceId":"` + sourceID.String() + `","currency":"COP","name":"  Viajes  "}`
	resp := doJSON(t, app, http.MethodPost, "/portfolios/cash/pockets", body)
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d, want 200: %s", resp.StatusCode, raw)
	}

	if got.SourceID != sourceID || got.Currency != money.COP {
		t.Errorf("account = %v %s, want %v COP", got.SourceID, got.Currency, sourceID)
	}
	// The name reaches the repository as typed; it is trimmed where it is
	// stored, so the unique key sees one spelling.
	if got.CleanName() != "Viajes" {
		t.Errorf("name = %q, want it to clean to Viajes", got.Name)
	}
}

// Every rule the input breaks is answered by the service, so the repository —
// here a hook left nil, which would panic — is never reached.
func TestHandlerCashPocketRefusesBeforeTheRepository(t *testing.T) {
	sourceID := uuid.New()

	cases := map[string]struct {
		method string
		path   string
		body   string
	}{
		"no platform":   {http.MethodPost, "/portfolios/cash/pockets", `{"currency":"COP","name":"Viajes"}`},
		"unsupported":   {http.MethodPost, "/portfolios/cash/pockets", `{"sourceId":"` + sourceID.String() + `","currency":"ARS","name":"Viajes"}`},
		"no name":       {http.MethodPost, "/portfolios/cash/pockets", `{"sourceId":"` + sourceID.String() + `","currency":"COP","name":"   "}`},
		"no new name":   {http.MethodPut, "/portfolios/cash/pockets/" + uuid.New().String(), `{"name":""}`},
		"move to self":  {http.MethodPost, "/portfolios/cash/movements/move", moveBody(sourceID, uuid.UUID{}, uuid.UUID{}, 1000)},
		"move nothing":  {http.MethodPost, "/portfolios/cash/movements/move", moveBody(sourceID, uuid.UUID{}, uuid.New(), 0)},
		"move no scope": {http.MethodPost, "/portfolios/cash/movements/move", `{"currency":"COP","toPocketId":"` + uuid.New().String() + `","amount":1000,"date":"2026-09-10T00:00:00Z"}`},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			app := newTestModule(t, new(fakeRepository{}), uuid.New(), "user")

			resp := doJSON(t, app, tc.method, tc.path, tc.body)
			if resp.StatusCode != http.StatusBadRequest {
				raw, _ := io.ReadAll(resp.Body)
				t.Errorf("status = %d, want 400: %s", resp.StatusCode, raw)
			}
		})
	}
}

func moveBody(sourceID, from, to uuid.UUID, amount int) string {
	body := map[string]any{
		"portfolioId": uuid.New().String(),
		"sourceId":    sourceID.String(),
		"currency":    "COP",
		"amount":      amount,
		"date":        "2026-09-10T00:00:00Z",
	}

	if from != (uuid.UUID{}) {
		body["fromPocketId"] = from.String()
	}

	if to != (uuid.UUID{}) {
		body["toPocketId"] = to.String()
	}

	raw, _ := json.Marshal(body)

	return string(raw)
}

// A pocket someone else owns is a 404, and the two blocks are 409s: money still
// in it, and a name its account already uses.
func TestHandlerCashPocketErrorsMapToStatuses(t *testing.T) {
	pocketID := uuid.New()
	rename := `{"name":"Viajes"}`

	cases := []struct {
		name   string
		err    error
		method string
		path   string
		body   string
		want   int
	}{
		{"someone else's pocket", ErrCashPocketNotFound, http.MethodPut, "/portfolios/cash/pockets/" + pocketID.String(), rename, http.StatusNotFound},
		{"deleting someone else's", ErrCashPocketNotFound, http.MethodDelete, "/portfolios/cash/pockets/" + pocketID.String(), "", http.StatusNotFound},
		{"a name already used", ErrCashPocketNameTaken, http.MethodPut, "/portfolios/cash/pockets/" + pocketID.String(), rename, http.StatusConflict},
		{"a pocket with money", ErrCashPocketNotEmpty, http.MethodDelete, "/portfolios/cash/pockets/" + pocketID.String(), "", http.StatusConflict},
		{"an inactive platform", invalidCashPocket("the platform is inactive"), http.MethodPut, "/portfolios/cash/pockets/" + pocketID.String(), rename, http.StatusBadRequest},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := new(fakeRepository{
				renameCashPocket: func(context.Context, uuid.UUID, uuid.UUID, RenameCashPocketInput) (CashPocket, error) {
					return CashPocket{}, tc.err
				},
				deleteCashPocket: func(context.Context, uuid.UUID, uuid.UUID) error { return tc.err },
			})
			app := newTestModule(t, repo, uuid.New(), "user")

			resp := doJSON(t, app, tc.method, tc.path, tc.body)
			if resp.StatusCode != tc.want {
				raw, _ := io.ReadAll(resp.Body)
				t.Errorf("status = %d, want %d: %s", resp.StatusCode, tc.want, raw)
			}
		})
	}
}

// A move that cannot be made against the balance it leaves is a 409, the same
// answer a withdrawal gets.
func TestHandlerMoveCash(t *testing.T) {
	userID, portfolioID, sourceID, pocketID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	var got CashMoveInput

	repo := new(fakeRepository{
		moveCash: func(_ context.Context, uid, pid, sid uuid.UUID, in CashMoveInput) (CashMove, error) {
			if uid != userID || pid != portfolioID || sid != sourceID {
				t.Errorf("ids = %v %v %v, want %v %v %v", uid, pid, sid, userID, portfolioID, sourceID)
			}
			got = in
			return CashMove{
				From: CashMovement{ID: uuid.New(), Kind: CashKindWithdrawal, Amount: "2000"},
				To:   CashMovement{ID: uuid.New(), Kind: CashKindDeposit, Amount: "2000"},
			}, nil
		},
	})
	app := newTestModule(t, repo, userID, "user")

	body := `{"portfolioId":"` + portfolioID.String() + `","sourceId":"` + sourceID.String() +
		`","currency":"COP","toPocketId":"` + pocketID.String() + `","amount":2000,"date":"2026-09-10T00:00:00.000Z","notes":"al viaje"}`
	resp := doJSON(t, app, http.MethodPost, "/portfolios/cash/movements/move", body)
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d, want 200: %s", resp.StatusCode, raw)
	}

	// No fromPocketId is the main account, which is the side money usually
	// leaves.
	if got.From != (uuid.UUID{}) || got.To != pocketID {
		t.Errorf("move = %v -> %v, want the main account -> %v", got.From, got.To, pocketID)
	}
	if !got.Amount.Equal(mustDecimal(t, "2000")) || got.Notes != "al viaje" {
		t.Errorf("input = %+v", got)
	}
	if want := time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC); !got.Date.Equal(want) {
		t.Errorf("date = %v, want %v", got.Date, want)
	}

	overdrawn := new(fakeRepository{
		moveCash: func(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, CashMoveInput) (CashMove, error) {
			return CashMove{}, ErrInsufficientCash
		},
	})
	resp = doJSON(t, newTestModule(t, overdrawn, userID, "user"), http.MethodPost, "/portfolios/cash/movements/move", body)
	if resp.StatusCode != http.StatusConflict {
		t.Errorf("status = %d, want 409", resp.StatusCode)
	}
}

// "move" is a path, not a movement called move: the literal route is matched
// before "/:txnId".
func TestHandlerMoveCashIsNotAMovementID(t *testing.T) {
	repo := new(fakeRepository{
		updateCashMovement: func(context.Context, uuid.UUID, uuid.UUID, CashMovementInput) (CashMovement, error) {
			t.Error("PUT reached the movement writer")
			return CashMovement{}, nil
		},
	})
	app := newTestModule(t, repo, uuid.New(), "user")

	// Nothing is registered for PUT on the literal path, so it answers 405 or
	// 404 — never the movement writer above.
	resp := doJSON(t, app, http.MethodPut, "/portfolios/cash/movements/move", `{"kind":"deposit","amount":1,"date":"2026-09-10T00:00:00Z"}`)
	if resp.StatusCode == http.StatusOK {
		t.Errorf("status = %d, want the literal path not to be read as a movement", resp.StatusCode)
	}
}

// The deposit endpoints (000048): what the form sends, what is refused before
// the repository is reached, and which status each block answers with.

// depositBody is the plan's worked example as the form sends it.
func depositBody(portfolioID, sourceID uuid.UUID) string {
	return `{"portfolioId":"` + portfolioID.String() + `","sourceId":"` + sourceID.String() + `",` +
		`"currency":"COP","name":"CDT 90 días","amount":"10000000",` +
		`"openedOn":"2026-09-01T00:00:00Z","maturesOn":"2026-11-30T00:00:00Z",` +
		`"annualRatePct":"10","withholdingPct":"4","posting":"daily","tiers":[]}`
}

func TestHandlerOpenFixedDeposit(t *testing.T) {
	userID, portfolioID, sourceID := uuid.New(), uuid.New(), uuid.New()
	pocketID := uuid.New()

	var got NewFixedDepositInput

	repo := new(fakeRepository{
		openFixedDeposit: func(_ context.Context, uid uuid.UUID, in NewFixedDepositInput) (CashPocket, error) {
			if uid != userID {
				t.Errorf("user = %v, want %v", uid, userID)
			}
			got = in

			return CashPocket{ID: pocketID, SourceID: sourceID, Currency: money.COP, Kind: PocketFixed}, nil
		},
		// Opening one computes the days it already earned, which reaches the
		// ledger through these two.
		getCashAccrualTargets: func(context.Context, time.Time, CashAccrualFilter) ([]CashAccrualTarget, error) {
			return nil, nil
		},
		getHeldCashInterest: func(context.Context, time.Time, CashAccrualFilter) ([]uuid.UUID, error) {
			return nil, nil
		},
	})
	app := newTestModule(t, repo, userID, "user")

	resp := doJSON(t, app, http.MethodPost, "/portfolios/cash/deposits", depositBody(portfolioID, sourceID))
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d, want 200: %s", resp.StatusCode, raw)
	}

	if got.PortfolioID != portfolioID || got.SourceID != sourceID || got.Currency != money.COP {
		t.Errorf("account = %v %v %s, want %v %v COP", got.PortfolioID, got.SourceID, got.Currency, portfolioID, sourceID)
	}
	if !got.Amount.Equal(mustDecimal(t, "10000000")) || !got.AnnualRatePct.Equal(mustDecimal(t, "10")) {
		t.Errorf("deposit = %s at %s %%, want 10000000 at 10 %%", got.Amount, got.AnnualRatePct)
	}
	if !got.OpenedOn.Equal(sept(1)) || got.MaturesOn == nil || !got.MaturesOn.Equal(time.Date(2026, time.November, 30, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("term = %v .. %v, want 2026-09-01 .. 2026-11-30", got.OpenedOn, got.MaturesOn)
	}
	if got.Posting != PostingDaily {
		t.Errorf("posting = %q, want daily", got.Posting)
	}
}

// Every rule the input breaks is answered by the service, so the repository —
// hooks left nil, which would panic — is never reached.
func TestHandlerFixedDepositRefusesBeforeTheRepository(t *testing.T) {
	portfolioID, sourceID, pocketID := uuid.New(), uuid.New(), uuid.New()
	deposit := depositBody(portfolioID, sourceID)

	cases := map[string]struct {
		path string
		body string
	}{
		"no portfolio":    {"/portfolios/cash/deposits", strings.Replace(deposit, `"portfolioId":"`+portfolioID.String()+`"`, `"portfolioId":null`, 1)},
		"nothing in it":   {"/portfolios/cash/deposits", strings.Replace(deposit, `"amount":"10000000"`, `"amount":"0"`, 1)},
		"no name":         {"/portfolios/cash/deposits", strings.Replace(deposit, `"name":"CDT 90 días"`, `"name":"  "`, 1)},
		"no opening day":  {"/portfolios/cash/deposits", strings.Replace(deposit, `"openedOn":"2026-09-01T00:00:00Z"`, `"openedOn":null`, 1)},
		"term before it":  {"/portfolios/cash/deposits", strings.Replace(deposit, `"maturesOn":"2026-11-30T00:00:00Z"`, `"maturesOn":"2026-09-01T00:00:00Z"`, 1)},
		"no closing day":  {"/portfolios/cash/pockets/" + pocketID.String() + "/close", `{}`},
		"a penalty below": {"/portfolios/cash/pockets/" + pocketID.String() + "/close", `{"closesOn":"` + time.Now().UTC().Format(time.DateOnly) + `T00:00:00Z","penalty":"-1"}`},
		"closed tomorrow": {"/portfolios/cash/pockets/" + pocketID.String() + "/close", `{"closesOn":"2999-01-01T00:00:00Z"}`},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			app := newTestModule(t, new(fakeRepository{}), uuid.New(), "user")

			resp := doJSON(t, app, http.MethodPost, tc.path, tc.body)
			if resp.StatusCode != http.StatusBadRequest {
				raw, _ := io.ReadAll(resp.Body)
				t.Errorf("status = %d, want 400: %s", resp.StatusCode, raw)
			}
		})
	}
}

// A deposit someone else owns is a 404; the writes it does not take, and one
// already closed, are 409s.
func TestHandlerFixedDepositErrorsMapToStatuses(t *testing.T) {
	pocketID := uuid.New()
	close := `{"closesOn":"` + time.Now().UTC().Format(time.DateOnly) + `T00:00:00Z","penalty":"0"}`

	cases := []struct {
		name string
		err  error
		want int
	}{
		{"someone else's deposit", ErrCashPocketNotFound, http.StatusNotFound},
		{"already closed", ErrCashPocketClosed, http.StatusConflict},
		{"a flexible pocket", invalidCashPocket("only a fixed deposit is cancelled"), http.StatusBadRequest},
		{"days already computed past it", cashRateInUse(sept(14), "the deposit can be cancelled from 2026-09-16"), http.StatusConflict},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := new(fakeRepository{
				endFixedDeposit: func(context.Context, uuid.UUID, uuid.UUID, time.Time) (CashPocket, error) {
					return CashPocket{}, tc.err
				},
			})
			app := newTestModule(t, repo, uuid.New(), "user")

			resp := doJSON(t, app, http.MethodPost, "/portfolios/cash/pockets/"+pocketID.String()+"/close", close)
			if resp.StatusCode != tc.want {
				raw, _ := io.ReadAll(resp.Body)
				t.Errorf("status = %d, want %d: %s", resp.StatusCode, tc.want, raw)
			}
		})
	}
}
