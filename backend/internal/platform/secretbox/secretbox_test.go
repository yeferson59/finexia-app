package secretbox

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"slices"
	"strings"
	"testing"
)

// testKey returns a deterministic-length, random 32-byte KEK encoded the way
// configuration supplies it.
func testKey(t *testing.T) string {
	t.Helper()

	key := make([]byte, keyLen)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("rand: %v", err)
	}

	return base64.StdEncoding.EncodeToString(key)
}

func testKeyring(t *testing.T) *Keyring {
	t.Helper()

	ring, err := NewKeyring([]string{"1:" + testKey(t)}, 1)
	if err != nil {
		t.Fatalf("NewKeyring: %v", err)
	}

	return ring
}

func TestSealOpenRoundTrip(t *testing.T) {
	ring := testKeyring(t)
	aad := AAD("11111111-1111-1111-1111-111111111111", "finnhub")
	secret := []byte("d0nt-l34k-me")

	sealed, err := ring.Seal(secret, aad)
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}

	got, err := ring.Open(sealed, aad)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	if !bytes.Equal(got, secret) {
		t.Fatalf("round trip = %q, want %q", got, secret)
	}
}

// Sealing the same secret twice must not produce the same bytes, otherwise the
// stored rows would reveal which users share an API key.
func TestSealIsNotDeterministic(t *testing.T) {
	ring := testKeyring(t)
	aad := AAD("11111111-1111-1111-1111-111111111111", "finnhub")

	first, err := ring.Seal([]byte("same-secret"), aad)
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}
	second, err := ring.Seal([]byte("same-secret"), aad)
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}

	if bytes.Equal(first.Ciphertext, second.Ciphertext) {
		t.Fatal("two seals of the same secret produced identical ciphertext")
	}
	if bytes.Equal(first.Nonce, second.Nonce) {
		t.Fatal("two seals reused the same nonce")
	}
}

// The property the whole design rests on: write access to the table must not
// grant read access to somebody else's key. Moving a sealed row to another
// user's AAD has to fail.
func TestOpenRejectsACiphertextMovedToAnotherOwner(t *testing.T) {
	ring := testKeyring(t)

	alice := AAD("11111111-1111-1111-1111-111111111111", "finnhub")
	bob := AAD("22222222-2222-2222-2222-222222222222", "finnhub")

	sealed, err := ring.Seal([]byte("alice-api-key"), alice)
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}

	if _, err := ring.Open(sealed, bob); !errors.Is(err, ErrDecrypt) {
		t.Fatalf("Open with another owner's AAD = %v, want ErrDecrypt", err)
	}
}

// Same row, same user, different provider: the provider is part of the binding
// too, so a Finnhub credential cannot be read back as an Alpha Vantage one.
func TestOpenRejectsAMismatchedProvider(t *testing.T) {
	ring := testKeyring(t)
	user := "11111111-1111-1111-1111-111111111111"

	sealed, err := ring.Seal([]byte("api-key"), AAD(user, "finnhub"))
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}

	if _, err := ring.Open(sealed, AAD(user, "alphavantage")); !errors.Is(err, ErrDecrypt) {
		t.Fatalf("Open with a different provider = %v, want ErrDecrypt", err)
	}
}

func TestOpenRejectsTamperedFields(t *testing.T) {
	aad := AAD("11111111-1111-1111-1111-111111111111", "finnhub")

	tests := map[string]func(*Sealed){
		"ciphertext": func(s *Sealed) { s.Ciphertext[0] ^= 0xff },
		"nonce":      func(s *Sealed) { s.Nonce[0] ^= 0xff },
		"wrappedDEK": func(s *Sealed) { s.WrappedDEK[0] ^= 0xff },
	}

	for name, tamper := range tests {
		t.Run(name, func(t *testing.T) {
			ring := testKeyring(t)

			sealed, err := ring.Seal([]byte("api-key"), aad)
			if err != nil {
				t.Fatalf("Seal: %v", err)
			}

			tamper(&sealed)

			if _, err := ring.Open(sealed, aad); !errors.Is(err, ErrDecrypt) {
				t.Fatalf("Open after tampering with %s = %v, want ErrDecrypt", name, err)
			}
		})
	}
}

