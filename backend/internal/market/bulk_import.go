package market

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"regexp"
	"slices"
	"strings"

	"github.com/xuri/excelize/v2"
	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"
)

// This file keeps the minimal spreadsheet-parsing toolbox the exchange-rate
// importer still needs. The canonical copies migrated to internal/portfolio
// with the transactions/assets importers (Fase 6); this duplicate dies with
// the market module in Fase 7 (see docs/TECH_DEBT.md).

var accentReplacer = strings.NewReplacer(
	"á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ü", "u", "ñ", "n",
	"Á", "a", "É", "e", "Í", "i", "Ó", "o", "Ú", "u", "Ü", "u", "Ñ", "n",
)

var nonWordSeparators = regexp.MustCompile(`[_\-./:()#]+`)
var multiSpace = regexp.MustCompile(`\s+`)

// normKey canonicalises a header or enum-like cell for synonym lookup.
func normKey(s string) string {
	s = accentReplacer.Replace(strings.ToLower(strings.TrimSpace(s)))
	s = nonWordSeparators.ReplaceAllString(s, " ")
	s = multiSpace.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

type importSource struct {
	sheets []string
	sheet  string
	rows   [][]string
}

func parseImportFile(data []byte, filename, sheet string) (importSource, error) {
	if strings.HasSuffix(strings.ToLower(filename), ".csv") {
		rows, err := parseCSV(data)
		if err != nil {
			return importSource{}, err
		}
		return importSource{sheets: []string{"CSV"}, sheet: "CSV", rows: rows}, nil
	}

	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return importSource{}, errors.New("invalid spreadsheet: could not open the file, expected .xlsx or .csv")
	}
	defer func() { _ = f.Close() }()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return importSource{}, errors.New("invalid spreadsheet: file has no sheets")
	}
	selected := sheets[0]
	for _, s := range sheets {
		if s == sheet {
			selected = s
			break
		}
	}

	rows, err := f.GetRows(selected)
	if err != nil {
		return importSource{}, fmt.Errorf("invalid spreadsheet: could not read sheet %q", selected)
	}
	return importSource{sheets: sheets, sheet: selected, rows: rows}, nil
}

func parseCSV(data []byte) ([][]string, error) {
	data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF}) // UTF-8 BOM

	// Detect the delimiter from the first line: "classic" exports use ',',
	// ';' (Excel with Spanish locale) or tabs.
	firstLine := data
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		firstLine = data[:i]
	}
	delimiter := ','
	best := bytes.Count(firstLine, []byte{','})
	for _, cand := range []byte{';', '\t'} {
		if n := bytes.Count(firstLine, []byte{cand}); n > best {
			best, delimiter = n, rune(cand)
		}
	}

	r := csv.NewReader(bytes.NewReader(data))
	r.Comma = delimiter
	r.FieldsPerRecord = -1
	r.LazyQuotes = true

	var rows [][]string
	for {
		record, err := r.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, errors.New("invalid spreadsheet: malformed CSV file")
		}
		rows = append(rows, record)
	}
	return rows, nil
}

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

func rowIsEmpty(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}

// mapSimpleHeaders matches spreadsheet headers against a fixed set of field
// synonyms (case/accent/punctuation-insensitive via normKey) and returns the
// 0-based column index for each recognised field. Used by the exchange-rate
// bulk importer, which — unlike the freeform transactions importer — works
// off a small, fixed schema instead of a user-driven mapping.
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
