package portfolio

import (
	"errors"
	"testing"
	"time"

	"uuid"
)

func planDay(month time.Month, day int) time.Time {
	return time.Date(2026, month, day, 0, 0, 0, 0, time.UTC)
}

// The example of §4 of docs/PLAN_FONDOS_INVERSION.md, figure by figure: a fund
// followed only by the money that went in and out and the balances the app
// showed.
func TestReplayBalanceFundPlanExample(t *testing.T) {
	entry := uuid.New()
	july, august, withdrawal := uuid.New(), uuid.New(), uuid.New()

	replay, err := replayBalanceFund(
		[]fundMovementFact{
			// Out of order on purpose: the replay sorts by day.
			{TxnID: withdrawal, EntryID: entry, Date: planDay(time.September, 20), Withdrawal: true, Amount: mustDecimal(t, "2000000")},
			{TxnID: july, EntryID: entry, Date: planDay(time.July, 1), Amount: mustDecimal(t, "10000000")},
			{TxnID: august, EntryID: entry, Date: planDay(time.August, 15), Amount: mustDecimal(t, "5000000")},
		},
		[]fundBalanceFact{
			{Date: planDay(time.September, 30), Balance: mustDecimal(t, "13050000")},
			{Date: planDay(time.July, 31), Balance: mustDecimal(t, "10080000")},
			{Date: planDay(time.August, 31), Balance: mustDecimal(t, "15110000")},
		},
	)
	if err != nil {
		t.Fatalf("replayBalanceFund: %v", err)
	}

	wantMoves := []struct {
		id              uuid.UUID
		quantity, price string
	}{
		{july, "100000", "100"},
		{august, "49603.17460317", "100.8"},
		{withdrawal, "19801.87618916", "101.0005305"},
	}

	if len(replay.Movements) != len(wantMoves) {
		t.Fatalf("movements = %d, want %d", len(replay.Movements), len(wantMoves))
	}

	for i, want := range wantMoves {
		got := replay.Movements[i]
		if got.TxnID != want.id {
			t.Fatalf("movement %d is %v, want %v", i, got.TxnID, want.id)
		}

		sameAmount(t, "quantity", got.Quantity.String(), want.quantity)
		sameAmount(t, "price", got.Price.String(), want.price)
	}

	wantMarks := []string{"100.8", "101.0005305", "100.53828551"}
	if len(replay.Marks) != len(wantMarks) {
		t.Fatalf("marks = %d, want %d", len(replay.Marks), len(wantMarks))
	}

	for i, want := range wantMarks {
		if replay.Marks[i].Skipped {
			t.Fatalf("mark %d skipped", i)
		}

		sameAmount(t, "unit value", replay.Marks[i].UnitValue.String(), want)
	}
}

// A movement and a balance on the same day: the balance is the close of the
// day, so it already holds the contribution, and the contribution trades at
// the balance before.
func TestReplayBalanceFundSameDay(t *testing.T) {
	entry := uuid.New()

	replay, err := replayBalanceFund(
		[]fundMovementFact{
			{TxnID: uuid.New(), EntryID: entry, Date: planDay(time.July, 1), Amount: mustDecimal(t, "1000")},
			{TxnID: uuid.New(), EntryID: entry, Date: planDay(time.July, 31), Amount: mustDecimal(t, "1000")},
		},
		[]fundBalanceFact{
			{Date: planDay(time.July, 30), Balance: mustDecimal(t, "1100")},
			{Date: planDay(time.July, 31), Balance: mustDecimal(t, "2100")},
		},
	)
	if err != nil {
		t.Fatalf("replayBalanceFund: %v", err)
	}

	// 1000 / 110 units, and on the 31st the 2100 closes on 19.09090909 units.
	sameAmount(t, "second contribution", replay.Movements[1].Quantity.String(), "9.09090909")
	sameAmount(t, "price", replay.Movements[1].Price.String(), "110")
	sameAmount(t, "unit value on the 31st", replay.Marks[1].UnitValue.String(), "110.00000001")
}

