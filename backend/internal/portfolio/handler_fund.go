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

	// ?withPositions=true also deletes the positions that still hold it.
	withPositions := c.Query("withPositions") == "true"

	if err := h.service.DeleteFund(c, userID, assetID, withPositions); err != nil {
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

func (h *handler) GetFundMovements(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	assetID, err := httpx.ParamUUID(c, "assetId")
	if err != nil {
		return httpx.BadRequest(c, "Invalid fund ID", err.Error())
	}

	movements, err := h.service.GetFundMovements(c, userID, assetID)
	if err != nil {
		return httpx.FromDomain(c, err, "Error retrieving fund movements", "Could not retrieve fund movements")
	}

	return httpx.OK(c, "Fund movements retrieved", "Fund movements retrieved successfully", movements)
}

// ContributeToFund puts money into a fund: an amount in one followed by
// balance, units at a unit value in one followed by units.
func (h *handler) ContributeToFund(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	assetID, err := httpx.ParamUUID(c, "assetId")
	if err != nil {
		return httpx.BadRequest(c, "Invalid fund ID", err.Error())
	}

	req, err := httpx.Bind[ContributeToFundRequestDTO](c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid request", err.Error())
	}

	movement, err := h.service.ContributeToFund(c, userID, assetID, req.Input())
	if err != nil {
		return httpx.FromDomain(c, err, "Error recording contribution", "Could not record contribution")
	}

	return httpx.OK(c, "Contribution recorded", "Contribution recorded successfully", movement)
}

// WithdrawFromFund takes money out of a position of a fund.
func (h *handler) WithdrawFromFund(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	assetID, err := httpx.ParamUUID(c, "assetId")
	if err != nil {
		return httpx.BadRequest(c, "Invalid fund ID", err.Error())
	}

	req, err := httpx.Bind[WithdrawFromFundRequestDTO](c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid request", err.Error())
	}

	movement, err := h.service.WithdrawFromFund(c, userID, assetID, req.Input())
	if err != nil {
		return httpx.FromDomain(c, err, "Error recording withdrawal", "Could not record withdrawal")
	}

	return httpx.OK(c, "Withdrawal recorded", "Withdrawal recorded successfully", movement)
}

// UpdateFundMovement restates a contribution or withdrawal.
func (h *handler) UpdateFundMovement(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	txnID, err := httpx.ParamUUID(c, "txnId")
	if err != nil {
		return httpx.BadRequest(c, "Invalid movement ID", err.Error())
	}

	req, err := httpx.Bind[UpdateFundMovementRequestDTO](c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid request", err.Error())
	}

	movement, err := h.service.UpdateFundMovement(c, userID, txnID, req.Input())
	if err != nil {
		return httpx.FromDomain(c, err, "Error updating fund movement", "Could not update fund movement")
	}

	return httpx.OK(c, "Fund movement updated", "Fund movement updated successfully", movement)
}

// DeleteFundMovement takes a contribution or withdrawal back.
func (h *handler) DeleteFundMovement(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	txnID, err := httpx.ParamUUID(c, "txnId")
	if err != nil {
		return httpx.BadRequest(c, "Invalid movement ID", err.Error())
	}

	if err := h.service.DeleteFundMovement(c, userID, txnID); err != nil {
		return httpx.FromDomain(c, err, "Error deleting fund movement", "Could not delete fund movement")
	}

	return httpx.OK(c, "Fund movement deleted", "Fund movement deleted successfully", nil)
}

// SaveFundMarks records several marks at once: the table a statement prints.
func (h *handler) SaveFundMarks(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	assetID, err := httpx.ParamUUID(c, "assetId")
	if err != nil {
		return httpx.BadRequest(c, "Invalid fund ID", err.Error())
	}

	req, err := httpx.Bind[SaveFundMarksRequestDTO](c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid request", err.Error())
	}

	saved, err := h.service.SaveFundMarks(c, userID, assetID, req.Input())
	if err != nil {
		return httpx.FromDomain(c, err, "Error saving fund marks", "Could not save fund marks")
	}

	return httpx.OK(c, "Fund marks saved", "Fund marks saved successfully", map[string]int{"saved": saved})
}

// GetFundPerformance answers how a fund did: returns by period, money, series.
func (h *handler) GetFundPerformance(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	assetID, err := httpx.ParamUUID(c, "assetId")
	if err != nil {
		return httpx.BadRequest(c, "Invalid fund ID", err.Error())
	}

	perf, err := h.service.GetFundPerformance(c, userID, assetID)
	if err != nil {
		return httpx.FromDomain(c, err, "Error retrieving fund performance", "Could not retrieve fund performance")
	}

	return httpx.OK(c, "Fund performance retrieved", "Fund performance retrieved successfully", perf)
}

// SearchPublicFunds finds funds of the SFC's catalog: GET /funds/catalog?q=.
func (h *handler) SearchPublicFunds(c fiber.Ctx) error {
	funds, err := h.service.SearchPublicFunds(c, c.Query("q"))
	if err != nil {
		return httpx.FromDomain(c, err, "Error searching published funds", "Could not search published funds")
	}

	return httpx.OK(c, "Published funds retrieved", "Published funds retrieved successfully", funds)
}

// LinkFund links a fund to one of the SFC's catalog, whose published unit
// values become its marks.
func (h *handler) LinkFund(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	assetID, err := httpx.ParamUUID(c, "assetId")
	if err != nil {
		return httpx.BadRequest(c, "Invalid fund ID", err.Error())
	}

	req, err := httpx.Bind[LinkFundRequestDTO](c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid request", err.Error())
	}

	fund, err := h.service.LinkFund(c, userID, assetID, req.PublicFundID)
	if err != nil {
		return httpx.FromDomain(c, err, "Error linking fund", "Could not link fund")
	}

	return httpx.OK(c, "Fund linked", "Fund linked successfully", fund)
}

// UnlinkFund ends a fund's link and takes back the marks it brought.
func (h *handler) UnlinkFund(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	assetID, err := httpx.ParamUUID(c, "assetId")
	if err != nil {
		return httpx.BadRequest(c, "Invalid fund ID", err.Error())
	}

	fund, err := h.service.UnlinkFund(c, userID, assetID)
	if err != nil {
		return httpx.FromDomain(c, err, "Error unlinking fund", "Could not unlink fund")
	}

	return httpx.OK(c, "Fund unlinked", "Fund unlinked successfully", fund)
}
