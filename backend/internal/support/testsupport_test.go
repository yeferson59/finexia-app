package support

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"

	"github.com/yeferson59/finexia-app/internal/platform/bold"
)

// memRepository keeps orders in a map and applies the same transition table the
// SQL does, so the service's flows can be followed end to end without Postgres.
// The SQL itself is pinned by postgres_db_test.go.
type memRepository struct {
	mu     sync.Mutex
	orders map[string]Contribution
	events map[string]PaymentEvent
	// failRecord makes RecordEvent fail, to check the handler's 500.
	failRecord error
	// deletedBefore is the cutoff of the last DeleteAbandoned, zero if none ran.
	deletedBefore time.Time
}

func newMemRepository() *memRepository {
	return new(memRepository{orders: map[string]Contribution{}, events: map[string]PaymentEvent{}})
}

func (m *memRepository) CreateContribution(_ context.Context, orderID string, amount int64, currency string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	m.orders[orderID] = Contribution{OrderID: orderID, Amount: amount, Currency: currency, Status: StatusCreated, CreatedAt: now, UpdatedAt: now}

	return nil
}

func (m *memRepository) GetContribution(_ context.Context, orderID string) (Contribution, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	c, ok := m.orders[orderID]
	if !ok {
		return Contribution{}, ErrContributionNotFound
	}

	return c, nil
}

func (m *memRepository) ApplyUpdate(_ context.Context, orderID string, u Update) (Contribution, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.apply(orderID, u)
}

func (m *memRepository) apply(orderID string, u Update) (Contribution, error) {
	c, ok := m.orders[orderID]
	if !ok {
		return Contribution{}, ErrContributionNotFound
	}

	if slices.Contains(transitionsTo(u.Status), c.Status) {
		c.Status = u.Status
		if u.TotalCharged != nil {
			c.TotalCharged = u.TotalCharged
		}
		if u.PaymentID != "" {
			c.PaymentID = u.PaymentID
		}
		if u.PaymentMethod != "" {
			c.PaymentMethod = u.PaymentMethod
		}
		if u.Status == StatusApproved && c.ApprovedAt == nil {
			c.ApprovedAt = new(time.Now())
		}
		m.orders[orderID] = c
	}

	return c, nil
}

func (m *memRepository) RecordEvent(_ context.Context, ev PaymentEvent, u *Update) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failRecord != nil {
		return false, m.failRecord
	}

	if _, seen := m.events[ev.ID]; seen {
		return true, nil
	}
	m.events[ev.ID] = ev

	if u != nil && ev.OrderID != "" {
		_, _ = m.apply(ev.OrderID, *u)
	}

	return false, nil
}

// plant stores an order as it would be after some history, for the admin and
// reconcile tests.
func (m *memRepository) plant(c Contribution) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if c.Currency == "" {
		c.Currency = "COP"
	}
	m.orders[c.OrderID] = c
}

// sorted returns the orders newest first, as the SQL does.
func (m *memRepository) sorted() []Contribution {
	all := make([]Contribution, 0, len(m.orders))
	for _, c := range m.orders {
		all = append(all, c)
	}
	slices.SortFunc(all, func(a, b Contribution) int { return b.CreatedAt.Compare(a.CreatedAt) })

	return all
}

func (m *memRepository) ListContributions(_ context.Context, status Status, offset, limit uint) ([]Contribution, uint, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var matched []Contribution
	for _, c := range m.sorted() {
		if status == "" || c.Status == status {
			matched = append(matched, c)
		}
	}

	count := uint(len(matched))
	if offset >= count {
		return []Contribution{}, count, nil
	}

	return matched[offset:min(offset+limit, count)], count, nil
}

func (m *memRepository) Summarize(_ context.Context, since time.Time) (Summary, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	summary := Summary{Counts: map[Status]int64{}}
	for _, s := range Statuses {
		summary.Counts[s] = 0
	}

	for _, c := range m.orders {
		summary.Counts[c.Status]++
		if c.Status != StatusApproved {
			continue
		}

		charged := c.Amount
		if c.TotalCharged != nil {
			charged = *c.TotalCharged
		}
		summary.ApprovedTotal += charged
		if c.ApprovedAt != nil && !c.ApprovedAt.Before(since) {
			summary.ApprovedRecent += charged
		}
		if c.ApprovedAt != nil && (summary.LastApprovedAt == nil || c.ApprovedAt.After(*summary.LastApprovedAt)) {
			summary.LastApprovedAt = c.ApprovedAt
		}
	}

	return summary, nil
}

func (m *memRepository) ListOpen(_ context.Context, createdBefore, createdAfter time.Time, limit int) ([]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var ids []string
	for _, c := range m.sorted() {
		open := c.Status == StatusPending || (c.Status == StatusCreated && !c.CreatedAt.Before(createdAfter))
		if c.CreatedAt.Before(createdBefore) && open && len(ids) < limit {
			ids = append(ids, c.OrderID)
		}
	}

	return ids, nil
}

func (m *memRepository) DeleteAbandoned(_ context.Context, before time.Time) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.deletedBefore = before

	var n int64
	for id, c := range m.orders {
		if c.Status == StatusCreated && c.CreatedAt.Before(before) {
			delete(m.orders, id)
			n++
		}
	}

	return n, nil
}

// fakeGateway answers the voucher query with a fixed payment or error, and
// counts the calls.
type fakeGateway struct {
	payment bold.Payment
	err     error
	calls   int
	// byOrder answers specific orders; the rest get payment/err.
	byOrder map[string]bold.Payment
}

func (f *fakeGateway) Payment(_ context.Context, orderID string) (bold.Payment, error) {
	f.calls++

	if f.byOrder != nil {
		if p, ok := f.byOrder[orderID]; ok {
			return p, nil
		}
	}

	return f.payment, f.err
}

// fakeGuard stands in for *auth.Module: it rejects the request, or lets it
// through having written the role the JWT gate writes, so httpx.RequireAdmin —
// which the module does not inject — is still what decides.
type fakeGuard struct {
	authStatus int
	role       string
}

func (g fakeGuard) RequireAuth() fiber.Handler {
	return func(c fiber.Ctx) error {
		if g.authStatus != 0 {
			return c.SendStatus(g.authStatus)
		}

		c.Locals(httpx.LocalRole, g.role)

		return c.Next()
	}
}

var adminGuard = fakeGuard{role: httpx.RoleAdmin}

const testSecret = "secreta"

var enabledConfig = Config{
	FrontendURL: "https://finexia.me/",
	APIKey:      "identidad",
	SecretKey:   testSecret,
}

// sign reproduces Bold's webhook signature for a test body.
func sign(body, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(base64.StdEncoding.EncodeToString([]byte(body))))

	return hex.EncodeToString(mac.Sum(nil))
}

// saleEvent builds a webhook body for one of our orders.
func saleEvent(eventID, eventType, orderID string, total int64) string {
	return `{"id":"` + eventID + `","type":"` + eventType + `","subject":"PAY-` + eventID + `",` +
		`"data":{"payment_id":"PAY-` + eventID + `","payment_method":"CARD",` +
		`"amount":{"currency":"COP","total":` + strconv.FormatInt(total, 10) + `},` +
		`"metadata":{"reference":"` + orderID + `"}}}`
}
