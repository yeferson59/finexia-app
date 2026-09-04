-- OAuth 2.1 authorization server, for the MCP endpoint only.
--
-- /mcp already accepted two credentials: a session's access token and a
-- personal access token from the settings screen. Both assume a human who can
-- paste a secret into a config file. The remote connectors — claude.ai, and
-- every MCP client that follows it — cannot: they only know how to register
-- themselves and walk an authorization flow, so a server that wants their
-- traffic has to be an authorization server, not just a resource server.
--
-- The whole surface is scoped to MCP on purpose. These tokens authenticate
-- nothing but /mcp, they carry one read-only scope, and no browser session is
-- ever minted from them. What this is not is a general-purpose OAuth provider
-- for the REST API: that would put every write endpoint behind a credential a
-- third party holds, which is a decision nobody has made.

-- Clients register themselves (RFC 7591) rather than being provisioned by hand:
-- an MCP client is software the user installs, and there is no operator in the
-- loop to issue it credentials. Registration is therefore open, which is safe
-- only because a registered client can still do nothing at all until a logged-in
-- user approves it on the consent screen.
CREATE TABLE oauth_clients (
  id                         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  -- The public identifier the client sends on every request. Not the primary
  -- key: it is chosen by us but travels in URLs and logs, so the internal id
  -- stays a UUID like every other table here.
  client_id                  TEXT         NOT NULL UNIQUE,
  -- Hex SHA-256, NULL for a public client. A client that runs on the user's
  -- own machine cannot keep a secret, so OAuth 2.1 has it authenticate with
  -- PKCE alone; storing a secret it would have to ship anyway buys nothing.
  client_secret_hash         CHAR(64),
  -- What the consent screen calls it. Client-supplied and therefore untrusted:
  -- it is escaped where it is rendered, and it is the only client-supplied
  -- string a user ever sees.
  client_name                VARCHAR(120) NOT NULL,
  -- The exact URIs a code may be delivered to. Matched literally at both
  -- /authorize and /token — no prefix, no wildcard, no normalisation — because
  -- every open-redirect in an OAuth server has come from loosening this.
  redirect_uris              TEXT[]       NOT NULL,
  grant_types                TEXT[]       NOT NULL,
  response_types             TEXT[]       NOT NULL,
  scope                      TEXT         NOT NULL,
  token_endpoint_auth_method TEXT         NOT NULL,
  client_uri                 TEXT,
  logo_uri                   TEXT,
  created_at                 TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- A pending consent: the authorization request, parked server-side while the
-- user is sent to the frontend to approve it.
--
-- The alternative is to carry the parameters through the consent screen as
-- query string and read them back on approval, which is how this is usually
-- done and is a mistake: every one of them — redirect_uri, scope, the PKCE
-- challenge — would then be attacker-supplied at the moment the code is minted,
-- and the validation done at /authorize would be worth nothing. Here the
-- browser carries an opaque id and nothing else.
CREATE TABLE oauth_authorization_requests (
  id                    UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  client_id             TEXT        NOT NULL REFERENCES oauth_clients(client_id) ON DELETE CASCADE,
  redirect_uri          TEXT        NOT NULL,
  scope                 TEXT        NOT NULL,
  -- The client's CSRF token, echoed back untouched on the redirect.
  state                 TEXT,
  code_challenge        TEXT        NOT NULL,
  code_challenge_method TEXT        NOT NULL,
  -- RFC 8707: which resource the token is being requested for. Recorded so the
  -- audience can be checked at /token rather than assumed.
  resource              TEXT,
  expires_at            TIMESTAMPTZ NOT NULL,
  created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- The authorization code, minted only after a user approved a request above.
--
-- Consumed rows are kept rather than deleted: a code presented twice is the
-- signature of a stolen code, and the only way to notice is to still have the
-- row that says this one was already spent. See the reuse handling in
-- service_oauth.go, which revokes the whole grant rather than just refusing.
CREATE TABLE oauth_authorization_codes (
  id                    UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  code_hash             CHAR(64)    NOT NULL UNIQUE,
  client_id             TEXT        NOT NULL REFERENCES oauth_clients(client_id) ON DELETE CASCADE,
  user_id               UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  -- Copied from the request, not re-read from the token call: the code is bound
  -- to the URI it was delivered to, and /token verifies the two match.
  redirect_uri          TEXT        NOT NULL,
  scope                 TEXT        NOT NULL,
  code_challenge        TEXT        NOT NULL,
  code_challenge_method TEXT        NOT NULL,
  resource              TEXT,
  consumed_at           TIMESTAMPTZ,
  expires_at            TIMESTAMPTZ NOT NULL,
  created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- One row per live authorization: the user's standing grant to one client, and
-- the access/refresh pair that currently represents it.
--
-- Access and refresh live in the same row because they are the same grant seen
-- at two timescales, and because revoking is then one delete rather than a
-- cascade nobody remembers to write. Refreshing rotates both hashes in place,
-- so a grant keeps its id and its created_at across a year of hourly
-- refreshes — which is what makes the row readable on a "connected apps"
-- screen as the thing the user actually consented to.
CREATE TABLE oauth_grants (
  id                 UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  client_id          TEXT        NOT NULL REFERENCES oauth_clients(client_id) ON DELETE CASCADE,
  user_id            UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  scope              TEXT        NOT NULL,
  -- Only hashes, for the reason mcp_tokens gives: a dump of this table
  -- authenticates nobody.
  access_token_hash  CHAR(64)    NOT NULL UNIQUE,
  access_expires_at  TIMESTAMPTZ NOT NULL,
  -- NULL when the client was not issued a refresh token.
  refresh_token_hash CHAR(64)    UNIQUE,
  refresh_expires_at TIMESTAMPTZ,
  resource           TEXT,
  -- Written at most once every few minutes by the guard, like mcp_tokens.
  last_used_at       TIMESTAMPTZ,
  created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Listing is always "this user's connected apps", and revoking a client is
-- "every grant of it".
CREATE INDEX idx_oauth_grants_user_id ON oauth_grants(user_id);
CREATE INDEX idx_oauth_grants_client_id ON oauth_grants(client_id);

-- The two sweeps the cleanup job runs.
CREATE INDEX idx_oauth_codes_expires_at ON oauth_authorization_codes(expires_at);
CREATE INDEX idx_oauth_requests_expires_at ON oauth_authorization_requests(expires_at);

-- A user re-approving a client they already approved should update the grant
-- they have, not accumulate a second one that no screen distinguishes from the
-- first. Scope is part of the key because a grant for different scopes is a
-- different consent.
CREATE UNIQUE INDEX uq_oauth_grants_user_client_scope ON oauth_grants(user_id, client_id, scope);
