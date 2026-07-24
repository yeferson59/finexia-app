package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/paginate"
)

func perform(t *testing.T, h fiber.Handler) (int, map[string]any) {
	t.Helper()
	app := fiber.New()
	app.Get("/", h)

	resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
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

func TestFromDomainMapping(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want int
	}{
		{"tagged too-many", AsTooManyRequests(errors.New("too many failed login attempts")), fiber.StatusTooManyRequests},
		{"tagged bad-request", AsBadRequest(errors.New("invalid credentials")), fiber.StatusBadRequest},
		{"tagged not-found", AsNotFound(errors.New("portfolio not found")), fiber.StatusNotFound},
		{"tagged conflict", AsConflict(errors.New("email already exists")), fiber.StatusConflict},
		// No substring fallback anymore: an untagged error is always 500,
		// regardless of what its message happens to contain.
		{"untagged with keyword", errors.New("invalid credentials"), fiber.StatusInternalServerError},
		{"untagged plain", errors.New("something else entirely"), fiber.StatusInternalServerError},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			status, payload := perform(t, func(c fiber.Ctx) error {
				return FromDomain(c, tc.err, "msg", "domain:action")
			})
			if status != tc.want {
				t.Errorf("status = %d, want %d", status, tc.want)
			}
			if success, _ := payload["success"].(bool); success {
				t.Error("success should be false")
			}
			if action, _ := payload["action"].(string); action != "domain:action" {
				t.Errorf("action = %q, want domain:action", action)
			}
		})
	}
}

func TestFromDomainTypedKindWins(t *testing.T) {
	sentinel := errors.New("portfolio not found")

	cases := []struct {
		name string
		err  error
		want int
	}{
		// A tagged NotFound must map to 404 even though the message also
		// contains "failed", which the frozen substring order would send to
		// 400 — this is the exact bug typed errors fix (TECH_DEBT #1).
		{"tag beats failed-substring", AsNotFound(errors.New("failed: portfolio not found")), fiber.StatusNotFound},
		{"AsBadRequest", AsBadRequest(errors.New("whatever")), fiber.StatusBadRequest},
		{"AsConflict", AsConflict(errors.New("whatever")), fiber.StatusConflict},
		{"AsTooManyRequests", AsTooManyRequests(errors.New("whatever")), fiber.StatusTooManyRequests},
		// A tag survives a further fmt.Errorf %w wrap.
		{"tag survives wrapping", fmt.Errorf("loading portfolio: %w", AsNotFound(sentinel)), fiber.StatusNotFound},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			status, _ := perform(t, func(c fiber.Ctx) error {
				return FromDomain(c, tc.err, "msg", "domain:action")
			})
			if status != tc.want {
				t.Errorf("status = %d, want %d", status, tc.want)
			}
		})
	}
}

func TestTaggedIsTransparentToErrorsIs(t *testing.T) {
	sentinel := errors.New("portfolio not found")
	wrapped := fmt.Errorf("loading portfolio: %w", AsNotFound(sentinel))

	if !errors.Is(wrapped, sentinel) {
		t.Error("errors.Is should still match the wrapped sentinel through the tag")
	}
	if got := Tagged(KindNotFound, nil); got != nil {
		t.Errorf("Tagged(kind, nil) = %v, want nil", got)
	}
}

func TestEnvelopes(t *testing.T) {
	t.Run("OK carries the success envelope", func(t *testing.T) {
		status, payload := perform(t, func(c fiber.Ctx) error {
			return OK(c, "msg", "det", fiber.Map{"k": "v"})
		})
		if status != fiber.StatusOK {
			t.Fatalf("status = %d, want 200", status)
		}
		for _, key := range []string{"success", "message", "details", "data", "timestamp"} {
			if _, ok := payload[key]; !ok {
				t.Errorf("missing envelope field %q", key)
			}
		}
	})

	t.Run("ErrorAction carries details and action", func(t *testing.T) {
		status, payload := perform(t, func(c fiber.Ctx) error {
			return ErrorAction(c, fiber.StatusForbidden, "msg", "det", "auth:x")
		})
		if status != fiber.StatusForbidden {
			t.Fatalf("status = %d, want 403", status)
		}
		if action, _ := payload["action"].(string); action != "auth:x" {
			t.Errorf("action = %q, want auth:x", action)
		}
		if details, _ := payload["details"].(string); details != "det" {
			t.Errorf("details = %q, want det", details)
		}
	})

	t.Run("Unauthorized returns 401", func(t *testing.T) {
		status, _ := perform(t, func(c fiber.Ctx) error {
			return Unauthorized(c, "msg", "det")
		})
		if status != fiber.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", status)
		}
	})
}

func TestPaginationMetadata(t *testing.T) {
	info := &paginate.PageInfo{Page: 2, Limit: 10, Offset: 10}
	meta := PaginationMetadata(info, 35, "itemsForPage", "totalItems")

	if meta["totalPages"] != uint(4) {
		t.Errorf("totalPages = %v, want 4", meta["totalPages"])
	}
	if meta["previous"] != true || meta["next"] != true {
		t.Errorf("previous/next = %v/%v, want true/true", meta["previous"], meta["next"])
	}
	if meta["itemsForPage"] != 10 || meta["totalItems"] != uint(35) {
		t.Errorf("unexpected keys: %v", meta)
	}
}
