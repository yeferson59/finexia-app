package support

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/yeferson59/finexia-app/internal/platform/bold"
)

func TestOrderID(t *testing.T) {
	id := newOrderID(time.UnixMilli(1_695_404_400_000))

	if len(id) > 60 {
		t.Errorf("len(%q) = %d, Bold takes at most 60", id, len(id))
	}
	if !isOrderID(id) {
		t.Errorf("isOrderID(%q) = false for an id this module minted", id)
	}
	if id == newOrderID(time.UnixMilli(1_695_404_400_000)) {
		t.Error("two ids minted in the same millisecond are equal")
	}

	for _, bad := range []string{"", "inv0334", "FNX-APOYO-1/../x", "FNX-APOYO-1695404400000-0123", id + "0"} {
		if isOrderID(bad) {
			t.Errorf("isOrderID(%q) = true", bad)
		}
	}
}

// TestTransitions walks the orders Bold's notifications can arrive in and
// checks where each sequence ends.
func TestTransitions(t *testing.T) {
	cases := []struct {
		name  string
		steps []Status
		want  Status
	}{
		{"approval", []Status{StatusApproved}, StatusApproved},
		{"retry after a rejection is approved", []Status{StatusRejected, StatusApproved}, StatusApproved},
		{"late rejection of the first attempt", []Status{StatusApproved, StatusRejected}, StatusApproved},
		{"late pending after approval", []Status{StatusApproved, StatusPending}, StatusApproved},
		{"PSE pending then rejected", []Status{StatusPending, StatusRejected}, StatusRejected},
		{"retry after a rejection is processing", []Status{StatusRejected, StatusPending}, StatusPending},
		{"void of an approval", []Status{StatusApproved, StatusVoided}, StatusVoided},
		{"nothing undoes a void", []Status{StatusVoided, StatusApproved, StatusPending, StatusRejected}, StatusVoided},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			status := StatusCreated
			for _, to := range tc.steps {
				if slices.Contains(transitionsTo(to), status) {
					status = to
				}
			}
			if status != tc.want {
				t.Fatalf("ended in %s, want %s", status, tc.want)
			}
		})
	}
}

func TestCreateCheckout(t *testing.T) {
	t.Run("stores the order and signs it", func(t *testing.T) {
		repo := newMemRepository()
		svc := newService(repo, new(fakeGateway{}), enabledConfig, nil)

		checkout, err := svc.CreateCheckout(context.Background(), 20_000)
		if err != nil {
			t.Fatalf("CreateCheckout: %v", err)
		}

		if !isOrderID(checkout.OrderID) {
			t.Errorf("OrderID = %q", checkout.OrderID)
		}
		if checkout.Amount != "20000" || checkout.Currency != "COP" || checkout.APIKey != "identidad" {
			t.Errorf("checkout = %+v", checkout)
		}
		if checkout.RedirectionURL != "https://finexia.me/apoyar" {
			t.Errorf("RedirectionURL = %q", checkout.RedirectionURL)
		}
		if checkout.RenderMode != "embedded" {
			t.Errorf("RenderMode = %q", checkout.RenderMode)
		}
		if want := bold.IntegritySignature(checkout.OrderID, 20_000, "COP", testSecret); checkout.IntegritySignature != want {
			t.Errorf("IntegritySignature = %s, want %s", checkout.IntegritySignature, want)
		}

		stored, err := repo.GetContribution(context.Background(), checkout.OrderID)
		if err != nil {
			t.Fatalf("the order was not stored: %v", err)
		}
		if stored.Amount != 20_000 || stored.Status != StatusCreated {
			t.Errorf("stored = %+v", stored)
		}
	})

	t.Run("refuses amounts out of range", func(t *testing.T) {
		svc := newService(newMemRepository(), nil, enabledConfig, nil)

		for _, amount := range []int64{0, MinAmount - 1, MaxAmount + 1} {
			if _, err := svc.CreateCheckout(context.Background(), amount); !errors.Is(err, ErrAmountOutOfRange) {
				t.Errorf("amount %d: err = %v, want ErrAmountOutOfRange", amount, err)
			}
		}
	})

	t.Run("is off without both keys", func(t *testing.T) {
		for _, cfg := range []Config{{APIKey: "k"}, {SecretKey: "s"}, {}} {
			svc := newService(newMemRepository(), nil, cfg, nil)
			if _, err := svc.CreateCheckout(context.Background(), 20_000); !errors.Is(err, ErrDisabled) {
				t.Errorf("cfg %+v: err = %v, want ErrDisabled", cfg, err)
			}
		}
	})
}

