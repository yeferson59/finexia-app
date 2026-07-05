CREATE TABLE IF NOT EXISTS invitations (
  id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
  email VARCHAR(255) NOT NULL,
  name VARCHAR(255) NOT NULL,
  role VARCHAR(30) NOT NULL DEFAULT 'customer',
  token_hash VARCHAR(64) NOT NULL UNIQUE,
  invited_by UUID,
  expires_at TIMESTAMPTZ NOT NULL,
  accepted_at TIMESTAMPTZ,
  revoked_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT fk_invitations_invited_by FOREIGN KEY (invited_by) REFERENCES users(id) ON DELETE SET NULL
);

-- At most one live (neither accepted nor revoked) invitation per email, so a
-- resend replaces the previous token instead of piling up redeemable links.
CREATE UNIQUE INDEX IF NOT EXISTS idx_invitations_email_active
  ON invitations (email)
  WHERE accepted_at IS NULL AND revoked_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_invitations_token_hash ON invitations (token_hash);
CREATE INDEX IF NOT EXISTS idx_invitations_email ON invitations (email);
