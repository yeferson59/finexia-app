DROP VIEW IF EXISTS portfolio_summary;

DROP TABLE IF EXISTS portfolio_snapshots;
DROP TABLE IF EXISTS exchange_rates;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS portfolio_entries;
DROP TABLE IF EXISTS assets;
DROP TABLE IF EXISTS portfolios;
DROP TABLE IF EXISTS risks;
DROP TABLE IF EXISTS investment_sources;

DROP FUNCTION IF EXISTS recalculate_avg_cost();
DROP FUNCTION IF EXISTS set_updated_at();

DROP TYPE IF EXISTS portfolio_entry_category;
DROP TYPE IF EXISTS transaction_type;
DROP TYPE IF EXISTS source_type;
DROP TYPE IF EXISTS asset_type;
DROP TYPE IF EXISTS portfolio_type;
