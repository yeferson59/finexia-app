package market

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
	"github.com/yeferson59/finexia-app/internal/platform/spreadsheet"
)

// maxExchangeRateImportRows bounds how many data rows a single exchange-rate
// upload may contain.
const maxExchangeRateImportRows = 5000

var exchangeRateHeaderSynonyms = map[string][]string{
	"fromcurrency": {"fromcurrency", "from currency", "from", "origen", "moneda origen", "desde"},
	"tocurrency":   {"tocurrency", "to currency", "to", "destino", "moneda destino", "hacia"},
	"rate":         {"rate", "tasa", "valor", "value", "precio"},
}

// ImportExchangeRatesFromFile parses an uploaded CSV/XLSX with columns
// fromCurrency, toCurrency and rate, upserting one currency pair per valid
// row (dated today). Invalid rows are skipped and reported.
func (s *Service) ImportExchangeRatesFromFile(ctx context.Context, data []byte, filename, sheet string) (ImportResultResponseDTO, error) {
	src, err := spreadsheet.ReadFile(data, filename, sheet)
	if err != nil {
		return ImportResultResponseDTO{}, httpx.AsBadRequest(err)
	}

	headerIdx := firstNonEmptyRow(src.Rows)
	if headerIdx == -1 {
		return ImportResultResponseDTO{}, httpx.AsBadRequest(errors.New("invalid spreadsheet: the file is empty"))
	}

	cols := mapSimpleHeaders(src.Rows[headerIdx], exchangeRateHeaderSynonyms)
	if missing := missingCols(cols, "fromcurrency", "tocurrency", "rate"); len(missing) > 0 {
		return ImportResultResponseDTO{}, httpx.AsBadRequest(fmt.Errorf("invalid spreadsheet: missing required columns: %s", strings.Join(missing, ", ")))
	}

	dataRows := src.Rows[headerIdx+1:]
	if len(dataRows) > maxExchangeRateImportRows {
		return ImportResultResponseDTO{}, httpx.AsTooManyRequests(fmt.Errorf("invalid spreadsheet: too many rows (max %d)", maxExchangeRateImportRows))
	}

	now := time.Now()
	result := ImportResultResponseDTO{Errors: []ImportResultErrorDTO{}}

	for i, row := range dataRows {
		if spreadsheet.RowIsEmpty(row) {
			continue
		}
		rowNumber := headerIdx + 2 + i
		result.TotalRows++

		from := strings.ToUpper(spreadsheet.CellAtIdx(row, cols["fromcurrency"]))
		to := strings.ToUpper(spreadsheet.CellAtIdx(row, cols["tocurrency"]))
		rateRaw := spreadsheet.CellAtIdx(row, cols["rate"])

		var rowErrs []string
		if len(from) != 3 {
			rowErrs = append(rowErrs, fmt.Sprintf("moneda origen inválida: %q", from))
		}
		if len(to) != 3 {
			rowErrs = append(rowErrs, fmt.Sprintf("moneda destino inválida: %q", to))
		}

		rate, err := parseDecimal(rateRaw)
		if err != nil {
			rowErrs = append(rowErrs, fmt.Sprintf("tasa no numérica: %q", rateRaw))
		} else if rate.IsZero() || rate.IsNeg() {
			rowErrs = append(rowErrs, fmt.Sprintf("la tasa debe ser mayor que 0: %q", rateRaw))
		}

		if len(rowErrs) > 0 {
			result.Skipped++
			if len(result.Errors) < 100 {
				result.Errors = append(result.Errors, ImportResultErrorDTO{Row: rowNumber, Message: strings.Join(rowErrs, "; ")})
			}
			continue
		}

		if _, err := s.repo.UpsertExchangeRate(ctx, from, to, rate, now); err != nil {
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
