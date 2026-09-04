package auth

// The OAuth 2.1 use cases: registration, the two halves of the authorization
// flow, the token endpoint, and the authentication the /mcp guard performs on
// an OAuth access token.
//
// The flow is split across three callers on purpose, and it is worth naming
// them because no single function here tells the whole story. The *client*
// calls /register and /authorize and /token. The *browser* — a logged-in user,
// on the frontend — calls the consent pair in the middle. Nothing the client
// sends reaches the code-minting step except through a row this server wrote
// and the browser only named by an opaque id.

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"

	"uuid"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

// consentPath is where the browser is sent to approve a request. It lives on
// the frontend because that is where the session is: the API is on another
// origin and holds no cookie, so it cannot know who is looking at the screen.
const consentPath = "/oauth/consent"

// RegisterOAuthClient implements RFC 7591. Registration is open — an MCP client
// is software the user installed, and there is no operator in the loop to
// provision it — which is safe only because a registered client can do nothing
// until a user approves it on the consent screen.
//
// The client's own metadata is not trusted anywhere it could matter: the name
// is escaped where it is rendered, and the redirect URIs are validated here and
// then matched literally forever after.
func (s *service) RegisterOAuthClient(ctx context.Context, reg OAuthClientRegistration) (OAuthClientCredentials, error) {
	name, err := normalizeClientName(reg.ClientName)
	if err != nil {
		return OAuthClientCredentials{}, err
	}

	if len(reg.RedirectURIs) == 0 {
		return OAuthClientCredentials{}, oauthInvalidRedirectURI("redirect_uris is required")
	}

	for _, uri := range reg.RedirectURIs {
		if err := validateRedirectURI(uri); err != nil {
			return OAuthClientCredentials{}, err
		}
	}

	scope, err := normalizeScope(reg.Scope)
	if err != nil {
		return OAuthClientCredentials{}, err
	}

	// The only response type this server has, and the only two grants. A client
	// asking for something else is refused rather than silently registered with
	// less than it asked for, so it finds out now instead of at /authorize.
	if err := requireSubset(reg.ResponseTypes, []string{"code"}, "response_types"); err != nil {
		return OAuthClientCredentials{}, err
	}

	if err := requireSubset(reg.GrantTypes, []string{"authorization_code", "refresh_token"}, "grant_types"); err != nil {
		return OAuthClientCredentials{}, err
	}

	authMethod := reg.TokenEndpointAuthMethod
	if authMethod == "" {
		// RFC 7591's default is client_secret_basic, but OAuth 2.1's advice for
		// a client that cannot hold a secret is "none". Defaulting the other way
		// would hand a secret to every client that did not think to ask not to
		// have one, and a secret shipped inside a desktop app is not a secret.
		authMethod = "none"
	}

	if authMethod != "none" && authMethod != "client_secret_post" && authMethod != "client_secret_basic" {
		return OAuthClientCredentials{}, oauthInvalidClientMetadata("unsupported token_endpoint_auth_method: " + authMethod)
	}

	clientID, err := generateClientID()
	if err != nil {
		return OAuthClientCredentials{}, err
	}

	var rawSecret, secretHash string
	if authMethod != "none" {
		if rawSecret, secretHash, err = generateOAuthToken("fnx_csec_"); err != nil {
			return OAuthClientCredentials{}, err
		}
	}

	client := oauthClient{
		ClientID:          clientID,
		SecretHash:        secretHash,
		Name:              name,
		RedirectURIs:      reg.RedirectURIs,
		GrantTypes:        []string{"authorization_code", "refresh_token"},
		ResponseTypes:     []string{"code"},
		Scope:             scope,
		TokenEndpointAuth: authMethod,
		ClientURI:         reg.ClientURI,
		LogoURI:           reg.LogoURI,
	}

	if err := s.stores.OAuth.CreateOAuthClient(ctx, client); err != nil {
		return OAuthClientCredentials{}, err
	}

	s.log.Info(ctx, "oauth client registered",
		logger.Str("client_id", clientID),
		logger.Str("client_name", name),
	)

	return OAuthClientCredentials{
		ClientID:         clientID,
		ClientSecret:     rawSecret,
		ClientIDIssuedAt: time.Now().UTC().Unix(),
		// 0 means "does not expire" per RFC 7591 §3.2.1. Rotating a secret
		// would need a client that knows how to re-register, and none does.
		ClientSecretExpiresAt: 0,
		ClientName:            name,
		RedirectURIs:          client.RedirectURIs,
		GrantTypes:            client.GrantTypes,
		ResponseTypes:         client.ResponseTypes,
		Scope:                 scope,
		TokenEndpointAuth:     authMethod,
	}, nil
}

