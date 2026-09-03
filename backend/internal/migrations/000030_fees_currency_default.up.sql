-- Make fees_currency fill itself in, the way fx_rate does.
--
-- 000029 added the column NOT NULL with no default, which left every writer that
-- does not name it failing on a not-null violation. fx_rate got away with
-- DEFAULT 1 because its "omitted" value is a constant; fees_currency's is not —
-- an omitted fee currency means "the same one the trade was quoted in", and a
-- column default cannot read another column of its own row.
--
-- That asymmetry is a landmine rather than a nuisance. The application always
-- names the column, because TransactionInput.Validate resolves it before any
-- insert; what breaks is everything else — an operator's backfill, a fixture, a
-- future writer that predates none of this and simply lists the columns it
-- knows. A NOT NULL column with no default asks every one of them to remember a
-- rule that is already written down twice.
--
-- A BEFORE trigger is what a column default would be if defaults could see the
-- row. It runs before the not-null constraint is checked, so an insert that
-- omits the column ends up with the right value instead of an error, and the
-- rule ends up stated in the schema rather than only in the callers.
--
-- On UPDATE too: setting fees_currency to NULL explicitly means the same thing
-- it means on insert, and answering it differently depending on the statement
-- would be a distinction with no meaning behind it.
CREATE OR REPLACE FUNCTION transactions_default_fees_currency()
RETURNS TRIGGER AS $$
BEGIN
  IF NEW.fees_currency IS NULL THEN
    NEW.fees_currency := NEW.currency;
  END IF;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_class WHERE relkind = 'r' AND relname = 'transactions') THEN
    EXECUTE 'DROP TRIGGER IF EXISTS trg_transactions_default_fees_currency ON transactions';
    EXECUTE 'CREATE TRIGGER trg_transactions_default_fees_currency
             BEFORE INSERT OR UPDATE ON transactions
             FOR EACH ROW EXECUTE FUNCTION transactions_default_fees_currency()';
  END IF;
END$$;
