package market

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"uuid"

	"github.com/gofiber/fiber/v3"
	"github.com/yeferson59/gofinance/v2/money"
)

// The catalog has two creation paths and the difference between them is the
// point: the operator's overwrites, the user's does not. These tests pin that
// difference, since a user reaching UpsertAsset would let anybody rewrite the
// metadata of an asset every other user is holding.

func TestContributeAsset(t *testing.T) {
	userID := uuid.New()

	t.Run("goes to CreateAssetIfAbsent, never to UpsertAsset", func(t *testing.T) {
		var gotUserID uuid.UUID
		var gotTicker, gotName string
		var gotCurrency money.Currency

		repo := new(fakeRepository{
			upsertAsset: func(context.Context, AssetSpec) (Asset, error) {
				t.Fatal("a contribution reached UpsertAsset, which overwrites rows other users hold")

				return Asset{}, nil
			},
			createAssetIfAbsent: func(_ context.Context, uid uuid.UUID, ticker, name string, _ AssetType, _ string, currency money.Currency) (Asset, error) {
				gotUserID, gotTicker, gotName, gotCurrency = uid, ticker, name, currency

				return Asset{Ticker: ticker}, nil
			},
		})

		svc := newTestServices(repo, newMemStorage())

		if _, err := svc.ContributeAsset(context.Background(), userID, "  ecopetrol  ", " Ecopetrol S.A. ", Stock, "BVC", money.COP); err != nil {
			t.Fatalf("ContributeAsset: %v", err)
		}

		if gotUserID != userID {
			t.Errorf("userID = %s, want %s", gotUserID, userID)
		}
		if gotTicker != "ECOPETROL" {
			t.Errorf("ticker = %q, want %q", gotTicker, "ECOPETROL")
		}
		if gotName != "Ecopetrol S.A." {
			t.Errorf("name = %q, want it trimmed", gotName)
		}
		if gotCurrency != money.COP {
			t.Errorf("currency = %q, want %q", gotCurrency, "COP")
		}
	})

	t.Run("an absent name falls back to the ticker", func(t *testing.T) {
		var gotName string
		repo := new(fakeRepository{
			createAssetIfAbsent: func(_ context.Context, _ uuid.UUID, _, name string, _ AssetType, _ string, _ money.Currency) (Asset, error) {
				gotName = name

				return Asset{}, nil
			},
		})

		if _, err := newTestServices(repo, newMemStorage()).ContributeAsset(context.Background(), userID, "GEB", "", Stock, "", money.COP); err != nil {
			t.Fatalf("ContributeAsset: %v", err)
		}

		if gotName != "GEB" {
			t.Errorf("name = %q, want the ticker", gotName)
		}
	})

	t.Run("rejects bad input before touching the repository", func(t *testing.T) {
		cases := []struct {
			name     string
			ticker   string
			asset    AssetType
			currency money.Currency
			want     error
		}{
			{"empty ticker", "   ", Stock, money.USD, errAssetTickerRequired},
			{"overlong ticker", strings.Repeat("A", maxTickerLen+1), Stock, money.USD, errAssetTickerTooLong},
			{"unknown type", "AAPL", AssetType("nft"), money.USD, errAssetTypeInvalid},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				repo := new(fakeRepository{
					createAssetIfAbsent: func(context.Context, uuid.UUID, string, string, AssetType, string, money.Currency) (Asset, error) {
						t.Fatal("invalid input reached the repository")

						return Asset{}, nil
					},
				})

				_, err := newTestServices(repo, newMemStorage()).ContributeAsset(context.Background(), userID, tc.ticker, "Name", tc.asset, "", tc.currency)
				if !errors.Is(err, tc.want) {
					t.Errorf("err = %v, want %v", err, tc.want)
				}
			})
		}
	})

	t.Run("stops at the daily quota", func(t *testing.T) {
		var since time.Time
		repo := new(fakeRepository{
			countContributed: func(_ context.Context, uid uuid.UUID, from time.Time) (int, error) {
				if uid != userID {
					t.Errorf("counted for %s, want %s", uid, userID)
				}
				since = from

				return maxContributedAssetsPerDay, nil
			},
			createAssetIfAbsent: func(context.Context, uuid.UUID, string, string, AssetType, string, money.Currency) (Asset, error) {
				t.Fatal("the quota was exceeded and the asset was created anyway")

				return Asset{}, nil
			},
		})

		_, err := newTestServices(repo, newMemStorage()).ContributeAsset(context.Background(), userID, "AAPL", "Apple", Stock, "", money.USD)
		if !errors.Is(err, ErrAssetQuotaExceeded) {
			t.Fatalf("err = %v, want ErrAssetQuotaExceeded", err)
		}

		// A rolling window, not a calendar day: a user who hits the cap at 23:00
		// should not get a fresh allowance an hour later.
		if elapsed := time.Since(since); elapsed < 23*time.Hour || elapsed > 25*time.Hour {
			t.Errorf("counted from %s, want roughly 24h ago", since)
		}
	})
}

