-- Approximate, human-readable location ("Bogotá, Colombia") resolved from the
-- session's IP, so the active-devices list can show where each session came
-- from. Nullable: private IPs and failed lookups simply leave it empty.
ALTER TABLE sessions ADD COLUMN location VARCHAR(120);
