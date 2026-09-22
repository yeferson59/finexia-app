package portfolio

import (
	"context"
	"errors"
	"testing"
	"time"

	"uuid"

	"github.com/yeferson59/gofinance/v2/money"
)

// Pockets, migration 000047. Same database contract as postgres_cash_db_test.go.

// pocket opens a flexible pocket on the fixture's account.
func (f cashFixture) pocket(t *testing.T, name string) CashPocket {
	t.Helper()

	pocket, err := f.repo.CreateCashPocket(context.Background(), f.userID, NewCashPocketInput{
		SourceID: f.sourceID,
		Currency: money.USD,
		Name:     name,
	})
	if err != nil {
		t.Fatalf("CreateCashPocket(%q): %v", name, err)
	}

	return pocket
}

// moveInto records a movement in one drawer of the fixture's account.
func (f cashFixture) moveInto(t *testing.T, pocketID uuid.UUID, kind CashMovementKind, amount string, date time.Time) CashMovement {
	t.Helper()

	m, err := f.repo.CreateCashMovement(context.Background(), f.userID, f.portfolioID, f.sourceID, pocketID, CashMovementInput{
		Kind: kind, Amount: mustDecimal(t, amount), Currency: money.USD, Date: date,
	})
	if err != nil {
		t.Fatalf("CreateCashMovement(%s %s in %v): %v", kind, amount, pocketID, err)
	}

	return m
}

// rateOn records a USD rate on one drawer of the fixture's account.
func (f cashFixture) rateOn(t *testing.T, pocketID uuid.UUID, pct string, from time.Time) CashRate {
	t.Helper()

	rate, err := f.repo.CreateCashRate(context.Background(), f.userID, NewCashRateInput{
		SourceID:      f.sourceID,
		Currency:      money.USD,
		PocketID:      pocketID,
		EffectiveFrom: from,
		CashRateInput: CashRateInput{
			AnnualRatePct:  mustDecimal(t, pct),
			WithholdingPct: mustDecimal(t, "0"),
			Posting:        PostingDaily,
		},
	})
	if err != nil {
		t.Fatalf("CreateCashRate(%s on %v): %v", pct, pocketID, err)
	}

	return rate
}

// A pocket and the main account are two balances of one platform inside one
// portfolio. The partial unique indexes let them coexist: the main account
// keeps the key it always had, and a pocket is one balance per portfolio.
func TestCashPocketAndMainAccountShareAPortfolio(t *testing.T) {
	f := newCashFixture(t)
	ctx := context.Background()

	viajes := f.pocket(t, "Viajes")
	main := f.mustMove(t, CashKindDeposit, "6000", cashDay)
	inPocket := f.moveInto(t, viajes.ID, CashKindDeposit, "2000", cashDay)

	if main.EntryID == inPocket.EntryID {
		t.Fatalf("both movements landed on %v, want two balances", main.EntryID)
	}

	balances, err := f.repo.GetCashBalancesByUserID(ctx, f.userID, money.XXX)
	if err != nil {
		t.Fatalf("GetCashBalancesByUserID: %v", err)
	}
	if len(balances) != 2 {
		t.Fatalf("balances = %d, want two", len(balances))
	}

	for _, b := range balances {
		switch b.EntryID {
		case main.EntryID:
			sameAmount(t, "the main account", b.Balance, "6000")
			if b.PocketID != nil || b.PocketName != "" {
				t.Errorf("the main account reads as pocket %v %q, want none", b.PocketID, b.PocketName)
			}
		case inPocket.EntryID:
			sameAmount(t, "the pocket", b.Balance, "2000")
			if b.PocketID == nil || *b.PocketID != viajes.ID || b.PocketName != "Viajes" || b.PocketKind != PocketFlexible {
				t.Errorf("pocket = %v %q %q, want Viajes, flexible", b.PocketID, b.PocketName, b.PocketKind)
			}
		default:
			t.Errorf("unexpected balance %+v", b)
		}
	}

	// A second deposit in each drawer lands on the balance that is already
	// there, which is what the two ON CONFLICT keys resolve.
	if again := f.mustMove(t, CashKindDeposit, "500", cashDay); again.EntryID != main.EntryID {
		t.Errorf("the second deposit opened %v, want the main account %v", again.EntryID, main.EntryID)
	}
	if again := f.moveInto(t, viajes.ID, CashKindDeposit, "500", cashDay); again.EntryID != inPocket.EntryID {
		t.Errorf("the second deposit opened %v, want the pocket %v", again.EntryID, inPocket.EntryID)
	}
}

