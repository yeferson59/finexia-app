package auth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"uuid"

	"github.com/gofiber/fiber/v3"

	"github.com/yeferson59/finexia-app/internal/identity"
	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

// The redirect URI these tests register. It is https and exact, which is all
// the server will accept for a hosted client.
const testRedirectURI = "https://claude.test/api/mcp/auth_callback"

// oauthHarness is one mounted module and everything a case needs to drive it:
// the router, the service behind it (for asserting on a token the HTTP surface
// only hands out), a session token the consent routes accept, the user it
// belongs to, and the store, for state no endpoint exposes.
//
// A struct rather than five return values because most cases want two of them,
// and five positional results turn every one of those into a row of blanks.
type oauthHarness struct {
	app     *fiber.App
	svc     *service
	session string
	userID  uuid.UUID
	store   *oauthMemStore
}

// oauthTestApp mounts the module over a real in-memory OAuth store.
func oauthTestApp(t *testing.T) oauthHarness {
	t.Helper()

	cfg := testConfig()
	userID := uuid.New()
	store := newOAuthMemStore()
	repo := store.attach(new(fakeRepository{}))

	svc := newService(testStores(repo), cfg, newMemStorage(), nil, nil, logger.Noop())

	token, err := svc.CreateJWToken(userID, "user", time.Now().UTC().Add(time.Hour))
	if err != nil {
		t.Fatalf("CreateJWToken: %v", err)
	}

	repo.getSessionByToken = func(context.Context, string) (identity.User, error) {
		return identity.User{
			ID:       userID,
			Role:     identity.Role{Name: "user"},
			Sessions: []identity.Session{{Token: token}},
		}, nil
	}

	m := newModule(Deps{Cfg: cfg, Storage: newMemStorage(), Log: logger.Noop()}, svc)

	app := fiber.New()
	m.Routes(app)

	return oauthHarness{app: app, svc: svc, session: token, userID: userID, store: store}
}

// pkcePair returns a verifier and the S256 challenge derived from it.
func pkcePair() (verifier, challenge string) {
	verifier = "finexia-test-verifier-0123456789-0123456789-0123456789"
	sum := sha256.Sum256([]byte(verifier))

	return verifier, base64.RawURLEncoding.EncodeToString(sum[:])
}

// registerTestClient walks RFC 7591 and returns the issued client id.
func registerTestClient(t *testing.T, app *fiber.App) string {
	t.Helper()

	body := `{"client_name":"Claude","redirect_uris":["` + testRedirectURI + `"],` +
		`"grant_types":["authorization_code","refresh_token"],"response_types":["code"],` +
		`"token_endpoint_auth_method":"none"}`

	req := httptest.NewRequest("POST", "/oauth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("POST /oauth/register: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != fiber.StatusCreated {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("register status = %d, want 201 (%s)", resp.StatusCode, raw)
	}

	var out OAuthClientCredentials
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decoding the registration: %v", err)
	}

	if out.ClientID == "" {
		t.Fatal("registration returned no client_id")
	}

	// A client that asked for "none" must not be handed a secret it would then
	// have to keep.
	if out.ClientSecret != "" {
		t.Error("a public client was issued a client_secret")
	}

	return out.ClientID
}

// authorizeURL builds one /authorize call.
func authorizeURL(clientID, challenge string, overrides map[string]string) string {
	q := url.Values{
		"response_type":         {"code"},
		"client_id":             {clientID},
		"redirect_uri":          {testRedirectURI},
		"scope":                 {MCPScope},
		"state":                 {"opaque-client-state"},
		"code_challenge":        {challenge},
		"code_challenge_method": {"S256"},
		"resource":              {"https://api.finexia.test/mcp"},
	}

	for k, v := range overrides {
		if v == "" {
			q.Del(k)

			continue
		}

		q.Set(k, v)
	}

	return "/oauth/authorize?" + q.Encode()
}

