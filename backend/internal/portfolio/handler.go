package portfolio

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

type handler struct {
	service *Service
}

const (
	LocalUserID = "auth_user_id"
	LocalToken  = "auth_token"
	LocalRole   = "auth_role"
)

// getUserIDTokenRole extracts the authenticated identity the JWT middleware
// stored in the request locals.
func getUserIDTokenRole(c fiber.Ctx) (uuid.UUID, string, string, error) {
	userIDStr, _ := c.Locals(LocalUserID).(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, "", "", err
	}

	token, _ := c.Locals(LocalToken).(string)
	role, _ := c.Locals(LocalRole).(string)
	if token == "" || role == "" {
		return uuid.Nil, "", "", errors.New("missing authenticated identity")
	}

	return userID, token, role, nil
}

func getParamUUID(c fiber.Ctx, paramName string) (uuid.UUID, error) {
	return uuid.Parse(c.Params(paramName))
}

func (h *handler) GetPortfolios(c fiber.Ctx) error {
	userID, _, _, err := getUserIDTokenRole(c)
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
	userID, _, _, err := getUserIDTokenRole(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	currency := strings.ToUpper(strings.TrimSpace(c.Query("currency")))
	if currency != "" && !IsSupportedDisplayCurrency(currency) {
		return httpx.BadRequest(c, "Unsupported currency", "currency must be one of: "+strings.Join(SupportedDisplayCurrencies, ", "))
	}

	var summaries []SummaryView
	if currency == "" {
		summaries, err = h.service.GetPortfoliosSummary(c, userID)
	} else {
		summaries, err = h.service.GetPortfoliosSummaryInCurrency(c, userID, currency)
	}
	if err != nil {
		return httpx.FromDomain(c, err, "Error retrieving portfolio summaries", "Could not retrieve portfolio summaries")
	}

	return httpx.OK(c, "Portfolio summaries retrieved", "Portfolio summaries retrieved successfully", summaries)
}

func (h *handler) GetPortfolio(c fiber.Ctx) error {
	userID, _, _, err := getUserIDTokenRole(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	portfolioID, err := getParamUUID(c, "id")
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
	userID, _, _, err := getUserIDTokenRole(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	var req CreatePortfolioRequestDTO

	if err := c.Bind().JSON(&req); err != nil {
		return httpx.BadRequest(c, "Invalid request", err.Error())
	}

	portfolioType := Type(req.Type)
	if !portfolioType.IsValid() {
		return httpx.BadRequest(c, "Invalid portfolio type", "Portfolio type must be one of the supported values: stocks, etfs, cryptos, bonds, cash, forex, real_estates, commodities, their combinations or diversified")
	}

	if req.RiskID == uuid.Nil {
		return httpx.BadRequest(c, "Invalid risk", "A valid risk level is required")
	}

	created, err := h.service.CreatePortfolio(c, userID, req.Name, req.Description, req.Currency, req.RiskID, portfolioType, req.PriceValue, req.IsDefault)
	if err != nil {
		return httpx.FromDomain(c, err, "Error creating portfolio", "Could not create portfolio")
	}

	return httpx.OK(c, "Portfolio created", "Portfolio created successfully", created)
}

func (h *handler) GetPortfolioTopTransaction(c fiber.Ctx) error {
	userID, _, _, err := getUserIDTokenRole(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	portfolioID, err := getParamUUID(c, "id")
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
	userID, _, _, err := getUserIDTokenRole(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	portfolioID, err := getParamUUID(c, "id")
	if err != nil {
		return httpx.BadRequest(c, "Invalid portfolio ID", err.Error())
	}

	var req UpdatePortfolioRequestDTO
	if err := c.Bind().JSON(&req); err != nil {
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
	userID, _, _, err := getUserIDTokenRole(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	var req CreatePlatformRequestDTO

	if err := c.Bind().JSON(&req); err != nil {
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
	userID, _, _, err := getUserIDTokenRole(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	platforms, err := h.service.GetPlatforms(c, userID)
	if err != nil {
		return httpx.FromDomain(c, err, "", "")
	}

	dtos := make([]PlatformResponseDTO, 0, len(platforms))
	for _, p := range platforms {
		dtos = append(dtos, newPlatformResponse(p))
	}

	return httpx.OK(c, "", "", dtos)
}

func (h *handler) UpdatePlatform(c fiber.Ctx) error {
	userID, _, _, err := getUserIDTokenRole(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	sourceID, err := getParamUUID(c, "id")
	if err != nil {
		return httpx.BadRequest(c, "Invalid platform ID", err.Error())
	}

	var req UpdatePlatformRequestDTO
	if err := c.Bind().JSON(&req); err != nil {
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
	userID, _, _, err := getUserIDTokenRole(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	sourceID, err := getParamUUID(c, "id")
	if err != nil {
		return httpx.BadRequest(c, "Invalid platform ID", err.Error())
	}

	if err := h.service.DeletePlatform(c, userID, sourceID); err != nil {
		return httpx.FromDomain(c, err, "Error deleting platform", "Could not delete platform")
	}

	return httpx.OK(c, "Platform deleted", "Platform deleted successfully", nil)
}
