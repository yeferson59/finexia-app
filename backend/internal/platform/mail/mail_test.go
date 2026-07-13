package mail

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/resend/resend-go/v3"
)

// capturedEmail holds the fields the fake Resend API received so tests can
// assert on the rendered message.
type capturedEmail struct {
	req resend.SendEmailRequest
}

// newTestService builds a mail.Service whose Resend client points at a local
// httptest server. status controls the fake API's HTTP response code; the
// captured request (if any) is written to got.
func newTestService(t *testing.T, status int, got *capturedEmail) *Service {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got != nil {
			_ = json.NewDecoder(r.Body).Decode(&got.req)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if status >= 200 && status < 300 {
			_, _ = w.Write([]byte(`{"id":"email_123"}`))
		} else {
			_, _ = w.Write([]byte(`{"statusCode":` + strconv.Itoa(status) + `,"message":"boom","name":"application_error"}`))
		}
	}))
	t.Cleanup(srv.Close)

	s, err := New("test-key", "Finexia <noreply@finexia.io>")
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Resend resolves request paths relative to BaseURL, so it needs a
	// trailing slash to behave like the default "https://api.resend.com/".
	base, err := url.Parse(srv.URL + "/")
	if err != nil {
		t.Fatalf("parse base url: %v", err)
	}
	s.client.BaseURL = base
	return s
}

func TestNewParsesTemplates(t *testing.T) {
	s, err := New("test-key", "noreply@finexia.io")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	for _, name := range []string{"waitlist_confirmation.html", "activity_alert.html", "weekly_summary.html"} {
		if s.tmpl.Lookup(name) == nil {
			t.Errorf("template %q was not parsed", name)
		}
	}
}

func TestSendWaitlistConfirmation(t *testing.T) {
	t.Run("renders the email and posts it", func(t *testing.T) {
		var got capturedEmail
		s := newTestService(t, http.StatusOK, &got)

		if err := s.SendWaitlistConfirmation("ada@example.com"); err != nil {
			t.Fatalf("SendWaitlistConfirmation: %v", err)
		}
		if len(got.req.To) != 1 || got.req.To[0] != "ada@example.com" {
			t.Errorf("To = %v, want [ada@example.com]", got.req.To)
		}
		if got.req.From != "Finexia <noreply@finexia.io>" {
			t.Errorf("From = %q", got.req.From)
		}
		if got.req.Subject == "" {
			t.Error("Subject should not be empty")
		}
		if !strings.Contains(got.req.Html, "ada@example.com") {
			t.Errorf("rendered HTML does not contain the recipient email:\n%s", got.req.Html)
		}
	})

	t.Run("surfaces API errors", func(t *testing.T) {
		s := newTestService(t, http.StatusInternalServerError, nil)
		if err := s.SendWaitlistConfirmation("ada@example.com"); err == nil {
			t.Fatal("expected an error when the API returns 500")
		}
	})
}

func TestSendActivityAlert(t *testing.T) {
	var got capturedEmail
	s := newTestService(t, http.StatusOK, &got)

	data := ActivityAlertData{
		UserName:        "Ada",
		AssetTicker:     "AAPL",
		AssetName:       "Apple Inc.",
		TransactionType: "BUY",
		Quantity:        "10",
		Price:           "192.53",
		Total:           "1925.30",
		Currency:        "USD",
		TransactionDate: "2026-07-03",
		DashboardURL:    "http://localhost:8080/dashboard",
	}

	if err := s.SendActivityAlert("ada@example.com", data); err != nil {
		t.Fatalf("SendActivityAlert: %v", err)
	}
	if !strings.Contains(got.req.Subject, "AAPL") {
		t.Errorf("Subject = %q, want it to mention the ticker", got.req.Subject)
	}
	for _, want := range []string{"Ada", "Apple Inc.", "192.53"} {
		if !strings.Contains(got.req.Html, want) {
			t.Errorf("rendered HTML missing %q", want)
		}
	}
}

func TestSendWeeklySummary(t *testing.T) {
	var got capturedEmail
	s := newTestService(t, http.StatusOK, &got)

	data := WeeklySummaryData{
		UserName:         "Ada",
		TotalValue:       "1100.00",
		TotalGainLoss:    "50.00",
		TotalGainLossPct: "4.76",
		GainLossColor:    "#22c97e",
		WeekLabel:        "Jun 26 - Jul 3",
		DashboardURL:     "http://localhost:8080/dashboard",
		Portfolios: []WeeklySummaryPortfolio{
			{Name: "Growth", Type: "stocks", TotalMarketValue: "600.00", TotalGainLoss: "100.00", TotalGainLossPct: "20.00", GainLossColor: "#22c97e"},
		},
	}

	if err := s.SendWeeklySummary("ada@example.com", data); err != nil {
		t.Fatalf("SendWeeklySummary: %v", err)
	}
	if !strings.Contains(got.req.Subject, "Jun 26 - Jul 3") {
		t.Errorf("Subject = %q, want it to include the week label", got.req.Subject)
	}
	for _, want := range []string{"1100.00", "Growth", "600.00"} {
		if !strings.Contains(got.req.Html, want) {
			t.Errorf("rendered HTML missing %q", want)
		}
	}
}
