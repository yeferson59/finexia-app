package services

import "strings"

// mapSimpleHeaders matches spreadsheet headers against a fixed set of field
// synonyms (case/accent/punctuation-insensitive via normKey) and returns the
// 0-based column index for each recognised field. Used by the assets and
// exchange-rate bulk importers, which — unlike the freeform transactions
// importer — work off a small, fixed schema instead of a user-driven mapping.
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
