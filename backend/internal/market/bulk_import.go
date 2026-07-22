package market

import (
	"errors"
	"regexp"
	"slices"
	"strings"

	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/platform/spreadsheet"
)

// This file keeps the fixed-schema import helpers the asset and exchange-rate
// bulk importers share. Generic spreadsheet reading (file → rows, cell access,
// header normalisation) lives in platform/spreadsheet.

var decimalCleanRe = regexp.MustCompile(`[^0-9,.\-()]`)

// parseDecimal converts human spreadsheet numbers ("1.234,56", "$ 1,234.56",
// "(120.50)") into a money.Decimal.
func parseDecimal(raw string) (money.Decimal, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return decimal.Zero, errors.New("empty value")
	}

	s = decimalCleanRe.ReplaceAllString(s, "")
	negative := strings.HasPrefix(s, "(") && strings.HasSuffix(s, ")")
	s = strings.Trim(s, "()")
	if strings.HasPrefix(s, "-") {
		negative = !negative
		s = strings.TrimPrefix(s, "-")
	}
	s = strings.ReplaceAll(s, "-", "")
	if s == "" {
		return decimal.Zero, errors.New("not a number")
	}

	lastComma := strings.LastIndex(s, ",")
	lastDot := strings.LastIndex(s, ".")
	switch {
	case lastComma >= 0 && lastDot >= 0:
		// Both present: the right-most separator is the decimal one.
		if lastComma > lastDot {
			s = strings.ReplaceAll(s, ".", "")
			s = strings.Replace(s, ",", ".", 1)
			s = strings.ReplaceAll(s, ",", "") // stray extra commas
		} else {
			s = strings.ReplaceAll(s, ",", "")
		}
	case lastComma >= 0:
		commas := strings.Count(s, ",")
		digitsAfter := len(s) - lastComma - 1
		if commas == 1 && digitsAfter != 3 {
			s = strings.Replace(s, ",", ".", 1)
		} else {
			// "1,234" / "1,234,567": thousands separators.
			s = strings.ReplaceAll(s, ",", "")
		}
	case strings.Count(s, ".") > 1:
		// "1.234.567": Spanish-locale thousands separators.
		s = strings.ReplaceAll(s, ".", "")
	}

	if negative {
		s = "-" + s
	}
	d, err := decimal.NewFromString(s)
	if err != nil {
		return decimal.Zero, errors.New("not a number")
	}
	return d, nil
}

// mapSimpleHeaders matches spreadsheet headers against a fixed set of field
// synonyms (case/accent/punctuation-insensitive via spreadsheet.NormKey) and
// returns the 0-based column index for each recognised field. Used by the
// fixed-schema importers (assets, exchange rates).
func mapSimpleHeaders(headers []string, synonyms map[string][]string) map[string]int {
	idx := make(map[string]int, len(synonyms))
	for i, h := range headers {
		key := spreadsheet.NormKey(h)
		if key == "" {
			continue
		}
		for field, syns := range synonyms {
			if _, assigned := idx[field]; assigned {
				continue
			}

			if slices.Contains(syns, key) {
				idx[field] = i
				break
			}
		}
	}
	return idx
}

// firstNonEmptyRow returns the index of the first row with at least one
// non-blank cell, or -1 if every row is empty.
func firstNonEmptyRow(rows [][]string) int {
	for i, row := range rows {
		if !spreadsheet.RowIsEmpty(row) {
			return i
		}
	}
	return -1
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