// requireSubset refuses metadata naming anything this server does not do.
func requireSubset(requested, supported []string, field string) error {
	for _, r := range requested {
		if !contains(supported, r) {
			return oauthInvalidClientMetadata("unsupported " + field + " value: " + r)
		}
	}

	return nil
}

func contains(haystack []string, needle string) bool {
	for _, h := range haystack {
		if h == needle {
			return true
		}
	}

	return false
}

// StartAuthorization validates an /authorize call and returns the URL the
// browser should be sent to next — the consent screen when the request is
// usable, the client's own redirect_uri carrying an error when it is not.
//
// The split between "returns a URL carrying an error" and "returns an error" is
// the security-relevant part of this endpoint, and it is decided by the first
// two checks. Until the client is known *and* the redirect_uri is one it
// registered, there is nowhere safe to send a response: redirecting to an
// unverified URI would let anyone learn which client ids exist by pointing one
// at a server they control. After those two pass, every remaining failure is
// the client's to handle and travels to the client, per RFC 6749 §4.1.2.1.
func (s *service) StartAuthorization(ctx context.Context, req OAuthAuthorizationRequest) (string, error) {
	client, err := s.stores.OAuth.GetOAuthClient(ctx, req.ClientID)
	if err != nil {
		if errors.Is(err, ErrOAuthClientNotFound) {
			return "", oauthInvalidRequest("unknown client_id")
		}

		return "", err
	}

	if !client.allowsRedirectURI(req.RedirectURI) {
		return "", oauthInvalidRedirectURI("redirect_uri is not registered for this client")
	}

	// From here on, errors go to the client.
	redirectErr := func(oerr *oauthError) (string, error) {
		target, berr := buildRedirect(req.RedirectURI, map[string]string{
			"error":             oerr.Code,
			"error_description": oerr.Description,
			"state":             req.State,
		})
		if berr != nil {
			return "", berr
		}

		return target, nil
	}

	if req.ResponseType != "code" {
		return redirectErr(oauthUnsupportedResponseType("only the authorization code flow is supported"))
	}

	// PKCE is mandatory, not negotiated. OAuth 2.1 requires it of every client,
	// public and confidential alike, and a server that accepts an authorization
	// request without it is a server whose codes are worth stealing.
	if req.CodeChallenge == "" {
		return redirectErr(oauthInvalidRequest("code_challenge is required"))
	}

	method := req.CodeChallengeMethod
	if method == "" {
		// RFC 7636's default is "plain", which OAuth 2.1 removed. Treating an
		// absent method as S256 rather than defaulting to a method this server
		// refuses is what keeps a client that omits the field working.
		method = oauthCodeChallengeS256
	}

	if method != oauthCodeChallengeS256 {
		return redirectErr(oauthInvalidRequest("code_challenge_method must be S256"))
	}

	scope, err := normalizeScope(req.Scope)
	if err != nil {
		oerr, _ := asOAuthError(err)

		return redirectErr(oerr)
	}

	resource, err := normalizeResource(req.Resource, s.mcpResourceURI())
	if err != nil {
		oerr, _ := asOAuthError(err)

		return redirectErr(oerr)
	}

	id, err := s.stores.OAuth.CreateAuthorizationRequest(ctx, pendingAuthorization{
		ClientID:            client.ClientID,
		RedirectURI:         req.RedirectURI,
		Scope:               scope,
		State:               req.State,
		CodeChallenge:       req.CodeChallenge,
		CodeChallengeMethod: method,
		Resource:            resource,
		ExpiresAt:           time.Now().UTC().Add(oauthRequestTTL),
	})
	if err != nil {
		return redirectErr(oauthServerError("could not start the authorization"))
	}

	return s.consentURL(id), nil
}

