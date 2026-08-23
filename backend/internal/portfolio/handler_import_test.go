package portfolio

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"uuid"

	"github.com/gofiber/fiber/v3"
)

// The import endpoints are the module's only multipart surface, so their
// handler layer (file field, size cap, JSON side-forms, required mapping) is
// worth covering separately from the parser, which transaction_import_test.go
// already exercises.

// importForm builds a multipart body. A field whose value is empty is skipped,
// so a test can leave one out; a nil file omits the file part entirely.
func importForm(t *testing.T, file []byte, filename string, fields map[string]string) (*bytes.Buffer, string) {
	t.Helper()
	body := new(bytes.Buffer{})
	w := multipart.NewWriter(body)

	if file != nil {
		part, err := w.CreateFormFile("file", filename)
		if err != nil {
			t.Fatalf("CreateFormFile: %v", err)
		}
		if _, err := part.Write(file); err != nil {
			t.Fatalf("write file part: %v", err)
		}
	}
	for name, value := range fields {
		if value == "" {
			continue
		}
		if err := w.WriteField(name, value); err != nil {
			t.Fatalf("WriteField(%q): %v", name, err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}
	return body, w.FormDataContentType()
}

func doMultipart(t *testing.T, app *fiber.App, target string, body *bytes.Buffer, contentType string) *http.Response {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, target, body)
	req.Header.Set("Content-Type", contentType)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("POST %s: %v", target, err)
	}
	return resp
}

// csvImport is a two-row spreadsheet in the semicolon/comma-decimal layout the
// Spanish-language exports use.
var csvImport = []byte("Fecha;Tipo;Ticker;Cantidad;Precio\n15/01/2024;Compra;VOO;2;430,10\n16/01/2024;Compra;VOO;1;431,00\n")

// fullMapping maps every column of csvImport, as the confirmed mapping the
// import step requires.
const fullMapping = `{"date":0,"type":1,"ticker":2,"quantity":3,"price":4}`

func TestHandlerPreviewTransactionsImport(t *testing.T) {
	userID := uuid.New()

	t.Run("parses the upload and suggests a mapping", func(t *testing.T) {
		app := newTestModule(t, new(fakeRepository{}), userID, "user")

		body, contentType := importForm(t, csvImport, "movimientos.csv", nil)
		resp := doMultipart(t, app, "/portfolios/transactions/import/preview", body, contentType)
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}

		_, data := decodeEnvelope(t, resp)
		var preview ImportPreviewResponseDTO
		if err := json.Unmarshal(data, &preview); err != nil {
			t.Fatalf("decode data: %v (%s)", err, data)
		}
		if len(preview.Headers) != 5 || preview.Headers[0] != "Fecha" {
			t.Errorf("headers = %v", preview.Headers)
		}
		if preview.SuggestedMapping.Ticker == nil || *preview.SuggestedMapping.Ticker != 2 {
			t.Errorf("suggested ticker column = %v, want 2", preview.SuggestedMapping.Ticker)
		}
	})

	t.Run("requires the file field", func(t *testing.T) {
		app := newTestModule(t, new(fakeRepository{}), userID, "user")

		body, contentType := importForm(t, nil, "", map[string]string{"sheet": "Sheet1"})
		resp := doMultipart(t, app, "/portfolios/transactions/import/preview", body, contentType)
		if resp.StatusCode != fiber.StatusBadRequest {
			t.Fatalf("status = %d, want 400 without a file", resp.StatusCode)
		}
	})

	t.Run("rejects a malformed mapping", func(t *testing.T) {
		app := newTestModule(t, new(fakeRepository{}), userID, "user")

		body, contentType := importForm(t, csvImport, "movimientos.csv", map[string]string{"mapping": `{`})
		resp := doMultipart(t, app, "/portfolios/transactions/import/preview", body, contentType)
		if resp.StatusCode != fiber.StatusBadRequest {
			t.Fatalf("status = %d, want 400", resp.StatusCode)
		}
	})

	t.Run("rejects a malformed defaults", func(t *testing.T) {
		app := newTestModule(t, new(fakeRepository{}), userID, "user")

		body, contentType := importForm(t, csvImport, "movimientos.csv", map[string]string{"defaults": `{`})
		resp := doMultipart(t, app, "/portfolios/transactions/import/preview", body, contentType)
		if resp.StatusCode != fiber.StatusBadRequest {
			t.Fatalf("status = %d, want 400", resp.StatusCode)
		}
	})
}

func TestHandlerImportTransactions(t *testing.T) {
	userID := uuid.New()
	portfolioID := uuid.New()
	sourceID := uuid.New()

	baseFields := func() map[string]string {
		return map[string]string{
			"portfolioId": portfolioID.String(),
			"sourceId":    sourceID.String(),
			"mapping":     fullMapping,
			"defaults":    `{"currency":"USD"}`,
		}
	}

	t.Run("persists the mapped rows", func(t *testing.T) {
		var persisted []ImportTransactionRow
		repo := new(fakeRepository{
			importEntryTransactions: func(_ context.Context, uid, pid, sid uuid.UUID, rows []ImportTransactionRow) (int, error) {
				if uid != userID || pid != portfolioID || sid != sourceID {
					t.Errorf("ids = %s/%s/%s, want %s/%s/%s", uid, pid, sid, userID, portfolioID, sourceID)
				}
				persisted = rows
				return len(rows), nil
			},
		})
		app := newTestModule(t, repo, userID, "user")

		body, contentType := importForm(t, csvImport, "movimientos.csv", baseFields())
		resp := doMultipart(t, app, "/portfolios/transactions/import", body, contentType)
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
		if len(persisted) != 2 {
			t.Fatalf("persisted %d rows, want 2", len(persisted))
		}
		if persisted[0].Ticker != "VOO" {
			t.Errorf("first row ticker = %q, want VOO", persisted[0].Ticker)
		}

		_, data := decodeEnvelope(t, resp)
		var result ImportResultResponseDTO
		if err := json.Unmarshal(data, &result); err != nil {
			t.Fatalf("decode data: %v (%s)", err, data)
		}
		if result.Imported != 2 || result.Skipped != 0 {
			t.Errorf("result = %+v", result)
		}
	})

	t.Run("requires a confirmed mapping", func(t *testing.T) {
		app := newTestModule(t, new(fakeRepository{}), userID, "user")

		fields := baseFields()
		delete(fields, "mapping")
		body, contentType := importForm(t, csvImport, "movimientos.csv", fields)
		resp := doMultipart(t, app, "/portfolios/transactions/import", body, contentType)
		if resp.StatusCode != fiber.StatusBadRequest {
			t.Fatalf("status = %d, want 400 without a mapping", resp.StatusCode)
		}
	})

	t.Run("rejects non-uuid target ids", func(t *testing.T) {
		app := newTestModule(t, new(fakeRepository{}), userID, "user")

		for _, field := range []string{"portfolioId", "sourceId"} {
			fields := baseFields()
			fields[field] = "not-a-uuid"
			body, contentType := importForm(t, csvImport, "movimientos.csv", fields)
			resp := doMultipart(t, app, "/portfolios/transactions/import", body, contentType)
			if resp.StatusCode != fiber.StatusBadRequest {
				t.Errorf("%s: status = %d, want 400", field, resp.StatusCode)
			}
		}
	})
}
