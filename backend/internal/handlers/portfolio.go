package handlers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/paginate"
	"github.com/google/uuid"
	"github.com/yeferson59/finexia-app/internal/dtos/portfolio"
	"github.com/yeferson59/finexia-app/internal/entities"
)

func (h *Handlers) GetPortfolios(c fiber.Ctx) error {
	userID, _, _, err := h.getUserIDTokenRole(c)
	if err != nil {
		return h.responseBadRequest(c, "Invalid user ID", err.Error())
	}

	portfolios, err := h.services.GetPortfolios(h.ctx, userID)
	if err != nil {
		return h.responseFromDomain(c, err, "Error retrieving portfolios", "Could not retrieve portfolios")
	}

	return h.responseStatusOk(c, "Portfolios retrieved", "Portfolios retrieved successfully", portfolios)
}

func (h *Handlers) GetPortfolio(c fiber.Ctx) error {
	userID, _, _, err := h.getUserIDTokenRole(c)
	if err != nil {
		return h.responseBadRequest(c, "Invalid user ID", err.Error())
	}

	portfolioID, err := h.getParamUUID(c, "id")
	if err != nil {
		return h.responseBadRequest(c, "Invalid portfolio ID", err.Error())
	}

	portfolioDetail, err := h.services.GetPortfolio(h.ctx, userID, portfolioID)
	if err != nil {
		return h.responseFromDomain(c, err, "Error retrieving portfolio", "Could not retrieve portfolio")
	}

	return h.responseStatusOk(c, "Portfolio retrieved", "Portfolio retrieved successfully", portfolio.NewPortfolioDetailResponse(portfolioDetail))
}

func (h *Handlers) GetPortfoliosRisks(c fiber.Ctx) error {
	risks, err := h.services.GetPortfoliosRisks(h.ctx)
	if err != nil {
		return h.responseFromDomain(c, err, "Error retrieving portfolio risks", "Could not retrieve portfolio risks")
	}

	return h.responseStatusOk(c, "Portfolio risks retrieved", "Portfolio risks retrieved successfully", risks)
}

func (h *Handlers) CreatePortfolio(c fiber.Ctx) error {
	userID, _, _, err := h.getUserIDTokenRole(c)
	if err != nil {
		return h.responseBadRequest(c, "Invalid user ID", err.Error())
	}

	var req portfolio.CreatePortfolioRequestDTO

	if err := c.Bind().JSON(&req); err != nil {
		return h.responseBadRequest(c, "Invalid request", err.Error())
	}

	portfolioType := entities.PortfolioType(req.Type)
	if !portfolioType.IsValid() {
		return h.responseBadRequest(c, "Invalid portfolio type", "Portfolio type must be one of the supported values: stocks, etfs, cryptos, bonds, cash, forex, real_estates, commodities, their combinations or diversified")
	}

	if req.RiskID == uuid.Nil {
		return h.responseBadRequest(c, "Invalid risk", "A valid risk level is required")
	}

	portfolio, err := h.services.CreatePortfolio(h.ctx, userID, req.Name, req.Description, req.Currency, req.RiskID, portfolioType, req.PriceValue, req.IsDefault)
	if err != nil {
		return h.responseFromDomain(c, err, "Error creating portfolio", "Could not create portfolio")
	}

	return h.responseStatusOk(c, "Portfolio created", "Portfolio created successfully", portfolio)
}

func (h *Handlers) CreatePlatform(c fiber.Ctx) error {
	userID, _, _, err := h.getUserIDTokenRole(c)
	if err != nil {
		return h.responseBadRequest(c, "Invalid user ID", err.Error())
	}

	var req portfolio.CreatePlatformRequestDTO

	if err := c.Bind().JSON(&req); err != nil {
		return h.responseBadRequest(c, "Invalid request", err.Error())
	}

	sourceType := entities.SourceType(req.Type)
	if !sourceType.IsValid() {
		return h.responseBadRequest(c, "Invalid source type", "Source type must be one of: broker, bank, tradingPlatform, neobank, defi, cryptoWallet, mutualFunds, brokerageHouse, other")
	}

	platform, err := h.services.CreatePlatform(h.ctx, userID, sourceType, req.Name, req.Description)
	if err != nil {
		return h.responseFromDomain(c, err, "Error creating platform", "Could not create platform")
	}

	return h.responseStatusOk(c, "Platform created", "Platform created successfully", platform)
}

func (h *Handlers) GetPlatforms(c fiber.Ctx) error {
	userID, _, _, err := h.getUserIDTokenRole(c)
	if err != nil {
		return h.responseBadRequest(c, "Invalid user ID", err.Error())
	}

	platforms, err := h.services.GetPlatforms(h.ctx, userID)
	if err != nil {
		return h.responseFromDomain(c, err, "", "")
	}

	return h.responseStatusOk(c, "", "", platforms)
}

func (h *Handlers) CreatePortfolioEntry(c fiber.Ctx) error {
	userID, _, _, err := h.getUserIDTokenRole(c)
	if err != nil {
		return h.responseBadRequest(c, "Invalid user ID", err.Error())
	}

	var req portfolio.CreatePortfolioEntryRequestDTO

	if err := c.Bind().JSON(&req); err != nil {
		return h.responseBadRequest(c, "Invalid request", err.Error())
	}

	AssetType := entities.AssetType(req.Category)
	category := AssetType.Transform()

	if !category.IsValid() {
		return h.responseBadRequest(c, "Invalid category", "Category must be one of: stocks, etf, crypto, bonds, cash, real_estate, commodities, other")
	}

	entry, err := h.services.CreatePortfolioEntry(h.ctx, userID, req.PortfolioID, req.AssetID, req.SourceID, req.Quantity, req.Price, req.CostCurrency, category, req.EntryDate, req.Notes)
	if err != nil {
		return h.responseFromDomain(c, err, "Error creating portfolio entry", "Could not create portfolio entry")
	}

	return h.responseStatusOk(c, "Portfolio entry created", "Portfolio entry created successfully", entry)
}

func (h *Handlers) GetAssets(c fiber.Ctx) error {
	paginateInfo, ok := paginate.FromContext(c)
	if !ok {
		return h.responseInternalServerError(c, "", "paginate info not found")
	}

	assests, err := h.services.GetAssets(h.ctx, uint(paginateInfo.Offset), uint(paginateInfo.Limit))
	if err != nil {
		return h.responseFromDomain(c, err, "", "")
	}

	return h.responseStatusOk(c, "", "", assests)
}
