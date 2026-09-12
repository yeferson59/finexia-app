-- Back to reading only the transactions as they are: a deletion or a quantity
-- edit again rewrites the flows of days whose snapshots it cannot touch.
DROP VIEW IF EXISTS growth_flow_events;

DROP TRIGGER IF EXISTS trg_entries_retire_transactions ON portfolio_entries;
DROP TRIGGER IF EXISTS trg_transactions_retire_on_resize ON transactions;
DROP TRIGGER IF EXISTS trg_transactions_retire_on_delete ON transactions;
DROP TRIGGER IF EXISTS trg_transactions_default_flow_recorded_at ON transactions;

DROP FUNCTION IF EXISTS entries_retire_transactions();
DROP FUNCTION IF EXISTS transactions_retire_on_resize();
DROP FUNCTION IF EXISTS transactions_retire_on_delete();
DROP FUNCTION IF EXISTS retire_transaction_version(transactions, UUID, CHAR);
DROP FUNCTION IF EXISTS transaction_version_seen(UUID, TIMESTAMPTZ);
DROP FUNCTION IF EXISTS entry_unit_market_value(UUID);
DROP FUNCTION IF EXISTS transactions_default_flow_recorded_at();

DROP TABLE IF EXISTS retired_transactions;

ALTER TABLE transactions DROP COLUMN IF EXISTS flow_recorded_at;
