package entities

import "testing"

func TestAssetTypeTransform(t *testing.T) {
	cases := []struct {
		in   AssetType
		want PortfolioEntryCategory
	}{
		{Stock, Stocks},
		{ETF, ETFs},
		{Crypto, Cryptos},
		{Bond, Bonds},
		{Cash, Cashs},
		{RealEstate, RealEstates},
		{Commodity, Commodities},
		{Other, Others},
		{AssetType("nonsense"), Others},
		{AssetType(""), Others},
	}

	for _, tc := range cases {
		if got := tc.in.Transform(); got != tc.want {
			t.Errorf("AssetType(%q).Transform() = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestSourceTypeIsValid(t *testing.T) {
	valid := []SourceType{Broker, Bank, TradingPlatform, Neobank, DeFi, CryptoWallet, MutualFunds, BrokerageHouse, OtherType}
	for _, s := range valid {
		if !s.IsValid() {
			t.Errorf("SourceType(%q).IsValid() = false, want true", s)
		}
	}

	invalid := []SourceType{"", "bank", "BROKER", "exchange"}
	for _, s := range invalid {
		if s.IsValid() {
			t.Errorf("SourceType(%q).IsValid() = true, want false", s)
		}
	}
}

func TestTransactionTypeIsValid(t *testing.T) {
	valid := []TransactionType{Buy, Sell, Dividend, Split, TransferIn, TransferOut, Fee, Interest}
	for _, tt := range valid {
		if !tt.IsValid() {
			t.Errorf("TransactionType(%q).IsValid() = false, want true", tt)
		}
	}

	invalid := []TransactionType{"", "BUY", "swap", "transfer"}
	for _, tt := range invalid {
		if tt.IsValid() {
			t.Errorf("TransactionType(%q).IsValid() = true, want false", tt)
		}
	}
}

func TestPortfolioEntryCategoryIsValid(t *testing.T) {
	valid := []PortfolioEntryCategory{Stocks, ETFs, Cryptos, Bonds, Cashs, RealEstates, Commodities, Others}
	for _, c := range valid {
		if !c.IsValid() {
			t.Errorf("PortfolioEntryCategory(%q).IsValid() = false, want true", c)
		}
	}

	invalid := []PortfolioEntryCategory{"", "stock", "crypto", "STOCKS"}
	for _, c := range invalid {
		if c.IsValid() {
			t.Errorf("PortfolioEntryCategory(%q).IsValid() = true, want false", c)
		}
	}
}

func TestPortfolioTypeIsValid(t *testing.T) {
	valid := []PortfolioType{
		PortfolioTypeStocks, PortfolioTypeETFs, PortfolioTypeCryptos, PortfolioTypeBonds,
		PortfolioTypeCash, PortfolioTypeForex, PortfolioTypeRealEstates, PortfolioTypeCommodities,
		PortfolioTypeForexStocks, PortfolioTypeForexETFs, PortfolioTypeForexCryptos, PortfolioTypeForexBonds,
		PortfolioTypeForexCash, PortfolioTypeForexRealStates, PortfolioTypeForexCommodities,
		PortfolioTypeStocksETFs, PortfolioTypeStocksCryptos, PortfolioTypeStocksBonds, PortfolioTypeStocksCash,
		PortfolioTypeStocksRealEstates, PortfolioTypeStocksCommodities,
		PortfolioTypeETFsCryptos, PortfolioTypeETFsBonds, PortfolioTypeETFsCash, PortfolioTypeETFsRealEstates,
		PortfolioTypeETFsCommodities,
		PortfolioTypeCryptosBonds, PortfolioTypeCryptosCash, PortfolioTypeCryptosRealEstates, PortfolioTypeCryptosCommodities,
		PortfolioTypeBondsCash, PortfolioTypeBondsRealEstates, PortfolioTypeBondsCommodities,
		PortfolioTypeCashRealEstates, PortfolioTypeCashCommodities,
		PortfolioTypeRealEstatesCommodities, PortfolioTypeDiversified,
	}
	for _, p := range valid {
		if !p.IsValid() {
			t.Errorf("PortfolioType(%q).IsValid() = false, want true", p)
		}
	}

	invalid := []PortfolioType{"", "STOCKS", "stocks_forex", "mixed", "crypto"}
	for _, p := range invalid {
		if p.IsValid() {
			t.Errorf("PortfolioType(%q).IsValid() = true, want false", p)
		}
	}
}