// consentURL is where the browser goes to approve. The id is the only thing
// that travels: every parameter the decision depends on stays in the row.
func (s *service) consentURL(id uuid.UUID) string {
	return strings.TrimSuffix(s.cfg.FrontendURL, "/") + consentPath + "?request=" + url.QueryEscape(id.String())
}

// mcpResourceURI is the canonical identifier of the thing being protected, in
// the RFC 8707 sense: the audience a token is minted for, and the "resource"
// field of the protected-resource metadata.
func (s *service) mcpResourceURI() string {
	return strings.TrimSuffix(s.cfg.PublicURL, "/") + "/mcp"
}

// issuer is this authorization server's identity. It has to be the origin the
// metadata is served from, byte for byte, or a client that checks the issuer
// against where it fetched the document — which every correct one does —
// rejects everything this server says.
func (s *service) issuer() string {
	return strings.TrimSuffix(s.cfg.PublicURL, "/")
}

// GetConsent describes a parked request for the screen that will approve it.
func (s *service) GetConsent(ctx context.Context, requestID uuid.UUID) (OAuthConsent, error) {
	pending, err := s.stores.OAuth.GetAuthorizationRequest(ctx, requestID)
	if err != nil {
		return OAuthConsent{}, oauthConsentError(err)
	}

	client, err := s.stores.OAuth.GetOAuthClient(ctx, pending.ClientID)
	if err != nil {
		return OAuthConsent{}, oauthConsentError(err)
	}

	return OAuthConsent{
		RequestID:   pending.ID,
		ClientName:  client.Name,
		ClientURI:   client.ClientURI,
		LogoURI:     client.LogoURI,
		RedirectURI: pending.RedirectURI,
		Scopes:      scopeList(pending.Scope),
		ExpiresAt:   pending.ExpiresAt,
	}, nil
}

// ApproveConsent mints the authorization code and returns where to send the
// browser. It is the only path that creates a code, and it takes the user id
// from the session rather than from anything the request carried.
func (s *service) ApproveConsent(ctx context.Context, requestID, userID uuid.UUID) (string, error) {
	pending, err := s.stores.OAuth.GetAuthorizationRequest(ctx, requestID)
	if err != nil {
		return "", oauthConsentError(err)
	}

	rawCode, codeHash, err := generateOAuthToken("")
	if err != nil {
		return "", err
	}

	now := time.Now().UTC()
	if err := s.stores.OAuth.CreateAuthorizationCode(ctx, codeHash, authorizationCode{
		ClientID:            pending.ClientID,
		UserID:              userID,
		RedirectURI:         pending.RedirectURI,
		Scope:               pending.Scope,
		CodeChallenge:       pending.CodeChallenge,
		CodeChallengeMethod: pending.CodeChallengeMethod,
		Resource:            pending.Resource,
		ExpiresAt:           now.Add(oauthCodeTTL),
	}); err != nil {
		return "", err
	}

	// Deleted rather than left to expire: the request has been answered, and a
	// second approval of the same screen must not mint a second code.
	if err := s.stores.OAuth.DeleteAuthorizationRequest(ctx, requestID); err != nil {
		s.log.Warn(ctx, "oauth: could not delete an approved authorization request",
			logger.Str("request_id", requestID.String()),
			logger.Err(err),
		)
	}

	s.log.Info(ctx, "oauth authorization approved",
		logger.Str("client_id", pending.ClientID),
		logger.Str("user_id", userID.String()),
	)

	return buildRedirect(pending.RedirectURI, map[string]string{
		"code":  rawCode,
		"state": pending.State,
	})
}

// DenyConsent answers the client with access_denied. The user said no, which
// the client is entitled to be told: leaving it hanging is what makes a
// connector spin until it times out.
func (s *service) DenyConsent(ctx context.Context, requestID uuid.UUID) (string, error) {
	pending, err := s.stores.OAuth.GetAuthorizationRequest(ctx, requestID)
	if err != nil {
		return "", oauthConsentError(err)
	}

	if err := s.stores.OAuth.DeleteAuthorizationRequest(ctx, requestID); err != nil {
		s.log.Warn(ctx, "oauth: could not delete a denied authorization request",
			logger.Str("request_id", requestID.String()),
			logger.Err(err),
		)
	}

	denied := oauthAccessDenied("the user declined the request")

	return buildRedirect(pending.RedirectURI, map[string]string{
		"error":             denied.Code,
		"error_description": denied.Description,
		"state":             pending.State,
	})
}

