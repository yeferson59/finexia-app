package mcp

import (
	"context"

	"github.com/yeferson59/gofinance/v2/money"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/yeferson59/finexia-app/internal/portfolio"
)

// Default and maximum page sizes for the tools that take a limit. The ceilings
// are what one tool result may carry without crowding out the conversation it
// is an answer inside of.
const (
	defaultTransactionLimit = 20
	maxTransactionLimit     = 200
)

// addPortfolioTools registers everything a caller can ask about their own
// positions. Each handler closes over the caller, which is what makes the
// per-request server the isolation boundary: there is no argument by which a
// tool could be pointed at another user's data.
func (m *Module) addPortfolioTools(s *mcpsdk.Server, c caller) {
	readTool(s, "list_portfolios", "List portfolios",
		"List the user's portfolios with what each one cost, what it is worth now and its gain or loss. Start here: the ids it returns are what the other portfolio tools accept.",
		func(ctx context.Context, in CurrencyInput) (PortfoliosOutput, error) {
			cur, err := parseCurrency(in.Currency)
			if err != nil {
				return PortfoliosOutput{}, err
			}

			var views []portfolio.SummaryView

			// Two calls rather than one with an empty currency: reporting each
			// portfolio in its own base currency is a different question from
			// converting them all into one, and the service keeps them apart.
			if cur == money.XXX {
				views, err = m.portfolios.GetPortfoliosSummary(ctx, c.userID)
			} else {
				views, err = m.portfolios.GetPortfoliosSummaryInCurrency(ctx, c.userID, cur)
			}

			if err != nil {
				return PortfoliosOutput{}, m.logToolError(ctx, "list_portfolios", c, err)
			}

			return PortfoliosOutput{Portfolios: portfolioSummaries(views)}, nil
		})

	readTool(s, "get_holdings", "Get holdings",
		"List what the user holds of each asset, totalled across every portfolio they own. Use this to answer how much of a given ticker they have; list_portfolios answers how the money is split between portfolios.",
		func(ctx context.Context, in CurrencyInput) (HoldingsOutput, error) {
			cur, err := parseCurrency(in.Currency)
			if err != nil {
				return HoldingsOutput{}, err
			}

			holdings, err := m.portfolios.GetAssetHoldings(ctx, c.userID, cur)
			if err != nil {
				return HoldingsOutput{}, m.logToolError(ctx, "get_holdings", c, err)
			}

			return HoldingsOutput{Holdings: holdingRows(holdings)}, nil
		})

	readTool(s, "get_allocation", "Get allocation",
		"Total the user's holdings per asset category (stocks, etfs, cryptos, bonds, cash, real estate, commodities), every slice in one currency so the shares add up.",
		func(ctx context.Context, in CurrencyInput) (AllocationOutput, error) {
			cur, err := parseCurrency(in.Currency)
			if err != nil {
				return AllocationOutput{}, err
			}

			items, err := m.portfolios.GetAssetAllocation(ctx, c.userID, cur)
			if err != nil {
				return AllocationOutput{}, m.logToolError(ctx, "get_allocation", c, err)
			}

			return AllocationOutput{Allocation: allocationSlices(items)}, nil
		})

	readTool(s, "list_recent_transactions", "List recent transactions",
		"List the user's most recent transactions across every portfolio, newest first.",
		func(ctx context.Context, in TransactionsInput) (TransactionsOutput, error) {
			limit := clampLimit(in.Limit, defaultTransactionLimit, maxTransactionLimit)

			txns, err := m.portfolios.GetRecentUserTransactions(ctx, c.userID, limit)
			if err != nil {
				return TransactionsOutput{}, m.logToolError(ctx, "list_recent_transactions", c, err)
			}

			return TransactionsOutput{Transactions: transactionRows(txns)}, nil
		})

	readTool(s, "get_portfolio_growth", "Get portfolio growth",
		"Return the account-wide value series from the daily snapshots, with the summary that separates how the total moved (deposits included) from what was actually earned.",
		func(ctx context.Context, in GrowthInput) (GrowthOutput, error) {
			cur, err := parseCurrency(in.Currency)
			if err != nil {
				return GrowthOutput{}, err
			}

			points, summary, err := m.portfolios.GetPortfolioGrowth(ctx, c.userID, cur, in.Period)
			if err != nil {
				return GrowthOutput{}, m.logToolError(ctx, "get_portfolio_growth", c, err)
			}

			return GrowthOutput{Summary: growthSummary(summary), Points: growthPoints(points)}, nil
		})

	readTool(s, "list_platforms", "List platforms",
		"List the brokers, banks, wallets and other platforms the user holds positions through, with what is held on each.",
		func(ctx context.Context, in CurrencyInput) (PlatformsOutput, error) {
			cur, err := parseCurrency(in.Currency)
			if err != nil {
				return PlatformsOutput{}, err
			}

			platforms, err := m.portfolios.GetPlatforms(ctx, c.userID, cur)
			if err != nil {
				return PlatformsOutput{}, m.logToolError(ctx, "list_platforms", c, err)
			}

			return PlatformsOutput{Platforms: platformRows(platforms)}, nil
		})
}

