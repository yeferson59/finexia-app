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

-- Remove risks inserted by the seed
DELETE FROM risks WHERE id IN (
  '25549e04-2eb7-4a05-9f07-4698324588ce',
  '62f30795-c2d5-4cdd-8c67-aef7d588aefa',
  '5bb911c5-a14f-4a50-aa8d-3f032baf1cf5'
);

-- Remove roles inserted by the seed
DELETE FROM roles WHERE id IN (
  '25549e04-2eb7-4a05-9f07-4698324588ce'
);


-- End of cleanup