// oauthConsentError maps the store's sentinels for the consent screen, which
// is an ordinary browser page and gets the envelope everything else does.
func oauthConsentError(err error) error {
	switch {
	case errors.Is(err, ErrOAuthRequestNotFound), errors.Is(err, ErrOAuthClientNotFound):
		return httpx.AsNotFound(ErrOAuthRequestNotFound)
	default:
		return err
	}
}

// OAuthTokenRequest is one /token call, already parsed out of the form body and
// the Authorization header.
type OAuthTokenRequest struct {
	GrantType    string
	Code         string
	RedirectURI  string
	CodeVerifier string
	RefreshToken string
	ClientID     string
	ClientSecret string
	Resource     string
}

// ExchangeOAuthToken is the token endpoint. Both grants end in the same place —
// a rotated pair on one grant row — so the difference between them is only how
// the caller proves they are entitled to it.
func (s *service) ExchangeOAuthToken(ctx context.Context, req OAuthTokenRequest) (OAuthTokens, error) {
	switch req.GrantType {
	case "authorization_code":
		return s.exchangeAuthorizationCode(ctx, req)
	case "refresh_token":
		return s.exchangeRefreshToken(ctx, req)
	default:
		return OAuthTokens{}, oauthUnsupportedGrantType("unsupported grant_type: " + req.GrantType)
	}
}

// authenticateClient resolves the caller of /token and checks whatever proof
// its registration says it has.
//
// A public client proves nothing here — PKCE is its proof, and it is checked
// against the code. That is not a gap: a client with no secret is one whose
// secret would have shipped inside a binary anyone can read, and pretending
// otherwise buys a false sense of a boundary.
func (s *service) authenticateClient(ctx context.Context, clientID, secret string) (oauthClient, error) {
	if clientID == "" {
		return oauthClient{}, oauthInvalidClient("client_id is required")
	}

	client, err := s.stores.OAuth.GetOAuthClient(ctx, clientID)
	if err != nil {
		if errors.Is(err, ErrOAuthClientNotFound) {
			return oauthClient{}, oauthInvalidClient("unknown client")
		}

		return oauthClient{}, err
	}

	if client.isPublic() {
		return client, nil
	}

	if secret == "" || !verifyClientSecret(secret, client.SecretHash) {
		return oauthClient{}, oauthInvalidClient("client authentication failed")
	}

	return client, nil
}

func (s *service) exchangeAuthorizationCode(ctx context.Context, req OAuthTokenRequest) (OAuthTokens, error) {
	client, err := s.authenticateClient(ctx, req.ClientID, req.ClientSecret)
	if err != nil {
		return OAuthTokens{}, err
	}

	if req.Code == "" {
		return OAuthTokens{}, oauthInvalidRequest("code is required")
	}

	code, claimed, err := s.stores.OAuth.ConsumeAuthorizationCode(ctx, hashOAuthToken(req.Code))
	if err != nil {
		if errors.Is(err, ErrOAuthCodeNotFound) {
			return OAuthTokens{}, oauthInvalidGrant("the authorization code is invalid")
		}

		return OAuthTokens{}, err
	}

	// A code presented twice was intercepted: whoever replayed it either raced
	// the legitimate client or stole what it already spent, and in both cases
	// the tokens minted from the first exchange are in doubt. RFC 6749 §10.5
	// asks for exactly this — revoke everything issued from that authorization,
	// not just refuse the second call.
	if !claimed {
		s.revokeAfterCodeReuse(ctx, code)

		return OAuthTokens{}, oauthInvalidGrant("the authorization code has already been used")
	}

	switch {
	case time.Now().UTC().After(code.ExpiresAt):
		return OAuthTokens{}, oauthInvalidGrant("the authorization code has expired")
	case code.ClientID != client.ClientID:
		// The code belongs to another client. Answered as an invalid grant
		// rather than as a client error: which client a code was minted for is
		// not something the caller is entitled to learn.
		return OAuthTokens{}, oauthInvalidGrant("the authorization code is invalid")
	case req.RedirectURI != "" && req.RedirectURI != code.RedirectURI:
		return OAuthTokens{}, oauthInvalidGrant("redirect_uri does not match the authorization request")
	case req.RedirectURI == "":
		return OAuthTokens{}, oauthInvalidRequest("redirect_uri is required")
	}

	if err := verifyPKCE(req.CodeVerifier, code.CodeChallenge, code.CodeChallengeMethod); err != nil {
		return OAuthTokens{}, err
	}

	if _, err := normalizeResource(req.Resource, s.mcpResourceURI()); err != nil {
		return OAuthTokens{}, err
	}

	return s.issueTokens(ctx, code.UserID, code.ClientID, code.Scope, code.Resource, nil)
}

