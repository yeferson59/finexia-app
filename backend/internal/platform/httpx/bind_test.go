package httpx

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
)

// credentialsDTO stands in for the request bodies this codebase actually binds:
// a login, a password reset, an invitation acceptance, a TOTP challenge, a
// market-data API key. Every one of them carries a field that must only ever
// arrive in a body.
type credentialsDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Token    string `json:"token"`
	APIKey   string `json:"apiKey"`
}

// bindFrom sends one request through a handler that binds credentialsDTO and
// reports what the DTO ended up holding.
func bindFrom(t *testing.T, target string, headers map[string]string, body string) credentialsDTO {
	t.Helper()

	var bound credentialsDTO
	app := fiber.New()
	app.Post("/bind/:token?", func(c fiber.Ctx) error {
		dto, err := Bind[credentialsDTO](c)
		bound = dto
		if err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest(fiber.MethodPost, target, strings.NewReader(body))
	if body != "" {
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	}
	for name, value := range headers {
		req.Header.Set(name, value)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	return bound
}

// TestBindIgnoresQueryString is the regression guard for the switch away from
// Fiber's Bind().All(). All() merged the query string into any field the body
// left unset, which made every credential-bearing endpoint accept its secret in
// the URL — where it is recorded by access logs, forwarded in Referer headers,
// and kept in browser history.
func TestBindIgnoresQueryString(t *testing.T) {
	got := bindFrom(t,
		"/bind?email=victim@example.com&password=hunter2&token=reset-token&apiKey=live_key",
		nil,
		`{"email":"real@example.com"}`,
	)

	if got.Email != "real@example.com" {
		t.Errorf("Email = %q, want the body's value", got.Email)
	}
	if got.Password != "" {
		t.Errorf("Password = %q, want empty: a password in the query string must not populate the DTO", got.Password)
	}
	if got.Token != "" {
		t.Errorf("Token = %q, want empty: a reset token in the query string must not populate the DTO", got.Token)
	}
	if got.APIKey != "" {
		t.Errorf("APIKey = %q, want empty: an API key in the query string must not populate the DTO", got.APIKey)
	}
}

// TestBindIgnoresHeadersAndCookies covers the other two sources All() merged
// from. A field the caller never sent could otherwise be filled from a header
// or cookie, which is not where any handler here means to read it.
func TestBindIgnoresHeadersAndCookies(t *testing.T) {
	got := bindFrom(t,
		"/bind",
		map[string]string{
			"password": "from-header",
			"apiKey":   "from-header",
			"Cookie":   "token=from-cookie; password=from-cookie",
		},
		`{"email":"real@example.com"}`,
	)

	if got.Password != "" {
		t.Errorf("Password = %q, want empty: request headers must not populate a body DTO", got.Password)
	}
	if got.APIKey != "" {
		t.Errorf("APIKey = %q, want empty: request headers must not populate a body DTO", got.APIKey)
	}
	if got.Token != "" {
		t.Errorf("Token = %q, want empty: cookies must not populate a body DTO", got.Token)
	}
}

// TestBindIgnoresPathParameters pins the last source. Handlers read path
// parameters explicitly with ParamUUID or c.Params, which is what keeps it
// visible at the call site which part of the URL they trust.
func TestBindIgnoresPathParameters(t *testing.T) {
	got := bindFrom(t, "/bind/token-from-path", nil, `{"email":"real@example.com"}`)

	if got.Token != "" {
		t.Errorf("Token = %q, want empty: path parameters must not populate a body DTO", got.Token)
	}
}

// TestBindReadsTheBody is the counterpart — the fix must not stop the intended
// source from working.
func TestBindReadsTheBody(t *testing.T) {
	got := bindFrom(t, "/bind", nil,
		`{"email":"user@example.com","password":"in-the-body","token":"t","apiKey":"k"}`)

	want := credentialsDTO{Email: "user@example.com", Password: "in-the-body", Token: "t", APIKey: "k"}
	if got != want {
		t.Errorf("Bind = %+v, want %+v", got, want)
	}
}
