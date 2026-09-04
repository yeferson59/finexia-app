package auth

import (
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"uuid"

	"github.com/gofiber/fiber/v3"

	"github.com/yeferson59/finexia-app/internal/identity"
	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

// mcpTokenTestApp mounts the whole module and returns an app plus an access
// token its RequireAuth gate accepts, so these tests traverse the chain a
// browser does — guard, limiter, handler — instead of calling the handler
// directly. The user id the routes will act for is returned too, since every
// store call must be scoped to it.
func mcpTokenTestApp(t *testing.T, repo *fakeRepository) (*fiber.App, string, uuid.UUID) {
	t.Helper()

	cfg := testConfig()
	userID := uuid.New()

	service := newService(testStores(repo), cfg, newMemStorage(), nil, nil, logger.Noop())

	token, err := service.CreateJWToken(userID, "user", time.Now().UTC().Add(time.Hour))
	if err != nil {
		t.Fatalf("CreateJWToken: %v", err)
	}

	// The gate validates the bearer token against the session store, so the
	// fake has to answer for it before any of these routes is reachable.
	repo.getSessionByToken = func(context.Context, string) (identity.User, error) {
		return identity.User{
			ID:       userID,
			Role:     identity.Role{Name: "user"},
			Sessions: []identity.Session{{Token: token}},
		}, nil
	}

	m := newModule(Deps{Cfg: cfg, Storage: newMemStorage(), Log: logger.Noop()}, service)

	app := fiber.New()
	m.Routes(app)

	return app, token, userID
}

// callMCP performs one authenticated request against the token endpoints and
// returns the status with the decoded envelope.
func callMCP(t *testing.T, app *fiber.App, accessToken, method, target, body string) (int, map[string]any) {
	t.Helper()

	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}

	req := httptest.NewRequest(method, target, reader)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, target, err)
	}
	defer func() { _ = resp.Body.Close() }()

	raw, _ := io.ReadAll(resp.Body)

	var payload map[string]any
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &payload); err != nil {
			t.Fatalf("invalid JSON response %q: %v", raw, err)
		}
	}

	return resp.StatusCode, payload
}

// TestMCPTokenEndpointServesTheSecretOnce pins what the whole design rests on:
// creation answers with the secret, and the listing — the only other way to
// reach a token — has no field that could carry it.
func TestMCPTokenEndpointServesTheSecretOnce(t *testing.T) {
	var minted MCPToken

	repo := new(fakeRepository{
		countMCPTokens: func(context.Context, uuid.UUID) (int, error) { return 0, nil },
		createMCPToken: func(_ context.Context, userID uuid.UUID, name, _, last4 string, expiresAt *time.Time) (MCPToken, error) {
			minted = MCPToken{
				ID: uuid.New(), Name: name, Last4: last4,
				ExpiresAt: expiresAt, CreatedAt: time.Now().UTC(),
			}

			return minted, nil
		},
	})
	repo.listMCPTokens = func(context.Context, uuid.UUID) ([]MCPToken, error) {
		return []MCPToken{minted}, nil
	}

	app, accessToken, _ := mcpTokenTestApp(t, repo)

	status, payload := callMCP(t, app, accessToken, "POST", "/auth/mcp-tokens", `{"name":"Claude Desktop"}`)
	if status != fiber.StatusCreated {
		t.Fatalf("create status = %d, want 201: %v", status, payload)
	}

	data, _ := payload["data"].(map[string]any)

	secret, _ := data["token"].(string)
	if !strings.HasPrefix(secret, mcpTokenPrefix) {
		t.Fatalf("created token = %q, want the %q prefix", secret, mcpTokenPrefix)
	}

	status, payload = callMCP(t, app, accessToken, "GET", "/auth/mcp-tokens", "")
	if status != fiber.StatusOK {
		t.Fatalf("list status = %d, want 200: %v", status, payload)
	}

	listed, _ := json.Marshal(payload["data"])
	if strings.Contains(string(listed), secret) {
		t.Fatal("the listing served the secret back")
	}

	if !strings.Contains(string(listed), minted.Last4) {
		t.Errorf("the listing does not show last4, which is how a token is recognised: %s", listed)
	}
}

