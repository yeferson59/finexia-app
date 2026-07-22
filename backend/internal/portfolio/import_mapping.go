package portfolio

// This file holds column-mapping and file-parsing helpers for the
// transaction importer: field synonyms, header detection and CSV/XLSX
// decoding. Split out of import.go to keep each file under ~500 lines.

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

var txnTypeSynonyms = map[string]TransactionType{
	"buy": Buy, "compra": Buy, "compras": Buy, "purchase": Buy,
	"bought": Buy, "adquisicion": Buy, "cpra": Buy,
	"sell": Sell, "venta": Sell, "ventas": Sell, "sale": Sell,
	"sold": Sell, "vender": Sell, "vta": Sell,
	"dividend": Dividend, "dividendo": Dividend, "dividendos": Dividend, "div": Dividend,
	"interest": Interest, "interes": Interest, "intereses": Interest, "rendimiento": Interest, "rendimientos": Interest,
	"fee": Fee, "fees": Fee, "comision": Fee, "comisiones": Fee, "cargo": Fee,
	"split": Split, "division": Split,
	"transfer in": TransferIn, "transfer_in": TransferIn, "transferencia entrada": TransferIn,
	"deposito": TransferIn, "deposit": TransferIn, "entrada": TransferIn, "ingreso": TransferIn, "aporte": TransferIn,
	"transfer out": TransferOut, "transfer_out": TransferOut, "transferencia salida": TransferOut,
	"retiro": TransferOut, "withdrawal": TransferOut, "salida": TransferOut, "egreso": TransferOut,
}

var categorySynonyms = map[string]AssetType{
	"stock": Stock, "stocks": Stock, "accion": Stock, "acciones": Stock,
	"equity": Stock, "equities": Stock,
	"etf": ETF, "etfs": ETF, "fondo": ETF, "fondos": ETF,
	"fund": ETF, "funds": ETF, "fondo indexado": ETF, "index fund": ETF,
	"crypto": Crypto, "cripto": Crypto, "criptomoneda": Crypto,
	"criptomonedas": Crypto, "cryptocurrency": Crypto, "criptos": Crypto,
	"bond": Bond, "bonds": Bond, "bono": Bond, "bonos": Bond,
	"renta fija": Bond, "cdt": Bond, "fixed income": Bond,
	"cash": Cash, "efectivo": Cash, "liquidez": Cash, "dinero": Cash,
	"real estate": RealEstate, "real_estate": RealEstate, "inmueble": RealEstate,
	"inmuebles": RealEstate, "bienes raices": RealEstate, "reit": RealEstate, "fibra": RealEstate,
	"commodity": Commodity, "commodities": Commodity, "materia prima": Commodity,
	"materias primas": Commodity, "oro": Commodity, "gold": Commodity, "plata": Commodity,
	"other": Other, "otro": Other, "otros": Other,
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
func suggestMapping(headers []string) ImportMappingDTO {
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
	return ImportMappingDTO{
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
	for i := range limit {
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
