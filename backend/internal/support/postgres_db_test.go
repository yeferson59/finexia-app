package support

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// The transition table and the idempotency of the webhook are both enforced in
// SQL — a conditional UPDATE and a primary key inside one transaction — so the
// in-memory repository the service tests use can only mirror them. These pin
// the real thing.
//
// It needs a database with the migrations applied:
//
//	TEST_DATABASE_URL=postgres://postgres:password@localhost:5432/postgres?sslmode=disable go test ./internal/support/
//
// Without that variable it skips, so `go test ./...` stays a no-setup command.

func supportTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL no está definida: se omite la prueba contra Postgres")
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatalf("pgxpool.New: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Fatalf("ping: %v", err)
	}

	t.Cleanup(pool.Close)

	return pool
}

// plantOrder inserts an order and registers its removal, events included.
func plantOrder(t *testing.T, repo *PostgresRepository, pool *pgxpool.Pool) string {
	t.Helper()

	orderID := newOrderID(time.Now())

	t.Cleanup(func() {
		ctx := context.Background()
		if _, err := pool.Exec(ctx, `DELETE FROM support_payment_events WHERE order_id = $1`, orderID); err != nil {
			t.Logf("cleanup events: %v", err)
		}
		if _, err := pool.Exec(ctx, `DELETE FROM support_contributions WHERE order_id = $1`, orderID); err != nil {
			t.Logf("cleanup order: %v", err)
		}
	})

	if err := repo.CreateContribution(context.Background(), orderID, 20_000, "COP"); err != nil {
		t.Fatalf("CreateContribution: %v", err)
	}

	return orderID
}

func TestPostgresApplyUpdate(t *testing.T) {
	pool := supportTestPool(t)
	repo := NewPostgresRepository(pool)
	ctx := context.Background()

	t.Run("a new order is created and readable", func(t *testing.T) {
		orderID := plantOrder(t, repo, pool)

		c, err := repo.GetContribution(ctx, orderID)
		if err != nil {
			t.Fatalf("GetContribution: %v", err)
		}
		if c.Status != StatusCreated || c.Amount != 20_000 || c.Currency != "COP" || c.TotalCharged != nil || c.ApprovedAt != nil {
			t.Errorf("contribution = %+v", c)
		}
	})

	t.Run("approval records what Bold charged and when", func(t *testing.T) {
		orderID := plantOrder(t, repo, pool)

		c, err := repo.ApplyUpdate(ctx, orderID, Update{Status: StatusApproved, TotalCharged: new(int64(20_000)), PaymentID: "PAY1", PaymentMethod: "CARD"})
		if err != nil {
			t.Fatalf("ApplyUpdate: %v", err)
		}
		if c.Status != StatusApproved || c.TotalCharged == nil || *c.TotalCharged != 20_000 || c.ApprovedAt == nil {
			t.Errorf("contribution = %+v", c)
		}
		if c.PaymentID != "PAY1" || c.PaymentMethod != "CARD" {
			t.Errorf("payment = %q / %q", c.PaymentID, c.PaymentMethod)
		}
	})

	t.Run("a late rejection does not undo an approval", func(t *testing.T) {
		orderID := plantOrder(t, repo, pool)

		if _, err := repo.ApplyUpdate(ctx, orderID, Update{Status: StatusApproved, PaymentID: "PAY2"}); err != nil {
			t.Fatal(err)
		}
		c, err := repo.ApplyUpdate(ctx, orderID, Update{Status: StatusRejected, PaymentID: "PAY1"})
		if err != nil {
			t.Fatalf("ApplyUpdate: %v", err)
		}
		if c.Status != StatusApproved || c.PaymentID != "PAY2" {
			t.Errorf("contribution = %+v, want approved with PAY2 untouched", c)
		}
	})

	t.Run("a retry after a rejection is approved", func(t *testing.T) {
		orderID := plantOrder(t, repo, pool)

		if _, err := repo.ApplyUpdate(ctx, orderID, Update{Status: StatusRejected}); err != nil {
			t.Fatal(err)
		}
		c, err := repo.ApplyUpdate(ctx, orderID, Update{Status: StatusApproved})
		if err != nil || c.Status != StatusApproved {
			t.Fatalf("contribution = %+v, err = %v", c, err)
		}
	})

	t.Run("an unknown order is not found", func(t *testing.T) {
		if _, err := repo.ApplyUpdate(ctx, newOrderID(time.Now()), Update{Status: StatusApproved}); err != ErrContributionNotFound {
			t.Fatalf("err = %v, want ErrContributionNotFound", err)
		}
	})
}