// TestMCPTokenEndpointScopesEveryCallToTheCaller: no route takes a user id, so
// the only one that can reach the store is the authenticated one.
func TestMCPTokenEndpointScopesEveryCallToTheCaller(t *testing.T) {
	var sawCreate, sawList, sawDelete uuid.UUID

	tokenID := uuid.New()
	repo := new(fakeRepository{
		countMCPTokens: func(context.Context, uuid.UUID) (int, error) { return 0, nil },
		createMCPToken: func(_ context.Context, userID uuid.UUID, name, _, last4 string, expiresAt *time.Time) (MCPToken, error) {
			sawCreate = userID

			return MCPToken{ID: tokenID, Name: name, Last4: last4, ExpiresAt: expiresAt}, nil
		},
		listMCPTokens: func(_ context.Context, userID uuid.UUID) ([]MCPToken, error) {
			sawList = userID

			return nil, nil
		},
		deleteMCPToken: func(_ context.Context, userID, _ uuid.UUID) error {
			sawDelete = userID

			return nil
		},
	})

	app, accessToken, userID := mcpTokenTestApp(t, repo)

	callMCP(t, app, accessToken, "POST", "/auth/mcp-tokens", `{"name":"laptop"}`)
	callMCP(t, app, accessToken, "GET", "/auth/mcp-tokens", "")
	callMCP(t, app, accessToken, "DELETE", "/auth/mcp-tokens/"+tokenID.String(), "")

	if sawCreate != userID || sawList != userID || sawDelete != userID {
		t.Errorf("store saw create=%s list=%s delete=%s, want the authenticated %s", sawCreate, sawList, sawDelete, userID)
	}
}

// TestMCPTokenEndpointStatuses maps each refusal to the status a client acts
// on: a bad field is worth correcting, a taken name is worth renaming, and an
// unknown id is worth forgetting.
func TestMCPTokenEndpointStatuses(t *testing.T) {
	for _, tc := range []struct {
		name       string
		method     string
		target     string
		body       string
		repo       func(*fakeRepository)
		wantStatus int
	}{
		{
			name: "a lifetime past the maximum is a 400", method: "POST", target: "/auth/mcp-tokens",
			body: `{"name":"too long","expiresInDays":400}`,
			repo: func(f *fakeRepository) {
				f.countMCPTokens = func(context.Context, uuid.UUID) (int, error) { return 0, nil }
			},
			wantStatus: fiber.StatusBadRequest,
		},
		{
			name: "a blank name is a 400", method: "POST", target: "/auth/mcp-tokens",
			body: `{"name":"   "}`,
			repo: func(f *fakeRepository) {
				f.countMCPTokens = func(context.Context, uuid.UUID) (int, error) { return 0, nil }
			},
			wantStatus: fiber.StatusBadRequest,
		},
		{
			name: "a name already in use is a 409", method: "POST", target: "/auth/mcp-tokens",
			body: `{"name":"Claude Desktop"}`,
			repo: func(f *fakeRepository) {
				f.countMCPTokens = func(context.Context, uuid.UUID) (int, error) { return 1, nil }
				f.createMCPToken = func(context.Context, uuid.UUID, string, string, string, *time.Time) (MCPToken, error) {
					return MCPToken{}, ErrMCPTokenNameTaken
				}
			},
			wantStatus: fiber.StatusConflict,
		},
		{
			name: "the limit is a 409", method: "POST", target: "/auth/mcp-tokens",
			body: `{"name":"one more"}`,
			repo: func(f *fakeRepository) {
				f.countMCPTokens = func(context.Context, uuid.UUID) (int, error) { return maxMCPTokensPerUser, nil }
			},
			wantStatus: fiber.StatusConflict,
		},
		{
			name: "deleting an unknown token is a 404", method: "DELETE", target: "/auth/mcp-tokens/" + uuid.New().String(),
			repo: func(f *fakeRepository) {
				f.deleteMCPToken = func(context.Context, uuid.UUID, uuid.UUID) error { return ErrMCPTokenNotFound }
			},
			wantStatus: fiber.StatusNotFound,
		},
		{
			name: "rotating an unknown token is a 404", method: "POST", target: "/auth/mcp-tokens/" + uuid.New().String() + "/rotate",
			repo: func(f *fakeRepository) {
				f.rotateMCPToken = func(context.Context, uuid.UUID, uuid.UUID, string, string, *time.Time) (MCPToken, error) {
					return MCPToken{}, ErrMCPTokenNotFound
				}
			},
			wantStatus: fiber.StatusNotFound,
		},
		{
			name: "a malformed id is a 400", method: "DELETE", target: "/auth/mcp-tokens/not-a-uuid",
			repo:       func(*fakeRepository) {},
			wantStatus: fiber.StatusBadRequest,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := new(fakeRepository{})
			tc.repo(repo)

			app, accessToken, _ := mcpTokenTestApp(t, repo)

			status, payload := callMCP(t, app, accessToken, tc.method, tc.target, tc.body)
			if status != tc.wantStatus {
				t.Errorf("status = %d, want %d: %v", status, tc.wantStatus, payload)
			}
		})
	}
}

