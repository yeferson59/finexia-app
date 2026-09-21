-- Postgres cannot drop a value from an enum without rebuilding the type (see
-- 000040's down migration). The value stays; 000054's down migration removes
-- every rule that gave it a meaning.
SELECT 1;
