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
)

// The catalog has two creation paths and the difference between them is the
// point: the operator's overwrites, the user's does not. These tests pin that
// difference, since a user reaching UpsertAsset would let anybody rewrite the
// metadata of an asset every other user is holding.

func TestContributeAsset(t *testing.T) {
	userID := uuid.New()

	t.Run("goes to CreateAssetIfAbsent, never to UpsertAsset", func(t *testing.T) {
		var gotUserID uuid.UUID
		var gotTicker, gotName, gotCurrency string

		repo := new(fakeRepository{
			upsertAsset: func(context.Context, string, string, AssetType, string, string) (Asset, error) {
				t.Fatal("a contribution reached UpsertAsset, which overwrites rows other users hold")

				return Asset{}, nil
			},
			createAssetIfAbsent: func(_ context.Context, uid uuid.UUID, ticker, name string, _ AssetType, _, currency string) (Asset, error) {
				gotUserID, gotTicker, gotName, gotCurrency = uid, ticker, name, currency

				return Asset{Ticker: ticker}, nil
			},
		})

		svc := newTestServices(repo, newMemStorage())

		if _, err := svc.ContributeAsset(context.Background(), userID, "  ecopetrol  ", " Ecopetrol S.A. ", Stock, "BVC", "cop"); err != nil {
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
		if gotCurrency != "COP" {
			t.Errorf("currency = %q, want %q", gotCurrency, "COP")
		}
	})

	t.Run("an absent name falls back to the ticker", func(t *testing.T) {
		var gotName string
		repo := new(fakeRepository{
			createAssetIfAbsent: func(_ context.Context, _ uuid.UUID, _, name string, _ AssetType, _, _ string) (Asset, error) {
				gotName = name

				return Asset{}, nil
			},
		})

		if _, err := newTestServices(repo, newMemStorage()).ContributeAsset(context.Background(), userID, "GEB", "", Stock, "", "COP"); err != nil {
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
			currency string
			want     error
		}{
			{"empty ticker", "   ", Stock, "USD", errAssetTickerRequired},
			{"overlong ticker", strings.Repeat("A", maxTickerLen+1), Stock, "USD", errAssetTickerTooLong},
			{"unknown type", "AAPL", AssetType("nft"), "USD", errAssetTypeInvalid},
			{"malformed currency", "AAPL", Stock, "DOLAR", errAssetCurrencyInvalid},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				repo := new(fakeRepository{
					createAssetIfAbsent: func(context.Context, uuid.UUID, string, string, AssetType, string, string) (Asset, error) {
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
			createAssetIfAbsent: func(context.Context, uuid.UUID, string, string, AssetType, string, string) (Asset, error) {
				t.Fatal("the quota was exceeded and the asset was created anyway")

				return Asset{}, nil
			},
		})

		_, err := newTestServices(repo, newMemStorage()).ContributeAsset(context.Background(), userID, "AAPL", "Apple", Stock, "", "USD")
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
			upsertAsset: func(context.Context, string, string, AssetType, string, string) (Asset, error) {
				t.Fatal("a non-admin reached the curating path")

				return Asset{}, nil
			},
			createAssetIfAbsent: func(_ context.Context, uid uuid.UUID, ticker, _ string, _ AssetType, _, _ string) (Asset, error) {
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
			upsertAsset: func(_ context.Context, ticker, _ string, _ AssetType, _, _ string) (Asset, error) {
				curated = true

				return Asset{ID: uuid.New(), Ticker: ticker, IsCurated: true}, nil
			},
			createAssetIfAbsent: func(context.Context, uuid.UUID, string, string, AssetType, string, string) (Asset, error) {
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
			createAssetIfAbsent: func(context.Context, uuid.UUID, string, string, AssetType, string, string) (Asset, error) {
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
