package auth

// The OAuth endpoints. Three of them answer a machine and one answers a
// browser, and that split decides the response shape every time: /register,
// /token and the two discovery documents speak the specs' own JSON, while the
// consent pair the frontend calls travels in the httpx envelope like every
// other route in this API.
//
// None of the spec-shaped responses use httpx.Success or httpx.Error. That is
// not an oversight — RFC 6749 §5.1/§5.2, RFC 7591 §3.2 and RFC 8414 §3.2 each
// define the body byte for byte, and a client parses it by that shape or fails.
// /mcp already carries the same exemption for JSON-RPC; see docs/API.md §1.1.

import (
	"encoding/base64"
	"html"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v3"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

// oauthPublicCORS lets a browser-based client read the endpoints that carry no
// cookie and no ambient authority.
//
// It is a wildcard because these four are genuinely public: the two discovery
// documents are constants, /register is open by design, and /token authorises
// the caller by the code and PKCE verifier in the body rather than by anything
// the browser attaches on its own. Credentials are explicitly not allowed, so
// no session cookie can ride along — which is what keeps the wildcard from
// being a CSRF surface.
func oauthPublicCORS(c fiber.Ctx) error {
	c.Set(fiber.HeaderAccessControlAllowOrigin, "*")
	c.Set(fiber.HeaderAccessControlAllowHeaders, "Content-Type, Authorization, MCP-Protocol-Version")
	c.Set(fiber.HeaderAccessControlAllowMethods, "GET, POST, OPTIONS")

	if c.Method() == fiber.MethodOptions {
		return c.SendStatus(fiber.StatusNoContent)
	}

	return c.Next()
}

// writeOAuthError answers in the OAuth error shape, with the no-store the
// specs require of anything that touched token material.
func writeOAuthError(c fiber.Ctx, err error) error {
	oerr, ok := asOAuthError(err)
	if !ok {
		return httpx.InternalServerError(c, "authorization failed", "auth:oauth")
	}

	noStore(c)

	if oerr.Code == "invalid_client" {
		// RFC 6749 §5.2: a 401 from the token endpoint carries a challenge.
		c.Set(fiber.HeaderWWWAuthenticate, `Basic realm="finexia"`)
	}

	return c.Status(oerr.Status).JSON(fiber.Map{
		"error":             oerr.Code,
		"error_description": oerr.Description,
	})
}

// noStore keeps tokens out of every cache between here and the client.
func noStore(c fiber.Ctx) {
	c.Set(fiber.HeaderCacheControl, "no-store")
	c.Set(fiber.HeaderPragma, "no-cache")
}

// protectedResourceMetadata answers RFC 9728 for /mcp. Served at both the
// resource-suffixed path the spec derives from "/mcp" and the bare one, because
// clients in the wild ask for either.
func (h *handler) protectedResourceMetadata(c fiber.Ctx) error {
	return c.JSON(h.service.ProtectedResourceMetadata())
}

// authorizationServerMetadata answers RFC 8414.
func (h *handler) authorizationServerMetadata(c fiber.Ctx) error {
	return c.JSON(h.service.ServerMetadata())
}

// registerOAuthClient implements RFC 7591 dynamic client registration.
func (h *handler) registerOAuthClient(c fiber.Ctx) error {
	reg, err := httpx.Bind[OAuthClientRegistration](c)
	if err != nil {
		return writeOAuthError(c, oauthInvalidClientMetadata("the registration body is not valid JSON"))
	}

	credentials, err := h.service.RegisterOAuthClient(c, reg)
	if err != nil {
		if _, ok := asOAuthError(err); ok {
			return writeOAuthError(c, err)
		}

		return httpx.InternalServerError(c, "registration failed", "auth:oauth:register")
	}

	noStore(c)

	return c.Status(fiber.StatusCreated).JSON(credentials)
}

// authorize is the front door of the flow, and the only endpoint here a person
// actually looks at. It answers with a redirect in every case it can, and with
// a rendered page only when redirecting would itself be the vulnerability —
// see StartAuthorization for which is which.
func (h *handler) authorize(c fiber.Ctx) error {
	req := OAuthAuthorizationRequest{
		ClientID:            c.Query("client_id"),
		RedirectURI:         c.Query("redirect_uri"),
		ResponseType:        c.Query("response_type"),
		Scope:               c.Query("scope"),
		State:               c.Query("state"),
		CodeChallenge:       c.Query("code_challenge"),
		CodeChallengeMethod: c.Query("code_challenge_method"),
		Resource:            c.Query("resource"),
	}

	target, err := h.service.StartAuthorization(c, req)
	if err != nil {
		return renderAuthorizeError(c, err)
	}

	return c.Redirect().Status(fiber.StatusFound).To(target)
}

// renderAuthorizeError is the dead end: a request whose client or redirect_uri
// could not be verified, which therefore has nowhere to be sent. A person is
// looking at this, so it is a page rather than a JSON body — and every value
// interpolated into it is escaped, because the only strings here came from the
// query string of a link somebody clicked.
func renderAuthorizeError(c fiber.Ctx, err error) error {
	oerr, ok := asOAuthError(err)
	if !ok {
		oerr = oauthServerError("the authorization request could not be processed")
	}

	c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)

	return c.Status(oerr.Status).SendString(`<!doctype html>
<html lang="en"><head><meta charset="utf-8"><title>Authorization error</title>
<meta name="viewport" content="width=device-width,initial-scale=1">
<style>
body{font-family:system-ui,-apple-system,"Segoe UI",sans-serif;background:#0b0b0d;color:#e8e8ea;
display:flex;align-items:center;justify-content:center;min-height:100vh;margin:0;padding:24px}
main{max-width:34rem}h1{font-size:1.25rem;margin:0 0 .75rem}
p{line-height:1.6;color:#a8a8b3;margin:0 0 .5rem}code{color:#e8b563}
</style></head>
<body><main>
<h1>This application could not be authorized</h1>
<p>` + html.EscapeString(oerr.Description) + `</p>
<p>Error code: <code>` + html.EscapeString(oerr.Code) + `</code></p>
<p>Nothing was shared. You can close this window and try connecting again from the application.</p>
</main></body></html>`)
}

