package auth

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
)

// newRBACApp builds an app that fakes the JWT middleware by injecting the
// given role into locals before the RBAC handler runs.
func newRBACApp(role string, handler fiber.Handler) *fiber.App {
	app := fiber.New()
	app.Get("/protected", func(c fiber.Ctx) error {
		if role != "" {
			c.Locals(LocalRole, role)
		}
		return c.Next()
	}, handler, func(c fiber.Ctx) error {
		return c.SendString("ok")
	})
	return app
}

func TestRequireRole(t *testing.T) {
	m := &Module{}

	cases := []struct {
		name       string
		role       string
		allowed    []string
		wantStatus int
	}{
		{"matching role passes", "admin", []string{"admin"}, fiber.StatusOK},
		{"role in list passes", "editor", []string{"admin", "editor"}, fiber.StatusOK},
		{"other role is forbidden", "user", []string{"admin"}, fiber.StatusForbidden},
		{"missing role is forbidden", "", []string{"admin"}, fiber.StatusForbidden},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			app := newRBACApp(tc.role, m.RequireRole(tc.allowed...))

			resp, err := app.Test(httptest.NewRequest("GET", "/protected", nil))
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

func TestRequireAdmin(t *testing.T) {
	m := &Module{}

	t.Run("admin passes", func(t *testing.T) {
		app := newRBACApp("admin", m.RequireAdmin())
		resp, err := app.Test(httptest.NewRequest("GET", "/protected", nil))
		if err != nil {
			t.Fatalf("app.Test: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("status = %d, want 200", resp.StatusCode)
		}
	})

	t.Run("regular user is forbidden", func(t *testing.T) {
		app := newRBACApp("user", m.RequireAdmin())
		resp, err := app.Test(httptest.NewRequest("GET", "/protected", nil))
		if err != nil {
			t.Fatalf("app.Test: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode != fiber.StatusForbidden {
			t.Errorf("status = %d, want 403", resp.StatusCode)
		}
	})
}