// follow performs a request and returns the status plus the Location header,
// without following the redirect.
func follow(t *testing.T, app *fiber.App, target string) (int, string) {
	t.Helper()

	resp, err := app.Test(httptest.NewRequest("GET", target, nil))
	if err != nil {
		t.Fatalf("GET %s: %v", target, err)
	}
	defer func() { _ = resp.Body.Close() }()

	return resp.StatusCode, resp.Header.Get("Location")
}

// postForm performs a form-encoded POST, the shape /token requires.
func postForm(t *testing.T, app *fiber.App, target string, form url.Values) (int, map[string]any) {
	t.Helper()

	req := httptest.NewRequest("POST", target, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("POST %s: %v", target, err)
	}
	defer func() { _ = resp.Body.Close() }()

	var out map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&out)

	return resp.StatusCode, out
}

// approve walks the consent screen's two calls and returns the redirect the
// browser is sent to.
func approve(t *testing.T, app *fiber.App, sessionToken, requestID string, approved bool) string {
	t.Helper()

	status, envelope := callMCP(t, app, sessionToken, "GET", "/auth/oauth/consent/"+requestID, "")
	if status != fiber.StatusOK {
		t.Fatalf("GET consent = %d, want 200 (%v)", status, envelope)
	}

	body := `{"approved":false}`
	if approved {
		body = `{"approved":true}`
	}

	status, envelope = callMCP(t, app, sessionToken, "POST", "/auth/oauth/consent/"+requestID, body)
	if status != fiber.StatusOK {
		t.Fatalf("POST consent = %d, want 200 (%v)", status, envelope)
	}

	data, _ := envelope["data"].(map[string]any)
	target, _ := data["redirectTo"].(string)

	if target == "" {
		t.Fatalf("consent returned no redirectTo: %v", envelope)
	}

	return target
}

// requestIDFrom extracts the consent id out of the /authorize redirect.
func requestIDFrom(t *testing.T, consentURL string) string {
	t.Helper()

	u, err := url.Parse(consentURL)
	if err != nil {
		t.Fatalf("parsing the consent URL %q: %v", consentURL, err)
	}

	id := u.Query().Get("request")
	if id == "" {
		t.Fatalf("no request id in %q", consentURL)
	}

	return id
}

// TestOAuthDiscoveryDocuments: the two documents a client reads before it does
// anything else. Their contents are the contract — a client configures itself
// from them once and never asks again — so this asserts the endpoints they
// advertise are the endpoints that exist.
func TestOAuthDiscoveryDocuments(t *testing.T) {
	h := oauthTestApp(t)
	app := h.app

	t.Run("protected resource metadata points at the authorization server", func(t *testing.T) {
		var doc OAuthProtectedResourceMetadata
		getJSON(t, app, "/.well-known/oauth-protected-resource/mcp", &doc)

		if doc.Resource != "https://api.finexia.test/mcp" {
			t.Errorf("resource = %q", doc.Resource)
		}

		if len(doc.AuthorizationServers) != 1 || doc.AuthorizationServers[0] != "https://api.finexia.test" {
			t.Errorf("authorization_servers = %v", doc.AuthorizationServers)
		}

		// A token in a query string ends up in access logs and Referer headers.
		if len(doc.BearerMethodsSupported) != 1 || doc.BearerMethodsSupported[0] != "header" {
			t.Errorf("bearer_methods_supported = %v, want [header] only", doc.BearerMethodsSupported)
		}
	})

	t.Run("the bare protected-resource path answers too", func(t *testing.T) {
		var doc OAuthProtectedResourceMetadata
		getJSON(t, app, "/.well-known/oauth-protected-resource", &doc)

		if doc.Resource == "" {
			t.Error("the bare path served an empty document")
		}
	})

	t.Run("authorization server metadata advertises only what is enforced", func(t *testing.T) {
		var doc OAuthServerMetadata
		getJSON(t, app, "/.well-known/oauth-authorization-server", &doc)

		if doc.Issuer != "https://api.finexia.test" {
			t.Errorf("issuer = %q", doc.Issuer)
		}

		for field, got := range map[string]string{
			"authorization_endpoint": doc.AuthorizationEndpoint,
			"token_endpoint":         doc.TokenEndpoint,
			"registration_endpoint":  doc.RegistrationEndpoint,
		} {
			if !strings.HasPrefix(got, doc.Issuer+"/oauth/") {
				t.Errorf("%s = %q, want it under the issuer", field, got)
			}
		}

		// OAuth 2.1 removed "plain", and StartAuthorization refuses it. The
		// document must not claim otherwise.
		if len(doc.CodeChallengeMethodsSupported) != 1 || doc.CodeChallengeMethodsSupported[0] != "S256" {
			t.Errorf("code_challenge_methods_supported = %v, want [S256] only", doc.CodeChallengeMethodsSupported)
		}
	})
}

