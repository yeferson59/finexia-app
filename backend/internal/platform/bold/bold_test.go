package bold

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIntegritySignatureMatchesBoldsExample(t *testing.T) {
	// The worked example in Bold's payment-button documentation.
	got := IntegritySignature("inv0334", 39400, "COP", "kgfq2nN0o52XqnuXZWIN2F")
	want := "620a64c6eab8858d0f96d4f818a1d77be5e9b9eb9dc681f527de1af54fc1b739"

	if got != want {
		t.Fatalf("IntegritySignature = %s, want %s", got, want)
	}
}

// The expected values were computed apart from this package, as
// hmac.new(key, base64.b64encode(body), sha256).hexdigest() in Python — the
// snippet Bold's webhook documentation gives.
func TestVerifyWebhook(t *testing.T) {
	body := []byte(`{"id":"evt-1","type":"SALE_APPROVED"}`)
	const signed = "3186a51a0ae38b88800bf0c7713789cd0f090187c4e4c0288e826ec72d4e8e80"
	const signedTestMode = "f7f8e6c9fd36440dc64f23566e83195486502a70cc7e83aff40a398c886e20c5"

	cases := []struct {
		name      string
		body      []byte
		signature string
		secret    string
		want      bool
	}{
		{"valid", body, signed, "test-secret", true},
		{"uppercase hex", body, strings.ToUpper(signed), "test-secret", true},
		{"test mode signs with an empty key", body, signedTestMode, "", true},
		{"wrong key", body, signed, "other", false},
		{"body changed after signing", []byte(`{"id":"evt-1","type":"SALE_REJECTED"}`), signed, "test-secret", false},
		{"no signature", body, "", "test-secret", false},
		{"no signature in test mode", body, "", "", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := VerifyWebhook(tc.body, tc.signature, tc.secret); got != tc.want {
				t.Fatalf("VerifyWebhook = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestParseEvent(t *testing.T) {
	t.Run("reads a sale notification", func(t *testing.T) {
		// Shaped like the example in Bold's webhook documentation, trimmed.
		body := []byte(`{
			"id": "3cf7ee8e-4ea4-4bd1-a9f4-ac1ad2a6cc6d",
			"type": "SALE_APPROVED",
			"subject": "CNPVI70CQc0EY",
			"source": "/payments",
			"spec_version": "1.0",
			"time": 1695404400000000000,
			"data": {
				"payment_id": "CNPVI70CQc0EY",
				"merchant_id": "CKKA859CGE",
				"amount": {"currency": "COP", "total": 20000, "taxes": [], "tip": 0},
				"payer_email": "someone@example.com",
				"payment_method": "CARD",
				"metadata": {"reference": "FNX-APOYO-1695404400000-0123456789abcdef"}
			},
			"datacontenttype": "application/json"
		}`)

		ev, err := ParseEvent(body)
		if err != nil {
			t.Fatalf("ParseEvent: %v", err)
		}

		if ev.ID != "3cf7ee8e-4ea4-4bd1-a9f4-ac1ad2a6cc6d" || ev.Type != EventSaleApproved || ev.PaymentID != "CNPVI70CQc0EY" {
			t.Errorf("event = %+v", ev)
		}
		if ev.Reference != "FNX-APOYO-1695404400000-0123456789abcdef" {
			t.Errorf("Reference = %q", ev.Reference)
		}
		if ev.Total == nil || *ev.Total != 20000 {
			t.Errorf("Total = %v, want 20000", ev.Total)
		}
		if ev.PaymentMethod != "CARD" {
			t.Errorf("PaymentMethod = %q", ev.PaymentMethod)
		}
	})

	t.Run("falls back to the subject for the payment id", func(t *testing.T) {
		ev, err := ParseEvent([]byte(`{"id":"e","type":"VOID_APPROVED","subject":"PAY1","data":{}}`))
		if err != nil {
			t.Fatalf("ParseEvent: %v", err)
		}
		if ev.PaymentID != "PAY1" || ev.Total != nil {
			t.Errorf("event = %+v", ev)
		}
	})

	for name, body := range map[string]string{
		"not json":      `nope`,
		"no id":         `{"type":"SALE_APPROVED","subject":"P"}`,
		"no type":       `{"id":"e","subject":"P"}`,
		"no payment id": `{"id":"e","type":"SALE_APPROVED","data":{}}`,
	} {
		t.Run("rejects "+name, func(t *testing.T) {
			if _, err := ParseEvent([]byte(body)); !errors.Is(err, ErrMalformedEvent) {
				t.Fatalf("err = %v, want ErrMalformedEvent", err)
			}
		})
	}
}

func TestClientPayment(t *testing.T) {
	t.Run("asks with the identity key and reads the voucher", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/FNX-APOYO-1" {
				t.Errorf("path = %q", r.URL.Path)
			}
			if got := r.Header.Get("Authorization"); got != "x-api-key identidad" {
				t.Errorf("Authorization = %q", got)
			}
			_, _ = w.Write([]byte(`{"transaction_id":"TX1","total":20000,"payment_method":"PSE","payment_status":"APPROVED"}`))
		}))
		defer srv.Close()

		c := New(srv.Client(), "identidad")
		c.baseURL = srv.URL + "/"

		p, err := c.Payment(context.Background(), "FNX-APOYO-1")
		if err != nil {
			t.Fatalf("Payment: %v", err)
		}
		if p.Status != StatusApproved || p.TransactionID != "TX1" || p.PaymentMethod != "PSE" {
			t.Errorf("payment = %+v", p)
		}
		if p.Total == nil || *p.Total != 20000 {
			t.Errorf("Total = %v", p.Total)
		}
	})

	t.Run("fails on anything but a 200 with a status", func(t *testing.T) {
		for name, handler := range map[string]http.HandlerFunc{
			"401": func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusUnauthorized) },
			"no status": func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(`{"total":1}`))
			},
			"not json": func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(`<html>`)) },
		} {
			t.Run(name, func(t *testing.T) {
				srv := httptest.NewServer(handler)
				defer srv.Close()

				c := New(srv.Client(), "k")
				c.baseURL = srv.URL + "/"

				if _, err := c.Payment(context.Background(), "X"); err == nil {
					t.Fatal("Payment returned no error")
				}
			})
		}
	})
}
