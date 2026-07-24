package marketing

import (
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func newTestApp(repo Repository, mail Mailer) *fiber.App {
	app := fiber.New()
	New(repo, mail).Routes(app)
	return app
}

func postWaitlist(t *testing.T, app *fiber.App, body string) (int, map[string]any) {
	t.Helper()
	req := httptest.NewRequest("POST", "/marketing/waitlists", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	raw, _ := io.ReadAll(resp.Body)
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("invalid JSON response %q: %v", raw, err)
	}
	return resp.StatusCode, payload
}

func TestCreateWaitlistRoute(t *testing.T) {
	t.Run("registers the email and returns the envelope", func(t *testing.T) {
		var saved string
		repo := &fakeRepository{
			saveWaitlistEmail: func(_ context.Context, email string) error {
				saved = email
				return nil
			},
		}
		mailer := &fakeMailer{}
		app := newTestApp(repo, mailer)

		status, payload := postWaitlist(t, app, `{"email":"new@example.com"}`)
		if status != fiber.StatusOK {
			t.Fatalf("status = %d, want 200", status)
		}
		if success, _ := payload["success"].(bool); !success {
			t.Error("success should be true")
		}
		data, _ := payload["data"].(map[string]any)
		if data["email"] != "new@example.com" {
			t.Errorf("data.email = %v", data["email"])
		}
		if saved != "new@example.com" {
			t.Errorf("saved = %q", saved)
		}
		if len(mailer.waitlistTo) != 1 {
			t.Errorf("confirmations = %v", mailer.waitlistTo)
		}
	})

	t.Run("maps a duplicate email to 409", func(t *testing.T) {
		repo := &fakeRepository{
			// The repository translates the DB unique violation into this
			// tagged sentinel; the fake stands in for that contract.
			saveWaitlistEmail: func(context.Context, string) error {
				return ErrWaitlistEmailExists
			},
		}
		app := newTestApp(repo, &fakeMailer{})

		status, payload := postWaitlist(t, app, `{"email":"dup@example.com"}`)
		if status != fiber.StatusConflict {
			t.Fatalf("status = %d, want 409", status)
		}
		if success, _ := payload["success"].(bool); success {
			t.Error("success should be false")
		}
	})

	t.Run("rejects a malformed body", func(t *testing.T) {
		app := newTestApp(&fakeRepository{}, &fakeMailer{})

		status, _ := postWaitlist(t, app, `{`)
		if status != fiber.StatusBadRequest {
			t.Fatalf("status = %d, want 400", status)
		}
	})
}

// fakeGuard stands in for *auth.Module: RequireAuth/RequireAdmin either let
// the request through or answer with the status the real guards would.
type fakeGuard struct {
	authStatus  int
	adminStatus int
}

func (g fakeGuard) handler(status int) fiber.Handler {
	return func(c fiber.Ctx) error {
		if status != 0 {
			return c.SendStatus(status)
		}
		return c.Next()
	}
}

func (g fakeGuard) RequireAuth() fiber.Handler  { return g.handler(g.authStatus) }
func (g fakeGuard) RequireAdmin() fiber.Handler { return g.handler(g.adminStatus) }

// TestListWaitlistRoute covers the admin listing this module took over from
// the user module (docs/TECH_DEBT.md #10). The path stays /users/waitlist, so
// the route carries its own guards instead of inheriting the user group's.
func TestListWaitlistRoute(t *testing.T) {
	newApp := func(repo Repository, guard *fakeGuard) *fiber.App {
		app := fiber.New()
		m := New(repo, &fakeMailer{})
		if guard != nil {
			m.SetAdminGuard(*guard, nil)
		}
		m.Routes(app)
		return app
	}

	get := func(t *testing.T, app *fiber.App) (int, map[string]any) {
		t.Helper()
		resp, err := app.Test(httptest.NewRequest("GET", "/users/waitlist?page=1&limit=10", nil))
		if err != nil {
			t.Fatalf("app.Test: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		raw, _ := io.ReadAll(resp.Body)
		var payload map[string]any
		_ = json.Unmarshal(raw, &payload)
		return resp.StatusCode, payload
	}

	t.Run("returns the paginated waitlist", func(t *testing.T) {
		repo := &fakeRepository{
			listWaitlist: func(_ context.Context, offset, limit uint) ([]Waitlist, uint, error) {
				if limit != 10 {
					t.Errorf("limit = %d, want 10", limit)
				}
				if offset != 0 {
					t.Errorf("offset = %d, want 0", offset)
				}
				return []Waitlist{{Email: "early@example.com"}}, 1, nil
			},
		}

		status, payload := get(t, newApp(repo, &fakeGuard{}))
		if status != fiber.StatusOK {
			t.Fatalf("status = %d, want 200", status)
		}
		data, _ := payload["data"].(map[string]any)
		items, _ := data["items"].([]any)
		if len(items) != 1 {
			t.Fatalf("items = %v, want 1 entry", data["items"])
		}
		first, _ := items[0].(map[string]any)
		if first["email"] != "early@example.com" {
			t.Errorf("items[0].email = %v", first["email"])
		}
	})

	t.Run("is gated by the shared guards", func(t *testing.T) {
		if status, _ := get(t, newApp(&fakeRepository{}, &fakeGuard{authStatus: fiber.StatusUnauthorized})); status != fiber.StatusUnauthorized {
			t.Errorf("status = %d, want 401 when RequireAuth rejects", status)
		}
		if status, _ := get(t, newApp(&fakeRepository{}, &fakeGuard{adminStatus: fiber.StatusForbidden})); status != fiber.StatusForbidden {
			t.Errorf("status = %d, want 403 when RequireAdmin rejects", status)
		}
	})

	t.Run("is not registered without a guard", func(t *testing.T) {
		// A missing guard must never expose the listing unprotected.
		if status, _ := get(t, newApp(&fakeRepository{}, nil)); status != fiber.StatusNotFound {
			t.Errorf("status = %d, want 404 when no guard was injected", status)
		}
	})
}
