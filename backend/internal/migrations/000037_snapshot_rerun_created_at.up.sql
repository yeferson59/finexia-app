-- Give a snapshot re-written later in its day the time of that later write.
--
-- The growth series hands each transaction to the first snapshot whose
-- created_at is not older than the transaction, reading created_at as "when this
-- row's totals were read". Until now a second run of the snapshot job on the
-- same date — a retry, or the catch-up a restart fires — refreshed the totals
-- and left created_at at the first write. On 2026-08-14 the job ran at 00:55 and
-- again at 22:00; the positions recorded between 01:41 and 03:03 were in that
-- day's value while their flows went to the 15th, and the series drew +10% on
-- the 14th and −7% on the 15th out of two days the market barely moved.
--
-- UpsertPortfolioSnapshot now stamps the read time on every write. The rows
-- already stored are repaired from the one trace a later run leaves: a pass
-- writes every portfolio at once, so when a row for a date was inserted later
-- than another row for the same date, that later pass refreshed the older row
-- as well, and the older row takes the later time. Rows one pass wrote together
-- sit milliseconds apart, hence the minute of slack. A date with no row to prove
-- a later pass is left as it is.
UPDATE portfolio_snapshots ps
   SET created_at = latest.written_at
  FROM (
    SELECT snapshot_date, MAX(created_at) AS written_at
    FROM portfolio_snapshots
    GROUP BY snapshot_date
  ) latest
 WHERE ps.snapshot_date = latest.snapshot_date
   AND ps.created_at < latest.written_at - INTERVAL '1 minute';
