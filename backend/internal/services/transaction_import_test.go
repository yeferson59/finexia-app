package services

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"

	portfoliodto "github.com/yeferson59/finexia-app/internal/dtos/portfolio"
	"github.com/yeferson59/finexia-app/internal/entities"
)

// buildXLSX creates an in-memory workbook with the given rows on one sheet.
func buildXLSX(t *testing.T, sheet string, rows [][]any) []byte {
	t.Helper()
	f := excelize.NewFile()
	defer func() { _ = f.Close() }()
	if sheet != "Sheet1" {
		if err := f.SetSheetName("Sheet1", sheet); err != nil {
			t.Fatalf("rename sheet: %v", err)
		}
	}
	for i, row := range rows {
		cell, _ := excelize.CoordinatesToCellName(1, i+1)
		if err := f.SetSheetRow(sheet, cell, &row); err != nil {
			t.Fatalf("set row %d: %v", i+1, err)
		}
	}
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		t.Fatalf("write workbook: %v", err)
	}
	return buf.Bytes()
}

func TestParseDecimal(t *testing.T) {
	cases := map[string]string{
		"10":         "10",
		"10.5":       "10.5",
		"10,5":       "10.5",
		"1.234,56":   "1234.56",
		"1,234.56":   "1234.56",
		"1,234":      "1234",
		"1.234.567":  "1234567",
		"$ 1,234.50": "1234.5",
		"USD 99":     "99",
		"(120.50)":   "-120.5",
		"-3,25":      "-3.25",
	}
	for input, want := range cases {
		got, err := parseDecimal(input)
		if err != nil {
			t.Errorf("parseDecimal(%q) unexpected error: %v", input, err)
			continue
		}
		if got.String() != want {
			t.Errorf("parseDecimal(%q) = %s, want %s", input, got.String(), want)
		}
	}

	for _, input := range []string{"", "abc", "--"} {
		if _, err := parseDecimal(input); err == nil {
			t.Errorf("parseDecimal(%q) expected error", input)
		}
	}
}

func TestParseImportDate(t *testing.T) {
	cases := []struct {
		input string
		order string
		want  string
	}{
		{"2024-01-15", "dmy", "2024-01-15"},
		{"2024/1/5", "dmy", "2024-01-05"},
		{"15/01/2024", "dmy", "2024-01-15"},
		{"01/15/2024", "mdy", "2024-01-15"},
		{"15/01/2024", "mdy", "2024-01-15"}, // self-corrects: 15 cannot be a month
		{"5/3/24", "dmy", "2024-03-05"},
		{"5/3/24", "mdy", "2024-05-03"},
		{"15-ene-2024", "dmy", "2024-01-15"},
		{"15 de enero de 2024", "dmy", "2024-01-15"},
		{"2024-01-15 00:00:00", "dmy", "2024-01-15"},
		{"45306", "dmy", "2024-01-15"}, // Excel serial date
	}
	for _, c := range cases {
		got, err := parseImportDate(c.input, c.order)
		if err != nil {
			t.Errorf("parseImportDate(%q, %s) unexpected error: %v", c.input, c.order, err)
			continue
		}
		if got.Format("2006-01-02") != c.want {
			t.Errorf("parseImportDate(%q, %s) = %s, want %s", c.input, c.order, got.Format("2006-01-02"), c.want)
		}
	}

	for _, input := range []string{"", "hola", "32/01/2024", "2024-13-01"} {
		if _, err := parseImportDate(input, "dmy"); err == nil {
			t.Errorf("parseImportDate(%q) expected error", input)
		}
	}
}

func TestInferDateOrder(t *testing.T) {
	if got := inferDateOrder([]string{"01/02/2024", "15/03/2024"}); got != "dmy" {
		t.Errorf("expected dmy, got %s", got)
	}
	if got := inferDateOrder([]string{"01/02/2024", "03/15/2024"}); got != "mdy" {
		t.Errorf("expected mdy, got %s", got)
	}
	if got := inferDateOrder([]string{"01/02/2024"}); got != "dmy" {
		t.Errorf("ambiguous dates should default to dmy, got %s", got)
	}
}

func TestNormalizeTxnTypeAndCategory(t *testing.T) {
	if typ, ok := normalizeTxnType("Compra"); !ok || typ != entities.Buy {
		t.Errorf("Compra should normalize to buy, got %v %v", typ, ok)
	}
	if typ, ok := normalizeTxnType("VENTA"); !ok || typ != entities.Sell {
		t.Errorf("VENTA should normalize to sell, got %v %v", typ, ok)
	}
	if typ, ok := normalizeTxnType("Dividendo"); !ok || typ != entities.Dividend {
		t.Errorf("Dividendo should normalize to dividend, got %v %v", typ, ok)
	}
	if _, ok := normalizeTxnType("???"); ok {
		t.Error("unknown type should not normalize")
	}
	if cat, ok := normalizeCategory("Acciones"); !ok || cat != entities.Stock {
		t.Errorf("Acciones should normalize to stock, got %v %v", cat, ok)
	}
	if cat, ok := normalizeCategory("Criptomonedas"); !ok || cat != entities.Crypto {
		t.Errorf("Criptomonedas should normalize to crypto, got %v %v", cat, ok)
	}
}

