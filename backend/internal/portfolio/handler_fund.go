package portfolio

// Fund handlers: the investment funds an owner follows and the marks that
// price them.

import (
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

func (h *handler) GetFunds(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	funds, err := h.service.GetFunds(c, userID)
	if err != nil {
		return httpx.FromDomain(c, err, "Error retrieving funds", "Could not retrieve funds")
	}

	return httpx.OK(c, "Funds retrieved", "Funds retrieved successfully", funds)
}

func (h *handler) GetFund(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	assetID, err := httpx.ParamUUID(c, "assetId")
	if err != nil {
		return httpx.BadRequest(c, "Invalid fund ID", err.Error())
	}

	fund, err := h.service.GetFund(c, userID, assetID)
	if err != nil {
		return httpx.FromDomain(c, err, "Error retrieving fund", "Could not retrieve fund")
	}

	return httpx.OK(c, "Fund retrieved", "Fund retrieved successfully", fund)
}

func (h *handler) CreateFund(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	req, err := httpx.Bind[CreateFundRequestDTO](c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid request", err.Error())
	}

	fund, err := h.service.CreateFund(c, userID, req.Input())
	if err != nil {
		return httpx.FromDomain(c, err, "Error creating fund", "Could not create fund")
	}

	return httpx.OK(c, "Fund created", "Fund created successfully", fund)
}

func (h *handler) DeleteFund(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	assetID, err := httpx.ParamUUID(c, "assetId")
	if err != nil {
		return httpx.BadRequest(c, "Invalid fund ID", err.Error())
	}

	if err := h.service.DeleteFund(c, userID, assetID); err != nil {
		return httpx.FromDomain(c, err, "Error deleting fund", "Could not delete fund")
	}

	return httpx.OK(c, "Fund deleted", "Fund deleted successfully", nil)
}

func (h *handler) GetFundMarks(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	assetID, err := httpx.ParamUUID(c, "assetId")
	if err != nil {
		return httpx.BadRequest(c, "Invalid fund ID", err.Error())
	}

	marks, err := h.service.GetFundMarks(c, userID, assetID)
	if err != nil {
		return httpx.FromDomain(c, err, "Error retrieving fund marks", "Could not retrieve fund marks")
	}

	return httpx.OK(c, "Fund marks retrieved", "Fund marks retrieved successfully", marks)
}

// SaveFundMark records what a fund was worth on a day, replacing the mark it
// had on that day.
func (h *handler) SaveFundMark(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	assetID, err := httpx.ParamUUID(c, "assetId")
	if err != nil {
		return httpx.BadRequest(c, "Invalid fund ID", err.Error())
	}

	req, err := httpx.Bind[SaveFundMarkRequestDTO](c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid request", err.Error())
	}

	mark, err := h.service.SaveFundMark(c, userID, assetID, req.Input())
	if err != nil {
		return httpx.FromDomain(c, err, "Error saving fund mark", "Could not save fund mark")
	}

	return httpx.OK(c, "Fund mark saved", "Fund mark saved successfully", mark)
}

// DeleteFundMark takes back a fund's mark. The day is the path's last segment,
// as YYYY-MM-DD: a mark is identified by the day it is for.
func (h *handler) DeleteFundMark(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	assetID, err := httpx.ParamUUID(c, "assetId")
	if err != nil {
		return httpx.BadRequest(c, "Invalid fund ID", err.Error())
	}

	date, err := time.Parse(time.DateOnly, c.Params("date"))
	if err != nil {
		return httpx.BadRequest(c, "Invalid date", "date must be YYYY-MM-DD")
	}

	if err := h.service.DeleteFundMark(c, userID, assetID, date); err != nil {
		return httpx.FromDomain(c, err, "Error deleting fund mark", "Could not delete fund mark")
	}

	return httpx.OK(c, "Fund mark deleted", "Fund mark deleted successfully", nil)
}
