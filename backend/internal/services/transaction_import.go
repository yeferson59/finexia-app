package services

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"

	portfoliodto "github.com/yeferson59/finexia-app/internal/dtos/portfolio"
	"github.com/yeferson59/finexia-app/internal/entities"
)

const (
	// maxImportRows bounds how many data rows a single upload may contain.
	maxImportRows = 5000
	// importPreviewRowCap bounds how many parsed rows the preview response
	// carries back to the client; totals always cover the whole file.
	importPreviewRowCap = 200

	maxTickerLen    = 20 // assets.ticker VARCHAR(20)
	maxAssetNameLen = 255
	maxNotesLen     = 500
)

// importField identifies a mappable transaction attribute.
type importField string

const (
	fieldDate      importField = "date"
	fieldType      importField = "type"
	fieldTicker    importField = "ticker"
	fieldAssetName importField = "assetName"
	fieldQuantity  importField = "quantity"
	fieldPrice     importField = "price"
	fieldFees      importField = "fees"
	fieldCurrency  importField = "currency"
	fieldCategory  importField = "category"
	fieldNotes     importField = "notes"
)

// requiredImportFields must be mapped before rows can be validated/imported.
var requiredImportFields = []importField{fieldDate, fieldTicker, fieldQuantity, fieldPrice}

// fieldSynonyms drives the automatic column detection. Every user keeps their
// own spreadsheet layout, so headers are matched against Spanish and English
// synonyms after stripping accents, case and punctuation.
var fieldSynonyms = map[importField][]string{
	fieldDate:      {"fecha", "date", "fecha operacion", "fecha de operacion", "fecha de compra", "fecha transaccion", "trade date", "dia", "day"},
	fieldType:      {"tipo", "type", "operacion", "tipo operacion", "tipo de operacion", "movimiento", "transaccion", "transaction type", "side", "orden"},
	fieldTicker:    {"ticker", "simbolo", "symbol", "codigo", "code", "activo codigo", "sigla"},
	fieldAssetName: {"activo", "asset", "nombre", "name", "empresa", "company", "descripcion", "description", "producto", "instrumento", "titulo"},
	fieldQuantity:  {"cantidad", "quantity", "qty", "unidades", "shares", "acciones", "titulos", "nominales", "cant", "numero de acciones", "units"},
	fieldPrice:     {"precio", "price", "precio unitario", "precio compra", "precio de compra", "cotizacion", "unit price", "valor unitario", "px"},
	fieldFees:      {"comision", "comisiones", "fee", "fees", "cargo", "cargos", "gastos", "costos", "commission", "costo transaccion"},
	fieldCurrency:  {"moneda", "currency", "divisa", "ccy"},
	fieldCategory:  {"categoria", "category", "tipo de activo", "tipo activo", "asset type", "clase", "clase de activo", "asset class"},
	fieldNotes:     {"notas", "nota", "notes", "comentario", "comentarios", "observaciones", "observacion", "memo", "detalle"},
}

var txnTypeSynonyms = map[string]entities.TransactionType{
	"buy": entities.Buy, "compra": entities.Buy, "compras": entities.Buy, "purchase": entities.Buy,
	"bought": entities.Buy, "adquisicion": entities.Buy, "cpra": entities.Buy,
	"sell": entities.Sell, "venta": entities.Sell, "ventas": entities.Sell, "sale": entities.Sell,
	"sold": entities.Sell, "vender": entities.Sell, "vta": entities.Sell,
	"dividend": entities.Dividend, "dividendo": entities.Dividend, "dividendos": entities.Dividend, "div": entities.Dividend,
	"interest": entities.Interest, "interes": entities.Interest, "intereses": entities.Interest, "rendimiento": entities.Interest, "rendimientos": entities.Interest,
	"fee": entities.Fee, "fees": entities.Fee, "comision": entities.Fee, "comisiones": entities.Fee, "cargo": entities.Fee,
	"split": entities.Split, "division": entities.Split,
	"transfer in": entities.TransferIn, "transfer_in": entities.TransferIn, "transferencia entrada": entities.TransferIn,
	"deposito": entities.TransferIn, "deposit": entities.TransferIn, "entrada": entities.TransferIn, "ingreso": entities.TransferIn, "aporte": entities.TransferIn,
	"transfer out": entities.TransferOut, "transfer_out": entities.TransferOut, "transferencia salida": entities.TransferOut,
	"retiro": entities.TransferOut, "withdrawal": entities.TransferOut, "salida": entities.TransferOut, "egreso": entities.TransferOut,
}

