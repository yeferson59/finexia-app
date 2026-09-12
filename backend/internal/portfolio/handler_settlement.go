package portfolio

import (
	"github.com/gofiber/fiber/v3"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

// ChangeEntrySettlement restates a position in the currency its account really
// settled in, with the rate each of its transactions converted at.
//
// The cost currency used to be fixed when a position was opened, so correcting
// it meant deleting the position and loading it again, history and all. This
// keeps the history and changes only what was wrong.
func (h *handler) ChangeEntrySettlement(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	entryID, err := httpx.ParamUUID(c, "entryId")
	if err != nil {
		return httpx.BadRequest(c, "Invalid entry ID", err.Error())
	}

	req, err := httpx.Bind[ChangeSettlementRequestDTO](c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid request", err.Error())
	}

	changed, err := h.service.ChangeEntrySettlement(c, userID, entryID, req.CostCurrency, req.RatesByTransaction())
	if err != nil {
		return httpx.FromDomain(c, err, "Error changing settlement currency", "Could not change the position's settlement currency")
	}

	return httpx.OK(c, "Settlement currency changed", "Settlement currency changed successfully",
		ChangeSettlementResponseDTO{CostCurrency: req.CostCurrency.String(), Transactions: changed})
}
