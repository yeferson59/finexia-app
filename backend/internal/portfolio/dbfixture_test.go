package portfolio

import (
	"context"
	"testing"

	"uuid"

	"github.com/jackc/pgx/v5/pgxpool"
)

// dropFixture registers the teardown every *_db_test.go fixture in this package
// needs, in the one order that works.
//
// Deleting the user is not enough, and never was. It cascades to
// investment_sources, and portfolio_entries.source_id points at that table with
// ON DELETE SET NULL against a NOT NULL column — so Postgres refuses the whole
// delete with 23502 and nothing goes. Every fixture here swallowed that error
// with `_, _ =`, so the suite stayed green while each run left its user, its
// portfolios, its entries and its assets behind. A development database that
// had run these tests for a while held hundreds of catalog rows nobody planted
// on purpose, which is a slow way to make the assets table stop resembling the
// thing the tests are written against.
//
// Dropping the entries first is what unblocks the cascade, and it is also what
// releases the assets: nothing cascades to those, because they are shared
// catalog rows and not the user's to delete. The three steps live in one
// t.Cleanup rather than three because the order is load-bearing and t.Cleanup
// runs last-in-first-out, which reads backwards at every call site.
//
// It returns the function a fixture calls to hand it each catalog row it
// plants. That is why the call goes at the top of a fixture, before the inserts
// that can t.Fatalf: the teardown is already registered by then, and the ids
// reach it afterwards. A fixture that plants no assets ignores the return.
//
// Failures are logged rather than ignored. Silence is what let this run for as
// long as it did, and a teardown that fails must not fail a test that passed.
func dropFixture(t *testing.T, pool *pgxpool.Pool, userID uuid.UUID) func(assetID uuid.UUID) {
	t.Helper()

	var planted []uuid.UUID

	t.Cleanup(func() {
		ctx := context.Background()

		if _, err := pool.Exec(ctx, `
			DELETE FROM portfolio_entries
			WHERE portfolio_id IN (SELECT id FROM portfolios WHERE user_id = $1)
		`, userID); err != nil {
			t.Logf("cleanup entries: %v", err)
		}

		if _, err := pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, userID); err != nil {
			t.Logf("cleanup user: %v", err)
		}

		if len(planted) == 0 {
			return
		}

		if _, err := pool.Exec(ctx, `DELETE FROM assets WHERE id = ANY($1)`, planted); err != nil {
			t.Logf("cleanup assets: %v", err)
		}
	})

	return func(assetID uuid.UUID) { planted = append(planted, assetID) }
}
