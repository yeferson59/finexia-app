package market

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/yeferson59/finexia-app/internal/platform/logger"
	"github.com/yeferson59/finexia-app/internal/platform/marketdata"
	"github.com/yeferson59/finexia-app/internal/platform/secretbox"
)

// verifySymbol is a widely covered ticker used to prove a key works before it
// is stored. It costs one request against the user's own quota, which is worth
// it: the alternative is a key that looks saved and silently produces nothing
// until the next morning's sync.
const verifySymbol = "AAPL"

// ListCredentials returns what a user may see about their own keys: never the
// key, never the ciphertext.
func (s *Service) ListCredentials(ctx context.Context, userID uuid.UUID) ([]Credential, error) {
	return s.repo.ListCredentials(ctx, userID)
}

// SaveCredential verifies an API key against its provider and, only if the
// provider accepts it, seals and stores it.
//
// The plaintext key lives in this function and in the provider client it builds,
// and nowhere else: it is never logged, never returned, and never written
// unsealed.
func (s *Service) SaveCredential(ctx context.Context, userID uuid.UUID, provider ProviderID, apiKey string) (Credential, error) {
	if !provider.IsValid() {
		return Credential{}, ErrInvalidProvider
	}

	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return Credential{}, ErrInvalidAPIKey
	}

	if err := s.probe(ctx, provider, apiKey); err != nil {
		return Credential{}, err
	}

	sealed, err := s.seal(userID, provider, apiKey)
	if err != nil {
		return Credential{}, err
	}

	now := time.Now().UTC()

	return s.repo.UpsertCredential(ctx, userID, sealed, last4(apiKey), &now)
}

func (s *Service) DeleteCredential(ctx context.Context, userID uuid.UUID, provider ProviderID) error {
	if !provider.IsValid() {
		return ErrInvalidProvider
	}

	return s.repo.DeleteCredential(ctx, userID, provider)
}

// VerifyCredential re-checks a stored key and records the verdict, so a key
// that was revoked at the provider shows up as invalid in the UI instead of
// just producing no prices.
func (s *Service) VerifyCredential(ctx context.Context, userID uuid.UUID, provider ProviderID) (Credential, error) {
	if !provider.IsValid() {
		return Credential{}, ErrInvalidProvider
	}

	apiKey, err := s.openCredential(ctx, userID, provider)
	if err != nil {
		return Credential{}, err
	}
	defer secretbox.Zero(apiKey)

	status, statusErr := CredentialActive, ""
	if err := s.probe(ctx, provider, string(apiKey)); err != nil {
		status, statusErr = statusFor(err), err.Error()
	}

	if err := s.repo.SetCredentialStatus(ctx, userID, provider, status, statusErr); err != nil {
		return Credential{}, err
	}

	creds, err := s.repo.ListCredentials(ctx, userID)
	if err != nil {
		return Credential{}, err
	}
	for _, c := range creds {
		if c.Provider == provider {
			return c, nil
		}
	}

	return Credential{}, ErrCredentialNotFound
}

// probe spends one request to confirm the provider accepts the key. A symbol
// the provider simply does not cover still proves the key itself is good, so
// ErrUnsupported counts as success.
func (s *Service) probe(ctx context.Context, provider ProviderID, apiKey string) error {
	chain, err := s.providers.For([]marketdata.Credential{{Provider: provider, APIKey: apiKey}})
	if err != nil {
		return err
	}

	if _, err := chain.FetchQuote(ctx, verifySymbol); err != nil {
		if errors.Is(err, marketdata.ErrUnsupported) || errors.Is(err, marketdata.ErrRateLimited) {
			return nil
		}

		return ErrInvalidAPIKey
	}

	return nil
}

// seal encrypts the key for storage, bound to the user and provider that own it.
func (s *Service) seal(userID uuid.UUID, provider ProviderID, apiKey string) (sealedCredential, error) {
	if s.keyring == nil {
		return sealedCredential{}, ErrKeyEncryptionOff
	}

	sealed, err := s.keyring.Seal([]byte(apiKey), credentialAAD(userID.String(), string(provider)))
	if err != nil {
		return sealedCredential{}, fmt.Errorf("seal credential: %w", err)
	}

	return sealedCredential{Provider: provider, Sealed: sealed}, nil
}

// openCredential decrypts one stored key. The caller owns the returned bytes
// and must Zero them.
func (s *Service) openCredential(ctx context.Context, userID uuid.UUID, provider ProviderID) ([]byte, error) {
	if s.keyring == nil {
		return nil, ErrKeyEncryptionOff
	}

	stored, err := s.repo.GetSealedCredential(ctx, userID, provider)
	if err != nil {
		return nil, err
	}

	return s.keyring.Open(stored.Sealed, credentialAAD(userID.String(), string(provider)))
}

// providerFor assembles the chain for one user from their stored keys. The
// plaintext keys are zeroed before returning: the clients hold their own copy
// for the duration of the sync, and nothing else needs them.
//
// It also reports the pace to leave between calls, taken from the slowest
// provider the user actually configured.
func (s *Service) providerFor(ctx context.Context, userID uuid.UUID) (marketdata.Provider, time.Duration, error) {
	if s.keyring == nil {
		return nil, 0, ErrKeyEncryptionOff
	}

	stored, err := s.repo.GetSealedCredentials(ctx, userID)
	if err != nil {
		return nil, 0, err
	}
	if len(stored) == 0 {
		return nil, 0, ErrNoCredentials
	}

	creds := make([]marketdata.Credential, 0, len(stored))
	pace := finnhubPace

	for _, sc := range stored {
		plain, err := s.keyring.Open(sc.Sealed, credentialAAD(userID.String(), string(sc.Provider)))
		if err != nil {
			// One unreadable row (e.g. its KEK was retired) must not sink the
			// keys that still work.
			s.log.Error(ctx, "cannot open stored credential", logger.Err(err), logger.Str("provider", string(sc.Provider)))
			continue
		}

		creds = append(creds, marketdata.Credential{Provider: sc.Provider, APIKey: string(plain)})
		secretbox.Zero(plain)

		if sc.Provider == AlphaVantage {
			pace = alphaVantagePace
		}
	}

	chain, err := s.providers.For(creds)
	if err != nil {
		return nil, 0, err
	}

	return chain, pace, nil
}

// statusFor maps a provider failure onto the status stored against the key.
// Only a rejection of the key itself marks it invalid; a spent quota leaves it
// active so tomorrow's run tries again.
func statusFor(err error) CredentialStatus {
	switch {
	case errors.Is(err, marketdata.ErrUnauthorized), errors.Is(err, ErrInvalidAPIKey):
		return CredentialInvalid
	case errors.Is(err, marketdata.ErrRateLimited):
		return CredentialRateLimited
	default:
		return CredentialActive
	}
}
