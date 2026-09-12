package market

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
	"github.com/yeferson59/finexia-app/internal/platform/spreadsheet"
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
	"sector":    {"sector", "industria", "industry", "sector economico", "rubro"},
	// The breakdown lives in one cell of its own rather than eleven columns, so
	// the shares in the file — which are every row but a handful — do not each
	// carry ten blanks. ParseSectorBreakdown documents the spelling.
	"sectorweights": {
		"sectorweights", "sector weights", "desglose", "desglose sectorial",
		"pesos", "pesos por sector", "industrias", "sector breakdown",
	},
}

// ImportAssetsFromFile parses an uploaded CSV/XLSX with columns
// ticker, name, assetType, currency (required) and exchange, sector,
// sectorWeights (optional), upserting one asset per valid row. Invalid rows are
// skipped and reported.
//
// A spreadsheet is how a catalog gets classified in bulk — eleven sectors over
// a few hundred tickers is not an afternoon of clicking through an edit form —
// so the sector column is read here even though it is optional. An unreadable
// label skips the row rather than importing the asset unclassified: the file
// said something about it, and quietly ignoring the one column the operator
// opened the spreadsheet to fill would report an import that did not happen.
//
// sectorWeights is the same column for the asset one industry cannot describe.
// It matters more here than anywhere else in the app: a fund's eleven weights
// are the most tedious thing in the catalog to enter by hand, and they are
// exactly what a paste from a fact sheet already has.
func (s *service) ImportAssetsFromFile(ctx context.Context, data []byte, filename, sheet string) (ImportResultResponseDTO, error) {
	src, err := spreadsheet.ReadFile(data, filename, sheet)
	if err != nil {
		return ImportResultResponseDTO{}, httpx.AsBadRequest(err)
	}

	headerIdx := firstNonEmptyRow(src.Rows)
	if headerIdx == -1 {
		return ImportResultResponseDTO{}, httpx.AsBadRequest(errors.New("invalid spreadsheet: the file is empty"))
	}

	cols := mapSimpleHeaders(src.Rows[headerIdx], assetHeaderSynonyms)
	if missing := missingCols(cols, "ticker", "name", "assettype", "currency"); len(missing) > 0 {
		return ImportResultResponseDTO{}, httpx.AsBadRequest(fmt.Errorf("invalid spreadsheet: missing required columns: %s", strings.Join(missing, ", ")))
	}

	dataRows := src.Rows[headerIdx+1:]
	if len(dataRows) > maxAssetImportRows {
		return ImportResultResponseDTO{}, httpx.AsTooManyRequests(fmt.Errorf("invalid spreadsheet: too many rows (max %d)", maxAssetImportRows))
	}

	exchangeIdx, hasExchange := cols["exchange"]
	sectorIdx, hasSector := cols["sector"]
	weightsIdx, hasWeights := cols["sectorweights"]
	result := ImportResultResponseDTO{Errors: []ImportResultErrorDTO{}}

	for i, row := range dataRows {
		if spreadsheet.RowIsEmpty(row) {
			continue
		}
		rowNumber := headerIdx + 2 + i
		result.TotalRows++

		ticker := strings.ToUpper(spreadsheet.CellAtIdx(row, cols["ticker"]))
		name := spreadsheet.CellAtIdx(row, cols["name"])
		assetTypeRaw := spreadsheet.CellAtIdx(row, cols["assettype"])
		currencyRaw := strings.ToUpper(spreadsheet.CellAtIdx(row, cols["currency"]))
		exchange := ""
		if hasExchange {
			exchange = spreadsheet.CellAtIdx(row, exchangeIdx)
		}
		sectorRaw := ""
		if hasSector {
			sectorRaw = spreadsheet.CellAtIdx(row, sectorIdx)
		}
		weightsRaw := ""
		if hasWeights {
			weightsRaw = spreadsheet.CellAtIdx(row, weightsIdx)
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

		assetType, ok := NormalizeAssetType(assetTypeRaw)
		if !ok {
			if assetTypeRaw == "" {
				rowErrs = append(rowErrs, "el tipo de activo está vacío")
			} else {
				rowErrs = append(rowErrs, fmt.Sprintf("tipo de activo no reconocido: %q", assetTypeRaw))
			}
		}

		currency, ok := normalizeCurrencyCode(currencyRaw)
		if !ok {
			rowErrs = append(rowErrs, fmt.Sprintf("moneda inválida: %q", currencyRaw))
		}

		sector, ok := NormalizeSector(sectorRaw)
		switch {
		case !ok:
			rowErrs = append(rowErrs, fmt.Sprintf("sector no reconocido: %q", sectorRaw))
		// assetType.IsValid() guards the second message rather than the row:
		// an unrecognised type has already failed above, and the type this
		// would print is the empty one the normaliser handed back, not
		// anything the file says.
		case sector != SectorNone && assetType.IsValid() && !assetType.HasSector():
			rowErrs = append(rowErrs, fmt.Sprintf("%q no lleva sector: solo acciones, ETFs, bonos y otros", assetTypeRaw))
		}

		// The breakdown's own two failures, told apart on purpose: a cell that
		// could not be read at all is a spelling problem the operator fixes in
		// the file, and one that could be read but does not add up is a number
		// problem they fix against the fact sheet. Its arithmetic is checked by
		// the same Validate the API calls, so a row the importer accepts is a
		// row the edit form would have accepted too.
		weights, ok := ParseSectorBreakdown(weightsRaw)
		switch {
		case !ok:
			rowErrs = append(rowErrs, fmt.Sprintf("desglose por industrias ilegible: %q (se espera «technology: 33.1; financials: 13.8»)", weightsRaw))
		case !weights.IsEmpty() && sector != SectorNone:
			rowErrs = append(rowErrs, "la fila trae sector y desglose: deja solo uno de los dos")
		case !weights.IsEmpty() && assetType.IsValid() && !assetType.HasSector():
			rowErrs = append(rowErrs, fmt.Sprintf("%q no lleva desglose por industrias: solo acciones, ETFs, bonos y otros", assetTypeRaw))
		default:
			if err := weights.Validate(); err != nil {
				rowErrs = append(rowErrs, err.Error())
			}
		}

		if len(rowErrs) > 0 {
			result.Skipped++
			if len(result.Errors) < 100 {
				result.Errors = append(result.Errors, ImportResultErrorDTO{Row: rowNumber, Message: strings.Join(rowErrs, "; ")})
			}
			continue
		}

		if _, err := s.repo.UpsertAsset(ctx, AssetSpec{
			Ticker:        ticker,
			Name:          name,
			AssetType:     assetType,
			Exchange:      exchange,
			Currency:      currency,
			Sector:        sector,
			SectorWeights: weights,
		}); err != nil {
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
