-- The voluntary contributions of /apoyar, paid through Bold's payment button.
--
-- A row is written when the backend signs an order, before the payer has seen
-- Bold's checkout, so every order Bold ever hears about is one this table
-- already knows. Its status then only moves forward on what Bold says: the
-- webhook, or the payment-voucher query when someone comes back from the
-- checkout. The URL Bold redirects to carries a status too, but anyone can type
-- that one, so it never writes here.
--
-- No user_id on purpose. The page is public and contributing changes nothing in
-- an account (the page says so), so there is nothing to tie a payment to — and
-- no personal data of the payer is stored either: Bold keeps that.
CREATE TYPE support_contribution_status AS ENUM (
  -- The order was signed; nobody has paid yet, and most never will: closing
  -- the checkout leaves the row here.
  'created',
  -- Bold is still processing (PSE, or a card in review).
  'pending',
  -- The last attempt failed. Not final: the checkout lets the payer try again
  -- on the same order.
  'rejected',
  'approved',
  -- An approved payment Bold later voided.
  'voided'
);

CREATE TABLE support_contributions (
  id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  -- The order id sent to Bold, which comes back as data.metadata.reference in
  -- the webhook and is the key of the voucher query. UNIQUE is the lookup index.
  order_id        VARCHAR(60) NOT NULL UNIQUE,
  -- Whole pesos, as Bold takes them: COP has no cents in practice.
  amount          BIGINT      NOT NULL CHECK (amount > 0),
  currency        CHAR(3)     NOT NULL DEFAULT 'COP',
  status          support_contribution_status NOT NULL DEFAULT 'created',
  -- What Bold reports once there is a transaction. Kept apart from amount:
  -- the order is what was asked, this is what was charged.
  total_charged   BIGINT,
  payment_id      VARCHAR(64),
  payment_method  VARCHAR(32),
  approved_at     TIMESTAMPTZ,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Listing and reconciling go by status, newest first.
CREATE INDEX idx_support_contributions_status ON support_contributions(status, created_at DESC);

-- Every webhook Bold delivered, by its own event id.
--
-- Bold retries a notification until it gets a 200 and may deliver one more
-- than once, so the event id is the idempotency key: the insert and the status
-- change it causes commit together, and a second delivery of the same id finds
-- the row and changes nothing.
--
-- Only what reconciliation needs is kept, not the payload: the payload carries
-- the payer's email and card details, and none of that belongs here.
CREATE TABLE support_payment_events (
  event_id     VARCHAR(64) PRIMARY KEY,
  type         VARCHAR(32) NOT NULL,
  payment_id   VARCHAR(64) NOT NULL,
  -- NULL or unknown when the payment was not one of ours (a payment link made
  -- from Bold's panel, say): it is still recorded, and still answered with 200.
  order_id     VARCHAR(60),
  total        BIGINT,
  received_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_support_payment_events_order ON support_payment_events(order_id);