var categorySynonyms = map[string]entities.AssetType{
	"stock": entities.Stock, "stocks": entities.Stock, "accion": entities.Stock, "acciones": entities.Stock,
	"equity": entities.Stock, "equities": entities.Stock,
	"etf": entities.ETF, "etfs": entities.ETF, "fondo": entities.ETF, "fondos": entities.ETF,
	"fund": entities.ETF, "funds": entities.ETF, "fondo indexado": entities.ETF, "index fund": entities.ETF,
	"crypto": entities.Crypto, "cripto": entities.Crypto, "criptomoneda": entities.Crypto,
	"criptomonedas": entities.Crypto, "cryptocurrency": entities.Crypto, "criptos": entities.Crypto,
	"bond": entities.Bond, "bonds": entities.Bond, "bono": entities.Bond, "bonos": entities.Bond,
	"renta fija": entities.Bond, "cdt": entities.Bond, "fixed income": entities.Bond,
	"cash": entities.Cash, "efectivo": entities.Cash, "liquidez": entities.Cash, "dinero": entities.Cash,
	"real estate": entities.RealEstate, "real_estate": entities.RealEstate, "inmueble": entities.RealEstate,
	"inmuebles": entities.RealEstate, "bienes raices": entities.RealEstate, "reit": entities.RealEstate, "fibra": entities.RealEstate,
	"commodity": entities.Commodity, "commodities": entities.Commodity, "materia prima": entities.Commodity,
	"materias primas": entities.Commodity, "oro": entities.Commodity, "gold": entities.Commodity, "plata": entities.Commodity,
	"other": entities.Other, "otro": entities.Other, "otros": entities.Other,
}

var currencySymbols = map[string]string{
	"$": "USD", "US$": "USD", "€": "EUR", "£": "GBP", "¥": "JPY",
}

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

// --- file parsing ---------------------------------------------------------

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

	// Detect the delimiter from the first line: exports "clásicos" use ',',
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

// --- header detection & mapping suggestion --------------------------------

// matchField scores how well a header matches a field's synonyms.
func matchField(header string, field importField) int {
	h := normKey(header)
	if h == "" {
		return 0
	}
	score := 0
	for _, syn := range fieldSynonyms[field] {
		switch {
		case h == syn:
			return 3
		case len(syn) >= 4 && strings.Contains(h, syn):
			if score < 2 {
				score = 2
			}
		case len(h) >= 4 && strings.Contains(syn, h):
			if score < 1 {
				score = 1
			}
		}
	}
	return score
}

