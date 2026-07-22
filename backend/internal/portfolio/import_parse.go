package portfolio

// This file holds value-parsing helpers for the transaction importer:
// decimals, dates (numeric and Spanish textual) and enum normalisation.
// Split out of import.go to keep each file under ~500 lines.

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"
)

var currencyCodeRe = regexp.MustCompile(`[A-Z]{3}`)
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

var spanishMonths = map[string]int{
	"ene": 1, "enero": 1, "jan": 1, "january": 1,
	"feb": 2, "febrero": 2, "february": 2,
	"mar": 3, "marzo": 3, "march": 3,
	"abr": 4, "abril": 4, "apr": 4, "april": 4,
	"may": 5, "mayo": 5,
	"jun": 6, "junio": 6, "june": 6,
	"jul": 7, "julio": 7, "july": 7,
	"ago": 8, "agosto": 8, "aug": 8, "august": 8,
	"sep": 9, "sept": 9, "septiembre": 9, "september": 9,
	"oct": 10, "octubre": 10, "october": 10,
	"nov": 11, "noviembre": 11, "november": 11,
	"dic": 12, "diciembre": 12, "dec": 12, "december": 12,
}

var textualDateRe = regexp.MustCompile(`^(\d{1,2})[\s\-/.]*(?:de\s+)?([a-z]+)[\s\-/.,]*(?:de\s+)?(\d{2,4})$`)
var numericDateRe = regexp.MustCompile(`^(\d{1,4})[\-/.](\d{1,2})[\-/.](\d{1,4})$`)

// parseImportDate accepts ISO dates, day/month/year and month/day/year (per
// dateOrder), Excel serial numbers and "15 ene 2024"-style textual dates.
func parseImportDate(raw, dateOrder string) (time.Time, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return time.Time{}, errors.New("empty date")
	}
	// Drop a time component ("2024-01-15 00:00:00", "15/01/2024T10:00").
	if i := strings.IndexAny(s, " T"); i > 0 && strings.ContainsAny(s[:i], "-/.") {
		s = s[:i]
	}

	// Excel serial date (formatted cells arrive as text, raw ones as serial).
	if serial, err := strconv.ParseFloat(s, 64); err == nil {
		if serial < 1 || serial > 200000 {
			return time.Time{}, errors.New("unrecognized date")
		}
		return excelize.ExcelDateToTime(serial, false)
	}

	if m := textualDateRe.FindStringSubmatch(normKey(s)); m != nil {
		month, ok := spanishMonths[m[2]]
		if !ok {
			return time.Time{}, errors.New("unrecognized date")
		}
		day, _ := strconv.Atoi(m[1])
		year := expandYear(m[3])
		return buildDate(year, month, day)
	}

	m := numericDateRe.FindStringSubmatch(s)
	if m == nil {
		return time.Time{}, errors.New("unrecognized date")
	}
	p0, _ := strconv.Atoi(m[1])
	p1, _ := strconv.Atoi(m[2])
	p2, _ := strconv.Atoi(m[3])

	if len(m[1]) == 4 { // ISO: yyyy-mm-dd
		return buildDate(p0, p1, p2)
	}

	year := expandYear(m[3])
	day, month := p0, p1
	if dateOrder == "mdy" {
		day, month = p1, p0
	}
	// Self-correct obvious mismatches regardless of the preferred order.
	if month > 12 && day <= 12 {
		day, month = month, day
	}
	return buildDate(year, month, day)
}

func expandYear(s string) int {
	y, _ := strconv.Atoi(s)
	if len(s) <= 2 {
		y += 2000
	}
	return y
}

func buildDate(year, month, day int) (time.Time, error) {
	if year < 1900 || year > 2200 || month < 1 || month > 12 || day < 1 || day > 31 {
		return time.Time{}, errors.New("unrecognized date")
	}
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	if t.Day() != day || int(t.Month()) != month {
		return time.Time{}, errors.New("unrecognized date")
	}
	return t, nil
}

// inferDateOrder inspects every date cell: any first component above 12 proves
// day-first, any second component above 12 proves month-first. Defaults to
// day-first, the common convention for our Spanish-speaking users.
func inferDateOrder(values []string) string {
	for _, v := range values {
		m := numericDateRe.FindStringSubmatch(strings.TrimSpace(v))
		if m == nil || len(m[1]) == 4 {
			continue
		}
		p0, _ := strconv.Atoi(m[1])
		p1, _ := strconv.Atoi(m[2])
		if p0 > 12 && p1 <= 12 {
			return "dmy"
		}
		if p1 > 12 && p0 <= 12 {
			return "mdy"
		}
	}
	return "dmy"
}

func normalizeTxnType(raw string) (TransactionType, bool) {
	if t, ok := txnTypeSynonyms[normKey(raw)]; ok {
		return t, true
	}
	return "", false
}

func normalizeCategory(raw string) (AssetType, bool) {
	if c, ok := categorySynonyms[normKey(raw)]; ok {
		return c, true
	}
	return "", false
}

func normalizeCurrency(raw string) (string, bool) {
	s := strings.ToUpper(strings.TrimSpace(raw))
	if s == "" {
		return "", false
	}
	if code, ok := currencySymbols[s]; ok {
		return code, true
	}
	if code := currencyCodeRe.FindString(s); code != "" {
		return code, true
	}
	return "", false
}

// --- row assembly -----------------------------------------------------------
