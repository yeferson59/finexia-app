package portfolio

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"uuid"

	"github.com/yeferson59/gofinance/v2/decimal"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

// The synthetic units of a fund followed by balance (D4 and D5 of
// docs/PLAN_FONDOS_INVERSION.md). The owner only knows money: what went in,
// what came out, and what the app says the fund holds on a day. Units are how
// that money is split into what the owner put in and what the fund made of it:
//
//   - the first contribution buys at a unit value of 100 — an index, base 100;
//   - a contribution or a withdrawal on day d trades at the unit value of the
//     last balance before d: units = amount / unit value;
//   - a balance S on day d — the close of the day, with its movements — fixes
//     the unit value at S / units held that day.
//
// The return that comes out is time-weighted: a large contribution is not
// mistaken for growth, because it buys units instead of moving their value.
//
// The replay is pure so it can be checked figure by figure against the plan's
// example. The repository reads the facts, replays them, and writes back the
// quantities, the prices and the unit values it derived.

var (
	// ErrFundNotEnoughUnits refuses a withdrawal larger than what its position
	// holds on its day — also one that became larger because an earlier
	// movement was edited. It is a conflict: the same withdrawal fits a position
	// that holds more.
	ErrFundNotEnoughUnits = httpx.AsConflict(errors.New("the position does not hold that much of the fund"))
	// ErrFundNoUnits refuses a balance on a day the fund held nothing: there
	// are no units for it to value.
	ErrFundNoUnits = httpx.AsConflict(errors.New("the fund held no units on that day"))
	// ErrFundBalanceManaged refuses the generic transaction writers on a fund
	// followed by balance: its quantities are derived from its movements, and a
	// transaction written by hand would stop matching them (D12).
	ErrFundBalanceManaged = httpx.AsConflict(errors.New("the fund is followed by balance; record its contributions and withdrawals from the funds screen"))
)

// fundOpeningUnitValue is what the first contribution buys at: an index, base
// 100. Not 1: unit values are kept at eight decimals, and the rounding of a unit
// value lands on every unit. At 1 a fund of ten million has ten million units
// and a balance read back from them is off by several cents; at 100 it has a
// hundredth of that, and the error stays under a cent.
var fundOpeningUnitValue = decimal.MustFromString("100")

// fundWithdrawalSlack is how far past a position's units a withdrawal may reach
// and still take exactly all of them: a cent of the fund's currency. A
// withdrawal of the whole balance, typed as the balance, lands a rounding away
// from the units it converts to, and that is not a withdrawal of more.
var fundWithdrawalSlack = decimal.MustFromString("0.01")

// fundMovementFact is a contribution or a withdrawal as the owner stated it.
type fundMovementFact struct {
	TxnID     uuid.UUID
	EntryID   uuid.UUID
	Date      time.Time
	CreatedAt time.Time
	// Withdrawal is a sale of units; otherwise it is a purchase.
	Withdrawal bool
	// Amount is the money that went in, or came out before the fees.
	Amount decimal.Decimal
	// All takes every unit the position holds.
	All bool
}

// fundBalanceFact is the balance the owner read on a day.
type fundBalanceFact struct {
	Date    time.Time
	Balance decimal.Decimal
}

// replayedMovement is a movement with the units it came to.
type replayedMovement struct {
	TxnID    uuid.UUID
	EntryID  uuid.UUID
	Quantity decimal.Decimal
	Price    decimal.Decimal
}

// replayedMark is a balance with the unit value it fixed. Skipped is a balance
// on a day the fund held no units: it values nothing and leaves the unit value
// where it was.
type replayedMark struct {
	Date      time.Time
	UnitValue decimal.Decimal
	Skipped   bool
}

// fundReplay is everything a replay derived.
type fundReplay struct {
	Movements []replayedMovement
	Marks     []replayedMark
}

