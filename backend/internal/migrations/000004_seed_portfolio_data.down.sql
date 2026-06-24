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
