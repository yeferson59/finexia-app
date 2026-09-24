package mcp

import (
	"context"
	"fmt"
	"time"

	"uuid"

	"github.com/yeferson59/gofinance/v2/money"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/yeferson59/finexia-app/internal/market"
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

	readTool(s, "get_sector_allocation", "Get sector allocation",
		"Total the user's holdings per industry (technology, healthcare, financials…), every slice in one currency so the shares add up. Use this for concentration questions — how much of the money rides on one industry — which get_allocation cannot answer: eight tickers across three portfolios can be one bet on semiconductors and still look diversified by asset type. Assets with no classification are reported in their own slice rather than dropped, so check it before calling a portfolio well spread.",
		func(ctx context.Context, in CurrencyInput) (SectorAllocationOutput, error) {
			cur, err := parseCurrency(in.Currency)
			if err != nil {
				return SectorAllocationOutput{}, err
			}

			items, err := m.portfolios.GetSectorAllocation(ctx, c.userID, cur)
			if err != nil {
				return SectorAllocationOutput{}, m.logToolError(ctx, "get_sector_allocation", c, err)
			}

			return SectorAllocationOutput{Allocation: sectorSlices(items)}, nil
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

	readTool(s, "get_cash_accounts", "Get cash accounts",
		"List the user's cash balances — what each platform holds in each currency, per portfolio — with the rate the account earns and what that interest has paid. Use this for questions about idle money and what it yields: how much is in cash, which balances earn nothing, how much interest came in this month. The interest counts as gain, never as money put in, so it shows up in the portfolio's gain and in its return.",
		func(ctx context.Context, in CurrencyInput) (CashOutput, error) {
			cur, err := parseCurrency(in.Currency)
			if err != nil {
				return CashOutput{}, err
			}

			balances, err := m.portfolios.GetCashBalances(ctx, c.userID, cur)
			if err != nil {
				return CashOutput{}, m.logToolError(ctx, "get_cash_accounts", c, err)
			}

			rates, err := m.portfolios.GetCashRates(ctx, c.userID)
			if err != nil {
				return CashOutput{}, m.logToolError(ctx, "get_cash_accounts", c, err)
			}

			return CashOutput{Accounts: cashAccountRows(balances, rates, time.Now())}, nil
		})

	readTool(s, "get_funds", "Get investment funds",
		"List the user's investment funds — collective investment funds, voluntary pension funds, money-market pockets: money that earns whatever it turns out to earn, not a fixed rate — with what each is worth at its latest recorded value, the money that went in and out, and its return over 30, 90, 180 and 365 days, the year so far and since inception, also annualized. Use it for questions about how a fund did. A fund priced at cost has no recorded value yet, so its zero gain must not be reported as a return; a period with no figure is one the history does not cover.",
		func(ctx context.Context, _ EmptyInput) (FundsOutput, error) {
			funds, err := m.portfolios.GetFunds(ctx, c.userID)
			if err != nil {
				return FundsOutput{}, m.logToolError(ctx, "get_funds", c, err)
			}

			rows := make([]FundRow, 0, len(funds))

			for _, f := range funds {
				perf, err := m.portfolios.GetFundPerformance(ctx, c.userID, f.AssetID)
				if err != nil {
					return FundsOutput{}, m.logToolError(ctx, "get_funds", c, err)
				}

				rows = append(rows, fundRow(f, perf))
			}

			return FundsOutput{Funds: rows}, nil
		})
}

func fundRow(f portfolio.Fund, perf portfolio.FundPerformance) FundRow {
	row := FundRow{
		Name:         f.Name,
		Platforms:    make([]string, 0, len(f.Positions)),
		Currency:     f.Currency.String(),
		Tracking:     string(f.Tracking),
		Value:        f.Value,
		Cost:         f.Cost,
		PricedAtCost: f.PricedAtCost,
		Invested:     perf.Invested,
		Withdrawn:    perf.Withdrawn,
		RealizedGain: perf.RealizedGain,
		Returns:      make([]FundReturn, 0, len(perf.Periods)),
	}

	seen := make(map[string]bool)
	for _, p := range f.Positions {
		if !seen[p.SourceName] {
			seen[p.SourceName] = true
			row.Platforms = append(row.Platforms, p.SourceName)
		}
	}

	if f.UnitValue != nil {
		row.UnitValue = *f.UnitValue
	}

	if f.ValuedOn != nil {
		row.ValuedOn = timeText(*f.ValuedOn)
	}

	if pf := f.PublicFund; pf != nil {
		row.PublishedBy = fmt.Sprintf("%s — %s (participación %d)", pf.EntityName, pf.FundName, pf.Participation)
	}

	if perf.UnrealizedGain != nil {
		row.UnrealizedGain = *perf.UnrealizedGain
	}

	for _, p := range perf.Periods {
		r := FundReturn{Period: p.Key}

		if p.From != nil {
			r.From = timeText(*p.From)
		}

		if p.Days != nil {
			r.Days = *p.Days
		}

		if p.Pct != nil {
			r.Pct = *p.Pct
		}

		if p.EAPct != nil {
			r.EAPct = *p.EAPct
		}

		row.Returns = append(row.Returns, r)
	}

	return row
}

// cashAccountKey names an account: a platform, a currency and a pocket of it,
// which is what a rate belongs to (000047). The main account is the empty
// pocket, so a platform without pockets keys exactly as it always did.
type cashAccountKey struct {
	sourceID uuid.UUID
	currency money.Currency
	pocketID uuid.UUID
}

// pocketKey is a pocket as the key holds it: the zero UUID for the main
// account.
func pocketKey(pocketID *uuid.UUID) uuid.UUID {
	if pocketID == nil {
		return uuid.UUID{}
	}

	return *pocketID
}

// ratesInEffect indexes the version of each account's rate that applies on the
// given day. Only the newest version of an account can still be running, so a
// version a later one follows is never in effect.
func ratesInEffect(rates []portfolio.CashRate, day time.Time) map[cashAccountKey]portfolio.CashRate {
	inEffect := make(map[cashAccountKey]portfolio.CashRate, len(rates))

	for _, r := range rates {
		if r.InEffectOn(day) {
			inEffect[cashAccountKey{sourceID: r.SourceID, currency: r.Currency, pocketID: pocketKey(r.PocketID)}] = r
		}
	}

	return inEffect
}

func cashAccountRows(balances []portfolio.CashBalance, rates []portfolio.CashRate, now time.Time) []CashAccount {
	inEffect := ratesInEffect(rates, now)
	out := make([]CashAccount, 0, len(balances))

	for _, b := range balances {
		row := CashAccount{
			PortfolioID:       b.PortfolioID.String(),
			PortfolioName:     b.PortfolioName,
			PlatformID:        b.SourceID.String(),
			Platform:          b.SourceName,
			Pocket:            b.PocketName,
			Balance:           b.Balance,
			Currency:          b.Currency.String(),
			Value:             b.Value,
			DisplayCurrency:   b.DisplayCurrency.String(),
			InterestEarned:    b.InterestEarned,
			InterestThisMonth: b.InterestThisMonth,
			PendingInterest:   b.PendingInterest,
			Movements:         b.Movements,
		}

		if b.LastAccrualDate != nil {
			row.LastAccrualDate = timeText(*b.LastAccrualDate)
		}

		if b.LastMovementDate != nil {
			row.LastMovementDate = timeText(*b.LastMovementDate)
		}

		if rate, ok := inEffect[cashAccountKey{sourceID: b.SourceID, currency: b.Currency, pocketID: pocketKey(b.PocketID)}]; ok {
			row.AnnualRatePct = rate.AnnualRatePct
			row.WithholdingPct = rate.WithholdingPct
			row.Posting = string(rate.Posting)
			row.RateFrom = timeText(rate.EffectiveFrom)

			for _, tier := range rate.Tiers {
				row.Tiers = append(row.Tiers, CashAccountTier{FromBalance: tier.FromBalance, AnnualRatePct: tier.AnnualRatePct})
			}
		}

		out = append(out, row)
	}

	return out
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
			Sector:          string(h.Sector),
			SectorWeights:   sectorWeightRows(h.SectorWeights),
			Portfolios:      h.Portfolios,
			PriceSource:     string(h.PriceSource),
		})
	}

	return out
}

// sectorWeightRows maps an asset's breakdown into the tool's own shape, as text
// for the same reason every other amount here is text: the client is a language
// model, and a JSON number is where a decimal loses its digits.
//
// nil for an asset with no breakdown, which is most of them, so the field drops
// out of the payload entirely rather than repeating an empty array on every row.
func sectorWeightRows(weights market.SectorBreakdown) []SectorWeight {
	if weights.IsEmpty() {
		return nil
	}

	out := make([]SectorWeight, 0, len(weights))

	for _, w := range weights {
		out = append(out, SectorWeight{Sector: string(w.Sector), Weight: w.Weight.String()})
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

func sectorSlices(items []portfolio.SectorAllocationItem) []SectorSlice {
	out := make([]SectorSlice, 0, len(items))

	for _, i := range items {
		out = append(out, SectorSlice{
			Sector:      string(i.Sector),
			MarketValue: i.MarketValue,
			Currency:    i.Currency.String(),
			Assets:      i.Assets,
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