// The account pays 8 % and its pocket 10 %, and each earns on what its own
// balances hold. The platform adds the two: a pocket never left it.
func TestCashPocketEarnsItsOwnRate(t *testing.T) {
	f := newCashFixture(t)
	ctx := context.Background()

	viajes := f.pocket(t, "Viajes")
	main := f.mustMove(t, CashKindDeposit, "10000", cashDay)
	inPocket := f.moveInto(t, viajes.ID, CashKindDeposit, "10000", cashDay)

	mainRate := f.rateOn(t, uuid.UUID{}, "8", cashDay)
	pocketRate := f.rateOn(t, viajes.ID, "10", cashDay)

	// Two versions of the same platform and currency, which before pockets one
	// unique key would have refused.
	if mainRate.PocketID != nil {
		t.Errorf("the main account's rate reads as pocket %v, want none", mainRate.PocketID)
	}
	if pocketRate.PocketID == nil || *pocketRate.PocketID != viajes.ID || pocketRate.PocketName != "Viajes" {
		t.Errorf("the pocket's rate = %v %q, want Viajes", pocketRate.PocketID, pocketRate.PocketName)
	}
	if !mainRate.Latest || !pocketRate.Latest {
		t.Errorf("latest = %v and %v, want both: each is the latest of its own drawer", mainRate.Latest, pocketRate.Latest)
	}

	f.accrue(t, main.EntryID, mainRate.ID, cashDay)
	f.accrue(t, inPocket.EntryID, pocketRate.ID, cashDay)

	// 10 000 × ((1.08)^(1/365) − 1) = 2.1087, 10 000 × ((1.10)^(1/365) − 1) = 2.6116.
	sameAmount(t, "the main account", f.balanceOf(t, main.EntryID).Balance, "10002.11")
	sameAmount(t, "the pocket", f.balanceOf(t, inPocket.EntryID).Balance, "10002.61")

	// The platform adds the two, as it did before either was a pocket.
	platforms, err := f.repo.GetPlatformsWithStats(ctx, f.userID, money.USD)
	if err != nil {
		t.Fatalf("GetPlatformsWithStats: %v", err)
	}
	if len(platforms) != 1 {
		t.Fatalf("platforms = %d, want one", len(platforms))
	}
	sameAmount(t, "the platform's value", platforms[0].MarketValue, "20004.72")
}

// A rate given to a pocket leaves the main account's alone, and each day is
// computed at the rate of its own drawer. The nightly run takes them both.
func TestCashAccrualTargetsSplitByPocket(t *testing.T) {
	f := newCashFixture(t)
	ctx := context.Background()

	viajes := f.pocket(t, "Viajes")
	main := f.mustMove(t, CashKindDeposit, "10000", cashDay)
	inPocket := f.moveInto(t, viajes.ID, CashKindDeposit, "10000", cashDay)

	mainRate := f.rateOn(t, uuid.UUID{}, "8", cashDay)
	pocketRate := f.rateOn(t, viajes.ID, "10", cashDay)

	targets, err := f.repo.GetCashAccrualTargets(ctx, cashDay, CashAccrualFilter{UserID: f.userID})
	if err != nil {
		t.Fatalf("GetCashAccrualTargets: %v", err)
	}
	if len(targets) != 2 {
		t.Fatalf("targets = %d, want one per drawer", len(targets))
	}

	for _, target := range targets {
		if len(target.Versions) != 1 {
			t.Fatalf("balance %v has %d versions, want only its own drawer's", target.EntryID, len(target.Versions))
		}

		want := mainRate.ID
		if target.EntryID == inPocket.EntryID {
			want = pocketRate.ID
		}

		if target.Versions[0].ID != want {
			t.Errorf("balance %v earns at %v, want %v", target.EntryID, target.Versions[0].ID, want)
		}
	}

	// Narrowed to the main account, the pocket is left out — and the other way
	// round. "Every drawer" and "the one with no pocket" are different runs.
	onMain, err := f.repo.GetCashAccrualTargets(ctx, cashDay, CashAccrualFilter{UserID: f.userID}.OnPocket(uuid.UUID{}))
	if err != nil {
		t.Fatalf("GetCashAccrualTargets on the main account: %v", err)
	}
	if len(onMain) != 1 || onMain[0].EntryID != main.EntryID {
		t.Errorf("the main account's targets = %+v, want only %v", onMain, main.EntryID)
	}

	onPocket, err := f.repo.GetCashAccrualTargets(ctx, cashDay, CashAccrualFilter{UserID: f.userID}.OnPocket(viajes.ID))
	if err != nil {
		t.Fatalf("GetCashAccrualTargets on the pocket: %v", err)
	}
	if len(onPocket) != 1 || onPocket[0].EntryID != inPocket.EntryID {
		t.Errorf("the pocket's targets = %+v, want only %v", onPocket, inPocket.EntryID)
	}
}

