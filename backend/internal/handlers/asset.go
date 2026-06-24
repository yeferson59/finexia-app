package handlers

import "github.com/gofiber/fiber/v3"

func (h *Handlers) SyncAssetPrices(c fiber.Ctx) error {
	assets, errs := h.services.SyncAssetPrices(h.ctx)

	if len(errs) > 0 && len(assets) == 0 {
		return h.responseInternalServerError(c, "Asset price sync failed", errs[0].Error())
	}

	return h.responseStatusOk(c, "Asset prices synced", "", assets)
}
