package portfolio

import (
	"context"
	"errors"
	"fmt"
	"time"

	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/market"
	"github.com/yeferson59/finexia-app/internal/platform/database"
)

// GetCashBalancesByUserID lists every cash position the user has, one row per
// portfolio, platform and currency.
//
// It values them with the rule the holdings and the summary use — the user's
// own price, then the catalog's, then the entry's cost — so a balance shows the
// same amount here as in the portfolio it belongs to. For the balances the app
// opens that rule lands on the catalog's price of one, and the balance is its
// quantity.
//
// Unlike the holdings it keeps the positions at zero. A share sold in full is no
// longer something the user has; an emptied account is still an account, and it
// is where the next deposit goes.
func (r *PostgresRepository) GetCashBalancesByUserID(ctx context.Context, userID uuid.UUID, displayCurrency money.Currency) ([]CashBalance, error) {
	// Rounded before becoming text for the reason the holdings give: past
	// nineteen decimals the engine that parses the text back reads zero.
	rows, err := r.db.Query(ctx, `
		SELECT
			pe.id,
			pe.portfolio_id,
			p.name,
			pe.source_id,
			COALESCE(s.name, ''),
			a.id,
			a.ticker,
			a.name,
			ROUND(pe.quantity::numeric * v.price, 8)::text,
			v.currency,
			ROUND(pe.quantity::numeric * v.price * COALESCE(fx.rate, 1), 8)::text,
			target.code,
			fx.rate IS NOT NULL,
			(SELECT COUNT(*) FROM transactions t WHERE t.entry_id = pe.id),
			(SELECT MAX(t.transaction_date) FROM transactions t WHERE t.entry_id = pe.id),
			pe.pocket_id,
			COALESCE(pk.name, ''),
			COALESCE(pk.kind::text, '')
		FROM portfolio_entries pe
		JOIN portfolios p ON p.id = pe.portfolio_id
		JOIN users u      ON u.id = p.user_id
		JOIN assets a     ON a.id = pe.asset_id
		LEFT JOIN investment_sources s  ON s.id = pe.source_id
		LEFT JOIN cash_pockets pk       ON pk.id = pe.pocket_id
		LEFT JOIN user_asset_prices uap ON uap.asset_id = a.id AND uap.user_id = p.user_id
		CROSS JOIN LATERAL (
			SELECT COALESCE(NULLIF($2::text, ''), u.preferred_currency, 'USD')::char(3) AS code
		) target
		CROSS JOIN LATERAL (
			SELECT
				COALESCE(uap.price::numeric, a.current_price::numeric, pe.price::numeric) AS price,
				CASE
					WHEN COALESCE(uap.price, a.current_price) IS NOT NULL
						THEN COALESCE(a.currency, pe.cost_currency)
					ELSE pe.cost_currency
				END AS currency
		) v
		CROSS JOIN LATERAL (
			SELECT fx_rate(p.user_id, v.currency, target.code) AS rate
		) fx
		WHERE p.user_id = $1
		  AND a.asset_type = 'cash'
		ORDER BY ROUND(pe.quantity::numeric * v.price * COALESCE(fx.rate, 1), 8) DESC, p.name, s.name, a.ticker, pk.name
	`, userID, currencyParam(displayCurrency))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	balances := make([]CashBalance, 0)
	for rows.Next() {
		var b CashBalance

		if err := rows.Scan(
			&b.EntryID,
			&b.PortfolioID,
			&b.PortfolioName,
			&b.SourceID,
			&b.SourceName,
			&b.AssetID,
			&b.Ticker,
			&b.Name,
			&b.Balance,
			&b.Currency,
			&b.Value,
			&b.DisplayCurrency,
			&b.FXConverted,
			&b.Movements,
			&b.LastMovementDate,
			&b.PocketID,
			&b.PocketName,
			&b.PocketKind,
		); err != nil {
			return nil, err
		}

		balances = append(balances, b)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return balances, r.addCashInterest(ctx, userID, displayCurrency, balances)
}

// cashMovementEditable is the SQL test behind CashMovement.Editable, shared by
// the reads and by the write that has to enforce it: a kind the cash screens
// know, at one unit per unit, with no conversion anywhere on the row.
const cashMovementEditable = `(
	t.type IN ('buy', 'transfer_in', 'sell', 'transfer_out', 'cash_interest')
	AND t.price = 1
	AND t.fx_rate = 1
	AND t.currency = pe.cost_currency
	AND t.fees_currency = t.currency
	AND pe.cost_currency = a.currency
)`

const cashMovementColumns = `
	t.id, t.entry_id, t.type,
	ROUND(t.quantity * t.price * t.fx_rate, 8)::text,
	pe.cost_currency,
	t.fees::text, t.fees_currency,
	t.transaction_date, COALESCE(t.notes, ''),
	` + cashMovementEditable + `,
	EXISTS (SELECT 1 FROM cash_interest_accruals ac WHERE ac.transaction_id = t.id),
	COALESCE((
		SELECT oa.ticker
		FROM transactions ot
		JOIN portfolio_entries ope ON ope.id = ot.entry_id
		JOIN assets oa             ON oa.id = ope.asset_id
		WHERE ot.id = t.credited_from
	), ''),
	pe.portfolio_id, p.name, pe.source_id, COALESCE(s.name, ''), a.ticker, t.created_at`

const cashMovementFrom = `
	FROM transactions t
	JOIN portfolio_entries pe ON pe.id = t.entry_id
	JOIN portfolios p         ON p.id = pe.portfolio_id
	JOIN assets a             ON a.id = pe.asset_id
	LEFT JOIN investment_sources s ON s.id = pe.source_id`

func scanCashMovement(row pgx.Row) (CashMovement, error) {
	var m CashMovement

	if err := row.Scan(
		&m.ID,
		&m.EntryID,
		&m.Type,
		&m.Amount,
		&m.Currency,
		&m.Fees,
		&m.FeesCurrency,
		&m.Date,
		&m.Notes,
		&m.Editable,
		&m.Automatic,
		&m.OriginTicker,
		&m.PortfolioID,
		&m.PortfolioName,
		&m.SourceID,
		&m.SourceName,
		&m.Ticker,
		&m.CreatedAt,
	); err != nil {
		return CashMovement{}, err
	}

	m.Kind = cashKindOf(m.Type)

	return m, nil
}

func (r *PostgresRepository) CountCashMovements(ctx context.Context, userID uuid.UUID) (int, error) {
	var total int

	err := r.db.QueryRow(ctx, `SELECT COUNT(*) `+cashMovementFrom+`
		WHERE p.user_id = $1 AND a.asset_type = 'cash'
	`, userID).Scan(&total)

	return total, err
}

// GetCashMovementsPaginated lists the transactions on every cash position the
// user has, most recent first — the same order as the account's ledger.
func (r *PostgresRepository) GetCashMovementsPaginated(ctx context.Context, userID uuid.UUID, limit, offset int) ([]CashMovement, error) {
	rows, err := r.db.Query(ctx, `SELECT `+cashMovementColumns+cashMovementFrom+`
		WHERE p.user_id = $1 AND a.asset_type = 'cash'
		ORDER BY t.transaction_date DESC, t.created_at DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	movements := make([]CashMovement, 0)
	for rows.Next() {
		m, err := scanCashMovement(rows)
		if err != nil {
			return nil, err
		}

		movements = append(movements, m)
	}

	return movements, rows.Err()
}

func getCashMovement(ctx context.Context, tx pgx.Tx, userID, txnID uuid.UUID) (CashMovement, error) {
	m, err := scanCashMovement(tx.QueryRow(ctx, `SELECT `+cashMovementColumns+cashMovementFrom+`
		WHERE t.id = $1 AND p.user_id = $2 AND a.asset_type = 'cash'
	`, txnID, userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return CashMovement{}, ErrCashMovementNotFound
	}

	return m, err
}

// requireCashAccount refuses a portfolio or a platform that is not the user's.
// The two are one answer: a request that names either wrongly is asking about
// an account it cannot see.
func requireCashAccount(ctx context.Context, tx pgx.Tx, userID, portfolioID, sourceID uuid.UUID) error {
	var owned bool

	if err := tx.QueryRow(ctx, `
		SELECT EXISTS (SELECT 1 FROM portfolios WHERE id = $1 AND user_id = $3)
		   AND EXISTS (SELECT 1 FROM investment_sources WHERE id = $2 AND user_id = $3)
	`, portfolioID, sourceID, userID).Scan(&owned); err != nil {
		return err
	}

	if !owned {
		return ErrPortfolioOrSourceNotFound
	}

	return nil
}

// requireWritablePocket turns the pocket a cash write names into the value
// portfolio_entries.pocket_id takes: nil for the main account, or the pocket
// itself once it is known to be the owner's and part of the account stated.
//
// A pocket carries its own platform and currency, so one that disagrees with
// what the request said is refused rather than quietly preferred: the caller
// believes it is writing somewhere else.
//
// A fixed deposit is refused outright (000048). It holds one deposit at one
// rate for one term, and a movement by hand would make it something else; the
// writes that do move its money — opening it, settling it — hold the pocket
// already and go straight to writeCashMovement, which is why this guard sits in
// front of that rather than inside it.
func requireWritablePocket(ctx context.Context, tx pgx.Tx, userID, pocketID, sourceID uuid.UUID, cur money.Currency) (*uuid.UUID, error) {
	if pocketID == (uuid.UUID{}) {
		return nil, nil
	}

	locked, err := lockCashPocket(ctx, tx, userID, pocketID)
	if err != nil {
		return nil, err
	}

	if locked.sourceID != sourceID || locked.currency != cur {
		return nil, invalidCash("that pocket belongs to another account, so it cannot hold a movement of this one")
	}

	if locked.kind == PocketFixed {
		return nil, fmt.Errorf("%w: cancel it to get the money back", ErrCashPocketFixed)
	}

	return &locked.id, nil
}

// CreateCashMovement records a movement on the balance a platform holds in one
// currency for one portfolio, opening that balance if there is none.
//
// pocketID names the drawer of the account it goes in: the zero UUID is the
// main account, which is every balance there was before pockets existed
// (000047).
//
// The balance it lands on is the one the owner would point at: a cash position
// of that platform and portfolio, in that currency, kept at one unit per unit.
// One the app opened wins over one created by hand, and the older of two of the
// same kind wins over the newer, so repeated deposits keep landing in one row.
// A position that fits none of that — cash bought with another currency, at a
// rate — is left alone: a movement at one unit per unit would change its cost.
//
// The position is locked before its balance is read, so two withdrawals racing
// each other cannot both see the money that only one of them can take.
func (r *PostgresRepository) CreateCashMovement(ctx context.Context, userID, portfolioID, sourceID, pocketID uuid.UUID, in CashMovementInput) (CashMovement, error) {
	var movement CashMovement

	if err := database.WithinTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		if err := requireCashAccount(ctx, tx, userID, portfolioID, sourceID); err != nil {
			return err
		}

		pocket, err := requireWritablePocket(ctx, tx, userID, pocketID, sourceID, in.Currency)
		if err != nil {
			return err
		}

		movement, err = writeCashMovement(ctx, tx, userID, portfolioID, sourceID, pocket, in)

		return err
	}); err != nil {
		return CashMovement{}, err
	}

	return movement, nil
}

// writeCashMovement is CreateCashMovement once the account and the pocket are
// known to be the owner's: find or open the balance, refuse the movement that
// would overdraw it, and record it. MoveCash calls it twice.
func writeCashMovement(ctx context.Context, tx pgx.Tx, userID, portfolioID, sourceID uuid.UUID, pocket *uuid.UUID, in CashMovementInput) (CashMovement, error) {
	txnIn := in.transactionInput(in.Currency)

	entryID, balance, found, err := lockCashEntry(ctx, tx, portfolioID, sourceID, pocket, in.Currency)
	if err != nil {
		return CashMovement{}, err
	}

	if balance.Add(balanceEffect(txnIn.Type, txnIn.Quantity)).IsNeg() {
		return CashMovement{}, fmt.Errorf("%w: the balance holds %s %s", ErrInsufficientCash, balance.String(), in.Currency)
	}

	if !found {
		assetID, err := ensureCashAsset(ctx, tx, in.Currency)
		if err != nil {
			return CashMovement{}, err
		}

		if entryID, err = openCashEntry(ctx, tx, portfolioID, assetID, sourceID, pocket, in.Currency, in.Date); err != nil {
			return CashMovement{}, err
		}
	}

	// TransactionInput.Validate would not catch this. It refuses a missing
	// rate between two currencies, not a rate of one, so a dollar deposit
	// into a position that costs in pesos would pass it and be averaged in as
	// one peso a dollar.
	var (
		costCurrency money.Currency
		atPar        bool
	)
	if err := tx.QueryRow(ctx, `
		SELECT cost_currency, cash_entry_at_par(id) FROM portfolio_entries WHERE id = $1
	`, entryID).Scan(&costCurrency, &atPar); err != nil {
		return CashMovement{}, err
	}

	if costCurrency != in.Currency || !atPar {
		return CashMovement{}, invalidCash("this platform already holds %s as a position that costs in %s at another price; record movements on it from the position", cashTicker(in.Currency), costCurrency)
	}

	settled, err := txnIn.Validate(costCurrency)
	if err != nil {
		return CashMovement{}, err
	}

	var txnID uuid.UUID
	if err := tx.QueryRow(ctx, `
		INSERT INTO transactions (entry_id, type, quantity, price, currency, fx_rate, fees, fees_currency, transaction_date, notes, credited_from, cost_basis)
		VALUES ($1::uuid, $2::transaction_type, $3::numeric, $4::numeric, $5::char(3), $6::numeric, $7::numeric, $8::char(3), $9::date, $10, $11::uuid, $12::numeric)
		RETURNING id
	`, entryID, settled.Type, settled.Quantity.String(), settled.Price.String(), settled.Currency,
		settled.FXRate.String(), settled.Fees.String(), settled.FeesCurrency, settled.TransactionDate, settled.Notes,
		in.linkedTo, decimalParam(in.costBasis)).Scan(&txnID); err != nil {
		return CashMovement{}, err
	}

	return getCashMovement(ctx, tx, userID, txnID)
}

// MoveCash moves money between two cash balances inside one portfolio: two
// drawers of one account — the main account and a pocket, or two pockets — or
// two accounts, which is the transfer from the app that holds the savings to
// the broker that is about to spend them.
//
// It is one transaction with two legs, a withdrawal on one side and a deposit
// on the other, so the money is never in both places or in neither. The two
// flows offset each other — same day, no fee, and the same value once the
// stated rate is applied — so the portfolio's net flow does not move, and
// neither does its return: the money changed hands inside the portfolio, it did
// not arrive or leave.
//
// The destination account is checked on its own: a transfer names two
// platforms, and the second one has to be the owner's too. The destination
// balance need not exist yet — writeCashMovement opens it, the same way a first
// deposit does — so money can be transferred to a broker that has never held
// cash.
//
// Both balances are locked before either is written, in the order of their
// position, so a move each way at the same time queues instead of deadlocking.
func (r *PostgresRepository) MoveCash(ctx context.Context, userID, portfolioID, sourceID uuid.UUID, in CashMoveInput) (CashMove, error) {
	var move CashMove

	// The same defaults the service applies, because the fields below are read
	// here: a caller that named only a drawer means the account the money is
	// already in, and reading its zero values as a platform and an arriving
	// amount would look up nothing and deposit nothing.
	in = in.withDefaults(sourceID)

	if err := database.WithinTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		if err := requireCashAccount(ctx, tx, userID, portfolioID, sourceID); err != nil {
			return err
		}

		if in.ToSource != sourceID {
			if err := requireCashAccount(ctx, tx, userID, portfolioID, in.ToSource); err != nil {
				return err
			}
		}

		from, err := requireWritablePocket(ctx, tx, userID, in.From, sourceID, in.Currency)
		if err != nil {
			return err
		}

		to, err := requireWritablePocket(ctx, tx, userID, in.To, in.ToSource, in.ToCurrency)
		if err != nil {
			return err
		}

		if _, err := tx.Exec(ctx, `
			SELECT pe.id
			FROM portfolio_entries pe
			WHERE pe.portfolio_id = $1
			  AND ((pe.source_id = $2 AND pe.pocket_id IS NOT DISTINCT FROM $3::uuid)
			    OR (pe.source_id = $4 AND pe.pocket_id IS NOT DISTINCT FROM $5::uuid))
			ORDER BY pe.id
			FOR UPDATE OF pe
		`, portfolioID, sourceID, from, in.ToSource, to); err != nil {
			return err
		}

		out, into := in.legs()

		if move.From, err = writeCashMovement(ctx, tx, userID, portfolioID, sourceID, from, out); err != nil {
			return err
		}

		move.To, err = writeCashMovement(ctx, tx, userID, portfolioID, in.ToSource, to, into)

		return err
	}); err != nil {
		return CashMove{}, err
	}

	return move, nil
}

// lockCashEntry finds and locks the balance a cash write lands on, and reads
// what it holds. found is false when there is none yet, which is a balance of
// zero.
//
// pocketID is which drawer of the account: nil is the main one, the balance
// that has no pocket. A pocket has exactly one balance per portfolio — its
// asset and its platform follow from the pocket — so the search for the right
// one among several only ever applies to the main account.
func lockCashEntry(ctx context.Context, tx pgx.Tx, portfolioID, sourceID uuid.UUID, pocketID *uuid.UUID, cur money.Currency) (uuid.UUID, decimal.Decimal, bool, error) {
	var (
		entryID  uuid.UUID
		quantity decimal.Decimal
	)

	err := tx.QueryRow(ctx, `
		SELECT pe.id, pe.quantity
		FROM portfolio_entries pe
		JOIN assets a ON a.id = pe.asset_id
		WHERE pe.portfolio_id = $1
		  AND pe.source_id    = $2
		  AND pe.pocket_id IS NOT DISTINCT FROM $5::uuid
		  AND a.asset_type    = 'cash'
		  AND a.currency      = $3::char(3)
		  AND pe.cost_currency = $3::char(3)
		  AND cash_entry_at_par(pe.id)
		ORDER BY (a.ticker = $4) DESC, pe.created_at
		LIMIT 1
		FOR UPDATE OF pe
	`, portfolioID, sourceID, cur, cashTicker(cur), pocketID).Scan(&entryID, &quantity)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.UUID{}, decimal.Zero, false, nil
	}

	if err != nil {
		return uuid.UUID{}, decimal.Zero, false, err
	}

	return entryID, quantity, true, nil
}

// openCashEntry opens the balance a movement needs and returns its position.
//
// The price is seeded at one and the trigger keeps it there until interest
// averages in at no cost (000042). The two statements differ only in the key
// they resolve a conflict on, and each has to name the predicate of its partial
// index (000047) or Postgres will not use it. On a conflict the position
// already existed: for the main account, opened by hand in another cost
// currency or at another price, which the caller then refuses to write into.
func openCashEntry(ctx context.Context, tx pgx.Tx, portfolioID, assetID, sourceID uuid.UUID, pocketID *uuid.UUID, cur money.Currency, date time.Time) (uuid.UUID, error) {
	conflict := `ON CONFLICT (portfolio_id, asset_id, COALESCE(source_id::TEXT, '')) WHERE pocket_id IS NULL`
	if pocketID != nil {
		conflict = `ON CONFLICT (portfolio_id, pocket_id) WHERE pocket_id IS NOT NULL`
	}

	var entryID uuid.UUID
	err := tx.QueryRow(ctx, `
		INSERT INTO portfolio_entries (portfolio_id, asset_id, source_id, pocket_id, quantity, price, cost_currency, entry_date, notes)
		VALUES ($1::uuid, $2::uuid, $3::uuid, $6::uuid, 0, 1, $4::char(3), $5::date, '')
		`+conflict+`
		DO UPDATE SET updated_at = NOW()
		RETURNING id
	`, portfolioID, assetID, sourceID, cur, date, pocketID).Scan(&entryID)

	return entryID, err
}

// ensureCashAsset returns the catalog row of the balance kept in cur, creating
// it the first time anyone needs it.
//
// It is curated — one row everybody shares, like a listed share — and carries
// a catalog price of one, which is what makes the valuation read the balance as
// its quantity without a provider to ask. The market sync skips cash assets, so
// nothing ever overwrites that price.
//
// The row is looked up again after the insert rather than returned by it: ON
// CONFLICT DO NOTHING returns nothing when a concurrent request created it
// first, and the second read sees that request's row once it has committed.
func ensureCashAsset(ctx context.Context, tx pgx.Tx, cur money.Currency) (uuid.UUID, error) {
	ticker := cashTicker(cur)

	find := func() (uuid.UUID, bool, error) {
		var (
			assetID       uuid.UUID
			assetType     market.AssetType
			assetCurrency money.Currency
		)

		err := tx.QueryRow(ctx, `
			SELECT id, asset_type, currency
			FROM assets
			WHERE ticker = $1 AND COALESCE(exchange, '') = ''
		`, ticker).Scan(&assetID, &assetType, &assetCurrency)
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.UUID{}, false, nil
		}

		if err != nil {
			return uuid.UUID{}, false, err
		}

		// A server error, not the caller's: somebody put a different asset under
		// the ticker this module owns, and writing a balance into it would file
		// money under a share.
		if assetType != market.Cash || assetCurrency != cur {
			return uuid.UUID{}, false, fmt.Errorf("catalog asset %s is a %s in %s, not the %s cash balance", ticker, assetType, assetCurrency, cur)
		}

		return assetID, true, nil
	}

	if assetID, ok, err := find(); err != nil || ok {
		return assetID, err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO assets (ticker, name, asset_type, exchange, currency, current_price, price_updated_at, is_curated, created_at, updated_at)
		VALUES ($1, $2, 'cash', NULL, $3::char(3), 1, NOW(), TRUE, NOW(), NOW())
		ON CONFLICT (ticker, COALESCE(exchange, '')) DO NOTHING
	`, ticker, cashAssetName(cur), cur); err != nil {
		return uuid.UUID{}, err
	}

	assetID, ok, err := find()
	if err == nil && !ok {
		err = fmt.Errorf("cash asset %s was not created", ticker)
	}

	return assetID, err
}

// lockedCashMovement is what an edit or a deletion has to know about the row it
// touches before it is allowed to: what it did to the balance, what the balance
// holds now, and whether it can be rewritten as a cash movement at all.
type lockedCashMovement struct {
	txnType      TransactionType
	quantity     decimal.Decimal
	balance      decimal.Decimal
	costCurrency money.Currency
	editable     bool
	// fixed is whether it sits in a fixed deposit. Its rows are the deposit that
	// opened it and the interest it earned, and neither is the owner's to
	// rewrite: what the deposit says is settled by cancelling it.
	fixed bool
}

func lockCashMovement(ctx context.Context, tx pgx.Tx, userID, txnID uuid.UUID) (lockedCashMovement, error) {
	var m lockedCashMovement

	err := tx.QueryRow(ctx, `
		SELECT t.type, t.quantity, pe.quantity, pe.cost_currency, `+cashMovementEditable+`,
		       EXISTS (SELECT 1 FROM cash_pockets pk WHERE pk.id = pe.pocket_id AND pk.kind = 'fixed')
		FROM transactions t
		JOIN portfolio_entries pe ON pe.id = t.entry_id
		JOIN portfolios p         ON p.id = pe.portfolio_id
		JOIN assets a             ON a.id = pe.asset_id
		WHERE t.id = $1 AND p.user_id = $2 AND a.asset_type = 'cash'
		FOR UPDATE OF pe
	`, txnID, userID).Scan(&m.txnType, &m.quantity, &m.balance, &m.costCurrency, &m.editable, &m.fixed)
	if errors.Is(err, pgx.ErrNoRows) {
		return m, ErrCashMovementNotFound
	}

	if err == nil && m.fixed {
		return m, fmt.Errorf("%w: cancel it or delete it whole", ErrCashPocketFixed)
	}

	return m, err
}

// UpdateCashMovement rewrites a movement in place, on the balance it is already
// on and in that balance's currency.
//
// The balance it would leave is checked by taking the old movement out and
// putting the new one in: turning a deposit into a withdrawal, or shrinking a
// deposit the balance has since spent, can empty an account as surely as a new
// withdrawal can.
func (r *PostgresRepository) UpdateCashMovement(ctx context.Context, userID, txnID uuid.UUID, in CashMovementInput) (CashMovement, error) {
	var movement CashMovement

	if err := database.WithinTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		current, err := lockCashMovement(ctx, tx, userID, txnID)
		if err != nil {
			return err
		}

		if current.txnType.isCashLinked() {
			return ErrCashCreditLinked
		}

		if !current.editable {
			return ErrCashMovementNotEditable
		}

		settled, err := in.transactionInput(current.costCurrency).Validate(current.costCurrency)
		if err != nil {
			return err
		}

		after := current.balance.
			Add(balanceEffect(current.txnType, current.quantity).Neg()).
			Add(balanceEffect(settled.Type, settled.Quantity))
		if after.IsNeg() {
			return fmt.Errorf("%w: the change would leave the balance at %s %s", ErrInsufficientCash, after.String(), current.costCurrency)
		}

		if _, err := tx.Exec(ctx, `
			UPDATE transactions SET
				type             = $1::transaction_type,
				quantity         = $2::numeric,
				price            = $3::numeric,
				currency         = $4::char(3),
				fx_rate          = $5::numeric,
				fees             = $6::numeric,
				fees_currency    = $7::char(3),
				transaction_date = $8::date,
				notes            = $9,
				updated_at       = NOW()
			WHERE id = $10
		`, settled.Type, settled.Quantity.String(), settled.Price.String(), settled.Currency, settled.FXRate.String(),
			settled.Fees.String(), settled.FeesCurrency, settled.TransactionDate, settled.Notes, txnID); err != nil {
			return err
		}

		movement, err = getCashMovement(ctx, tx, userID, txnID)

		return err
	}); err != nil {
		return CashMovement{}, err
	}

	return movement, nil
}