func getJSON(t *testing.T, app *fiber.App, target string, out any) {
	t.Helper()

	resp, err := app.Test(httptest.NewRequest("GET", target, nil))
	if err != nil {
		t.Fatalf("GET %s: %v", target, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("GET %s = %d, want 200", target, resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		t.Fatalf("decoding %s: %v", target, err)
	}
}

// TestOAuthAuthorizationCodeFlow walks the whole thing, in the order a real
// client does it, and ends where the point of it is: an access token the /mcp
// guard accepts for the user who approved it.
func TestOAuthAuthorizationCodeFlow(t *testing.T) {
	h := oauthTestApp(t)
	app, svc, sessionToken, userID, store := h.app, h.svc, h.session, h.userID, h.store

	clientID := registerTestClient(t, app)
	verifier, challenge := pkcePair()

	status, location := follow(t, app, authorizeURL(clientID, challenge, nil))
	if status != fiber.StatusFound {
		t.Fatalf("/authorize = %d, want 302", status)
	}

	if !strings.HasPrefix(location, "https://finexia.test/oauth/consent?request=") {
		t.Fatalf("/authorize sent the browser to %q, want the consent screen", location)
	}

	redirect := approve(t, app, sessionToken, requestIDFrom(t, location), true)

	u, err := url.Parse(redirect)
	if err != nil {
		t.Fatalf("parsing the approval redirect: %v", err)
	}

	if u.Scheme+"://"+u.Host+u.Path != testRedirectURI {
		t.Errorf("approved redirect went to %q, want the registered URI", redirect)
	}

	// The client's CSRF token comes back untouched, which is the only way it
	// can match the request it started.
	if got := u.Query().Get("state"); got != "opaque-client-state" {
		t.Errorf("state = %q, want it echoed back", got)
	}

	code := u.Query().Get("code")
	if code == "" {
		t.Fatalf("no code in %q", redirect)
	}

	status, tokens := postForm(t, app, "/oauth/token", url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {testRedirectURI},
		"client_id":     {clientID},
		"code_verifier": {verifier},
	})
	if status != fiber.StatusOK {
		t.Fatalf("/token = %d, want 200 (%v)", status, tokens)
	}

	accessToken, _ := tokens["access_token"].(string)
	refreshToken, _ := tokens["refresh_token"].(string)

	if !strings.HasPrefix(accessToken, oauthAccessPrefix) {
		t.Errorf("access_token = %q, want the %s prefix", accessToken, oauthAccessPrefix)
	}

	if tokens["token_type"] != "Bearer" || tokens["scope"] != MCPScope {
		t.Errorf("token response = %v", tokens)
	}

	// The point of all of it: the guard resolves this token to the user who
	// stood in front of the consent screen, and to nobody else.
	gotUser, role, err := svc.AuthenticateOAuthToken(context.Background(), accessToken)
	if err != nil {
		t.Fatalf("AuthenticateOAuthToken: %v", err)
	}

	if gotUser != userID || role != "user" {
		t.Errorf("token resolved to (%v, %q), want (%v, \"user\")", gotUser, role, userID)
	}

	t.Run("refreshing rotates both halves", func(t *testing.T) {
		status, refreshed := postForm(t, app, "/oauth/token", url.Values{
			"grant_type":    {"refresh_token"},
			"refresh_token": {refreshToken},
			"client_id":     {clientID},
		})
		if status != fiber.StatusOK {
			t.Fatalf("refresh = %d, want 200 (%v)", status, refreshed)
		}

		newAccess, _ := refreshed["access_token"].(string)
		if newAccess == "" || newAccess == accessToken {
			t.Error("refresh did not issue a new access token")
		}

		// The old access token must die with the rotation, or revoking a
		// connection would mean nothing for a further hour.
		if _, _, err := svc.AuthenticateOAuthToken(context.Background(), accessToken); err == nil {
			t.Error("the previous access token still authenticates after a refresh")
		}

		if _, _, err := svc.AuthenticateOAuthToken(context.Background(), newAccess); err != nil {
			t.Errorf("the rotated access token does not authenticate: %v", err)
		}
	})

	t.Run("re-approving the same client does not add a second grant", func(t *testing.T) {
		before := store.grantCount()

		_, location := follow(t, app, authorizeURL(clientID, challenge, nil))
		redirect := approve(t, app, sessionToken, requestIDFrom(t, location), true)

		u, _ := url.Parse(redirect)
		postForm(t, app, "/oauth/token", url.Values{
			"grant_type":    {"authorization_code"},
			"code":          {u.Query().Get("code")},
			"redirect_uri":  {testRedirectURI},
			"client_id":     {clientID},
			"code_verifier": {verifier},
		})

		if after := store.grantCount(); after != before {
			t.Errorf("grant count went from %d to %d; re-approving must update the grant, not add one", before, after)
		}
	})
}

// TestAuthorizeRefusesWhatItCannotRedirect covers the one decision that makes
// /authorize safe: a request whose client or redirect_uri could not be verified
// is rendered, never redirected. Redirecting either would let an attacker probe
// for registered clients using a URI they control.
func TestAuthorizeRefusesWhatItCannotRedirect(t *testing.T) {
	h := oauthTestApp(t)
	app := h.app
	clientID := registerTestClient(t, app)
	_, challenge := pkcePair()

	for _, tc := range []struct {
		name   string
		target string
	}{
		{
			name:   "unknown client",
			target: authorizeURL("fnx_client_nope", challenge, nil),
		},
		{
			name:   "unregistered redirect_uri",
			target: authorizeURL(clientID, challenge, map[string]string{"redirect_uri": "https://evil.test/steal"}),
		},
		{
			// A registered URI with anything appended is a different URI. The
			// match is literal precisely so this fails.
			name:   "registered prefix with a suffix",
			target: authorizeURL(clientID, challenge, map[string]string{"redirect_uri": testRedirectURI + "/../evil"}),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			status, location := follow(t, app, tc.target)

			if location != "" {
				t.Fatalf("answered with a redirect to %q; nothing may be sent to an unverified URI", location)
			}

			if status != fiber.StatusBadRequest {
				t.Errorf("status = %d, want 400", status)
			}
		})
	}
}

