package portfolio

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"uuid"

	"github.com/gofiber/fiber/v3"
	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"
)

func TestHandlerChangeEntrySettlement(t *testing.T) {
	userID, entryID, txnID := uuid.New(), uuid.New(), uuid.New()
	path := "/portfolios/entries/" + entryID.String() + "/settlement"

	t.Run("restates the position with the rates sent", func(t *testing.T) {
		var (
			gotEntry    uuid.UUID
			gotCurrency money.Currency
			gotRates    map[uuid.UUID]decimal.Decimal
		)
		repo := new(fakeRepository{
			changeEntrySettlement: func(_ context.Context, uid, eid uuid.UUID, currency money.Currency, rates map[uuid.UUID]decimal.Decimal) (int, error) {
				if uid != userID {
					t.Errorf("userID = %s, want %s", uid, userID)
				}
				gotEntry, gotCurrency, gotRates = eid, currency, rates

				return 1, nil
			},
		})
		app := newTestModule(t, repo, userID, "user")

		body := `{"costCurrency":"USD","rates":[{"transactionId":"` + txnID.String() + `","fxRate":"1.1698"}]}`
		resp := doJSON(t, app, http.MethodPut, path, body)
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
		if gotEntry != entryID || gotCurrency != money.USD {
			t.Errorf("entry/currency = %s/%s, want %s/USD", gotEntry, gotCurrency, entryID)
		}
		if rate, ok := gotRates[txnID]; !ok || rate.String() != "1.1698" {
			t.Errorf("rates = %v, want %s at 1.1698", gotRates, txnID)
		}

		var out struct {
			Data ChangeSettlementResponseDTO `json:"data"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if out.Data.CostCurrency != "USD" || out.Data.Transactions != 1 {
			t.Errorf("response = %+v, want USD and 1 transaction", out.Data)
		}
	})

	t.Run("a refused rate is a 400", func(t *testing.T) {
		repo := new(fakeRepository{
			changeEntrySettlement: func(context.Context, uuid.UUID, uuid.UUID, money.Currency, map[uuid.UUID]decimal.Decimal) (int, error) {
				return 0, fmt.Errorf("transaction of 2026-08-18: %w", ErrTransactionFXRate)
			},
		})
		app := newTestModule(t, repo, userID, "user")

		resp := doJSON(t, app, http.MethodPut, path, `{"costCurrency":"USD"}`)
		if resp.StatusCode != fiber.StatusBadRequest {
			t.Errorf("status = %d, want 400", resp.StatusCode)
		}
	})

	t.Run("a position the caller does not own is a 404", func(t *testing.T) {
		repo := new(fakeRepository{
			changeEntrySettlement: func(context.Context, uuid.UUID, uuid.UUID, money.Currency, map[uuid.UUID]decimal.Decimal) (int, error) {
				return 0, ErrEntryNotFound
			},
		})
		app := newTestModule(t, repo, userID, "user")

		resp := doJSON(t, app, http.MethodPut, path, `{"costCurrency":"USD"}`)
		if resp.StatusCode != fiber.StatusNotFound {
			t.Errorf("status = %d, want 404", resp.StatusCode)
		}
	})

	t.Run("a missing currency or a malformed id never reaches the repository", func(t *testing.T) {
		called := false
		repo := new(fakeRepository{
			changeEntrySettlement: func(context.Context, uuid.UUID, uuid.UUID, money.Currency, map[uuid.UUID]decimal.Decimal) (int, error) {
				called = true

				return 0, nil
			},
		})
		app := newTestModule(t, repo, userID, "user")

		if resp := doJSON(t, app, http.MethodPut, path, `{"rates":[]}`); resp.StatusCode != fiber.StatusBadRequest {
			t.Errorf("missing currency: status = %d, want 400", resp.StatusCode)
		}
		if resp := doJSON(t, app, http.MethodPut, "/portfolios/entries/not-a-uuid/settlement", `{"costCurrency":"USD"}`); resp.StatusCode != fiber.StatusBadRequest {
			t.Errorf("malformed id: status = %d, want 400", resp.StatusCode)
		}
		if called {
			t.Error("an invalid request reached the repository")
		}
	})
}
