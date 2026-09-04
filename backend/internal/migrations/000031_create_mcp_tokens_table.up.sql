-- Personal access tokens for the MCP endpoint.
--
-- An MCP client — Claude Desktop, Claude Code, any other — is configured once
-- and then runs unattended, so it cannot use the access token the browser gets:
-- that one lives for minutes and is refreshed through a cookie no external
-- client holds. These tokens are the credential built for that shape of caller:
-- long-lived, created and revoked by the user from the settings screen, and
-- accepted on /mcp only.
--
-- Only the SHA-256 of the token is stored. The secret itself is shown once, at
-- creation and at rotation, and is unrecoverable afterwards — a dump of this
-- table gives an attacker nothing to authenticate with, which is why there is
-- no encrypted-secret column here the way market_credentials has one (that key
-- has to be replayed to a provider; this one only ever has to be compared).
CREATE TABLE mcp_tokens (
  id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id      UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  -- What the user called it ("Claude Desktop", "portátil"), so a token can be
  -- revoked by recognising it rather than by matching a fragment.
  name         VARCHAR(60) NOT NULL,
  -- Hex SHA-256 of the raw token, the same width and encoding refresh_tokens
  -- uses. UNIQUE gives the lookup index authentication needs, so no separate
  -- index is declared for it.
  token_hash   CHAR(64)    NOT NULL UNIQUE,
  -- Trailing fragment of the secret, the only plaintext kept, so the list can
  -- show "····a3f9" next to each name.
  last4        VARCHAR(4)  NOT NULL,
  -- NULL means the token does not expire; the user chooses at creation.
  expires_at   TIMESTAMPTZ,
  -- Written at most once every few minutes by the guard, so the settings
  -- screen can show which tokens are actually in use — and which are not.
  last_used_at TIMESTAMPTZ,
  -- Set when the secret is replaced in place. The row keeps its name and its
  -- creation date, which is what makes a rotation readable as one.
  rotated_at   TIMESTAMPTZ,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Listing is always "this user's tokens".
CREATE INDEX idx_mcp_tokens_user_id ON mcp_tokens(user_id);

-- One name per user, case-insensitively: the name is how the user tells two
-- tokens apart when revoking one, so duplicates would defeat its only purpose.
CREATE UNIQUE INDEX uq_mcp_tokens_user_name ON mcp_tokens(user_id, lower(name));
