package market

import (
	"context"

	"github.com/google/uuid"
	"github.com/yeferson59/gofinance/v2/money"
)

// Asset catalog use cases. The market module owns the asset lifecycle; the
// portfolio module consumes these through the interfaces it declares.

func (s *Service) GetAssets(ctx context.Context, offset, limit uint) ([]Asset, error) {
	return s.repo.GetAssets(ctx, offset, limit)
}

func (s *Service) SearchAssets(ctx context.Context, search string, offset, limit uint) ([]Asset, error) {
	return s.repo.SearchAssets(ctx, search, offset, limit)
}

func (s *Service) GetAssetByID(ctx context.Context, assetID uuid.UUID) (Asset, error) {
	return s.repo.GetAssetByID(ctx, assetID)
}

func (s *Service) CreateAsset(ctx context.Context, ticker, name string, assetType AssetType, exchange, currency string) (Asset, error) {
	return s.repo.UpsertAsset(ctx, ticker, name, assetType, exchange, currency)
}

func (s *Service) UpdateAssetPrice(ctx context.Context, assetID uuid.UUID, price money.Money) (Asset, error) {
	return s.repo.UpdateAssetPrice(ctx, assetID, price)
}
