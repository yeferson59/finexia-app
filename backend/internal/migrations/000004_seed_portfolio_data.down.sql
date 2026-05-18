-- 000004_seed_portfolio_data.down.sql
-- Removes the synthetic seed data inserted by 000004_seed_portfolio_data.up.sql
-- This down script only removes the user-independent inserts (assets and exchange_rates)
-- because user-dependent data is managed by the application/user-creation flow.

-- Remove exchange rates inserted by the seed
DELETE FROM exchange_rates WHERE id IN (
  '81818181-8181-8181-8181-818181818181',
  '82828282-8282-8282-8282-828282828282',
  '83838383-8383-8383-8383-838383838383'
);

-- Remove assets inserted by the seed
DELETE FROM assets WHERE id IN (
  'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
  'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
  'cccccccc-cccc-cccc-cccc-cccccccccccc',
  'dddddddd-dddd-dddd-dddd-dddddddddddd',
  'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee',
  'ffffffff-ffff-ffff-ffff-ffffffffffff'
);

-- End of cleanup
