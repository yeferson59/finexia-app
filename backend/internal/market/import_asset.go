package market

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
	"github.com/yeferson59/finexia-app/internal/platform/spreadsheet"
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

// ImportAssetsFromFile parses an uploaded CSV/XLSX with columns
// ticker, name, assetType, currency (required) and exchange (optional),
// upserting one asset per valid row. Invalid rows are skipped and reported.
func (s *Service) ImportAssetsFromFile(ctx context.Context, data []byte, filename, sheet string) (ImportResultResponseDTO, error) {
	src, err := spreadsheet.ReadFile(data, filename, sheet)
	if err != nil {
		return ImportResultResponseDTO{}, httpx.AsBadRequest(err)
	}

	headerIdx := firstNonEmptyRow(src.Rows)
	if headerIdx == -1 {
		return ImportResultResponseDTO{}, httpx.AsBadRequest(errors.New("invalid spreadsheet: the file is empty"))
	}

	cols := mapSimpleHeaders(src.Rows[headerIdx], assetHeaderSynonyms)
	if missing := missingCols(cols, "ticker", "name", "assettype", "currency"); len(missing) > 0 {
		return ImportResultResponseDTO{}, httpx.AsBadRequest(fmt.Errorf("invalid spreadsheet: missing required columns: %s", strings.Join(missing, ", ")))
	}

	dataRows := src.Rows[headerIdx+1:]
	if len(dataRows) > maxAssetImportRows {
		return ImportResultResponseDTO{}, httpx.AsTooManyRequests(fmt.Errorf("invalid spreadsheet: too many rows (max %d)", maxAssetImportRows))
	}

	exchangeIdx, hasExchange := cols["exchange"]
	result := ImportResultResponseDTO{Errors: []ImportResultErrorDTO{}}

	for i, row := range dataRows {
		if spreadsheet.RowIsEmpty(row) {
			continue
		}
		rowNumber := headerIdx + 2 + i
		result.TotalRows++

		ticker := strings.ToUpper(spreadsheet.CellAtIdx(row, cols["ticker"]))
		name := spreadsheet.CellAtIdx(row, cols["name"])
		assetTypeRaw := spreadsheet.CellAtIdx(row, cols["assettype"])
		currency := strings.ToUpper(spreadsheet.CellAtIdx(row, cols["currency"]))
		exchange := ""
		if hasExchange {
			exchange = spreadsheet.CellAtIdx(row, exchangeIdx)
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

		assetType, ok := NormalizeAssetType(assetTypeRaw)
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
