package handlers

import (
	"strings"

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

	portfolios, err := h.services.GetPortfolios(c, userID)
	if err != nil {
		return h.responseFromDomain(c, err, "Error retrieving portfolios", "Could not retrieve portfolios")
	}

	return h.responseStatusOk(c, "Portfolios retrieved", "Portfolios retrieved successfully", portfolios)
}

func (h *Handlers) GetPortfoliosSummary(c fiber.Ctx) error {
	userID, _, _, err := h.getUserIDTokenRole(c)
	if err != nil {
		return h.responseBadRequest(c, "Invalid user ID", err.Error())
	}

	summaries, err := h.services.GetPortfoliosSummary(c, userID)
	if err != nil {
		return h.responseFromDomain(c, err, "Error retrieving portfolio summaries", "Could not retrieve portfolio summaries")
	}

	return h.responseStatusOk(c, "Portfolio summaries retrieved", "Portfolio summaries retrieved successfully", summaries)
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

	portfolioDetail, err := h.services.GetPortfolio(c, userID, portfolioID)
	if err != nil {
		return h.responseFromDomain(c, err, "Error retrieving portfolio", "Could not retrieve portfolio")
	}

	return h.responseStatusOk(c, "Portfolio retrieved", "Portfolio retrieved successfully", portfolio.NewPortfolioDetailResponse(portfolioDetail))
}

func (h *Handlers) GetPortfoliosRisks(c fiber.Ctx) error {
	risks, err := h.services.GetPortfoliosRisks(c)
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

	portfolio, err := h.services.CreatePortfolio(c, userID, req.Name, req.Description, req.Currency, req.RiskID, portfolioType, req.PriceValue, req.IsDefault)
	if err != nil {
		return h.responseFromDomain(c, err, "Error creating portfolio", "Could not create portfolio")
	}

	return h.responseStatusOk(c, "Portfolio created", "Portfolio created successfully", portfolio)
}

func (h *Handlers) GetPortfolioTopTransaction(c fiber.Ctx) error {
	userID, _, _, err := h.getUserIDTokenRole(c)
	if err != nil {
		return h.responseBadRequest(c, "Invalid user ID", err.Error())
	}

	portfolioID, err := h.getParamUUID(c, "id")
	if err != nil {
		return h.responseBadRequest(c, "Invalid portfolio ID", err.Error())
	}

	dto, err := h.services.GetPortfolioTopTransaction(c, userID, portfolioID)
	if err != nil {
		return h.responseFromDomain(c, err, "Error retrieving top transaction", "Could not retrieve top transaction")
	}

	return h.responseStatusOk(c, "Top transaction retrieved", "Top transaction retrieved successfully", dto)
}

func (h *Handlers) UpdatePortfolio(c fiber.Ctx) error {
	userID, _, _, err := h.getUserIDTokenRole(c)
	if err != nil {
		return h.responseBadRequest(c, "Invalid user ID", err.Error())
	}

	portfolioID, err := h.getParamUUID(c, "id")
	if err != nil {
		return h.responseBadRequest(c, "Invalid portfolio ID", err.Error())
	}

	var req portfolio.UpdatePortfolioRequestDTO
	if err := c.Bind().JSON(&req); err != nil {
		return h.responseBadRequest(c, "Invalid request", err.Error())
	}

	portfolioType := entities.PortfolioType(req.Type)
	if req.Type != "" && !portfolioType.IsValid() {
		return h.responseBadRequest(c, "Invalid portfolio type", "Portfolio type must be one of the supported values")
	}

	riskID, err := uuid.Parse(req.RiskID)
	if err != nil {
		return h.responseBadRequest(c, "Invalid risk ID", err.Error())
	}

	updated, err := h.services.UpdatePortfolio(c, userID, portfolioID, req.Name, req.Description, portfolioType, riskID, req.IsDefault)
	if err != nil {
		return h.responseFromDomain(c, err, "Error updating portfolio", "Could not update portfolio")
	}

	return h.responseStatusOk(c, "Portfolio updated", "Portfolio updated successfully", updated)
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
		return h.responseBadRequest(c, "Invalid source type", "Source type must be one of: broker, investment_bank, trading_platform, neobank, de_fi, crypto_wallet, mutual_funds, brokerage_house, other")
	}

	platform, err := h.services.CreatePlatform(c, userID, sourceType, req.Name, req.Description)
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

	platforms, err := h.services.GetPlatforms(c, userID)
	if err != nil {
		return h.responseFromDomain(c, err, "", "")
	}

	dtos := make([]portfolio.PlatformResponseDTO, 0, len(platforms))
	for _, p := range platforms {
		dtos = append(dtos, portfolio.PlatformResponseDTO{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			SourceType:  string(p.SourceType),
			IsActive:    p.IsActive,
			CreatedAt:   p.CreatedAt,
			UpdatedAt:   p.UpdatedAt,
			Investments: p.Investments,
			TotalValue:  p.TotalValue,
		})
	}

	return h.responseStatusOk(c, "", "", dtos)
}

