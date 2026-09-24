package portfolio

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"testing"
	"time"

	"uuid"

	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/platform/logger"
	"github.com/yeferson59/finexia-app/internal/platform/marketdata/sfc"
)

// fakePublicFunds stands for the SFC's feed.
type fakePublicFunds struct {
	funds  []sfc.Fund
	values []sfc.UnitValue
	err    error
	// asked records every UnitValues call, as key and day.
	asked []string
}

func (f *fakePublicFunds) LatestFunds(context.Context) ([]sfc.Fund, error) {
	return f.funds, f.err
}

func (f *fakePublicFunds) UnitValues(_ context.Context, key sfc.FundKey, from time.Time) ([]sfc.UnitValue, error) {
	f.asked = append(f.asked, key.String()+"@"+from.Format(time.DateOnly))

	var out []sfc.UnitValue

	for _, v := range f.values {
		if !v.Date.Before(from) {
			out = append(out, v)
		}
	}

	return out, f.err
}

func sfcDay(s string) time.Time {
	d, err := time.Parse(time.DateOnly, s)
	if err != nil {
		panic(err)
	}

	return d
}

var latamKey = sfc.FundKey{EntityType: 5, Entity: 31, Fund: 3644, Compartment: 1, Participation: 501}

func latamFund() PublicFund {
	return newPublicFund(sfc.Fund{
		Key: latamKey, EntityName: "Fiduciaria Bancolombia S.A.", Name: "FONDO DE INVERSIÓN COLECTIVA RENTA ACCIONES LATAM",
		Kind: "FIC DE TIPO GENERAL", UnitValue: "77723.164524", Date: sfcDay("2026-09-22"), Investors: 6383,
	})
}

func TestNormalizeSearch(t *testing.T) {
	cases := map[string]string{
		"FONDO DE INVERSIÓN COLECTIVA":     "fondo de inversion colectiva",
		"  Fiduciaria   Bancolombia S.A. ": "fiduciaria bancolombia s a",
		"Compañía — Señor Ñandú":           "compania senor nandu",
		"":                                 "",
	}

	for in, want := range cases {
		if got := normalizeSearch(in); got != want {
			t.Errorf("normalizeSearch(%q) = %q, want %q", in, got, want)
		}
	}

	if got := latamFund().searchText(); got != "fiduciaria bancolombia s a fondo de inversion colectiva renta acciones latam 3644 501" {
		t.Errorf("searchText = %q", got)
	}
}

func TestPublicFundSearchWords(t *testing.T) {
	words, err := publicFundSearchWords("  Rénta  LATAM ")
	if err != nil || !slices.Equal(words, []string{"renta", "latam"}) {
		t.Fatalf("words = %q, %v", words, err)
	}

	for _, bad := range []string{"", " ", "a", "--", string(make([]byte, 101))} {
		if _, err := publicFundSearchWords(bad); !errors.Is(err, ErrInvalidPublicFundSearch) {
			t.Errorf("publicFundSearchWords(%q) = %v, want ErrInvalidPublicFundSearch", bad, err)
		}
	}
}

func TestPublicFundValues(t *testing.T) {
	got := publicFundValues([]sfc.UnitValue{
		{Date: sfcDay("2026-09-20"), Value: "77700.123456789"},
		{Date: sfcDay("2026-09-21"), Value: "abc"},
		{Date: sfcDay("2026-09-22"), Value: "0"},
		{Date: sfcDay("2026-09-23"), Value: "1000000000000"},
	})

	// A value that does not fit a mark is dropped, and the one kept is rounded
	// to what a mark stores.
	if len(got) != 1 || got[0].UnitValue.String() != "77700.12345679" || !got[0].Date.Equal(sfcDay("2026-09-20")) {
		t.Fatalf("values = %+v", got)
	}
}

func TestPublicValueOn(t *testing.T) {
	values := []PublicFundValue{
		{Date: sfcDay("2026-09-01"), UnitValue: mustDecimal(t, "100")},
		{Date: sfcDay("2026-09-10"), UnitValue: mustDecimal(t, "101")},
		{Date: sfcDay("2026-09-11"), UnitValue: mustDecimal(t, "102")},
	}

	cases := []struct {
		day  string
		want string
	}{
		{"2026-09-10", "101"},
		{"2026-09-15", "102"},
		{"2026-09-18", "102"},
		{"2026-09-19", ""}, // eight days after the last value
		{"2026-08-31", ""}, // before the first
		{"2026-09-05", "100"},
	}

	for _, tc := range cases {
		got, ok := publicValueOn(values, sfcDay(tc.day))
		if tc.want == "" {
			if ok {
				t.Errorf("publicValueOn(%s) = %s, want none", tc.day, got)
			}

			continue
		}

		if !ok || got.String() != tc.want {
			t.Errorf("publicValueOn(%s) = %s (%v), want %s", tc.day, got, ok, tc.want)
		}
	}
}

