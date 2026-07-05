-- ip_address was VARCHAR(14), too short even for a full IPv4 address
-- ("255.255.255.255" is 15 chars) and far too short for IPv6 (up to 45 chars).
ALTER TABLE sessions ALTER COLUMN ip_address TYPE VARCHAR(45);