// suggestMapping assigns each field its best-matching column, greedily by
// score so an exact "precio" beats a partial "precio total".
func suggestMapping(headers []string) portfoliodto.ImportMappingDTO {
	type cand struct {
		field importField
		col   int
		score int
	}
	var cands []cand
	for _, field := range []importField{
		fieldDate, fieldType, fieldTicker, fieldAssetName, fieldQuantity,
		fieldPrice, fieldFees, fieldCurrency, fieldCategory, fieldNotes,
	} {
		for col, header := range headers {
			if score := matchField(header, field); score > 0 {
				cands = append(cands, cand{field, col, score})
			}
		}
	}
	// Stable selection sort: highest score first; ties keep field priority
	// order (slice order) so required fields win contested columns.
	assignedField := map[importField]bool{}
	assignedCol := map[int]bool{}
	result := map[importField]int{}
	for range cands {
		bestIdx := -1
		for i, c := range cands {
			if assignedField[c.field] || assignedCol[c.col] {
				continue
			}
			if bestIdx == -1 || c.score > cands[bestIdx].score {
				bestIdx = i
			}
		}
		if bestIdx == -1 {
			break
		}
		chosen := cands[bestIdx]
		assignedField[chosen.field] = true
		assignedCol[chosen.col] = true
		result[chosen.field] = chosen.col
	}

	toPtr := func(f importField) *int {
		if col, ok := result[f]; ok {
			c := col
			return &c
		}
		return nil
	}
	return portfoliodto.ImportMappingDTO{
		Date:      toPtr(fieldDate),
		Type:      toPtr(fieldType),
		Ticker:    toPtr(fieldTicker),
		AssetName: toPtr(fieldAssetName),
		Quantity:  toPtr(fieldQuantity),
		Price:     toPtr(fieldPrice),
		Fees:      toPtr(fieldFees),
		Currency:  toPtr(fieldCurrency),
		Category:  toPtr(fieldCategory),
		Notes:     toPtr(fieldNotes),
	}
}

// detectHeaderRow scans the first rows of the sheet and picks the one whose
// cells match the most field synonyms, so title rows above the real table
// ("Mis inversiones 2024") are skipped. Falls back to the first non-empty row.
func detectHeaderRow(rows [][]string) int {
	firstNonEmpty := -1
	bestRow, bestScore := -1, 0
	limit := min(len(rows), 10)
	for i := 0; i < limit; i++ {
		if rowIsEmpty(rows[i]) {
			continue
		}
		if firstNonEmpty == -1 {
			firstNonEmpty = i
		}
		score := 0
		for _, field := range []importField{
			fieldDate, fieldType, fieldTicker, fieldAssetName, fieldQuantity,
			fieldPrice, fieldFees, fieldCurrency, fieldCategory, fieldNotes,
		} {
			for _, header := range rows[i] {
				if matchField(header, field) >= 2 {
					score++
					break
				}
			}
		}
		if score > bestScore {
			bestRow, bestScore = i, score
		}
	}
	if bestScore >= 2 {
		return bestRow
	}
	return firstNonEmpty
}

func rowIsEmpty(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}

func cellAt(row []string, idx *int) string {
	if idx == nil || *idx < 0 || *idx >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[*idx])
}

// --- value normalisation ---------------------------------------------------

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

func normalizeTxnType(raw string) (entities.TransactionType, bool) {
	if t, ok := txnTypeSynonyms[normKey(raw)]; ok {
		return t, true
	}
	return "", false
}

