package httpx

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// runWithLocals mounts h behind a middleware that seeds the request locals the
// auth middleware would set, then reports what Identity made of them.
func runWithLocals(t *testing.T, locals map[string]any) (uuid.UUID, string, string, error) {
	t.Helper()

	var (
		gotID    uuid.UUID
		gotToken string
		gotRole  string
		gotErr   error
	)

	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error {
		for key, value := range locals {
			c.Locals(key, value)
		}
		return c.Next()
	}, func(c fiber.Ctx) error {
		gotID, gotToken, gotRole, gotErr = Identity(c)
		return c.SendStatus(fiber.StatusOK)
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	return gotID, gotToken, gotRole, gotErr
}

func TestIdentity(t *testing.T) {
	userID := uuid.New()

	t.Run("returns the seeded identity", func(t *testing.T) {
		id, token, role, err := runWithLocals(t, map[string]any{
			LocalUserID: userID.String(),
			LocalToken:  "raw-token",
			LocalRole:   "admin",
		})
		if err != nil {
			t.Fatalf("Identity: %v", err)
		}
		if id != userID || token != "raw-token" || role != "admin" {
			t.Errorf("Identity = (%s, %q, %q), want (%s, %q, %q)", id, token, role, userID, "raw-token", "admin")
		}
	})

	t.Run("errors when the request never passed the auth middleware", func(t *testing.T) {
		if _, _, _, err := runWithLocals(t, nil); err == nil {
			t.Error("Identity returned no error for empty locals")
		}
	})

	t.Run("errors on a malformed user id", func(t *testing.T) {
		_, _, _, err := runWithLocals(t, map[string]any{
			LocalUserID: "not-a-uuid",
			LocalToken:  "raw-token",
			LocalRole:   "user",
		})
		if err == nil {
			t.Error("Identity accepted a malformed user id")
		}
	})

	t.Run("errors on an incomplete claim set", func(t *testing.T) {
		// A valid id is not enough: the token and role must be there too, or
		// downstream handlers would act on a half-populated identity.
		for _, missing := range []string{LocalToken, LocalRole} {
			locals := map[string]any{
				LocalUserID: userID.String(),
				LocalToken:  "raw-token",
				LocalRole:   "user",
			}
			delete(locals, missing)

			_, _, _, err := runWithLocals(t, locals)
			if !errors.Is(err, ErrNoIdentity) {
				t.Errorf("without %s: err = %v, want ErrNoIdentity", missing, err)
			}
		}
	})
}

func TestParamUUID(t *testing.T) {
	want := uuid.New()

	run := func(t *testing.T, target string) (uuid.UUID, error) {
		t.Helper()
		var (
			got uuid.UUID
			err error
		)
		app := fiber.New()
		app.Get("/things/:id", func(c fiber.Ctx) error {
			got, err = ParamUUID(c, "id")
			return c.SendStatus(fiber.StatusOK)
		})
		resp, terr := app.Test(httptest.NewRequest("GET", target, nil))
		if terr != nil {
			t.Fatalf("app.Test: %v", terr)
		}
		defer func() { _ = resp.Body.Close() }()
		return got, err
	}

	t.Run("parses a uuid parameter", func(t *testing.T) {
		got, err := run(t, "/things/"+want.String())
		if err != nil {
			t.Fatalf("ParamUUID: %v", err)
		}
		if got != want {
			t.Errorf("ParamUUID = %s, want %s", got, want)
		}
	})

	t.Run("errors on a non-uuid parameter", func(t *testing.T) {
		if _, err := run(t, "/things/not-a-uuid"); err == nil {
			t.Error("ParamUUID accepted a non-uuid parameter")
		}
	})
}
