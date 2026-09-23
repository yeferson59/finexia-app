package support

import (
	"encoding/json"
	"errors"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/yeferson59/finexia-app/internal/platform/bold"
	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

func newTestApp(svc *service) *fiber.App {
	return newGuardedApp(svc, adminGuard)
}

func newGuardedApp(svc *service, guard fakeGuard) *fiber.App {
	app := fiber.New()
	new(Module{service: svc, handler: newHandler(svc), authMiddl: guard, limiter: httpx.OrPassThrough(nil)}).Routes(app)

	return app
}

func call(t *testing.T, app *fiber.App, method, path, body string, headers map[string]string) (int, map[string]any) {
	t.Helper()

	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	raw, _ := io.ReadAll(resp.Body)
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("invalid JSON response %q: %v", raw, err)
	}

	return resp.StatusCode, payload
}

func TestConfigRoute(t *testing.T) {
	for name, tc := range map[string]struct {
		cfg  Config
		want bool
	}{
		"enabled":  {enabledConfig, true},
		"disabled": {Config{}, false},
	} {
		t.Run(name, func(t *testing.T) {
			status, payload := call(t, newTestApp(newService(newMemRepository(), nil, tc.cfg, nil)), "GET", "/support", "", nil)
			if status != fiber.StatusOK {
				t.Fatalf("status = %d", status)
			}
			data, _ := payload["data"].(map[string]any)
			if data["enabled"] != tc.want || data["minAmount"] != float64(MinAmount) || data["currency"] != "COP" {
				t.Errorf("data = %v", data)
			}
		})
	}
}

func TestCheckoutRoutes(t *testing.T) {
	t.Run("creates and then reads an order", func(t *testing.T) {
		app := newTestApp(newService(newMemRepository(), nil, enabledConfig, nil))

		status, payload := call(t, app, "POST", "/support/checkouts", `{"amount":20000}`, nil)
		if status != fiber.StatusCreated {
			t.Fatalf("create status = %d, payload = %v", status, payload)
		}
		data, _ := payload["data"].(map[string]any)
		orderID, _ := data["orderId"].(string)
		if !isOrderID(orderID) || data["amount"] != "20000" || data["integritySignature"] == "" {
			t.Fatalf("checkout = %v", data)
		}

		status, payload = call(t, app, "GET", "/support/checkouts/"+orderID, "", nil)
		if status != fiber.StatusOK {
			t.Fatalf("read status = %d", status)
		}
		data, _ = payload["data"].(map[string]any)
		if data["status"] != "created" || data["amount"] != float64(20000) {
			t.Errorf("contribution = %v", data)
		}
		if _, leaked := data["PaymentID"]; leaked {
			t.Error("the payment id is internal and must not be serialised")
		}
	})

	t.Run("answers 400 to a bad amount", func(t *testing.T) {
		app := newTestApp(newService(newMemRepository(), nil, enabledConfig, nil))

		for _, body := range []string{`{"amount":999}`, `{"amount":2000001}`, `{"amount":"20000"}`, `nope`} {
			if status, _ := call(t, app, "POST", "/support/checkouts", body, nil); status != fiber.StatusBadRequest {
				t.Errorf("%s: status = %d, want 400", body, status)
			}
		}
	})

	t.Run("answers 503 when contributions are off", func(t *testing.T) {
		app := newTestApp(newService(newMemRepository(), nil, Config{}, nil))

		if status, _ := call(t, app, "POST", "/support/checkouts", `{"amount":20000}`, nil); status != fiber.StatusServiceUnavailable {
			t.Errorf("status = %d, want 503", status)
		}
	})

	t.Run("answers 404 to an unknown order", func(t *testing.T) {
		app := newTestApp(newService(newMemRepository(), nil, enabledConfig, nil))

		if status, _ := call(t, app, "GET", "/support/checkouts/inv0334", "", nil); status != fiber.StatusNotFound {
			t.Errorf("status = %d, want 404", status)
		}
	})
}

func TestWebhookRoute(t *testing.T) {
	post := func(t *testing.T, app *fiber.App, body, signature string) int {
		t.Helper()
		status, _ := call(t, app, "POST", "/support/webhooks/bold", body, map[string]string{bold.SignatureHeader: signature})
		return status
	}

	body := saleEvent("evt-1", bold.EventSaleApproved, newOrderID(time.Now()), 20_000)

	t.Run("200 for a signed notification", func(t *testing.T) {
		app := newTestApp(newService(newMemRepository(), nil, enabledConfig, nil))
		if status := post(t, app, body, sign(body, testSecret)); status != fiber.StatusOK {
			t.Fatalf("status = %d, want 200", status)
		}
	})

	t.Run("401 for a bad signature", func(t *testing.T) {
		app := newTestApp(newService(newMemRepository(), nil, enabledConfig, nil))
		if status := post(t, app, body, sign(body, "otra")); status != fiber.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", status)
		}
	})

	t.Run("400 for a signed body that is not an event", func(t *testing.T) {
		app := newTestApp(newService(newMemRepository(), nil, enabledConfig, nil))
		junk := `{"hello":"world"}`
		if status := post(t, app, junk, sign(junk, testSecret)); status != fiber.StatusBadRequest {
			t.Fatalf("status = %d, want 400", status)
		}
	})

	t.Run("503 when the keys are missing, so Bold retries", func(t *testing.T) {
		app := newTestApp(newService(newMemRepository(), nil, Config{}, nil))
		if status := post(t, app, body, sign(body, "")); status != fiber.StatusServiceUnavailable {
			t.Fatalf("status = %d, want 503", status)
		}
	})

	t.Run("500 when it cannot be stored, so Bold retries", func(t *testing.T) {
		repo := newMemRepository()
		repo.failRecord = errors.New("connection refused")
		app := newTestApp(newService(repo, nil, enabledConfig, nil))
		if status := post(t, app, body, sign(body, testSecret)); status != fiber.StatusInternalServerError {
			t.Fatalf("status = %d, want 500", status)
		}
	})
}