func TestGroupLinkedFundsAndValuesSince(t *testing.T) {
	a, b := uuid.New(), uuid.New()
	groups := groupLinkedFunds([]LinkedFund{
		{UserID: a, PublicFundID: "1-1-1-1-1"},
		{UserID: b, PublicFundID: "1-1-1-1-1"},
		{UserID: a, PublicFundID: "2-2-2-2-2"},
	})

	if len(groups) != 2 || len(groups[0]) != 2 || len(groups[1]) != 1 {
		t.Fatalf("groups = %+v", groups)
	}

	values := []PublicFundValue{{Date: sfcDay("2026-09-20")}, {Date: sfcDay("2026-09-21")}, {Date: sfcDay("2026-09-22")}}

	if got := valuesSince(values, sfcDay("2026-09-21")); len(got) != 2 || !got[0].Date.Equal(sfcDay("2026-09-21")) {
		t.Errorf("valuesSince(21) = %+v", got)
	}

	if got := valuesSince(values, sfcDay("2026-09-23")); len(got) != 0 {
		t.Errorf("valuesSince(23) = %+v", got)
	}
}

func publicFundService(repo *fakeRepository, feed *fakePublicFunds) *service {
	svc := newService(repo, testConfig(), nil, nil, nil, logger.Noop())
	if feed != nil {
		svc.publicFunds = feed
	}

	return svc
}

func TestLinkFund(t *testing.T) {
	userID, assetID := uuid.New(), uuid.New()
	public := latamFund()

	newRepo := func(fund Fund, linked *[]PublicFundValue) *fakeRepository {
		return new(fakeRepository{
			getFund: func(context.Context, uuid.UUID, uuid.UUID) (Fund, error) { return fund, nil },
			getPublicFund: func(_ context.Context, id string) (PublicFund, error) {
				if id != public.ID {
					return PublicFund{}, ErrPublicFundNotFound
				}

				return public, nil
			},
			getFundMovements: func(context.Context, uuid.UUID, uuid.UUID) ([]FundMovement, error) {
				// Most recent first, as the repository answers.
				return []FundMovement{{Date: sfcDay("2026-09-15")}, {Date: sfcDay("2026-09-10")}}, nil
			},
			linkFund: func(_ context.Context, _, _ uuid.UUID, id string, values []PublicFundValue) (Fund, error) {
				*linked = values
				fund.PublicFund = &PublicFundLink{ID: id}

				return fund, nil
			},
		})
	}

	feed := &fakePublicFunds{values: []sfc.UnitValue{
		{Date: sfcDay("2026-09-09"), Value: "77000"},
		{Date: sfcDay("2026-09-10"), Value: "77100"},
		{Date: sfcDay("2026-09-22"), Value: "77723.164524"},
	}}

	t.Run("imports from the first purchase", func(t *testing.T) {
		var linked []PublicFundValue

		fund, err := publicFundService(newRepo(Fund{Tracking: FundUnits, Currency: money.COP}, &linked), feed).
			LinkFund(context.Background(), userID, assetID, public.ID)
		if err != nil {
			t.Fatalf("LinkFund: %v", err)
		}

		if fund.PublicFund == nil || fund.PublicFund.ID != public.ID {
			t.Errorf("fund = %+v", fund)
		}

		if last := feed.asked[len(feed.asked)-1]; last != "5-31-3644-1-501@2026-09-10" {
			t.Errorf("asked = %q, want from the first purchase", last)
		}

		if len(linked) != 2 || linked[1].UnitValue.String() != "77723.164524" {
			t.Errorf("linked values = %+v", linked)
		}
	})

	t.Run("refuses a fund that is not linkable", func(t *testing.T) {
		for _, fund := range []Fund{
			{Tracking: FundBalance, Currency: money.COP},
			{Tracking: FundUnits, Currency: money.USD},
		} {
			var linked []PublicFundValue

			_, err := publicFundService(newRepo(fund, &linked), feed).LinkFund(context.Background(), userID, assetID, public.ID)
			if !errors.Is(err, ErrFundNotLinkable) {
				t.Errorf("LinkFund(%s, %s) = %v, want ErrFundNotLinkable", fund.Tracking, fund.Currency, err)
			}
		}
	})

	t.Run("an unknown published fund", func(t *testing.T) {
		var linked []PublicFundValue

		_, err := publicFundService(newRepo(Fund{Tracking: FundUnits, Currency: money.COP}, &linked), feed).
			LinkFund(context.Background(), userID, assetID, "9-9-9-9-9")
		if !errors.Is(err, ErrPublicFundNotFound) {
			t.Errorf("err = %v, want ErrPublicFundNotFound", err)
		}
	})

	t.Run("the SFC does not answer", func(t *testing.T) {
		var linked []PublicFundValue

		down := &fakePublicFunds{err: errors.New("status 500")}

		_, err := publicFundService(newRepo(Fund{Tracking: FundUnits, Currency: money.COP}, &linked), down).
			LinkFund(context.Background(), userID, assetID, public.ID)
		if !errors.Is(err, ErrPublicFundsUnavailable) || linked != nil {
			t.Errorf("err = %v (linked %v), want ErrPublicFundsUnavailable and nothing written", err, linked)
		}
	})
}

