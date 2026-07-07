package handlers

import (
	"strings"

	"github.com/gofiber/fiber/v3"

	"github.com/yeferson59/finexia-app/internal/dtos/portfolio"
	"github.com/yeferson59/finexia-app/internal/entities"
)

func (h *Handlers) SyncAssetPrices(c fiber.Ctx) error {
	assets, errs := h.services.SyncAssetPrices(c)

	if len(errs) > 0 && len(assets) == 0 {
		return h.responseInternalServerError(c, "Asset price sync failed", errs[0].Error())
	}

	return h.responseStatusOk(c, "Asset prices synced", "", assets)
}

func (h *Handlers) SyncSingleAsset(c fiber.Ctx) error {
	assetID, err := h.getParamUUID(c, "id")
	if err != nil {
		return h.responseBadRequest(c, "Invalid asset ID", err.Error())
	}

	asset, err := h.services.SyncAssetByID(c, assetID)
	if err != nil {
		return h.responseFromDomain(c, err, "Asset sync failed", err.Error())
	}

	return h.responseStatusOk(c, "Asset price synced", "", asset)
}

func (h *Handlers) ImportAssets(c fiber.Ctx) error {
	data, filename, err := readImportFile(c)
	if err != nil {
		return h.responseBadRequest(c, "Invalid file", err.Error())
	}

	result, err := h.services.ImportAssetsFromFile(c, data, filename, c.FormValue("sheet"))
	if err != nil {
		return h.responseFromDomain(c, err, "Error importing assets", "Could not import the uploaded assets")
	}

	return h.responseStatusOk(c, "Assets imported", "Spreadsheet imported successfully", result)
}

func (h *Handlers) CreateAsset(c fiber.Ctx) error {
	var req portfolio.CreateAssetRequestDTO
	if err := c.Bind().JSON(&req); err != nil {
		return h.responseBadRequest(c, "Invalid request", err.Error())
	}

	req.Ticker = strings.ToUpper(strings.TrimSpace(req.Ticker))
	req.Currency = strings.ToUpper(strings.TrimSpace(req.Currency))

	assetType := entities.AssetType(req.AssetType)
	if !assetType.IsValid() {
		return h.responseBadRequest(c, "Invalid asset type", "Asset type must be one of: stock, etf, crypto, bond, cash, real_estate, commodity, other")
	}

	asset, err := h.services.CreateAsset(c, req.Ticker, req.Name, assetType, req.Exchange, req.Currency)
	if err != nil {
		return h.responseFromDomain(c, err, "Error creating asset", "Could not create asset")
	}

	return h.responseSuccess(c, fiber.StatusCreated, "Asset created", "Asset created successfully", asset)
}
