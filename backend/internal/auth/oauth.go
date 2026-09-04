package auth

// The OAuth 2.1 authorization server that fronts /mcp: the domain types, the
// token and code formats, and the rules that decide what a request is allowed
// to be — none of which the repository or the handler should be deciding for
// themselves.
//
// Why this exists at all, when /mcp already takes a personal access token: a
// remote MCP connector has nowhere to put one. It is software the user did not
// install and cannot configure, so the only credential it can obtain is one it
// negotiates for itself. That negotiation is this file.

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net/url"
	"slices"
	"strings"
	"time"

	"uuid"
)

const (
	// oauthAccessPrefix and oauthRefreshPrefix mark a bearer token as an OAuth
	// one, the way mcpTokenPrefix marks a personal access token. The guard
	// picks a store by prefix, so the two schemes never fall through to each
	// other and a rejected token is reported against the scheme it claimed.
	oauthAccessPrefix  = "fnx_oat_"
	oauthRefreshPrefix = "fnx_ort_"
	// oauthTokenBytes is the entropy behind each prefix: 256 bits, matching
	// every other credential this module mints.
	oauthTokenBytes = 32
	// oauthClientIDBytes is smaller because a client id is an identifier, not
	// a secret — it travels in URLs, logs and the consent screen. It is random
	// only so it cannot be guessed into a collision.
	oauthClientIDBytes = 16
)

const (
	// MCPScope is the only scope this server issues. It is read-only because
	// the MCP tool surface is, and naming it explicitly is what lets a future
	// write scope be a *different* consent rather than a silent widening of
	// this one.
	MCPScope = "mcp:read"

	// oauthCodeChallengeS256 is the only PKCE method accepted. OAuth 2.1 drops
	// "plain" entirely: it proves nothing an interceptor of the request could
	// not also produce.
	oauthCodeChallengeS256 = "S256"

	// The verifier length bounds come from RFC 7636 §4.1. They are checked
	// rather than assumed because a client sending a short verifier is a client
	// whose PKCE is not protecting anything.
	oauthMinVerifierLen = 43
	oauthMaxVerifierLen = 128
)

const (
	// oauthRequestTTL is how long a parked consent waits for the user. It has
	// to outlast a login — the user may arrive at the consent screen logged
	// out — without leaving an approvable request lying around all day.
	oauthRequestTTL = 15 * time.Minute
	// oauthCodeTTL bounds the window between the redirect and the token call.
	// The client redeems immediately; this is slack for a slow network, not a
	// window anyone is meant to sit in. RFC 6749 asks for ten minutes at most.
	oauthCodeTTL = 5 * time.Minute
	// oauthAccessTTL is deliberately short. The refresh token is what gives a
	// connector its long life, and it is the one a revocation reaches, so the
	// access token's only job is to keep the guard from hitting the grant row's
	// revocation state more often than this.
	oauthAccessTTL = time.Hour
	// oauthRefreshTTL is the real lifetime of a connection. Thirty days of
	// silence and the user reconnects; any use inside the window rolls it
	// forward, so an active connector never expires.
	oauthRefreshTTL = 30 * 24 * time.Hour
	// oauthGrantTouchInterval mirrors mcpTokenTouchInterval: last_used_at
	// answers "is this connection alive?", which minutes of resolution answer
	// as well as seconds.
	oauthGrantTouchInterval = 5 * time.Minute
)

// maxOAuthClientNameLen mirrors oauth_clients.client_name VARCHAR(120).
const maxOAuthClientNameLen = 120

// oauthError is a failure in the shape the OAuth specs require: a machine
// -readable code, a human-readable description, and the status that carries
// them. It is not an httpx error because these responses do not travel in the
// httpx envelope — RFC 6749 §5.2 defines the body, the same way JSON-RPC
// defines /mcp's, and a client parses it by that shape or not at all.
type oauthError struct {
	Code        string
	Description string
	Status      int
	// Redirectable marks the errors that must be delivered to the client's
	// redirect_uri rather than rendered. The distinction is the security-
	// relevant part of /authorize: an error about *which* URI to use can never
	// be sent to that URI, or an attacker learns whether a client id exists by
	// pointing it at a URI they control.
	Redirectable bool
}

func (e *oauthError) Error() string { return e.Code + ": " + e.Description }

// The error codes this server can answer with, from RFC 6749 §4.1.2.1 and
// §5.2, RFC 7591 §3.2.2, and RFC 8707 §2.
func oauthInvalidRequest(desc string) *oauthError {
	return &oauthError{Code: "invalid_request", Description: desc, Status: 400, Redirectable: true}
}

func oauthInvalidClient(desc string) *oauthError {
	// 401, and the handler adds WWW-Authenticate when the client tried to
	// authenticate: this is the one token-endpoint failure that is about who
	// the caller is rather than what they asked for.
	return &oauthError{Code: "invalid_client", Description: desc, Status: 401}
}

