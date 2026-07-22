package portfolio

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/market"
	"github.com/yeferson59/finexia-app/internal/platform/spreadsheet"
)

type importOutcome struct {
	preview ImportPreviewResponseDTO
	valid   []ImportTransactionRow
}

// quantityRequired reports whether a transaction type moves units and thus
// needs a positive quantity; cash-flow rows (dividends, fees, interest) may
// leave it empty.
func quantityRequired(t TransactionType) bool {
	switch t {
	case Buy, Sell, TransferIn, TransferOut, Split:
		return true
	default:
		return false
	}
}

func applyImportDefaults(defaults ImportDefaultsDTO) (ImportDefaultsDTO, error) {
	out := defaults
	if strings.TrimSpace(out.Type) == "" {
		out.Type = string(Buy)
	}
	if _, ok := normalizeTxnType(out.Type); !ok {
		return out, fmt.Errorf("invalid default transaction type %q", out.Type)
	}
	if strings.TrimSpace(out.Currency) == "" {
		out.Currency = "USD"
	}
	cur, ok := normalizeCurrency(out.Currency)
	if !ok {
		return out, fmt.Errorf("invalid default currency %q", out.Currency)
	}
	out.Currency = cur
	if strings.TrimSpace(out.Category) == "" {
		out.Category = string(market.Stock)
	}
	if _, ok := market.NormalizeAssetType(out.Category); !ok {
		return out, fmt.Errorf("invalid default category %q", out.Category)
	}
	switch out.DateFormat {
	case "", "auto":
		out.DateFormat = "auto"
	case "dmy", "mdy":
	default:
		return out, fmt.Errorf("invalid date format %q", out.DateFormat)
	}
	return out, nil
}

func missingRequiredFields(m *ImportMappingDTO) []string {
	var missing []string
	for _, f := range requiredImportFields {
		var ptr *int
		switch f {
		case fieldDate:
			ptr = m.Date
		case fieldTicker:
			ptr = m.Ticker
		case fieldQuantity:
			ptr = m.Quantity
		case fieldPrice:
			ptr = m.Price
		}
		if ptr == nil {
			missing = append(missing, string(f))
		}
	}
	return missing
}

// buildImport parses the sheet with the given (or suggested) mapping and
// returns the full per-row preview plus the validated rows ready to persist.
func buildImport(src spreadsheet.Source, mapping *ImportMappingDTO, defaults ImportDefaultsDTO) (importOutcome, error) {
	defaults, err := applyImportDefaults(defaults)
	if err != nil {
		return importOutcome{}, err
	}

	headerIdx := detectHeaderRow(src.Rows)
	if headerIdx == -1 {
		return importOutcome{}, errors.New("invalid spreadsheet: the file is empty")
	}
	headers := make([]string, len(src.Rows[headerIdx]))
	for i, h := range src.Rows[headerIdx] {
		headers[i] = strings.TrimSpace(h)
	}

	suggested := suggestMapping(headers)
	if mapping == nil {
		mapping = &suggested
	}

	dataRows := src.Rows[headerIdx+1:]
	if len(dataRows) > maxImportRows {
		return importOutcome{}, fmt.Errorf("invalid spreadsheet: too many rows (max %d)", maxImportRows)
	}

	preview := ImportPreviewResponseDTO{
		Sheets:           src.Sheets,
		Sheet:            src.Sheet,
		HeaderRow:        headerIdx + 1,
		Headers:          headers,
		SuggestedMapping: suggested,
		MissingFields:    missingRequiredFields(mapping),
		Rows:             []ImportRowDTO{},
	}

	dateOrder := defaults.DateFormat
	if dateOrder == "auto" {
		var dates []string
		for _, row := range dataRows {
			if v := spreadsheet.CellAt(row, mapping.Date); v != "" {
				dates = append(dates, v)
			}
		}
		dateOrder = inferDateOrder(dates)
	}

	var valid []ImportTransactionRow
	for i, row := range dataRows {
		if spreadsheet.RowIsEmpty(row) {
			continue
		}
		rowNumber := headerIdx + 2 + i // 1-based sheet row number
		dto := ImportRowDTO{RowNumber: rowNumber, Raw: row, Errors: []string{}}
		preview.TotalRows++

		if len(preview.MissingFields) > 0 {
			preview.InvalidRows++
			preview.Rows = append(preview.Rows, dto)
			continue
		}

		entity, rowErrs := buildImportRow(row, rowNumber, mapping, defaults, dateOrder, &dto)
		if len(rowErrs) > 0 {
			dto.Errors = rowErrs
			preview.InvalidRows++
		} else {
			dto.Valid = true
			preview.ValidRows++
			valid = append(valid, entity)
		}
		preview.Rows = append(preview.Rows, dto)
	}

	return importOutcome{preview: preview, valid: valid}, nil
}