// Tiers are of a drawer, not of the platform: the account's own money decides
// which step it is on, and the pocket's does not push it up one.
func TestCashPocketTiersCountOnlyTheirOwnDrawer(t *testing.T) {
	f := newCashFixture(t)

	viajes := f.pocket(t, "Viajes")
	main := f.mustMove(t, CashKindDeposit, "4000", cashDay)
	f.moveInto(t, viajes.ID, CashKindDeposit, "4000", cashDay)

	rate, err := f.repo.CreateCashRate(context.Background(), f.userID, NewCashRateInput{
		SourceID:      f.sourceID,
		Currency:      money.USD,
		EffectiveFrom: cashDay,
		CashRateInput: CashRateInput{
			AnnualRatePct:  mustDecimal(t, "12"),
			WithholdingPct: mustDecimal(t, "0"),
			Posting:        PostingDaily,
			Tiers:          tierSteps(t, "5000", "8"),
		},
	})
	if err != nil {
		t.Fatalf("CreateCashRate: %v", err)
	}

	f.accrue(t, main.EntryID, rate.ID, cashDay)

	// The two drawers hold 8 000 between them, but the main account holds 4 000,
	// so the whole of it sits in the first step: 4 000 × ((1.12)^(1/365) − 1).
	row, _ := f.accrualOn(t, main.EntryID, cashDay)
	sameAmount(t, "net", row.net, "1.24215102")
	sameAmount(t, "the rate of the day", f.accrualRate(t, main.EntryID, cashDay), "0.12")
}

// Moving money between drawers is one transaction with two legs. The flows
// cancel out, so the portfolio's net flow — and its return — does not move.
func TestMoveCashLeavesTheNetFlowAlone(t *testing.T) {
	f := newCashFixture(t)
	ctx := context.Background()

	viajes := f.pocket(t, "Viajes")
	main := f.mustMove(t, CashKindDeposit, "6000", cashDay)

	move, err := f.repo.MoveCash(ctx, f.userID, f.portfolioID, f.sourceID, CashMoveInput{
		Currency: money.USD,
		To:       viajes.ID,
		Amount:   mustDecimal(t, "2000"),
		Date:     cashDay,
		Notes:    "para el viaje",
	})
	if err != nil {
		t.Fatalf("MoveCash: %v", err)
	}

	if move.From.Kind != CashKindWithdrawal || move.To.Kind != CashKindDeposit {
		t.Errorf("legs = %s and %s, want a withdrawal and a deposit", move.From.Kind, move.To.Kind)
	}
	if move.From.EntryID != main.EntryID {
		t.Errorf("the withdrawal left %v, want the main account %v", move.From.EntryID, main.EntryID)
	}

	sameAmount(t, "the main account", f.balanceOf(t, main.EntryID).Balance, "4000")
	sameAmount(t, "the pocket", f.balanceOf(t, move.To.EntryID).Balance, "2000")

	// The two legs are a transfer_out and a transfer_in of the same amount on
	// the same day, so what the portfolio took in is the first deposit alone.
	var netFlow string
	if err := f.pool.QueryRow(ctx, `
		SELECT ROUND(COALESCE(SUM(
			CASE t.type
				WHEN 'transfer_in'  THEN  t.quantity * t.price * t.fx_rate
				WHEN 'transfer_out' THEN -(t.quantity * t.price * t.fx_rate)
				ELSE 0
			END
		), 0), 8)::text
		FROM transactions t
		JOIN portfolio_entries pe ON pe.id = t.entry_id
		WHERE pe.portfolio_id = $1
	`, f.portfolioID).Scan(&netFlow); err != nil {
		t.Fatalf("read the net flow: %v", err)
	}
	sameAmount(t, "net flow", netFlow, "6000")

	// More than it holds is refused, and nothing of the move is written.
	_, err = f.repo.MoveCash(ctx, f.userID, f.portfolioID, f.sourceID, CashMoveInput{
		Currency: money.USD, To: viajes.ID, Amount: mustDecimal(t, "99999"), Date: cashDay,
	})
	if !errors.Is(err, ErrInsufficientCash) {
		t.Errorf("moving more than it holds = %v, want ErrInsufficientCash", err)
	}
	sameAmount(t, "the pocket after the refusal", f.balanceOf(t, move.To.EntryID).Balance, "2000")
}

