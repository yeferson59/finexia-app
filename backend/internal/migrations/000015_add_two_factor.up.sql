-- Two-factor authentication (TOTP), opt-in per user: rows only exist for
-- users who started a setup, and "enabled" stays FALSE until the user proves
-- possession of the authenticator by confirming a valid code.
CREATE TABLE IF NOT EXISTS user_two_factor(
  user_id UUID PRIMARY KEY NOT NULL,
  secret VARCHAR(64) NOT NULL,
  enabled BOOLEAN NOT NULL DEFAULT FALSE,
  confirmed_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT fk_user_two_factor_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Single-use recovery codes, stored as SHA-256 hashes so a database leak
-- never exposes a usable code.
CREATE TABLE IF NOT EXISTS user_two_factor_recovery_codes(
  id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL,
  code_hash VARCHAR(64) NOT NULL,
  used_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT fk_two_factor_recovery_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_two_factor_recovery_user_id ON user_two_factor_recovery_codes(user_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_two_factor_recovery_code_hash ON user_two_factor_recovery_codes(code_hash);