// Withdrawing everything takes every unit whatever the replay makes of them,
// at the price the money says; a withdrawal typed as the whole balance lands a
// rounding away and still takes exactly all of it.
func TestReplayBalanceFundWithdrawals(t *testing.T) {
	entry := uuid.New()
	opening := fundMovementFact{TxnID: uuid.New(), EntryID: entry, Date: planDay(time.July, 1), Amount: mustDecimal(t, "1000")}
	balance := fundBalanceFact{Date: planDay(time.July, 31), Balance: mustDecimal(t, "1003")}

	t.Run("all", func(t *testing.T) {
		replay, err := replayBalanceFund([]fundMovementFact{
			opening,
			{TxnID: uuid.New(), EntryID: entry, Date: planDay(time.August, 5), Withdrawal: true, All: true, Amount: mustDecimal(t, "1004.5")},
		}, []fundBalanceFact{balance})
		if err != nil {
			t.Fatalf("replayBalanceFund: %v", err)
		}

		sameAmount(t, "quantity", replay.Movements[1].Quantity.String(), "10")
		sameAmount(t, "price", replay.Movements[1].Price.String(), "100.45")
	})

	t.Run("the balance typed as money", func(t *testing.T) {
		replay, err := replayBalanceFund([]fundMovementFact{
			opening,
			{TxnID: uuid.New(), EntryID: entry, Date: planDay(time.August, 5), Withdrawal: true, Amount: mustDecimal(t, "1003.004")},
		}, []fundBalanceFact{balance})
		if err != nil {
			t.Fatalf("replayBalanceFund: %v", err)
		}

		sameAmount(t, "quantity", replay.Movements[1].Quantity.String(), "10")
	})

	t.Run("more than there is", func(t *testing.T) {
		_, err := replayBalanceFund([]fundMovementFact{
			opening,
			{TxnID: uuid.New(), EntryID: entry, Date: planDay(time.August, 5), Withdrawal: true, Amount: mustDecimal(t, "1100")},
		}, []fundBalanceFact{balance})
		if !errors.Is(err, ErrFundNotEnoughUnits) {
			t.Fatalf("replayBalanceFund = %v, want ErrFundNotEnoughUnits", err)
		}
	})

	t.Run("from another portfolio's position", func(t *testing.T) {
		_, err := replayBalanceFund([]fundMovementFact{
			opening,
			{TxnID: uuid.New(), EntryID: uuid.New(), Date: planDay(time.August, 5), Withdrawal: true, Amount: mustDecimal(t, "10")},
		}, []fundBalanceFact{balance})
		if !errors.Is(err, ErrFundNotEnoughUnits) {
			t.Fatalf("replayBalanceFund = %v, want ErrFundNotEnoughUnits", err)
		}
	})
}

// A balance on a day nothing was held values nothing: it is skipped and the
// unit value stays where it was.
func TestReplayBalanceFundSkipsABalanceWithNoUnits(t *testing.T) {
	entry := uuid.New()

	replay, err := replayBalanceFund(
		[]fundMovementFact{{TxnID: uuid.New(), EntryID: entry, Date: planDay(time.July, 10), Amount: mustDecimal(t, "500")}},
		[]fundBalanceFact{
			{Date: planDay(time.July, 5), Balance: mustDecimal(t, "100")},
			{Date: planDay(time.July, 31), Balance: mustDecimal(t, "510")},
		},
	)
	if err != nil {
		t.Fatalf("replayBalanceFund: %v", err)
	}

	if !replay.Marks[0].Skipped {
		t.Error("the balance before any contribution was used")
	}

	sameAmount(t, "the contribution", replay.Movements[0].Quantity.String(), "5")
	sameAmount(t, "unit value", replay.Marks[1].UnitValue.String(), "102")
}

// Two portfolios hold the fund; the balance is the whole statement, and it
// values the units of both.
func TestReplayBalanceFundAcrossPortfolios(t *testing.T) {
	a, b := uuid.New(), uuid.New()

	replay, err := replayBalanceFund(
		[]fundMovementFact{
			{TxnID: uuid.New(), EntryID: a, Date: planDay(time.July, 1), Amount: mustDecimal(t, "600")},
			{TxnID: uuid.New(), EntryID: b, Date: planDay(time.July, 1), Amount: mustDecimal(t, "400")},
		},
		[]fundBalanceFact{{Date: planDay(time.July, 31), Balance: mustDecimal(t, "1050")}},
	)
	if err != nil {
		t.Fatalf("replayBalanceFund: %v", err)
	}

	sameAmount(t, "unit value", replay.Marks[0].UnitValue.String(), "105")
}