// revokeAfterCodeReuse destroys every grant the replayed code's client holds
// for its user. Failures are logged rather than returned: the caller is being
// refused either way, and a store that broke must not turn a detected replay
// into a successful one.
func (s *service) revokeAfterCodeReuse(ctx context.Context, code authorizationCode) {
	s.log.Error(ctx, "oauth: authorization code replayed; revoking the client's grants",
		logger.Str("client_id", code.ClientID),
		logger.Str("user_id", code.UserID.String()),
	)

	revoked, err := s.stores.OAuth.DeleteOAuthGrantsForClient(ctx, code.UserID, code.ClientID)
	if err != nil {
		s.log.Error(ctx, "oauth: could not revoke grants after a code replay", logger.Err(err))

		return
	}

	s.log.Warn(ctx, "oauth: grants revoked after a code replay", logger.Int64("revoked", revoked))
}

func (s *service) exchangeRefreshToken(ctx context.Context, req OAuthTokenRequest) (OAuthTokens, error) {
	client, err := s.authenticateClient(ctx, req.ClientID, req.ClientSecret)
	if err != nil {
		return OAuthTokens{}, err
	}

	if req.RefreshToken == "" {
		return OAuthTokens{}, oauthInvalidRequest("refresh_token is required")
	}

	grant, err := s.stores.OAuth.GetGrantByRefreshToken(ctx, hashOAuthToken(req.RefreshToken))
	if err != nil {
		if errors.Is(err, ErrOAuthGrantNotFound) {
			// Also what a revoked connection gets: the row is gone, so there is
			// nothing to distinguish it from a token that never existed — which
			// is the answer the client needs either way, "authorize again".
			return OAuthTokens{}, oauthInvalidGrant("the refresh token is invalid or was revoked")
		}

		return OAuthTokens{}, err
	}

	switch {
	case grant.ClientID != client.ClientID:
		return OAuthTokens{}, oauthInvalidGrant("the refresh token is invalid or was revoked")
	case grant.RefreshExpiresAt != nil && time.Now().UTC().After(*grant.RefreshExpiresAt):
		return OAuthTokens{}, oauthInvalidGrant("the refresh token has expired")
	}

	return s.issueTokens(ctx, grant.UserID, grant.ClientID, grant.Scope, grant.Resource, &grant.ID)
}

// issueTokens mints a pair and writes it to the grant.
//
// grantID distinguishes the two callers: a refresh rotates the row it was given,
// an authorization creates or replaces the row for (user, client, scope). Both
// replace the access token as well as the refresh token, so a rotation never
// leaves the previous access token alive.
func (s *service) issueTokens(
	ctx context.Context,
	userID uuid.UUID,
	clientID, scope, resource string,
	grantID *uuid.UUID,
) (OAuthTokens, error) {
	accessRaw, accessHash, err := generateOAuthToken(oauthAccessPrefix)
	if err != nil {
		return OAuthTokens{}, err
	}

	refreshRaw, refreshHash, err := generateOAuthToken(oauthRefreshPrefix)
	if err != nil {
		return OAuthTokens{}, err
	}

	now := time.Now().UTC()
	accessExpiresAt := now.Add(oauthAccessTTL)
	refreshExpiresAt := now.Add(oauthRefreshTTL)

	if grantID != nil {
		err = s.stores.OAuth.RotateGrantTokens(ctx, *grantID, accessHash, accessExpiresAt, refreshHash, &refreshExpiresAt)
	} else {
		_, err = s.stores.OAuth.UpsertOAuthGrant(ctx, userID, clientID, scope, resource, accessHash, accessExpiresAt, refreshHash, &refreshExpiresAt)
	}

	if err != nil {
		if errors.Is(err, ErrOAuthGrantNotFound) {
			return OAuthTokens{}, oauthInvalidGrant("the refresh token is invalid or was revoked")
		}

		return OAuthTokens{}, err
	}

	return OAuthTokens{
		AccessToken:  accessRaw,
		TokenType:    "Bearer",
		ExpiresIn:    int(oauthAccessTTL.Seconds()),
		RefreshToken: refreshRaw,
		Scope:        scope,
	}, nil
}

