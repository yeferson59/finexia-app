-- What a contribution or a withdrawal of a fund followed by balance was, in
-- money (docs/PLAN_FONDOS_INVERSION.md, D4 and D5).
--
-- A fund followed by 'balance' (000057) has synthetic units: the first
-- contribution buys at a unit value of 1, a balance on a day fixes the unit
-- value as balance / units held, and every contribution or withdrawal trades at
-- the unit value of the last balance before its day. The quantity and price of
-- its transactions are therefore derived, and change whenever an earlier
-- movement or balance does. What does not change is what the owner stated: the
-- pesos that went in or came out. That is the fact, and it is kept here, so
-- replaying the fund always starts from what was said and never from a quantity
-- an earlier replay rounded.
--
--   amount         what went in, or what came out before the fees, in the
--                  fund's currency.
--   withdraws_all  a withdrawal of everything the position held: its quantity
--                  is all the units, whatever they are after a replay, and its
--                  price is amount / units.
--
-- One row per transaction of such a fund, deleted with it.
CREATE TABLE IF NOT EXISTS fund_movements (
  txn_id        UUID PRIMARY KEY REFERENCES transactions(id) ON DELETE CASCADE,
  amount        NUMERIC(20, 8) NOT NULL CHECK (amount > 0),
  withdraws_all BOOLEAN NOT NULL DEFAULT FALSE
);
