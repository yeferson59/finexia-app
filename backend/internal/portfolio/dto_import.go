package portfolio

// ImportMappingDTO maps each transaction field to a 0-based column index of
// the uploaded spreadsheet. A nil entry means the field is not present in the
// file and falls back to the import defaults (or empty).
type ImportMappingDTO struct {
	Date      *int `json:"date"`
	Type      *int `json:"type"`
	Ticker    *int `json:"ticker"`
	AssetName *int `json:"assetName"`
	Quantity  *int `json:"quantity"`
	Price     *int `json:"price"`
	Fees      *int `json:"fees"`
	Currency  *int `json:"currency"`
	Category  *int `json:"category"`
	Notes     *int `json:"notes"`
}

// ImportDefaultsDTO holds fallback values applied to rows whose column is not
// mapped or whose cell is empty.
type ImportDefaultsDTO struct {
	// Type is the transaction type used when the file has no type column
	// (defaults to "buy").
	Type string `json:"type"`
	// Currency is the ISO-4217 code used when the file has no currency column
	// (defaults to "USD").
	Currency string `json:"currency"`
	// Category is the asset type used when the file has no category column
	// (defaults to "stock").
	Category string `json:"category"`
	// DateFormat disambiguates numeric dates: "auto", "dmy" or "mdy".
	DateFormat string `json:"dateFormat"`
}

// ImportRowDTO is one spreadsheet row after applying the column mapping and
// normalising every value. Invalid rows carry per-row error messages so the
// user can fix the file (or the mapping) before committing the import.
type ImportRowDTO struct {
	RowNumber int      `json:"rowNumber"`
	Raw       []string `json:"raw"`
	Date      string   `json:"date"`
	Type      string   `json:"type"`
	Ticker    string   `json:"ticker"`
	AssetName string   `json:"assetName"`
	Quantity  string   `json:"quantity"`
	Price     string   `json:"price"`
	Fees      string   `json:"fees"`
	Currency  string   `json:"currency"`
	Category  string   `json:"category"`
	Notes     string   `json:"notes"`
	Valid     bool     `json:"valid"`
	Errors    []string `json:"errors"`
}

type ImportPreviewResponseDTO struct {
	Sheets           []string         `json:"sheets"`
	Sheet            string           `json:"sheet"`
	HeaderRow        int              `json:"headerRow"`
	Headers          []string         `json:"headers"`
	SuggestedMapping ImportMappingDTO `json:"suggestedMapping"`
	// MissingFields lists required fields (date, ticker, quantity, price) that
	// the active mapping leaves unassigned; rows are not validated until the
	// user maps them.
	MissingFields []string `json:"missingFields"`
	TotalRows     int      `json:"totalRows"`
	ValidRows     int      `json:"validRows"`
	InvalidRows   int      `json:"invalidRows"`
	// Rows is capped (see importPreviewRowCap); TotalRows/ValidRows/
	// InvalidRows always cover the whole file.
	Rows []ImportRowDTO `json:"rows"`
}

type ImportResultErrorDTO struct {
	Row     int    `json:"row"`
	Message string `json:"message"`
}

type ImportResultResponseDTO struct {
	TotalRows int                    `json:"totalRows"`
	Imported  int                    `json:"imported"`
	Skipped   int                    `json:"skipped"`
	Errors    []ImportResultErrorDTO `json:"errors"`
}
