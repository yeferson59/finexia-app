package market

import (
	"time"

	"github.com/yeferson59/finexia-app/internal/platform/marketdata"
	"github.com/yeferson59/finexia-app/internal/platform/secretbox"
)

// ProviderID identifies a market-data provider a user can bring their own key
// for. It is an alias, not a new type: the same identifier is persisted in
// market_credentials.provider, forms part of the AAD binding a sealed key to
// its row, and selects the client in marketdata/providers, so having one type
// across all three removes a class of conversion mistakes.
type ProviderID = marketdata.ProviderName

const (
	AlphaVantage = marketdata.AlphaVantage
	Finnhub      = marketdata.Finnhub
)

// SupportedProviders is the order a user's keys are tried in.
var SupportedProviders = marketdata.SupportedProviders

// CredentialStatus tracks whether a stored key still works. It is set when the
// user verifies a key and whenever the sync job gets a definitive answer from
// the provider, so a key that stopped working can be surfaced in the UI instead
// of silently producing no prices.
type CredentialStatus string

const (
	CredentialActive      CredentialStatus = "active"
	CredentialInvalid     CredentialStatus = "invalid"
	CredentialRateLimited CredentialStatus = "rate_limited"
)

// Credential is the only projection of a stored key that may leave the service
// layer. It has no field capable of carrying the API key or the ciphertext —
// that absence is the point, and it is what the handler tests pin down.
type Credential struct {
	Provider ProviderID `json:"provider"`
	// Last4 is the trailing fragment of the key, kept so the UI can show which
	// key is configured ("····3f9a") without holding the secret.
	Last4          string           `json:"last4"`
	Status         CredentialStatus `json:"status"`
	LastVerifiedAt *time.Time       `json:"lastVerifiedAt"`
	LastError      string           `json:"lastError,omitempty"`
	CreatedAt      time.Time        `json:"createdAt"`
	UpdatedAt      time.Time        `json:"updatedAt"`
}

// sealedCredential carries the encrypted key between the repository and the
// service. It is deliberately unexported: nothing outside this package can name
// the type, so a sealed key cannot end up in a DTO or a handler response by
// accident.
type sealedCredential struct {
	Provider ProviderID
	Sealed   secretbox.Sealed
}

// credentialAAD binds a sealed key to the user and provider it belongs to.
// Sealing and opening must agree on it, which is what makes a ciphertext copied
// into another user's row fail to decrypt.
func credentialAAD(userID, provider string) []byte {
	return secretbox.AAD(userID, provider)
}

// last4 returns the trailing fragment stored alongside the ciphertext. Short
// keys yield whatever they have; the value is cosmetic and never used to
// authenticate anything.
func last4(apiKey string) string {
	runes := []rune(apiKey)
	if len(runes) <= 4 {
		return string(runes)
	}

	return string(runes[len(runes)-4:])
}