func TestContribution(t *testing.T) {
	ctx := context.Background()

	newOrder := func(t *testing.T, gw *fakeGateway) (*service, *memRepository, string) {
		t.Helper()
		repo := newMemRepository()
		svc := newService(repo, gw, enabledConfig, nil)
		checkout, err := svc.CreateCheckout(ctx, 20_000)
		if err != nil {
			t.Fatalf("CreateCheckout: %v", err)
		}
		return svc, repo, checkout.OrderID
	}

	t.Run("brings an open order up to date from Bold", func(t *testing.T) {
		gw := new(fakeGateway{payment: bold.Payment{Status: bold.StatusApproved, Total: new(int64(20_000)), TransactionID: "TX1", PaymentMethod: "PSE"}})
		svc, _, orderID := newOrder(t, gw)

		c, err := svc.Contribution(ctx, orderID)
		if err != nil {
			t.Fatalf("Contribution: %v", err)
		}
		if c.Status != StatusApproved || c.TotalCharged == nil || *c.TotalCharged != 20_000 || c.ApprovedAt == nil {
			t.Errorf("contribution = %+v", c)
		}
		if c.PaymentID != "TX1" || c.PaymentMethod != "PSE" {
			t.Errorf("payment = %q / %q", c.PaymentID, c.PaymentMethod)
		}

		// Settled now: the next read does not ask Bold again.
		if _, err := svc.Contribution(ctx, orderID); err != nil {
			t.Fatal(err)
		}
		if gw.calls != 1 {
			t.Errorf("gateway calls = %d, want 1", gw.calls)
		}
	})

	t.Run("keeps the stored status when Bold does not answer", func(t *testing.T) {
		svc, _, orderID := newOrder(t, new(fakeGateway{err: errors.New("timeout")}))

		c, err := svc.Contribution(ctx, orderID)
		if err != nil {
			t.Fatalf("Contribution: %v", err)
		}
		if c.Status != StatusCreated {
			t.Errorf("Status = %s, want created", c.Status)
		}
	})

	t.Run("an order nobody paid stays created", func(t *testing.T) {
		svc, _, orderID := newOrder(t, new(fakeGateway{payment: bold.Payment{Status: bold.StatusNoTransaction}}))

		c, err := svc.Contribution(ctx, orderID)
		if err != nil || c.Status != StatusCreated {
			t.Fatalf("contribution = %+v, err = %v", c, err)
		}
	})

	t.Run("an unknown or malformed id is not found, without asking Bold", func(t *testing.T) {
		gw := new(fakeGateway{})
		svc := newService(newMemRepository(), gw, enabledConfig, nil)

		for _, id := range []string{"inv0334", newOrderID(time.Now())} {
			if _, err := svc.Contribution(ctx, id); !errors.Is(err, ErrContributionNotFound) {
				t.Errorf("%q: err = %v, want ErrContributionNotFound", id, err)
			}
		}
		if gw.calls != 0 {
			t.Errorf("gateway calls = %d, want 0", gw.calls)
		}
	})
}

