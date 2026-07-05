package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	portfoliodto "github.com/yeferson59/finexia-app/internal/dtos/portfolio"
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
func (s *Services) ImportAssetsFromFile(ctx context.Context, data []byte, filename, sheet string) (portfoliodto.ImportResultResponseDTO, error) {
	src, err := parseImportFile(data, filename, sheet)
	if err != nil {
		return portfoliodto.ImportResultResponseDTO{}, err
	}

	headerIdx := firstNonEmptyRow(src.rows)
	if headerIdx == -1 {
		return portfoliodto.ImportResultResponseDTO{}, errors.New("invalid spreadsheet: the file is empty")
	}

	cols := mapSimpleHeaders(src.rows[headerIdx], assetHeaderSynonyms)
	if missing := missingCols(cols, "ticker", "name", "assettype", "currency"); len(missing) > 0 {
		return portfoliodto.ImportResultResponseDTO{}, fmt.Errorf("invalid spreadsheet: missing required columns: %s", strings.Join(missing, ", "))
	}

	dataRows := src.rows[headerIdx+1:]
	if len(dataRows) > maxAssetImportRows {
		return portfoliodto.ImportResultResponseDTO{}, fmt.Errorf("invalid spreadsheet: too many rows (max %d)", maxAssetImportRows)
	}

	exchangeIdx, hasExchange := cols["exchange"]
	result := portfoliodto.ImportResultResponseDTO{Errors: []portfoliodto.ImportResultErrorDTO{}}

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
				result.Errors = append(result.Errors, portfoliodto.ImportResultErrorDTO{Row: rowNumber, Message: strings.Join(rowErrs, "; ")})
			}
			continue
		}

		if _, err := s.repos.UpsertAsset(ctx, ticker, name, assetType, exchange, currency); err != nil {
			result.Skipped++
			if len(result.Errors) < 100 {
				result.Errors = append(result.Errors, portfoliodto.ImportResultErrorDTO{Row: rowNumber, Message: err.Error()})
			}
			continue
		}
		result.Imported++
	}

	return result, nil
}
