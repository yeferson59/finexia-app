-- Index the growth series' flow lookup by the column it actually searches on.
--
-- 000027 attributed every transaction to the first snapshot on or after its
-- transaction_date, and indexed transactions for that. It was the wrong date.
--
-- A snapshot is a record of what the portfolio held when the daily job ran, and
-- past snapshots are never recomputed. So a position registered today with a
-- trade date of two months ago moves the series today — that is the first
-- snapshot that ever saw it — while its trade date points at a snapshot that
-- was already written without it. Netting the deposit out on the older date
-- subtracted money from a day whose value had not moved and left the day the
-- value did move with nothing to net, so a single backdated purchase was
-- counted as a loss on one day and as a windfall on another. On an account
-- loaded with an existing portfolio, which is every new account, that turned a
-- +10% year into a +270% one.
--
-- The queries now match transactions.created_at against portfolio_snapshots
-- .created_at: a snapshot reflects a transaction exactly when it was computed
-- after the transaction was recorded, whatever day the trade itself happened.
CREATE INDEX IF NOT EXISTS idx_snapshots_portfolio_created ON portfolio_snapshots(portfolio_id, created_at);

CREATE INDEX IF NOT EXISTS idx_transactions_entry_created ON transactions(entry_id, created_at);

DROP INDEX IF EXISTS idx_transactions_entry_date;
