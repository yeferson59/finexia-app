package support

import (
	"context"
	"time"
)

// Repository declares only what this module needs from persistence.
type Repository interface {
	CreateContribution(ctx context.Context, orderID string, amount int64, currency string) error
	GetContribution(ctx context.Context, orderID string) (Contribution, error)
	// ApplyUpdate moves an order to u.Status if transitionsTo allows it from
	// where it is, and returns the order as it stands afterwards either way.
	ApplyUpdate(ctx context.Context, orderID string, u Update) (Contribution, error)
	// RecordEvent stores a webhook and, in the same transaction, applies u to
	// its order when there is one. It reports duplicate — and changes nothing —
	// when the event id was already recorded.
	RecordEvent(ctx context.Context, ev PaymentEvent, u *Update) (duplicate bool, err error)

	// ListContributions pages through the orders, newest first, optionally
	// only those in one status.
	ListContributions(ctx context.Context, status Status, offset, limit uint) ([]Contribution, uint, error)
	// Summarize counts the orders by status and adds up what was approved,
	// in total and since `since`.
	Summarize(ctx context.Context, since time.Time) (Summary, error)
	// ListOpen returns up to limit orders worth asking Bold about: pending ones,
	// and created ones younger than createdAfter. Only orders older than
	// createdBefore, so a webhook on its way gets there first.
	ListOpen(ctx context.Context, createdBefore, createdAfter time.Time, limit int) ([]string, error)
	// DeleteAbandoned removes the orders still in created from before `before`:
	// checkouts somebody opened and closed without paying.
	DeleteAbandoned(ctx context.Context, before time.Time) (int64, error)
}
