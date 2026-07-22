package portfolio

// Transaction, entry and asset-transaction HTTP handlers. Split out of
// handler.go to keep each file under ~500 lines.

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/paginate"
	"github.com/google/uuid"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

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
