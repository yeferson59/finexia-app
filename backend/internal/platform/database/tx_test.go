package database

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
)

// fakeTx embeds pgx.Tx so only the lifecycle methods need real behavior;
// anything else panics loudly if a test touches it.
type fakeTx struct {
	pgx.Tx

	committed  bool
	rolledBack bool
	commitErr  error
}

func (t *fakeTx) Commit(context.Context) error {
	t.committed = true
	return t.commitErr
}

func (t *fakeTx) Rollback(context.Context) error {
	if !t.committed {
		t.rolledBack = true
	}
	return nil
}

type fakeBeginner struct {
	tx       *fakeTx
	beginErr error
}

func (b *fakeBeginner) Begin(context.Context) (pgx.Tx, error) {
	if b.beginErr != nil {
		return nil, b.beginErr
	}
	return b.tx, nil
}

func TestWithinTx(t *testing.T) {
	ctx := context.Background()

	t.Run("commits when fn succeeds", func(t *testing.T) {
		tx := &fakeTx{}
		err := WithinTx(ctx, &fakeBeginner{tx: tx}, func(context.Context, pgx.Tx) error {
			return nil
		})
		if err != nil {
			t.Fatalf("WithinTx = %v, want nil", err)
		}
		if !tx.committed {
			t.Error("transaction was not committed")
		}
		if tx.rolledBack {
			t.Error("transaction should not be rolled back after commit")
		}
	})

	t.Run("rolls back when fn fails", func(t *testing.T) {
		tx := &fakeTx{}
		wantErr := errors.New("boom")
		err := WithinTx(ctx, &fakeBeginner{tx: tx}, func(context.Context, pgx.Tx) error {
			return wantErr
		})
		if !errors.Is(err, wantErr) {
			t.Fatalf("WithinTx = %v, want %v", err, wantErr)
		}
		if tx.committed {
			t.Error("transaction should not be committed")
		}
		if !tx.rolledBack {
			t.Error("transaction was not rolled back")
		}
	})

	t.Run("rolls back when fn panics", func(t *testing.T) {
		tx := &fakeTx{}
		func() {
			defer func() { _ = recover() }()
			_ = WithinTx(ctx, &fakeBeginner{tx: tx}, func(context.Context, pgx.Tx) error {
				panic("boom")
			})
		}()
		if tx.committed {
			t.Error("transaction should not be committed")
		}
		if !tx.rolledBack {
			t.Error("transaction was not rolled back")
		}
	})

	t.Run("propagates Begin errors", func(t *testing.T) {
		wantErr := errors.New("no connection")
		err := WithinTx(ctx, &fakeBeginner{beginErr: wantErr}, func(context.Context, pgx.Tx) error {
			t.Error("fn should not run when Begin fails")
			return nil
		})
		if !errors.Is(err, wantErr) {
			t.Fatalf("WithinTx = %v, want %v", err, wantErr)
		}
	})

	t.Run("returns the commit error", func(t *testing.T) {
		wantErr := errors.New("serialization failure")
		tx := &fakeTx{commitErr: wantErr}
		err := WithinTx(ctx, &fakeBeginner{tx: tx}, func(context.Context, pgx.Tx) error {
			return nil
		})
		if !errors.Is(err, wantErr) {
			t.Fatalf("WithinTx = %v, want %v", err, wantErr)
		}
	})
}
