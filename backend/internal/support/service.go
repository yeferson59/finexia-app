package support

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/yeferson59/finexia-app/internal/platform/bold"
	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

// Config is the part of the environment this module reads.
type Config struct {
	// FrontendURL is the app's origin; Bold sends the payer back to /apoyar on it.
	FrontendURL string
	// APIKey and SecretKey are Bold's identity and secret keys. Without both,
	// contributions are off and every route says so.
	APIKey    string
	SecretKey string
	// WebhookTestMode verifies webhooks with the empty key Bold's test
	// environment signs them with. The composition root refuses it in production.
	WebhookTestMode bool
}

func (c Config) enabled() bool {
	return c.APIKey != "" && c.SecretKey != ""
}

func (c Config) webhookSecret() string {
	if c.WebhookTestMode {
		return ""
	}

	return c.SecretKey
}

// gateway is the one question this module asks Bold; *bold.Client answers it.
type gateway interface {
	Payment(ctx context.Context, orderID string) (bold.Payment, error)
}

type service struct {
	repo    Repository
	gateway gateway
	cfg     Config
	log     logger.Logger
	now     func() time.Time
}

func newService(repo Repository, gw gateway, cfg Config, log logger.Logger) *service {
	if log == nil {
		log = logger.Noop()
	}

	return new(service{repo: repo, gateway: gw, cfg: cfg, log: log, now: time.Now})
}

// Enabled reports whether contributions can be taken right now.
func (s *service) Enabled() bool {
	return s.cfg.enabled()
}

// CreateCheckout stores a new order and signs it for the payment button.
//
// The row goes in before the payer sees Bold, so every webhook for an order of
// ours finds it.
func (s *service) CreateCheckout(ctx context.Context, amount int64) (Checkout, error) {
	if !s.cfg.enabled() {
		return Checkout{}, ErrDisabled
	}

	if amount < MinAmount || amount > MaxAmount {
		return Checkout{}, ErrAmountOutOfRange
	}

	orderID := newOrderID(s.now())
	if err := s.repo.CreateContribution(ctx, orderID, amount, bold.Currency); err != nil {
		return Checkout{}, err
	}

	return Checkout{
		OrderID:            orderID,
		Currency:           bold.Currency,
		Amount:             strconv.FormatInt(amount, 10),
		APIKey:             s.cfg.APIKey,
		IntegritySignature: bold.IntegritySignature(orderID, amount, bold.Currency, s.cfg.SecretKey),
		RedirectionURL:     strings.TrimRight(s.cfg.FrontendURL, "/") + "/apoyar",
		Description:        "Aporte a Finexia",
		RenderMode:         "embedded",
	}, nil
}

// Contribution returns an order's status for the page the payer comes back to.
//
// The payer usually arrives before the webhook does, so an order that is not
// settled yet is asked about to Bold and brought up to date on the way. If Bold
// does not answer, the stored status is the answer: never the one in the URL.
func (s *service) Contribution(ctx context.Context, orderID string) (Contribution, error) {
	if !isOrderID(orderID) {
		return Contribution{}, ErrContributionNotFound
	}

	c, err := s.repo.GetContribution(ctx, orderID)
	if err != nil || c.Status.settled() || !s.cfg.enabled() || s.gateway == nil {
		return c, err
	}

	refreshed, err := s.refresh(ctx, c)
	if errors.Is(err, errGateway) {
		s.log.Warn(ctx, "support: bold payment voucher failed", logger.Str("orderId", orderID), logger.Err(err))
		return c, nil
	}

	return refreshed, err
}

// errGateway marks a failure to get an answer from Bold, as opposed to one
// storing it: the page shrugs the first off and shows what is stored.
var errGateway = errors.New("bold did not answer")

// refresh asks Bold about c and records what it says. An answer that changes
// nothing — no transaction yet, a status it does not know — returns c as is.
func (s *service) refresh(ctx context.Context, c Contribution) (Contribution, error) {
	payment, err := s.gateway.Payment(ctx, c.OrderID)
	if err != nil {
		return c, errors.Join(errGateway, err)
	}

	to, ok := statusFromVoucher(payment.Status)
	if !ok {
		return c, nil
	}

	return s.repo.ApplyUpdate(ctx, c.OrderID, Update{
		Status:        to,
		TotalCharged:  payment.Total,
		PaymentID:     payment.TransactionID,
		PaymentMethod: payment.PaymentMethod,
	})
}

