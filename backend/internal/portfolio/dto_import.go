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
	// FXRate is the rate that row's trade settled at, into the account currency
	// named once for the whole file by ImportDefaultsDTO.CostCurrency. Unmapped
	// means every row settled at 1, which is what a single-currency statement
	// says and what every import before this field did.
	FXRate   *int `json:"fxRate"`
	Category *int `json:"category"`
	Notes    *int `json:"notes"`
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
	// CostCurrency is the account's: the one the broker debited, and the one
	// every position this import opens will carry its cost in. It belongs to the
	// upload rather than to a row because a statement is one account, so its
	// settlement currency is a property of the file.
	//
	// Empty means "the same as the row's", which is what every import before
	// this field produced and what a single-currency statement means. Setting it
	// to something else is what makes the FXRate column meaningful.
	CostCurrency string `json:"costCurrency"`
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
	FXRate    string   `json:"fxRate"`
	// CostCurrency is echoed per row even though it comes from the defaults, so
	// the preview table can show what each row will actually cost in without the
	// client having to join the row against the upload's settings.
	CostCurrency string   `json:"costCurrency"`
	Category     string   `json:"category"`
	Notes        string   `json:"notes"`
	Valid        bool     `json:"valid"`
	Errors       []string `json:"errors"`
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

// ImportResultErrorDTO reports why a single row was skipped during the
// transaction commit.
type ImportResultErrorDTO struct {
	Row     int    `json:"row"`
	Message string `json:"message"`
}

// ImportResultResponseDTO summarises a transaction import commit.
type ImportResultResponseDTO struct {
	TotalRows int                    `json:"totalRows"`
	Imported  int                    `json:"imported"`
	Skipped   int                    `json:"skipped"`
	Errors    []ImportResultErrorDTO `json:"errors"`
}
