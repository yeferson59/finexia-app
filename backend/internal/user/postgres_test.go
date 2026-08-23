package user

import (
	"context"
	"errors"
	"testing"

	"uuid"

	"github.com/jackc/pgx/v5"
)

// stubRow answers one QueryRow: either an error, or a fill that writes into the
// scan destinations the caller passed.
type stubRow struct {
	err  error
	fill func(dest []any)
}

func (r stubRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	if r.fill != nil {
		r.fill(dest)
	}

	return nil
}

// stubQuerier stands in for a pool or an open transaction: it hands out rows in
// order and records what each statement received, which is how the tests below
// assert that the role id looked up first is the one the INSERT uses.
type stubQuerier struct {
	rows []pgx.Row
	sqls []string
	args [][]any
}

func (q *stubQuerier) QueryRow(_ context.Context, sql string, args ...any) pgx.Row {
	q.sqls = append(q.sqls, sql)
	q.args = append(q.args, args)

	if len(q.rows) == 0 {
		return stubRow{err: errors.New("unexpected query: " + sql)}
	}

	row := q.rows[0]
	q.rows = q.rows[1:]

	return row
}

// insertedUserRow fills the eleven RETURNING columns; only the ones the tests
// assert carry a meaningful value.
func insertedUserRow(id uuid.UUID, name, email string, roleID uuid.UUID) stubRow {
	return stubRow{fill: func(dest []any) {
		*dest[0].(*uuid.UUID) = id
		*dest[1].(*string) = name
		*dest[2].(*string) = email
		*dest[5].(*uuid.UUID) = roleID
	}}
}

// TestInsertUserUsesLookedUpRole is the invariant that lets auth's sign-up run
// this statement inside its own transaction: InsertUser resolves the customer
// role and feeds that id to the INSERT, touching nothing else.
func TestInsertUserUsesLookedUpRole(t *testing.T) {
	roleID := uuid.New()
	userID := uuid.New()

	q := new(stubQuerier{rows: []pgx.Row{
		stubRow{fill: func(dest []any) { *dest[0].(*uuid.UUID) = roleID }},
		insertedUserRow(userID, "Jane Doe", "jane@example.com", roleID),
	}})

	got, err := InsertUser(context.Background(), q, "Jane Doe", "jane@example.com")
	if err != nil {
		t.Fatalf("InsertUser: %v", err)
	}

	if got.ID != userID || got.Name != "Jane Doe" || got.Email != "jane@example.com" {
		t.Errorf("user = %+v, want the inserted row echoed back", got)
	}
	if got.Role.Name != "customer" {
		t.Errorf("Role.Name = %q, want %q", got.Role.Name, "customer")
	}

	if len(q.sqls) != 2 {
		t.Fatalf("ran %d statements, want 2 (role lookup + insert)", len(q.sqls))
	}
	if q.args[0][0] != "customer" {
		t.Errorf("role lookup arg = %v, want %q", q.args[0][0], "customer")
	}
	// The role id must travel from the lookup into the INSERT, not be
	// re-derived or defaulted.
	if q.args[1][2] != roleID {
		t.Errorf("insert role_id = %v, want the looked-up %v", q.args[1][2], roleID)
	}
}

// TestInsertUserStopsOnFailure checks both statements propagate their error
// untagged, and that a failed role lookup never reaches the INSERT — the caller
// owns the transaction, so InsertUser must leave it clean to roll back.
func TestInsertUserStopsOnFailure(t *testing.T) {
	boom := errors.New("boom")

	t.Run("role lookup fails", func(t *testing.T) {
		q := new(stubQuerier{rows: []pgx.Row{stubRow{err: boom}}})

		if _, err := InsertUser(context.Background(), q, "Jane", "jane@example.com"); !errors.Is(err, boom) {
			t.Fatalf("err = %v, want %v", err, boom)
		}
		if len(q.sqls) != 1 {
			t.Errorf("ran %d statements, want 1: the insert must not run", len(q.sqls))
		}
	})

	t.Run("insert fails", func(t *testing.T) {
		q := new(stubQuerier{rows: []pgx.Row{
			stubRow{fill: func(dest []any) { *dest[0].(*uuid.UUID) = uuid.New() }},
			stubRow{err: boom},
		}})

		if _, err := InsertUser(context.Background(), q, "Jane", "jane@example.com"); !errors.Is(err, boom) {
			t.Fatalf("err = %v, want %v", err, boom)
		}
	})
}