func TestCreateLinkedFund(t *testing.T) {
	public := latamFund()
	feed := &fakePublicFunds{values: []sfc.UnitValue{
		{Date: sfcDay("2026-09-08"), Value: "77000"},
		{Date: sfcDay("2026-09-22"), Value: "77723.164524"},
	}}

	var created NewFundInput

	repo := new(fakeRepository{
		getPublicFund: func(context.Context, string) (PublicFund, error) { return public, nil },
		createFund: func(_ context.Context, _ uuid.UUID, in NewFundInput) (Fund, error) {
			created = in

			return Fund{Name: in.CleanName()}, nil
		},
	})
	svc := publicFundService(repo, feed)

	in := NewFundInput{
		PortfolioID: uuid.New(), SourceID: uuid.New(), Currency: money.COP, Tracking: FundUnits,
		Date: sfcDay("2026-09-10"), Units: mustDecimal(t, "10"), PublicFundID: public.ID,
	}

	if _, err := svc.CreateFund(context.Background(), uuid.New(), in); err != nil {
		t.Fatalf("CreateFund: %v", err)
	}

	// The name and the unit value come from the catalog, the unit value as of
	// the last day published before the purchase.
	if created.Name != public.FundName || created.UnitValue.String() != "77000" || len(created.publicValues) != 1 {
		t.Errorf("created = name %q, unit value %s, %d values", created.Name, created.UnitValue, len(created.publicValues))
	}

	in.Date = sfcDay("2026-09-20")

	if _, err := svc.CreateFund(context.Background(), uuid.New(), in); !errors.Is(err, ErrInvalidFund) {
		t.Errorf("a purchase twelve days after the last value = %v, want ErrInvalidFund", err)
	}

	in.Tracking, in.Amount, in.Units = FundBalance, mustDecimal(t, "1000"), mustDecimal(t, "0")

	if _, err := svc.CreateFund(context.Background(), uuid.New(), in); !errors.Is(err, ErrFundNotLinkable) {
		t.Errorf("a linked fund followed by balance = %v, want ErrFundNotLinkable", err)
	}
}

