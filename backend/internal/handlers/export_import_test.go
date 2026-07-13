package handlers

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/yeferson59/finexia-app/internal/entities"
)

func TestExportSummaryHandlerSmoke(t *testing.T) {
	userID := uuid.New()
	repo := &stubRepository{
		getPortfoliosSummaryByUserID: func(context.Context, uuid.UUID) ([]entities.PortfolioSummaryView, error) {
			return []entities.PortfolioSummaryView{}, nil
		},
		getAssetAllocationByUserID: func(context.Context, uuid.UUID) ([]entities.AllocationItem, error) {
			return []entities.AllocationItem{}, nil
		},
	}
	h := newTestHandlers(repo)
	app := fiber.New()
	app.Use(authed(userID))
	app.Get("/portfolios/export/summary", h.ExportSummary)

	req := httptest.NewRequest("GET", "/portfolios/export/summary", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if cd := resp.Header.Get(fiber.HeaderContentDisposition); cd != `attachment; filename="resumen-mensual.xlsx"` {
		t.Errorf("Content-Disposition = %q, want xlsx attachment", cd)
	}
}

// buildImportForm assembles the multipart body the import endpoints expect:
// a "file" part plus optional extra form fields.
func buildImportForm(t *testing.T, filename, content string, fields map[string]string) (*bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	if _, err := part.Write([]byte(content)); err != nil {
		t.Fatalf("write file part: %v", err)
	}
	for k, v := range fields {
		if err := w.WriteField(k, v); err != nil {
			t.Fatalf("WriteField(%s): %v", k, err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}
	return &buf, w.FormDataContentType()
}

const sampleImportCSV = "date,type,symbol,quantity,price\n" +
	"2024-01-02,buy,AAPL,1,100\n"

func TestPreviewTransactionsImportHandlerSmoke(t *testing.T) {
	userID := uuid.New()

	t.Run("parses an uploaded CSV without touching the repository", func(t *testing.T) {
		h := newTestHandlers(&stubRepository{})
		app := fiber.New()
		app.Use(authed(userID))
		app.Post("/portfolios/transactions/import/preview", h.PreviewTransactionsImport)

		body, contentType := buildImportForm(t, "transactions.csv", sampleImportCSV, nil)
		req := httptest.NewRequest("POST", "/portfolios/transactions/import/preview", body)
		req.Header.Set("Content-Type", contentType)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
	})

	t.Run("rejects a request without a file", func(t *testing.T) {
		h := newTestHandlers(&stubRepository{})
		app := fiber.New()
		app.Use(authed(userID))
		app.Post("/portfolios/transactions/import/preview", h.PreviewTransactionsImport)

		_, status, payload := doJSON(t, app, "POST", "/portfolios/transactions/import/preview", "")
		if status != fiber.StatusBadRequest {
			t.Errorf("status = %d, want 400", status)
		}
		if success, _ := payload["success"].(bool); success {
			t.Error("success should be false")
		}
	})
}

func TestImportTransactionsHandlerSmoke(t *testing.T) {
	userID := uuid.New()

	t.Run("rejects an import without a confirmed mapping", func(t *testing.T) {
		h := newTestHandlers(&stubRepository{})
		app := fiber.New()
		app.Use(authed(userID))
		app.Post("/portfolios/transactions/import", h.ImportTransactions)

		body, contentType := buildImportForm(t, "transactions.csv", sampleImportCSV, map[string]string{
			"portfolioId": uuid.NewString(),
			"sourceId":    uuid.NewString(),
		})
		req := httptest.NewRequest("POST", "/portfolios/transactions/import", body)
		req.Header.Set("Content-Type", contentType)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != fiber.StatusBadRequest {
			t.Errorf("status = %d, want 400", resp.StatusCode)
		}
	})

	t.Run("rejects an invalid portfolio id", func(t *testing.T) {
		h := newTestHandlers(&stubRepository{})
		app := fiber.New()
		app.Use(authed(userID))
		app.Post("/portfolios/transactions/import", h.ImportTransactions)

		body, contentType := buildImportForm(t, "transactions.csv", sampleImportCSV, map[string]string{
			"portfolioId": "not-a-uuid",
			"sourceId":    uuid.NewString(),
		})
		req := httptest.NewRequest("POST", "/portfolios/transactions/import", body)
		req.Header.Set("Content-Type", contentType)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != fiber.StatusBadRequest {
			t.Errorf("status = %d, want 400", resp.StatusCode)
		}
	})
}
