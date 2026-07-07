// Package geoip resolves public IP addresses to a coarse, human-readable
// location ("Bogotá, Colombia") for security notifications. Lookups are
// best-effort: any failure yields an empty string so callers can fall back
// gracefully instead of blocking or erroring an alert email.
package geoip

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"slices"
	"strings"
	"time"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
}

func New() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 5 * time.Second},
		baseURL:    "https://ipwho.is",
	}
}

// Locate returns "City, Region, Country" (deduplicated, best available
// subset) for a public IP. Private, loopback, and unparsable addresses never
// leave the process: they resolve to "" immediately, as does any API failure.
func (c *Client) Locate(ctx context.Context, ip string) string {
	parsed := net.ParseIP(strings.TrimSpace(ip))
	if parsed == nil || parsed.IsLoopback() || parsed.IsPrivate() || parsed.IsUnspecified() || parsed.IsLinkLocalUnicast() || parsed.IsLinkLocalMulticast() {
		return ""
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/"+parsed.String()+"?fields=success,city,region,country", nil)
	if err != nil {
		return ""
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return ""
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	var body struct {
		Success bool   `json:"success"`
		City    string `json:"city"`
		Region  string `json:"region"`
		Country string `json:"country"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil || !body.Success {
		return ""
	}

	parts := make([]string, 0, 3)
	for _, part := range []string{body.City, body.Region, body.Country} {
		if part = strings.TrimSpace(part); part != "" && !slices.Contains(parts, part) {
			parts = append(parts, part)
		}
	}
	return strings.Join(parts, ", ")
}
