package portfolio

import (
	"context"
	"strings"

	"github.com/google/uuid"
)

func (s *Service) CreatePlatform(ctx context.Context, userID uuid.UUID, sourceType SourceType, name, description string) (InvestmentSource, error) {
	platform, err := s.repo.CreatePlatform(ctx, userID, sourceType, strings.ToLower(name), description)
	if err != nil {
		return InvestmentSource{}, err
	}

	return platform, nil
}

func (s *Service) GetPlatforms(ctx context.Context, userID uuid.UUID) ([]PlatformStats, error) {
	return s.repo.GetPlatformsWithStats(ctx, userID)
}

func (s *Service) UpdatePlatform(ctx context.Context, userID, sourceID uuid.UUID, name, description string, sourceType SourceType, isActive bool) (PlatformStats, error) {
	return s.repo.UpdatePlatform(ctx, userID, sourceID, name, description, sourceType, isActive)
}

func (s *Service) DeletePlatform(ctx context.Context, userID, sourceID uuid.UUID) error {
	return s.repo.DeletePlatform(ctx, userID, sourceID)
}
