package support

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/paginate"

	"github.com/yeferson59/finexia-app/internal/platform/bold"
	"github.com/yeferson59/finexia-app/internal/platform/httpx"
	"github.com/yeferson59/finexia-app/pkg/dtos"
)

// checkoutRequest is the body of POST /support/checkouts. The service checks
// the range again: the validator is the first answer, not the rule.
type checkoutRequest struct {
	Amount int64 `json:"amount" validate:"required,min=1000,max=2000000"`
}

type handler struct {
	service *service
}

func newHandler(svc *service) *handler {
	return new(handler{svc})
}

// config serves GET /support: whether the page can take contributions, and
// between which amounts.
func (h *handler) config(c fiber.Ctx) error {
	return httpx.OK(c, "support", "", fiber.Map{
		"enabled":   h.service.Enabled(),
		"currency":  bold.Currency,
		"minAmount": MinAmount,
		"maxAmount": MaxAmount,
	})
}

// createCheckout serves POST /support/checkouts.
func (h *handler) createCheckout(c fiber.Ctx) error {
	req, err := httpx.Bind[checkoutRequest](c)
	if err != nil {
		return httpx.BadRequest(c, "invalid amount", "amount must be a whole number of pesos between 1000 and 2000000")
	}

	checkout, err := h.service.CreateCheckout(c, req.Amount)
	if errors.Is(err, ErrDisabled) {
		return httpx.ErrorAction(c, fiber.StatusServiceUnavailable, "contributions unavailable", "", "support:disabled")
	}
	if err != nil {
		return httpx.FromDomain(c, err, "failed to create checkout", "support:checkout")
	}

	return httpx.Success(c, fiber.StatusCreated, "checkout created", "", checkout)
}

// getCheckout serves GET /support/checkouts/:orderId.
func (h *handler) getCheckout(c fiber.Ctx) error {
	contribution, err := h.service.Contribution(c, c.Params("orderId"))
	if err != nil {
		return httpx.FromDomain(c, err, "failed to read contribution", "support:status")
	}

	return httpx.OK(c, "contribution", "", contribution)
}

// boldWebhook serves POST /support/webhooks/bold, which Bold calls through the
// app's public origin (docs/API.md §1.6).
//
// The signature covers the raw body, so the handler reads c.Body() and never
// binds it.
func (h *handler) boldWebhook(c fiber.Ctx) error {
	err := h.service.HandleWebhook(c, c.Body(), c.Get(bold.SignatureHeader))

	switch {
	case err == nil:
		return httpx.OK(c, "received", "", nil)
	case errors.Is(err, ErrInvalidSignature):
		return httpx.Unauthorized(c, "invalid signature", "")
	case errors.Is(err, ErrDisabled):
		// 503 and not 200: Bold retries for a day, which leaves time to
		// configure the keys without losing the notification.
		return httpx.ErrorAction(c, fiber.StatusServiceUnavailable, "contributions unavailable", "", "support:disabled")
	default:
		return httpx.FromDomain(c, err, "failed to process webhook", "support:webhook")
	}
}

// adminContribution is an order as the admin screen lists it: the public
// shape plus what identifies the payment in Bold's panel.
type adminContribution struct {
	Contribution
	PaymentID     string    `json:"paymentId"`
	PaymentMethod string    `json:"paymentMethod"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// listContributions (admin) serves GET /support/contributions?status=.
func (h *handler) listContributions(c fiber.Ctx) error {
	pageInfo, ok := paginate.FromContext(c)
	if !ok {
		return httpx.InternalServerError(c, "", "paginate info not found")
	}

	items, count, err := h.service.ListContributions(c, c.Query("status"), uint(pageInfo.Offset), uint(pageInfo.Limit))
	if errors.Is(err, ErrInvalidStatus) {
		return httpx.BadRequest(c, "invalid status", "status must be one of created, pending, rejected, approved, voided")
	}
	if err != nil {
		return httpx.FromDomain(c, err, "failed to list contributions", "support:list")
	}

	out := make([]adminContribution, len(items))
	for i, item := range items {
		out[i] = adminContribution{
			Contribution:  item,
			PaymentID:     item.PaymentID,
			PaymentMethod: item.PaymentMethod,
			UpdatedAt:     item.UpdatedAt,
		}
	}

	return httpx.OK(c, "contributions", "", dtos.FilterPagination[[]adminContribution]{
		Items:    out,
		MetaData: httpx.PaginationMetadata(pageInfo, count),
	})
}

// summary (admin) serves GET /support/contributions/summary.
func (h *handler) summary(c fiber.Ctx) error {
	summary, err := h.service.Summary(c)
	if err != nil {
		return httpx.FromDomain(c, err, "failed to summarize contributions", "support:summary")
	}

	return httpx.OK(c, "summary", "", summary)
}
