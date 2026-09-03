package portfolio

// Transaction, entry and asset-transaction HTTP handlers. Split out of
// handler.go to keep each file under ~500 lines.

import (
	"strings"

	"uuid"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/paginate"

	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/market"
	"github.com/yeferson59/finexia-app/internal/platform/currency"
	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

func (h *handler) CreatePortfolioEntry(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	req, err := httpx.Bind[CreatePortfolioEntryRequestDTO](c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid request", err.Error())
	}

	txnType := TransactionType(req.TransactionType)
	if req.TransactionType == "" {
		txnType = Buy
	} else if !txnType.IsValid() {
		return httpx.BadRequest(c, "Invalid transaction type", "Type must be one of: buy, sell, dividend, split, transfer_in, transfer_out, fee, interest")
	}

	// No category travels with the request any more: the class of a position is
	// the type of the asset it holds, and the catalogue is what stores that.
	entry, err := h.service.CreatePortfolioEntry(c, userID, req.PortfolioID, req.AssetID, req.SourceID, req.CostCurrency, req.Input(txnType))
	if err != nil {
		return httpx.FromDomain(c, err, "Error creating portfolio entry", "Could not create portfolio entry")
	}

	return httpx.OK(c, "Portfolio entry created", "Portfolio entry created successfully", entry)
}

func (h *handler) UpdateAssetPrice(c fiber.Ctx) error {
	if _, _, _, err := httpx.Identity(c); err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	assetID, err := httpx.ParamUUID(c, "id")
	if err != nil {
		return httpx.BadRequest(c, "Invalid asset ID", err.Error())
	}

	req, err := httpx.Bind[UpdateAssetPriceRequestDTO](c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid request", err.Error())
	}

	asset, err := h.assets.UpdateAssetPrice(c, assetID, req.Price)
	if err != nil {
		return httpx.FromDomain(c, err, "Error updating asset price", "Could not update asset price")
	}

	return httpx.OK(c, "Asset price updated", "Asset price updated successfully", asset)
}

func (h *handler) GetAssetAllocation(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	// Same contract as the summary's ?currency=, because the dashboard shows
	// the two side by side and a donut in dollars beside totals in pesos is a
	// worse answer than either. Omitted falls back to the user's stored
	// preference rather than to no conversion at all: these categories are
	// summed across portfolios that may be denominated differently, so there is
	// no such thing as an unconverted total here.

	var req CurrencyDTO
	if err := c.Bind().Query(&req); err != nil {
		return httpx.BadRequest(c, "Unprocess query currency", "no process currency")
	}

	if req.Currency != money.XXX && !currency.IsSupported(req.Currency) {
		return httpx.BadRequest(c, "Unsupported currency", "currency must be one of: "+currency.List())
	}

	items, err := h.service.GetAssetAllocation(c, userID, req.Currency)
	if err != nil {
		return httpx.FromDomain(c, err, "Error retrieving asset allocation", "Could not retrieve asset allocation")
	}

	return httpx.OK(c, "Asset allocation retrieved", "Asset allocation retrieved successfully", NewAllocationResponse(items))
}

// GetAssetHoldings answers the consolidated view of what the user owns: one row
// per asset, totalled across every portfolio. Same ?currency= contract as the
// allocation — and for the same reason: the rows are summed across portfolios
// that may be denominated differently, so there is no unconverted answer to
// give.
func (h *handler) GetAssetHoldings(c fiber.Ctx) error {
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

	holdings, err := h.service.GetAssetHoldings(c, userID, req.Currency)
	if err != nil {
		return httpx.FromDomain(c, err, "Error retrieving asset holdings", "Could not retrieve asset holdings")
	}

	return httpx.OK(c, "Asset holdings retrieved", "Asset holdings retrieved successfully", NewAssetHoldingsResponse(holdings))
}

func (h *handler) GetUserTransactions(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
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
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	entryID, err := httpx.ParamUUID(c, "entryId")
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
	userID, _, _, err := httpx.Identity(c)
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
	entryID, err := httpx.ParamUUID(c, "entryId")
	if err != nil {
		return httpx.BadRequest(c, "Invalid entry ID", err.Error())
	}

	req, err := httpx.Bind[CreateTransactionRequestDTO](c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid request", err.Error())
	}

	return h.upsertTransaction(c, req.Type, func(userID uuid.UUID, txnType TransactionType) (Transaction, error) {
		return h.service.CreateTransaction(c, userID, entryID, req.Input(txnType))
	}, "Error creating transaction", "Could not create transaction", "Transaction created", "Transaction created successfully")
}

func (h *handler) UpdateTransaction(c fiber.Ctx) error {
	txnID, err := httpx.ParamUUID(c, "txnId")
	if err != nil {
		return httpx.BadRequest(c, "Invalid transaction ID", err.Error())
	}

	req, err := httpx.Bind[UpdateTransactionRequestDTO](c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid request", err.Error())
	}

	return h.upsertTransaction(c, req.Type, func(userID uuid.UUID, txnType TransactionType) (Transaction, error) {
		return h.service.UpdateTransaction(c, userID, txnID, req.Input(txnType))
	}, "Error updating transaction", "Could not update transaction", "Transaction updated", "Transaction updated successfully")
}

func (h *handler) DeleteTransaction(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	txnID, err := httpx.ParamUUID(c, "txnId")
	if err != nil {
		return httpx.BadRequest(c, "Invalid transaction ID", err.Error())
	}

	if err := h.service.DeleteTransaction(c, userID, txnID); err != nil {
		return httpx.FromDomain(c, err, "Error deleting transaction", "Could not delete transaction")
	}

	return httpx.OK(c, "Transaction deleted", "Transaction deleted successfully", nil)
}

func (h *handler) GetAssets(c fiber.Ctx) error {
	paginateInfo, ok := paginate.FromContext(c)
	if !ok {
		return httpx.InternalServerError(c, "", "paginate info not found")
	}

	userID, _, role, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	// Admins see the whole table because they moderate what users contribute;
	// everybody else sees the curated rows plus their own.
	view := market.CatalogView{ViewerID: userID, All: role == httpx.RoleAdmin}

	search := strings.TrimSpace(c.Query("search"))

	var assets []market.Asset
	if search != "" {
		assets, err = h.assets.SearchAssets(c, view, search, uint(paginateInfo.Offset), uint(paginateInfo.Limit))
	} else {
		assets, err = h.assets.GetAssets(c, view, uint(paginateInfo.Offset), uint(paginateInfo.Limit))
	}

	if err != nil {
		return httpx.FromDomain(c, err, "", "")
	}

	return httpx.OK(c, "", "", assets)
}

func (h *handler) GetPortfolioGrowth(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	period := c.Query("period", "ALL")

	// La serie agregada suma portafolios que pueden tener bases distintas, así
	// que necesita una moneda a la que llevarlo todo. Vacío significa la del
	// perfil, igual que en el resumen y en la asignación.
	var req CurrencyDTO
	if err := c.Bind().Query(&req); err != nil {
		return httpx.BadRequest(c, "Unprocess query currency", "no process currency")
	}

	if req.Currency != money.XXX && !currency.IsSupported(req.Currency) {
		return httpx.BadRequest(c, "Unsupported currency", "currency must be one of: "+currency.List())
	}

	points, summary, err := h.service.GetPortfolioGrowth(c, userID, req.Currency, period)
	if err != nil {
		return httpx.FromDomain(c, err, "Error retrieving portfolio growth", "Could not retrieve portfolio growth data")
	}

	return httpx.OK(c, "Portfolio growth retrieved", "Portfolio growth retrieved successfully",
		NewGrowthResponse(points, summary))
}

func (h *handler) GetPortfolioGrowthByID(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	portfolioID, err := httpx.ParamUUID(c, "id")
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
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	portfolioID, err := httpx.ParamUUID(c, "id")
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
