package auth

// Personal access tokens for the MCP endpoint: the domain types, the token
// format and the rules about it that neither the repository nor the handler
// should be deciding for themselves.

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"uuid"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

const (
	// mcpTokenPrefix marks a bearer token as one of these rather than a JWT.
	//
	// It is what lets one guard serve both credentials without trying each in
	// turn: the prefix decides which path a request takes, so a malformed token
	// fails once, against the scheme it claimed, instead of being reported as
	// whichever check happened to run last. It is also what makes a leaked
	// token greppable — secret scanners match on prefixes, not on entropy.
	mcpTokenPrefix = "fnx_mcp_"
	// mcpTokenBytes is the entropy behind the prefix. 256 bits, the same as a
	// refresh token: these live far longer, so they get no less.
	mcpTokenBytes = 32
	// maxMCPTokensPerUser bounds how many a single account can hold. It is not
	// a security boundary — every one of them is the same user's — but a
	// forgotten token is a live credential, and a list nobody can read is a
	// list nobody revokes from.
	maxMCPTokensPerUser = 10
	// maxMCPTokenNameLen mirrors mcp_tokens.name VARCHAR(60).
	maxMCPTokenNameLen = 60
	// mcpTokenTouchInterval is how stale last_used_at is allowed to get before
	// the guard writes it again. Every MCP call would otherwise be a write, and
	// the column exists to answer "is this token still in use?", which minutes
	// of resolution answer as well as seconds.
	mcpTokenTouchInterval = 5 * time.Minute
)

// The lifetime is expressed in days everywhere it crosses a boundary — the
// request body, the service signature, this file — because that is the unit the
// user picks in and the only one that survives the round trip unambiguously.
const (
	// DefaultMCPTokenExpiryDays applies when the caller names no lifetime.
	// Ninety days is short enough that an abandoned token stops working on its
	// own, and long enough not to be a monthly chore.
	DefaultMCPTokenExpiryDays = 90
	// MaxMCPTokenExpiryDays caps an explicit lifetime. A token meant to outlive
	// this is what "no expiry" is for: unbounded should be a choice the user
	// made, not the accumulated result of picking a large number.
	MaxMCPTokenExpiryDays = 365
)

var (
	// ErrMCPTokenNotFound is returned for a token that is not this user's,
	// which is also what a caller guessing ids gets.
	ErrMCPTokenNotFound = errors.New("mcp token not found")
	// ErrMCPTokenNameTaken keeps the list readable; see uq_mcp_tokens_user_name.
	ErrMCPTokenNameTaken = errors.New("you already have a token with that name")
	// ErrTooManyMCPTokens is maxMCPTokensPerUser, refused.
	ErrTooManyMCPTokens = errors.New("you have reached the limit of MCP tokens; delete one before creating another")
	// ErrInvalidMCPToken is every authentication failure: unknown, expired, or
	// belonging to a deleted account. The caller is told no more than that.
	ErrInvalidMCPToken = errors.New("invalid mcp token")

	// The input rules, as sentinels rather than errors built at the return
	// site, so a test can match the rule that fired instead of its wording.
	ErrMCPTokenNameRequired  = errors.New("the token needs a name")
	ErrMCPTokenNameTooLong   = errors.New("the name is too long")
	ErrMCPTokenExpiryInvalid = errors.New("the lifetime cannot be negative")
	ErrMCPTokenExpiryTooLong = errors.New("the maximum lifetime is 365 days; choose no expiry instead")
)

// MCPToken is the projection of a token that may leave the service layer. It
// has no field capable of carrying the secret — that absence is the point, and
// it is what keeps a token from being served back to anyone, its owner
// included, after the one moment it is shown.
type MCPToken struct {
	ID         uuid.UUID  `json:"id"`
	Name       string     `json:"name"`
	Last4      string     `json:"last4"`
	ExpiresAt  *time.Time `json:"expiresAt"`
	LastUsedAt *time.Time `json:"lastUsedAt"`
	RotatedAt  *time.Time `json:"rotatedAt"`
	CreatedAt  time.Time  `json:"createdAt"`
	// Expired is derived, not stored: an expired token stays in the list until
	// the user deletes it or the cleanup job does, and the list is where they
	// find out it stopped working.
	Expired bool `json:"expired"`
}

