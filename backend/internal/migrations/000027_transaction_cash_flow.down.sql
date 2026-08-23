DROP INDEX IF EXISTS idx_transactions_entry_date;

DROP FUNCTION IF EXISTS transaction_cash_flow(transaction_type, NUMERIC, NUMERIC, NUMERIC);
