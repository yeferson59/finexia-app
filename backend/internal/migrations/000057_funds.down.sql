-- The unit values the owners wrote were copied to user_asset_prices; without
-- the marks behind them they are prices nobody can explain, so they go and the
-- fund positions fall back to their cost. The snapshots keep the values the
-- marks restated them to: they are history, and nothing here can say what they
-- were before.
DELETE FROM user_asset_prices uap
 USING assets a
 WHERE a.id = uap.asset_id
   AND a.asset_type = 'fund'
   AND uap.source = 'user';

DROP TABLE IF EXISTS fund_snapshot_values;
DROP TABLE IF EXISTS fund_marks;
DROP TABLE IF EXISTS user_funds;
DROP TYPE IF EXISTS fund_tracking;
