package market

// ImportResultErrorDTO reports why a single spreadsheet row was skipped.
type ImportResultErrorDTO struct {
	Row     int    `json:"row"`
	Message string `json:"message"`
}

// ImportResultResponseDTO summarises a bulk import (assets or exchange rates):
// how many rows were seen, imported, skipped and why.
type ImportResultResponseDTO struct {
	TotalRows int                    `json:"totalRows"`
	Imported  int                    `json:"imported"`
	Skipped   int                    `json:"skipped"`
	Errors    []ImportResultErrorDTO `json:"errors"`
}
