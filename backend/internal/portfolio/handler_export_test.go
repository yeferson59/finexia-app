package portfolio

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

// The export endpoints stream a spreadsheet rather than the JSON envelope, so
// they are asserted on the download headers and on the workbook itself: the
// rows have to reach the file, not just the handler (docs/TECH_DEBT.md #11).

// readSheet parses the xlsx body and returns the rows of the named sheet.
func readSheet(t *testing.T, resp *http.Response, sheet string) [][]string {
	t.Helper()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	f, err := excelize.OpenReader(bytes.NewReader(body))
	if err != nil {
		t.Fatalf("open xlsx (%d bytes): %v", len(body), err)
	}
	defer func() { _ = f.Close() }()

	rows, err := f.GetRows(sheet)
	if err != nil {
		t.Fatalf("GetRows(%q): %v", sheet, err)
	}
	return rows
}

// assertSpreadsheetDownload checks the response is an attachment with the
// expected filename.
func assertSpreadsheetDownload(t *testing.T, resp *http.Response, filename string) {
	t.Helper()
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if got, want := resp.Header.Get("Content-Type"), "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"; got != want {
		t.Errorf("Content-Type = %q, want %q", got, want)
	}
	if got, want := resp.Header.Get("Content-Disposition"), `attachment; filename="`+filename+`"`; got != want {
		t.Errorf("Content-Disposition = %q, want %q", got, want)
	}
}

func TestHandlerExportSummary(t *testing.T) {
	userID := uuid.New()
	repo := &fakeRepository{
		getPortfoliosSummaryByUserID: func(context.Context, uuid.UUID) ([]SummaryView, error) {
			return []SummaryView{{
				Name: "Growth", Type: Type("stocks"), BaseCurrency: "USD", RiskName: "moderate",
				TotalPositions: 4, TotalCostBase: "900", TotalMarketValue: "1000",
				TotalGainLoss: "100", TotalGainLossPct: "11.11",
			}}, nil
		},
		getAssetAllocationByUserID: func(context.Context, uuid.UUID) ([]AllocationItem, error) {
			return []AllocationItem{{Category: Stocks, MarketValue: "1000"}}, nil
		},
	}
	app := newTestModule(t, repo, userID, "user")

	resp := do(t, app, http.MethodGet, "/portfolios/export/summary")
	assertSpreadsheetDownload(t, resp, "resumen-mensual.xlsx")

	rows := readSheet(t, resp, "Portafolios")
	if len(rows) != 2 {
		t.Fatalf("rows = %d, want a header plus one portfolio", len(rows))
	}
	if rows[0][0] != "Nombre" {
		t.Errorf("header[0] = %q, want %q", rows[0][0], "Nombre")
	}
	if rows[1][0] != "Growth" || rows[1][2] != "USD" || rows[1][3] != "moderate" {
		t.Errorf("portfolio row = %v", rows[1])
	}
}

func TestHandlerExportTransactions(t *testing.T) {
	userID := uuid.New()
	txnDate := time.Date(2026, time.April, 7, 0, 0, 0, 0, time.UTC)
	var gotLimit int
	repo := &fakeRepository{
		getRecentTransactionsByUserID: func(_ context.Context, _ uuid.UUID, limit int) ([]Transaction, error) {
			gotLimit = limit
			return []Transaction{{
				ID: uuid.New(), Type: Buy, Currency: "USD",
				TransactionDate: txnDate, Notes: "first buy",
			}}, nil
		},
	}
	app := newTestModule(t, repo, userID, "user")

	resp := do(t, app, http.MethodGet, "/portfolios/export/transactions")
	assertSpreadsheetDownload(t, resp, "transacciones.xlsx")

	// The export pulls the whole history, not the dashboard's recent slice.
	if gotLimit != 10000 {
		t.Errorf("limit = %d, want 10000", gotLimit)
	}

	rows := readSheet(t, resp, "Transacciones")
	if len(rows) != 2 {
		t.Fatalf("rows = %d, want a header plus one transaction", len(rows))
	}
	if rows[1][0] != "2026-04-07" || rows[1][1] != string(Buy) {
		t.Errorf("transaction row = %v", rows[1])
	}
}

func TestHandlerExportRiskMetrics(t *testing.T) {
	userID := uuid.New()
	var gotHasSince bool
	repo := &fakeRepository{
		getPortfolioGrowthByUserID: func(_ context.Context, _ uuid.UUID, hasSince bool, _ time.Time) ([]GrowthPoint, error) {
			gotHasSince = hasSince
			return []GrowthPoint{{
				Date:       time.Date(2026, time.January, 15, 0, 0, 0, 0, time.UTC),
				TotalValue: "1000", TotalCostBase: "900", GainLoss: "100", GainLossPct: "11.11",
			}}, nil
		},
	}
	app := newTestModule(t, repo, userID, "user")

	resp := do(t, app, http.MethodGet, "/portfolios/export/risk")
	assertSpreadsheetDownload(t, resp, "riesgo-volatilidad.xlsx")

	// The report is the full history: the handler hard-codes the ALL period.
	if gotHasSince {
		t.Error("hasSince = true, want the unbounded range")
	}

	rows := readSheet(t, resp, "Historial de crecimiento")
	if len(rows) != 2 {
		t.Fatalf("rows = %d, want a header plus one point", len(rows))
	}
	if rows[1][0] != "2026-01-15" || rows[1][1] != "1000" {
		t.Errorf("growth row = %v", rows[1])
	}
}