func portfolioSummaries(views []portfolio.SummaryView) []PortfolioSummary {
	out := make([]PortfolioSummary, 0, len(views))

	for _, v := range views {
		out = append(out, PortfolioSummary{
			ID:                v.ID.String(),
			Name:              v.Name,
			Description:       v.Description,
			Type:              string(v.Type),
			Risk:              v.RiskName,
			BaseCurrency:      v.BaseCurrency.String(),
			DisplayCurrency:   v.DisplayCurrency.String(),
			IsDefault:         v.IsDefault,
			Positions:         v.TotalPositions,
			CostBasis:         v.TotalCostBase,
			MarketValue:       v.TotalMarketValue,
			GainLoss:          v.TotalGainLoss,
			GainLossPct:       v.TotalGainLossPct,
			PricedWithOwnKey:  v.PositionsPricedOwn,
			PricedManually:    v.PositionsPricedManual,
			ValuedAtCost:      v.PositionsAtCost,
			PositionsUnvalued: v.PositionsUnconverted,
			ConvertedFromBase: v.FXConverted,
		})
	}

	return out
}

func holdingRows(holdings []portfolio.AssetHolding) []Holding {
	out := make([]Holding, 0, len(holdings))

	for _, h := range holdings {
		out = append(out, Holding{
			AssetID:         h.AssetID.String(),
			Ticker:          h.Ticker,
			Name:            h.Name,
			AssetType:       string(h.AssetType),
			Exchange:        h.Exchange,
			Quantity:        h.Quantity,
			Currency:        h.Currency.String(),
			MarketPrice:     h.MarketPrice,
			MarketValue:     h.MarketValue,
			DisplayCurrency: h.DisplayCurrency.String(),
			Portfolios:      h.Portfolios,
			PriceSource:     string(h.PriceSource),
		})
	}

	return out
}

func allocationSlices(items []portfolio.AllocationItem) []AllocationSlice {
	out := make([]AllocationSlice, 0, len(items))

	for _, i := range items {
		out = append(out, AllocationSlice{
			Category:    string(i.Category),
			MarketValue: i.MarketValue,
			Currency:    i.Currency.String(),
			Unconverted: i.PositionsUnconverted,
		})
	}

	return out
}

func transactionRows(txns []portfolio.Transaction) []Transaction {
	out := make([]Transaction, 0, len(txns))

	for _, t := range txns {
		row := Transaction{
			ID:       t.ID.String(),
			Ticker:   t.Entry.Asset.Ticker,
			Type:     string(t.Type),
			Quantity: t.Quantity.String(),
			Price:    t.Price.String(),
			Currency: t.Currency.String(),
			FXRate:   t.FXRate.String(),
			Fees:     t.Fees.String(),
			Date:     timeText(t.TransactionDate),
			Notes:    t.Notes,
		}

		// The asset name and the two secondary currencies are only present on
		// reads that join the entry; left empty they are omitted rather than
		// reported as XXX, which is a currency code and would read as one.
		row.AssetName = t.Entry.Asset.Name

		if t.CostCurrency != money.XXX {
			row.CostCurrency = t.CostCurrency.String()
		}

		if t.FeesCurrency != money.XXX {
			row.FeesCurrency = t.FeesCurrency.String()
		}

		out = append(out, row)
	}

	return out
}

func growthSummary(s portfolio.GrowthSummary) GrowthSummary {
	return GrowthSummary{
		FirstDate:    timeText(s.FirstDate),
		InitialValue: s.InitialValue,
		CurrentValue: s.CurrentValue,
		GrowthPct:    s.TotalGrowthPct,
		GainLoss:     s.GainLoss,
		GainLossPct:  s.GainLossPct,
		Currency:     s.Currency.String(),
	}
}

func growthPoints(points []portfolio.GrowthPoint) []GrowthPoint {
	out := make([]GrowthPoint, 0, len(points))

	for _, p := range points {
		out = append(out, GrowthPoint{
			Date:        timeText(p.Date),
			TotalValue:  p.TotalValue,
			CostBasis:   p.TotalCostBase,
			GainLoss:    p.GainLoss,
			GainLossPct: p.GainLossPct,
			Currency:    p.Currency.String(),
			NetFlow:     p.NetFlow,
			Unconverted: p.PortfoliosUnconverted,
		})
	}

	return out
}

func platformRows(platforms []portfolio.PlatformStats) []Platform {
	out := make([]Platform, 0, len(platforms))

	for _, p := range platforms {
		out = append(out, Platform{
			ID:               p.ID.String(),
			Name:             p.Name,
			Description:      p.Description,
			Type:             string(p.SourceType),
			IsActive:         p.IsActive,
			Positions:        p.Investments,
			Assets:           p.Assets,
			Portfolios:       p.Portfolios,
			CostBasis:        p.TotalValue,
			MarketValue:      p.MarketValue,
			DisplayCurrency:  p.DisplayCurrency.String(),
			PricedWithOwnKey: p.PositionsPricedOwn,
			PricedManually:   p.PositionsPricedManual,
			ValuedAtCost:     p.PositionsAtCost,
		})
	}

	return out
}
