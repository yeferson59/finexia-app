// Package spreadsheet holds the domain-agnostic reading of uploaded CSV/XLSX
// files shared by the importers (portfolio transactions, market assets). It
// knows nothing about the business fields — it only turns bytes into rows and
// offers small helpers to read and normalise cells.
package spreadsheet

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/xuri/excelize/v2"
)

// Source is a parsed spreadsheet: the rows of the selected sheet plus the list
// of available sheets.
type Source struct {
	Sheets []string
	Sheet  string
	Rows   [][]string
}

// ReadFile parses an uploaded file (.csv or .xlsx). For workbooks it selects
// the requested sheet when present, otherwise the first one.
func ReadFile(data []byte, filename, sheet string) (Source, error) {
	if strings.HasSuffix(strings.ToLower(filename), ".csv") {
		rows, err := parseCSV(data)
		if err != nil {
			return Source{}, err
		}
		return Source{Sheets: []string{"CSV"}, Sheet: "CSV", Rows: rows}, nil
	}

	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return Source{}, errors.New("invalid spreadsheet: could not open the file, expected .xlsx or .csv")
	}
	defer func() { _ = f.Close() }()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return Source{}, errors.New("invalid spreadsheet: file has no sheets")
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
		return Source{}, fmt.Errorf("invalid spreadsheet: could not read sheet %q", selected)
	}
	return Source{Sheets: sheets, Sheet: selected, Rows: rows}, nil
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

var accentReplacer = strings.NewReplacer(
	"á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ü", "u", "ñ", "n",
	"Á", "a", "É", "e", "Í", "i", "Ó", "o", "Ú", "u", "Ü", "u", "Ñ", "n",
)

var nonWordSeparators = regexp.MustCompile(`[_\-./:()#]+`)
var multiSpace = regexp.MustCompile(`\s+`)

// NormKey lower-cases, strips accents and collapses punctuation/whitespace so a
// spreadsheet header can be matched against a synonym list.
func NormKey(s string) string {
	s = accentReplacer.Replace(strings.ToLower(strings.TrimSpace(s)))
	s = nonWordSeparators.ReplaceAllString(s, " ")
	s = multiSpace.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

// RowIsEmpty reports whether every cell in a row is blank.
func RowIsEmpty(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}

// CellAt reads a cell by an optional column index (nil = unmapped), returning
// "" when the index is nil or out of range.
func CellAt(row []string, idx *int) string {
	if idx == nil || *idx < 0 || *idx >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[*idx])
}

// CellAtIdx reads a cell by a plain column index, returning "" when out of range.
func CellAtIdx(row []string, idx int) string {
	if idx < 0 || idx >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[idx])
}
