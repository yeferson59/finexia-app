package portfolio

import (
	"context"

	"uuid"

	"github.com/yeferson59/finexia-app/internal/market"
)

// This file satisfies market.Holdings. The market module needs to know which
// assets a user holds so a personal API quota is spent only on those, but it
// must not import portfolio — the dependency runs portfolio → market, and
// app/arch_test.go pins that direction. So market declares the interface and
// this module implements it.

var _ market.Holdings = (*service)(nil)

// HeldAssetIDs lists the assets the user actually owns across every portfolio.
// Under BYO-key this bounds the sync to what the user cares about: walking the
// whole catalog would burn a personal free-tier quota on assets they do not
// hold.
func (s *service) HeldAssetIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	return s.repo.GetHeldAssetIDs(ctx, userID)
}

// RequiredCurrencyPairs lists the conversions the user's portfolios actually
// need: from each currency their holdings are priced in, to each base currency
// their portfolios report in. Same reasoning as above — only fetch what this
// user's screens will display.
func (s *service) RequiredCurrencyPairs(ctx context.Context, userID uuid.UUID) ([]market.CurrencyPair, error) {
	pairs, err := s.repo.GetRequiredCurrencyPairs(ctx, userID)
	if err != nil {
		return nil, err
	}

	out := make([]market.CurrencyPair, 0, len(pairs))
	for _, p := range pairs {
		out = append(out, market.CurrencyPair{From: p.From, To: p.To})
	}

	return out, nil
}
