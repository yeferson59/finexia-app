package handlers

import (
	"strings"

	"github.com/gofiber/fiber/v3"

	"github.com/yeferson59/finexia-app/internal/dtos/portfolio"
	"github.com/yeferson59/finexia-app/internal/entities"
)

func (h *Handlers) SyncAssetPrices(c fiber.Ctx) error {
	assets, errs := h.services.SyncAssetPrices(h.ctx)

	if len(errs) > 0 && len(assets) == 0 {
		return h.responseInternalServerError(c, "Asset price sync failed", errs[0].Error())
	}

	return h.responseStatusOk(c, "Asset prices synced", "", assets)
}

func (h *Handlers) CreateAsset(c fiber.Ctx) error {
	var req portfolio.CreateAssetRequestDTO
	if err := c.Bind().JSON(&req); err != nil {
		return h.responseBadRequest(c, "Invalid request", err.Error())
	}

	req.Ticker = strings.ToUpper(strings.TrimSpace(req.Ticker))
	req.Currency = strings.ToUpper(strings.TrimSpace(req.Currency))

	asset, err := h.services.CreateAsset(h.ctx, req.Ticker, req.Name, entities.AssetType(req.AssetType), req.Exchange, req.Currency)
	if err != nil {
		return h.responseFromDomain(c, err, "Error creating asset", "Could not create asset")
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Asset created",
		"details": "Asset created successfully",
		"data":    asset,
	})
}
