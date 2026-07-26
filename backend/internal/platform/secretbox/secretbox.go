// Package secretbox seals third-party credentials (today: the market-data API
// keys each user brings) so that a leaked database is not enough to recover
// them. It implements envelope encryption with AES-256-GCM: every secret gets
// its own random data key (DEK), and the DEK travels wrapped under a key
// encryption key (KEK) that lives in the environment — never in the database.
//
// Two properties matter and both come from the AAD:
//
//   - The payload is bound to caller-supplied associated data (owner and
//     provider). Copying a ciphertext into another user's row makes Open fail,
//     so write access to the table does not grant access to someone else's key.
//   - The wrapped DEK is bound to its KEK version, so a retired KEK cannot be
//     forced back into use by editing the stored version.
//
// The KEK is held in memory only. Callers get the plaintext for the duration of
// one provider call and are expected to Zero it afterwards.
package secretbox

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// keyLen is the AES-256 key size, required for both the KEK and every DEK.
const keyLen = 32

// aadSep joins the parts of a caller AAD. It is a control character, so it
// cannot appear in the UUIDs and provider slugs that make up the AAD and the
// encoding stays unambiguous.
const aadSep = "\x1f"

// dekAADPrefix domain-separates the DEK-wrapping layer from the payload layer,
// so a wrapped DEK can never be opened as if it were a payload.
const dekAADPrefix = "finexia/secretbox/dek/v"

var (
	ErrNoKeys           = errors.New("secretbox: no KEK configured")
	ErrUnknownKEK       = errors.New("secretbox: unknown KEK version")
	ErrMalformedSealed  = errors.New("secretbox: malformed sealed value")
	ErrDecrypt          = errors.New("secretbox: cannot decrypt (wrong key or altered data)")
	errKeyLen           = errors.New("secretbox: KEK must be 32 bytes")
	errMalformedKeysEnv = errors.New(`secretbox: keys must be formatted as "version:base64,version:base64"`)
)

// Sealed is the storable form of a secret: everything except the KEK. The three
// byte slices map one-to-one onto the wrapped_dek, nonce and ciphertext columns.
type Sealed struct {
	// KEKVersion identifies which KEK wrapped the DEK, so rotation can retire a
	// key without rewriting every row at once.
	KEKVersion int
	// WrappedDEK is the DEK sealed under the KEK, prefixed by its own nonce.
	WrappedDEK []byte
	// Nonce is the payload nonce, used with the DEK.
	Nonce []byte
	// Ciphertext is the secret sealed under the DEK.
	Ciphertext []byte
}

// Keyring holds every KEK the process accepts: the active one for sealing, plus
// any retired versions still needed to open rows that have not been rewrapped.
type Keyring struct {
	keys   map[int][]byte
	active int
}

// NewKeyring parses the KEK material from configuration. keys is a comma
// separated list of "version:base64" entries and active names the version used
// for new seals; both come straight from the environment.
//
// It is deliberately strict — a missing or short KEK is a configuration error
// that must stop the process at startup, not a condition to paper over with a
// default, which would silently store secrets under a guessable key.
func NewKeyring(keys, active string) (*Keyring, error) {
	keys = strings.TrimSpace(keys)
	if keys == "" {
		return nil, ErrNoKeys
	}

	parsed := make(map[int][]byte)

	for entry := range strings.SplitSeq(keys, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}

		rawVersion, encoded, found := strings.Cut(entry, ":")
		if !found {
			return nil, errMalformedKeysEnv
		}

		version, err := strconv.Atoi(strings.TrimSpace(rawVersion))
		if err != nil {
			return nil, fmt.Errorf("%w: bad version %q", errMalformedKeysEnv, rawVersion)
		}
		if _, duplicate := parsed[version]; duplicate {
			return nil, fmt.Errorf("secretbox: KEK version %d declared twice", version)
		}

		key, err := base64.StdEncoding.DecodeString(strings.TrimSpace(encoded))
		if err != nil {
			return nil, fmt.Errorf("secretbox: KEK version %d is not valid base64: %w", version, err)
		}
		if len(key) != keyLen {
			return nil, fmt.Errorf("%w: version %d has %d", errKeyLen, version, len(key))
		}

		parsed[version] = key
	}

	if len(parsed) == 0 {
		return nil, ErrNoKeys
	}

	activeVersion, err := strconv.Atoi(strings.TrimSpace(active))
	if err != nil {
		return nil, fmt.Errorf("secretbox: active KEK version %q is not a number", active)
	}
	if _, ok := parsed[activeVersion]; !ok {
		return nil, fmt.Errorf("%w: active version %d was not supplied", ErrUnknownKEK, activeVersion)
	}

	return &Keyring{keys: parsed, active: activeVersion}, nil
}

// ActiveVersion reports the KEK version new seals are written under. Callers use
// it to decide which rows still need a Rewrap after a rotation.
func (k *Keyring) ActiveVersion() int { return k.active }

