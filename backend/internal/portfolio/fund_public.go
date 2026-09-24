package portfolio

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"uuid"

	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
	"github.com/yeferson59/finexia-app/internal/platform/marketdata/sfc"
)

// Funds whose unit value the Superintendencia Financiera publishes (§12 of
// docs/PLAN_FONDOS_INVERSION.md, migration 000059).
//
// Every Colombian FIC reports the value of its unit each day, and the SFC
// publishes them as open data. The owner of a fund followed by units can link
// it to its entry in that catalog, and from then on the published values are
// its marks: nobody has to copy them from the statement. They are written into
// fund_marks with source = 'public', so the price, the snapshots and the
// returns follow them the way they follow a mark the owner typed.
//
// Two rules keep the owner in charge. A mark they wrote is never overwritten by
// a published one on the same day. And unlinking takes back every published
// mark, so a link made to the wrong type of participation leaves nothing
// behind.
//
// A fund followed by balance cannot be linked: its units are synthetic, and a
// published unit value means nothing against them.

var (
	// ErrPublicFundNotFound answers for a catalog entry that does not exist.
	ErrPublicFundNotFound = httpx.AsNotFound(errors.New("public fund not found"))
	// ErrFundNotLinkable refuses to link a fund whose units are not the
	// published fund's: one followed by balance, or held in another currency.
	ErrFundNotLinkable = httpx.AsConflict(errors.New("only a fund followed by units, in COP, can be linked to a published fund"))
	// ErrPublicFundsUnavailable means the SFC did not answer. The same request
	// can work later.
	ErrPublicFundsUnavailable = httpx.AsUnavailable(errors.New("the published fund values could not be read"))
	// ErrInvalidPublicFundSearch rejects a search too short to narrow a
	// thousand funds down.
	ErrInvalidPublicFundSearch = httpx.AsBadRequest(errors.New("invalid public fund search"))
)

// publicFundCurrency is what the SFC's unit values are stated in.
const publicFundCurrency = money.COP

// FundMarkSource says who wrote a mark.
type FundMarkSource string

const (
	// FundMarkByUser is a mark the owner wrote. It always wins.
	FundMarkByUser FundMarkSource = "user"
	// FundMarkPublished is a unit value the SFC published, written through the
	// fund's link.
	FundMarkPublished FundMarkSource = "public"
)

const (
	// minPublicFundSearch is the shortest search that is run: two letters
	// already match hundreds of funds.
	minPublicFundSearch = 2
	// maxPublicFundSearch keeps a search inside what anyone types.
	maxPublicFundSearch = 100
	// publicFundSearchLimit is how many funds a search answers with.
	publicFundSearchLimit = 25
	// publicFundStaleDays hides from a search a fund the SFC stopped
	// publishing: liquidated or merged, it cannot price anything new.
	publicFundStaleDays = 60
	// publicFundBackfillDays is how far back a link imports when the fund has
	// no purchase to start from.
	publicFundBackfillDays = 365
	// publicFundOpeningLookback is how many days before a purchase a published
	// value can be to stand for its unit value, when the owner left it out:
	// the day before a long weekend, not a month ago.
	publicFundOpeningLookback = 7
)

// PublicFundSource reads the SFC's open data. Satisfied by *sfc.Client.
type PublicFundSource interface {
	LatestFunds(ctx context.Context) ([]sfc.Fund, error)
	UnitValues(ctx context.Context, key sfc.FundKey, from time.Time) ([]sfc.UnitValue, error)
}

// PublicFund is one type of participation of one fund in the SFC's catalog.
type PublicFund struct {
	ID            string `json:"id"`
	EntityName    string `json:"entityName"`
	FundName      string `json:"fundName"`
	FundKind      string `json:"fundKind"`
	FundCode      int    `json:"fundCode"`
	Participation int    `json:"participation"`
	// UnitValue is the latest one published, on ValueDate, in COP.
	UnitValue string    `json:"unitValue"`
	ValueDate time.Time `json:"valueDate"`
	Investors int       `json:"investors"`

	key sfc.FundKey
}

// Key is the fund's key in the feed.
func (f PublicFund) Key() sfc.FundKey {
	return f.key
}

// PublicFundLink is what a fund is linked to, as a fund shows it.
type PublicFundLink struct {
	ID            string `json:"id"`
	EntityName    string `json:"entityName"`
	FundName      string `json:"fundName"`
	Participation int    `json:"participation"`
}

