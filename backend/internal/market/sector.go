package market

import (
	"regexp"
	"strings"

	"github.com/yeferson59/gofinance/v2/decimal"

	"github.com/yeferson59/finexia-app/internal/platform/spreadsheet"
)

// Sector is the line of business behind an asset — the answer to "if this one
// industry has a bad year, how much of my money is in it?".
//
// It is a different axis from AssetType and not a finer version of it: a type
// says what the instrument *is* (a share, a bond, a coin), a sector says what
// the money behind it *does*. Two stocks share a type and can sit in unrelated
// industries; a stock and a corporate bond of the same issuer share a sector
// and nothing else.
//
// Unlike the free-text columns beside it, this one needs no length check: the
// vocabulary is closed, every value below fits the assets.sector column
// (VARCHAR(32)) with room to spare, and nothing that is not one of them can be
// stored — NormalizeSector and IsValid are what a hand-typed label has to get
// past.
//
// The vocabulary is the eleven GICS sectors, lower-cased and underscored like
// AssetType. Not because the standard is authoritative here — nothing in this
// app is licensed from it — but because it is the split every broker statement
// and every provider already maps onto, so a user filling this in by hand and a
// provider feed added later land on the same set.
type Sector string

const (
	SectorTechnology     Sector = "technology"
	SectorHealthcare     Sector = "healthcare"
	SectorFinancials     Sector = "financials"
	SectorConsumerDisc   Sector = "consumer_discretionary"
	SectorConsumerStaple Sector = "consumer_staples"
	SectorEnergy         Sector = "energy"
	SectorIndustrials    Sector = "industrials"
	SectorMaterials      Sector = "materials"
	SectorUtilities      Sector = "utilities"
	SectorRealEstate     Sector = "real_estate"
	SectorCommunication  Sector = "communication_services"
)

// SectorUnclassified and SectorNotApplicable are the two buckets a sector
// breakdown adds and the catalog never stores. They are the difference between
// the two kinds of blank a NULL column would otherwise flatten into one:
//
//   - Unclassified is an asset that has a sector and nobody has filled it in.
//     It is work to do, and telling the user how much money is sitting in it is
//     the point — a breakdown that quietly dropped these rows would report
//     shares of a total that is not the user's money.
//   - NotApplicable is an asset with no sector to have: a coin, a currency
//     balance, a flat, a bar of gold. Nothing to fill in, ever.
//
// They are Sectors rather than a separate type so one enum covers every row a
// breakdown can produce, and IsValid rejects both so neither can be written to
// a catalog row.
const (
	SectorUnclassified  Sector = "unclassified"
	SectorNotApplicable Sector = "not_applicable"
)

// Sectors is the storable vocabulary, in the order a UI that offers a choice
// should list it. Excludes the two buckets above, which are derived and never
// stored.
var Sectors = []Sector{
	SectorTechnology,
	SectorCommunication,
	SectorHealthcare,
	SectorFinancials,
	SectorConsumerDisc,
	SectorConsumerStaple,
	SectorIndustrials,
	SectorEnergy,
	SectorMaterials,
	SectorUtilities,
	SectorRealEstate,
}

// IsValid reports whether the sector is one a catalog row may carry. The empty
// sector is not valid and is not meant to be: "no sector" is expressed by
// leaving the column NULL, which is what SectorNone below is for.
func (s Sector) IsValid() bool {
	switch s {
	case SectorTechnology, SectorCommunication, SectorHealthcare, SectorFinancials,
		SectorConsumerDisc, SectorConsumerStaple, SectorIndustrials, SectorEnergy,
		SectorMaterials, SectorUtilities, SectorRealEstate:
		return true
	default:
		return false
	}
}

// SectorNone is the unset sector: what a create or an edit sends to leave the
// column NULL, and what an asset nobody has classified reads back as.
const SectorNone Sector = ""

// HasSector reports whether an asset of this type can be classified at all.
//
// The line is drawn at "is there an operating business behind this". A share, a
// fund and a bond all have an issuer that works in an industry — a bond ETF is
// financials, a corporate bond is whatever its issuer is — while a coin, a cash
// balance, a flat and a bar of gold have none, and asking a user to pick one for
// them would be asking them to invent an answer.
//
// Other is on the classifiable side deliberately: it is the bucket the importer
// drops labels it could not place into, so it holds things that do have a
// sector more often than not, and leaving it out would hide them from the
// breakdown instead of prompting for them.
func (a AssetType) HasSector() bool {
	switch a {
	case Stock, ETF, Bond, Other:
		return true
	default:
		return false
	}
}