// Seal encrypts plaintext under a fresh DEK and wraps that DEK under the active
// KEK. aad binds the result to its owner: Open only succeeds when given the same
// value, which is what stops a sealed secret from being replayed into another
// row. Build it with AAD.
func (k *Keyring) Seal(plaintext, aad []byte) (Sealed, error) {
	dek := make([]byte, keyLen)
	if _, err := rand.Read(dek); err != nil {
		return Sealed{}, fmt.Errorf("secretbox: generate DEK: %w", err)
	}
	defer Zero(dek)

	payloadAEAD, err := newAEAD(dek)
	if err != nil {
		return Sealed{}, err
	}

	nonce, err := newNonce(payloadAEAD.NonceSize())
	if err != nil {
		return Sealed{}, err
	}

	wrapped, err := k.wrapDEK(dek, k.active)
	if err != nil {
		return Sealed{}, err
	}

	return Sealed{
		KEKVersion: k.active,
		WrappedDEK: wrapped,
		Nonce:      nonce,
		Ciphertext: payloadAEAD.Seal(nil, nonce, plaintext, aad),
	}, nil
}

// Open reverses Seal. It returns ErrDecrypt when the KEK is wrong, the stored
// bytes were altered, or aad does not match the one used to seal — the three are
// deliberately indistinguishable to the caller.
//
// The returned plaintext should be passed to Zero once the caller is done.
func (k *Keyring) Open(s Sealed, aad []byte) ([]byte, error) {
	dek, err := k.unwrapDEK(s.WrappedDEK, s.KEKVersion)
	if err != nil {
		return nil, err
	}
	defer Zero(dek)

	payloadAEAD, err := newAEAD(dek)
	if err != nil {
		return nil, err
	}

	if len(s.Nonce) != payloadAEAD.NonceSize() {
		return nil, ErrMalformedSealed
	}

	plaintext, err := payloadAEAD.Open(nil, s.Nonce, s.Ciphertext, aad)
	if err != nil {
		return nil, ErrDecrypt
	}

	return plaintext, nil
}

// Rewrap moves a sealed value onto the active KEK without touching the payload,
// which is what makes rotation cheap: the secret itself is never decrypted, so
// rotating does not require reaching out to any provider. It needs no AAD
// because the caller's binding lives on the payload layer, which is left alone.
func (k *Keyring) Rewrap(s Sealed) (Sealed, error) {
	if s.KEKVersion == k.active {
		return s, nil
	}

	dek, err := k.unwrapDEK(s.WrappedDEK, s.KEKVersion)
	if err != nil {
		return Sealed{}, err
	}
	defer Zero(dek)

	wrapped, err := k.wrapDEK(dek, k.active)
	if err != nil {
		return Sealed{}, err
	}

	s.KEKVersion = k.active
	s.WrappedDEK = wrapped

	return s, nil
}

// wrapDEK seals a DEK under the KEK of the given version, prefixing the nonce so
// the result is a single column value. The version goes in as AAD, which is what
// prevents a stored row from claiming a different (say, retired) KEK version.
func (k *Keyring) wrapDEK(dek []byte, version int) ([]byte, error) {
	kek, ok := k.keys[version]
	if !ok {
		return nil, fmt.Errorf("%w: %d", ErrUnknownKEK, version)
	}

	kekAEAD, err := newAEAD(kek)
	if err != nil {
		return nil, err
	}

	nonce, err := newNonce(kekAEAD.NonceSize())
	if err != nil {
		return nil, err
	}

	return kekAEAD.Seal(nonce, nonce, dek, dekAAD(version)), nil
}

// unwrapDEK is wrapDEK's inverse; the caller owns the returned DEK and must Zero it.
func (k *Keyring) unwrapDEK(wrapped []byte, version int) ([]byte, error) {
	kek, ok := k.keys[version]
	if !ok {
		return nil, fmt.Errorf("%w: %d", ErrUnknownKEK, version)
	}

	kekAEAD, err := newAEAD(kek)
	if err != nil {
		return nil, err
	}

	nonceSize := kekAEAD.NonceSize()
	if len(wrapped) < nonceSize {
		return nil, ErrMalformedSealed
	}

	dek, err := kekAEAD.Open(nil, wrapped[:nonceSize], wrapped[nonceSize:], dekAAD(version))
	if err != nil {
		return nil, ErrDecrypt
	}

	return dek, nil
}

// AAD builds the associated data that binds a sealed secret to its owner. Pass
// the same parts, in the same order, to Seal and Open — for a market credential
// that is the user ID and the provider slug.
func AAD(parts ...string) []byte {
	return []byte(strings.Join(parts, aadSep))
}

// Zero overwrites b in place. Used on DEKs and on decrypted secrets once the
// call that needed them has returned, so they do not linger in reusable memory.
func Zero(b []byte) { clear(b) }

func dekAAD(version int) []byte {
	return []byte(dekAADPrefix + strconv.Itoa(version))
}

func newAEAD(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("secretbox: new cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("secretbox: new GCM: %w", err)
	}

	return aead, nil
}

func newNonce(size int) ([]byte, error) {
	nonce := make([]byte, size)
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("secretbox: generate nonce: %w", err)
	}

	return nonce, nil
}
