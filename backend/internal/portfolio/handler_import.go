package portfolio

import (
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

// maxImportFileSize bounds uploaded spreadsheets; classic personal trackers
// with a few thousand rows stay well under this.
const maxImportFileSize = 8 << 20 // 8 MiB

func readImportFile(c fiber.Ctx) ([]byte, string, error) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return nil, "", errors.New("missing file: attach the spreadsheet in the \"file\" field")
	}
	if fileHeader.Size > maxImportFileSize {
		return nil, "", errors.New("file too large: the maximum size is 8 MB")
	}
	file, err := fileHeader.Open()
	if err != nil {
		return nil, "", err
	}
	defer func(f multipart.File) { _ = f.Close() }(file)

	data, err := io.ReadAll(io.LimitReader(file, maxImportFileSize+1))
	if err != nil {
		return nil, "", err
	}
	if len(data) > maxImportFileSize {
		return nil, "", errors.New("file too large: the maximum size is 8 MB")
	}
	return data, fileHeader.Filename, nil
}

func parseImportForms(c fiber.Ctx) (*ImportMappingDTO, ImportDefaultsDTO, error) {
	var mapping *ImportMappingDTO
	if raw := c.FormValue("mapping"); raw != "" {
		mapping = &ImportMappingDTO{}
		if err := json.Unmarshal([]byte(raw), mapping); err != nil {
			return nil, ImportDefaultsDTO{}, errors.New("invalid mapping: malformed JSON")
		}
	}

	var defaults ImportDefaultsDTO
	if raw := c.FormValue("defaults"); raw != "" {
		if err := json.Unmarshal([]byte(raw), &defaults); err != nil {
			return nil, ImportDefaultsDTO{}, errors.New("invalid defaults: malformed JSON")
		}
	}
	return mapping, defaults, nil
}

// PreviewTransactionsImport parses an uploaded spreadsheet (multipart field
// "file", optional "sheet", "mapping" and "defaults" JSON fields) and returns
// headers, a suggested column mapping and per-row validation, without writing
// anything.
func (h *handler) PreviewTransactionsImport(c fiber.Ctx) error {
	if _, _, _, err := getUserIDTokenRole(c); err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	data, filename, err := readImportFile(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid file", err.Error())
	}

	mapping, defaults, err := parseImportForms(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid request", err.Error())
	}

	preview, err := h.service.PreviewTransactionImport(data, filename, c.FormValue("sheet"), mapping, defaults)
	if err != nil {
		return httpx.FromDomain(c, err, "Error parsing spreadsheet", "Could not parse the uploaded file")
	}

	return httpx.OK(c, "Import preview generated", "Spreadsheet parsed successfully", preview)
}

// ImportTransactions re-parses the uploaded spreadsheet with the confirmed
// mapping and persists every valid row into the given portfolio/platform.
func (h *handler) ImportTransactions(c fiber.Ctx) error {
	userID, _, _, err := getUserIDTokenRole(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	portfolioID, err := uuid.Parse(c.FormValue("portfolioId"))
	if err != nil {
		return httpx.BadRequest(c, "Invalid portfolio ID", "portfolioId must be a valid UUID")
	}
	sourceID, err := uuid.Parse(c.FormValue("sourceId"))
	if err != nil {
		return httpx.BadRequest(c, "Invalid source ID", "sourceId must be a valid UUID")
	}

	data, filename, err := readImportFile(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid file", err.Error())
	}

	mapping, defaults, err := parseImportForms(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid request", err.Error())
	}
	if mapping == nil {
		return httpx.BadRequest(c, "Invalid request", "mapping is required to import transactions")
	}

	result, err := h.service.ImportTransactionsFromFile(c, userID, portfolioID, sourceID, data, filename, c.FormValue("sheet"), *mapping, defaults)
	if err != nil {
		return httpx.FromDomain(c, err, "Error importing transactions", "Could not import the uploaded transactions")
	}

	return httpx.OK(c, "Transactions imported", "Spreadsheet imported successfully", result)
}
