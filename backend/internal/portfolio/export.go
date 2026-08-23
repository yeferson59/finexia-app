package portfolio

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/xuri/excelize/v2"
	"github.com/yeferson59/gofinance/v2/decimal"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

func (h *handler) ExportSummary(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	summaries, err := h.service.GetPortfoliosSummary(c, userID)
	if err != nil {
		return httpx.FromDomain(c, err, "Error generating report", "Could not retrieve portfolio data")
	}

	// No currency asked for: the spreadsheet reports in the user's own
	// preference, the same default the allocation endpoint applies.
	allocationItems, err := h.service.GetAssetAllocation(c, userID, "")
	if err != nil {
		return httpx.FromDomain(c, err, "Error generating report", "Could not retrieve allocation data")
	}
	allocation := NewAllocationResponse(allocationItems)

	f := excelize.NewFile()
	defer func() { _ = f.Close() }()

	// Sheet 1: Portafolios
	portfolioSheet := "Portafolios"
	_ = f.SetSheetName("Sheet1", portfolioSheet)
	portfolioHeaders := []string{
		"Nombre", "Tipo", "Moneda", "Riesgo",
		"Posiciones", "Costo Base", "Valor de Mercado", "Ganancia/Pérdida", "Rentabilidad %",
	}
	for i, header := range portfolioHeaders {
		col, _ := excelize.ColumnNumberToName(i + 1)
		_ = f.SetCellValue(portfolioSheet, col+"1", header)
	}
	for i, s := range summaries {
		row := strconv.Itoa(i + 2)
		_ = f.SetCellValue(portfolioSheet, "A"+row, s.Name)
		_ = f.SetCellValue(portfolioSheet, "B"+row, string(s.Type))
		_ = f.SetCellValue(portfolioSheet, "C"+row, s.BaseCurrency)
		_ = f.SetCellValue(portfolioSheet, "D"+row, s.RiskName)
		_ = f.SetCellValue(portfolioSheet, "E"+row, s.TotalPositions)
		_ = f.SetCellValue(portfolioSheet, "F"+row, s.TotalCostBase)
		_ = f.SetCellValue(portfolioSheet, "G"+row, s.TotalMarketValue)
		_ = f.SetCellValue(portfolioSheet, "H"+row, s.TotalGainLoss)
		_ = f.SetCellValue(portfolioSheet, "I"+row, s.TotalGainLossPct)
	}

	// Sheet 2: Asignación
	allocSheet := "Asignación"
	_, _ = f.NewSheet(allocSheet)
	allocHeaders := []string{"Categoría", "Valor de Mercado", "Porcentaje"}
	for i, header := range allocHeaders {
		col, _ := excelize.ColumnNumberToName(i + 1)
		_ = f.SetCellValue(allocSheet, col+"1", header)
	}
	for i, a := range allocation {
		row := strconv.Itoa(i + 2)
		_ = f.SetCellValue(allocSheet, "A"+row, a.Category)
		_ = f.SetCellValue(allocSheet, "B"+row, a.MarketValue)
		_ = f.SetCellValue(allocSheet, "C"+row, a.Percent)
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return httpx.InternalServerError(c, "Error serializing report", err.Error())
	}

	c.Set(fiber.HeaderContentType, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set(fiber.HeaderContentDisposition, `attachment; filename="resumen-mensual.xlsx"`)
	return c.Send(buf.Bytes())
}

func (h *handler) ExportTransactions(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	txns, err := h.service.GetRecentUserTransactions(c, userID, 10000)
	if err != nil {
		return httpx.FromDomain(c, err, "Error generating report", "Could not retrieve transactions")
	}
	dtos := NewUserTransactionListResponse(txns)

	f := excelize.NewFile()
	defer func() { _ = f.Close() }()

	sheet := "Transacciones"
	_ = f.SetSheetName("Sheet1", sheet)
	headers := []string{"Fecha", "Tipo", "Activo", "Ticker", "Cantidad", "Precio", "Comisiones", "Moneda", "Notas"}
	for i, header := range headers {
		col, _ := excelize.ColumnNumberToName(i + 1)
		_ = f.SetCellValue(sheet, col+"1", header)
	}
	for i, t := range dtos {
		row := strconv.Itoa(i + 2)
		_ = f.SetCellValue(sheet, "A"+row, t.TransactionDate.Format("2006-01-02"))
		_ = f.SetCellValue(sheet, "B"+row, t.Type)
		_ = f.SetCellValue(sheet, "C"+row, t.AssetName)
		_ = f.SetCellValue(sheet, "D"+row, t.AssetTicker)
		_ = f.SetCellValue(sheet, "E"+row, t.Quantity)
		_ = f.SetCellValue(sheet, "F"+row, t.Price)
		_ = f.SetCellValue(sheet, "G"+row, t.Fees)
		_ = f.SetCellValue(sheet, "H"+row, t.Currency)
		_ = f.SetCellValue(sheet, "I"+row, t.Notes)
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return httpx.InternalServerError(c, "Error serializing report", err.Error())
	}

	c.Set(fiber.HeaderContentType, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set(fiber.HeaderContentDisposition, `attachment; filename="transacciones.xlsx"`)
	return c.Send(buf.Bytes())
}

const (
	riskMetricsSheet   = "Métricas de riesgo"
	riskHistorySheet   = "Historial de crecimiento"
	riskMonthlySheet   = "Rentabilidad mensual"
	riskUnavailableTag = "Sin historial suficiente"
)

// ExportRiskMetrics publishes the risk report as three sheets: the figures, the
// month-by-month return behind them, and the raw series they were computed on.
//
// It used to be the series alone, which is what the file was named after but
// not what it contained: a reader who opened "riesgo-volatilidad.xlsx" got a
// column of valuations and had to work out the volatility themselves. The
// numbers are the ones the reports page shows, from the same series and the
// same formulas, so downloading the file to check one is worth doing.
func (h *handler) ExportRiskMetrics(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	points, _, err := h.service.GetPortfolioGrowth(c, userID, "", "ALL")
	if err != nil {
		return httpx.FromDomain(c, err, "Error generating report", "Could not retrieve growth data")
	}

	metrics := BuildGrowthMetrics(points)

	f := excelize.NewFile()
	defer func() { _ = f.Close() }()

	_ = f.SetSheetName("Sheet1", riskMetricsSheet)
	writeRiskMetricsSheet(f, metrics)

	_, _ = f.NewSheet(riskMonthlySheet)
	writeRiskMonthlySheet(f, metrics)

	_, _ = f.NewSheet(riskHistorySheet)
	writeRiskHistorySheet(f, points)

	buf, err := f.WriteToBuffer()
	if err != nil {
		return httpx.InternalServerError(c, "Error serializing report", err.Error())
	}

	c.Set(fiber.HeaderContentType, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set(fiber.HeaderContentDisposition, `attachment; filename="riesgo-volatilidad.xlsx"`)
	return c.Send(buf.Bytes())
}

// writeHeaders lays the first row of a sheet.
func writeHeaders(f *excelize.File, sheet string, headers []string) {
	for i, header := range headers {
		col, _ := excelize.ColumnNumberToName(i + 1)
		_ = f.SetCellValue(sheet, col+"1", header)
	}
}

// pct renders a fraction as the percentage the report speaks in, or the
// "not enough history" tag when the metric could not be computed at all. The
// tag matters: a blank cell reads as zero, and zero volatility is a claim.
func pct(value decimal.Decimal, available bool) string {
	if !available {
		return riskUnavailableTag
	}

	return value.Mul(oneHundred).RoundBank(2).StringFixed(2)
}

func writeRiskMetricsSheet(f *excelize.File, m GrowthMetrics) {
	writeHeaders(f, riskMetricsSheet, []string{"Métrica", "Valor", "Unidad", "Cómo se calcula"})

	if m.Points < 2 {
		_ = f.SetCellValue(riskMetricsSheet, "A2", "Historial")
		_ = f.SetCellValue(riskMetricsSheet, "B2", riskUnavailableTag)
		_ = f.SetCellValue(riskMetricsSheet, "C2", "")
		_ = f.SetCellValue(riskMetricsSheet, "D2", "Hacen falta al menos dos cierres diarios de la cartera.")
		return
	}

	currency := m.Currency
	if currency == "" {
		currency = "USD"
	}

	rows := [][4]string{
		{"Periodo cubierto", m.FirstDate.Format("2006-01-02") + " a " + m.LastDate.Format("2006-01-02"), "fecha", "Del primer cierre diario al último."},
		{"Días de historial", strconv.Itoa(m.SpanDays), "días", "Días naturales entre esos dos cierres."},
		{"Puntos de la serie", strconv.Itoa(m.Points), "puntos", "Un cierre diario por punto."},
		{"Rentabilidad del periodo", pct(m.TotalReturn, true), "%", "Retornos de cada tramo encadenados, con los aportes y retiros descontados (Dietz modificada)."},
		{"Rentabilidad anualizada", pct(m.Annualized, m.HasAnnualized), "%", "(1 + rentabilidad del periodo) ^ (365,25 / días) − 1. Desde 90 días de historial, el mismo umbral que la volatilidad y el Sharpe."},
		volatilityRow(m),
		{"Máxima caída", pct(m.MaxDrawdown, true), "%", "Peor bajada desde un máximo del índice de rentabilidad, no del saldo. Se mide tramo a tramo, así que puede caer dentro de un mes que cerró en positivo."},
		{"Ratio de Sharpe", ratio(m.Sharpe, m.HasSharpe), "veces", "Retorno medio de los tramos ÷ volatilidad, × √(tramos por año), con tasa libre de riesgo 0. No es la rentabilidad anualizada de arriba ÷ volatilidad: aquella es compuesta y sale más alta."},
		{"Mejor mes", monthCell(m.Best, m.HasMonthReturns), "%", monthNote(m, "más alto")},
		{"Peor mes", monthCell(m.Worst, m.HasMonthReturns), "%", monthNote(m, "más bajo")},
		{"Valor actual", m.CurrentValue.RoundBank(2).StringFixed(2), currency, "Valor de mercado de la cuenta en el último cierre."},
		{"Capital invertido", m.InvestedCost.RoundBank(2).StringFixed(2), currency, "Coste de las posiciones abiertas en el último cierre."},
		{"Ganancia / pérdida", m.GainLoss.RoundBank(2).StringFixed(2), currency, "Valor de mercado menos capital invertido."},
		{"Aporte neto del periodo", m.NetFlow.RoundBank(2).StringFixed(2), currency, "Dinero puesto menos dinero sacado: la suma de la columna «Aporte neto» del historial. No es rentabilidad."},
	}

	for i, entry := range rows {
		row := strconv.Itoa(i + 2)
		for col, value := range entry {
			name, _ := excelize.ColumnNumberToName(col + 1)
			_ = f.SetCellValue(riskMetricsSheet, name+row, value)
		}
	}
}

// volatilityRow names the dispersion figure after what it actually is. Below
// the ninety days that annualizing asks for, the sheet still publishes it —
// dispersion converges long before a mean does — but as the per-subperiod
// figure it is, because the two differ by a factor of twenty on a daily series.
func volatilityRow(m GrowthMetrics) [4]string {
	if !m.VolatilityAnnualized {
		return [4]string{
			"Volatilidad por tramo", pct(m.Volatility, m.HasVolatility), "%",
			"Desviación típica muestral de los retornos de cada tramo, sin anualizar: eso pide 90 días de historial. Desde 10 tramos.",
		}
	}

	return [4]string{
		"Volatilidad anualizada", pct(m.Volatility, m.HasVolatility), "%",
		"Desviación típica muestral de los retornos × √(tramos por año). Desde 10 tramos y 90 días.",
	}
}

// ratio renders the Sharpe ratio, which is a multiple and not a percentage.
func ratio(value decimal.Decimal, available bool) string {
	if !available {
		return riskUnavailableTag
	}

	return value.RoundBank(2).StringFixed(2)
}

// monthCell renders "2026-03: +1,20" style content as the month and its return.
// A month the history does not cover whole carries the same asterisk the
// reports page uses, so the two never disagree on which figure is comparable.
func monthCell(month MonthReturn, available bool) string {
	if !available {
		return riskUnavailableTag
	}

	cell := month.Month + " " + month.Rate.Mul(oneHundred).RoundBank(2).StringFixed(2)
	if month.Partial {
		cell += " *"
	}

	return cell
}

// monthNote explains what the best and worst month leave out, and warns when
// the history is so short that the pick had to be a partial month after all.
func monthNote(m GrowthMetrics, extreme string) string {
	note := "Mes completo con el retorno encadenado " + extreme + "; los meses parciales —aquel en el que empieza el historial y el que está en curso— no compiten."
	if m.MonthsPartialOnly {
		note += " Todavía no hay ningún mes completo: la cifra sale de uno parcial y va marcada con *."
	}

	return note
}

func writeRiskMonthlySheet(f *excelize.File, m GrowthMetrics) {
	// The third column is what keeps a reader from comparing three days against
	// a full month: the rate of a partial month is real, only not comparable.
	writeHeaders(f, riskMonthlySheet, []string{"Mes", "Rentabilidad %", "Mes completo"})

	for i, month := range MonthlyReturns(m.Subperiod) {
		row := strconv.Itoa(i + 2)
		whole := "sí"
		if month.Partial {
			whole = "no"
		}

		_ = f.SetCellValue(riskMonthlySheet, "A"+row, month.Month)
		_ = f.SetCellValue(riskMonthlySheet, "B"+row, month.Rate.Mul(oneHundred).RoundBank(2).StringFixed(2))
		_ = f.SetCellValue(riskMonthlySheet, "C"+row, whole)
	}
}

func writeRiskHistorySheet(f *excelize.File, points []GrowthPoint) {
	// The flow column is what makes the rest auditable: without it a reader
	// cannot tell a day the market moved from a day the owner deposited.
	writeHeaders(f, riskHistorySheet, []string{
		"Fecha", "Valor Total", "Costo Base", "Ganancia/Pérdida", "Rentabilidad %", "Aporte neto",
	})

	for i, p := range points {
		row := strconv.Itoa(i + 2)
		netFlow := p.NetFlow
		if netFlow == "" {
			netFlow = "0"
		}

		_ = f.SetCellValue(riskHistorySheet, "A"+row, p.Date.Format("2006-01-02"))
		_ = f.SetCellValue(riskHistorySheet, "B"+row, p.TotalValue)
		_ = f.SetCellValue(riskHistorySheet, "C"+row, p.TotalCostBase)
		_ = f.SetCellValue(riskHistorySheet, "D"+row, p.GainLoss)
		_ = f.SetCellValue(riskHistorySheet, "E"+row, p.GainLossPct)
		_ = f.SetCellValue(riskHistorySheet, "F"+row, netFlow)
	}
}
