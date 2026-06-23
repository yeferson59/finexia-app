-- 000004_seed_portfolio_data.up.sql
-- Idempotent seed for tables without user dependencies (assets, exchange_rates)
-- This seed intentionally removes all user-dependent inserts (users, roles, accounts,
-- sessions, portfolios, investment_sources, portfolio_entries, transactions, snapshots)
-- because those will be associated to the application's normal user creation flow.
-- Safe to run multiple times.

-- ASSETS (idempotent: insert only if same ticker+exchange combination doesn't exist)
INSERT INTO assets (id, ticker, name, asset_type, exchange, currency, current_price, price_updated_at, created_at, updated_at)
SELECT 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'AAPL', 'Apple Inc.', 'stock', 'NASDAQ', 'USD', 195.00, NOW(), NOW(), NOW()
WHERE NOT EXISTS (
  SELECT 1 FROM assets WHERE ticker = 'AAPL' AND COALESCE(exchange, '') = 'NASDAQ'
);

INSERT INTO assets (id, ticker, name, asset_type, exchange, currency, current_price, price_updated_at, created_at, updated_at)
SELECT 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'MSFT', 'Microsoft Corporation', 'stock', 'NASDAQ', 'USD', 330.00, NOW(), NOW(), NOW()
WHERE NOT EXISTS (
  SELECT 1 FROM assets WHERE ticker = 'MSFT' AND COALESCE(exchange, '') = 'NASDAQ'
);

INSERT INTO assets (id, ticker, name, asset_type, exchange, currency, current_price, price_updated_at, created_at, updated_at)
SELECT 'cccccccc-cccc-cccc-cccc-cccccccccccc', 'SPY', 'SPDR S&P 500 ETF Trust', 'etf', 'NYSEARCA', 'USD', 450.00, NOW(), NOW(), NOW()
WHERE NOT EXISTS (
  SELECT 1 FROM assets WHERE ticker = 'SPY' AND COALESCE(exchange, '') = 'NYSEARCA'
);

INSERT INTO assets (id, ticker, name, asset_type, exchange, currency, current_price, price_updated_at, created_at, updated_at)
SELECT 'dddddddd-dddd-dddd-dddd-dddddddddddd', 'BTC-USD', 'Bitcoin', 'crypto', 'Coinbase', 'USD', 60000.00, NOW(), NOW(), NOW()
WHERE NOT EXISTS (
  SELECT 1 FROM assets WHERE ticker = 'BTC-USD' AND COALESCE(exchange, '') = 'Coinbase'
);

INSERT INTO assets (id, ticker, name, asset_type, exchange, currency, current_price, price_updated_at, created_at, updated_at)
SELECT 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 'ETH-USD', 'Ethereum', 'crypto', 'Coinbase', 'USD', 4000.00, NOW(), NOW(), NOW()
WHERE NOT EXISTS (
  SELECT 1 FROM assets WHERE ticker = 'ETH-USD' AND COALESCE(exchange, '') = 'Coinbase'
);

INSERT INTO assets (id, ticker, name, asset_type, exchange, currency, current_price, price_updated_at, created_at, updated_at)
SELECT 'ffffffff-ffff-ffff-ffff-ffffffffffff', 'BND', 'Vanguard Total Bond Market ETF', 'bond', 'NASDAQ', 'USD', 100.00, NOW(), NOW(), NOW()
WHERE NOT EXISTS (
  SELECT 1 FROM assets WHERE ticker = 'BND' AND COALESCE(exchange, '') = 'NASDAQ'
);

-- EXCHANGE RATES (idempotent by from_currency, to_currency, rate_date)
INSERT INTO exchange_rates (id, from_currency, to_currency, rate, rate_date, fetched_at)
VALUES
  ('81818181-8181-8181-8181-818181818181', 'EUR', 'USD', 1.20, '2021-06-01', NOW()),
  ('82828282-8282-8282-8282-828282828282', 'GBP', 'USD', 1.39, '2021-06-01', NOW()),
  ('83838383-8383-8383-8383-838383838383', 'USD', 'USD', 1.00000000, '2021-06-01', NOW())
ON CONFLICT (from_currency, to_currency, rate_date) DO UPDATE
  SET rate = EXCLUDED.rate,
      fetched_at = NOW();

INSERT INTO risks (id, name, description, created_at, updated_at)
VALUES
  ('25549e04-2eb7-4a05-9f07-4698324588ce', 'Bajo Riesgo', 'Inversiones conservadoras', NOW(), NOW()),
  ('62f30795-c2d5-4cdd-8c67-aef7d588aefa', 'Riesgo Moderado', 'Balance entre riesgo y retorno', NOW(), NOW()),
  ('5bb911c5-a14f-4a50-aa8d-3f032baf1cf5', 'Alto Riesgo', 'Busca máximo crecimiento', NOW(), NOW())
ON CONFLICT (id) DO UPDATE
  SET name = EXCLUDED.name,
      description = EXCLUDED.description,
      updated_at = NOW();

INSERT INTO roles (id, name, description, created_at, updated_at)
VALUES
  ('25549e04-2eb7-4a05-9f07-4698324588ce', 'customer', 'Rol por defecto para usuarios autenticados', NOW(), NOW())
ON CONFLICT (id) DO UPDATE
  SET name = EXCLUDED.name,
      description = EXCLUDED.description,
      updated_at = NOW();

-- End of simplified, user-independent seed
