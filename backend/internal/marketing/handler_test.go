package marketing

import (
	"context"
	"encoding/json"
	"errors"
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
			saveWaitlistEmail: func(context.Context, string) error {
				return errors.New("duplicate key value violates unique constraint")
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