func TestPostgresRecordEvent(t *testing.T) {
	pool := supportTestPool(t)
	repo := NewPostgresRepository(pool)
	ctx := context.Background()

	t.Run("records once and applies once, even delivered concurrently", func(t *testing.T) {
		orderID := plantOrder(t, repo, pool)
		ev := PaymentEvent{ID: "test-" + orderID, Type: "SALE_APPROVED", PaymentID: "PAY1", OrderID: orderID, Total: new(int64(20_000))}
		u := Update{Status: StatusApproved, TotalCharged: ev.Total, PaymentID: ev.PaymentID}

		var (
			wg         sync.WaitGroup
			mu         sync.Mutex
			duplicates int
		)
		for range 5 {
			wg.Go(func() {
				dup, err := repo.RecordEvent(ctx, ev, &u)
				if err != nil {
					t.Errorf("RecordEvent: %v", err)
					return
				}
				if dup {
					mu.Lock()
					duplicates++
					mu.Unlock()
				}
			})
		}
		wg.Wait()

		if duplicates != 4 {
			t.Errorf("duplicates = %d, want 4 of 5 deliveries", duplicates)
		}

		var events int
		if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM support_payment_events WHERE event_id = $1`, ev.ID).Scan(&events); err != nil {
			t.Fatal(err)
		}
		if events != 1 {
			t.Errorf("event rows = %d, want 1", events)
		}

		if c, _ := repo.GetContribution(ctx, orderID); c.Status != StatusApproved {
			t.Errorf("Status = %s, want approved", c.Status)
		}
	})

	t.Run("a payment that is not ours is recorded without an order", func(t *testing.T) {
		eventID := "test-foreign-" + newOrderID(time.Now())
		t.Cleanup(func() {
			_, _ = pool.Exec(context.Background(), `DELETE FROM support_payment_events WHERE event_id = $1`, eventID)
		})

		dup, err := repo.RecordEvent(ctx, PaymentEvent{ID: eventID, Type: "SALE_APPROVED", PaymentID: "LNK"}, &Update{Status: StatusApproved})
		if err != nil || dup {
			t.Fatalf("dup = %v, err = %v", dup, err)
		}
	})
}

// plantAt inserts an order with a given status and age, for the admin and
// reconcile queries, and registers its removal.
func plantAt(t *testing.T, pool *pgxpool.Pool, status Status, age time.Duration, amount int64) string {
	t.Helper()

	orderID := newOrderID(time.Now())
	t.Cleanup(func() {
		if _, err := pool.Exec(context.Background(), `DELETE FROM support_contributions WHERE order_id = $1`, orderID); err != nil {
			t.Logf("cleanup order: %v", err)
		}
	})

	createdAt := time.Now().Add(-age)
	var approvedAt *time.Time
	if status == StatusApproved {
		approvedAt = &createdAt
	}

	if _, err := pool.Exec(context.Background(),
		`INSERT INTO support_contributions (order_id, amount, status, approved_at, created_at, updated_at)
		 VALUES ($1, $2, $3::text::support_contribution_status, $4, $5, $5)`,
		orderID, amount, string(status), approvedAt, createdAt,
	); err != nil {
		t.Fatalf("plant: %v", err)
	}

	return orderID
}

func TestPostgresAdminQueries(t *testing.T) {
	pool := supportTestPool(t)
	repo := NewPostgresRepository(pool)
	ctx := context.Background()

	t.Run("the status filter only returns that status", func(t *testing.T) {
		approved := plantAt(t, pool, StatusApproved, time.Minute, 20_000)
		rejected := plantAt(t, pool, StatusRejected, time.Minute, 20_000)

		items, count, err := repo.ListContributions(ctx, StatusRejected, 0, 1000)
		if err != nil {
			t.Fatalf("ListContributions: %v", err)
		}
		if count != uint(len(items)) && len(items) < 1000 {
			t.Errorf("count = %d for %d items", count, len(items))
		}

		var sawRejected bool
		for _, c := range items {
			if c.Status != StatusRejected {
				t.Errorf("%s is %s under the rejected filter", c.OrderID, c.Status)
			}
			if c.OrderID == approved {
				t.Error("the approved order came back under the rejected filter")
			}
			sawRejected = sawRejected || c.OrderID == rejected
		}
		if !sawRejected {
			t.Error("the rejected order is missing")
		}

		all, allCount, err := repo.ListContributions(ctx, "", 0, 1000)
		if err != nil || allCount < 2 || len(all) < 2 {
			t.Fatalf("unfiltered: %d of %d, err = %v", len(all), allCount, err)
		}
	})

	t.Run("the summary adds approved orders, recent ones apart", func(t *testing.T) {
		since := time.Now().Add(-RecentWindow)
		before, err := repo.Summarize(ctx, since)
		if err != nil {
			t.Fatalf("Summarize: %v", err)
		}

		plantAt(t, pool, StatusApproved, time.Hour, 20_000)
		plantAt(t, pool, StatusApproved, 40*24*time.Hour, 50_000)
		plantAt(t, pool, StatusRejected, time.Hour, 99_000)

		after, err := repo.Summarize(ctx, since)
		if err != nil {
			t.Fatalf("Summarize: %v", err)
		}

		if got := after.ApprovedTotal - before.ApprovedTotal; got != 70_000 {
			t.Errorf("approved total grew by %d, want 70000", got)
		}
		if got := after.ApprovedRecent - before.ApprovedRecent; got != 20_000 {
			t.Errorf("recent total grew by %d, want 20000", got)
		}
		if got := after.Counts[StatusApproved] - before.Counts[StatusApproved]; got != 2 {
			t.Errorf("approved count grew by %d, want 2", got)
		}
		if _, ok := after.Counts[StatusVoided]; !ok {
			t.Error("a status with no orders is missing from the counts")
		}
	})

	t.Run("open orders and abandoned ones", func(t *testing.T) {
		now := time.Now()
		fresh := plantAt(t, pool, StatusCreated, 5*time.Minute, 20_000)
		open := plantAt(t, pool, StatusCreated, 2*time.Hour, 20_000)
		pending := plantAt(t, pool, StatusPending, 30*24*time.Hour, 20_000)
		abandoned := plantAt(t, pool, StatusCreated, 8*24*time.Hour, 20_000)
		settled := plantAt(t, pool, StatusApproved, 8*24*time.Hour, 20_000)

		ids, err := repo.ListOpen(ctx, now.Add(-reconcileGrace), now.Add(-abandonAfter), 10_000)
		if err != nil {
			t.Fatalf("ListOpen: %v", err)
		}
		got := map[string]bool{}
		for _, id := range ids {
			got[id] = true
		}
		for id, want := range map[string]bool{fresh: false, open: true, pending: true, abandoned: false, settled: false} {
			if got[id] != want {
				t.Errorf("ListOpen has %s = %v, want %v", id, got[id], want)
			}
		}

		if _, err := repo.DeleteAbandoned(ctx, now.Add(-abandonAfter)); err != nil {
			t.Fatalf("DeleteAbandoned: %v", err)
		}
		if _, err := repo.GetContribution(ctx, abandoned); err != ErrContributionNotFound {
			t.Errorf("abandoned: err = %v, want deleted", err)
		}
		for _, kept := range []string{fresh, open, pending, settled} {
			if _, err := repo.GetContribution(ctx, kept); err != nil {
				t.Errorf("%s was deleted: %v", kept, err)
			}
		}
	})
}