func TestImportPublicFundValues(t *testing.T) {
	alice, bob, carol := uuid.New(), uuid.New(), uuid.New()
	other := "85-27-98186-1-502"

	feed := &fakePublicFunds{values: []sfc.UnitValue{
		{Date: sfcDay("2026-09-20"), Value: "100"},
		{Date: sfcDay("2026-09-21"), Value: "101"},
		{Date: sfcDay("2026-09-22"), Value: "102"},
	}}

	imported := map[uuid.UUID][]string{}

	repo := new(fakeRepository{
		getLinkedFunds: func(context.Context) ([]LinkedFund, error) {
			return []LinkedFund{
				{UserID: alice, PublicFundID: latamFund().ID, Since: sfcDay("2026-09-22")},
				{UserID: bob, PublicFundID: latamFund().ID, Since: sfcDay("2026-09-21")},
				{UserID: carol, PublicFundID: other, Since: sfcDay("2026-09-20")},
			}, nil
		},
		getPublicFund: func(_ context.Context, id string) (PublicFund, error) {
			if id == other {
				return PublicFund{}, ErrPublicFundNotFound
			}

			return latamFund(), nil
		},
		importPublicMarks: func(_ context.Context, userID, _ uuid.UUID, _ string, values []PublicFundValue) (int, error) {
			for _, v := range values {
				imported[userID] = append(imported[userID], v.Date.Format(time.DateOnly))
			}

			return len(values), nil
		},
	})

	written, errs := publicFundService(repo, feed).ImportPublicFundValues(context.Background())

	// One read for the two owners of the same fund, from the earlier of them.
	if !slices.Equal(feed.asked, []string{"5-31-3644-1-501@2026-09-21"}) {
		t.Errorf("asked = %q", feed.asked)
	}

	// Each owner gets what is new to them; the fund missing from the catalog
	// fails alone.
	if written != 3 || len(errs) != 1 || !errors.Is(errs[0], ErrPublicFundNotFound) {
		t.Errorf("written = %d, errs = %v", written, errs)
	}

	if !slices.Equal(imported[alice], []string{"2026-09-22"}) || !slices.Equal(imported[bob], []string{"2026-09-21", "2026-09-22"}) {
		t.Errorf("imported = %v", imported)
	}
}

func TestSearchPublicFundsFillsAnEmptyCatalog(t *testing.T) {
	var (
		upserted []PublicFund
		searched []string
	)

	repo := new(fakeRepository{
		countPublicFunds: func(context.Context) (int, error) { return len(upserted), nil },
		upsertPublicFunds: func(_ context.Context, funds []PublicFund) (int, error) {
			upserted = funds

			return len(funds), nil
		},
		searchPublicFunds: func(_ context.Context, words []string, _ time.Time, limit int) ([]PublicFund, error) {
			searched = words

			return upserted[:min(limit, len(upserted))], nil
		},
	})

	feed := &fakePublicFunds{funds: []sfc.Fund{{Key: latamKey, Name: "LATAM", UnitValue: "1", Date: sfcDay("2026-09-22")}}}

	funds, err := publicFundService(repo, feed).SearchPublicFunds(context.Background(), "latam")
	if err != nil || len(funds) != 1 || funds[0].ID != latamKey.String() || !slices.Equal(searched, []string{"latam"}) {
		t.Fatalf("funds = %+v, %v (searched %q)", funds, err, searched)
	}

	if _, err := publicFundService(repo, feed).SearchPublicFunds(context.Background(), "x"); !errors.Is(err, ErrInvalidPublicFundSearch) {
		t.Errorf("a one-letter search = %v", err)
	}
}

// The routes answer a short search with 400, a fund that cannot take a
// published value with 409, and an SFC that cannot be read with 503.
func TestHandlerPublicFundStatuses(t *testing.T) {
	assetID := uuid.New()
	public := latamFund()
	fund := Fund{AssetID: assetID, Tracking: FundUnits, Currency: money.COP}

	repo := new(fakeRepository{
		getFund:          func(context.Context, uuid.UUID, uuid.UUID) (Fund, error) { return fund, nil },
		getPublicFund:    func(context.Context, string) (PublicFund, error) { return public, nil },
		getFundMovements: func(context.Context, uuid.UUID, uuid.UUID) ([]FundMovement, error) { return nil, nil },
	})

	// The module is built without a feed, which is how an SFC that is down
	// looks to a link.
	app := newTestModule(t, repo, uuid.New(), "user")
	link := `{"publicFundId":"` + public.ID + `"}`

	if resp := doJSON(t, app, http.MethodGet, "/portfolios/funds/catalog?q=x", ""); resp.StatusCode != http.StatusBadRequest {
		t.Errorf("short search: status = %d, want 400", resp.StatusCode)
	}

	if resp := doJSON(t, app, http.MethodPut, "/portfolios/funds/"+assetID.String()+"/link", link); resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("SFC down: status = %d, want 503", resp.StatusCode)
	}

	fund.Tracking = FundBalance

	if resp := doJSON(t, app, http.MethodPut, "/portfolios/funds/"+assetID.String()+"/link", link); resp.StatusCode != http.StatusConflict {
		t.Errorf("balance fund: status = %d, want 409", resp.StatusCode)
	}
}
