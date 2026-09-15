package portfolio

// Cash rate handlers: the rates cash accounts earn, and the four writes.

import (
	"github.com/gofiber/fiber/v3"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

func (h *handler) GetCashRates(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	rates, err := h.service.GetCashRates(c, userID)
	if err != nil {
		return httpx.FromDomain(c, err, "Error retrieving cash rates", "Could not retrieve cash rates")
	}

	return httpx.OK(c, "Cash rates retrieved", "Cash rates retrieved successfully", rates)
}

func (h *handler) CreateCashRate(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	req, err := httpx.Bind[CreateCashRateRequestDTO](c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid request", err.Error())
	}

	rate, err := h.service.CreateCashRate(c, userID, req.Input())
	if err != nil {
		return httpx.FromDomain(c, err, "Error recording cash rate", "Could not record cash rate")
	}

	return httpx.OK(c, "Cash rate recorded", "Cash rate recorded successfully", rate)
}

func (h *handler) UpdateCashRate(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	rateID, err := httpx.ParamUUID(c, "rateId")
	if err != nil {
		return httpx.BadRequest(c, "Invalid rate ID", err.Error())
	}

	req, err := httpx.Bind[UpdateCashRateRequestDTO](c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid request", err.Error())
	}

	rate, err := h.service.UpdateCashRate(c, userID, rateID, req.Input())
	if err != nil {
		return httpx.FromDomain(c, err, "Error updating cash rate", "Could not update cash rate")
	}

	return httpx.OK(c, "Cash rate updated", "Cash rate updated successfully", rate)
}

func (h *handler) EndCashRate(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	rateID, err := httpx.ParamUUID(c, "rateId")
	if err != nil {
		return httpx.BadRequest(c, "Invalid rate ID", err.Error())
	}

	req, err := httpx.Bind[EndCashRateRequestDTO](c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid request", err.Error())
	}

	rate, err := h.service.EndCashRate(c, userID, rateID, req.EndsOn)
	if err != nil {
		return httpx.FromDomain(c, err, "Error ending cash rate", "Could not end cash rate")
	}

	return httpx.OK(c, "Cash rate ended", "Cash rate ended successfully", rate)
}

func (h *handler) DeleteCashRate(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	rateID, err := httpx.ParamUUID(c, "rateId")
	if err != nil {
		return httpx.BadRequest(c, "Invalid rate ID", err.Error())
	}

	if err := h.service.DeleteCashRate(c, userID, rateID); err != nil {
		return httpx.FromDomain(c, err, "Error deleting cash rate", "Could not delete cash rate")
	}

	return httpx.OK(c, "Cash rate deleted", "Cash rate deleted successfully", nil)
}