// TestAuthorizeSendsRecoverableErrorsToTheClient is the other half: once the
// client and the URI are known, the client is entitled to hear what went wrong
// — at its own redirect_uri, with its state, per RFC 6749 §4.1.2.1.
func TestAuthorizeSendsRecoverableErrorsToTheClient(t *testing.T) {
	h := oauthTestApp(t)
	app := h.app
	clientID := registerTestClient(t, app)
	_, challenge := pkcePair()

	for _, tc := range []struct {
		name      string
		overrides map[string]string
		wantError string
	}{
		{
			name:      "no PKCE challenge",
			overrides: map[string]string{"code_challenge": ""},
			wantError: "invalid_request",
		},
		{
			// OAuth 2.1 removed "plain"; accepting it would make PKCE decorative.
			name:      "plain PKCE",
			overrides: map[string]string{"code_challenge_method": "plain"},
			wantError: "invalid_request",
		},
		{
			name:      "implicit flow",
			overrides: map[string]string{"response_type": "token"},
			wantError: "unsupported_response_type",
		},
		{
			name:      "a scope this server does not issue",
			overrides: map[string]string{"scope": "mcp:write"},
			wantError: "invalid_scope",
		},
		{
			// RFC 8707: a token minted here is for this resource, so a client
			// asking for another one is refused rather than quietly served.
			name:      "a resource belonging to another server",
			overrides: map[string]string{"resource": "https://api.elsewhere.test/mcp"},
			wantError: "invalid_target",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			status, location := follow(t, app, authorizeURL(clientID, challenge, tc.overrides))

			if status != fiber.StatusFound {
				t.Fatalf("status = %d, want a 302 carrying the error", status)
			}

			u, err := url.Parse(location)
			if err != nil {
				t.Fatalf("parsing %q: %v", location, err)
			}

			if got := u.Query().Get("error"); got != tc.wantError {
				t.Errorf("error = %q, want %q (%s)", got, tc.wantError, location)
			}

			if got := u.Query().Get("state"); got != "opaque-client-state" {
				t.Errorf("state = %q, want it echoed even on an error", got)
			}

			if u.Query().Get("code") != "" {
				t.Error("a failed authorization carried a code")
			}
		})
	}
}

