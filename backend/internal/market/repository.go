package market

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/yeferson59/gofinance/v2/money"
)

type Repository interface {
	// Exchange rates
	UpsertExchangeRate(ctx context.Context, from, to string, rate money.Decimal, rateDate time.Time) (ExchangeRate, error)
	GetExchangeRates(ctx context.Context, offset, limit uint) ([]ExchangeRate, error)
	UpdateExchangeRateByID(ctx context.Context, id uuid.UUID, rate money.Decimal) (ExchangeRate, error)

	// Assets (catalog owned by this module; portfolio reads them via AssetReader)
	GetAssetByID(ctx context.Context, assetID uuid.UUID) (Asset, error)
	GetAssets(ctx context.Context, offset, limit uint) ([]Asset, error)
	SearchAssets(ctx context.Context, search string, offset, limit uint) ([]Asset, error)
	UpsertAsset(ctx context.Context, ticker, name string, assetType AssetType, exchange, currency string) (Asset, error)
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
	UpsertUserExchangeRate(ctx context.Context, userID uuid.UUID, from, to string, rate money.Decimal, source ProviderID, fetchedAt time.Time) error
}
