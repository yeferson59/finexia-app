package market

import (
	"github.com/gofiber/fiber/v3"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

// The responses below carry Credential values, which have no field for the API
// key or the ciphertext. That is the guarantee: there is no shape of this
// handler that can serve a key back, not even to its owner.

func (h *handler) ListCredentials(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid identity", err.Error())
	}

	creds, err := h.service.ListCredentials(c, userID)
	if err != nil {
		return httpx.FromDomain(c, err, "Error reading credentials", "Could not read your market data keys")
	}

	return httpx.OK(c, "", "", creds)
}

func (h *handler) SaveCredential(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid identity", err.Error())
	}

	provider := ProviderID(c.Params("provider"))

	var req SaveCredentialRequestDTO
	if err := c.Bind().JSON(&req); err != nil {
		return httpx.BadRequest(c, "Invalid request", err.Error())
	}

	cred, err := h.service.SaveCredential(c, userID, provider, req.APIKey)
	if err != nil {
		// err may name the provider but never the key: every provider error is
		// built through marketdata.Errorf, which scrubs it.
		return httpx.FromDomain(c, err, "Could not save the key", credentialFailureDetail(err))
	}

	return httpx.OK(c, "Key saved", "The key was verified against the provider and stored encrypted", cred)
}

func (h *handler) VerifyCredential(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid identity", err.Error())
	}

	cred, err := h.service.VerifyCredential(c, userID, ProviderID(c.Params("provider")))
	if err != nil {
		return httpx.FromDomain(c, err, "Could not verify the key", credentialFailureDetail(err))
	}

	return httpx.OK(c, "Key verified", "", cred)
}

func (h *handler) DeleteCredential(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid identity", err.Error())
	}

	if err := h.service.DeleteCredential(c, userID, ProviderID(c.Params("provider"))); err != nil {
		return httpx.FromDomain(c, err, "Could not delete the key", "The key could not be deleted")
	}

	return httpx.OK(c, "Key deleted", "", nil)
}

// SyncMarketData refreshes the caller's own holdings with the caller's own
// keys. It replaces the admin-only global sync, which no longer has a key to
// run under.
func (h *handler) SyncMarketData(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid identity", err.Error())
	}

	assetIDs, err := h.holdings.HeldAssetIDs(c, userID)
	if err != nil {
		return httpx.FromDomain(c, err, "Could not read your holdings", "")
	}

	prices, errs := h.service.SyncAssetsForUser(c, userID, assetIDs)
	if len(errs) > 0 && len(prices) == 0 {
		return httpx.FromDomain(c, errs[0], "Market data sync failed", credentialFailureDetail(errs[0]))
	}

	return httpx.OK(c, "Market data synced", "", prices)
}

// credentialFailureDetail keeps provider text out of the response body unless
// it is one of our own domain errors.
//
// The old handlers returned errs[0].Error() straight to the client. With the
// operator's key that was merely untidy; with a user's key it would be a way to
// echo provider output — and the key rides in Alpha Vantage's URL — back over
// the wire. Only the messages this package authored are returned now.
func credentialFailureDetail(err error) string {
	switch {
	case err == nil:
		return ""
	case isDomainCredentialError(err):
		return err.Error()
	default:
		return "The market data provider could not be reached. Try again in a few minutes."
	}
}