// token is the token endpoint. The body is form-encoded, not JSON: RFC 6749
// §4.1.3 says so, and every client sends it that way.
func (h *handler) token(c fiber.Ctx) error {
	clientID, clientSecret := c.FormValue("client_id"), c.FormValue("client_secret")

	// client_secret_basic: the credentials arrive in the Authorization header
	// instead of the body, and take precedence — a client that authenticated in
	// the header has said which identity it is claiming.
	if id, secret, ok := basicAuth(c.Get(fiber.HeaderAuthorization)); ok {
		clientID, clientSecret = id, secret
	}

	tokens, err := h.service.ExchangeOAuthToken(c, OAuthTokenRequest{
		GrantType:    c.FormValue("grant_type"),
		Code:         c.FormValue("code"),
		RedirectURI:  c.FormValue("redirect_uri"),
		CodeVerifier: c.FormValue("code_verifier"),
		RefreshToken: c.FormValue("refresh_token"),
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Resource:     c.FormValue("resource"),
	})
	if err != nil {
		if _, ok := asOAuthError(err); ok {
			return writeOAuthError(c, err)
		}

		return httpx.InternalServerError(c, "the token request failed", "auth:oauth:token")
	}

	noStore(c)

	return c.JSON(tokens)
}

// basicAuth decodes a client_secret_basic header. Per RFC 6749 §2.3.1 both
// halves are form-urlencoded before being joined, which matters for a secret
// containing a colon — this server never mints one, but a client may have been
// registered elsewhere and re-pointed here.
func basicAuth(header string) (string, string, bool) {
	const scheme = "Basic "

	if len(header) <= len(scheme) || !strings.EqualFold(header[:len(scheme)], scheme) {
		return "", "", false
	}

	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(header[len(scheme):]))
	if err != nil {
		return "", "", false
	}

	id, secret, found := strings.Cut(string(decoded), ":")
	if !found {
		return "", "", false
	}

	unescapedID, err := url.QueryUnescape(id)
	if err != nil {
		unescapedID = id
	}

	unescapedSecret, err := url.QueryUnescape(secret)
	if err != nil {
		unescapedSecret = secret
	}

	return unescapedID, unescapedSecret, true
}

// getOAuthConsent describes a parked request for the screen that will approve
// it. Session-guarded: the consent screen is a page in the dashboard, and the
// frontend calls this with the user's own access token.
func (h *handler) getOAuthConsent(c fiber.Ctx) error {
	requestID, err := httpx.ParamUUID(c, "id")
	if err != nil {
		return httpx.BadRequest(c, "invalid request id", "auth:oauth:consent")
	}

	consent, err := h.service.GetConsent(c, requestID)
	if err != nil {
		return httpx.FromDomain(c, err, "this authorization request is no longer valid", "auth:oauth:consent")
	}

	return httpx.OK(c, "authorization request retrieved", "", consent)
}

// decideOAuthConsent records the user's answer and returns where to send the
// browser next. The redirect is returned rather than performed: the caller is
// the frontend's server-side load, not the browser, so it is the one that has
// to issue the redirect the user's browser will follow.
func (h *handler) decideOAuthConsent(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "invalid user id", "auth:oauth:consent")
	}

	requestID, err := httpx.ParamUUID(c, "id")
	if err != nil {
		return httpx.BadRequest(c, "invalid request id", "auth:oauth:consent")
	}

	req, err := httpx.Bind[OAuthConsentDecisionDTO](c)
	if err != nil {
		return httpx.BadRequest(c, "invalid request body", "auth:oauth:consent")
	}

	var target string
	if req.Approved {
		target, err = h.service.ApproveConsent(c, requestID, userID)
	} else {
		target, err = h.service.DenyConsent(c, requestID)
	}

	if err != nil {
		return httpx.FromDomain(c, err, "this authorization request is no longer valid", "auth:oauth:consent")
	}

	return httpx.OK(c, "authorization request answered", "", fiber.Map{"redirectTo": target})
}

// listOAuthGrants and revokeOAuthGrant are the connected-applications surface:
// what a user has approved, and the button that takes it back.
func (h *handler) listOAuthGrants(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "invalid user id", "auth:oauth:grants:list")
	}

	grants, err := h.service.ListOAuthGrants(c, userID)
	if err != nil {
		return httpx.FromDomain(c, err, "failed to list connected applications", "auth:oauth:grants:list")
	}

	return httpx.OK(c, "connected applications retrieved", "", grants)
}

func (h *handler) revokeOAuthGrant(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "invalid user id", "auth:oauth:grants:revoke")
	}

	grantID, err := httpx.ParamUUID(c, "id")
	if err != nil {
		return httpx.BadRequest(c, "invalid grant id", "auth:oauth:grants:revoke")
	}

	if err := h.service.RevokeOAuthGrant(c, userID, grantID); err != nil {
		return httpx.FromDomain(c, err, "failed to disconnect the application", "auth:oauth:grants:revoke")
	}

	return httpx.OK(c, "application disconnected", "its access stopped working immediately", nil)
}