// replayBalanceFund derives the units of every movement and the unit value of
// every balance, in the order they happened: by day, and within a day the
// movements before the balance, which is the close of the day.
func replayBalanceFund(movements []fundMovementFact, balances []fundBalanceFact) (fundReplay, error) {
	moves := append([]fundMovementFact(nil), movements...)
	sort.SliceStable(moves, func(i, j int) bool {
		a, b := moves[i], moves[j]
		if !a.Date.Equal(b.Date) {
			return a.Date.Before(b.Date)
		}

		if !a.CreatedAt.Equal(b.CreatedAt) {
			return a.CreatedAt.Before(b.CreatedAt)
		}

		return a.TxnID.String() < b.TxnID.String()
	})

	marks := append([]fundBalanceFact(nil), balances...)
	sort.SliceStable(marks, func(i, j int) bool { return marks[i].Date.Before(marks[j].Date) })

	var (
		out      fundReplay
		value    = fundOpeningUnitValue
		held     = make(map[uuid.UUID]decimal.Decimal)
		total    decimal.Decimal
		nextMark int
	)

	// closeDaysBefore applies every balance dated before day: those days closed
	// before anything on day moved.
	closeDaysBefore := func(day time.Time, inclusive bool) error {
		for nextMark < len(marks) {
			m := marks[nextMark]
			if m.Date.After(day) || (!inclusive && m.Date.Equal(day)) {
				return nil
			}

			nextMark++

			if !total.IsPos() {
				out.Marks = append(out.Marks, replayedMark{Date: m.Date, Skipped: true})

				continue
			}

			v, err := m.Balance.Div(total)
			if err != nil {
				return err
			}

			v = v.RoundHAZ(8)
			if !v.IsPos() {
				return fmt.Errorf("%w: a balance of %s on %s is too small to value %s units",
					ErrInvalidFundMark, m.Balance, m.Date.Format(time.DateOnly), total)
			}

			value = v
			out.Marks = append(out.Marks, replayedMark{Date: m.Date, UnitValue: v})
		}

		return nil
	}

	for _, mv := range moves {
		if err := closeDaysBefore(mv.Date, false); err != nil {
			return fundReplay{}, err
		}

		quantity, price, err := replayMovement(mv, value, held[mv.EntryID])
		if err != nil {
			return fundReplay{}, err
		}

		if mv.Withdrawal {
			held[mv.EntryID] = held[mv.EntryID].Sub(quantity)
			total = total.Sub(quantity)
		} else {
			held[mv.EntryID] = held[mv.EntryID].Add(quantity)
			total = total.Add(quantity)
		}

		out.Movements = append(out.Movements, replayedMovement{
			TxnID: mv.TxnID, EntryID: mv.EntryID, Quantity: quantity, Price: price,
		})
	}

	// The balances after the last movement close their days too.
	if len(marks) > 0 {
		if err := closeDaysBefore(marks[len(marks)-1].Date, true); err != nil {
			return fundReplay{}, err
		}
	}

	return out, nil
}

// replayMovement is the quantity and price of one movement at the unit value
// of its day, against what its position holds.
func replayMovement(mv fundMovementFact, value, held decimal.Decimal) (quantity, price decimal.Decimal, err error) {
	day := mv.Date.Format(time.DateOnly)

	if !mv.Withdrawal {
		q, err := mv.Amount.Div(value)
		if err != nil {
			return decimal.Decimal{}, decimal.Decimal{}, err
		}

		q = q.RoundHAZ(8)
		if !q.IsPos() {
			return decimal.Decimal{}, decimal.Decimal{}, fmt.Errorf("%w: the contribution of %s is too small", ErrInvalidFund, day)
		}

		return q, value, nil
	}

	if !held.IsPos() {
		return decimal.Decimal{}, decimal.Decimal{}, fmt.Errorf("%w: nothing was left on %s", ErrFundNotEnoughUnits, day)
	}

	if mv.All {
		p, err := mv.Amount.Div(held)
		if err != nil {
			return decimal.Decimal{}, decimal.Decimal{}, err
		}

		return held, p.RoundHAZ(8), nil
	}

	q, err := mv.Amount.Div(value)
	if err != nil {
		return decimal.Decimal{}, decimal.Decimal{}, err
	}

	q = q.RoundHAZ(8)

	if q.GreaterThan(held) {
		// A cent past what is there is the whole of it, typed as money.
		if q.Sub(held).Mul(value).LessThanOrEqual(fundWithdrawalSlack) {
			return held, value, nil
		}

		return decimal.Decimal{}, decimal.Decimal{}, fmt.Errorf("%w: on %s it held %s at %s",
			ErrFundNotEnoughUnits, day, held.Mul(value).RoundHAZ(2), value)
	}

	return q, value, nil
}