// sectorSynonyms maps what people and providers actually write to the eleven
// values above. Same mechanism as categorySynonyms, and for the same reason: a
// spreadsheet column is free text, and the alternative to a synonym table is
// rejecting rows over spelling.
//
// Keys are normalised by spreadsheet.NormKey (lower-cased, accent-stripped,
// punctuation-collapsed), so "Tecnología", "TECNOLOGIA" and "tecnologia" all
// arrive here as the same string and only one entry is needed.
//
// It carries the provider spellings too — Finnhub's finnhubIndustry and Alpha
// Vantage's OVERVIEW.Sector — even though nothing fetches them today. They cost
// a line each, and they are the mapping a BYO-key backfill would need on the
// day it is written.
var sectorSynonyms = map[string]Sector{
	// Technology
	"technology": SectorTechnology, "tech": SectorTechnology,
	"tecnologia": SectorTechnology, "information technology": SectorTechnology,
	"it": SectorTechnology, "software": SectorTechnology, "hardware": SectorTechnology,
	"semiconductors": SectorTechnology, "semiconductores": SectorTechnology,
	"electronic technology": SectorTechnology, "technology services": SectorTechnology,

	// Communication services
	"communication services": SectorCommunication, "communications": SectorCommunication,
	"comunicaciones": SectorCommunication, "servicios de comunicacion": SectorCommunication,
	"telecom": SectorCommunication, "telecomunicaciones": SectorCommunication,
	"telecommunication": SectorCommunication, "media": SectorCommunication,
	"medios": SectorCommunication, "entretenimiento": SectorCommunication,

	// Health care
	"healthcare": SectorHealthcare, "health care": SectorHealthcare,
	"salud": SectorHealthcare, "sanidad": SectorHealthcare,
	"pharmaceuticals": SectorHealthcare, "farmaceutica": SectorHealthcare,
	"biotechnology": SectorHealthcare, "biotecnologia": SectorHealthcare,
	"life sciences": SectorHealthcare, "health technology": SectorHealthcare,

	// Financials
	"financials": SectorFinancials, "financial services": SectorFinancials,
	"finance": SectorFinancials, "finanzas": SectorFinancials,
	"servicios financieros": SectorFinancials, "banking": SectorFinancials,
	"banca": SectorFinancials, "bancos": SectorFinancials, "banks": SectorFinancials,
	"insurance": SectorFinancials, "seguros": SectorFinancials,

	// Consumer discretionary
	"consumer discretionary": SectorConsumerDisc, "consumo discrecional": SectorConsumerDisc,
	"consumer cyclical": SectorConsumerDisc, "consumo ciclico": SectorConsumerDisc,
	"retail": SectorConsumerDisc, "comercio": SectorConsumerDisc,
	"automobiles": SectorConsumerDisc, "automoviles": SectorConsumerDisc,
	"hotels restaurants leisure": SectorConsumerDisc, "ocio": SectorConsumerDisc,

	// Consumer staples
	"consumer staples": SectorConsumerStaple, "consumo basico": SectorConsumerStaple,
	"consumer defensive": SectorConsumerStaple, "consumo defensivo": SectorConsumerStaple,
	"consumer products": SectorConsumerStaple, "food products": SectorConsumerStaple,
	"alimentacion": SectorConsumerStaple, "beverages": SectorConsumerStaple,
	"bebidas": SectorConsumerStaple, "tobacco": SectorConsumerStaple,

	// Energy
	"energy": SectorEnergy, "energia": SectorEnergy,
	"oil gas": SectorEnergy, "petroleo": SectorEnergy, "petroleo y gas": SectorEnergy,
	"energy transportation": SectorEnergy,

	// Industrials
	"industrials": SectorIndustrials, "industrial": SectorIndustrials,
	"industria": SectorIndustrials, "industriales": SectorIndustrials,
	"manufacturing": SectorIndustrials, "manufactura": SectorIndustrials,
	"aerospace defense": SectorIndustrials, "aeroespacial": SectorIndustrials,
	"machinery": SectorIndustrials, "maquinaria": SectorIndustrials,
	"logistics transportation": SectorIndustrials, "transporte": SectorIndustrials,
	"construccion": SectorIndustrials, "construction": SectorIndustrials,

	// Materials
	"materials": SectorMaterials, "materiales": SectorMaterials,
	"basic materials": SectorMaterials, "materiales basicos": SectorMaterials,
	"chemicals": SectorMaterials, "quimica": SectorMaterials,
	"metals mining": SectorMaterials, "mineria": SectorMaterials,
	"mining": SectorMaterials, "packaging": SectorMaterials,

	// Utilities
	"utilities": SectorUtilities, "servicios publicos": SectorUtilities,
	"utilidades": SectorUtilities, "electricidad": SectorUtilities,
	"agua": SectorUtilities, "gas": SectorUtilities,

	// Real estate
	"real estate": SectorRealEstate, "inmobiliario": SectorRealEstate,
	"inmobiliaria": SectorRealEstate, "bienes raices": SectorRealEstate,
	"reit": SectorRealEstate, "reits": SectorRealEstate,
}

