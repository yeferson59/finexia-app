package helpers

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// TOTP implementation per RFC 6238 with the parameters every mainstream
// authenticator app (Google Authenticator, Authy, 1Password, ...) uses by
// default: HMAC-SHA1, 6 digits, 30-second time step.
const (
	totpDigits = 6
	totpPeriod = 30 * time.Second
)

// base32NoPadding matches what authenticator apps expect when the secret is
// typed manually: uppercase base32 without '=' padding.
var base32NoPadding = base32.StdEncoding.WithPadding(base32.NoPadding)

// GenerateTOTPSecret returns a new random 160-bit secret encoded in base32,
// the key size recommended by RFC 4226 for HMAC-SHA1.
func GenerateTOTPSecret() (string, error) {
	buf := make([]byte, 20)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return base32NoPadding.EncodeToString(buf), nil
}

// TOTPTimestep returns the RFC 6238 time-step counter for the given instant.
// Callers use it to remember which step a code was accepted for, so the same
// code can never be replayed within its validity window.
func TOTPTimestep(t time.Time) int64 {
	return t.Unix() / int64(totpPeriod.Seconds())
}

// TOTPCode computes the 6-digit code for the given secret and instant.
func TOTPCode(secret string, t time.Time) (string, error) {
	return totpCodeAt(secret, TOTPTimestep(t))
}

func totpCodeAt(secret string, timestep int64) (string, error) {
	key, err := base32NoPadding.DecodeString(strings.ToUpper(strings.ReplaceAll(secret, " ", "")))
	if err != nil {
		return "", fmt.Errorf("invalid totp secret: %w", err)
	}

	var counter [8]byte
	binary.BigEndian.PutUint64(counter[:], uint64(timestep))

	mac := hmac.New(sha1.New, key)
	mac.Write(counter[:])
	sum := mac.Sum(nil)

	// Dynamic truncation (RFC 4226 §5.3).
	offset := sum[len(sum)-1] & 0x0f
	value := binary.BigEndian.Uint32(sum[offset:offset+4]) & 0x7fffffff

	return fmt.Sprintf("%06d", value%1_000_000), nil
}

// VerifyTOTP checks a submitted code against the secret, accepting one step
// of clock drift on either side. It reports whether the code matched and, if
// so, the time step it matched at (for replay tracking). The comparison is
// constant-time.
func VerifyTOTP(secret, code string, t time.Time) (bool, int64) {
	code = strings.ReplaceAll(strings.TrimSpace(code), " ", "")
	if len(code) != totpDigits {
		return false, 0
	}

	current := TOTPTimestep(t)
	for _, step := range []int64{current, current - 1, current + 1} {
		expected, err := totpCodeAt(secret, step)
		if err != nil {
			return false, 0
		}

		if subtle.ConstantTimeCompare([]byte(expected), []byte(code)) == 1 {
			return true, step
		}
	}

	return false, 0
}

// BuildOTPAuthURL renders the otpauth:// provisioning URI that authenticator
// apps consume, either through a QR code or as a tappable link.
func BuildOTPAuthURL(issuer, accountName, secret string) string {
	label := url.PathEscape(issuer + ":" + accountName)
	params := url.Values{}

	params.Set("secret", secret)
	params.Set("issuer", issuer)
	params.Set("algorithm", "SHA1")
	params.Set("digits", strconv.Itoa(totpDigits))
	params.Set("period", strconv.Itoa(int(totpPeriod.Seconds())))

	return "otpauth://totp/" + label + "?" + params.Encode()
}
