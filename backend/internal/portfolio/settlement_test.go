package portfolio

import (
	"errors"
	"strings"
	"testing"
	"time"

	"uuid"

	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"
)

func settlementDecimal(t *testing.T, raw string) decimal.Decimal {
	t.Helper()

	d, err := decimal.NewFromString(raw)
	if err != nil {
		t.Fatalf("decimal %q: %v", raw, err)
	}

	return d
}

func settlementMoney(t *testing.T, raw string, currency money.Currency) money.Money {
	t.Helper()

	m, err := money.NewMoneyFromString(raw, currency)
	if err != nil {
		t.Fatalf("money %q: %v", raw, err)
	}

	return m
}

// settlementTxn is a trade the way the repository reads it back.
func settlementTxn(t *testing.T, txnType TransactionType, currency, feesCurrency money.Currency, fees string) Transaction {
	t.Helper()

	return Transaction{
		ID:              uuid.New(),
		Type:            txnType,
		Quantity:        settlementDecimal(t, "1.7986"),
		Price:           settlementMoney(t, "14.323", currency),
		Currency:        currency,
		FXRate:          decimal.One,
		Fees:            settlementMoney(t, fees, feesCurrency),
		FeesCurrency:    feesCurrency,
		TransactionDate: time.Date(2026, time.August, 18, 0, 0, 0, 0, time.UTC),
	}
}

// IBCZ.DE: a EUR ETF recorded as costing in EUR, bought from a USD account.
func TestPlanSettlementRestatesAEuroPositionInDollars(t *testing.T) {
	txn := settlementTxn(t, Buy, money.EUR, money.EUR, "0")

	plan, err := PlanSettlement([]Transaction{txn}, money.USD, map[uuid.UUID]decimal.Decimal{
		txn.ID: settlementDecimal(t, "1.1698"),
	})
	if err != nil {
		t.Fatalf("PlanSettlement: %v", err)
	}

	if len(plan) != 1 || plan[0].TransactionID != txn.ID {
		t.Fatalf("plan = %+v, want the one transaction", plan)
	}
	if plan[0].FXRate.String() != "1.1698" {
		t.Errorf("rate = %s, want 1.1698", plan[0].FXRate)
	}
	// The fee sat on the fill, in the trade's currency, and stays there.
	if plan[0].FeesCurrency != money.EUR {
		t.Errorf("fees currency = %s, want EUR", plan[0].FeesCurrency)
	}
}

// A currency does not convert into itself at anything but 1, so a rate sent for
// a transaction already quoted in the new cost currency is not taken.
func TestPlanSettlementSettlesTheNewCostCurrencyItselfAtOne(t *testing.T) {
	txn := settlementTxn(t, Buy, money.USD, money.USD, "0")

	plan, err := PlanSettlement([]Transaction{txn}, money.USD, map[uuid.UUID]decimal.Decimal{
		txn.ID: settlementDecimal(t, "1.1698"),
	})
	if err != nil {
		t.Fatalf("PlanSettlement: %v", err)
	}
	if !plan[0].FXRate.Equal(decimal.One) {
		t.Errorf("rate = %s, want 1", plan[0].FXRate)
	}
}

// Without the rate of the day the only number left would be today's, which is
// the re-translation the rate column exists to stop.
func TestPlanSettlementRefusesAMissingRate(t *testing.T) {
	txn := settlementTxn(t, Buy, money.EUR, money.EUR, "0")

	_, err := PlanSettlement([]Transaction{txn}, money.USD, nil)
	if !errors.Is(err, ErrTransactionFXRate) {
		t.Fatalf("err = %v, want ErrTransactionFXRate", err)
	}
	if !strings.Contains(err.Error(), "2026-08-18") {
		t.Errorf("err = %q, want it to name the transaction's date", err)
	}
}

func TestPlanSettlementLetsASplitThroughWithoutARate(t *testing.T) {
	txn := settlementTxn(t, Split, money.EUR, money.EUR, "0")

	plan, err := PlanSettlement([]Transaction{txn}, money.USD, nil)
	if err != nil {
		t.Fatalf("PlanSettlement: %v", err)
	}
	if !plan[0].FXRate.Equal(decimal.One) {
		t.Errorf("rate = %s, want 1", plan[0].FXRate)
	}
}

// A commission billed to the account was in the account's currency, and the
// account's currency is what is being corrected.
func TestPlanSettlementMovesAnAccountFeeWithTheAccount(t *testing.T) {
	txn := settlementTxn(t, Buy, money.USD, money.EUR, "1.5")
	txn.FXRate = settlementDecimal(t, "0.9")

	plan, err := PlanSettlement([]Transaction{txn}, money.USD, nil)
	if err != nil {
		t.Fatalf("PlanSettlement: %v", err)
	}
	if plan[0].FeesCurrency != money.USD {
		t.Errorf("fees currency = %s, want USD", plan[0].FeesCurrency)
	}
}

func TestPlanSettlementRefusesARateForAnotherPosition(t *testing.T) {
	txn := settlementTxn(t, Buy, money.EUR, money.EUR, "0")

	_, err := PlanSettlement([]Transaction{txn}, money.USD, map[uuid.UUID]decimal.Decimal{
		txn.ID:     settlementDecimal(t, "1.1698"),
		uuid.New(): settlementDecimal(t, "1.1698"),
	})
	if !errors.Is(err, ErrTransactionNotFound) {
		t.Fatalf("err = %v, want ErrTransactionNotFound", err)
	}
}

func TestPlanSettlementRefusesAPositionWithoutTransactions(t *testing.T) {
	if _, err := PlanSettlement(nil, money.USD, nil); !errors.Is(err, ErrSettlementWithoutTransactions) {
		t.Fatalf("err = %v, want ErrSettlementWithoutTransactions", err)
	}
}