func (h *Handlers) UpdatePlatform(c fiber.Ctx) error {
	userID, _, _, err := h.getUserIDTokenRole(c)
	if err != nil {
		return h.responseBadRequest(c, "Invalid user ID", err.Error())
	}

	sourceID, err := h.getParamUUID(c, "id")
	if err != nil {
		return h.responseBadRequest(c, "Invalid platform ID", err.Error())
	}

	var req portfolio.UpdatePlatformRequestDTO
	if err := c.Bind().JSON(&req); err != nil {
		return h.responseBadRequest(c, "Invalid request", err.Error())
	}

	sourceType := entities.SourceType(req.Type)
	if req.Type != "" && !sourceType.IsValid() {
		return h.responseBadRequest(c, "Invalid source type", "Source type must be one of: broker, investment_bank, trading_platform, neobank, de_fi, crypto_wallet, mutual_funds, brokerage_house, other")
	}

	p, err := h.services.UpdatePlatform(c, userID, sourceID, req.Name, req.Description, sourceType, req.IsActive)
	if err != nil {
		return h.responseFromDomain(c, err, "Error updating platform", "Could not update platform")
	}

	return h.responseStatusOk(c, "Platform updated", "Platform updated successfully", portfolio.PlatformResponseDTO{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		SourceType:  string(p.SourceType),
		IsActive:    p.IsActive,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
		Investments: p.Investments,
		TotalValue:  p.TotalValue,
	})
}

func (h *Handlers) DeletePlatform(c fiber.Ctx) error {
	userID, _, _, err := h.getUserIDTokenRole(c)
	if err != nil {
		return h.responseBadRequest(c, "Invalid user ID", err.Error())
	}

	sourceID, err := h.getParamUUID(c, "id")
	if err != nil {
		return h.responseBadRequest(c, "Invalid platform ID", err.Error())
	}

	if err := h.services.DeletePlatform(c, userID, sourceID); err != nil {
		return h.responseFromDomain(c, err, "Error deleting platform", "Could not delete platform")
	}

	return h.responseStatusOk(c, "Platform deleted", "Platform deleted successfully", nil)
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

	txnType := entities.TransactionType(req.TransactionType)
	if req.TransactionType == "" {
		txnType = entities.Buy
	} else if !txnType.IsValid() {
		return h.responseBadRequest(c, "Invalid transaction type", "Type must be one of: buy, sell, dividend, split, transfer_in, transfer_out, fee, interest")
	}

	entry, err := h.services.CreatePortfolioEntry(c, userID, req.PortfolioID, req.AssetID, req.SourceID, txnType, req.Quantity, req.Price, req.CostCurrency, category, req.EntryDate, req.Notes)
	if err != nil {
		return h.responseFromDomain(c, err, "Error creating portfolio entry", "Could not create portfolio entry")
	}

	return h.responseStatusOk(c, "Portfolio entry created", "Portfolio entry created successfully", entry)
}

func (h *Handlers) UpdateAssetPrice(c fiber.Ctx) error {
	if _, _, _, err := h.getUserIDTokenRole(c); err != nil {
		return h.responseBadRequest(c, "Invalid user ID", err.Error())
	}

	assetID, err := h.getParamUUID(c, "id")
	if err != nil {
		return h.responseBadRequest(c, "Invalid asset ID", err.Error())
	}

	var req portfolio.UpdateAssetPriceRequestDTO

	if err := c.Bind().JSON(&req); err != nil {
		return h.responseBadRequest(c, "Invalid request", err.Error())
	}

	asset, err := h.services.UpdateAssetPrice(c, assetID, req.Price)
	if err != nil {
		return h.responseFromDomain(c, err, "Error updating asset price", "Could not update asset price")
	}

	return h.responseStatusOk(c, "Asset price updated", "Asset price updated successfully", asset)
}