// newAssetApp mounts the market routes acting as a caller with the given role.
func newAssetApp(t *testing.T, repo Repository, userID uuid.UUID, role string) *fiber.App {
	t.Helper()

	noopLimiter := func(c fiber.Ctx) error { return c.Next() }
	mod := New(Deps{
		Service:           newTestServices(repo, newMemStorage()),
		AuthMiddleware:    stubAuth{userID: userID, role: role},
		Limiter:           noopLimiter,
		CredentialLimiter: noopLimiter,
	})

	app := fiber.New()
	mod.Routes(app)

	return app
}

func TestHandlerCreateAsset(t *testing.T) {
	userID := uuid.New()
	body := `{"ticker":"geb","name":"Grupo Energía Bogotá","assetType":"stock","currency":"cop"}`

	t.Run("a user contributes rather than curates", func(t *testing.T) {
		var contributed bool
		repo := new(fakeRepository{
			upsertAsset: func(context.Context, AssetSpec) (Asset, error) {
				t.Fatal("a non-admin reached the curating path")

				return Asset{}, nil
			},
			createAssetIfAbsent: func(_ context.Context, uid uuid.UUID, ticker, _ string, _ AssetType, _ string, _ money.Currency) (Asset, error) {
				contributed = true

				if uid != userID {
					t.Errorf("userID = %s, want the caller %s", uid, userID)
				}

				return Asset{ID: uuid.New(), Ticker: ticker}, nil
			},
		})

		resp := request(t, newAssetApp(t, repo, userID, "user"), http.MethodPost, "/assets", body)
		if resp.StatusCode != fiber.StatusCreated {
			t.Fatalf("status = %d, want 201", resp.StatusCode)
		}
		if !contributed {
			t.Error("the request never reached CreateAssetIfAbsent")
		}
	})

	t.Run("an admin curates", func(t *testing.T) {
		var curated bool
		repo := new(fakeRepository{
			upsertAsset: func(_ context.Context, spec AssetSpec) (Asset, error) {
				curated = true

				return Asset{ID: uuid.New(), Ticker: spec.Ticker, IsCurated: true}, nil
			},
			createAssetIfAbsent: func(context.Context, uuid.UUID, string, string, AssetType, string, money.Currency) (Asset, error) {
				t.Fatal("an admin was routed through the contribution path")

				return Asset{}, nil
			},
		})

		resp := request(t, newAssetApp(t, repo, userID, "admin"), http.MethodPost, "/assets", body)
		if resp.StatusCode != fiber.StatusCreated {
			t.Fatalf("status = %d, want 201", resp.StatusCode)
		}
		if !curated {
			t.Error("the admin request never reached UpsertAsset")
		}
	})

	// The admin form posts the sector as free text — the select sends the
	// canonical value, a re-import can send anything — so what the handler does
	// with that string is the contract the form depends on.
	t.Run("the sector arrives normalised from whatever the form sent", func(t *testing.T) {
		for _, spelling := range []string{"technology", "Tecnología", "Technology"} {
			var got Sector
			repo := new(fakeRepository{
				upsertAsset: func(_ context.Context, spec AssetSpec) (Asset, error) {
					got = spec.Sector

					return Asset{ID: uuid.New(), Ticker: spec.Ticker}, nil
				},
			})

			withSector := `{"ticker":"aapl","name":"Apple","assetType":"stock","currency":"usd","sector":"` + spelling + `"}`

			resp := request(t, newAssetApp(t, repo, userID, "admin"), http.MethodPost, "/assets", withSector)
			if resp.StatusCode != fiber.StatusCreated {
				t.Fatalf("%q: status = %d, want 201", spelling, resp.StatusCode)
			}
			if got != SectorTechnology {
				t.Errorf("%q reached the service as %q, want technology", spelling, got)
			}
		}
	})

	// A label nobody recognises is refused rather than dropped: the form said
	// something about this asset, and creating the row without it would report
	// a success for a field that never landed.
	t.Run("an unknown sector is a 400 and writes nothing", func(t *testing.T) {
		repo := new(fakeRepository{
			upsertAsset: func(context.Context, AssetSpec) (Asset, error) {
				t.Fatal("a row was written despite an unusable sector")

				return Asset{}, nil
			},
		})

		bad := `{"ticker":"aapl","name":"Apple","assetType":"stock","currency":"usd","sector":"ganaderia"}`

		resp := request(t, newAssetApp(t, repo, userID, "admin"), http.MethodPost, "/assets", bad)
		if resp.StatusCode != fiber.StatusBadRequest {
			t.Fatalf("status = %d, want 400", resp.StatusCode)
		}
	})

	// The form hides the select for a type that cannot carry one, so an omitted
	// sector is the ordinary path for half the catalog and must not fail.
	t.Run("an omitted sector creates the asset unclassified", func(t *testing.T) {
		var got Sector
		var called bool
		repo := new(fakeRepository{
			upsertAsset: func(_ context.Context, spec AssetSpec) (Asset, error) {
				got, called = spec.Sector, true

				return Asset{ID: uuid.New(), Ticker: spec.Ticker}, nil
			},
		})

		resp := request(t, newAssetApp(t, repo, userID, "admin"), http.MethodPost, "/assets", body)
		if resp.StatusCode != fiber.StatusCreated {
			t.Fatalf("status = %d, want 201", resp.StatusCode)
		}
		if !called || got != SectorNone {
			t.Errorf("sector = %q (reached=%t), want none", got, called)
		}
	})

	// The whole reason the breakdown exists: an ETF that is every industry at
	// once had two bad answers before it, and both of them ended up on a chart.
	t.Run("a fund posts its breakdown instead of a sector", func(t *testing.T) {
		var got AssetSpec
		repo := new(fakeRepository{
			upsertAsset: func(_ context.Context, spec AssetSpec) (Asset, error) {
				got = spec

				return Asset{ID: uuid.New(), Ticker: spec.Ticker}, nil
			},
		})

		// The two JSON shapes a client may send a decimal in — quoted and bare
		// — and a label in the user's own language, since the breakdown goes
		// through the same normaliser the single sector does.
		fund := `{"ticker":"voo","name":"Vanguard S&P 500","assetType":"etf","currency":"usd",` +
			`"sectorWeights":[{"sector":"Tecnología","weight":"33.1"},{"sector":"financials","weight":13.8}]}`

		resp := request(t, newAssetApp(t, repo, userID, "admin"), http.MethodPost, "/assets", fund)
		if resp.StatusCode != fiber.StatusCreated {
			t.Fatalf("status = %d, want 201", resp.StatusCode)
		}

		if len(got.SectorWeights) != 2 {
			t.Fatalf("weights = %+v, want two rows", got.SectorWeights)
		}
		if got.SectorWeights[0].Sector != SectorTechnology || got.SectorWeights[0].Weight.String() != "33.1" {
			t.Errorf("first weight = %+v, want technology 33.1", got.SectorWeights[0])
		}
		if got.SectorWeights[1].Sector != SectorFinancials || got.SectorWeights[1].Weight.String() != "13.8" {
			t.Errorf("second weight = %+v, want financials 13.8", got.SectorWeights[1])
		}
		if got.Sector != SectorNone {
			t.Errorf("sector = %q, want none beside a breakdown", got.Sector)
		}
	})

	// Arithmetic the form should have caught still has to be refused here: the
	// API is reachable without the form, and a breakdown adding to 160 %% would
	// hand a chart shares of more money than the user has.
	t.Run("a breakdown that cannot be true is a 400 and writes nothing", func(t *testing.T) {
		for _, bad := range []string{
			`"sectorWeights":[{"sector":"technology","weight":80},{"sector":"energy","weight":80}]`,
			`"sectorWeights":[{"sector":"technology","weight":50},{"sector":"technology","weight":20}]`,
			`"sectorWeights":[{"sector":"ganaderia","weight":50}]`,
			`"sector":"technology","sectorWeights":[{"sector":"energy","weight":50}]`,
		} {
			repo := new(fakeRepository{
				upsertAsset: func(context.Context, AssetSpec) (Asset, error) {
					t.Fatalf("a row was written despite %s", bad)

					return Asset{}, nil
				},
			})

			body := `{"ticker":"voo","name":"Vanguard","assetType":"etf","currency":"usd",` + bad + `}`

			resp := request(t, newAssetApp(t, repo, userID, "admin"), http.MethodPost, "/assets", body)
			if resp.StatusCode != fiber.StatusBadRequest {
				t.Errorf("%s: status = %d, want 400", bad, resp.StatusCode)
			}
		}
	})

	t.Run("the quota answers 429", func(t *testing.T) {
		repo := new(fakeRepository{
			countContributed: func(context.Context, uuid.UUID, time.Time) (int, error) {
				return maxContributedAssetsPerDay, nil
			},
		})

		resp := request(t, newAssetApp(t, repo, userID, "user"), http.MethodPost, "/assets", body)
		if resp.StatusCode != fiber.StatusTooManyRequests {
			t.Fatalf("status = %d, want 429", resp.StatusCode)
		}
	})

	t.Run("a rejected input says why, an internal failure does not", func(t *testing.T) {
		repo := new(fakeRepository{
			createAssetIfAbsent: func(context.Context, uuid.UUID, string, string, AssetType, string, money.Currency) (Asset, error) {
				return Asset{}, errors.New(`pq: duplicate key value violates unique constraint "idx_assets_ticker_exchange"`)
			},
		})
		app := newAssetApp(t, repo, userID, "user")

		bad := request(t, app, http.MethodPost, "/assets", `{"ticker":"AAPL","name":"Apple","assetType":"nft","currency":"USD"}`)
		if bad.StatusCode != fiber.StatusBadRequest {
			t.Fatalf("status = %d, want 400", bad.StatusCode)
		}
		if action := actionOf(t, bad); !strings.Contains(action, "tipo de activo") {
			t.Errorf("action = %q, want it to name the offending field", action)
		}

		boom := request(t, app, http.MethodPost, "/assets", body)
		if boom.StatusCode != fiber.StatusInternalServerError {
			t.Fatalf("status = %d, want 500", boom.StatusCode)
		}
		if action := actionOf(t, boom); strings.Contains(action, "constraint") {
			t.Errorf("action = %q, want the schema kept out of the response", action)
		}
	})
}

