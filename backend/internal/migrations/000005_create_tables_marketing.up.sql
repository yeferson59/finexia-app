DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'waitlist_status') THEN
        CREATE TYPE waitlist_status AS ENUM ('pending', 'invited', 'registered');
    END IF;
END;
$$;

CREATE TABLE IF NOT EXISTS waitlist (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    status waitlist_status DEFAULT 'pending',
    invited_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS waitlist_status_idx ON waitlist (status);
CREATE INDEX IF NOT EXISTS waitlist_email_idx ON waitlist (email);
