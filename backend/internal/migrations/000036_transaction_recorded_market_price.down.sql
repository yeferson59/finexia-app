-- Back to netting every transaction at its cost, loaded history included.
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_class WHERE relkind = 'r' AND relname = 'transactions') THEN
    EXECUTE 'DROP TRIGGER IF EXISTS trg_transactions_record_market_price ON transactions';
  END IF;
END$$;

DROP FUNCTION IF EXISTS transactions_record_market_price();
DROP FUNCTION IF EXISTS transaction_holding_flow(transaction_type, NUMERIC, NUMERIC);

ALTER TABLE transactions DROP COLUMN IF EXISTS recorded_market_price;