func (h *Handlers) GetAssetAllocation(c fiber.Ctx) error {
	userID, _, _, err := h.getUserIDTokenRole(c)
	if err != nil {
		return h.responseBadRequest(c, "Invalid user ID", err.Error())
	}

	items, err := h.services.GetAssetAllocation(c, userID)
	if err != nil {
		return h.responseFromDomain(c, err, "Error retrieving asset allocation", "Could not retrieve asset allocation")
	}

	return h.responseStatusOk(c, "Asset allocation retrieved", "Asset allocation retrieved successfully", portfolio.NewAllocationResponse(items))
}

func (h *Handlers) GetUserTransactions(c fiber.Ctx) error {
	userID, _, _, err := h.getUserIDTokenRole(c)
	if err != nil {
		return h.responseBadRequest(c, "Invalid user ID", err.Error())
	}

	txns, err := h.services.GetRecentUserTransactions(c, userID, 50)
	if err != nil {
		return h.responseFromDomain(c, err, "Error retrieving transactions", "Could not retrieve transactions")
	}

	return h.responseStatusOk(c, "Transactions retrieved", "Transactions retrieved successfully", portfolio.NewUserTransactionListResponse(txns))
}

func (h *Handlers) GetTransactions(c fiber.Ctx) error {
	userID, _, _, err := h.getUserIDTokenRole(c)
	if err != nil {
		return h.responseBadRequest(c, "Invalid user ID", err.Error())
	}

	entryID, err := h.getParamUUID(c, "entryId")
	if err != nil {
		return h.responseBadRequest(c, "Invalid entry ID", err.Error())
	}

	txns, err := h.services.GetTransactionsByEntry(c, userID, entryID)
	if err != nil {
		return h.responseFromDomain(c, err, "Error retrieving transactions", "Could not retrieve transactions")
	}

	return h.responseStatusOk(c, "Transactions retrieved", "Transactions retrieved successfully", portfolio.NewTransactionListResponse(txns))
}

// upsertTransaction resolves the caller's user ID, validates the transaction
// type, invokes the caller-supplied service call, and formats the response.
// CreateTransaction and UpdateTransaction only differ in which ID path param
// they read, which request DTO they bind, and which service method they
// call, so those are the only parts left in each of them.
func (h *Handlers) upsertTransaction(c fiber.Ctx, rawType string, call func(userID uuid.UUID, txnType entities.TransactionType) (entities.Transaction, error), failMessage, failDetails, okMessage, okDetails string) error {
	userID, _, _, err := h.getUserIDTokenRole(c)
	if err != nil {
		return h.responseBadRequest(c, "Invalid user ID", err.Error())
	}

	txnType := entities.TransactionType(rawType)
	if !txnType.IsValid() {
		return h.responseBadRequest(c, "Invalid transaction type", "Type must be one of: buy, sell, dividend, split, transfer_in, transfer_out, fee, interest")
	}

	txn, err := call(userID, txnType)
	if err != nil {
		return h.responseFromDomain(c, err, failMessage, failDetails)
	}

	return h.responseStatusOk(c, okMessage, okDetails, portfolio.NewTransactionResponse(txn))
}

func (h *Handlers) CreateTransaction(c fiber.Ctx) error {
	entryID, err := h.getParamUUID(c, "entryId")
	if err != nil {
		return h.responseBadRequest(c, "Invalid entry ID", err.Error())
	}

	var req portfolio.CreateTransactionRequestDTO
	if err := c.Bind().JSON(&req); err != nil {
		return h.responseBadRequest(c, "Invalid request", err.Error())
	}

	return h.upsertTransaction(c, req.Type, func(userID uuid.UUID, txnType entities.TransactionType) (entities.Transaction, error) {
		return h.services.CreateTransaction(c, userID, entryID, txnType, req.Quantity, req.Price, req.Currency, req.Fees, req.TransactionDate, req.Notes)
	}, "Error creating transaction", "Could not create transaction", "Transaction created", "Transaction created successfully")
}

