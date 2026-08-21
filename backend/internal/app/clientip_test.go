package app

import (
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yeferson59/finexia-app/internal/platform/cache"
	"github.com/yeferson59/finexia-app/internal/platform/config"
	"github.com/yeferson59/finexia-app/internal/platform/logger"
	"github.com/yeferson59/finexia-app/internal/platform/mail"
)

// The client IP is a security control, not a log field: every rate limiter in
// the app keys on it, which is what stands between the public auth endpoints
// and unlimited credential stuffing, mail bombing, and lockout-based denial of
// service against every account at once.
//
// Fiber hands c.IP() the ProxyHeader *verbatim* unless EnableIPValidation is
// set, and it consults that header whenever the immediate peer is trusted —
// which the loopback and private ranges the composition root trusts make true
// for every request in a containerised deployment. So without validation
// c.IP() is a string the client picked, and rotating it per request gives each
// request its own bucket in every limiter.
//
// These tests go over a real loopback listener rather than app.Test: the peer
// address is the whole precondition, and app.Test synthesises a request whose
// remote address (0.0.0.0) is trusted by nothing, which would make the app
// look safe for the wrong reason.

// probeApp composes the real App — so the fiber.Config under test is the one
// the composition root builds — and serves it on 127.0.0.1, which puts the
// request's peer inside the trusted loopback range.
func probeApp(t *testing.T) string {
	t.Helper()

	pool, err := pgxpool.New(t.Context(), "postgres://user:pass@127.0.0.1:1/finexia_test")
	if err != nil {
		t.Fatalf("pgxpool.New: %v", err)
	}
	t.Cleanup(pool.Close)

	mailService, err := mail.New("", "test@example.com")
	if err != nil {
		t.Fatalf("mail.New: %v", err)
	}

	a, err := New(Deps{
		Envs: new(config.EnvConfig{
			Port:               "0",
			Environment:        "test",
			JWTSecret:          "kP4vN8xQ2mR7wL5tZ9bC3jH6yF1sD0aG",
			JWTAccessDuration:  15 * time.Minute,
			JWTRefreshDuration: 30 * 24 * time.Hour,
			PublicURL:          "http://localhost:8080",
			CORSEnabled:        true,
			CORSOrigin:         []string{"http://localhost:5173"},
			TrustProxy:         true,
		}),
		DB:      pool,
		Cache:   cache.Conn("127.0.0.1", "1", "", 0),
		Mail:    mailService,
		Keyring: testKeyring(t),
		Log:     logger.Noop(),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Only the probe route: wire() would start the schedulers, and what is
	// under test is the fiber.Config that New already built.
	a.fiber.Get("/__probe/client-ip", func(c fiber.Ctx) error {
		return c.SendString(c.IP())
	})

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	go func() { _ = a.fiber.Listener(ln) }()
	t.Cleanup(func() {
		_ = a.fiber.ShutdownWithTimeout(2 * time.Second)
	})

	return "http://" + ln.Addr().String()
}

// observedIP reports what c.IP() resolved to for a request carrying the given
// X-Forwarded-For value.
func observedIP(t *testing.T, baseURL, forwardedFor string) string {
	t.Helper()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, baseURL+"/__probe/client-ip", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if forwardedFor != "" {
		req.Header.Set(fiber.HeaderXForwardedFor, forwardedFor)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("probe request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body := make([]byte, 256)
	n, _ := resp.Body.Read(body)

	return string(body[:n])
}

// TestClientIPRejectsForgedForwardedForValues is the regression guard: a
// non-IP X-Forwarded-For must never become the rate-limit key. Before
// EnableIPValidation was set, every one of these strings came back from
// c.IP() unchanged.
func TestClientIPRejectsForgedForwardedForValues(t *testing.T) {
	baseURL := probeApp(t)

	for _, forged := range []string{
		"not-an-ip",
		"attacker-bucket-0001",
		"'; DROP TABLE sessions; --",
	} {
		got := observedIP(t, baseURL, forged)

		if got == forged {
			t.Errorf("c.IP() returned X-Forwarded-For verbatim (%q); a client-chosen string must never become the rate-limit key", forged)
		}
		if net.ParseIP(got) == nil {
			t.Errorf("c.IP() = %q for header %q, want a parsable address", got, forged)
		}
	}
}

// TestClientIPKeepsOneBucketAcrossForgedHeaders states the property the
// limiters actually depend on: an attacker rotating the header keeps landing
// on the same key, because a per-request key is an unlimited budget.
func TestClientIPKeepsOneBucketAcrossForgedHeaders(t *testing.T) {
	baseURL := probeApp(t)

	first := observedIP(t, baseURL, "bucket-a")
	second := observedIP(t, baseURL, "bucket-b")

	if first != second {
		t.Errorf("rotating X-Forwarded-For changed the client IP (%q then %q); every forged value would get its own rate-limit budget", first, second)
	}
}

// TestClientIPUsesRealClientBehindTrustedProxies is the other half of the fix:
// rejecting forged values would be worthless if it also broke real proxies, so
// a well-formed chain must still resolve to the client and not to a hop.
func TestClientIPUsesRealClientBehindTrustedProxies(t *testing.T) {
	baseURL := probeApp(t)

	// 203.0.113.7 is the client; 10.0.0.8 and 192.168.1.1 are private-range
	// hops the deployment trusts, so they must be skipped.
	if got := observedIP(t, baseURL, "203.0.113.7, 10.0.0.8, 192.168.1.1"); got != "203.0.113.7" {
		t.Errorf("c.IP() = %q, want the client address 203.0.113.7 from the forwarded chain", got)
	}
}
