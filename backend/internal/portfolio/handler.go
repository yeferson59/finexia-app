package portfolio

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/paginate"
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

func (h *handler) CreatePortfolioEntry(c fiber.Ctx) error {
	userID, _, _, err := getUserIDTokenRole(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	var req CreatePortfolioEntryRequestDTO

	if err := c.Bind().JSON(&req); err != nil {
		return httpx.BadRequest(c, "Invalid request", err.Error())
	}

	assetType := AssetType(req.Category)
	category := assetType.Transform()

	if !category.IsValid() {
		return httpx.BadRequest(c, "Invalid category", "Category must be one of: stocks, etf, crypto, bonds, cash, real_estate, commodities, other")
	}

	txnType := TransactionType(req.TransactionType)
	if req.TransactionType == "" {
		txnType = Buy
	} else if !txnType.IsValid() {
		return httpx.BadRequest(c, "Invalid transaction type", "Type must be one of: buy, sell, dividend, split, transfer_in, transfer_out, fee, interest")
	}

	entry, err := h.service.CreatePortfolioEntry(c, userID, req.PortfolioID, req.AssetID, req.SourceID, txnType, req.Quantity, req.Price, req.CostCurrency, category, req.EntryDate, req.Notes)
	if err != nil {
		return httpx.FromDomain(c, err, "Error creating portfolio entry", "Could not create portfolio entry")
	}

	return httpx.OK(c, "Portfolio entry created", "Portfolio entry created successfully", entry)
}

func (h *handler) UpdateAssetPrice(c fiber.Ctx) error {
	if _, _, _, err := getUserIDTokenRole(c); err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	assetID, err := getParamUUID(c, "id")
	if err != nil {
		return httpx.BadRequest(c, "Invalid asset ID", err.Error())
	}

	var req UpdateAssetPriceRequestDTO

	if err := c.Bind().JSON(&req); err != nil {
		return httpx.BadRequest(c, "Invalid request", err.Error())
	}

	asset, err := h.service.UpdateAssetPrice(c, assetID, req.Price)
	if err != nil {
		return httpx.FromDomain(c, err, "Error updating asset price", "Could not update asset price")
	}

	return httpx.OK(c, "Asset price updated", "Asset price updated successfully", asset)
}

func (h *handler) GetAssetAllocation(c fiber.Ctx) error {
	userID, _, _, err := getUserIDTokenRole(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	items, err := h.service.GetAssetAllocation(c, userID)
	if err != nil {
		return httpx.FromDomain(c, err, "Error retrieving asset allocation", "Could not retrieve asset allocation")
	}

	return httpx.OK(c, "Asset allocation retrieved", "Asset allocation retrieved successfully", NewAllocationResponse(items))
}

func (h *handler) GetUserTransactions(c fiber.Ctx) error {
	userID, _, _, err := getUserIDTokenRole(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	txns, err := h.service.GetRecentUserTransactions(c, userID, 50)
	if err != nil {
		return httpx.FromDomain(c, err, "Error retrieving transactions", "Could not retrieve transactions")
	}

	return httpx.OK(c, "Transactions retrieved", "Transactions retrieved successfully", NewUserTransactionListResponse(txns))
}

func (h *handler) GetTransactions(c fiber.Ctx) error {
	userID, _, _, err := getUserIDTokenRole(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	entryID, err := getParamUUID(c, "entryId")
	if err != nil {
		return httpx.BadRequest(c, "Invalid entry ID", err.Error())
	}

	txns, err := h.service.GetTransactionsByEntry(c, userID, entryID)
	if err != nil {
		return httpx.FromDomain(c, err, "Error retrieving transactions", "Could not retrieve transactions")
	}

	return httpx.OK(c, "Transactions retrieved", "Transactions retrieved successfully", NewTransactionListResponse(txns))
}

// upsertTransaction resolves the caller's user ID, validates the transaction
// type, invokes the caller-supplied service call, and formats the response.
// CreateTransaction and UpdateTransaction only differ in which ID path param
// they read, which request DTO they bind, and which service method they
// call, so those are the only parts left in each of them.
func (h *handler) upsertTransaction(c fiber.Ctx, rawType string, call func(userID uuid.UUID, txnType TransactionType) (Transaction, error), failMessage, failDetails, okMessage, okDetails string) error {
	userID, _, _, err := getUserIDTokenRole(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	txnType := TransactionType(rawType)
	if !txnType.IsValid() {
		return httpx.BadRequest(c, "Invalid transaction type", "Type must be one of: buy, sell, dividend, split, transfer_in, transfer_out, fee, interest")
	}

	txn, err := call(userID, txnType)
	if err != nil {
		return httpx.FromDomain(c, err, failMessage, failDetails)
	}

	return httpx.OK(c, okMessage, okDetails, NewTransactionResponse(txn))
}

func (h *handler) CreateTransaction(c fiber.Ctx) error {
	entryID, err := getParamUUID(c, "entryId")
	if err != nil {
		return httpx.BadRequest(c, "Invalid entry ID", err.Error())
	}

	var req CreateTransactionRequestDTO
	if err := c.Bind().JSON(&req); err != nil {
		return httpx.BadRequest(c, "Invalid request", err.Error())
	}

	return h.upsertTransaction(c, req.Type, func(userID uuid.UUID, txnType TransactionType) (Transaction, error) {
		return h.service.CreateTransaction(c, userID, entryID, txnType, req.Quantity, req.Price, req.Currency, req.Fees, req.TransactionDate, req.Notes)
	}, "Error creating transaction", "Could not create transaction", "Transaction created", "Transaction created successfully")
}

func (h *handler) UpdateTransaction(c fiber.Ctx) error {
	txnID, err := getParamUUID(c, "txnId")
	if err != nil {
		return httpx.BadRequest(c, "Invalid transaction ID", err.Error())
	}

	var req UpdateTransactionRequestDTO
	if err := c.Bind().JSON(&req); err != nil {
		return httpx.BadRequest(c, "Invalid request", err.Error())
	}

	return h.upsertTransaction(c, req.Type, func(userID uuid.UUID, txnType TransactionType) (Transaction, error) {
		return h.service.UpdateTransaction(c, userID, txnID, txnType, req.Quantity, req.Price, req.Currency, req.Fees, req.TransactionDate, req.Notes)
	}, "Error updating transaction", "Could not update transaction", "Transaction updated", "Transaction updated successfully")
}

func (h *handler) GetAssets(c fiber.Ctx) error {
	paginateInfo, ok := paginate.FromContext(c)
	if !ok {
		return httpx.InternalServerError(c, "", "paginate info not found")
	}

	search := strings.TrimSpace(c.Query("search"))

	var assets []Asset
	var err error
	if search != "" {
		assets, err = h.service.SearchAssets(c, search, uint(paginateInfo.Offset), uint(paginateInfo.Limit))
	} else {
		assets, err = h.service.GetAssets(c, uint(paginateInfo.Offset), uint(paginateInfo.Limit))
	}

	if err != nil {
		return httpx.FromDomain(c, err, "", "")
	}

	return httpx.OK(c, "", "", assets)
}

func (h *handler) GetPortfolioGrowth(c fiber.Ctx) error {
	userID, _, _, err := getUserIDTokenRole(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	period := c.Query("period", "ALL")

	points, summary, err := h.service.GetPortfolioGrowth(c, userID, period)
	if err != nil {
		return httpx.FromDomain(c, err, "Error retrieving portfolio growth", "Could not retrieve portfolio growth data")
	}

	return httpx.OK(c, "Portfolio growth retrieved", "Portfolio growth retrieved successfully",
		NewGrowthResponse(points, summary))
}

func (h *handler) GetPortfolioGrowthByID(c fiber.Ctx) error {
	userID, _, _, err := getUserIDTokenRole(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	portfolioID, err := getParamUUID(c, "id")
	if err != nil {
		return httpx.BadRequest(c, "Invalid portfolio ID", err.Error())
	}

	period := c.Query("period", "ALL")

	points, summary, err := h.service.GetPortfolioGrowthByID(c, userID, portfolioID, period)
	if err != nil {
		return httpx.FromDomain(c, err, "Error retrieving portfolio growth", "Could not retrieve portfolio growth data")
	}

	return httpx.OK(c, "Portfolio growth retrieved", "Portfolio growth retrieved successfully",
		NewGrowthResponse(points, summary))
}

func (h *handler) GetAssetTransactions(c fiber.Ctx) error {
	userID, _, _, err := getUserIDTokenRole(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	portfolioID, err := getParamUUID(c, "id")
	if err != nil {
		return httpx.BadRequest(c, "Invalid portfolio ID", err.Error())
	}

	ticker := strings.TrimSpace(c.Params("symbol"))
	if ticker == "" {
		return httpx.BadRequest(c, "Invalid symbol", "symbol path parameter is required")
	}

	paginateInfo, ok := paginate.FromContext(c)
	if !ok {
		return httpx.InternalServerError(c, "", "paginate info not found")
	}

	page := paginateInfo.Page
	limit := paginateInfo.Limit

	txns, total, err := h.service.GetAssetTransactionsPaginated(c, userID, portfolioID, ticker, page, limit)
	if err != nil {
		return httpx.FromDomain(c, err, "Error retrieving asset transactions", "Could not retrieve asset transactions")
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + limit - 1) / limit
	}

	return httpx.OK(c, "Asset transactions retrieved", "Asset transactions retrieved successfully",
		PaginatedTransactionsDTO{
			Data:       NewTransactionListResponse(txns),
			Total:      total,
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
		},
	)
}
