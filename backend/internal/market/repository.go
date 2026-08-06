package market

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"
)

type Repository interface {
	// Exchange rates
	UpsertExchangeRate(ctx context.Context, from, to string, rate decimal.Decimal, rateDate time.Time) (ExchangeRate, error)
	GetExchangeRates(ctx context.Context, offset, limit uint) ([]ExchangeRate, error)
	UpdateExchangeRateByID(ctx context.Context, id uuid.UUID, rate decimal.Decimal) (ExchangeRate, error)

	// Assets (catalog owned by this module; portfolio reads them via AssetReader)
	GetAssetByID(ctx context.Context, assetID uuid.UUID) (Asset, error)
	GetAssets(ctx context.Context, view CatalogView, offset, limit uint) ([]Asset, error)
	SearchAssets(ctx context.Context, view CatalogView, search string, offset, limit uint) ([]Asset, error)
	// UpsertAsset writes a curated row, overwriting the metadata of an existing
	// one. Operator-only: the overwrite is what makes it unsafe to expose to
	// users, since the ticker is the conflict target and any user could reach
	// somebody else's row through it.
	UpsertAsset(ctx context.Context, ticker, name string, assetType AssetType, exchange, currency string) (Asset, error)
	// CreateAssetIfAbsent is the user-facing counterpart: it inserts only when
	// the ticker is new and otherwise returns the existing row untouched, so a
	// contribution can never rewrite curated data. Either way the asset joins
	// the user's own catalog.
	CreateAssetIfAbsent(ctx context.Context, userID uuid.UUID, ticker, name string, assetType AssetType, exchange, currency string) (Asset, error)
	// CountAssetsContributedBy bounds how much one user can add per day.
	CountAssetsContributedBy(ctx context.Context, userID uuid.UUID, since time.Time) (int, error)
	UpdateAssetPrice(ctx context.Context, assetID uuid.UUID, price money.Money) (Asset, error)

	CredentialStore
}

// CredentialStore holds each user's own provider keys, sealed, plus the prices
// and rates those keys fetched. Both halves live together because they share
// one invariant: everything here is scoped to the user who owns the key, and
// none of it may be served to anybody else.
//
// The sealed material is only ever named by the unexported sealedCredential
// type, so no caller outside this package can hold a ciphertext.
type CredentialStore interface {
	UpsertCredential(ctx context.Context, userID uuid.UUID, cred sealedCredential, keyLast4 string, status CredentialStatus, verifiedAt *time.Time) (Credential, error)
	ListCredentials(ctx context.Context, userID uuid.UUID) ([]Credential, error)
	GetSealedCredentials(ctx context.Context, userID uuid.UUID) ([]sealedCredential, error)
	GetSealedCredential(ctx context.Context, userID uuid.UUID, provider ProviderID) (sealedCredential, error)
	DeleteCredential(ctx context.Context, userID uuid.UUID, provider ProviderID) error
	SetCredentialStatus(ctx context.Context, userID uuid.UUID, provider ProviderID, status CredentialStatus, lastErr string) error
	UsersWithCredentials(ctx context.Context) ([]uuid.UUID, error)

	UpsertUserAssetPrice(ctx context.Context, userID, assetID uuid.UUID, price money.Money, currency string, source ProviderID, fetchedAt time.Time) error
	UpsertUserExchangeRate(ctx context.Context, userID uuid.UUID, from, to string, rate decimal.Decimal, source ProviderID, fetchedAt time.Time) error
}
