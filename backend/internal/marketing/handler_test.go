package marketing

import (
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

func newTestApp(repo Repository, mail Mailer) *fiber.App {
	app := fiber.New()
	New(Deps{Service: newService(repo, mail), AuthMiddl: fakeGuard{}}).Routes(app)
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
		repo := new(fakeRepository{
			saveWaitlistEmail: func(_ context.Context, email string) error {
				saved = email
				return nil
			},
		})
		mailer := new(fakeMailer{})
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
		repo := new(fakeRepository{
			// The repository translates the DB unique violation into this
			// tagged sentinel; the fake stands in for that contract.
			saveWaitlistEmail: func(context.Context, string) error {
				return ErrWaitlistEmailExists
			},
		})
		app := newTestApp(repo, new(fakeMailer{}))

		status, payload := postWaitlist(t, app, `{"email":"dup@example.com"}`)
		if status != fiber.StatusConflict {
			t.Fatalf("status = %d, want 409", status)
		}
		if success, _ := payload["success"].(bool); success {
			t.Error("success should be false")
		}
	})

	t.Run("rejects a malformed body", func(t *testing.T) {
		app := newTestApp(new(fakeRepository{}), new(fakeMailer{}))

		status, _ := postWaitlist(t, app, `{`)
		if status != fiber.StatusBadRequest {
			t.Fatalf("status = %d, want 400", status)
		}
	})
}

// fakeGuard stands in for *auth.Module. It is the only guard this module
// injects, and it plays the part the real one does: reject the request, or let
// it through having written the locals the JWT gate writes.
//
// role is one of those locals rather than a second status field, because the
// admin check is httpx.RequireAdmin — it reads the role out of the locals and
// is not injected. A guard that answered 403 by itself would let the route drop
// that check without a test noticing, which is exactly what happened once.
type fakeGuard struct {
	authStatus int
	role       string
}

func (g fakeGuard) RequireAuth() fiber.Handler {
	return func(c fiber.Ctx) error {
		if g.authStatus != 0 {
			return c.SendStatus(g.authStatus)
		}

		c.Locals(httpx.LocalRole, g.role)

		return c.Next()
	}
}

// adminGuard is the caller the listing is meant for.
var adminGuard = fakeGuard{role: httpx.RoleAdmin}

// TestListWaitlistRoute covers the admin listing this module serves even
// though the path stays /users/waitlist: the route carries its own guards
// instead of inheriting the user group's.
func TestListWaitlistRoute(t *testing.T) {
	newApp := func(repo Repository, guard fakeGuard) *fiber.App {
		app := fiber.New()
		New(Deps{Service: newService(repo, new(fakeMailer{})), AuthMiddl: guard}).Routes(app)
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

	t.Run("returns the paginated waitlist to an admin", func(t *testing.T) {
		repo := new(fakeRepository{
			listWaitlist: func(_ context.Context, offset, limit uint) ([]Waitlist, uint, error) {
				if limit != 10 {
					t.Errorf("limit = %d, want 10", limit)
				}
				if offset != 0 {
					t.Errorf("offset = %d, want 0", offset)
				}
				return []Waitlist{{Email: "early@example.com"}}, 1, nil
			},
		})

		status, payload := get(t, newApp(repo, adminGuard))
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

	// The waitlist is the email address of everybody who ever signed up, so
	// what is being guarded is reaching the handler at all — not the status
	// code, which a later refactor could keep while the read went through.
	//
	// The repository records that instead of being left unstubbed. An unstubbed
	// fake panics on the nil hook, and a panic in a Fiber handler takes the
	// whole test binary down: the package reports one crash and no results,
	// which is how this route stayed unguarded through two commits.
	t.Run("is gated by the shared guards", func(t *testing.T) {
		cases := []struct {
			name  string
			guard fakeGuard
			want  int
		}{
			{"no session", fakeGuard{authStatus: fiber.StatusUnauthorized}, fiber.StatusUnauthorized},
			// The case the route actually lost: signed in, but not an admin.
			{"signed in as a plain user", fakeGuard{role: "user"}, fiber.StatusForbidden},
			{"signed in with no role at all", fakeGuard{}, fiber.StatusForbidden},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				reached := false
				repo := new(fakeRepository{
					listWaitlist: func(context.Context, uint, uint) ([]Waitlist, uint, error) {
						reached = true

						return nil, 0, nil
					},
				})

				if status, _ := get(t, newApp(repo, tc.guard)); status != tc.want {
					t.Errorf("status = %d, want %d", status, tc.want)
				}
				if reached {
					t.Error("the handler read the waitlist; the guards must stop the request before it")
				}
			})
		}
	})

	t.Run("refuses to build without a guard", func(t *testing.T) {
		// A missing guard must never expose the listing unprotected. It used
		// to drop the route instead, which was safe but silent: the listing
		// simply 404'd and only a user would notice. Failing at construction
		// puts the wiring bug where it belongs — at boot.
		defer func() {
			if recover() == nil {
				t.Error("New returned normally without a guard, want a panic")
			}
		}()

		New(Deps{Service: newService(new(fakeRepository{}), new(fakeMailer{}))})
	})
}