func TestOpenRejectsAShortNonce(t *testing.T) {
	ring := testKeyring(t)
	aad := AAD("11111111-1111-1111-1111-111111111111", "finnhub")

	sealed, err := ring.Seal([]byte("api-key"), aad)
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}

	sealed.Nonce = sealed.Nonce[:len(sealed.Nonce)-1]

	if _, err := ring.Open(sealed, aad); !errors.Is(err, ErrMalformedSealed) {
		t.Fatalf("Open with a short nonce = %v, want ErrMalformedSealed", err)
	}
}

// A row must not be openable by claiming a KEK version the process holds but
// that did not wrap it.
func TestOpenRejectsAForgedKEKVersion(t *testing.T) {
	ring, err := NewKeyring([]string{"1:" + testKey(t), "2:" + testKey(t)}, 1)
	if err != nil {
		t.Fatalf("NewKeyring: %v", err)
	}

	aad := AAD("11111111-1111-1111-1111-111111111111", "finnhub")

	sealed, err := ring.Seal([]byte("api-key"), aad)
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}

	sealed.KEKVersion = 2

	if _, err := ring.Open(sealed, aad); !errors.Is(err, ErrDecrypt) {
		t.Fatalf("Open with a forged version = %v, want ErrDecrypt", err)
	}
}

func TestOpenRejectsAnUnknownKEKVersion(t *testing.T) {
	ring := testKeyring(t)
	aad := AAD("11111111-1111-1111-1111-111111111111", "finnhub")

	sealed, err := ring.Seal([]byte("api-key"), aad)
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}

	sealed.KEKVersion = 99

	if _, err := ring.Open(sealed, aad); !errors.Is(err, ErrUnknownKEK) {
		t.Fatalf("Open with an unknown version = %v, want ErrUnknownKEK", err)
	}
}

// Rotation: a row sealed under the retired KEK still opens, and Rewrap moves it
// onto the active one without ever needing the provider key back.
func TestRewrapMovesARowOntoTheActiveKEK(t *testing.T) {
	keys := []string{"1:" + testKey(t), "2:" + testKey(t)}
	aad := AAD("11111111-1111-1111-1111-111111111111", "finnhub")
	secret := []byte("api-key")

	old, err := NewKeyring(keys, 1)
	if err != nil {
		t.Fatalf("NewKeyring: %v", err)
	}

	sealed, err := old.Seal(secret, aad)
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}

	rotated, err := NewKeyring(keys, 2)
	if err != nil {
		t.Fatalf("NewKeyring: %v", err)
	}

	// Still readable before the rewrap, because version 1 is retired, not dropped.
	if _, err := rotated.Open(sealed, aad); err != nil {
		t.Fatalf("Open of a pre-rotation row: %v", err)
	}

	rewrapped, err := rotated.Rewrap(sealed)
	if err != nil {
		t.Fatalf("Rewrap: %v", err)
	}

	if rewrapped.KEKVersion != 2 {
		t.Fatalf("KEKVersion after Rewrap = %d, want 2", rewrapped.KEKVersion)
	}
	if bytes.Equal(rewrapped.WrappedDEK, sealed.WrappedDEK) {
		t.Fatal("Rewrap left the wrapped DEK unchanged")
	}
	// The payload is untouched: that is what makes rotation cheap.
	if !bytes.Equal(rewrapped.Ciphertext, sealed.Ciphertext) {
		t.Fatal("Rewrap re-encrypted the payload; it should only re-wrap the DEK")
	}

	got, err := rotated.Open(rewrapped, aad)
	if err != nil {
		t.Fatalf("Open after Rewrap: %v", err)
	}
	if !bytes.Equal(got, secret) {
		t.Fatalf("after Rewrap = %q, want %q", got, secret)
	}
}

