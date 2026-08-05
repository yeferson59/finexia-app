package mail

import (
	"encoding/json"
	"html"
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

// visibleText undoes the entity escaping html/template applies (a leading "+"
// becomes "&#43;", which every mail client renders back as "+"), so the
// assertions below read the text a subscriber sees rather than the wire form.
func visibleText(rendered string) string {
	return html.UnescapeString(rendered)
}

func TestWeeklySummaryRendersTheWeeklyChange(t *testing.T) {
	base := WeeklySummaryData{
		UserName:         "Ada",
		TotalValue:       "1100.00",
		TotalGainLoss:    "50.00",
		TotalGainLossPct: "4.76",
		GainLossColor:    "#22c97e",
		WeekLabel:        "Semana 31 — 2026",
		DashboardURL:     "http://localhost:8080/dashboard",
	}

	t.Run("a gain shows the amount, the percentage and the day compared against", func(t *testing.T) {
		var got capturedEmail
		s := newTestService(t, http.StatusOK, &got)

		data := base
		data.HasWeekChange = true
		data.WeekChangeValue = "+42.50"
		data.WeekChangePct = "+4.02"
		data.WeekChangeColor = "#22c97e"
		data.WeekChangeSince = "29 jul"

		if err := s.SendWeeklySummary("ada@example.com", data); err != nil {
			t.Fatalf("SendWeeklySummary: %v", err)
		}
		body := visibleText(got.req.Html)
		for _, want := range []string{"+42.50", "(+4.02%)", "Desde el 29 jul"} {
			if !strings.Contains(body, want) {
				t.Errorf("rendered HTML missing %q", want)
			}
		}
	})

	t.Run("a loss carries its own sign", func(t *testing.T) {
		var got capturedEmail
		s := newTestService(t, http.StatusOK, &got)

		data := base
		data.HasWeekChange = true
		data.WeekChangeValue = "-118.20"
		data.WeekChangePct = "-9.71"
		data.WeekChangeColor = "#e05a5a"
		data.WeekChangeSince = "29 jul"

		if err := s.SendWeeklySummary("ada@example.com", data); err != nil {
			t.Fatalf("SendWeeklySummary: %v", err)
		}
		body := visibleText(got.req.Html)
		for _, want := range []string{"-118.20", "(-9.71%)", "#e05a5a"} {
			if !strings.Contains(body, want) {
				t.Errorf("rendered HTML missing %q", want)
			}
		}
	})

	t.Run("without a baseline the block is left out entirely", func(t *testing.T) {
		var got capturedEmail
		s := newTestService(t, http.StatusOK, &got)

		if err := s.SendWeeklySummary("ada@example.com", base); err != nil {
			t.Fatalf("SendWeeklySummary: %v", err)
		}
		if strings.Contains(got.req.Html, "Desde el") {
			t.Error("the weekly change block should be hidden with no baseline")
		}
		if !strings.Contains(got.req.Html, "1100.00") {
			t.Error("the rest of the digest should still render")
		}
	})

	t.Run("each portfolio row shows its own movement above its all-time return", func(t *testing.T) {
		var got capturedEmail
		s := newTestService(t, http.StatusOK, &got)

		data := base
		data.HasWeekChange = true
		data.WeekChangeValue = "+42.50"
		data.WeekChangePct = "+4.02"
		data.WeekChangeColor = "#22c97e"
		data.WeekChangeSince = "29 jul"
		data.Portfolios = []WeeklySummaryPortfolio{{
			Name: "Acciones USA", Type: "stocks",
			TotalMarketValue: "8200.30 USD", TotalGainLossPct: "19.40", GainLossColor: "#22c97e",
			HasWeekChange: true, WeekChangeValue: "+150.20", WeekChangePct: "+1.87", WeekChangeColor: "#22c97e",
		}}

		if err := s.SendWeeklySummary("ada@example.com", data); err != nil {
			t.Fatalf("SendWeeklySummary: %v", err)
		}
		body := visibleText(got.req.Html)
		for _, want := range []string{"+150.20", "(+1.87%)", "19.40% total", "· variación desde el 29 jul"} {
			if !strings.Contains(body, want) {
				t.Errorf("rendered HTML missing %q", want)
			}
		}
	})

	t.Run("a portfolio with no history keeps its all-time return as the coloured figure", func(t *testing.T) {
		var got capturedEmail
		s := newTestService(t, http.StatusOK, &got)

		data := base
		data.Portfolios = []WeeklySummaryPortfolio{{
			Name: "Cripto", Type: "cryptos",
			TotalMarketValue: "4250.50 USD", TotalGainLossPct: "-3.10", GainLossColor: "#e05a5a",
		}}

		if err := s.SendWeeklySummary("ada@example.com", data); err != nil {
			t.Fatalf("SendWeeklySummary: %v", err)
		}
		body := visibleText(got.req.Html)
		if strings.Contains(body, "% total") {
			t.Error("without a weekly figure the return should not be demoted to a caption")
		}
		if !strings.Contains(body, "-3.10%") {
			t.Error("rendered HTML missing the all-time return")
		}
	})

	t.Run("an amount without a usable percentage omits the parenthesis", func(t *testing.T) {
		var got capturedEmail
		s := newTestService(t, http.StatusOK, &got)

		data := base
		data.HasWeekChange = true
		data.WeekChangeValue = "+1100.00"
		data.WeekChangeColor = "#22c97e"
		data.WeekChangeSince = "29 jul"

		if err := s.SendWeeklySummary("ada@example.com", data); err != nil {
			t.Fatalf("SendWeeklySummary: %v", err)
		}
		body := visibleText(got.req.Html)
		if !strings.Contains(body, "+1100.00") {
			t.Error("rendered HTML missing the change amount")
		}
		if strings.Contains(body, "()") || strings.Contains(body, "(%)") {
			t.Error("an empty percentage should not render an empty parenthesis")
		}
	})
}
