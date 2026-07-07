-- Sessions (and the refresh tokens that back them) are deleted on logout and
-- swept on expiry, so relying on the sessions table to recognize a returning
-- IP meant every re-login after a logout was flagged as an unknown device,
-- even from the exact same address. This table persists independently of
-- session lifecycle so "have we seen this IP for this user" survives logout.
CREATE TABLE known_login_ips (
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    ip_address    VARCHAR(45) NOT NULL,
    first_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, ip_address)
);