// Once the old KEK is dropped from the environment, rows that were rewrapped
// keep working and rows that were not are simply unreadable — never silently
// wrong.
func TestRotationDropsTheRetiredKEK(t *testing.T) {
	firstKey, secondKey := []string{"1:" + testKey(t)}, []string{"2:" + testKey(t)}
	aad := AAD("11111111-1111-1111-1111-111111111111", "finnhub")

	old, err := NewKeyring(firstKey, 1)
	if err != nil {
		t.Fatalf("NewKeyring: %v", err)
	}

	sealed, err := old.Seal([]byte("api-key"), aad)
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}

	only2, err := NewKeyring(secondKey, 2)
	if err != nil {
		t.Fatalf("NewKeyring: %v", err)
	}

	if _, err := only2.Open(sealed, aad); !errors.Is(err, ErrUnknownKEK) {
		t.Fatalf("Open after dropping the KEK = %v, want ErrUnknownKEK", err)
	}
}

func TestRewrapIsANoopOnTheActiveVersion(t *testing.T) {
	ring := testKeyring(t)
	aad := AAD("11111111-1111-1111-1111-111111111111", "finnhub")

	sealed, err := ring.Seal([]byte("api-key"), aad)
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}

	rewrapped, err := ring.Rewrap(sealed)
	if err != nil {
		t.Fatalf("Rewrap: %v", err)
	}

	if !bytes.Equal(rewrapped.WrappedDEK, sealed.WrappedDEK) {
		t.Fatal("Rewrap rewrapped a row that was already on the active KEK")
	}
}

// The reason the package exists: what lands in Postgres must not contain the
// secret in any recoverable form.
func TestSealedBytesDoNotContainThePlaintext(t *testing.T) {
	ring := testKeyring(t)
	secret := []byte("SUPERSECRETAPIKEY1234")

	sealed, err := ring.Seal(secret, AAD("11111111-1111-1111-1111-111111111111", "finnhub"))
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}

	// Everything the database row holds, concatenated.
	row := slices.Concat(sealed.WrappedDEK, sealed.Nonce, sealed.Ciphertext)

	if bytes.Contains(row, secret) {
		t.Fatal("the stored row contains the plaintext secret")
	}
	if strings.Contains(base64.StdEncoding.EncodeToString(row), base64.StdEncoding.EncodeToString(secret)) {
		t.Fatal("the stored row contains the base64 of the plaintext secret")
	}
}

func TestZeroWipesTheBuffer(t *testing.T) {
	buf := []byte("api-key")

	Zero(buf)

	for i, b := range buf {
		if b != 0 {
			t.Fatalf("byte %d = %d, want 0", i, b)
		}
	}
}

func TestAADPartsAreUnambiguous(t *testing.T) {
	// Without a separator these two would collapse to the same AAD, letting a
	// credential be opened under the wrong binding.
	if bytes.Equal(AAD("ab", "c"), AAD("a", "bc")) {
		t.Fatal("AAD is ambiguous across part boundaries")
	}
}

func TestNewKeyringRejectsBadConfiguration(t *testing.T) {
	valid := testKey(t)

	tests := map[string]struct {
		keys   []string
		active uint8
	}{
		"empty":               {[]string{""}, 1},
		"only separators":     {[]string{"", ""}, 1},
		"no version":          {[]string{valid}, 1},
		"non-numeric version": {[]string{"one:" + valid}, 1},
		"bad base64":          {[]string{"1:not-base64!!"}, 1},
		"short key":           {[]string{"1:" + base64.StdEncoding.EncodeToString([]byte("too-short"))}, 1},
		"duplicate version":   {[]string{"1:" + valid, "1:" + valid}, 1},
		"unknown active":      {[]string{"1:" + valid}, 7},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := NewKeyring(tc.keys, tc.active); err == nil {
				t.Fatal("NewKeyring accepted an invalid configuration")
			}
		})
	}
}

func TestNewKeyringAcceptsMultipleVersions(t *testing.T) {
	ring, err := NewKeyring([]string{" 1:" + testKey(t), "2:" + testKey(t) + " "}, 2)
	if err != nil {
		t.Fatalf("NewKeyring: %v", err)
	}

	if ring.ActiveVersion() != 2 {
		t.Fatalf("ActiveVersion = %d, want 2", ring.ActiveVersion())
	}
	if len(ring.keys) != 2 {
		t.Fatalf("parsed %d keys, want 2", len(ring.keys))
	}
}
