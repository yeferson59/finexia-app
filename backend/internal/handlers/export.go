package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/xuri/excelize/v2"

	"github.com/yeferson59/finexia-app/internal/dtos/portfolio"
)

func (h *Handlers) ExportSummary(c fiber.Ctx) error {
	userID, _, _, err := h.getUserIDTokenRole(c)
	if err != nil {
		return h.responseBadRequest(c, "Invalid user ID", err.Error())
	}

	summaries, err := h.services.GetPortfoliosSummary(h.ctx, userID)
	if err != nil {
		return h.responseFromDomain(c, err, "Error generating report", "Could not retrieve portfolio data")
	}

	allocationItems, err := h.services.GetAssetAllocation(h.ctx, userID)
	if err != nil {
		return h.responseFromDomain(c, err, "Error generating report", "Could not retrieve allocation data")
	}
	allocation := portfolio.NewAllocationResponse(allocationItems)

	f := excelize.NewFile()
	defer f.Close()

	// Sheet 1: Portafolios
	portfolioSheet := "Portafolios"
	f.SetSheetName("Sheet1", portfolioSheet)
	portfolioHeaders := []string{
		"Nombre", "Tipo", "Moneda", "Riesgo",
		"Posiciones", "Costo Base", "Valor de Mercado", "Ganancia/Pérdida", "Rentabilidad %",
	}
	for i, header := range portfolioHeaders {
		col, _ := excelize.ColumnNumberToName(i + 1)
		f.SetCellValue(portfolioSheet, col+"1", header)
	}
	for i, s := range summaries {
		row := strconv.Itoa(i + 2)
		f.SetCellValue(portfolioSheet, "A"+row, s.Name)
		f.SetCellValue(portfolioSheet, "B"+row, string(s.Type))
		f.SetCellValue(portfolioSheet, "C"+row, s.BaseCurrency)
		f.SetCellValue(portfolioSheet, "D"+row, s.RiskName)
		f.SetCellValue(portfolioSheet, "E"+row, s.TotalPositions)
		f.SetCellValue(portfolioSheet, "F"+row, s.TotalCostBase)
		f.SetCellValue(portfolioSheet, "G"+row, s.TotalMarketValue)
		f.SetCellValue(portfolioSheet, "H"+row, s.TotalGainLoss)
		f.SetCellValue(portfolioSheet, "I"+row, s.TotalGainLossPct)
	}

	// Sheet 2: Asignación
	allocSheet := "Asignación"
	f.NewSheet(allocSheet)
	allocHeaders := []string{"Categoría", "Valor de Mercado", "Porcentaje"}
	for i, header := range allocHeaders {
		col, _ := excelize.ColumnNumberToName(i + 1)
		f.SetCellValue(allocSheet, col+"1", header)
	}
	for i, a := range allocation {
		row := strconv.Itoa(i + 2)
		f.SetCellValue(allocSheet, "A"+row, a.Category)
		f.SetCellValue(allocSheet, "B"+row, a.MarketValue)
		f.SetCellValue(allocSheet, "C"+row, a.Percent)
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return h.responseInternalServerError(c, "Error serializing report", err.Error())
	}

	c.Set(fiber.HeaderContentType, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set(fiber.HeaderContentDisposition, `attachment; filename="resumen-mensual.xlsx"`)
	return c.Send(buf.Bytes())
}

func (h *Handlers) ExportTransactions(c fiber.Ctx) error {
	userID, _, _, err := h.getUserIDTokenRole(c)
	if err != nil {
		return h.responseBadRequest(c, "Invalid user ID", err.Error())
	}

	txns, err := h.services.GetRecentUserTransactions(h.ctx, userID, 10000)
	if err != nil {
		return h.responseFromDomain(c, err, "Error generating report", "Could not retrieve transactions")
	}
	dtos := portfolio.NewUserTransactionListResponse(txns)

	f := excelize.NewFile()
	defer f.Close()

	sheet := "Transacciones"
	f.SetSheetName("Sheet1", sheet)
	headers := []string{"Fecha", "Tipo", "Activo", "Ticker", "Cantidad", "Precio", "Comisiones", "Moneda", "Notas"}
	for i, header := range headers {
		col, _ := excelize.ColumnNumberToName(i + 1)
		f.SetCellValue(sheet, col+"1", header)
	}
	for i, t := range dtos {
		row := strconv.Itoa(i + 2)
		f.SetCellValue(sheet, "A"+row, t.TransactionDate.Format("2006-01-02"))
		f.SetCellValue(sheet, "B"+row, t.Type)
		f.SetCellValue(sheet, "C"+row, t.AssetName)
		f.SetCellValue(sheet, "D"+row, t.AssetTicker)
		f.SetCellValue(sheet, "E"+row, t.Quantity)
		f.SetCellValue(sheet, "F"+row, t.Price)
		f.SetCellValue(sheet, "G"+row, t.Fees)
		f.SetCellValue(sheet, "H"+row, t.Currency)
		f.SetCellValue(sheet, "I"+row, t.Notes)
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return h.responseInternalServerError(c, "Error serializing report", err.Error())
	}

	c.Set(fiber.HeaderContentType, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set(fiber.HeaderContentDisposition, `attachment; filename="transacciones.xlsx"`)
	return c.Send(buf.Bytes())
}

func (h *Handlers) ExportRiskMetrics(c fiber.Ctx) error {
	userID, _, _, err := h.getUserIDTokenRole(c)
	if err != nil {
		return h.responseBadRequest(c, "Invalid user ID", err.Error())
	}

	points, _, err := h.services.GetPortfolioGrowth(h.ctx, userID, "ALL")
	if err != nil {
		return h.responseFromDomain(c, err, "Error generating report", "Could not retrieve growth data")
	}

	f := excelize.NewFile()
	defer f.Close()

	sheet := "Historial de crecimiento"
	f.SetSheetName("Sheet1", sheet)
	headers := []string{"Fecha", "Valor Total", "Costo Base", "Ganancia/Pérdida", "Rentabilidad %"}
	for i, header := range headers {
		col, _ := excelize.ColumnNumberToName(i + 1)
		f.SetCellValue(sheet, col+"1", header)
	}
	for i, p := range points {
		row := strconv.Itoa(i + 2)
		f.SetCellValue(sheet, "A"+row, p.Date.Format("2006-01-02"))
		f.SetCellValue(sheet, "B"+row, p.TotalValue)
		f.SetCellValue(sheet, "C"+row, p.TotalCostBase)
		f.SetCellValue(sheet, "D"+row, p.GainLoss)
		f.SetCellValue(sheet, "E"+row, p.GainLossPct)
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return h.responseInternalServerError(c, "Error serializing report", err.Error())
	}

	c.Set(fiber.HeaderContentType, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set(fiber.HeaderContentDisposition, `attachment; filename="riesgo-volatilidad.xlsx"`)
	return c.Send(buf.Bytes())
}