func TestHandleWebhook(t *testing.T) {
	ctx := context.Background()

	setup := func(t *testing.T, cfg Config) (*service, *memRepository, string) {
		t.Helper()
		repo := newMemRepository()
		svc := newService(repo, nil, cfg, nil)
		checkout, err := svc.CreateCheckout(ctx, 20_000)
		if err != nil {
			t.Fatalf("CreateCheckout: %v", err)
		}
		return svc, repo, checkout.OrderID
	}

	t.Run("approves the order a signed notification names", func(t *testing.T) {
		svc, repo, orderID := setup(t, enabledConfig)
		body := saleEvent("evt-1", bold.EventSaleApproved, orderID, 20_000)

		if err := svc.HandleWebhook(ctx, []byte(body), sign(body, testSecret)); err != nil {
			t.Fatalf("HandleWebhook: %v", err)
		}

		c, _ := repo.GetContribution(ctx, orderID)
		if c.Status != StatusApproved || c.PaymentID != "PAY-evt-1" {
			t.Errorf("contribution = %+v", c)
		}
	})

	t.Run("a redelivered event changes nothing the second time", func(t *testing.T) {
		svc, repo, orderID := setup(t, enabledConfig)

		approve := saleEvent("evt-a", bold.EventSaleApproved, orderID, 20_000)
		void := saleEvent("evt-v", bold.EventVoidApproved, orderID, 20_000)
		for _, body := range []string{approve, void, approve} {
			if err := svc.HandleWebhook(ctx, []byte(body), sign(body, testSecret)); err != nil {
				t.Fatalf("HandleWebhook: %v", err)
			}
		}

		c, _ := repo.GetContribution(ctx, orderID)
		if c.Status != StatusVoided {
			t.Errorf("Status = %s, want voided", c.Status)
		}
		if len(repo.events) != 2 {
			t.Errorf("events recorded = %d, want 2", len(repo.events))
		}
	})

	t.Run("refuses a bad signature and records nothing", func(t *testing.T) {
		svc, repo, orderID := setup(t, enabledConfig)
		body := saleEvent("evt-1", bold.EventSaleApproved, orderID, 20_000)

		for _, sig := range []string{"", sign(body, "otra"), sign(body, "")} {
			if err := svc.HandleWebhook(ctx, []byte(body), sig); !errors.Is(err, ErrInvalidSignature) {
				t.Errorf("signature %q: err = %v, want ErrInvalidSignature", sig, err)
			}
		}
		if len(repo.events) != 0 {
			t.Errorf("events recorded = %d, want 0", len(repo.events))
		}
	})

	t.Run("test mode verifies with the empty key, and only with it", func(t *testing.T) {
		cfg := enabledConfig
		cfg.WebhookTestMode = true
		svc, repo, orderID := setup(t, cfg)
		body := saleEvent("evt-1", bold.EventSaleApproved, orderID, 20_000)

		if err := svc.HandleWebhook(ctx, []byte(body), sign(body, testSecret)); !errors.Is(err, ErrInvalidSignature) {
			t.Fatalf("real key in test mode: err = %v, want ErrInvalidSignature", err)
		}
		if err := svc.HandleWebhook(ctx, []byte(body), sign(body, "")); err != nil {
			t.Fatalf("empty key in test mode: %v", err)
		}
		if c, _ := repo.GetContribution(ctx, orderID); c.Status != StatusApproved {
			t.Errorf("Status = %s, want approved", c.Status)
		}
	})

	t.Run("acknowledges payments that are not ours", func(t *testing.T) {
		svc, repo, _ := setup(t, enabledConfig)
		body := saleEvent("evt-link", bold.EventSaleApproved, "LNK_ABC123", 50_000)

		if err := svc.HandleWebhook(ctx, []byte(body), sign(body, testSecret)); err != nil {
			t.Fatalf("HandleWebhook: %v", err)
		}
		if ev := repo.events["evt-link"]; ev.OrderID != "" {
			t.Errorf("OrderID = %q, want empty for a foreign reference", ev.OrderID)
		}
	})

	t.Run("a signed body that is not an event is malformed", func(t *testing.T) {
		svc, _, _ := setup(t, enabledConfig)
		body := `{"hello":"world"}`

		if err := svc.HandleWebhook(ctx, []byte(body), sign(body, testSecret)); !errors.Is(err, ErrMalformedEvent) {
			t.Fatalf("err = %v, want ErrMalformedEvent", err)
		}
	})

	t.Run("is off without keys", func(t *testing.T) {
		svc := newService(newMemRepository(), nil, Config{}, nil)
		body := saleEvent("evt-1", bold.EventSaleApproved, "x", 1)

		if err := svc.HandleWebhook(ctx, []byte(body), sign(body, "")); !errors.Is(err, ErrDisabled) {
			t.Fatalf("err = %v, want ErrDisabled", err)
		}
	})
}

func TestCheckoutNeverCarriesTheSecret(t *testing.T) {
	svc := newService(newMemRepository(), nil, enabledConfig, nil)

	checkout, err := svc.CreateCheckout(context.Background(), 20_000)
	if err != nil {
		t.Fatal(err)
	}

	for _, field := range []string{checkout.OrderID, checkout.Amount, checkout.APIKey, checkout.IntegritySignature, checkout.RedirectionURL, checkout.Description} {
		if strings.Contains(field, testSecret) {
			t.Fatalf("a checkout field contains the secret key: %q", field)
		}
	}
}