func TestPreviewSuggestsMappingFromSpanishHeaders(t *testing.T) {
	data := buildXLSX(t, "Mis inversiones", [][]any{
		{"Registro de inversiones 2024"}, // title row above the real table
		{},
		{"Fecha", "Operación", "Símbolo", "Empresa", "Cantidad", "Precio Unitario", "Comisión", "Moneda"},
		{"15/01/2024", "Compra", "aapl", "Apple Inc.", "10", "180,50", "1,20", "USD"},
		{"20/02/2024", "Venta", "AAPL", "Apple Inc.", "5", "195.30", "", "USD"},
		{"03/03/2024", "Dividendo", "AAPL", "Apple Inc.", "", "12", "", ""},
	})

	svc := newTestServices(&fakeRepository{}, nil)
	preview, err := svc.PreviewTransactionImport(data, "inversiones.xlsx", "", nil, portfoliodto.ImportDefaultsDTO{})
	if err != nil {
		t.Fatalf("preview: %v", err)
	}

	if preview.HeaderRow != 3 {
		t.Errorf("expected header row 3, got %d", preview.HeaderRow)
	}
	if len(preview.MissingFields) != 0 {
		t.Fatalf("expected no missing fields, got %v", preview.MissingFields)
	}
	m := preview.SuggestedMapping
	check := func(name string, got *int, want int) {
		if got == nil || *got != want {
			t.Errorf("suggested %s = %v, want %d", name, got, want)
		}
	}
	check("date", m.Date, 0)
	check("type", m.Type, 1)
	check("ticker", m.Ticker, 2)
	check("assetName", m.AssetName, 3)
	check("quantity", m.Quantity, 4)
	check("price", m.Price, 5)
	check("fees", m.Fees, 6)
	check("currency", m.Currency, 7)

	if preview.TotalRows != 3 || preview.ValidRows != 3 || preview.InvalidRows != 0 {
		t.Fatalf("expected 3 valid rows, got total=%d valid=%d invalid=%d (rows: %+v)",
			preview.TotalRows, preview.ValidRows, preview.InvalidRows, preview.Rows)
	}

	buyRow := preview.Rows[0]
	if buyRow.Type != "buy" || buyRow.Ticker != "AAPL" || buyRow.Date != "2024-01-15" ||
		buyRow.Quantity != "10" || buyRow.Price != "180.5" || buyRow.Fees != "1.2" {
		t.Errorf("unexpected normalized buy row: %+v", buyRow)
	}
	divRow := preview.Rows[2]
	if divRow.Type != "dividend" || divRow.Quantity != "0" || divRow.Currency != "USD" {
		t.Errorf("unexpected normalized dividend row: %+v", divRow)
	}
}

func TestPreviewReportsRowErrors(t *testing.T) {
	data := buildXLSX(t, "Sheet1", [][]any{
		{"Fecha", "Ticker", "Cantidad", "Precio"},
		{"15/01/2024", "AAPL", "10", "180.5"},
		{"no-es-fecha", "MSFT", "5", "300"},
		{"20/01/2024", "", "5", "300"},
		{"21/01/2024", "TSLA", "abc", "250"},
	})

	svc := newTestServices(&fakeRepository{}, nil)
	preview, err := svc.PreviewTransactionImport(data, "archivo.xlsx", "", nil, portfoliodto.ImportDefaultsDTO{})
	if err != nil {
		t.Fatalf("preview: %v", err)
	}

	if preview.ValidRows != 1 || preview.InvalidRows != 3 {
		t.Fatalf("expected 1 valid / 3 invalid, got %d/%d", preview.ValidRows, preview.InvalidRows)
	}
	if preview.Rows[1].Valid || len(preview.Rows[1].Errors) == 0 {
		t.Errorf("row with bad date should be invalid with errors: %+v", preview.Rows[1])
	}
	if preview.Rows[2].Valid || len(preview.Rows[2].Errors) == 0 {
		t.Errorf("row with empty ticker should be invalid: %+v", preview.Rows[2])
	}
	if preview.Rows[3].Valid || len(preview.Rows[3].Errors) == 0 {
		t.Errorf("row with non-numeric quantity should be invalid: %+v", preview.Rows[3])
	}
}