// sectorConjunctions are the characters NormKey leaves alone and a sector label
// is full of: providers write "Oil & Gas", "Metals & Mining", "Oil, Gas &
// Consumable Fuels". They are dropped rather than added to NormKey's own list,
// which matches spreadsheet headers where an ampersand is rare and meaningful.
var sectorConjunctions = regexp.MustCompile(`[&,+]+`)

// sectorKey is NormKey plus that: the spelling a label is looked up under.
func sectorKey(raw string) string {
	key := sectorConjunctions.ReplaceAllString(spreadsheet.NormKey(raw), " ")

	return strings.Join(strings.Fields(key), " ")
}

// NormalizeSector maps a free-form sector label to a known Sector,
// accent- and case-insensitively.
//
// An empty input is not an error: it is the ordinary case of a row that does
// not classify its asset, and it returns SectorNone with ok true so callers do
// not have to tell "left blank" apart from "unrecognised" themselves. Only a
// non-empty label nobody recognises returns false.
func NormalizeSector(raw string) (Sector, bool) {
	key := sectorKey(raw)
	if key == "" {
		return SectorNone, true
	}

	if s, ok := sectorSynonyms[key]; ok {
		return s, true
	}

	// A caller may send the canonical value straight through — the API, the
	// seed, a re-import of this app's own export — and those never reach the
	// table: NormKey has already turned "consumer_discretionary" into a spaced
	// spelling, so the underscores go back before the check.
	if s := Sector(strings.ReplaceAll(key, " ", "_")); s.IsValid() {
		return s, true
	}

	return SectorNone, false
}

// SectorWeight is one industry's share of a single asset, as a percentage.
//
// A percentage and not a fraction because of where the number comes from: a
// fund fact sheet that reads "Information Technology 33.1%". The value is
// transcribed, not computed, and asking the operator to divide by a hundred
// first would be asking for the one arithmetic step that gets typed wrong.
type SectorWeight struct {
	Sector Sector          `json:"sector"`
	Weight decimal.Decimal `json:"weight"`
}

// SectorBreakdown is what an asset is made of when one sector cannot say it.
//
// It exists because the single sector beside it is the right answer for a share
// and for a sector fund — Apple is technology, XLK is technology — and the
// wrong answer for the instrument most people actually hold. VOO is the whole
// S&P 500: filing it under technology would put two thirds of the position in
// industries it is not in, and leaving it blank would report the largest
// holding in the portfolio as work somebody has to go and do.
//
// The two are exclusive, and the exclusivity is validated rather than modelled:
// an asset carries either a Sector or a SectorBreakdown, never both, so there
// is one place to read an asset's classification from and no way for the two to
// disagree. Empty means the asset has no breakdown, which is the ordinary case.
//
// The weights are not required to add up to 100 — see Validate — and the
// allocation normalises over whatever total it finds. They are ordered by
// nothing in particular here; the repository reads them back heaviest first,
// which is the order a fact sheet prints and the order a reader wants.
type SectorBreakdown []SectorWeight

// IsEmpty reports whether the asset has no breakdown. It reads better than a
// len() at the call sites that ask "is this classified the plain way or the
// weighted way", which is most of them.
func (b SectorBreakdown) IsEmpty() bool { return len(b) == 0 }

// Total adds the weights up. It is what Validate bounds and what a UI shows
// beside the inputs so the operator can see 97.3 % and decide whether that is
// the fund's cash sleeve or a row they forgot.
func (b SectorBreakdown) Total() decimal.Decimal {
	total := decimal.Zero
	for _, w := range b {
		total = total.Add(w.Weight)
	}

	return total
}