// PublicFundValue is one published unit value, as it is written as a mark.
type PublicFundValue struct {
	Date      time.Time
	UnitValue decimal.Decimal
}

// LinkedFund is one fund some owner linked to a published one, and the day
// the next import starts from.
type LinkedFund struct {
	UserID       uuid.UUID
	AssetID      uuid.UUID
	PublicFundID string
	// Since is the day after the latest published mark it has, or — before it
	// has any — the day its first purchase was made.
	Since time.Time
}

// newPublicFund reads a fund of the feed into the catalog's shape.
func newPublicFund(f sfc.Fund) PublicFund {
	return PublicFund{
		ID:            f.Key.String(),
		EntityName:    f.EntityName,
		FundName:      f.Name,
		FundKind:      f.Kind,
		FundCode:      f.Key.Fund,
		Participation: f.Key.Participation,
		UnitValue:     f.UnitValue,
		ValueDate:     f.Date,
		Investors:     f.Investors,
		key:           f.Key,
	}
}

// searchText is what a search is matched against: the names, lower case and
// without accents, and the fund's codes, so "3644" finds a fund by the code a
// statement prints.
func (f PublicFund) searchText() string {
	return normalizeSearch(fmt.Sprintf("%s %s %d %d", f.EntityName, f.FundName, f.FundCode, f.Participation))
}

// normalizeSearch folds text for matching: lower case, accents off, spaces
// collapsed.
func normalizeSearch(s string) string {
	var b strings.Builder

	space := true

	for _, r := range strings.ToLower(s) {
		r = foldAccent(r)

		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)

			space = false
		case !space:
			b.WriteByte(' ')

			space = true
		}
	}

	return strings.TrimSpace(b.String())
}

func foldAccent(r rune) rune {
	switch r {
	case 'á', 'à', 'ä', 'â':
		return 'a'
	case 'é', 'è', 'ë', 'ê':
		return 'e'
	case 'í', 'ì', 'ï', 'î':
		return 'i'
	case 'ó', 'ò', 'ö', 'ô':
		return 'o'
	case 'ú', 'ù', 'ü', 'û':
		return 'u'
	case 'ñ':
		return 'n'
	}

	return r
}

// publicFundSearchWords splits a search into the words every match must hold.
func publicFundSearchWords(q string) ([]string, error) {
	q = strings.TrimSpace(q)

	if utf8.RuneCountInString(q) > maxPublicFundSearch {
		return nil, fmt.Errorf("%w: at most %d characters", ErrInvalidPublicFundSearch, maxPublicFundSearch)
	}

	words := strings.Fields(normalizeSearch(q))
	if len(strings.Join(words, "")) < minPublicFundSearch {
		return nil, fmt.Errorf("%w: type at least %d letters", ErrInvalidPublicFundSearch, minPublicFundSearch)
	}

	return words, nil
}

// publicFundValues reads the feed's values into marks, dropping any that do
// not fit a mark: a feed is not more trusted than the owner.
func publicFundValues(in []sfc.UnitValue) []PublicFundValue {
	out := make([]PublicFundValue, 0, len(in))

	for _, v := range in {
		d, err := decimal.NewFromString(v.Value)
		if err != nil || validateUnitValue(d, invalidFundMark) != nil {
			continue
		}

		out = append(out, PublicFundValue{Date: cashRateDay(v.Date), UnitValue: d.RoundHAZ(8)})
	}

	return out
}

// publicValueOn is the published value that stands for a purchase on day: the
// one of that day, or the latest before it within the lookback. values are
// oldest first.
func publicValueOn(values []PublicFundValue, day time.Time) (decimal.Decimal, bool) {
	day = cashRateDay(day)
	earliest := day.AddDate(0, 0, -publicFundOpeningLookback)

	for i := len(values) - 1; i >= 0; i-- {
		d := values[i].Date
		if d.After(day) {
			continue
		}

		if d.Before(earliest) {
			break
		}

		return values[i].UnitValue, true
	}

	return decimal.Decimal{}, false
}

// linkable checks that a fund can take a published unit value.
func (f Fund) linkable() error {
	if f.Tracking != FundUnits || f.Currency != publicFundCurrency {
		return ErrFundNotLinkable
	}

	return nil
}
