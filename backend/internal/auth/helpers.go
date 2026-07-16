package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func generateRefreshToken() (raw, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return
	}
	raw = base64.URLEncoding.EncodeToString(b)
	h := sha256.Sum256(b)
	hash = hex.EncodeToString(h[:])
	return
}

func hashRefreshToken(raw string) (string, error) {
	b, err := base64.URLEncoding.DecodeString(raw)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:]), nil
}

// comparePassword checks a plaintext password against its bcrypt hash. It
// replaces the ComparePassword method the Account entity used to carry:
// identity structs hold data only, behavior stays in the module.
func comparePassword(hashed, plain string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain))
}

// truncate keeps a string within the column limits (ip VARCHAR(45),
// user_agent VARCHAR(255)) so an oversized header can never fail the insert.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

// sanitizeIP discards anything that isn't a real IP literal. c.IP() is
// fed by a client-influenced header (X-Forwarded-For); a malformed or
// spoofed value must never be persisted as a "known" login IP or shown back
// to the user in a security alert as if it were their real address.
func sanitizeIP(ipAddress string) string {
	if net.ParseIP(strings.TrimSpace(ipAddress)) == nil {
		return ""
	}
	return ipAddress
}

// humanizeExpiry renders a token lifetime for the email copy, preferring
// whole days when the duration divides evenly.
func humanizeExpiry(d time.Duration) string {
	hours := int(d.Hours())
	if hours <= 0 {
		return "poco tiempo"
	}
	if hours%24 == 0 {
		days := hours / 24
		if days == 1 {
			return "1 día"
		}
		return fmt.Sprintf("%d días", days)
	}
	if hours == 1 {
		return "1 hora"
	}
	return fmt.Sprintf("%d horas", hours)
}

// locateIP resolves the approximate location of an IP for security alert
// emails. Bounded by its own timeout so a slow lookup can only delay the
// (already asynchronous) email, never the request that triggered it.
func (s *Service) locateIP(ipAddress string) string {
	if s.geo == nil {
		return ""
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.geo.Locate(ctx, ipAddress)
}
