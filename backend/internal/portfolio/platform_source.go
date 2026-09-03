package portfolio

import (
	"context"
	"strings"

	"uuid"

	"github.com/yeferson59/gofinance/v2/money"
)

func (s *service) CreatePlatform(ctx context.Context, userID uuid.UUID, sourceType SourceType, name, description string) (InvestmentSource, error) {
	platform, err := s.repo.CreatePlatform(ctx, userID, sourceType, strings.ToLower(name), description)
	if err != nil {
		return InvestmentSource{}, err
	}

	return platform, nil
}

// GetPlatforms lists the user's platforms with their totals converted into
// displayCurrency. An empty displayCurrency leaves them in the account's
// preferred currency; either way the answer says which one it used.
func (s *service) GetPlatforms(ctx context.Context, userID uuid.UUID, displayCurrency money.Currency) ([]PlatformStats, error) {
	return s.repo.GetPlatformsWithStats(ctx, userID, displayCurrency)
}

func (s *service) UpdatePlatform(ctx context.Context, userID, sourceID uuid.UUID, name, description string, sourceType SourceType, isActive bool) (PlatformStats, error) {
	return s.repo.UpdatePlatform(ctx, userID, sourceID, name, description, sourceType, isActive)
}

func (s *service) DeletePlatform(ctx context.Context, userID, sourceID uuid.UUID) error {
	return s.repo.DeletePlatform(ctx, userID, sourceID)
}