// TestTokenEndpointRejections covers the checks that stand between an
// intercepted code and a working token.
func TestTokenEndpointRejections(t *testing.T) {
	h := oauthTestApp(t)
	app, sessionToken := h.app, h.session
	clientID := registerTestClient(t, app)
	verifier, challenge := pkcePair()

	// mintCode runs the flow up to a fresh, unspent code.
	mintCode := func(t *testing.T) string {
		t.Helper()

		_, location := follow(t, app, authorizeURL(clientID, challenge, nil))
		redirect := approve(t, app, sessionToken, requestIDFrom(t, location), true)
		u, _ := url.Parse(redirect)

		return u.Query().Get("code")
	}

	for _, tc := range []struct {
		name      string
		form      func(code string) url.Values
		wantError string
	}{
		{
			name: "the wrong PKCE verifier",
			form: func(code string) url.Values {
				return url.Values{
					"grant_type": {"authorization_code"}, "code": {code},
					"redirect_uri": {testRedirectURI}, "client_id": {clientID},
					"code_verifier": {"another-verifier-that-is-long-enough-to-pass-length"},
				}
			},
			wantError: "invalid_grant",
		},
		{
			name: "no PKCE verifier at all",
			form: func(code string) url.Values {
				return url.Values{
					"grant_type": {"authorization_code"}, "code": {code},
					"redirect_uri": {testRedirectURI}, "client_id": {clientID},
				}
			},
			wantError: "invalid_grant",
		},
		{
			name: "a redirect_uri other than the one the code was minted for",
			form: func(code string) url.Values {
				return url.Values{
					"grant_type": {"authorization_code"}, "code": {code},
					"redirect_uri": {"https://evil.test/steal"}, "client_id": {clientID},
					"code_verifier": {verifier},
				}
			},
			wantError: "invalid_grant",
		},
		{
			name: "an unknown client",
			form: func(code string) url.Values {
				return url.Values{
					"grant_type": {"authorization_code"}, "code": {code},
					"redirect_uri": {testRedirectURI}, "client_id": {"fnx_client_nope"},
					"code_verifier": {verifier},
				}
			},
			wantError: "invalid_client",
		},
		{
			name: "a grant type this server does not have",
			form: func(string) url.Values {
				return url.Values{"grant_type": {"password"}, "client_id": {clientID}}
			},
			wantError: "unsupported_grant_type",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			status, body := postForm(t, app, "/oauth/token", tc.form(mintCode(t)))

			if status == fiber.StatusOK {
				t.Fatalf("the request succeeded: %v", body)
			}

			if got, _ := body["error"].(string); got != tc.wantError {
				t.Errorf("error = %q, want %q (%v)", got, tc.wantError, body)
			}

			if _, minted := body["access_token"]; minted {
				t.Error("a rejected token request still returned an access token")
			}
		})
	}
}