// buildImportRow normalises and validates a single data row. It fills the
// preview DTO with the normalised values and returns the persistable row
// when it has no errors.
func buildImportRow(
	row []string,
	rowNumber int,
	mapping *ImportMappingDTO,
	defaults ImportDefaultsDTO,
	dateOrder string,
	dto *ImportRowDTO,
) (ImportTransactionRow, []string) {
	var errs []string
	entity := ImportTransactionRow{RowNumber: rowNumber}

	// Transaction type: mapped column wins, empty cells fall back to default.
	typeRaw := spreadsheet.CellAt(row, mapping.Type)
	txnType, _ := normalizeTxnType(defaults.Type)
	if typeRaw != "" {
		if t, ok := normalizeTxnType(typeRaw); ok {
			txnType = t
		} else {
			errs = append(errs, fmt.Sprintf("tipo de operación no reconocido: %q", typeRaw))
		}
	}
	entity.Type = txnType
	dto.Type = string(txnType)

	// Date.
	if dateRaw := spreadsheet.CellAt(row, mapping.Date); dateRaw == "" {
		errs = append(errs, "la fecha está vacía")
	} else if date, err := parseImportDate(dateRaw, dateOrder); err != nil {
		errs = append(errs, fmt.Sprintf("fecha no reconocida: %q", dateRaw))
	} else {
		entity.Date = date
		dto.Date = date.Format("2006-01-02")
	}

	// Ticker.
	ticker := strings.ToUpper(spreadsheet.CellAt(row, mapping.Ticker))
	switch {
	case ticker == "":
		errs = append(errs, "el ticker/símbolo está vacío")
	case len(ticker) > maxTickerLen:
		errs = append(errs, fmt.Sprintf("el ticker supera %d caracteres: %q", maxTickerLen, ticker))
	default:
		entity.Ticker = ticker
		dto.Ticker = ticker
	}

	// Asset name (falls back to the ticker).
	name := spreadsheet.CellAt(row, mapping.AssetName)
	if name == "" {
		name = ticker
	}
	if len(name) > maxAssetNameLen {
		name = name[:maxAssetNameLen]
	}
	entity.AssetName = name
	dto.AssetName = name

	// Quantity.
	qtyRaw := spreadsheet.CellAt(row, mapping.Quantity)
	if qtyRaw == "" {
		if quantityRequired(txnType) {
			errs = append(errs, "la cantidad está vacía")
		} else {
			entity.Quantity = decimal.Zero
			dto.Quantity = "0"
		}
	} else if qty, err := parseDecimal(qtyRaw); err != nil {
		errs = append(errs, fmt.Sprintf("cantidad no numérica: %q", qtyRaw))
	} else if qty.InexactFloat64() < 0 || (quantityRequired(txnType) && qty.IsZero()) {
		errs = append(errs, fmt.Sprintf("la cantidad debe ser mayor que 0: %q", qtyRaw))
	} else {
		entity.Quantity = qty
		dto.Quantity = qty.String()
	}

	// Price.
	priceRaw := spreadsheet.CellAt(row, mapping.Price)
	if priceRaw == "" {
		if txnType == Buy || txnType == Sell {
			errs = append(errs, "el precio está vacío")
		} else {
			entity.Price = money.FromDecimal(decimal.Zero, money.USD)
			dto.Price = "0"
		}
	} else if price, err := parseDecimal(priceRaw); err != nil {
		errs = append(errs, fmt.Sprintf("precio no numérico: %q", priceRaw))
	} else if price.InexactFloat64() < 0 {
		errs = append(errs, fmt.Sprintf("el precio no puede ser negativo: %q", priceRaw))
	} else {
		entity.Price = money.FromDecimal(price, money.USD)
		dto.Price = price.String()
	}

	// Fees (optional).
	feesRaw := spreadsheet.CellAt(row, mapping.Fees)
	entity.Fees = money.FromDecimal(decimal.Zero, money.USD)
	dto.Fees = "0"
	if feesRaw != "" {
		if fees, err := parseDecimal(feesRaw); err != nil {
			errs = append(errs, fmt.Sprintf("comisión no numérica: %q", feesRaw))
		} else if fees.InexactFloat64() < 0 {
			errs = append(errs, fmt.Sprintf("la comisión no puede ser negativa: %q", feesRaw))
		} else {
			entity.Fees = money.FromDecimal(fees, money.USD)
			dto.Fees = fees.String()
		}
	}

	// Currency.
	currency := defaults.Currency
	if curRaw := spreadsheet.CellAt(row, mapping.Currency); curRaw != "" {
		if cur, ok := normalizeCurrency(curRaw); ok {
			currency = cur
		} else {
			errs = append(errs, fmt.Sprintf("moneda no reconocida: %q", curRaw))
		}
	}
	entity.Currency = currency
	dto.Currency = currency

	// Category / asset type.
	assetType, _ := market.NormalizeAssetType(defaults.Category)
	if catRaw := spreadsheet.CellAt(row, mapping.Category); catRaw != "" {
		if cat, ok := market.NormalizeAssetType(catRaw); ok {
			assetType = cat
		} else {
			assetType = market.Other
		}
	}
	entity.AssetType = assetType
	entity.Category = entryCategoryFor(assetType)
	dto.Category = string(assetType)

	// Notes.
	notes := spreadsheet.CellAt(row, mapping.Notes)
	if len(notes) > maxNotesLen {
		notes = notes[:maxNotesLen]
	}
	entity.Notes = notes
	dto.Notes = notes

	return entity, errs
}