func TestAdminRoutes(t *testing.T) {
	now := time.Now()
	seed := func() *memRepository {
		repo := newMemRepository()
		repo.plant(Contribution{OrderID: "a", Amount: 20_000, Status: StatusApproved, TotalCharged: new(int64(20_000)), PaymentID: "PAY-A", PaymentMethod: "CARD", ApprovedAt: new(now), CreatedAt: now.Add(-time.Hour)})
		repo.plant(Contribution{OrderID: "b", Amount: 50_000, Status: StatusApproved, ApprovedAt: new(now.Add(-40 * 24 * time.Hour)), CreatedAt: now.Add(-40 * 24 * time.Hour)})
		repo.plant(Contribution{OrderID: "c", Amount: 10_000, Status: StatusRejected, CreatedAt: now.Add(-2 * time.Hour)})
		repo.plant(Contribution{OrderID: "d", Amount: 10_000, Status: StatusCreated, CreatedAt: now})
		return repo
	}

	t.Run("only an admin gets through", func(t *testing.T) {
		for name, tc := range map[string]struct {
			guard fakeGuard
			want  int
		}{
			"no session": {fakeGuard{authStatus: fiber.StatusUnauthorized}, fiber.StatusUnauthorized},
			"customer":   {fakeGuard{role: "customer"}, fiber.StatusForbidden},
		} {
			t.Run(name, func(t *testing.T) {
				app := newGuardedApp(newService(seed(), nil, enabledConfig, nil), tc.guard)
				for _, path := range []string{"/support/contributions", "/support/contributions/summary"} {
					req := httptest.NewRequest("GET", path, nil)
					resp, err := app.Test(req)
					if err != nil {
						t.Fatal(err)
					}
					_ = resp.Body.Close()
					if resp.StatusCode != tc.want {
						t.Errorf("%s: status = %d, want %d", path, resp.StatusCode, tc.want)
					}
				}
			})
		}
	})

	t.Run("the payer's routes stay public", func(t *testing.T) {
		app := newGuardedApp(newService(seed(), nil, enabledConfig, nil), fakeGuard{authStatus: fiber.StatusUnauthorized})
		if status, _ := call(t, app, "GET", "/support", "", nil); status != fiber.StatusOK {
			t.Errorf("GET /support status = %d, want 200", status)
		}
	})

	t.Run("lists newest first, with the payment's details", func(t *testing.T) {
		app := newTestApp(newService(seed(), nil, enabledConfig, nil))

		status, payload := call(t, app, "GET", "/support/contributions?page=1&limit=2", "", nil)
		if status != fiber.StatusOK {
			t.Fatalf("status = %d, payload = %v", status, payload)
		}
		data, _ := payload["data"].(map[string]any)
		items, _ := data["items"].([]any)
		if len(items) != 2 {
			t.Fatalf("items = %v", items)
		}
		first, _ := items[0].(map[string]any)
		if first["orderId"] != "d" {
			t.Errorf("first = %v, want the newest order", first["orderId"])
		}
		second, _ := items[1].(map[string]any)
		if second["paymentId"] != "PAY-A" || second["paymentMethod"] != "CARD" {
			t.Errorf("second = %v", second)
		}
		meta, _ := data["metaData"].(map[string]any)
		if meta["total"] != float64(4) || meta["next"] != true {
			t.Errorf("metaData = %v", meta)
		}
	})

	t.Run("filters by status and refuses an unknown one", func(t *testing.T) {
		app := newTestApp(newService(seed(), nil, enabledConfig, nil))

		_, payload := call(t, app, "GET", "/support/contributions?status=approved", "", nil)
		data, _ := payload["data"].(map[string]any)
		if items, _ := data["items"].([]any); len(items) != 2 {
			t.Errorf("approved items = %d, want 2", len(items))
		}

		if status, _ := call(t, app, "GET", "/support/contributions?status=paid", "", nil); status != fiber.StatusBadRequest {
			t.Errorf("status = %d, want 400", status)
		}
	})

	t.Run("summarizes what was approved", func(t *testing.T) {
		app := newTestApp(newService(seed(), nil, enabledConfig, nil))

		status, payload := call(t, app, "GET", "/support/contributions/summary", "", nil)
		if status != fiber.StatusOK {
			t.Fatalf("status = %d", status)
		}
		data, _ := payload["data"].(map[string]any)
		// b has no total from Bold: its own amount counts. It is also older than
		// the recent window.
		if data["approvedTotal"] != float64(70_000) || data["approvedRecent"] != float64(20_000) {
			t.Errorf("totals = %v / %v", data["approvedTotal"], data["approvedRecent"])
		}
		counts, _ := data["counts"].(map[string]any)
		if counts["approved"] != float64(2) || counts["voided"] != float64(0) {
			t.Errorf("counts = %v", counts)
		}
	})
}