// TestAuthorizationCodeReplayRevokesTheGrant is RFC 6749 §10.5. A code offered
// twice was intercepted, so the tokens already minted from it are in doubt —
// refusing the second call and leaving the first one's tokens alive would hand
// the attacker exactly what they stole.
func TestAuthorizationCodeReplayRevokesTheGrant(t *testing.T) {
	h := oauthTestApp(t)
	app, svc, sessionToken, store := h.app, h.svc, h.session, h.store
	clientID := registerTestClient(t, app)
	verifier, challenge := pkcePair()

	_, location := follow(t, app, authorizeURL(clientID, challenge, nil))
	redirect := approve(t, app, sessionToken, requestIDFrom(t, location), true)
	u, _ := url.Parse(redirect)
	code := u.Query().Get("code")

	form := url.Values{
		"grant_type": {"authorization_code"}, "code": {code},
		"redirect_uri": {testRedirectURI}, "client_id": {clientID},
		"code_verifier": {verifier},
	}

	status, first := postForm(t, app, "/oauth/token", form)
	if status != fiber.StatusOK {
		t.Fatalf("the first exchange failed: %v", first)
	}

	accessToken, _ := first["access_token"].(string)

	if store.grantCount() != 1 {
		t.Fatalf("grant count = %d after the first exchange, want 1", store.grantCount())
	}

	status, second := postForm(t, app, "/oauth/token", form)
	if status == fiber.StatusOK {
		t.Fatal("the replayed code was accepted")
	}

	if got, _ := second["error"].(string); got != "invalid_grant" {
		t.Errorf("error = %q, want invalid_grant", got)
	}

	if n := store.grantCount(); n != 0 {
		t.Errorf("grant count = %d after a replay, want 0: the whole grant must be revoked", n)
	}

	if _, _, err := svc.AuthenticateOAuthToken(context.Background(), accessToken); err == nil {
		t.Error("the token minted from the replayed code still authenticates")
	}
}

// TestConsentDenialTellsTheClient: a user saying no is an answer the client is
// entitled to, and it must not mint anything.
func TestConsentDenialTellsTheClient(t *testing.T) {
	h := oauthTestApp(t)
	app, sessionToken, store := h.app, h.session, h.store
	clientID := registerTestClient(t, app)
	_, challenge := pkcePair()

	_, location := follow(t, app, authorizeURL(clientID, challenge, nil))
	redirect := approve(t, app, sessionToken, requestIDFrom(t, location), false)

	u, err := url.Parse(redirect)
	if err != nil {
		t.Fatalf("parsing %q: %v", redirect, err)
	}

	if got := u.Query().Get("error"); got != "access_denied" {
		t.Errorf("error = %q, want access_denied", got)
	}

	if u.Query().Get("code") != "" {
		t.Error("a denied request still produced a code")
	}

	if store.grantCount() != 0 {
		t.Error("a denied request created a grant")
	}
}