// hundred is the ceiling every weight and every total is held to. Built from a
// string so the parser that reads every other decimal in this app is what reads
// this one too.
var hundred = func() decimal.Decimal {
	d, err := decimal.NewFromString("100")
	if err != nil {
		panic("market: 100 is not a decimal: " + err.Error())
	}

	return d
}()

// Validate checks a breakdown against the four things that make it readable.
//
//   - Every sector is one of the eleven. The two derived buckets are rejected
//     with everything else: "unclassified" is what an asset with no breakdown
//     already reports, so a row claiming 8 % of it would be claiming a share of
//     a bucket that means the absence of a share.
//   - No sector appears twice. A fact sheet lists each once, so a repeat is a
//     transcription slip, and silently adding the two would hide it.
//   - Every weight is above 0 and at most 100. A zero-weight row says nothing a
//     missing row does not, and a negative one says nothing at all.
//   - The total is at most 100.
//
// What it deliberately does not require is a total of exactly 100. Fact sheets
// do not add to a hundred — a few tenths sit in cash, futures and receivables —
// and a hand-typed breakdown is further off still. Rejecting those would mean
// rejecting every real fund's published numbers, so the allocation normalises
// over the total it finds instead and an incomplete breakdown still accounts
// for the whole position. A total above 100 is the opposite case: not an
// incomplete transcription but a wrong one, and normalising it would quietly
// turn somebody's typo into a chart.
func (b SectorBreakdown) Validate() error {
	if b.IsEmpty() {
		return nil
	}

	seen := make(map[Sector]struct{}, len(b))
	for _, w := range b {
		if !w.Sector.IsValid() {
			return errAssetSectorInvalid
		}

		if _, dup := seen[w.Sector]; dup {
			return errAssetSectorWeightDuplicate
		}
		seen[w.Sector] = struct{}{}

		if !w.Weight.IsPos() || w.Weight.GreaterThan(hundred) {
			return errAssetSectorWeightRange
		}
	}

	if b.Total().GreaterThan(hundred) {
		return errAssetSectorWeightsTotal
	}

	return nil
}

// sectorWeightPairs splits the cell a spreadsheet carries a breakdown in, and
// sectorWeightSplit splits one pair into its two halves.
//
// One cell rather than eleven columns, because the sheet an operator builds has
// one row per asset and eleven mostly-empty columns would make every share in
// the file carry ten blanks. The spelling is the one somebody types by hand:
//
//	technology:33.1; financials:13.8; healthcare:10.2
//
// Both separators are generous on purpose. A newline inside a quoted CSV cell
// is what a paste from a PDF fact sheet produces, "=" is what a spreadsheet
// user reaches for when ":" looks like a time, and the percent sign comes along
// with the number whenever the column was copied rather than retyped.
var (
	sectorWeightPairs = regexp.MustCompile(`[;|\n\r]+`)
	sectorWeightSplit = regexp.MustCompile(`\s*[:=]\s*`)
)

// ParseSectorBreakdown reads a breakdown out of one spreadsheet cell.
//
// An empty cell is not an error, exactly like an empty sector column: it is the
// ordinary row that does not carry one, and it returns an empty breakdown with
// ok true. Anything else that cannot be read is false, and the importer skips
// the row rather than storing an asset whose fund the file described and the
// catalog does not.
//
// The sector half goes through NormalizeSector, so "Tecnología: 33,1" works for
// the same reason the plain sector column accepts it. The number half accepts a
// decimal comma — the separator most of this app's users type — because a
// weight between 0 and 100 has no thousands separator for it to be confused
// with.
func ParseSectorBreakdown(raw string) (SectorBreakdown, bool) {
	if strings.TrimSpace(raw) == "" {
		return nil, true
	}

	breakdown := make(SectorBreakdown, 0, len(Sectors))

	for _, pair := range sectorWeightPairs.Split(raw, -1) {
		if strings.TrimSpace(pair) == "" {
			continue
		}

		parts := sectorWeightSplit.Split(strings.TrimSpace(pair), 2)
		if len(parts) != 2 {
			return nil, false
		}

		sector, ok := NormalizeSector(parts[0])
		if !ok || sector == SectorNone {
			return nil, false
		}

		number := strings.TrimSuffix(strings.TrimSpace(parts[1]), "%")
		number = strings.ReplaceAll(strings.TrimSpace(number), ",", ".")

		weight, err := decimal.NewFromString(number)
		if err != nil {
			return nil, false
		}

		breakdown = append(breakdown, SectorWeight{Sector: sector, Weight: weight})
	}

	if breakdown.IsEmpty() {
		return nil, false
	}

	return breakdown, true
}
