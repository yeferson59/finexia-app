package portfolio

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

// The export endpoints stream a spreadsheet rather than the JSON envelope, so
// they are asserted on the download headers and on the workbook itself: the
// rows have to reach the file, not just the handler.

// openWorkbook parses the xlsx body. The body can only be drained once, so a
// report with several sheets is opened here and read through sheetRows.
func openWorkbook(t *testing.T, resp *http.Response) *excelize.File {
	t.Helper()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	f, err := excelize.OpenReader(bytes.NewReader(body))
	if err != nil {
		t.Fatalf("open xlsx (%d bytes): %v", len(body), err)
	}
	t.Cleanup(func() { _ = f.Close() })
	return f
}

// sheetRows returns the rows of one sheet of an already-opened workbook.
func sheetRows(t *testing.T, f *excelize.File, sheet string) [][]string {
	t.Helper()
	rows, err := f.GetRows(sheet)
	if err != nil {
		t.Fatalf("GetRows(%q): %v", sheet, err)
	}
	return rows
}

// readSheet is the one-sheet shorthand.
func readSheet(t *testing.T, resp *http.Response, sheet string) [][]string {
	t.Helper()
	return sheetRows(t, openWorkbook(t, resp), sheet)
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
	repo := new(fakeRepository{
		getPortfoliosSummaryByUserID: func(context.Context, uuid.UUID) ([]SummaryView, error) {
			return []SummaryView{{
				Name: "Growth", Type: Type("stocks"), BaseCurrency: "USD", RiskName: "moderate",
				TotalPositions: 4, TotalCostBase: "900", TotalMarketValue: "1000",
				TotalGainLoss: "100", TotalGainLossPct: "11.11",
			}}, nil
		},
		getAssetAllocationByUserID: func(context.Context, uuid.UUID, string) ([]AllocationItem, error) {
			return []AllocationItem{{Category: Stocks, MarketValue: "1000"}}, nil
		},
	})
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
	repo := new(fakeRepository{
		getRecentTransactionsByUserID: func(_ context.Context, _ uuid.UUID, limit int) ([]Transaction, error) {
			gotLimit = limit
			return []Transaction{{
				ID: uuid.New(), Type: Buy, Currency: "USD",
				TransactionDate: txnDate, Notes: "first buy",
			}}, nil
		},
	})
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

// cellFor finds a row of the metrics sheet by its label and returns its value.
func cellFor(t *testing.T, rows [][]string, label string) string {
	t.Helper()
	for _, row := range rows {
		if len(row) > 1 && row[0] == label {
			return row[1]
		}
	}
	t.Fatalf("no hay fila %q en %v", label, rows)
	return ""
}

// riskSeries is a month of daily points with one contribution in the middle:
// enough history for the risk figures, and a flow to prove they ignore it.
func riskSeries() []GrowthPoint {
	points := make([]GrowthPoint, 0, 31)
	for i := range 31 {
		value := 1000 + i*4
		flow := "0"
		if i == 15 {
			// A deposit that doubles the account without earning anything.
			value += 1000
			flow = "1000"
		}
		if i > 15 {
			value += 1000
		}
		points = append(points, GrowthPoint{
			Date:          time.Date(2026, time.January, 1+i, 0, 0, 0, 0, time.UTC),
			Currency:      "USD",
			TotalValue:    strconv.Itoa(value),
			TotalCostBase: "900",
			GainLoss:      "100",
			GainLossPct:   "11.11",
			NetFlow:       flow,
		})
	}
	return points
}

func TestHandlerExportRiskMetrics(t *testing.T) {
	userID := uuid.New()
	var gotHasSince bool
	repo := new(fakeRepository{
		getPortfolioGrowthByUserID: func(_ context.Context, _ uuid.UUID, _ string, hasSince bool, _ time.Time) ([]GrowthPoint, error) {
			gotHasSince = hasSince
			return riskSeries(), nil
		},
	})
	app := newTestModule(t, repo, userID, "user")

	resp := do(t, app, http.MethodGet, "/portfolios/export/risk")
	assertSpreadsheetDownload(t, resp, "riesgo-volatilidad.xlsx")

	// The report is the full history: the handler hard-codes the ALL period.
	if gotHasSince {
		t.Error("hasSince = true, want the unbounded range")
	}

	book := openWorkbook(t, resp)

	// The file is named after metrics, so it has to contain them and not just
	// the series they come from.
	metrics := sheetRows(t, book, "Métricas de riesgo")
	if got := cellFor(t, metrics, "Puntos de la serie"); got != "31" {
		t.Errorf("points = %q, want 31", got)
	}
	if got := cellFor(t, metrics, "Volatilidad anualizada"); got == "Sin historial suficiente" {
		t.Error("volatility withheld on a month of daily points")
	}
	if got := cellFor(t, metrics, "Ratio de Sharpe"); got == "Sin historial suficiente" {
		t.Error("sharpe withheld on a month of daily points")
	}
	// Thirty days is short of the ninety the annualized figure needs, and the
	// cell says so rather than leaving a blank that reads as zero.
	if got := cellFor(t, metrics, "Rentabilidad anualizada"); got != "Sin historial suficiente" {
		t.Errorf("annualized = %q, want the shortfall spelled out", got)
	}
	// The deposit of day 16 doubled the account; the return must not.
	total, err := strconv.ParseFloat(cellFor(t, metrics, "Rentabilidad del periodo"), 64)
	if err != nil {
		t.Fatalf("total return is not a number: %v", err)
	}
	if total > 20 {
		t.Errorf("total return = %.2f%%, the deposit leaked into the return", total)
	}

	monthly := sheetRows(t, book, "Rentabilidad mensual")
	if len(monthly) != 2 || monthly[1][0] != "2026-01" {
		t.Errorf("monthly sheet = %v, want a header plus January", monthly)
	}

	rows := sheetRows(t, book, "Historial de crecimiento")
	if len(rows) != 32 {
		t.Fatalf("rows = %d, want a header plus 31 points", len(rows))
	}
	if rows[1][0] != "2026-01-01" || rows[1][1] != "1000" {
		t.Errorf("growth row = %v", rows[1])
	}
	// The flow column is what makes the metrics auditable from the raw series.
	if rows[16][5] != "1000" {
		t.Errorf("net flow on the contribution day = %q, want 1000", rows[16][5])
	}
}

func TestHandlerExportRiskMetricsWithoutHistory(t *testing.T) {
	repo := new(fakeRepository{
		getPortfolioGrowthByUserID: func(context.Context, uuid.UUID, string, bool, time.Time) ([]GrowthPoint, error) {
			return nil, nil
		},
	})
	app := newTestModule(t, repo, uuid.New(), "user")

	resp := do(t, app, http.MethodGet, "/portfolios/export/risk")
	assertSpreadsheetDownload(t, resp, "riesgo-volatilidad.xlsx")

	// An empty account still gets a readable file, saying why it is empty.
	metrics := readSheet(t, resp, "Métricas de riesgo")
	if got := cellFor(t, metrics, "Historial"); got != "Sin historial suficiente" {
		t.Errorf("empty report says %q", got)
	}
}
