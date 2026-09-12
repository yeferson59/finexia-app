package market

import (
	"context"
	"fmt"
	"strings"
	"time"

	"uuid"

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

func (s *service) GetAssets(ctx context.Context, view CatalogView, offset, limit uint) ([]Asset, error) {
	return s.repo.GetAssets(ctx, view, offset, limit)
}

func (s *service) SearchAssets(ctx context.Context, view CatalogView, search string, offset, limit uint) ([]Asset, error) {
	return s.repo.SearchAssets(ctx, view, search, offset, limit)
}

func (s *service) GetAssetByID(ctx context.Context, assetID uuid.UUID) (Asset, error) {
	return s.repo.GetAssetByID(ctx, assetID)
}

// CreateAsset curates a catalog row. Operator-only, and the seed runs through
// it too.
func (s *service) CreateAsset(ctx context.Context, spec AssetSpec) (Asset, error) {
	input, err := normalizeAssetSpec(spec)
	if err != nil {
		return Asset{}, err
	}

	return s.repo.UpsertAsset(ctx, input)
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
//
// The sector is not among its arguments, unlike CreateAsset's. A contribution
// names an instrument the catalog was missing, and the user adding COLCAP knows
// its ticker and currency; asking them to also file it under an industry would
// be asking for a guess that every other holder of that row then inherits. The
// asset arrives unclassified, which the breakdown reports as such.
func (s *service) ContributeAsset(ctx context.Context, userID uuid.UUID, ticker, name string, assetType AssetType, exchange string, currency money.Currency) (Asset, error) {
	input, err := normalizeAssetSpec(AssetSpec{
		Ticker:    ticker,
		Name:      name,
		AssetType: assetType,
		Exchange:  exchange,
		Currency:  currency,
	})
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

	return s.repo.CreateAssetIfAbsent(ctx, userID, input.Ticker, input.Name, input.AssetType, input.Exchange, input.Currency)
}

// normalizeAssetSpec trims, upper-cases and validates what every write path
// receives. It lives in the service rather than the handler because the user
// path is no longer the only untrusted one: a request body, a spreadsheet row
// and an operator's edit reach the same table, and the column limits have to
// hold for all three.
func normalizeAssetSpec(spec AssetSpec) (AssetSpec, error) {
	if !spec.Currency.Valid() {
		return AssetSpec{}, errAssetCurrencyInvalid
	}

	in := AssetSpec{
		Ticker:        strings.ToUpper(strings.TrimSpace(spec.Ticker)),
		Name:          strings.TrimSpace(spec.Name),
		AssetType:     spec.AssetType,
		Exchange:      strings.TrimSpace(spec.Exchange),
		Currency:      spec.Currency,
		Sector:        spec.Sector,
		SectorWeights: spec.SectorWeights,
	}

	switch {
	case in.Ticker == "":
		return AssetSpec{}, errAssetTickerRequired
	case len(in.Ticker) > maxTickerLen:
		return AssetSpec{}, fmt.Errorf("%w: %d", errAssetTickerTooLong, maxTickerLen)
	}

	// A missing name is not worth a rejection: the ticker is what identifies
	// the asset, and the importer already falls back to it.
	if in.Name == "" {
		in.Name = in.Ticker
	}

	if len(in.Name) > maxAssetNameLen {
		in.Name = in.Name[:maxAssetNameLen]
	}

	if len(in.Exchange) > maxExchangeLen {
		return AssetSpec{}, fmt.Errorf("%w: %d", errAssetExchangeTooLong, maxExchangeLen)
	}

	if !in.AssetType.IsValid() {
		return AssetSpec{}, errAssetTypeInvalid
	}

	// A sector is optional, so only a value that is neither empty nor known is
	// an error. Classifying an asset whose type has none — a coin, a cash
	// balance — is rejected rather than silently stored: the breakdown reads
	// those from the type, so a sector written there would be a number nothing
	// displays and the next reader would have to explain.
	switch {
	case in.Sector == SectorNone:
	case !in.Sector.IsValid():
		return AssetSpec{}, errAssetSectorInvalid
	case !in.AssetType.HasSector():
		return AssetSpec{}, errAssetSectorNotApplicable
	}

	// The breakdown is the same field in a different shape, so it answers to the
	// same two rules — it only goes on something with a business behind it, and
	// it cannot sit beside a single sector — plus its own arithmetic. Checked
	// here and not only at the database because the constraint that matters most
	// is the one no column can express: that these eleven numbers are a reading
	// of one asset and not of an asset and a half.
	if !in.SectorWeights.IsEmpty() {
		switch {
		case in.Sector != SectorNone:
			return AssetSpec{}, errAssetSectorBoth
		case !in.AssetType.HasSector():
			return AssetSpec{}, errAssetSectorNotApplicable
		}

		if err := in.SectorWeights.Validate(); err != nil {
			return AssetSpec{}, err
		}
	}

	return in, nil
}

// UpdateAsset rewrites a catalog row. Operator-only, and the one path that can
// change what an existing asset says about itself.
//
// Everything a create validates, an edit validates the same way: the row ends up
// in the same table with the same column limits, and the ticker it is renamed to
// is as untrusted as the one it was created with.
//
// The currency is the field with a consequence beyond itself. assets stores the
// manual price as a bare numeric and reads its currency from this column, so
// re-denominating an asset silently reinterprets whatever price is there — 190
// dollars becomes 190 pesos. When the edit carries a new price that is not a
// problem, the two are written together; when it does not, the repository drops
// the stale number rather than let the catalog quote a figure in a currency
// nobody entered it in.
func (s *service) UpdateAsset(ctx context.Context, assetID uuid.UUID, upd AssetUpdate) (Asset, error) {
	input, err := normalizeAssetSpec(AssetSpec{
		Ticker:        upd.Ticker,
		Name:          upd.Name,
		AssetType:     upd.AssetType,
		Exchange:      upd.Exchange,
		Currency:      upd.Currency,
		Sector:        upd.Sector,
		SectorWeights: upd.SectorWeights,
	})
	if err != nil {
		return Asset{}, err
	}

	if upd.Price != nil && !upd.Price.IsPositive() {
		return Asset{}, errAssetPriceInvalid
	}

	return s.repo.UpdateAsset(ctx, assetID, AssetUpdate{
		Ticker:        input.Ticker,
		Name:          input.Name,
		AssetType:     input.AssetType,
		Exchange:      input.Exchange,
		Currency:      input.Currency,
		Sector:        input.Sector,
		SectorWeights: input.SectorWeights,
		IsCurated:     upd.IsCurated,
		Price:         upd.Price,
	})
}

func (s *service) UpdateAssetPrice(ctx context.Context, assetID uuid.UUID, price money.Money) (Asset, error) {
	return s.repo.UpdateAssetPrice(ctx, assetID, price)
}