// HandleWebhook verifies and records one of Bold's notifications.
//
// Bold wants a 200 within two seconds and retries for a day otherwise, so this
// does one transaction and nothing that waits on anyone else. A notification
// for a payment that is not one of our orders is recorded and acknowledged all
// the same: refusing it would only make Bold retry it five times.
func (s *service) HandleWebhook(ctx context.Context, body []byte, signature string) error {
	if !s.cfg.enabled() {
		return ErrDisabled
	}

	if !bold.VerifyWebhook(body, signature, s.cfg.webhookSecret()) {
		return ErrInvalidSignature
	}

	ev, err := bold.ParseEvent(body)
	if err != nil {
		return errors.Join(ErrMalformedEvent, err)
	}

	recorded := PaymentEvent{ID: ev.ID, Type: ev.Type, PaymentID: ev.PaymentID, Total: ev.Total}
	if isOrderID(ev.Reference) {
		recorded.OrderID = ev.Reference
	}

	var update *Update
	if to, ok := statusFromEvent(ev.Type); ok {
		update = new(Update{
			Status:        to,
			TotalCharged:  ev.Total,
			PaymentID:     ev.PaymentID,
			PaymentMethod: ev.PaymentMethod,
		})
	}

	duplicate, err := s.repo.RecordEvent(ctx, recorded, update)
	if err != nil {
		return err
	}

	s.log.Info(ctx, "support: bold webhook",
		logger.Str("eventId", ev.ID),
		logger.Str("type", ev.Type),
		logger.Str("orderId", recorded.OrderID),
		logger.Bool("duplicate", duplicate),
	)

	return nil
}

// ErrInvalidStatus is a status filter that is not one of Statuses.
var ErrInvalidStatus = errors.New("invalid status filter")

// ListContributions pages through the orders for the admin screen.
func (s *service) ListContributions(ctx context.Context, rawStatus string, offset, limit uint) ([]Contribution, uint, error) {
	status, ok := parseStatus(rawStatus)
	if !ok {
		return nil, 0, ErrInvalidStatus
	}

	return s.repo.ListContributions(ctx, status, offset, limit)
}

// Summary adds up the contributions for the admin screen.
func (s *service) Summary(ctx context.Context) (Summary, error) {
	return s.repo.Summarize(ctx, s.now().Add(-RecentWindow))
}

// The reconciliation's timing. Each value is a trade-off worth stating:
const (
	// reconcileGrace leaves a fresh order to its webhook, which normally lands
	// seconds after the payment. Asking Bold sooner would only race it.
	reconcileGrace = 15 * time.Minute
	// abandonAfter is how long an order may sit in created before it is taken
	// for a checkout nobody paid. Every hourly run asks Bold about it until
	// then, so a payment that did happen is found long before the order goes.
	abandonAfter = 7 * 24 * time.Hour
	// reconcileBatch caps the voucher queries of one run; the rest wait for
	// the next.
	reconcileBatch = 100
)

// ReconcileCounts is what one reconciliation did.
type ReconcileCounts struct {
	Checked   int
	Updated   int
	Failed    int
	Abandoned int64
}

// Reconcile does what the webhook would have done had it arrived: it asks Bold
// about every order still open and records what Bold says. Then it deletes the
// orders nobody paid.
//
// A webhook can be lost — the app down for longer than Bold's day of retries,
// the webhook URL mistyped in Bold's panel — and a payer who closes the tab
// never comes back to /apoyar to trigger the query either. Without this, both
// orders would say `created` forever, one of them wrongly.
//
// A failed query is counted and skipped, not returned: one order Bold cannot
// answer about must not stop the rest. But the deletion only runs after a batch
// Bold answered in full, and never with contributions off: an order is only
// known to be abandoned if Bold was asked and said so, and deleting one that
// was in fact paid is the one mistake here that cannot be put right later.
func (s *service) Reconcile(ctx context.Context) (ReconcileCounts, error) {
	var counts ReconcileCounts

	if !s.cfg.enabled() || s.gateway == nil {
		return counts, nil
	}

	now := s.now()

	open, err := s.repo.ListOpen(ctx, now.Add(-reconcileGrace), now.Add(-abandonAfter), reconcileBatch)
	if err != nil {
		return counts, err
	}

	for _, orderID := range open {
		if err := ctx.Err(); err != nil {
			return counts, err
		}

		counts.Checked++

		before, err := s.repo.GetContribution(ctx, orderID)
		if err != nil {
			counts.Failed++
			continue
		}

		after, err := s.refresh(ctx, before)
		if err != nil {
			s.log.Warn(ctx, "support: reconcile failed", logger.Str("orderId", orderID), logger.Err(err))
			counts.Failed++
			continue
		}

		if after.Status != before.Status {
			counts.Updated++
		}
	}

	if counts.Failed > 0 {
		return counts, nil
	}

	counts.Abandoned, err = s.repo.DeleteAbandoned(ctx, now.Add(-abandonAfter))

	return counts, err
}