func (h *Handlers) UpdateTransaction(c fiber.Ctx) error {
	txnID, err := h.getParamUUID(c, "txnId")
	if err != nil {
		return h.responseBadRequest(c, "Invalid transaction ID", err.Error())
	}

	var req portfolio.UpdateTransactionRequestDTO
	if err := c.Bind().JSON(&req); err != nil {
		return h.responseBadRequest(c, "Invalid request", err.Error())
	}

	return h.upsertTransaction(c, req.Type, func(userID uuid.UUID, txnType entities.TransactionType) (entities.Transaction, error) {
		return h.services.UpdateTransaction(c, userID, txnID, txnType, req.Quantity, req.Price, req.Currency, req.Fees, req.TransactionDate, req.Notes)
	}, "Error updating transaction", "Could not update transaction", "Transaction updated", "Transaction updated successfully")
}

func (h *Handlers) GetAssets(c fiber.Ctx) error {
	paginateInfo, ok := paginate.FromContext(c)
	if !ok {
		return h.responseInternalServerError(c, "", "paginate info not found")
	}

	search := strings.TrimSpace(c.Query("search"))

	var assets []entities.Asset
	var err error
	if search != "" {
		assets, err = h.services.SearchAssets(c, search, uint(paginateInfo.Offset), uint(paginateInfo.Limit))
	} else {
		assets, err = h.services.GetAssets(c, uint(paginateInfo.Offset), uint(paginateInfo.Limit))
	}

	if err != nil {
		return h.responseFromDomain(c, err, "", "")
	}

	return h.responseStatusOk(c, "", "", assets)
}

func (h *Handlers) GetPortfolioGrowth(c fiber.Ctx) error {
	userID, _, _, err := h.getUserIDTokenRole(c)
	if err != nil {
		return h.responseBadRequest(c, "Invalid user ID", err.Error())
	}

	period := c.Query("period", "ALL")

	points, summary, err := h.services.GetPortfolioGrowth(c, userID, period)
	if err != nil {
		return h.responseFromDomain(c, err, "Error retrieving portfolio growth", "Could not retrieve portfolio growth data")
	}

	return h.responseStatusOk(c, "Portfolio growth retrieved", "Portfolio growth retrieved successfully",
		portfolio.NewGrowthResponse(points, summary))
}

func (h *Handlers) GetPortfolioGrowthByID(c fiber.Ctx) error {
	userID, _, _, err := h.getUserIDTokenRole(c)
	if err != nil {
		return h.responseBadRequest(c, "Invalid user ID", err.Error())
	}

	portfolioID, err := h.getParamUUID(c, "id")
	if err != nil {
		return h.responseBadRequest(c, "Invalid portfolio ID", err.Error())
	}

	period := c.Query("period", "ALL")

	points, summary, err := h.services.GetPortfolioGrowthByID(c, userID, portfolioID, period)
	if err != nil {
		return h.responseFromDomain(c, err, "Error retrieving portfolio growth", "Could not retrieve portfolio growth data")
	}

	return h.responseStatusOk(c, "Portfolio growth retrieved", "Portfolio growth retrieved successfully",
		portfolio.NewGrowthResponse(points, summary))
}

func (h *Handlers) GetAssetTransactions(c fiber.Ctx) error {
	userID, _, _, err := h.getUserIDTokenRole(c)
	if err != nil {
		return h.responseBadRequest(c, "Invalid user ID", err.Error())
	}

	portfolioID, err := h.getParamUUID(c, "id")
	if err != nil {
		return h.responseBadRequest(c, "Invalid portfolio ID", err.Error())
	}

	ticker := strings.TrimSpace(c.Params("symbol"))
	if ticker == "" {
		return h.responseBadRequest(c, "Invalid symbol", "symbol path parameter is required")
	}

	paginateInfo, ok := paginate.FromContext(c)
	if !ok {
		return h.responseInternalServerError(c, "", "paginate info not found")
	}

	page := paginateInfo.Page
	limit := paginateInfo.Limit

	txns, total, err := h.services.GetAssetTransactionsPaginated(c, userID, portfolioID, ticker, page, limit)
	if err != nil {
		return h.responseFromDomain(c, err, "Error retrieving asset transactions", "Could not retrieve asset transactions")
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + limit - 1) / limit
	}

	return h.responseStatusOk(c, "Asset transactions retrieved", "Asset transactions retrieved successfully",
		portfolio.PaginatedTransactionsDTO{
			Data:       portfolio.NewTransactionListResponse(txns),
			Total:      total,
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
		},
	)
}
