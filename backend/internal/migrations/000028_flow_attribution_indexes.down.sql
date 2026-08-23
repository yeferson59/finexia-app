CREATE INDEX IF NOT EXISTS idx_transactions_entry_date ON transactions(entry_id, transaction_date);

DROP INDEX IF EXISTS idx_transactions_entry_created;

DROP INDEX IF EXISTS idx_snapshots_portfolio_created;