// Money moves to another platform, converting on the way: the pesos the savings
// app held leave it and dollars arrive at the broker, which is the transfer
// that happens before a purchase is funded from the broker's own cash.
//
// The broker has never held cash here, so the move also has to open the balance
// it lands on, the way a first deposit would.
func TestMoveCashCrossesPlatformsAndCurrencies(t *testing.T) {
	f := newCashFixture(t)
	ctx := context.Background()

	broker := uuid.New()
	f.exec(t, `INSERT INTO investment_sources (id, user_id, name, source_type)
	           VALUES ($1, $2, 'broker', 'broker')`, broker, f.userID)

	savings, err := f.repo.CreateCashMovement(ctx, f.userID, f.portfolioID, f.sourceID, uuid.UUID{}, CashMovementInput{
		Kind:     CashKindDeposit,
		Amount:   mustDecimal(t, "600000"),
		Currency: money.COP,
		Date:     cashDay,
	})
	if err != nil {
		t.Fatalf("the first deposit in pesos: %v", err)
	}

	move, err := f.repo.MoveCash(ctx, f.userID, f.portfolioID, f.sourceID, CashMoveInput{
		Currency:   money.COP,
		ToSource:   broker,
		ToCurrency: money.USD,
		Amount:     mustDecimal(t, "400000"),
		ToAmount:   mustDecimal(t, "98.50"),
		Date:       cashDay,
		Notes:      "para comprar AAPL",
	})
	if err != nil {
		t.Fatalf("MoveCash across platforms: %v", err)
	}

	if move.From.Currency != money.COP || move.To.Currency != money.USD {
		t.Errorf("legs in %s and %s, want COP and USD", move.From.Currency, move.To.Currency)
	}
	if move.From.SourceID != f.sourceID || move.To.SourceID != broker {
		t.Errorf("legs on %v and %v, want the bank %v and the broker %v",
			move.From.SourceID, move.To.SourceID, f.sourceID, broker)
	}

	sameAmount(t, "what the savings app keeps", f.balanceOf(t, savings.EntryID).Balance, "200000")
	// To the cent the statement says: the deposit is the amount the move stated,
	// not one recomputed from a rate.
	sameAmount(t, "what reached the broker", f.balanceOf(t, move.To.EntryID).Balance, "98.5")

	// More than the origin holds is refused whole: neither leg is written, so
	// the broker does not end up with money the bank never sent.
	_, err = f.repo.MoveCash(ctx, f.userID, f.portfolioID, f.sourceID, CashMoveInput{
		Currency: money.COP, ToSource: broker, ToCurrency: money.USD,
		Amount: mustDecimal(t, "9000000"), ToAmount: mustDecimal(t, "2216"), Date: cashDay,
	})
	if !errors.Is(err, ErrInsufficientCash) {
		t.Errorf("moving more than it holds = %v, want ErrInsufficientCash", err)
	}
	sameAmount(t, "the broker after the refusal", f.balanceOf(t, move.To.EntryID).Balance, "98.5")

	// A platform that is not the owner's is not a destination, whatever it holds.
	stranger := uuid.New()
	f.exec(t, `INSERT INTO users (id, name, email, role_id, preferred_currency)
	           VALUES ($1, 'somebody else', $2, (SELECT id FROM roles WHERE name = 'customer'), 'USD')`,
		stranger, stranger.String()+"@probe.test")
	theirs := uuid.New()
	f.exec(t, `INSERT INTO investment_sources (id, user_id, name, source_type)
	           VALUES ($1, $2, 'their broker', 'broker')`, theirs, stranger)

	_, err = f.repo.MoveCash(ctx, f.userID, f.portfolioID, f.sourceID, CashMoveInput{
		Currency: money.COP, ToSource: theirs, ToCurrency: money.COP,
		Amount: mustDecimal(t, "1000"), Date: cashDay,
	})
	if !errors.Is(err, ErrPortfolioOrSourceNotFound) {
		t.Errorf("moving to somebody else's platform = %v, want ErrPortfolioOrSourceNotFound", err)
	}

	f.exec(t, `DELETE FROM investment_sources WHERE user_id = $1`, stranger)
	f.exec(t, `DELETE FROM users WHERE id = $1`, stranger)
}

