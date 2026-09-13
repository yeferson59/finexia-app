package portfolio

// Cash balance handlers: the balances, their movements, and the three writes.

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/paginate"
	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/platform/currency"
	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

// GetCashBalances answers what each platform holds in cash, per portfolio and
// currency. Same ?currency= contract as the holdings: every row carries its
// value in one currency, so the balances add up.
func (h *handler) GetCashBalances(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	var req CurrencyDTO
	if err := c.Bind().Query(&req); err != nil {
		return httpx.BadRequest(c, "Unprocess query currency", "no process currency")
	}

	if req.Currency != money.XXX && !currency.IsSupported(req.Currency) {
		return httpx.BadRequest(c, "Unsupported currency", "currency must be one of: "+currency.List())
	}

	balances, err := h.service.GetCashBalances(c, userID, req.Currency)
	if err != nil {
		return httpx.FromDomain(c, err, "Error retrieving cash balances", "Could not retrieve cash balances")
	}

	return httpx.OK(c, "Cash balances retrieved", "Cash balances retrieved successfully", balances)
}

func (h *handler) GetCashMovements(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	paginateInfo, ok := paginate.FromContext(c)
	if !ok {
		return httpx.InternalServerError(c, "", "paginate info not found")
	}

	page := paginateInfo.Page
	limit := paginateInfo.Limit

	movements, total, err := h.service.GetCashMovements(c, userID, page, limit)
	if err != nil {
		return httpx.FromDomain(c, err, "Error retrieving cash movements", "Could not retrieve cash movements")
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + limit - 1) / limit
	}

	return httpx.OK(c, "Cash movements retrieved", "Cash movements retrieved successfully",
		PaginatedCashMovementsDTO{
			Data:       movements,
			Total:      total,
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
		},
	)
}

func (h *handler) CreateCashMovement(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	req, err := httpx.Bind[CreateCashMovementRequestDTO](c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid request", err.Error())
	}

	movement, err := h.service.CreateCashMovement(c, userID, req.PortfolioID, req.SourceID, req.Input())
	if err != nil {
		return httpx.FromDomain(c, err, "Error recording cash movement", "Could not record cash movement")
	}

	return httpx.OK(c, "Cash movement recorded", "Cash movement recorded successfully", movement)
}

func (h *handler) UpdateCashMovement(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	txnID, err := httpx.ParamUUID(c, "txnId")
	if err != nil {
		return httpx.BadRequest(c, "Invalid movement ID", err.Error())
	}

	req, err := httpx.Bind[UpdateCashMovementRequestDTO](c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid request", err.Error())
	}

	movement, err := h.service.UpdateCashMovement(c, userID, txnID, req.Input())
	if err != nil {
		return httpx.FromDomain(c, err, "Error updating cash movement", "Could not update cash movement")
	}

	return httpx.OK(c, "Cash movement updated", "Cash movement updated successfully", movement)
}

func (h *handler) DeleteCashMovement(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	txnID, err := httpx.ParamUUID(c, "txnId")
	if err != nil {
		return httpx.BadRequest(c, "Invalid movement ID", err.Error())
	}

	if err := h.service.DeleteCashMovement(c, userID, txnID); err != nil {
		return httpx.FromDomain(c, err, "Error deleting cash movement", "Could not delete cash movement")
	}

	return httpx.OK(c, "Cash movement deleted", "Cash movement deleted successfully", nil)
}