func oauthInvalidGrant(desc string) *oauthError {
	return &oauthError{Code: "invalid_grant", Description: desc, Status: 400}
}

func oauthInvalidScope(desc string) *oauthError {
	return &oauthError{Code: "invalid_scope", Description: desc, Status: 400, Redirectable: true}
}

func oauthAccessDenied(desc string) *oauthError {
	return &oauthError{Code: "access_denied", Description: desc, Status: 403, Redirectable: true}
}

func oauthUnsupportedGrantType(desc string) *oauthError {
	return &oauthError{Code: "unsupported_grant_type", Description: desc, Status: 400}
}

func oauthUnsupportedResponseType(desc string) *oauthError {
	return &oauthError{Code: "unsupported_response_type", Description: desc, Status: 400, Redirectable: true}
}

func oauthInvalidClientMetadata(desc string) *oauthError {
	return &oauthError{Code: "invalid_client_metadata", Description: desc, Status: 400}
}

func oauthInvalidRedirectURI(desc string) *oauthError {
	return &oauthError{Code: "invalid_redirect_uri", Description: desc, Status: 400}
}

func oauthInvalidTarget(desc string) *oauthError {
	return &oauthError{Code: "invalid_target", Description: desc, Status: 400, Redirectable: true}
}

func oauthServerError(desc string) *oauthError {
	return &oauthError{Code: "server_error", Description: desc, Status: 500, Redirectable: true}
}

// asOAuthError recovers the typed error, so a handler can answer in the OAuth
// shape without every layer below it having to return one.
func asOAuthError(err error) (*oauthError, bool) {
	var e *oauthError

	return e, errors.As(err, &e)
}

// ErrOAuthRequestNotFound is a consent id that names nothing — expired, already
// used, or invented. The consent screen shows it as "this request is no longer
// valid", which is all any of the three deserve.
var ErrOAuthRequestNotFound = errors.New("authorization request not found or expired")

// ErrOAuthGrantNotFound is a grant that is not this user's, which is also what
// a caller guessing ids gets.
var ErrOAuthGrantNotFound = errors.New("connected application not found")

// ErrOAuthClientNotFound is an unregistered client id. It never reaches the
// caller as itself: /authorize renders it and /token answers invalid_client,
// because saying which of the two a client got wrong is how an attacker
// enumerates registrations.
var ErrOAuthClientNotFound = errors.New("oauth client not found")

// ErrOAuthCodeNotFound is a code hash nothing matches, kept distinct from a
// code that was already spent — the second is a replay and is answered far
// more aggressively. See ConsumeAuthorizationCode.
var ErrOAuthCodeNotFound = errors.New("authorization code not found")

// ErrInvalidOAuthToken is every access-token failure the guard can hit:
// unknown, expired, revoked, or belonging to a banned account. One error for
// all four, so the response cannot be used to tell live tokens from dead ones —
// the same reason ErrInvalidMCPToken is one error.
var ErrInvalidOAuthToken = errors.New("invalid oauth token")

// oauthClient is a registered client, as the flow needs it.
type oauthClient struct {
	ClientID          string
	SecretHash        string
	Name              string
	RedirectURIs      []string
	GrantTypes        []string
	ResponseTypes     []string
	Scope             string
	TokenEndpointAuth string
	ClientURI         string
	LogoURI           string
	CreatedAt         time.Time
}

// isPublic reports whether this client authenticates with PKCE alone. A client
// that shipped no secret cannot be asked for one later.
func (c oauthClient) isPublic() bool {
	return c.SecretHash == ""
}

// allowsRedirectURI matches literally. Not by prefix, not after normalising
// the path, not ignoring the query: every open redirect an OAuth server has
// ever had came from one of those three relaxations.
func (c oauthClient) allowsRedirectURI(uri string) bool {
	return slices.Contains(c.RedirectURIs, uri)
}