// DeleteCashMovement removes a movement unless that would take the balance
// below zero — deleting a deposit whose money a later withdrawal already took.
// Any movement on a cash position can go, including one the cash screens cannot
// edit: removing a row does not reprice it. The one exception is a row that
// belongs to another transaction — a dividend's or a sale's credit, a
// purchase's debit — which goes with it.
func (r *PostgresRepository) DeleteCashMovement(ctx context.Context, userID, txnID uuid.UUID) error {
	return database.WithinTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		current, err := lockCashMovement(ctx, tx, userID, txnID)
		if err != nil {
			return err
		}

		if current.txnType.isCashLinked() {
			return ErrCashCreditLinked
		}

		after := current.balance.Add(balanceEffect(current.txnType, current.quantity).Neg())
		if after.IsNeg() {
			return fmt.Errorf("%w: removing it would leave the balance at %s %s", ErrInsufficientCash, after.String(), current.costCurrency)
		}

		_, err = tx.Exec(ctx, `DELETE FROM transactions WHERE id = $1`, txnID)

		return err
	})
}

// requireTypeAllowed refuses a transaction whose type the position's asset
// cannot take — cash_interest on anything but a cash balance — and, on any
// position, a row that belongs to another transaction: the credit of a dividend
// or a sale, the debit of a purchase, which only that transaction writes. The
// generic transaction writers call it; the cash writers only ever reach cash
// positions.
func requireTypeAllowed(ctx context.Context, tx pgx.Tx, entryID uuid.UUID, t TransactionType) error {
	if t.isCashLinked() {
		return ErrCashCreditLinked
	}

	var assetType market.AssetType

	if err := tx.QueryRow(ctx, `
		SELECT a.asset_type
		FROM portfolio_entries pe
		JOIN assets a ON a.id = pe.asset_id
		WHERE pe.id = $1
	`, entryID).Scan(&assetType); err != nil {
		return err
	}

	if !t.AllowedOn(assetType) {
		return ErrCashInterestOutsideCash
	}

	return nil
}

// requireWritableEntry refuses a position that belongs to a fixed deposit
// (000048). The generic transaction writers call it, next to requireTypeAllowed
// and for the same reason: the positions screen reaches every position there
// is, and a deposit recorded there would stop being one deposit at one rate.
func requireWritableEntry(ctx context.Context, tx pgx.Tx, entryID uuid.UUID) error {
	var fixed bool

	if err := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM portfolio_entries pe
			JOIN cash_pockets pk ON pk.id = pe.pocket_id
			WHERE pe.id = $1 AND pk.kind = 'fixed'
		)
	`, entryID).Scan(&fixed); err != nil {
		return err
	}

	if fixed {
		return fmt.Errorf("%w: cancel it or delete it whole", ErrCashPocketFixed)
	}

	return nil
}
