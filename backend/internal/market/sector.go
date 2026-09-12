package market

import (
	"regexp"
	"strings"

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
