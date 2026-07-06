package geoip

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestClient(handler http.HandlerFunc) (*Client, *httptest.Server) {
	server := httptest.NewServer(handler)
	return &Client{
		httpClient: &http.Client{Timeout: time.Second},
		baseURL:    server.URL,
	}, server
}

func TestLocate(t *testing.T) {
	t.Run("resolves a public IP to city, region and country", func(t *testing.T) {
		client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/203.0.113.9" {
				t.Errorf("path = %q, want /203.0.113.9", r.URL.Path)
			}
			_, _ = w.Write([]byte(`{"success":true,"city":"Bogotá","region":"Bogota D.C.","country":"Colombia"}`))
		})
		defer server.Close()

		if got := client.Locate(context.Background(), "203.0.113.9"); got != "Bogotá, Bogota D.C., Colombia" {
			t.Fatalf("Locate = %q", got)
		}
	})

	t.Run("deduplicates repeated fields and skips empty ones", func(t *testing.T) {
		client, server := newTestClient(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"success":true,"city":"Singapore","region":"","country":"Singapore"}`))
		})
		defer server.Close()

		if got := client.Locate(context.Background(), "203.0.113.9"); got != "Singapore" {
			t.Fatalf("Locate = %q", got)
		}
	})

	t.Run("returns empty on API failure responses", func(t *testing.T) {
		client, server := newTestClient(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"success":false}`))
		})
		defer server.Close()

		if got := client.Locate(context.Background(), "203.0.113.9"); got != "" {
			t.Fatalf("Locate = %q, want empty", got)
		}
	})

	t.Run("never queries the API for private or invalid addresses", func(t *testing.T) {
		client, server := newTestClient(func(http.ResponseWriter, *http.Request) {
			t.Error("unexpected API call")
		})
		defer server.Close()

		for _, ip := range []string{"", "not-an-ip", "127.0.0.1", "10.0.0.4", "192.168.1.20", "::1", "0.0.0.0"} {
			if got := client.Locate(context.Background(), ip); got != "" {
				t.Fatalf("Locate(%q) = %q, want empty", ip, got)
			}
		}
	})
}
