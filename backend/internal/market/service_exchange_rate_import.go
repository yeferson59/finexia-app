package market

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/yeferson59/finexia-app/internal/portfolio"
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
func (s *Service) ImportExchangeRatesFromFile(ctx context.Context, data []byte, filename, sheet string) (portfolio.ImportResultResponseDTO, error) {
	src, err := parseImportFile(data, filename, sheet)
	if err != nil {
		return portfolio.ImportResultResponseDTO{}, err
	}

	headerIdx := firstNonEmptyRow(src.rows)
	if headerIdx == -1 {
		return portfolio.ImportResultResponseDTO{}, errors.New("invalid spreadsheet: the file is empty")
	}

	cols := mapSimpleHeaders(src.rows[headerIdx], exchangeRateHeaderSynonyms)
	if missing := missingCols(cols, "fromcurrency", "tocurrency", "rate"); len(missing) > 0 {
		return portfolio.ImportResultResponseDTO{}, fmt.Errorf("invalid spreadsheet: missing required columns: %s", strings.Join(missing, ", "))
	}

	dataRows := src.rows[headerIdx+1:]
	if len(dataRows) > maxExchangeRateImportRows {
		return portfolio.ImportResultResponseDTO{}, fmt.Errorf("invalid spreadsheet: too many rows (max %d)", maxExchangeRateImportRows)
	}

	now := time.Now()
	result := portfolio.ImportResultResponseDTO{Errors: []portfolio.ImportResultErrorDTO{}}

	for i, row := range dataRows {
		if rowIsEmpty(row) {
			continue
		}
		rowNumber := headerIdx + 2 + i
		result.TotalRows++

		from := strings.ToUpper(cellAtIdx(row, cols["fromcurrency"]))
		to := strings.ToUpper(cellAtIdx(row, cols["tocurrency"]))
		rateRaw := cellAtIdx(row, cols["rate"])

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
				result.Errors = append(result.Errors, portfolio.ImportResultErrorDTO{Row: rowNumber, Message: strings.Join(rowErrs, "; ")})
			}
			continue
		}

		if _, err := s.repo.UpsertExchangeRate(ctx, from, to, rate, now); err != nil {
			result.Skipped++
			if len(result.Errors) < 100 {
				result.Errors = append(result.Errors, portfolio.ImportResultErrorDTO{Row: rowNumber, Message: err.Error()})
			}
			continue
		}
		result.Imported++
	}

	return result, nil
}
