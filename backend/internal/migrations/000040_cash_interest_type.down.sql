-- Postgres cannot drop a value from an enum without rebuilding the type, and
-- transaction_type is referenced by two tables, a view and several functions.
-- The value stays. 000041's down migration removes every rule that gave it a
-- meaning, so a cash_interest row left behind reads as a row that moves no
-- quantity and no flow.
SELECT 1;
