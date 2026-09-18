-- Postgres cannot drop a value from an enum without rebuilding the type (see
-- 000040's down migration). The value stays; 000050's down migration removes
-- every row of it and every rule that gave it a meaning.
SELECT 1;