// A pocket is deleted while it never held anything, and held back once it does.
func TestCashPocketIsDeletedOnlyWhileEmpty(t *testing.T) {
	f := newCashFixture(t)
	ctx := context.Background()

	viajes := f.pocket(t, "Viajes")
	if err := f.repo.DeleteCashPocket(ctx, f.userID, viajes.ID); err != nil {
		t.Fatalf("deleting an unused pocket: %v", err)
	}

	held := f.pocket(t, "Viajes")
	deposit := f.moveInto(t, held.ID, CashKindDeposit, "2000", cashDay)

	if err := f.repo.DeleteCashPocket(ctx, f.userID, held.ID); !errors.Is(err, ErrCashPocketNotEmpty) {
		t.Errorf("deleting a pocket with money = %v, want ErrCashPocketNotEmpty", err)
	}

	// Emptied but with its history, it is still refused: the movements happened,
	// and deleting the pocket would delete them.
	out := f.moveInto(t, held.ID, CashKindWithdrawal, "2000", cashDay)

	pockets, err := f.repo.GetCashPocketsByUserID(ctx, f.userID)
	if err != nil {
		t.Fatalf("GetCashPocketsByUserID: %v", err)
	}
	if len(pockets) != 1 {
		t.Fatalf("pockets = %d, want one", len(pockets))
	}
	sameAmount(t, "what it holds", pockets[0].Balance, "0")
	if pockets[0].Movements != 2 {
		t.Errorf("movements = %d, want the two it took", pockets[0].Movements)
	}

	if err := f.repo.DeleteCashPocket(ctx, f.userID, held.ID); !errors.Is(err, ErrCashPocketNotEmpty) {
		t.Errorf("deleting an emptied pocket = %v, want ErrCashPocketNotEmpty", err)
	}

	// With its whole history gone, the balance it opened is an empty shell and
	// goes with it.
	for _, id := range []uuid.UUID{out.ID, deposit.ID} {
		if err := f.repo.DeleteCashMovement(ctx, f.userID, id); err != nil {
			t.Fatalf("DeleteCashMovement: %v", err)
		}
	}

	if err := f.repo.DeleteCashPocket(ctx, f.userID, held.ID); err != nil {
		t.Fatalf("deleting a pocket with nothing left: %v", err)
	}

	var entries int
	if err := f.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM portfolio_entries WHERE pocket_id = $1
	`, held.ID).Scan(&entries); err != nil {
		t.Fatalf("count the balances left: %v", err)
	}
	if entries != 0 {
		t.Errorf("balances left = %d, want none", entries)
	}
}

// Two pockets of one account cannot share a name; two accounts can. Renaming
// runs into the same key.
func TestCashPocketNamesAreUniquePerAccount(t *testing.T) {
	f := newCashFixture(t)
	ctx := context.Background()

	f.pocket(t, "Viajes")

	if _, err := f.repo.CreateCashPocket(ctx, f.userID, NewCashPocketInput{
		SourceID: f.sourceID, Currency: money.USD, Name: "Viajes",
	}); !errors.Is(err, ErrCashPocketNameTaken) {
		t.Errorf("a second Viajes = %v, want ErrCashPocketNameTaken", err)
	}

	// The same name in another currency of the platform is another account.
	if _, err := f.repo.CreateCashPocket(ctx, f.userID, NewCashPocketInput{
		SourceID: f.sourceID, Currency: money.COP, Name: "Viajes",
	}); err != nil {
		t.Errorf("Viajes in COP: %v, want it allowed", err)
	}

	casa := f.pocket(t, "Casa")
	if _, err := f.repo.RenameCashPocket(ctx, f.userID, casa.ID, RenameCashPocketInput{Name: "Viajes"}); !errors.Is(err, ErrCashPocketNameTaken) {
		t.Errorf("renaming onto a taken name = %v, want ErrCashPocketNameTaken", err)
	}

	renamed, err := f.repo.RenameCashPocket(ctx, f.userID, casa.ID, RenameCashPocketInput{Name: "  La casa  "})
	if err != nil {
		t.Fatalf("RenameCashPocket: %v", err)
	}
	if renamed.Name != "La casa" {
		t.Errorf("name = %q, want it trimmed to %q", renamed.Name, "La casa")
	}
}

// Someone else's pocket is not found, and one of another account cannot take
// this account's movements or its rate.
func TestCashPocketOfAnotherAccountIsRefused(t *testing.T) {
	f := newCashFixture(t)
	ctx := context.Background()

	other := uuid.New()
	f.exec(t, `INSERT INTO investment_sources (id, user_id, name, source_type)
	           VALUES ($1, $2, 'broker', 'broker')`, other, f.userID)

	elsewhere, err := f.repo.CreateCashPocket(ctx, f.userID, NewCashPocketInput{
		SourceID: other, Currency: money.USD, Name: "Viajes",
	})
	if err != nil {
		t.Fatalf("CreateCashPocket on the other platform: %v", err)
	}

	if _, err := f.repo.CreateCashMovement(ctx, f.userID, f.portfolioID, f.sourceID, elsewhere.ID, CashMovementInput{
		Kind: CashKindDeposit, Amount: mustDecimal(t, "100"), Currency: money.USD, Date: cashDay,
	}); !errors.Is(err, ErrInvalidCashMovement) {
		t.Errorf("a movement into another account's pocket = %v, want ErrInvalidCashMovement", err)
	}

	if _, err := f.repo.CreateCashRate(ctx, f.userID, NewCashRateInput{
		SourceID: f.sourceID, Currency: money.USD, PocketID: elsewhere.ID, EffectiveFrom: cashDay,
		CashRateInput: CashRateInput{AnnualRatePct: mustDecimal(t, "8"), Posting: PostingDaily},
	}); !errors.Is(err, ErrInvalidCashRate) {
		t.Errorf("a rate on another account's pocket = %v, want ErrInvalidCashRate", err)
	}

	// A pocket that is not the caller's is not found at all.
	stranger := uuid.New()
	if _, err := f.repo.RenameCashPocket(ctx, stranger, elsewhere.ID, RenameCashPocketInput{Name: "x"}); !errors.Is(err, ErrCashPocketNotFound) {
		t.Errorf("renaming someone else's pocket = %v, want ErrCashPocketNotFound", err)
	}
	if err := f.repo.DeleteCashPocket(ctx, stranger, elsewhere.ID); !errors.Is(err, ErrCashPocketNotFound) {
		t.Errorf("deleting someone else's pocket = %v, want ErrCashPocketNotFound", err)
	}
}

// Recalculating one drawer leaves the other as it was: each earns its own rate
// on its own balances.
func TestRecalculateCashInterestTakesOneDrawer(t *testing.T) {
	f := newCashFixture(t)
	ctx := context.Background()

	viajes := f.pocket(t, "Viajes")
	main := f.mustMove(t, CashKindDeposit, "10000", cashDay)
	inPocket := f.moveInto(t, viajes.ID, CashKindDeposit, "10000", cashDay)

	mainRate := f.rateOn(t, uuid.UUID{}, "8", cashDay)
	pocketRate := f.rateOn(t, viajes.ID, "10", cashDay)

	f.accrue(t, main.EntryID, mainRate.ID, cashDay)
	f.accrue(t, inPocket.EntryID, pocketRate.ID, cashDay)

	cleared, err := f.repo.ClearCashInterest(ctx, CashAccrualFilter{
		UserID: f.userID, SourceID: f.sourceID, Currency: money.USD,
	}.OnPocket(viajes.ID), cashDay)
	if err != nil {
		t.Fatalf("ClearCashInterest: %v", err)
	}
	if cleared.Balances != 1 || cleared.Days != 1 {
		t.Errorf("cleared = %+v, want one day on one balance", cleared)
	}

	if n := f.countInterest(t, main.EntryID); n != 1 {
		t.Errorf("the main account kept %d credits, want its own", n)
	}
	if n := f.countInterest(t, inPocket.EntryID); n != 0 {
		t.Errorf("the pocket kept %d credits, want none", n)
	}
}
