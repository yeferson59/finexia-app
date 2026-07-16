package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/yeferson59/finexia-app/internal/auth"
)

func TestResponseFromDomain(t *testing.T) {
	h := &Handlers{}

	cases := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{"invalid maps to 400", errors.New("invalid credentials"), fiber.StatusBadRequest},
		{"failed maps to 400", errors.New("failed to read file"), fiber.StatusBadRequest},
		{"not found maps to 404", errors.New("portfolio not found"), fiber.StatusNotFound},
		{"already exists maps to 409", errors.New("user already exists"), fiber.StatusConflict},
		{"duplicate maps to 409", errors.New("duplicate key value"), fiber.StatusConflict},
		{"unknown maps to 500", errors.New("connection reset"), fiber.StatusInternalServerError},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/", func(c fiber.Ctx) error {
				return h.responseFromDomain(c, tc.err, "message", "action")
			})

			resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
			if err != nil {
				t.Fatalf("app.Test: %v", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != tc.wantStatus {
				t.Errorf("status = %d, want %d", resp.StatusCode, tc.wantStatus)
			}

			body, _ := io.ReadAll(resp.Body)
			var payload map[string]any
			if err := json.Unmarshal(body, &payload); err != nil {
				t.Fatalf("invalid JSON body: %v", err)
			}
			if success, _ := payload["success"].(bool); success {
				t.Error("success = true, want false for error responses")
			}
			if payload["message"] != "message" {
				t.Errorf("message = %v, want %q", payload["message"], "message")
			}
		})
	}
}

func TestResponseStatusOk(t *testing.T) {
	h := &Handlers{}
	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error {
		return h.responseStatusOk(c, "done", "details", fiber.Map{"id": 1})
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("invalid JSON body: %v", err)
	}
	if success, _ := payload["success"].(bool); !success {
		t.Error("success = false, want true")
	}
	if payload["message"] != "done" {
		t.Errorf("message = %v, want done", payload["message"])
	}
	if payload["data"] == nil {
		t.Error("expected data in response")
	}
	if payload["timestamp"] == nil {
		t.Error("expected timestamp in response")
	}
}

func TestGetParamUUID(t *testing.T) {
	h := &Handlers{}

	newApp := func(capture *uuid.UUID, captureErr *error) *fiber.App {
		app := fiber.New()
		app.Get("/items/:id", func(c fiber.Ctx) error {
			id, err := h.getParamUUID(c, "id")
			*capture, *captureErr = id, err
			return c.SendStatus(fiber.StatusOK)
		})
		return app
	}

	t.Run("valid uuid", func(t *testing.T) {
		var got uuid.UUID
		var gotErr error
		app := newApp(&got, &gotErr)
		want := uuid.New()

		if _, err := app.Test(httptest.NewRequest("GET", "/items/"+want.String(), nil)); err != nil {
			t.Fatalf("app.Test: %v", err)
		}
		if gotErr != nil {
			t.Fatalf("getParamUUID: %v", gotErr)
		}
		if got != want {
			t.Errorf("got %s, want %s", got, want)
		}
	})

	t.Run("invalid uuid", func(t *testing.T) {
		var got uuid.UUID
		var gotErr error
		app := newApp(&got, &gotErr)

		if _, err := app.Test(httptest.NewRequest("GET", "/items/not-a-uuid", nil)); err != nil {
			t.Fatalf("app.Test: %v", err)
		}
		if gotErr == nil {
			t.Fatal("expected error for a malformed UUID")
		}
		if got != uuid.Nil {
			t.Errorf("got %s, want uuid.Nil", got)
		}
	})
}

func TestGetUserIDTokenRole(t *testing.T) {
	h := &Handlers{}

	run := func(t *testing.T, userID, token, role string) (uuid.UUID, string, string, error) {
		t.Helper()
		var (
			gotID    uuid.UUID
			gotToken string
			gotRole  string
			gotErr   error
		)
		app := fiber.New()
		app.Get("/", func(c fiber.Ctx) error {
			if userID != "" {
				c.Locals(auth.LocalUserID, userID)
			}
			if token != "" {
				c.Locals(auth.LocalToken, token)
			}
			if role != "" {
				c.Locals(auth.LocalRole, role)
			}
			gotID, gotToken, gotRole, gotErr = h.getUserIDTokenRole(c)
			return c.SendStatus(fiber.StatusOK)
		})
		if _, err := app.Test(httptest.NewRequest("GET", "/", nil)); err != nil {
			t.Fatalf("app.Test: %v", err)
		}
		return gotID, gotToken, gotRole, gotErr
	}

	t.Run("complete identity", func(t *testing.T) {
		want := uuid.New()
		id, token, role, err := run(t, want.String(), "the-token", "admin")
		if err != nil {
			t.Fatalf("getUserIDTokenRole: %v", err)
		}
		if id != want || token != "the-token" || role != "admin" {
			t.Errorf("got (%s, %s, %s), want (%s, the-token, admin)", id, token, role, want)
		}
	})

	t.Run("missing user id", func(t *testing.T) {
		if _, _, _, err := run(t, "", "the-token", "admin"); err == nil {
			t.Fatal("expected error when user id is missing")
		}
	})

	t.Run("invalid user id", func(t *testing.T) {
		if _, _, _, err := run(t, "not-a-uuid", "the-token", "admin"); err == nil {
			t.Fatal("expected error for malformed user id")
		}
	})

	t.Run("missing token", func(t *testing.T) {
		if _, _, _, err := run(t, uuid.NewString(), "", "admin"); err == nil {
			t.Fatal("expected error when token is missing")
		}
	})

	t.Run("missing role", func(t *testing.T) {
		if _, _, _, err := run(t, uuid.NewString(), "the-token", ""); err == nil {
			t.Fatal("expected error when role is missing")
		}
	})
}
