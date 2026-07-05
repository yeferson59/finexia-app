package helpers

import (
	"encoding/base32"
	"strings"
	"testing"
	"time"
)

// rfc6238Secret is the ASCII seed "12345678901234567890" from RFC 6238
// Appendix B, encoded the way our helpers expect it.
var rfc6238Secret = base32.StdEncoding.WithPadding(base32.NoPadding).
	EncodeToString([]byte("12345678901234567890"))

// TestTOTPCodeRFC6238Vectors checks the SHA-1 test vectors from RFC 6238
// Appendix B (truncated to our 6 digits from the published 8).
func TestTOTPCodeRFC6238Vectors(t *testing.T) {
	cases := []struct {
		unix int64
		want string
	}{
		{59, "287082"},
		{1111111109, "081804"},
		{1111111111, "050471"},
		{1234567890, "005924"},
		{2000000000, "279037"},
		{20000000000, "353130"},
	}

	for _, tc := range cases {
		got, err := TOTPCode(rfc6238Secret, time.Unix(tc.unix, 0).UTC())
		if err != nil {
			t.Fatalf("TOTPCode(t=%d): %v", tc.unix, err)
		}
		if got != tc.want {
			t.Errorf("TOTPCode(t=%d) = %s, want %s", tc.unix, got, tc.want)
		}
	}
}

func TestVerifyTOTPWindow(t *testing.T) {
	now := time.Unix(1111111111, 0).UTC()

	current, _ := TOTPCode(rfc6238Secret, now)
	previous, _ := TOTPCode(rfc6238Secret, now.Add(-30*time.Second))
	next, _ := TOTPCode(rfc6238Secret, now.Add(30*time.Second))
	tooOld, _ := TOTPCode(rfc6238Secret, now.Add(-60*time.Second))

	for name, code := range map[string]string{"current": current, "previous": previous, "next": next} {
		if ok, _ := VerifyTOTP(rfc6238Secret, code, now); !ok {
			t.Errorf("VerifyTOTP rejected %s step code %s", name, code)
		}
	}

	if ok, _ := VerifyTOTP(rfc6238Secret, tooOld, now); ok {
		t.Error("VerifyTOTP accepted a code two steps old")
	}
	if ok, _ := VerifyTOTP(rfc6238Secret, "000000", now); ok {
		t.Error("VerifyTOTP accepted a wrong code")
	}
	if ok, _ := VerifyTOTP(rfc6238Secret, "12345", now); ok {
		t.Error("VerifyTOTP accepted a 5-digit code")
	}

	// Codes typed with spaces (as some apps display them) must still verify.
	spaced := current[:3] + " " + current[3:]
	if ok, _ := VerifyTOTP(rfc6238Secret, spaced, now); !ok {
		t.Errorf("VerifyTOTP rejected spaced code %q", spaced)
	}
}

func TestVerifyTOTPReturnsMatchedStep(t *testing.T) {
	now := time.Unix(1111111111, 0).UTC()
	code, _ := TOTPCode(rfc6238Secret, now)

	ok, step := VerifyTOTP(rfc6238Secret, code, now)
	if !ok {
		t.Fatal("expected code to verify")
	}
	if want := TOTPTimestep(now); step != want {
		t.Errorf("matched step = %d, want %d", step, want)
	}
}

func TestGenerateTOTPSecret(t *testing.T) {
	a, err := GenerateTOTPSecret()
	if err != nil {
		t.Fatalf("GenerateTOTPSecret: %v", err)
	}
	b, err := GenerateTOTPSecret()
	if err != nil {
		t.Fatalf("GenerateTOTPSecret: %v", err)
	}
	if a == b {
		t.Error("expected unique secrets")
	}
	// 20 bytes → 32 base32 chars, no padding.
	if len(a) != 32 || strings.Contains(a, "=") {
		t.Errorf("unexpected secret format: %q", a)
	}
	if _, err := TOTPCode(a, time.Now()); err != nil {
		t.Errorf("generated secret does not produce codes: %v", err)
	}
}

func TestBuildOTPAuthURL(t *testing.T) {
	url := BuildOTPAuthURL("Finexia", "user@example.com", "ABC234")

	for _, want := range []string{
		"otpauth://totp/Finexia:user@example.com?",
		"secret=ABC234",
		"issuer=Finexia",
		"algorithm=SHA1",
		"digits=6",
		"period=30",
	} {
		if !strings.Contains(url, want) {
			t.Errorf("otpauth URL %q missing %q", url, want)
		}
	}
}