func normalizeCategory(raw string) (entities.AssetType, bool) {
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

type importOutcome struct {
	preview portfoliodto.ImportPreviewResponseDTO
	valid   []entities.ImportTransactionRow
}

// quantityRequired reports whether a transaction type moves units and thus
// needs a positive quantity; cash-flow rows (dividends, fees, interest) may
// leave it empty.
func quantityRequired(t entities.TransactionType) bool {
	switch t {
	case entities.Buy, entities.Sell, entities.TransferIn, entities.TransferOut, entities.Split:
		return true
	default:
		return false
	}
}

func applyImportDefaults(defaults portfoliodto.ImportDefaultsDTO) (portfoliodto.ImportDefaultsDTO, error) {
	out := defaults
	if strings.TrimSpace(out.Type) == "" {
		out.Type = string(entities.Buy)
	}
	if _, ok := normalizeTxnType(out.Type); !ok {
		return out, fmt.Errorf("invalid default transaction type %q", out.Type)
	}
	if strings.TrimSpace(out.Currency) == "" {
		out.Currency = "USD"
	}
	cur, ok := normalizeCurrency(out.Currency)
	if !ok {
		return out, fmt.Errorf("invalid default currency %q", out.Currency)
	}
	out.Currency = cur
	if strings.TrimSpace(out.Category) == "" {
		out.Category = string(entities.Stock)
	}
	if _, ok := normalizeCategory(out.Category); !ok {
		return out, fmt.Errorf("invalid default category %q", out.Category)
	}
	switch out.DateFormat {
	case "", "auto":
		out.DateFormat = "auto"
	case "dmy", "mdy":
	default:
		return out, fmt.Errorf("invalid date format %q", out.DateFormat)
	}
	return out, nil
}

func missingRequiredFields(m *portfoliodto.ImportMappingDTO) []string {
	var missing []string
	for _, f := range requiredImportFields {
		var ptr *int
		switch f {
		case fieldDate:
			ptr = m.Date
		case fieldTicker:
			ptr = m.Ticker
		case fieldQuantity:
			ptr = m.Quantity
		case fieldPrice:
			ptr = m.Price
		}
		if ptr == nil {
			missing = append(missing, string(f))
		}
	}
	return missing
}

// buildImport parses the sheet with the given (or suggested) mapping and
// returns the full per-row preview plus the validated rows ready to persist.
func buildImport(src importSource, mapping *portfoliodto.ImportMappingDTO, defaults portfoliodto.ImportDefaultsDTO) (importOutcome, error) {
	defaults, err := applyImportDefaults(defaults)
	if err != nil {
		return importOutcome{}, err
	}

	headerIdx := detectHeaderRow(src.rows)
	if headerIdx == -1 {
		return importOutcome{}, errors.New("invalid spreadsheet: the file is empty")
	}
	headers := make([]string, len(src.rows[headerIdx]))
	for i, h := range src.rows[headerIdx] {
		headers[i] = strings.TrimSpace(h)
	}

	suggested := suggestMapping(headers)
	if mapping == nil {
		mapping = &suggested
	}

	dataRows := src.rows[headerIdx+1:]
	if len(dataRows) > maxImportRows {
		return importOutcome{}, fmt.Errorf("invalid spreadsheet: too many rows (max %d)", maxImportRows)
	}

	preview := portfoliodto.ImportPreviewResponseDTO{
		Sheets:           src.sheets,
		Sheet:            src.sheet,
		HeaderRow:        headerIdx + 1,
		Headers:          headers,
		SuggestedMapping: suggested,
		MissingFields:    missingRequiredFields(mapping),
		Rows:             []portfoliodto.ImportRowDTO{},
	}

	dateOrder := defaults.DateFormat
	if dateOrder == "auto" {
		var dates []string
		for _, row := range dataRows {
			if v := cellAt(row, mapping.Date); v != "" {
				dates = append(dates, v)
			}
		}
		dateOrder = inferDateOrder(dates)
	}

	var valid []entities.ImportTransactionRow
	for i, row := range dataRows {
		if rowIsEmpty(row) {
			continue
		}
		rowNumber := headerIdx + 2 + i // 1-based sheet row number
		dto := portfoliodto.ImportRowDTO{RowNumber: rowNumber, Raw: row, Errors: []string{}}
		preview.TotalRows++

		if len(preview.MissingFields) > 0 {
			preview.InvalidRows++
			preview.Rows = append(preview.Rows, dto)
			continue
		}

		entity, rowErrs := buildImportRow(row, rowNumber, mapping, defaults, dateOrder, &dto)
		if len(rowErrs) > 0 {
			dto.Errors = rowErrs
			preview.InvalidRows++
		} else {
			dto.Valid = true
			preview.ValidRows++
			valid = append(valid, entity)
		}
		preview.Rows = append(preview.Rows, dto)
	}

	return importOutcome{preview: preview, valid: valid}, nil
}

// buildImportRow normalises and validates a single data row. It fills the
// preview DTO with the normalised values and returns the persistable entity
// when the row has no errors.
func buildImportRow(
	row []string,
	rowNumber int,
	mapping *portfoliodto.ImportMappingDTO,
	defaults portfoliodto.ImportDefaultsDTO,
	dateOrder string,
	dto *portfoliodto.ImportRowDTO,
) (entities.ImportTransactionRow, []string) {
	var errs []string
	entity := entities.ImportTransactionRow{RowNumber: rowNumber}

	// Transaction type: mapped column wins, empty cells fall back to default.
	typeRaw := cellAt(row, mapping.Type)
	txnType, _ := normalizeTxnType(defaults.Type)
	if typeRaw != "" {
		if t, ok := normalizeTxnType(typeRaw); ok {
			txnType = t
		} else {
			errs = append(errs, fmt.Sprintf("tipo de operación no reconocido: %q", typeRaw))
		}
	}
	entity.Type = txnType
	dto.Type = string(txnType)

	// Date.
	if dateRaw := cellAt(row, mapping.Date); dateRaw == "" {
		errs = append(errs, "la fecha está vacía")
	} else if date, err := parseImportDate(dateRaw, dateOrder); err != nil {
		errs = append(errs, fmt.Sprintf("fecha no reconocida: %q", dateRaw))
	} else {
		entity.Date = date
		dto.Date = date.Format("2006-01-02")
	}

	// Ticker.
	ticker := strings.ToUpper(cellAt(row, mapping.Ticker))
	switch {
	case ticker == "":
		errs = append(errs, "el ticker/símbolo está vacío")
	case len(ticker) > maxTickerLen:
		errs = append(errs, fmt.Sprintf("el ticker supera %d caracteres: %q", maxTickerLen, ticker))
	default:
		entity.Ticker = ticker
		dto.Ticker = ticker
	}

	// Asset name (falls back to the ticker).
	name := cellAt(row, mapping.AssetName)
	if name == "" {
		name = ticker
	}
	if len(name) > maxAssetNameLen {
		name = name[:maxAssetNameLen]
	}
	entity.AssetName = name
	dto.AssetName = name

	// Quantity.
	qtyRaw := cellAt(row, mapping.Quantity)
	if qtyRaw == "" {
		if quantityRequired(txnType) {
			errs = append(errs, "la cantidad está vacía")
		} else {
			entity.Quantity = decimal.Zero
			dto.Quantity = "0"
		}
	} else if qty, err := parseDecimal(qtyRaw); err != nil {
		errs = append(errs, fmt.Sprintf("cantidad no numérica: %q", qtyRaw))
	} else if qty.InexactFloat64() < 0 || (quantityRequired(txnType) && qty.IsZero()) {
		errs = append(errs, fmt.Sprintf("la cantidad debe ser mayor que 0: %q", qtyRaw))
	} else {
		entity.Quantity = qty
		dto.Quantity = qty.String()
	}

	// Price.
	priceRaw := cellAt(row, mapping.Price)
	if priceRaw == "" {
		if txnType == entities.Buy || txnType == entities.Sell {
			errs = append(errs, "el precio está vacío")
		} else {
			entity.Price = money.FromDecimal(decimal.Zero, money.USD)
			dto.Price = "0"
		}
	} else if price, err := parseDecimal(priceRaw); err != nil {
		errs = append(errs, fmt.Sprintf("precio no numérico: %q", priceRaw))
	} else if price.InexactFloat64() < 0 {
		errs = append(errs, fmt.Sprintf("el precio no puede ser negativo: %q", priceRaw))
	} else {
		entity.Price = money.FromDecimal(price, money.USD)
		dto.Price = price.String()
	}

	// Fees (optional).
	feesRaw := cellAt(row, mapping.Fees)
	entity.Fees = money.FromDecimal(decimal.Zero, money.USD)
	dto.Fees = "0"
	if feesRaw != "" {
		if fees, err := parseDecimal(feesRaw); err != nil {
			errs = append(errs, fmt.Sprintf("comisión no numérica: %q", feesRaw))
		} else if fees.InexactFloat64() < 0 {
			errs = append(errs, fmt.Sprintf("la comisión no puede ser negativa: %q", feesRaw))
		} else {
			entity.Fees = money.FromDecimal(fees, money.USD)
			dto.Fees = fees.String()
		}
	}

	// Currency.
	currency := defaults.Currency
	if curRaw := cellAt(row, mapping.Currency); curRaw != "" {
		if cur, ok := normalizeCurrency(curRaw); ok {
			currency = cur
		} else {
			errs = append(errs, fmt.Sprintf("moneda no reconocida: %q", curRaw))
		}
	}
	entity.Currency = currency
	dto.Currency = currency

	// Category / asset type.
	assetType, _ := normalizeCategory(defaults.Category)
	if catRaw := cellAt(row, mapping.Category); catRaw != "" {
		if cat, ok := normalizeCategory(catRaw); ok {
			assetType = cat
		} else {
			assetType = entities.Other
		}
	}
	entity.AssetType = assetType
	entity.Category = assetType.Transform()
	dto.Category = string(assetType)

	// Notes.
	notes := cellAt(row, mapping.Notes)
	if len(notes) > maxNotesLen {
		notes = notes[:maxNotesLen]
	}
	entity.Notes = notes
	dto.Notes = notes

	return entity, errs
}

// --- public service API -----------------------------------------------------

// PreviewTransactionImport parses an uploaded spreadsheet and returns its
// headers, a suggested column mapping and every row normalised + validated
// with the active mapping, without touching the database.
func (s *Services) PreviewTransactionImport(
	data []byte,
	filename, sheet string,
	mapping *portfoliodto.ImportMappingDTO,
	defaults portfoliodto.ImportDefaultsDTO,
) (portfoliodto.ImportPreviewResponseDTO, error) {
	src, err := parseImportFile(data, filename, sheet)
	if err != nil {
		return portfoliodto.ImportPreviewResponseDTO{}, err
	}
	out, err := buildImport(src, mapping, defaults)
	if err != nil {
		return portfoliodto.ImportPreviewResponseDTO{}, err
	}
	if len(out.preview.Rows) > importPreviewRowCap {
		out.preview.Rows = out.preview.Rows[:importPreviewRowCap]
	}
	return out.preview, nil
}

// ImportTransactionsFromFile re-parses the uploaded file with the confirmed
// mapping and persists every valid row (asset + position + transaction) in a
// single database transaction. Invalid rows are skipped and reported back.
func (s *Services) ImportTransactionsFromFile(
	ctx context.Context,
	userID, portfolioID, sourceID uuid.UUID,
	data []byte,
	filename, sheet string,
	mapping portfoliodto.ImportMappingDTO,
	defaults portfoliodto.ImportDefaultsDTO,
) (portfoliodto.ImportResultResponseDTO, error) {
	src, err := parseImportFile(data, filename, sheet)
	if err != nil {
		return portfoliodto.ImportResultResponseDTO{}, err
	}
	out, err := buildImport(src, &mapping, defaults)
	if err != nil {
		return portfoliodto.ImportResultResponseDTO{}, err
	}
	if len(out.preview.MissingFields) > 0 {
		return portfoliodto.ImportResultResponseDTO{},
			fmt.Errorf("invalid mapping: missing required columns: %s", strings.Join(out.preview.MissingFields, ", "))
	}

	result := portfoliodto.ImportResultResponseDTO{
		TotalRows: out.preview.TotalRows,
		Skipped:   out.preview.InvalidRows,
		Errors:    []portfoliodto.ImportResultErrorDTO{},
	}
	for _, row := range out.preview.Rows {
		if !row.Valid && len(result.Errors) < 100 {
			result.Errors = append(result.Errors, portfoliodto.ImportResultErrorDTO{
				Row:     row.RowNumber,
				Message: strings.Join(row.Errors, "; "),
			})
		}
	}

	if len(out.valid) == 0 {
		return result, nil
	}

	imported, err := s.repos.ImportEntryTransactions(ctx, userID, portfolioID, sourceID, out.valid)
	if err != nil {
		return portfoliodto.ImportResultResponseDTO{}, err
	}
	result.Imported = imported
	return result, nil
}