// OAuthClientRegistration is the RFC 7591 request body, and the subset of it
// this server honours. Everything else a client sends is ignored rather than
// refused: registration failing over metadata nobody reads would break clients
// for no gain.
type OAuthClientRegistration struct {
	ClientName              string   `json:"client_name"`
	RedirectURIs            []string `json:"redirect_uris"`
	GrantTypes              []string `json:"grant_types"`
	ResponseTypes           []string `json:"response_types"`
	Scope                   string   `json:"scope"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
	ClientURI               string   `json:"client_uri"`
	LogoURI                 string   `json:"logo_uri"`
}

// OAuthClientCredentials is the registration response. The secret appears here
// and nowhere else, and only for a confidential client.
type OAuthClientCredentials struct {
	ClientID              string   `json:"client_id"`
	ClientSecret          string   `json:"client_secret,omitempty"`
	ClientIDIssuedAt      int64    `json:"client_id_issued_at"`
	ClientSecretExpiresAt int64    `json:"client_secret_expires_at"`
	ClientName            string   `json:"client_name"`
	RedirectURIs          []string `json:"redirect_uris"`
	GrantTypes            []string `json:"grant_types"`
	ResponseTypes         []string `json:"response_types"`
	Scope                 string   `json:"scope"`
	TokenEndpointAuth     string   `json:"token_endpoint_auth_method"`
}

// OAuthAuthorizationRequest is one parsed /authorize call, before anything has
// been decided about it.
type OAuthAuthorizationRequest struct {
	ClientID            string
	RedirectURI         string
	ResponseType        string
	Scope               string
	State               string
	CodeChallenge       string
	CodeChallengeMethod string
	Resource            string
}

// OAuthConsent is what the consent screen renders: who is asking, for what,
// and nothing the browser could tamper with into meaning something else. The
// redirect URI is included so a user can see where they are about to be sent.
type OAuthConsent struct {
	RequestID   uuid.UUID `json:"requestId"`
	ClientName  string    `json:"clientName"`
	ClientURI   string    `json:"clientUri,omitempty"`
	LogoURI     string    `json:"logoUri,omitempty"`
	RedirectURI string    `json:"redirectUri"`
	Scopes      []string  `json:"scopes"`
	ExpiresAt   time.Time `json:"expiresAt"`
}

// pendingAuthorization is the stored form of a parked consent.
type pendingAuthorization struct {
	ID                  uuid.UUID
	ClientID            string
	RedirectURI         string
	Scope               string
	State               string
	CodeChallenge       string
	CodeChallengeMethod string
	Resource            string
	ExpiresAt           time.Time
}

// authorizationCode is a minted code, read back at the token endpoint.
type authorizationCode struct {
	ID                  uuid.UUID
	ClientID            string
	UserID              uuid.UUID
	RedirectURI         string
	Scope               string
	CodeChallenge       string
	CodeChallengeMethod string
	Resource            string
	ConsumedAt          *time.Time
	ExpiresAt           time.Time
}

// OAuthTokens is the RFC 6749 §5.1 token response. The field names are the
// spec's, not this codebase's, because a client parses them by name.
type OAuthTokens struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope"`
}

// OAuthGrant is a connected application as the settings screen lists it. It
// carries no token material, for the same reason MCPToken has no secret field.
type OAuthGrant struct {
	ID         uuid.UUID  `json:"id"`
	ClientName string     `json:"clientName"`
	ClientURI  string     `json:"clientUri,omitempty"`
	Scopes     []string   `json:"scopes"`
	LastUsedAt *time.Time `json:"lastUsedAt"`
	CreatedAt  time.Time  `json:"createdAt"`
}

// oauthGrantIdentity is what authenticating an access token yields: the same
// three things the MCP guard needs from any credential, plus the grant to
// touch. Unexported so a row carrying token state cannot reach a DTO.
type oauthGrantIdentity struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	Role       string
	Scope      string
	ExpiresAt  time.Time
	LastUsedAt *time.Time
}

// grantRefresh is the state a refresh call reads back before rotating it.
type grantRefresh struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	ClientID         string
	Scope            string
	Resource         string
	RefreshExpiresAt *time.Time
}

// generateOAuthToken mints one credential: the raw secret to hand out and the
// hash to store. Unpadded URL encoding for the same reason the personal access
// tokens use it — these are pasted and logged, and "=" and "/" invite quoting
// mistakes.
func generateOAuthToken(prefix string) (raw, hash string, err error) {
	b := make([]byte, oauthTokenBytes)
	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}

	raw = prefix + base64.RawURLEncoding.EncodeToString(b)

	return raw, hashOAuthToken(raw), nil
}

// hashOAuthToken is the stored form: the whole raw string, prefix included, so
// what is looked up is exactly what the client sent.
func hashOAuthToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))

	return hex.EncodeToString(sum[:])
}