// --- public service API -----------------------------------------------------

// PreviewTransactionImport parses an uploaded spreadsheet and returns its
// headers, a suggested column mapping and every row normalised + validated
// with the active mapping, without touching the database.
func (s *Service) PreviewTransactionImport(
	data []byte,
	filename, sheet string,
	mapping *ImportMappingDTO,
	defaults ImportDefaultsDTO,
) (ImportPreviewResponseDTO, error) {
	src, err := spreadsheet.ReadFile(data, filename, sheet)
	if err != nil {
		return ImportPreviewResponseDTO{}, err
	}
	out, err := buildImport(src, mapping, defaults)
	if err != nil {
		return ImportPreviewResponseDTO{}, err
	}
	if len(out.preview.Rows) > importPreviewRowCap {
		out.preview.Rows = out.preview.Rows[:importPreviewRowCap]
	}
	return out.preview, nil
}

// ImportTransactionsFromFile re-parses the uploaded file with the confirmed
// mapping and persists every valid row (asset + position + transaction) in a
// single database transaction. Invalid rows are skipped and reported back.
func (s *Service) ImportTransactionsFromFile(
	ctx context.Context,
	userID, portfolioID, sourceID uuid.UUID,
	data []byte,
	filename, sheet string,
	mapping ImportMappingDTO,
	defaults ImportDefaultsDTO,
) (ImportResultResponseDTO, error) {
	src, err := spreadsheet.ReadFile(data, filename, sheet)
	if err != nil {
		return ImportResultResponseDTO{}, err
	}
	out, err := buildImport(src, &mapping, defaults)
	if err != nil {
		return ImportResultResponseDTO{}, err
	}
	if len(out.preview.MissingFields) > 0 {
		return ImportResultResponseDTO{},
			fmt.Errorf("invalid mapping: missing required columns: %s", strings.Join(out.preview.MissingFields, ", "))
	}

	result := ImportResultResponseDTO{
		TotalRows: out.preview.TotalRows,
		Skipped:   out.preview.InvalidRows,
		Errors:    []ImportResultErrorDTO{},
	}
	for _, row := range out.preview.Rows {
		if !row.Valid && len(result.Errors) < 100 {
			result.Errors = append(result.Errors, ImportResultErrorDTO{
				Row:     row.RowNumber,
				Message: strings.Join(row.Errors, "; "),
			})
		}
	}

	if len(out.valid) == 0 {
		return result, nil
	}

	imported, err := s.repo.ImportEntryTransactions(ctx, userID, portfolioID, sourceID, out.valid)
	if err != nil {
		return ImportResultResponseDTO{}, err
	}
	result.Imported = imported
	return result, nil
}
