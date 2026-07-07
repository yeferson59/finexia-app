package handlers

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/paginate"

	"github.com/yeferson59/finexia-app/internal/dtos/portfolio"
)

func (h *Handlers) SyncExchangeRates(c fiber.Ctx) error {
	rates, errs := h.services.SyncExchangeRates(c)

	if len(errs) > 0 && len(rates) == 0 {
		return h.responseInternalServerError(c, "Exchange rate sync failed", errs[0].Error())
	}

	return h.responseStatusOk(c, "Exchange rates synced", "", rates)
}

func (h *Handlers) GetExchangeRates(c fiber.Ctx) error {
	paginateInfo, ok := paginate.FromContext(c)
	if !ok {
		return h.responseInternalServerError(c, "", "paginate info not found")
	}

	rates, err := h.services.GetExchangeRates(c, uint(paginateInfo.Offset), uint(paginateInfo.Limit))
	if err != nil {
		return h.responseFromDomain(c, err, "", "")
	}

	return h.responseStatusOk(c, "", "", rates)
}

func (h *Handlers) CreateExchangeRate(c fiber.Ctx) error {
	var req portfolio.CreateExchangeRateRequestDTO
	if err := c.Bind().JSON(&req); err != nil {
		return h.responseBadRequest(c, "Invalid request", err.Error())
	}

	from := strings.ToUpper(strings.TrimSpace(req.FromCurrency))
	to := strings.ToUpper(strings.TrimSpace(req.ToCurrency))

	rate, err := h.services.CreateExchangeRate(c, from, to, req.Rate)
	if err != nil {
		return h.responseFromDomain(c, err, "Error creating exchange rate", "Could not create exchange rate")
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Exchange rate created",
		"details": "Exchange rate created successfully",
		"data":    rate,
	})
}

func (h *Handlers) UpdateExchangeRate(c fiber.Ctx) error {
	id, err := h.getParamUUID(c, "id")
	if err != nil {
		return h.responseBadRequest(c, "Invalid exchange rate ID", err.Error())
	}

	var req portfolio.UpdateExchangeRateRequestDTO
	if err := c.Bind().JSON(&req); err != nil {
		return h.responseBadRequest(c, "Invalid request", err.Error())
	}

	rate, err := h.services.UpdateExchangeRate(c, id, req.Rate)
	if err != nil {
		return h.responseFromDomain(c, err, "Error updating exchange rate", "Could not update exchange rate")
	}

	return h.responseStatusOk(c, "Exchange rate updated", "Exchange rate updated successfully", rate)
}

func (h *Handlers) ImportExchangeRates(c fiber.Ctx) error {
	data, filename, err := readImportFile(c)
	if err != nil {
		return h.responseBadRequest(c, "Invalid file", err.Error())
	}

	result, err := h.services.ImportExchangeRatesFromFile(c, data, filename, c.FormValue("sheet"))
	if err != nil {
		return h.responseFromDomain(c, err, "Error importing exchange rates", "Could not import the uploaded exchange rates")
	}

	return h.responseStatusOk(c, "Exchange rates imported", "Spreadsheet imported successfully", result)
}

func (h *Handlers) SyncSingleExchangeRate(c fiber.Ctx) error {
	id, err := h.getParamUUID(c, "id")
	if err != nil {
		return h.responseBadRequest(c, "Invalid exchange rate ID", err.Error())
	}

	rate, err := h.services.SyncExchangeRateByID(c, id)
	if err != nil {
		return h.responseFromDomain(c, err, "Exchange rate sync failed", err.Error())
	}

	return h.responseStatusOk(c, "Exchange rate synced", "", rate)
}