// TestConsentRequiresASession: the consent routes are what bind a code to a
// person, so an unauthenticated caller must not reach them — otherwise anyone
// holding a request id could approve on someone else's behalf.
func TestConsentRequiresASession(t *testing.T) {
	h := oauthTestApp(t)
	app, sessionToken := h.app, h.session
	clientID := registerTestClient(t, app)
	_, challenge := pkcePair()

	_, location := follow(t, app, authorizeURL(clientID, challenge, nil))
	requestID := requestIDFrom(t, location)

	for _, method := range []string{"GET", "POST"} {
		t.Run(method+" without a session", func(t *testing.T) {
			req := httptest.NewRequest(method, "/auth/oauth/consent/"+requestID, strings.NewReader(`{"approved":true}`))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("%s: %v", method, err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != fiber.StatusUnauthorized {
				t.Errorf("status = %d, want 401", resp.StatusCode)
			}
		})
	}

	// And the request survived the unauthenticated attempts, so the legitimate
	// user can still complete it.
	if redirect := approve(t, app, sessionToken, requestID, true); !strings.Contains(redirect, "code=") {
		t.Errorf("the request was consumed by the rejected attempts: %q", redirect)
	}
}

// TestConsentRequestIsSingleUse: one approval, one code. A screen re-submitted
// must not mint a second one.
func TestConsentRequestIsSingleUse(t *testing.T) {
	h := oauthTestApp(t)
	app, sessionToken := h.app, h.session
	clientID := registerTestClient(t, app)
	_, challenge := pkcePair()

	_, location := follow(t, app, authorizeURL(clientID, challenge, nil))
	requestID := requestIDFrom(t, location)

	approve(t, app, sessionToken, requestID, true)

	status, envelope := callMCP(t, app, sessionToken, "POST", "/auth/oauth/consent/"+requestID, `{"approved":true}`)
	if status != fiber.StatusNotFound {
		t.Errorf("re-approving = %d, want 404 (%v)", status, envelope)
	}
}

// TestRegistrationRefusesUnsafeRedirectURIs: the registered URI is the whole
// boundary — everything after registration matches against it literally — so
// this is the last place a plain-http or fragment-bearing URI can be stopped.
func TestRegistrationRefusesUnsafeRedirectURIs(t *testing.T) {
	h := oauthTestApp(t)
	app := h.app

	for _, tc := range []struct{ name, uri string }{
		{name: "plain http to a public host", uri: "http://evil.test/callback"},
		{name: "a fragment", uri: "https://claude.test/cb#frag"},
		{name: "a relative URI", uri: "/callback"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := `{"client_name":"X","redirect_uris":["` + tc.uri + `"],"token_endpoint_auth_method":"none"}`

			req := httptest.NewRequest("POST", "/oauth/register", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("register: %v", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != fiber.StatusBadRequest {
				t.Errorf("status = %d, want 400 for %q", resp.StatusCode, tc.uri)
			}
		})
	}

	t.Run("loopback http is allowed for native clients", func(t *testing.T) {
		body := `{"client_name":"Native","redirect_uris":["http://127.0.0.1:8976/callback"],"token_endpoint_auth_method":"none"}`

		req := httptest.NewRequest("POST", "/oauth/register", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("register: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != fiber.StatusCreated {
			t.Errorf("status = %d, want 201: RFC 8252 allows loopback", resp.StatusCode)
		}
	})
}

// TestRevokingAGrantKillsItsTokens: the disconnect button in the settings
// screen has to take effect on the next call, not when a token expires.
func TestRevokingAGrantKillsItsTokens(t *testing.T) {
	h := oauthTestApp(t)
	app, svc, sessionToken := h.app, h.svc, h.session
	clientID := registerTestClient(t, app)
	verifier, challenge := pkcePair()

	_, location := follow(t, app, authorizeURL(clientID, challenge, nil))
	redirect := approve(t, app, sessionToken, requestIDFrom(t, location), true)
	u, _ := url.Parse(redirect)

	_, tokens := postForm(t, app, "/oauth/token", url.Values{
		"grant_type": {"authorization_code"}, "code": {u.Query().Get("code")},
		"redirect_uri": {testRedirectURI}, "client_id": {clientID},
		"code_verifier": {verifier},
	})

	accessToken, _ := tokens["access_token"].(string)
	refreshToken, _ := tokens["refresh_token"].(string)

	status, envelope := callMCP(t, app, sessionToken, "GET", "/auth/oauth-grants", "")
	if status != fiber.StatusOK {
		t.Fatalf("listing grants = %d (%v)", status, envelope)
	}

	listed, _ := envelope["data"].([]any)
	if len(listed) != 1 {
		t.Fatalf("listed %d grants, want 1", len(listed))
	}

	first, _ := listed[0].(map[string]any)
	grantID, _ := first["id"].(string)

	if name, _ := first["clientName"].(string); name != "Claude" {
		t.Errorf("clientName = %q, want the registered name", name)
	}

	status, envelope = callMCP(t, app, sessionToken, "DELETE", "/auth/oauth-grants/"+grantID, "")
	if status != fiber.StatusOK {
		t.Fatalf("revoking = %d (%v)", status, envelope)
	}

	if _, _, err := svc.AuthenticateOAuthToken(context.Background(), accessToken); err == nil {
		t.Error("the access token still works after the grant was revoked")
	}

	status, body := postForm(t, app, "/oauth/token", url.Values{
		"grant_type": {"refresh_token"}, "refresh_token": {refreshToken}, "client_id": {clientID},
	})
	if status == fiber.StatusOK {
		t.Errorf("the refresh token still works after the grant was revoked: %v", body)
	}
}
