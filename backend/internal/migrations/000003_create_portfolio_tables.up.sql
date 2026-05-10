CREATE TYPE asset_type AS ENUM (
  'stock',
  'etf',
  'crypto',
  'bond',
  'cash',
  'other'
);

CREATE TYPE source_type AS ENUM (
  'broker',
  'bank',
  'crypto_exchange',
  'platform',
  'excel',
  'manual'
);

CREATE TYPE transaction_type AS ENUM (
  'buy',
  'sell',
  'dividend',
  'split',
  'transfer_in',
  'transfer_out',
  'fee',
  'interest'
);

CREATE TYPE portfolio_category AS ENUM (
  'stocks',
  'etf',
  'crypto',
  'bonds',
  'cash',
  'real_estate',
  'commodities',
  'other'
);

CREATE TABLE investment_sources (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL,
  name VARCHAR(255) NOT NULL,
  source_type source_type NOT NULL,
  description VARCHAR(500),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT fk_sources_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_sources_user ON investment_sources(user_id);

CREATE TABLE portfolios (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL,
  name VARCHAR(255) NOT NULL,
  description VARCHAR(500),
  base_currency CHAR(3) NOT NULL DEFAULT 'USD',
  is_default BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT fk_portfolios_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_portfolios_user ON portfolios(user_id);
CREATE UNIQUE INDEX idx_portfolios_default ON portfolios(user_id) WHERE is_default = TRUE;


CREATE TABLE assets (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  ticker VARCHAR(20) NOT NULL,
  name VARCHAR(255) NOT NULL,
  asset_type asset_type NOT NULL,
  exchange VARCHAR(100),
  currency CHAR(3) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_assets_ticker_exchange ON assets(ticker, COALESCE(exchange, ''));
CREATE INDEX idx_assets_type ON assets(asset_type);


CREATE TABLE portfolio_entries (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  portfolio_id UUID NOT NULL,
  asset_id UUID NOT NULL,
  source_id UUID,
  quantity NUMERIC(20, 8) NOT NULL DEFAULT 0,
  avg_cost_price NUMERIC(20, 8) NOT NULL,
  cost_currency CHAR(3) NOT NULL,
  category portfolio_category NOT NULL DEFAULT 'other',
  entry_date DATE NOT NULL,
  notes VARCHAR(500),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT fk_portfolio_entries_portfolio FOREIGN KEY (portfolio_id) REFERENCES portfolios(id) ON DELETE CASCADE,
  CONSTRAINT fk_portfolio_entries_asset FOREIGN KEY (asset_id) REFERENCES assets(id),
  CONSTRAINT fk_portfolio_entries_source FOREIGN KEY (source_id) REFERENCES investment_sources(id) ON DELETE SET NULL
);

CREATE UNIQUE INDEX idx_entries_portfolio_asset_source ON portfolio_entries(portfolio_id, asset_id, COALESCE(source_id::TEXT, ''));
CREATE INDEX idx_entries_portfolio ON portfolio_entries(portfolio_id);
CREATE INDEX idx_entries_asset     ON portfolio_entries(asset_id);

CREATE TABLE transactions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  entry_id UUID NOT NULL,
  type transaction_type NOT NULL,
  quantity NUMERIC(20, 8) NOT NULL DEFAULT 0,
  price NUMERIC(20, 8) NOT NULL,
  currency CHAR(3) NOT NULL,
  fees NUMERIC(20, 8) NOT NULL DEFAULT 0,
  transaction_date DATE NOT NULL,
  notes VARCHAR(500),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT fk_transactions_entry FOREIGN KEY (entry_id) REFERENCES portfolio_entries(id) ON DELETE CASCADE
);

CREATE INDEX idx_transactions_entry ON transactions(entry_id);
CREATE INDEX idx_transactions_date  ON transactions(transaction_date DESC);

CREATE TABLE exchange_rates (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  from_currency CHAR(3) NOT NULL,
  to_currency CHAR(3) NOT NULL,
  rate NUMERIC(20, 8) NOT NULL,
  rate_date DATE NOT NULL,
  fetched_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_rates_pair_date ON exchange_rates(from_currency, to_currency, rate_date);
CREATE INDEX idx_rates_date ON exchange_rates(rate_date DESC);

CREATE TABLE portfolio_snapshots (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  portfolio_id UUID NOT NULL,
  snapshot_date DATE NOT NULL,
  total_value NUMERIC(20, 8) NOT NULL,
  currency CHAR(3) NOT NULL,
  allocation JSONB NOT NULL DEFAULT '{}',
  total_gain_loss NUMERIC(20, 8) NOT NULL DEFAULT 0,
  total_gain_loss_pct NUMERIC(10, 6) NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT fk_snapshots_portfolio FOREIGN KEY (portfolio_id) REFERENCES portfolios(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_snapshots_portfolio_date ON portfolio_snapshots(portfolio_id, snapshot_date);
CREATE INDEX idx_snapshots_date ON portfolio_snapshots(snapshot_date DESC);

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;


CREATE TRIGGER trg_portfolios_updated_at BEFORE UPDATE ON portfolios FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_entries_updated_at BEFORE UPDATE ON portfolio_entries FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE OR REPLACE FUNCTION recalculate_avg_cost()
RETURNS TRIGGER AS $$
DECLARE
  v_total_qty   NUMERIC;
  v_total_cost  NUMERIC;
BEGIN
  SELECT
    SUM(CASE WHEN type IN ('buy', 'transfer_in') THEN quantity ELSE -quantity END),
    SUM(CASE WHEN type IN ('buy', 'transfer_in') THEN quantity * price ELSE 0 END)
  INTO v_total_qty, v_total_cost
  FROM transactions
  WHERE entry_id = NEW.entry_id
    AND type IN ('buy', 'sell', 'transfer_in', 'transfer_out');

  IF v_total_qty > 0 THEN
    UPDATE portfolio_entries
    SET
      quantity       = v_total_qty,
      avg_cost_price = v_total_cost / v_total_qty,
      updated_at     = NOW()
    WHERE id = NEW.entry_id;
  END IF;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_recalculate_avg_cost AFTER INSERT OR UPDATE ON transactions FOR EACH ROW EXECUTE FUNCTION recalculate_avg_cost();

CREATE OR REPLACE VIEW portfolio_summary AS
SELECT
  p.id                AS portfolio_id,
  p.user_id,
  p.name              AS portfolio_name,
  p.base_currency,
  COUNT(pe.id)        AS total_positions,
  SUM(
    pe.quantity * pe.avg_cost_price *
    COALESCE(
      (SELECT er.rate FROM exchange_rates er
       WHERE er.from_currency = pe.cost_currency
         AND er.to_currency   = p.base_currency
       ORDER BY er.rate_date DESC LIMIT 1),
      1
    )
  )                   AS total_cost_base,
  p.created_at
FROM portfolios p
LEFT JOIN portfolio_entries pe ON pe.portfolio_id = p.id
GROUP BY p.id;
