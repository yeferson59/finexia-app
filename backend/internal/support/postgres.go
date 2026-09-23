package support

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresRepository is the pgx-backed implementation of Repository.
type PostgresRepository struct {
	db *pgxpool.Pool
}

var _ Repository = (*PostgresRepository)(nil)

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return new(PostgresRepository{db})
}

// querier is what the pool and a transaction have in common, so the update
// runs the same inside RecordEvent's transaction and outside it.
type querier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

const contributionColumns = `order_id, amount, currency, status, total_charged,
	COALESCE(payment_id, ''), COALESCE(payment_method, ''), approved_at, created_at, updated_at`

const selectContribution = `SELECT ` + contributionColumns + ` FROM support_contributions WHERE order_id = $1`

// scanContribution reads contributionColumns, in that order.
func scanContribution(row pgx.Row) (Contribution, error) {
	var c Contribution

	err := row.Scan(
		&c.OrderID, &c.Amount, &c.Currency, &c.Status, &c.TotalCharged,
		&c.PaymentID, &c.PaymentMethod, &c.ApprovedAt, &c.CreatedAt, &c.UpdatedAt,
	)

	return c, err
}

func (r *PostgresRepository) CreateContribution(ctx context.Context, orderID string, amount int64, currency string) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO support_contributions (order_id, amount, currency) VALUES ($1, $2, $3)`,
		orderID, amount, currency,
	)

	return err
}

func (r *PostgresRepository) GetContribution(ctx context.Context, orderID string) (Contribution, error) {
	return getContribution(ctx, r.db, orderID)
}

func (r *PostgresRepository) ApplyUpdate(ctx context.Context, orderID string, u Update) (Contribution, error) {
	return applyUpdate(ctx, r.db, orderID, u)
}

func (r *PostgresRepository) RecordEvent(ctx context.Context, ev PaymentEvent, u *Update) (bool, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var orderID *string
	if ev.OrderID != "" {
		orderID = &ev.OrderID
	}

	tag, err := tx.Exec(ctx,
		`INSERT INTO support_payment_events (event_id, type, payment_id, order_id, total)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (event_id) DO NOTHING`,
		ev.ID, ev.Type, ev.PaymentID, orderID, ev.Total,
	)
	if err != nil {
		return false, err
	}

	if tag.RowsAffected() == 0 {
		return true, nil
	}

	// An order this table does not know — not minted here, or minted against
	// another database — is still recorded above and still acknowledged.
	if u != nil && ev.OrderID != "" {
		if _, err := applyUpdate(ctx, tx, ev.OrderID, *u); err != nil && !errors.Is(err, ErrContributionNotFound) {
			return false, err
		}
	}

	return false, tx.Commit(ctx)
}

func getContribution(ctx context.Context, q querier, orderID string) (Contribution, error) {
	c, err := scanContribution(q.QueryRow(ctx, selectContribution, orderID))
	if errors.Is(err, pgx.ErrNoRows) {
		return Contribution{}, ErrContributionNotFound
	}

	return c, err
}

// applyUpdate is a single conditional UPDATE: the transition table becomes the
// WHERE clause, so two deliveries racing on the same order cannot both read an
// old status and both write.
func applyUpdate(ctx context.Context, q querier, orderID string, u Update) (Contribution, error) {
	from := transitionsTo(u.Status)
	allowed := make([]string, len(from))
	for i, s := range from {
		allowed[i] = string(s)
	}

	_, err := q.Exec(ctx,
		`UPDATE support_contributions SET
			status         = $2::text::support_contribution_status,
			total_charged  = COALESCE($3, total_charged),
			payment_id     = COALESCE(NULLIF($4, ''), payment_id),
			payment_method = COALESCE(NULLIF($5, ''), payment_method),
			approved_at    = CASE WHEN $2 = 'approved' THEN COALESCE(approved_at, NOW()) ELSE approved_at END,
			updated_at     = NOW()
		 WHERE order_id = $1 AND status::text = ANY($6::text[])`,
		orderID, string(u.Status), u.TotalCharged, u.PaymentID, u.PaymentMethod, allowed,
	)
	if err != nil {
		return Contribution{}, err
	}

	return getContribution(ctx, q, orderID)
}

// ListContributions takes an empty status as "any", inside the WHERE clause,
// so one statement serves the listing with and without a filter.
func (r *PostgresRepository) ListContributions(ctx context.Context, status Status, offset, limit uint) ([]Contribution, uint, error) {
	const filter = `WHERE $1 = '' OR status::text = $1`

	var count uint
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM support_contributions `+filter, string(status)).Scan(&count); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(ctx,
		`SELECT `+contributionColumns+` FROM support_contributions `+filter+`
		 ORDER BY created_at DESC, order_id
		 LIMIT $2 OFFSET $3`,
		string(status), limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]Contribution, 0, limit)
	for rows.Next() {
		c, err := scanContribution(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, c)
	}

	return items, count, rows.Err()
}

func (r *PostgresRepository) Summarize(ctx context.Context, since time.Time) (Summary, error) {
	summary := Summary{Counts: make(map[Status]int64, len(Statuses))}
	for _, s := range Statuses {
		summary.Counts[s] = 0
	}

	rows, err := r.db.Query(ctx, `SELECT status::text, COUNT(*) FROM support_contributions GROUP BY status`)
	if err != nil {
		return Summary{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			status string
			count  int64
		)
		if err := rows.Scan(&status, &count); err != nil {
			return Summary{}, err
		}
		summary.Counts[Status(status)] = count
	}
	if err := rows.Err(); err != nil {
		return Summary{}, err
	}

	err = r.db.QueryRow(ctx,
		`SELECT
			COALESCE(SUM(COALESCE(total_charged, amount)), 0),
			COALESCE(SUM(COALESCE(total_charged, amount)) FILTER (WHERE approved_at >= $1), 0),
			MAX(approved_at)
		 FROM support_contributions WHERE status = 'approved'`,
		since,
	).Scan(&summary.ApprovedTotal, &summary.ApprovedRecent, &summary.LastApprovedAt)

	return summary, err
}

// ListOpen goes newest first: a recent order is the one most likely to have
// been paid, and the batch limit leaves the oldest for the next run.
func (r *PostgresRepository) ListOpen(ctx context.Context, createdBefore, createdAfter time.Time, limit int) ([]string, error) {
	rows, err := r.db.Query(ctx,
		`SELECT order_id FROM support_contributions
		 WHERE created_at < $1
		   AND (status = 'pending' OR (status = 'created' AND created_at >= $2))
		 ORDER BY created_at DESC
		 LIMIT $3`,
		createdBefore, createdAfter, limit,
	)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, pgx.RowTo[string])
}

func (r *PostgresRepository) DeleteAbandoned(ctx context.Context, before time.Time) (int64, error) {
	tag, err := r.db.Exec(ctx,
		`DELETE FROM support_contributions WHERE status = 'created' AND created_at < $1`,
		before,
	)
	if err != nil {
		return 0, err
	}

	return tag.RowsAffected(), nil
}
