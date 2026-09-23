// Package support owns the voluntary contributions of /apoyar: it signs the
// order Bold's payment button opens, stores it, and follows its status through
// Bold's webhook and the payment-voucher query.
//
// Nothing here is tied to an account. The page is public and contributing
// changes nothing in the product, so a contribution is an order and a status,
// not a user's record.
package support

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"
	"strconv"
	"time"

	"github.com/yeferson59/finexia-app/internal/platform/bold"
)

// Bold charges nothing under $1.000; the ceiling is ours, well below the
// account's own limits, so a typo with three extra zeros cannot get through.
const (
	MinAmount int64 = 1_000
	MaxAmount int64 = 2_000_000
)

type Status string

const (
	StatusCreated  Status = "created"
	StatusPending  Status = "pending"
	StatusRejected Status = "rejected"
	StatusApproved Status = "approved"
	StatusVoided   Status = "voided"
)

// settled reports whether nothing Bold can still say would change the status
// in a way the page cares about, so there is no point asking again.
// A rejection is not settled: the checkout lets the payer retry on the same
// order.
func (s Status) settled() bool {
	return s == StatusApproved || s == StatusVoided
}

// transitionsTo lists the statuses an order may be in for it to move to `to`.
//
// Bold's notifications can arrive late, twice, or out of order, and the voucher
// query can race the webhook. The table is what keeps a late message from
// undoing a newer one:
//
//   - a rejection only lands on an order still waiting, so a stale rejection
//     of a first attempt cannot overwrite the approval of the retry;
//   - an approval lands on anything but a void, including a rejection, which is
//     exactly the retry case;
//   - a void lands on anything that is not already void. Bold voids approved
//     payments, but the approval's own notification may be the one that got
//     lost.
func transitionsTo(to Status) []Status {
	switch to {
	case StatusPending:
		return []Status{StatusCreated, StatusPending, StatusRejected}
	case StatusRejected:
		return []Status{StatusCreated, StatusPending}
	case StatusApproved:
		return []Status{StatusCreated, StatusPending, StatusRejected, StatusApproved}
	case StatusVoided:
		return []Status{StatusCreated, StatusPending, StatusRejected, StatusApproved}
	default:
		return nil
	}
}

// statusFromVoucher reads the voucher query. NO_TRANSACTION_FOUND, and anything
// unknown, leaves the order as it is.
func statusFromVoucher(status string) (Status, bool) {
	switch status {
	case bold.StatusApproved:
		return StatusApproved, true
	case bold.StatusProcessing, bold.StatusPending:
		return StatusPending, true
	case bold.StatusRejected, bold.StatusFailed:
		return StatusRejected, true
	case bold.StatusVoided:
		return StatusVoided, true
	default:
		return "", false
	}
}

// statusFromEvent reads a webhook. A rejected void changes nothing: the payment
// stays what it was.
func statusFromEvent(eventType string) (Status, bool) {
	switch eventType {
	case bold.EventSaleApproved:
		return StatusApproved, true
	case bold.EventSaleRejected:
		return StatusRejected, true
	case bold.EventVoidApproved:
		return StatusVoided, true
	default:
		return "", false
	}
}

// Contribution is one order as the app stores it.
type Contribution struct {
	OrderID  string `json:"orderId"`
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
	Status   Status `json:"status"`
	// TotalCharged is what Bold reports it charged, once there is a transaction.
	TotalCharged  *int64     `json:"totalCharged"`
	PaymentID     string     `json:"-"`
	PaymentMethod string     `json:"-"`
	ApprovedAt    *time.Time `json:"approvedAt"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"-"`
}

// Statuses lists every status, in the order an order moves through them.
var Statuses = []Status{StatusCreated, StatusPending, StatusRejected, StatusApproved, StatusVoided}

// parseStatus reads a status from a query string; "" means any.
func parseStatus(raw string) (Status, bool) {
	if raw == "" {
		return "", true
	}

	for _, s := range Statuses {
		if string(s) == raw {
			return s, true
		}
	}

	return "", false
}

// Summary is the admin's view of the contributions as a whole.
type Summary struct {
	// Counts has every status, at zero when there is none.
	Counts map[Status]int64 `json:"counts"`
	// ApprovedTotal adds up what Bold charged for every approved order — the
	// order's own amount when Bold did not report a total.
	ApprovedTotal int64 `json:"approvedTotal"`
	// ApprovedRecent is the same sum over the last RecentWindow.
	ApprovedRecent int64      `json:"approvedRecent"`
	LastApprovedAt *time.Time `json:"lastApprovedAt"`
}

// RecentWindow is the span Summary.ApprovedRecent covers.
const RecentWindow = 30 * 24 * time.Hour

// Update is a status change with what Bold reported alongside it. Empty fields
// keep what the row already has.
type Update struct {
	Status        Status
	TotalCharged  *int64
	PaymentID     string
	PaymentMethod string
}

// PaymentEvent is a webhook notification as it is recorded.
type PaymentEvent struct {
	ID        string
	Type      string
	PaymentID string
	// OrderID is empty when the payment did not come from one of our orders.
	OrderID string
	Total   *int64
}

// Checkout is what the browser hands to `new BoldCheckout(…)`, field for field.
type Checkout struct {
	OrderID  string `json:"orderId"`
	Currency string `json:"currency"`
	// Amount goes as text: that is how the payment button takes it.
	Amount             string `json:"amount"`
	APIKey             string `json:"apiKey"`
	IntegritySignature string `json:"integritySignature"`
	RedirectionURL     string `json:"redirectionUrl"`
	Description        string `json:"description"`
	// RenderMode "embedded" opens Bold's checkout in a modal over the page.
	RenderMode string `json:"renderMode"`
}

const orderPrefix = "FNX-APOYO-"

// orderIDPattern accepts only order ids this module mints. The id arrives from
// a URL and from webhooks, so it is checked before it reaches a query or Bold.
var orderIDPattern = regexp.MustCompile(`^FNX-APOYO-\d{13}-[0-9a-f]{16}$`)

// newOrderID mints an order id: Bold takes up to 60 characters of
// alphanumerics, `-` and `_`, and recommends a timestamp so ids never repeat.
// The random tail is what makes an id unguessable, since anyone holding one can
// read that order's status.
func newOrderID(now time.Time) string {
	var tail [8]byte
	_, _ = rand.Read(tail[:])

	return orderPrefix + strconv.FormatInt(now.UnixMilli(), 10) + "-" + hex.EncodeToString(tail[:])
}

func isOrderID(value string) bool {
	return orderIDPattern.MatchString(value)
}
