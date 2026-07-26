package market

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/yeferson59/gofinance/v2/money"
)

// Asset catalog use cases. The market module owns the asset lifecycle; the
// portfolio module consumes these through the interfaces it declares.
//
// Creating an asset comes in two shapes. CreateAsset is the operator's: it
// curates, so its rows are visible to everybody and it may overwrite the
// metadata of a ticker that already exists. ContributeAsset is the user's: it
// never overwrites and its rows are visible only to the users who asked for
// them. The split is the whole point — before BYO-key the catalog was the
// operator's because the operator paid the provider quota, and the only way a
// user could add to it was the side door in the transaction importer.

func (s *Service) GetAssets(ctx context.Context, view CatalogView, offset, limit uint) ([]Asset, error) {
	return s.repo.GetAssets(ctx, view, offset, limit)
}

func (s *Service) SearchAssets(ctx context.Context, view CatalogView, search string, offset, limit uint) ([]Asset, error) {
	return s.repo.SearchAssets(ctx, view, search, offset, limit)
}

func (s *Service) GetAssetByID(ctx context.Context, assetID uuid.UUID) (Asset, error) {
	return s.repo.GetAssetByID(ctx, assetID)
}

// CreateAsset curates a catalog row. Operator-only, and the seed runs through
// it too.
func (s *Service) CreateAsset(ctx context.Context, ticker, name string, assetType AssetType, exchange, currency string) (Asset, error) {
	input, err := normalizeAssetInput(ticker, name, assetType, exchange, currency)
	if err != nil {
		return Asset{}, err
	}

	return s.repo.UpsertAsset(ctx, input.ticker, input.name, input.assetType, input.exchange, input.currency)
}

// maxContributedAssetsPerDay bounds how many new catalog rows one user can
// create in a rolling day.
//
// It is not a rate limit in disguise — the group's limiter already covers
// bursts. It bounds the lasting effect: a contributed row outlives the request
// that made it, and a moderator has to look at it. A number well above what
// anybody adds by hand in a day, and well below what makes the catalog somebody
// else's problem.
const maxContributedAssetsPerDay = 50

// ContributeAsset adds an asset to the shared catalog on a user's behalf, or
// hands back the one that is already there.
//
// The response does not say which of the two happened, deliberately: from the
// caller's side both mean "the ticker you asked for is now in your catalog",
// and distinguishing them would report on whether another user had already
// contributed that ticker.
func (s *Service) ContributeAsset(ctx context.Context, userID uuid.UUID, ticker, name string, assetType AssetType, exchange, currency string) (Asset, error) {
	input, err := normalizeAssetInput(ticker, name, assetType, exchange, currency)
	if err != nil {
		return Asset{}, err
	}

	contributed, err := s.repo.CountAssetsContributedBy(ctx, userID, time.Now().UTC().Add(-24*time.Hour))
	if err != nil {
		return Asset{}, err
	}

	if contributed >= maxContributedAssetsPerDay {
		return Asset{}, ErrAssetQuotaExceeded
	}

	return s.repo.CreateAssetIfAbsent(ctx, userID, input.ticker, input.name, input.assetType, input.exchange, input.currency)
}

type assetInput struct {
	ticker    string
	name      string
	assetType AssetType
	exchange  string
	currency  string
}

// normalizeAssetInput trims, upper-cases and validates what both creation paths
// receive. It lives in the service rather than the handler because the user
// path is no longer the only untrusted one: a request body and a spreadsheet
// row reach the same table, and the column limits have to hold for both.
func normalizeAssetInput(ticker, name string, assetType AssetType, exchange, currency string) (assetInput, error) {
	in := assetInput{
		ticker:    strings.ToUpper(strings.TrimSpace(ticker)),
		name:      strings.TrimSpace(name),
		assetType: assetType,
		exchange:  strings.TrimSpace(exchange),
		currency:  strings.ToUpper(strings.TrimSpace(currency)),
	}

	switch {
	case in.ticker == "":
		return assetInput{}, errAssetTickerRequired
	case len(in.ticker) > maxTickerLen:
		return assetInput{}, fmt.Errorf("%w: %d", errAssetTickerTooLong, maxTickerLen)
	}

	// A missing name is not worth a rejection: the ticker is what identifies
	// the asset, and the importer already falls back to it.
	if in.name == "" {
		in.name = in.ticker
	}

	if len(in.name) > maxAssetNameLen {
		in.name = in.name[:maxAssetNameLen]
	}

	if len(in.exchange) > maxExchangeLen {
		return assetInput{}, fmt.Errorf("%w: %d", errAssetExchangeTooLong, maxExchangeLen)
	}

	if !in.assetType.IsValid() {
		return assetInput{}, errAssetTypeInvalid
	}

	if len(in.currency) != 3 {
		return assetInput{}, errAssetCurrencyInvalid
	}

	return in, nil
}

func (s *Service) UpdateAssetPrice(ctx context.Context, assetID uuid.UUID, price money.Money) (Asset, error) {
	return s.repo.UpdateAssetPrice(ctx, assetID, price)
}
