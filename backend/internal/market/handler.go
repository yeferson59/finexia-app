package market

import (
	"errors"
	"io"
	"mime/multipart"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/paginate"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

type handler struct {
	service *Service
	// holdings answers "which assets does this user own", supplied by the
	// composition root from the portfolio module. market must not import
	// portfolio, so it is consumed through the interface declared here.
	holdings Holdings
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

// CreateAsset adds an asset to the catalog, curating it when an admin asks and
// contributing it when anybody else does.
//
// One route, two behaviours, instead of two routes: from the client's side this
// is the same request — "put this instrument in the catalog" — and the caller
// does not get to choose which of the two it becomes.
func (h *handler) CreateAsset(c fiber.Ctx) error {
	userID, _, role, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid identity", err.Error())
	}

	var req CreateAssetRequestDTO
	if err := c.Bind().JSON(&req); err != nil {
		return httpx.BadRequest(c, "Invalid request", err.Error())
	}

	assetType := AssetType(req.AssetType)

	var asset Asset
	if role == httpx.RoleAdmin {
		asset, err = h.service.CreateAsset(c, req.Ticker, req.Name, assetType, req.Exchange, req.Currency)
	} else {
		asset, err = h.service.ContributeAsset(c, userID, req.Ticker, req.Name, assetType, req.Exchange, req.Currency)
	}

	if err != nil {
		return httpx.FromDomain(c, err, "Error creating asset", assetFailureDetail(err))
	}

	return httpx.Success(c, fiber.StatusCreated, "Asset created", "Asset created successfully", asset)
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
	id, err := httpx.ParamUUID(c, "id")
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
