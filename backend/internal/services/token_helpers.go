package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"net"
	"strings"
)

// Temporary copies of helpers whose originals moved to the auth module with
// its core (Fase 4, PR A). The password-reset and invitation flows still
// living here need them until their own sub-areas migrate (PR B and PR C),
// at which point this file is deleted.

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