// AuthenticateOAuthToken resolves a presented access token to the identity the
// /mcp guard puts in the request locals, in the same shape
// AuthenticateMCPToken returns — so the guard stays unaware of which of the
// three credentials it just accepted.
//
// Uncached, for the reason the personal access tokens are: revoking a
// connection has to take effect on the next call, not when a cache entry
// happens to expire, and one indexed lookup per tool call is the cheaper side
// of that trade.
func (s *service) AuthenticateOAuthToken(ctx context.Context, raw string) (uuid.UUID, string, error) {
	if !looksLikeOAuthToken(raw) {
		return uuid.Nil(), "", httpx.AsBadRequest(ErrInvalidOAuthToken)
	}

	grant, err := s.stores.OAuth.GetGrantByAccessToken(ctx, hashOAuthToken(raw))
	if err != nil {
		// An unknown hash is the ordinary case — a revoked connection, an
		// expired token the client has not refreshed yet, a probe — so it is
		// not logged as a failure. A store that broke is.
		if !errors.Is(err, ErrOAuthGrantNotFound) {
			s.log.Error(ctx, "oauth token lookup failed", logger.Err(err))
		}

		return uuid.Nil(), "", httpx.AsBadRequest(ErrInvalidOAuthToken)
	}

	now := time.Now().UTC()
	if now.After(grant.ExpiresAt) {
		return uuid.Nil(), "", httpx.AsBadRequest(ErrInvalidOAuthToken)
	}

	s.touchOAuthGrant(ctx, grant, now)

	return grant.UserID, grant.Role, nil
}

// touchOAuthGrant keeps last_used_at fresh enough to answer "is this connection
// alive?" without writing on every call. A failed write is swallowed: the
// column is for the settings screen, and losing one update must never turn a
// working tool call into a 401.
func (s *service) touchOAuthGrant(ctx context.Context, grant oauthGrantIdentity, now time.Time) {
	if grant.LastUsedAt != nil && now.Sub(*grant.LastUsedAt) < oauthGrantTouchInterval {
		return
	}

	if err := s.stores.OAuth.TouchOAuthGrant(ctx, grant.ID); err != nil {
		s.log.Warn(ctx, "oauth grant touch failed",
			logger.Str("grant_id", grant.ID.String()),
			logger.Err(err),
		)
	}
}

// ListOAuthGrants returns the user's connected applications.
func (s *service) ListOAuthGrants(ctx context.Context, userID uuid.UUID) ([]OAuthGrant, error) {
	return s.stores.OAuth.ListOAuthGrants(ctx, userID)
}

// RevokeOAuthGrant disconnects one application. Both its tokens die with the
// row, so the next tool call it makes is a 401 and the next refresh is a dead
// end — there is no window in which a revoked connector still reads anything.
func (s *service) RevokeOAuthGrant(ctx context.Context, userID, grantID uuid.UUID) error {
	if err := s.stores.OAuth.DeleteOAuthGrant(ctx, userID, grantID); err != nil {
		if errors.Is(err, ErrOAuthGrantNotFound) {
			return httpx.AsNotFound(err)
		}

		return err
	}

	s.log.Info(ctx, "oauth grant revoked",
		logger.Str("user_id", userID.String()),
		logger.Str("grant_id", grantID.String()),
	)

	return nil
}
