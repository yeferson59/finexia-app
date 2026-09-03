-- Back to a column that must be named by every writer.
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_class WHERE relkind = 'r' AND relname = 'transactions') THEN
    EXECUTE 'DROP TRIGGER IF EXISTS trg_transactions_default_fees_currency ON transactions';
  END IF;
END$$;

DROP FUNCTION IF EXISTS transactions_default_fees_currency();
