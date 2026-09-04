package portfolio

import (
	"uuid"

	"github.com/gofiber/fiber/v3"

	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/platform/currency"
	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

type handler struct {
	service *service
	assets  AssetReader
}

func newHandler(svc *service, assets AssetReader) *handler {
	return new(handler{svc, assets})
}

func (h *handler) GetPortfolios(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	portfolios, err := h.service.GetPortfolios(c, userID)
	if err != nil {
		return httpx.FromDomain(c, err, "Error retrieving portfolios", "Could not retrieve portfolios")
	}

	return httpx.OK(c, "Portfolios retrieved", "Portfolios retrieved successfully", portfolios)
}

func (h *handler) GetPortfoliosSummary(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	var req CurrencyDTO
	if err := c.Bind().Query(&req); err != nil {
		return httpx.BadRequest(c, "Unprocess query currency", "no process currency")
	}

	if req.Currency != money.XXX && !currency.IsSupported(req.Currency) {
		return httpx.BadRequest(c, "Unsupported currency", "currency must be one of: "+currency.List())
	}

	var summaries []SummaryView
	if req.Currency == money.XXX {
		summaries, err = h.service.GetPortfoliosSummary(c, userID)
	} else {
		summaries, err = h.service.GetPortfoliosSummaryInCurrency(c, userID, req.Currency)
	}
	if err != nil {
		return httpx.FromDomain(c, err, "Error retrieving portfolio summaries", "Could not retrieve portfolio summaries")
	}

	return httpx.OK(c, "Portfolio summaries retrieved", "Portfolio summaries retrieved successfully", summaries)
}

func (h *handler) GetPortfolio(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	portfolioID, err := httpx.ParamUUID(c, "id")
	if err != nil {
		return httpx.BadRequest(c, "Invalid portfolio ID", err.Error())
	}

	portfolioDetail, err := h.service.GetPortfolio(c, userID, portfolioID)
	if err != nil {
		return httpx.FromDomain(c, err, "Error retrieving portfolio", "Could not retrieve portfolio")
	}

	return httpx.OK(c, "Portfolio retrieved", "Portfolio retrieved successfully", NewPortfolioDetailResponse(portfolioDetail))
}

func (h *handler) GetPortfoliosRisks(c fiber.Ctx) error {
	risks, err := h.service.GetPortfoliosRisks(c)
	if err != nil {
		return httpx.FromDomain(c, err, "Error retrieving portfolio risks", "Could not retrieve portfolio risks")
	}

	return httpx.OK(c, "Portfolio risks retrieved", "Portfolio risks retrieved successfully", risks)
}

func (h *handler) CreatePortfolio(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	req, err := httpx.Bind[CreatePortfolioRequestDTO](c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid request", err.Error())
	}

	portfolioType := Type(req.Type)
	if !portfolioType.IsValid() {
		return httpx.BadRequest(c, "Invalid portfolio type", "Portfolio type must be one of the supported values: stocks, etfs, cryptos, bonds, cash, forex, real_estates, commodities, their combinations or diversified")
	}

	if req.RiskID == uuid.Nil() {
		return httpx.BadRequest(c, "Invalid risk", "A valid risk level is required")
	}

	created, err := h.service.CreatePortfolio(c, userID, req.Name, req.Description, req.Currency, req.RiskID, portfolioType, req.PriceValue, req.IsDefault)
	if err != nil {
		return httpx.FromDomain(c, err, "Error creating portfolio", "Could not create portfolio")
	}

	return httpx.OK(c, "Portfolio created", "Portfolio created successfully", created)
}

func (h *handler) GetPortfolioTopTransaction(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	portfolioID, err := httpx.ParamUUID(c, "id")
	if err != nil {
		return httpx.BadRequest(c, "Invalid portfolio ID", err.Error())
	}

	dto, err := h.service.GetPortfolioTopTransaction(c, userID, portfolioID)
	if err != nil {
		return httpx.FromDomain(c, err, "Error retrieving top transaction", "Could not retrieve top transaction")
	}

	return httpx.OK(c, "Top transaction retrieved", "Top transaction retrieved successfully", dto)
}

func (h *handler) UpdatePortfolio(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	portfolioID, err := httpx.ParamUUID(c, "id")
	if err != nil {
		return httpx.BadRequest(c, "Invalid portfolio ID", err.Error())
	}

	req, err := httpx.Bind[UpdatePortfolioRequestDTO](c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid request", err.Error())
	}

	portfolioType := Type(req.Type)
	if req.Type != "" && !portfolioType.IsValid() {
		return httpx.BadRequest(c, "Invalid portfolio type", "Portfolio type must be one of the supported values")
	}

	riskID, err := uuid.Parse(req.RiskID)
	if err != nil {
		return httpx.BadRequest(c, "Invalid risk ID", err.Error())
	}

	updated, err := h.service.UpdatePortfolio(c, userID, portfolioID, req.Name, req.Description, portfolioType, riskID, req.IsDefault)
	if err != nil {
		return httpx.FromDomain(c, err, "Error updating portfolio", "Could not update portfolio")
	}

	return httpx.OK(c, "Portfolio updated", "Portfolio updated successfully", updated)
}

func (h *handler) CreatePlatform(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	req, err := httpx.Bind[CreatePlatformRequestDTO](c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid request", err.Error())
	}

	sourceType := SourceType(req.Type)
	if !sourceType.IsValid() {
		return httpx.BadRequest(c, "Invalid source type", "Source type must be one of: broker, investment_bank, trading_platform, neobank, de_fi, crypto_wallet, mutual_funds, brokerage_house, other")
	}

	platform, err := h.service.CreatePlatform(c, userID, sourceType, req.Name, req.Description)
	if err != nil {
		return httpx.FromDomain(c, err, "Error creating platform", "Could not create platform")
	}

	return httpx.OK(c, "Platform created", "Platform created successfully", platform)
}

func (h *handler) GetPlatforms(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	// Same contract as the summary's ?currency=: the platform totals are a sum
	// over entries in mixed currencies, so the caller has to be able to name
	// the one they want them in.
	var req CurrencyDTO
	if err := c.Bind().Query(&req); err != nil {
		return httpx.BadRequest(c, "Unprocess query currency", "no process currency")
	}

	if req.Currency != money.XXX && !currency.IsSupported(req.Currency) {
		return httpx.BadRequest(c, "Unsupported currency", "currency must be one of: "+currency.List())
	}

	platforms, err := h.service.GetPlatforms(c, userID, req.Currency)
	if err != nil {
		return httpx.FromDomain(c, err, "", "")
	}

	return httpx.OK(c, "", "", NewPlatformListResponse(platforms))
}

func (h *handler) UpdatePlatform(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	sourceID, err := httpx.ParamUUID(c, "id")
	if err != nil {
		return httpx.BadRequest(c, "Invalid platform ID", err.Error())
	}

	req, err := httpx.Bind[UpdatePlatformRequestDTO](c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid request", err.Error())
	}

	sourceType := SourceType(req.Type)
	if req.Type != "" && !sourceType.IsValid() {
		return httpx.BadRequest(c, "Invalid source type", "Source type must be one of: broker, investment_bank, trading_platform, neobank, de_fi, crypto_wallet, mutual_funds, brokerage_house, other")
	}

	p, err := h.service.UpdatePlatform(c, userID, sourceID, req.Name, req.Description, sourceType, req.IsActive)
	if err != nil {
		return httpx.FromDomain(c, err, "Error updating platform", "Could not update platform")
	}

	return httpx.OK(c, "Platform updated", "Platform updated successfully", newPlatformResponse(p))
}

func (h *handler) DeletePlatform(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	sourceID, err := httpx.ParamUUID(c, "id")
	if err != nil {
		return httpx.BadRequest(c, "Invalid platform ID", err.Error())
	}

	if err := h.service.DeletePlatform(c, userID, sourceID); err != nil {
		return httpx.FromDomain(c, err, "Error deleting platform", "Could not delete platform")
	}

	return httpx.OK(c, "Platform deleted", "Platform deleted successfully", nil)
}