// MCPTokenSecret is the only shape that carries the secret. It is returned by
// exactly two use cases — creation and rotation — and never persisted, cached
// or logged.
type MCPTokenSecret struct {
	MCPToken
	// Token is the raw credential, in the form the client sends it back:
	// Authorization: Bearer fnx_mcp_…
	Token string `json:"token"`
}

// mcpTokenIdentity is what authenticating a token yields: the account behind
// it and the two fields the guard needs to decide anything. It stays
// unexported so a row carrying a token's identity cannot end up in a DTO.
type mcpTokenIdentity struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	Role       string
	ExpiresAt  *time.Time
	LastUsedAt *time.Time
}

// generateMCPToken mints a token and returns it three ways: the raw secret to
// hand the user once, the hash to store, and the fragment to display.
func generateMCPToken() (raw, hash, last4 string, err error) {
	b := make([]byte, mcpTokenBytes)
	if _, err = rand.Read(b); err != nil {
		return "", "", "", err
	}

	// Unpadded URL encoding: the token is pasted into JSON config files and
	// shell environments, where "=" and "/" invite quoting mistakes.
	secret := base64.RawURLEncoding.EncodeToString(b)
	raw = mcpTokenPrefix + secret

	return raw, hashMCPToken(raw), secret[len(secret)-4:], nil
}

// hashMCPToken is the stored form. The whole raw token is hashed, prefix
// included, so what the guard looks up is exactly the string the client sent
// and there is no decoding step in front of the lookup that could fail.
func hashMCPToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))

	return hex.EncodeToString(sum[:])
}

// looksLikeMCPToken reports whether a bearer token claims to be one of these.
// It is a claim, not a verdict: only the store can say whether it is real.
func looksLikeMCPToken(raw string) bool {
	return strings.HasPrefix(raw, mcpTokenPrefix)
}

// normalizeMCPTokenName trims a name and refuses what the column cannot hold.
// Returning a 400 here rather than letting Postgres reject the insert is what
// keeps the error the user reads about their input.
func normalizeMCPTokenName(name string) (string, error) {
	name = strings.TrimSpace(name)

	switch {
	case name == "":
		return "", httpx.AsBadRequest(ErrMCPTokenNameRequired)
	case len([]rune(name)) > maxMCPTokenNameLen:
		return "", httpx.AsBadRequest(ErrMCPTokenNameTooLong)
	}

	return name, nil
}

// resolveMCPTokenExpiry turns the requested number of days into the absolute
// instant the row stores. Zero means the caller asked for a token that does
// not expire; a negative or oversized value is the caller's mistake.
func resolveMCPTokenExpiry(days int, now time.Time) (*time.Time, error) {
	switch {
	case days < 0:
		return nil, httpx.AsBadRequest(ErrMCPTokenExpiryInvalid)
	case days == 0:
		return nil, nil
	case days > MaxMCPTokenExpiryDays:
		return nil, httpx.AsBadRequest(ErrMCPTokenExpiryTooLong)
	}

	at := now.Add(time.Duration(days) * 24 * time.Hour)

	return &at, nil
}

// mcpTokenDomainError tags the store's sentinels with the status they mean.
// The repository returns them bare so it stays a persistence layer; the
// mapping belongs here, next to the rules those errors enforce.
func mcpTokenDomainError(err error) error {
	switch {
	case errors.Is(err, ErrMCPTokenNotFound):
		return httpx.AsNotFound(err)
	case errors.Is(err, ErrMCPTokenNameTaken):
		return httpx.AsConflict(err)
	default:
		return err
	}
}

// isMCPTokenNotFound separates the ordinary authentication failure — a hash
// nothing matches — from a store that actually broke.
func isMCPTokenNotFound(err error) bool {
	return errors.Is(err, ErrMCPTokenNotFound)
}