// generateClientID mints the public identifier a registration answers with.
func generateClientID() (string, error) {
	b := make([]byte, oauthClientIDBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return "fnx_client_" + base64.RawURLEncoding.EncodeToString(b), nil
}

// looksLikeOAuthToken reports whether a bearer token claims to be an OAuth
// access token. A claim, not a verdict: only the store can say it is real.
func looksLikeOAuthToken(raw string) bool {
	return strings.HasPrefix(raw, oauthAccessPrefix)
}

// verifyPKCE checks a code verifier against the challenge stored with the code.
//
// The comparison is constant-time even though both sides are public-ish values:
// the challenge is stored server-side and the verifier is the secret that
// proves the caller started the flow, so a timing oracle here is a timing
// oracle on the one thing standing between an intercepted code and a token.
func verifyPKCE(verifier, challenge, method string) error {
	switch {
	case method != oauthCodeChallengeS256:
		return oauthInvalidGrant("unsupported code_challenge_method")
	case len(verifier) < oauthMinVerifierLen || len(verifier) > oauthMaxVerifierLen:
		return oauthInvalidGrant("code_verifier must be between 43 and 128 characters")
	}

	sum := sha256.Sum256([]byte(verifier))
	computed := base64.RawURLEncoding.EncodeToString(sum[:])

	if subtle.ConstantTimeCompare([]byte(computed), []byte(challenge)) != 1 {
		return oauthInvalidGrant("code_verifier does not match the challenge")
	}

	return nil
}

// verifyClientSecret compares a presented secret against the stored hash in
// constant time.
func verifyClientSecret(presented, storedHash string) bool {
	return subtle.ConstantTimeCompare([]byte(hashOAuthToken(presented)), []byte(storedHash)) == 1
}

// validateRedirectURI decides whether a URI may receive an authorization code.
//
// The three shapes allowed are the three RFC 8252 and OAuth 2.1 recognise: an
// https URL for a server-side or hosted client, a loopback http URL for a
// native client that listens on an ephemeral port, and a private-use scheme for
// a native client the OS routes by scheme. Plain http to anywhere else would
// hand the code to whoever is on the network path.
//
// A fragment is refused outright: the code is appended to the query, and a URI
// that already has a fragment cannot receive one unambiguously.
func validateRedirectURI(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return oauthInvalidRedirectURI("redirect_uri is not a valid URI: " + raw)
	}

	if u.Fragment != "" || strings.Contains(raw, "#") {
		return oauthInvalidRedirectURI("redirect_uri must not contain a fragment: " + raw)
	}

	switch u.Scheme {
	case "https":
		if u.Host == "" {
			return oauthInvalidRedirectURI("redirect_uri must name a host: " + raw)
		}

		return nil
	case "http":
		if !isLoopbackHost(u.Hostname()) {
			return oauthInvalidRedirectURI("http redirect_uri is only allowed on loopback: " + raw)
		}

		return nil
	case "":
		return oauthInvalidRedirectURI("redirect_uri must be absolute: " + raw)
	default:
		// A private-use scheme (com.example.app:/callback). The OS decides who
		// receives it, so there is nothing for this server to check beyond it
		// being a scheme at all.
		return nil
	}
}

func isLoopbackHost(host string) bool {
	return host == "localhost" || host == "127.0.0.1" || host == "::1"
}

// normalizeScope reduces a requested scope string to what this server issues.
//
// An empty request means "the default", which is the only scope there is. A
// request naming anything else is refused rather than silently narrowed: a
// client that asked for more than it gets should find out now, not when a tool
// call fails.
func normalizeScope(requested string) (string, error) {
	if strings.TrimSpace(requested) == "" {
		return MCPScope, nil
	}

	for _, s := range strings.Fields(requested) {
		if s != MCPScope {
			return "", oauthInvalidScope("unsupported scope: " + s)
		}
	}

	return MCPScope, nil
}

// scopeList splits a stored scope string for presentation.
func scopeList(scope string) []string {
	return strings.Fields(scope)
}

// normalizeResource applies RFC 8707 to the resource indicator a client sends.
//
// The check is what keeps a token minted for this server from being replayed
// against another one: the client says which resource it wants, and this server
// only signs off on itself. A missing indicator is accepted — older clients do
// not send one — but a wrong one is refused rather than ignored.
func normalizeResource(requested, canonical string) (string, error) {
	if strings.TrimSpace(requested) == "" {
		return canonical, nil
	}

	if trimTrailingSlash(requested) != trimTrailingSlash(canonical) {
		return "", oauthInvalidTarget("this authorization server issues tokens for " + canonical + " only")
	}

	return canonical, nil
}

func trimTrailingSlash(s string) string {
	return strings.TrimSuffix(s, "/")
}

// normalizeClientName trims what the consent screen will show and refuses what
// the column cannot hold. A client that registers without a name gets one, so
// the screen never has to render an empty attribution.
func normalizeClientName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "Unnamed MCP client", nil
	}

	if len([]rune(name)) > maxOAuthClientNameLen {
		return "", oauthInvalidClientMetadata("client_name is too long")
	}

	return name, nil
}

// buildRedirect appends the authorization response to a client's redirect URI,
// preserving any query the client registered.
func buildRedirect(redirectURI string, params map[string]string) (string, error) {
	u, err := url.Parse(redirectURI)
	if err != nil {
		return "", err
	}

	q := u.Query()
	for k, v := range params {
		if v != "" {
			q.Set(k, v)
		}
	}

	u.RawQuery = q.Encode()

	return u.String(), nil
}