// TestRotateMCPTokenEndpointWithoutABody: rotating is a one-click action in the
// UI, so an absent body means "keep the default lifetime" rather than a
// malformed request.
func TestRotateMCPTokenEndpointWithoutABody(t *testing.T) {
	tokenID := uuid.New()

	var gotExpiry *time.Time

	repo := new(fakeRepository{
		rotateMCPToken: func(_ context.Context, _, id uuid.UUID, _, last4 string, expiresAt *time.Time) (MCPToken, error) {
			gotExpiry = expiresAt

			return MCPToken{ID: id, Name: "Claude Desktop", Last4: last4, ExpiresAt: expiresAt}, nil
		},
	})

	app, accessToken, _ := mcpTokenTestApp(t, repo)

	status, payload := callMCP(t, app, accessToken, "POST", "/auth/mcp-tokens/"+tokenID.String()+"/rotate", "")
	if status != fiber.StatusOK {
		t.Fatalf("status = %d, want 200: %v", status, payload)
	}

	if gotExpiry == nil {
		t.Fatal("an absent body produced a token that never expires; the default lifetime should apply")
	}

	data, _ := payload["data"].(map[string]any)
	if secret, _ := data["token"].(string); !strings.HasPrefix(secret, mcpTokenPrefix) {
		t.Errorf("rotation did not return a new secret: %v", data)
	}
}

// TestMCPTokenEndpointsRequireASession: they mint and destroy credentials, so
// they sit behind the same gate as everything else under /auth.
//
// Both failures answer 401, which is the guard's own answer rather than
// jwtware's default: a missing Authorization header is an unauthenticated
// request, not a malformed one. What this test is for is the part that matters
// either way — the store is never reached without a session.
func TestMCPTokenEndpointsRequireASession(t *testing.T) {
	for _, tc := range []struct {
		name       string
		bearer     string
		wantStatus int
	}{
		{name: "no header", bearer: "", wantStatus: fiber.StatusUnauthorized},
		{name: "not a session", bearer: "Bearer nonsense", wantStatus: fiber.StatusUnauthorized},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Every hook nil: any store call panics rather than answering.
			app, _, _ := mcpTokenTestApp(t, new(fakeRepository{}))

			req := httptest.NewRequest("GET", "/auth/mcp-tokens", nil)
			if tc.bearer != "" {
				req.Header.Set("Authorization", tc.bearer)
			}

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("app.Test: %v", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != tc.wantStatus {
				t.Errorf("status = %d, want %d", resp.StatusCode, tc.wantStatus)
			}
		})
	}
}
