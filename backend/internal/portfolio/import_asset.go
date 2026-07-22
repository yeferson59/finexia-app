package portfolio

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// maxAssetImportRows bounds how many data rows a single asset upload may
// contain — bulk asset catalogs are small compared to transaction history.
const maxAssetImportRows = 5000

var assetHeaderSynonyms = map[string][]string{
	"ticker":    {"ticker", "symbol", "simbolo", "codigo"},
	"name":      {"name", "nombre", "descripcion", "description"},
	"assettype": {"assettype", "asset type", "tipo", "type", "categoria", "category"},
	"exchange":  {"exchange", "bolsa", "mercado"},
	"currency":  {"currency", "moneda", "divisa", "ccy"},
}

// mapSimpleHeaders matches spreadsheet headers against a fixed set of field
// synonyms (case/accent/punctuation-insensitive via normKey) and returns the
// 0-based column index for each recognised field. Used by importers that —
// unlike the freeform transactions importer — work off a small, fixed schema
// instead of a user-driven mapping.
func mapSimpleHeaders(headers []string, synonyms map[string][]string) map[string]int {
	idx := make(map[string]int, len(synonyms))
	for i, h := range headers {
		key := normKey(h)
		if key == "" {
			continue
		}
		for field, syns := range synonyms {
			if _, assigned := idx[field]; assigned {
				continue
			}
			for _, syn := range syns {
				if key == syn {
					idx[field] = i
					break
				}
			}
		}
	}
	return idx
}

// firstNonEmptyRow returns the index of the first row with at least one
// non-blank cell, or -1 if every row is empty.
func firstNonEmptyRow(rows [][]string) int {
	for i, row := range rows {
		if !rowIsEmpty(row) {
			return i
		}
	}
	return -1
}

// cellAtIdx reads a cell by plain column index, returning "" when out of range.
func cellAtIdx(row []string, idx int) string {
	if idx < 0 || idx >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[idx])
}

// missingCols reports which of the required fields were not found in cols.
func missingCols(cols map[string]int, required ...string) []string {
	var missing []string
	for _, field := range required {
		if _, ok := cols[field]; !ok {
			missing = append(missing, field)
		}
	}
	return missing
}

// ImportAssetsFromFile parses an uploaded CSV/XLSX with columns
// ticker, name, assetType, currency (required) and exchange (optional),
// upserting one asset per valid row. Invalid rows are skipped and reported.
func (s *Service) ImportAssetsFromFile(ctx context.Context, data []byte, filename, sheet string) (ImportResultResponseDTO, error) {
	src, err := parseImportFile(data, filename, sheet)
	if err != nil {
		return ImportResultResponseDTO{}, err
	}

	headerIdx := firstNonEmptyRow(src.rows)
	if headerIdx == -1 {
		return ImportResultResponseDTO{}, errors.New("invalid spreadsheet: the file is empty")
	}

	cols := mapSimpleHeaders(src.rows[headerIdx], assetHeaderSynonyms)
	if missing := missingCols(cols, "ticker", "name", "assettype", "currency"); len(missing) > 0 {
		return ImportResultResponseDTO{}, fmt.Errorf("invalid spreadsheet: missing required columns: %s", strings.Join(missing, ", "))
	}

	dataRows := src.rows[headerIdx+1:]
	if len(dataRows) > maxAssetImportRows {
		return ImportResultResponseDTO{}, fmt.Errorf("invalid spreadsheet: too many rows (max %d)", maxAssetImportRows)
	}

	exchangeIdx, hasExchange := cols["exchange"]
	result := ImportResultResponseDTO{Errors: []ImportResultErrorDTO{}}

	for i, row := range dataRows {
		if rowIsEmpty(row) {
			continue
		}
		rowNumber := headerIdx + 2 + i
		result.TotalRows++

		ticker := strings.ToUpper(cellAtIdx(row, cols["ticker"]))
		name := cellAtIdx(row, cols["name"])
		assetTypeRaw := cellAtIdx(row, cols["assettype"])
		currency := strings.ToUpper(cellAtIdx(row, cols["currency"]))
		exchange := ""
		if hasExchange {
			exchange = cellAtIdx(row, exchangeIdx)
		}

		var rowErrs []string
		switch {
		case ticker == "":
			rowErrs = append(rowErrs, "el ticker está vacío")
		case len(ticker) > maxTickerLen:
			rowErrs = append(rowErrs, fmt.Sprintf("el ticker supera %d caracteres: %q", maxTickerLen, ticker))
		}

		if name == "" {
			name = ticker
		}
		if len(name) > maxAssetNameLen {
			name = name[:maxAssetNameLen]
		}

		assetType, ok := normalizeCategory(assetTypeRaw)
		if !ok {
			if assetTypeRaw == "" {
				rowErrs = append(rowErrs, "el tipo de activo está vacío")
			} else {
				rowErrs = append(rowErrs, fmt.Sprintf("tipo de activo no reconocido: %q", assetTypeRaw))
			}
		}

		if len(currency) != 3 {
			rowErrs = append(rowErrs, fmt.Sprintf("moneda inválida: %q", currency))
		}

		if len(rowErrs) > 0 {
			result.Skipped++
			if len(result.Errors) < 100 {
				result.Errors = append(result.Errors, ImportResultErrorDTO{Row: rowNumber, Message: strings.Join(rowErrs, "; ")})
			}
			continue
		}

		if _, err := s.repo.UpsertAsset(ctx, ticker, name, assetType, exchange, currency); err != nil {
			result.Skipped++
			if len(result.Errors) < 100 {
				result.Errors = append(result.Errors, ImportResultErrorDTO{Row: rowNumber, Message: err.Error()})
			}
			continue
		}
		result.Imported++
	}

	return result, nil
}
