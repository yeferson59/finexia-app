package market

import (
	"errors"
	"io"
	"mime/multipart"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/paginate"
	"github.com/google/uuid"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

type handler struct {
	service *Service
}

func (h *handler) SyncAssetPrices(c fiber.Ctx) error {
	assets, errs := h.service.SyncAssetPrices(c)

	if len(errs) > 0 && len(assets) == 0 {
		return httpx.InternalServerError(c, "Asset price sync failed", errs[0].Error())
	}

	return httpx.OK(c, "Asset prices synced", "", assets)
}

func (h *handler) SyncSingleAsset(c fiber.Ctx) error {
	assetID, err := getParamUUID(c, "id")
	if err != nil {
		return httpx.BadRequest(c, "Invalid asset ID", err.Error())
	}

	asset, err := h.service.SyncAssetByID(c, assetID)
	if err != nil {
		return httpx.FromDomain(c, err, "Asset sync failed", err.Error())
	}

	return httpx.OK(c, "Asset price synced", "", asset)
}

func (h *handler) ImportAssets(c fiber.Ctx) error {
	data, filename, err := readImportFile(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid file", err.Error())
	}

	result, err := h.service.ImportAssetsFromFile(c, data, filename, c.FormValue("sheet"))
	if err != nil {
		return httpx.FromDomain(c, err, "Error importing assets", "Could not import the uploaded assets")
	}

	return httpx.OK(c, "Assets imported", "Spreadsheet imported successfully", result)
}

func (h *handler) CreateAsset(c fiber.Ctx) error {
	var req CreateAssetRequestDTO
	if err := c.Bind().JSON(&req); err != nil {
		return httpx.BadRequest(c, "Invalid request", err.Error())
	}

	req.Ticker = strings.ToUpper(strings.TrimSpace(req.Ticker))
	req.Currency = strings.ToUpper(strings.TrimSpace(req.Currency))

	assetType := AssetType(req.AssetType)
	if !assetType.IsValid() {
		return httpx.BadRequest(c, "Invalid asset type", "Asset type must be one of: stock, etf, crypto, bond, cash, real_estate, commodity, other")
	}

	asset, err := h.service.CreateAsset(c, req.Ticker, req.Name, assetType, req.Exchange, req.Currency)
	if err != nil {
		return httpx.FromDomain(c, err, "Error creating asset", "Could not create asset")
	}

	return httpx.Success(c, fiber.StatusCreated, "Asset created", "Asset created successfully", asset)
}

func (h *handler) SyncExchangeRates(c fiber.Ctx) error {
	rates, errs := h.service.SyncExchangeRates(c)

	if len(errs) > 0 && len(rates) == 0 {
		return httpx.InternalServerError(c, "Exchange rate sync failed", errs[0].Error())
	}

	return httpx.OK(c, "Exchange rates synced", "", rates)
}

func (h *handler) GetExchangeRates(c fiber.Ctx) error {
	paginateInfo, ok := paginate.FromContext(c)
	if !ok {
		return httpx.InternalServerError(c, "", "paginate info not found")
	}

	rates, err := h.service.GetExchangeRates(c, uint(paginateInfo.Offset), uint(paginateInfo.Limit))
	if err != nil {
		return httpx.FromDomain(c, err, "", "")
	}

	return httpx.OK(c, "", "", rates)
}

func (h *handler) CreateExchangeRate(c fiber.Ctx) error {
	var req CreateExchangeRateRequestDTO
	if err := c.Bind().JSON(&req); err != nil {
		return httpx.BadRequest(c, "Invalid request", err.Error())
	}

	from := strings.ToUpper(strings.TrimSpace(req.FromCurrency))
	to := strings.ToUpper(strings.TrimSpace(req.ToCurrency))

	rate, err := h.service.CreateExchangeRate(c, from, to, req.Rate)
	if err != nil {
		return httpx.FromDomain(c, err, "Error creating exchange rate", "Could not create exchange rate")
	}

	return httpx.Success(c, fiber.StatusCreated, "Exchange rate created", "Exchange rate created successfully", rate)
}

func (h *handler) UpdateExchangeRate(c fiber.Ctx) error {
	id, err := getParamUUID(c, "id")
	if err != nil {
		return httpx.BadRequest(c, "Invalid exchange rate ID", err.Error())
	}

	var req UpdateExchangeRateRequestDTO
	if err := c.Bind().JSON(&req); err != nil {
		return httpx.BadRequest(c, "Invalid request", err.Error())
	}

	rate, err := h.service.UpdateExchangeRate(c, id, req.Rate)
	if err != nil {
		return httpx.FromDomain(c, err, "Error updating exchange rate", "Could not update exchange rate")
	}

	return httpx.OK(c, "Exchange rate updated", "Exchange rate updated successfully", rate)
}

func (h *handler) ImportExchangeRates(c fiber.Ctx) error {
	data, filename, err := readImportFile(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid file", err.Error())
	}

	result, err := h.service.ImportExchangeRatesFromFile(c, data, filename, c.FormValue("sheet"))
	if err != nil {
		return httpx.FromDomain(c, err, "Error importing exchange rates", "Could not import the uploaded exchange rates")
	}

	return httpx.OK(c, "Exchange rates imported", "Spreadsheet imported successfully", result)
}

func (h *handler) SyncSingleExchangeRate(c fiber.Ctx) error {
	id, err := getParamUUID(c, "id")
	if err != nil {
		return httpx.BadRequest(c, "Invalid exchange rate ID", err.Error())
	}

	rate, err := h.service.SyncExchangeRateByID(c, id)
	if err != nil {
		return httpx.FromDomain(c, err, "Exchange rate sync failed", err.Error())
	}

	return httpx.OK(c, "Exchange rate synced", "", rate)
}

// maxImportFileSize bounds uploaded spreadsheets; classic personal trackers
// with a few thousand rows stay well under this.
const maxImportFileSize = 8 << 20 // 8 MiB

func readImportFile(c fiber.Ctx) ([]byte, string, error) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return nil, "", errors.New("missing file: attach the spreadsheet in the \"file\" field")
	}
	if fileHeader.Size > maxImportFileSize {
		return nil, "", errors.New("file too large: the maximum size is 8 MB")
	}
	file, err := fileHeader.Open()
	if err != nil {
		return nil, "", err
	}
	defer func(f multipart.File) { _ = f.Close() }(file)

	data, err := io.ReadAll(io.LimitReader(file, maxImportFileSize+1))
	if err != nil {
		return nil, "", err
	}
	if len(data) > maxImportFileSize {
		return nil, "", errors.New("file too large: the maximum size is 8 MB")
	}
	return data, fileHeader.Filename, nil
}

func getParamUUID(c fiber.Ctx, paramName string) (uuid.UUID, error) {
	return uuid.Parse(c.Params(paramName))
}
