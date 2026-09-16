package portfolio

// Cash pocket handlers: the subaccounts a cash account's money can sit in, and
// the move between two of them.

import (
	"github.com/gofiber/fiber/v3"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

func (h *handler) GetCashPockets(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	pockets, err := h.service.GetCashPockets(c, userID)
	if err != nil {
		return httpx.FromDomain(c, err, "Error retrieving cash pockets", "Could not retrieve cash pockets")
	}

	return httpx.OK(c, "Cash pockets retrieved", "Cash pockets retrieved successfully", pockets)
}

func (h *handler) CreateCashPocket(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	req, err := httpx.Bind[CreateCashPocketRequestDTO](c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid request", err.Error())
	}

	pocket, err := h.service.CreateCashPocket(c, userID, req.Input())
	if err != nil {
		return httpx.FromDomain(c, err, "Error creating cash pocket", "Could not create cash pocket")
	}

	return httpx.OK(c, "Cash pocket created", "Cash pocket created successfully", pocket)
}

func (h *handler) RenameCashPocket(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	pocketID, err := httpx.ParamUUID(c, "pocketId")
	if err != nil {
		return httpx.BadRequest(c, "Invalid pocket ID", err.Error())
	}

	req, err := httpx.Bind[RenameCashPocketRequestDTO](c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid request", err.Error())
	}

	pocket, err := h.service.RenameCashPocket(c, userID, pocketID, req.Input())
	if err != nil {
		return httpx.FromDomain(c, err, "Error renaming cash pocket", "Could not rename cash pocket")
	}

	return httpx.OK(c, "Cash pocket renamed", "Cash pocket renamed successfully", pocket)
}

func (h *handler) DeleteCashPocket(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	pocketID, err := httpx.ParamUUID(c, "pocketId")
	if err != nil {
		return httpx.BadRequest(c, "Invalid pocket ID", err.Error())
	}

	if err := h.service.DeleteCashPocket(c, userID, pocketID); err != nil {
		return httpx.FromDomain(c, err, "Error deleting cash pocket", "Could not delete cash pocket")
	}

	return httpx.OK(c, "Cash pocket deleted", "Cash pocket deleted successfully", nil)
}

func (h *handler) MoveCash(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	req, err := httpx.Bind[MoveCashRequestDTO](c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid request", err.Error())
	}

	move, err := h.service.MoveCash(c, userID, req.PortfolioID, req.SourceID, req.Input())
	if err != nil {
		return httpx.FromDomain(c, err, "Error moving cash", "Could not move cash")
	}

	return httpx.OK(c, "Cash moved", "Cash moved successfully", move)
}