// actionOf reads the "action" field of an error envelope, which is where
// FromDomain puts the detail a client shows the user.
func actionOf(t *testing.T, resp *http.Response) string {
	t.Helper()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var envelope struct {
		Action string `json:"action"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		t.Fatalf("decode body %q: %v", raw, err)
	}

	return envelope.Action
}

// Editing a catalog row is the third write to this table, and the only one that
// can reach a row by id: it renames, re-denominates and un-publishes assets
// other users may be holding. These tests pin the two things that makes
// dangerous — the guard on the route, and what happens to a manual price when
// the currency underneath it changes.

func TestUpdateAsset(t *testing.T) {
	assetID := uuid.New()

	t.Run("normalizes the same way a create does", func(t *testing.T) {
		var got AssetUpdate

		repo := new(fakeRepository{
			updateAsset: func(_ context.Context, id uuid.UUID, upd AssetUpdate) (Asset, error) {
				if id != assetID {
					t.Errorf("assetID = %s, want %s", id, assetID)
				}
				got = upd

				return Asset{ID: id, Ticker: upd.Ticker}, nil
			},
		})

		_, err := newTestServices(repo, newMemStorage()).UpdateAsset(context.Background(), assetID, AssetUpdate{
			Ticker:    "  ecopetrol ",
			Name:      "  Ecopetrol S.A. ",
			AssetType: Stock,
			Exchange:  " BVC ",
			Currency:  money.COP,
		})
		if err != nil {
			t.Fatalf("UpdateAsset: %v", err)
		}

		if got.Ticker != "ECOPETROL" {
			t.Errorf("ticker = %q, want %q", got.Ticker, "ECOPETROL")
		}
		if got.Name != "Ecopetrol S.A." {
			t.Errorf("name = %q, want it trimmed", got.Name)
		}
		if got.Exchange != "BVC" {
			t.Errorf("exchange = %q, want it trimmed", got.Exchange)
		}
		// Neither was sent, and neither may be invented: a nil price keeps the
		// stored one, a nil flag keeps the audience.
		if got.Price != nil || got.IsCurated != nil {
			t.Errorf("price = %v, isCurated = %v, want both nil", got.Price, got.IsCurated)
		}
	})

	t.Run("rejects bad input before touching the repository", func(t *testing.T) {
		zero := mustMoney(t, "0", money.USD)
		negative := mustMoney(t, "-5", money.USD)

		cases := []struct {
			name string
			upd  AssetUpdate
			want error
		}{
			{"empty ticker", AssetUpdate{Ticker: "  ", AssetType: Stock, Currency: money.USD}, errAssetTickerRequired},
			{"unknown type", AssetUpdate{Ticker: "AAPL", AssetType: AssetType("nft"), Currency: money.USD}, errAssetTypeInvalid},
			{"invalid currency", AssetUpdate{Ticker: "AAPL", AssetType: Stock, Currency: money.Currency(200)}, errAssetCurrencyInvalid},
			{"zero price", AssetUpdate{Ticker: "AAPL", AssetType: Stock, Currency: money.USD, Price: &zero}, errAssetPriceInvalid},
			{"negative price", AssetUpdate{Ticker: "AAPL", AssetType: Stock, Currency: money.USD, Price: &negative}, errAssetPriceInvalid},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				repo := new(fakeRepository{
					updateAsset: func(context.Context, uuid.UUID, AssetUpdate) (Asset, error) {
						t.Fatal("invalid input reached the repository")

						return Asset{}, nil
					},
				})

				_, err := newTestServices(repo, newMemStorage()).UpdateAsset(context.Background(), assetID, tc.upd)
				if !errors.Is(err, tc.want) {
					t.Errorf("err = %v, want %v", err, tc.want)
				}
			})
		}
	})
}

func TestHandlerUpdateAsset(t *testing.T) {
	assetID := uuid.New()
	body := `{"ticker":"geb","name":"Grupo Energía Bogotá","assetType":"stock","currency":"COP","exchange":"BVC","isCurated":false,"price":{"value":"2450.50","currency":"COP"}}`

	t.Run("a non-admin cannot edit the catalog", func(t *testing.T) {
		repo := new(fakeRepository{
			updateAsset: func(context.Context, uuid.UUID, AssetUpdate) (Asset, error) {
				t.Fatal("a non-admin reached the edit path")

				return Asset{}, nil
			},
		})

		resp := request(t, newAssetApp(t, repo, uuid.New(), "user"), http.MethodPatch, "/assets/"+assetID.String(), body)
		if resp.StatusCode != fiber.StatusForbidden {
			t.Fatalf("status = %d, want 403", resp.StatusCode)
		}
	})

	t.Run("an admin's edit carries every field, including the ones whose zero value means something", func(t *testing.T) {
		var got AssetUpdate

		repo := new(fakeRepository{
			updateAsset: func(_ context.Context, id uuid.UUID, upd AssetUpdate) (Asset, error) {
				if id != assetID {
					t.Errorf("assetID = %s, want %s", id, assetID)
				}
				got = upd

				return Asset{ID: id, Ticker: upd.Ticker}, nil
			},
		})

		resp := request(t, newAssetApp(t, repo, uuid.New(), "admin"), http.MethodPatch, "/assets/"+assetID.String(), body)
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}

		if got.Ticker != "GEB" || got.Currency != money.COP || got.Exchange != "BVC" {
			t.Errorf("update = %+v, want the body's fields normalized", got)
		}
		// The one a plain bool would have swallowed: un-curating is what takes
		// a row off the shared catalog, and it is spelled `false`.
		if got.IsCurated == nil || *got.IsCurated {
			t.Errorf("isCurated = %v, want an explicit false", got.IsCurated)
		}
		if got.Price == nil || got.Price.String() != "2450.5" {
			t.Errorf("price = %v, want 2450.50", got.Price)
		}
	})

	t.Run("a rename onto an existing ticker answers 409", func(t *testing.T) {
		repo := new(fakeRepository{
			updateAsset: func(context.Context, uuid.UUID, AssetUpdate) (Asset, error) {
				return Asset{}, errAssetDuplicate
			},
		})

		resp := request(t, newAssetApp(t, repo, uuid.New(), "admin"), http.MethodPatch, "/assets/"+assetID.String(), body)
		if resp.StatusCode != fiber.StatusConflict {
			t.Fatalf("status = %d, want 409", resp.StatusCode)
		}
		if action := actionOf(t, resp); !strings.Contains(action, "ya existe") {
			t.Errorf("action = %q, want it to say the ticker is taken", action)
		}
	})

	t.Run("an internal failure keeps the schema out of the response", func(t *testing.T) {
		repo := new(fakeRepository{
			updateAsset: func(context.Context, uuid.UUID, AssetUpdate) (Asset, error) {
				return Asset{}, errors.New(`pq: null value in column "currency" violates not-null constraint`)
			},
		})

		resp := request(t, newAssetApp(t, repo, uuid.New(), "admin"), http.MethodPatch, "/assets/"+assetID.String(), body)
		if resp.StatusCode != fiber.StatusInternalServerError {
			t.Fatalf("status = %d, want 500", resp.StatusCode)
		}
		if action := actionOf(t, resp); strings.Contains(action, "constraint") {
			t.Errorf("action = %q, want the schema kept out of the response", action)
		}
	})
}

// mustMoney builds an amount from the text a request body would carry, which is
// also how the price reaches the service.
func mustMoney(t *testing.T, value string, currency money.Currency) money.Money {
	t.Helper()

	m, err := money.NewMoneyFromString(value, currency)
	if err != nil {
		t.Fatalf("money %q: %v", value, err)
	}

	return m
}
