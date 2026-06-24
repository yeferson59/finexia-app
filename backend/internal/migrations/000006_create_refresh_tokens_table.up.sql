CREATE TABLE refresh_tokens (
  id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash  CHAR(64)    NOT NULL UNIQUE,
  family_id   UUID        NOT NULL,
  session_id  UUID        REFERENCES sessions(id) ON DELETE CASCADE,
  ip_address  VARCHAR(45),
  user_agent  VARCHAR(255),
  expires_at  TIMESTAMPTZ NOT NULL,
  used_at     TIMESTAMPTZ,
  revoked_at  TIMESTAMPTZ,
  created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);
CREATE INDEX idx_refresh_tokens_family_id  ON refresh_tokens(family_id);
CREATE INDEX idx_refresh_tokens_user_id    ON refresh_tokens(user_id);
