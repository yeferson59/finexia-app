package handlers

import "github.com/gofiber/fiber/v3"

func (h *Handlers) SyncExchangeRates(c fiber.Ctx) error {
	rates, errs := h.services.SyncExchangeRates(c.Context())

	if len(errs) > 0 && len(rates) == 0 {
		return h.responseInternalServerError(c, "Exchange rate sync failed", errs[0].Error())
	}

	return h.responseStatusOk(c, "Exchange rates synced", "", rates)
}