func TestPreviewMissingRequiredColumns(t *testing.T) {
	data := buildXLSX(t, "Sheet1", [][]any{
		{"Columna A", "Columna B"},
		{"algo", "otra cosa"},
	})

	svc := newTestServices(&fakeRepository{}, nil)
	preview, err := svc.PreviewTransactionImport(data, "archivo.xlsx", "", nil, portfoliodto.ImportDefaultsDTO{})
	if err != nil {
		t.Fatalf("preview: %v", err)
	}
	if len(preview.MissingFields) != 4 {
		t.Fatalf("expected 4 missing fields, got %v", preview.MissingFields)
	}
	if preview.ValidRows != 0 || preview.InvalidRows != 1 {
		t.Errorf("rows must not validate while required columns are unmapped: %+v", preview)
	}
}

func TestPreviewParsesCSVWithSemicolons(t *testing.T) {
	csvData := []byte("Fecha;Tipo;Ticker;Cantidad;Precio\n15/01/2024;Compra;VOO;2;430,10\n16/01/2024;Compra;VOO;1;431,00\n")

	svc := newTestServices(&fakeRepository{}, nil)
	preview, err := svc.PreviewTransactionImport(csvData, "movimientos.csv", "", nil, portfoliodto.ImportDefaultsDTO{})
	if err != nil {
		t.Fatalf("preview: %v", err)
	}
	if preview.ValidRows != 2 {
		t.Fatalf("expected 2 valid rows, got %+v", preview)
	}
	if preview.Rows[0].Price != "430.1" {
		t.Errorf("comma decimal should normalize, got %q", preview.Rows[0].Price)
	}
}

func TestImportTransactionsFromFile(t *testing.T) {
	data := buildXLSX(t, "Sheet1", [][]any{
		{"Fecha", "Tipo", "Ticker", "Cantidad", "Precio", "Comisión"},
		{"15/01/2024", "Compra", "AAPL", "10", "180.5", "1.5"},
		{"16/01/2024", "Compra", "VOO", "2", "430", ""},
		{"fecha-mala", "Compra", "MSFT", "1", "300", ""},
	})

	userID, portfolioID, sourceID := uuid.New(), uuid.New(), uuid.New()
	var captured []entities.ImportTransactionRow
	repo := &fakeRepository{
		importEntryTransactions: func(_ context.Context, gotUser, gotPortfolio, gotSource uuid.UUID, rows []entities.ImportTransactionRow) (int, error) {
			if gotUser != userID || gotPortfolio != portfolioID || gotSource != sourceID {
				t.Errorf("unexpected ids passed to repository")
			}
			captured = rows
			return len(rows), nil
		},
	}
	svc := newTestServices(repo, nil)

	mapping := portfoliodto.ImportMappingDTO{}
	col := func(i int) *int { return &i }
	mapping.Date, mapping.Type, mapping.Ticker, mapping.Quantity, mapping.Price, mapping.Fees =
		col(0), col(1), col(2), col(3), col(4), col(5)

	result, err := svc.ImportTransactionsFromFile(
		context.Background(), userID, portfolioID, sourceID,
		data, "archivo.xlsx", "", mapping, portfoliodto.ImportDefaultsDTO{Currency: "USD"},
	)
	if err != nil {
		t.Fatalf("import: %v", err)
	}

	if result.TotalRows != 3 || result.Imported != 2 || result.Skipped != 1 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if len(result.Errors) != 1 || result.Errors[0].Row != 4 {
		t.Fatalf("expected one error on sheet row 4, got %+v", result.Errors)
	}
	if len(captured) != 2 {
		t.Fatalf("expected 2 rows persisted, got %d", len(captured))
	}
	first := captured[0]
	if first.Ticker != "AAPL" || first.Type != entities.Buy || first.Quantity.String() != "10" ||
		first.Price.String() != "180.5" || first.Fees.String() != "1.5" || first.Currency != "USD" ||
		!first.Date.Equal(time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("unexpected first persisted row: %+v", first)
	}
	if first.AssetName != "AAPL" {
		t.Errorf("asset name should fall back to ticker, got %q", first.AssetName)
	}
}

func TestImportRejectsMissingMapping(t *testing.T) {
	data := buildXLSX(t, "Sheet1", [][]any{
		{"Fecha", "Ticker"},
		{"15/01/2024", "AAPL"},
	})
	svc := newTestServices(&fakeRepository{}, nil)

	_, err := svc.ImportTransactionsFromFile(
		context.Background(), uuid.New(), uuid.New(), uuid.New(),
		data, "archivo.xlsx", "", portfoliodto.ImportMappingDTO{}, portfoliodto.ImportDefaultsDTO{},
	)
	if err == nil || !strings.Contains(err.Error(), "missing required columns") {
		t.Fatalf("expected missing-columns error, got %v", err)
	}
}

func TestParseDecimalNegativeWithCurrencySymbol(t *testing.T) {
	got, err := parseDecimal("$ (120.50)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.String() != "-120.5" {
		t.Fatalf("parseDecimal(\"$ (120.50)\") = %s, want -120.5", got.String())
	}
}
