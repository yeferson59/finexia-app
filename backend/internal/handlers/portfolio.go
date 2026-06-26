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

func (h *Handlers) GetPortfoliosSummary(c fiber.Ctx) error {
	userID, _, _, err := h.getUserIDTokenRole(c)
	if err != nil {
		return h.responseBadRequest(c, "Invalid user ID", err.Error())
	}

	summaries, err := h.services.GetPortfoliosSummary(h.ctx, userID)
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

	p, err := h.services.UpdatePlatform(h.ctx, userID, sourceID, req.Name, req.Description, sourceType, req.IsActive)
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

	if err := h.services.DeletePlatform(h.ctx, userID, sourceID); err != nil {
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

	entry, err := h.services.CreatePortfolioEntry(h.ctx, userID, req.PortfolioID, req.AssetID, req.SourceID, txnType, req.Quantity, req.Price, req.CostCurrency, category, req.EntryDate, req.Notes)
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

	asset, err := h.services.UpdateAssetPrice(h.ctx, assetID, req.Price)
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

	items, err := h.services.GetAssetAllocation(h.ctx, userID)
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

	txns, err := h.services.GetRecentUserTransactions(h.ctx, userID, 50)
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

	txns, err := h.services.GetTransactionsByEntry(h.ctx, userID, entryID)
	if err != nil {
		return h.responseFromDomain(c, err, "Error retrieving transactions", "Could not retrieve transactions")
	}

	return h.responseStatusOk(c, "Transactions retrieved", "Transactions retrieved successfully", portfolio.NewTransactionListResponse(txns))
}

func (h *Handlers) CreateTransaction(c fiber.Ctx) error {
	userID, _, _, err := h.getUserIDTokenRole(c)
	if err != nil {
		return h.responseBadRequest(c, "Invalid user ID", err.Error())
	}

	entryID, err := h.getParamUUID(c, "entryId")
	if err != nil {
		return h.responseBadRequest(c, "Invalid entry ID", err.Error())
	}

	var req portfolio.CreateTransactionRequestDTO
	if err := c.Bind().JSON(&req); err != nil {
		return h.responseBadRequest(c, "Invalid request", err.Error())
	}

	txnType := entities.TransactionType(req.Type)
	if !txnType.IsValid() {
		return h.responseBadRequest(c, "Invalid transaction type", "Type must be one of: buy, sell, dividend, split, transfer_in, transfer_out, fee, interest")
	}

	txn, err := h.services.CreateTransaction(h.ctx, userID, entryID, txnType, req.Quantity, req.Price, req.Currency, req.Fees, req.TransactionDate, req.Notes)
	if err != nil {
		return h.responseFromDomain(c, err, "Error creating transaction", "Could not create transaction")
	}

	return h.responseStatusOk(c, "Transaction created", "Transaction created successfully", portfolio.NewTransactionResponse(txn))
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

func (h *Handlers) GetPortfolioGrowth(c fiber.Ctx) error {
	userID, _, _, err := h.getUserIDTokenRole(c)
	if err != nil {
		return h.responseBadRequest(c, "Invalid user ID", err.Error())
	}

	period := c.Query("period", "ALL")

	points, summary, err := h.services.GetPortfolioGrowth(h.ctx, userID, period)
	if err != nil {
		return h.responseFromDomain(c, err, "Error retrieving portfolio growth", "Could not retrieve portfolio growth data")
	}

	return h.responseStatusOk(c, "Portfolio growth retrieved", "Portfolio growth retrieved successfully",
		portfolio.NewGrowthResponse(points, summary))
}
