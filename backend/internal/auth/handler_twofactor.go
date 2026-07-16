package auth

import (
	"errors"

	"github.com/gofiber/fiber/v3"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

// twoFactorLogin finishes a login whose password already checked out but
// whose account has 2FA enabled. Public endpoint, gated by the auth limiter.
func (h *handler) twoFactorLogin(c fiber.Ctx) error {
	var req TwoFactorLoginRequestDTO
	if err := c.Bind().Body(&req); err != nil {
		return httpx.BadRequest(c, "invalid request body", "auth:2fa:login")
	}

	result, err := h.service.CompleteTwoFactorLogin(c, req.Token, req.Code, c.IP(), c.Get("User-Agent"))
	if err != nil {
		if errors.Is(err, ErrTwoFactorPendingInvalid) {
			return httpx.ErrorAction(c, fiber.StatusUnauthorized, "two-factor session expired", "log in with your password again", "auth:2fa:expired")
		}

		if errors.Is(err, ErrTwoFactorInvalidCode) {
			return httpx.ErrorAction(c, fiber.StatusUnauthorized, "invalid two-factor code", "the code is incorrect or was already used", "auth:2fa:invalid-code")
		}

		return httpx.FromDomain(c, err, "failed to verify two-factor code", "auth:2fa:login")
	}

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    result.RawRefreshToken,
		Path:     "/",
		HTTPOnly: true,
		Secure:   h.cfg.Environment == "production",
		SameSite: "Strict",
		MaxAge:   int(h.cfg.JWTRefreshDuration.Seconds()),
	})

	return httpx.OK(c, "successfully logged in", "valid two-factor code",
		LoginResponseDTO{ID: result.ID, AccessToken: result.AccessToken})
}

// twoFactorStatus reports whether the authenticated user has 2FA enabled.
func (h *handler) twoFactorStatus(c fiber.Ctx) error {
	userID, _, _, err := getUserIDTokenRole(c)
	if err != nil {
		return httpx.BadRequest(c, "invalid user id", "auth:2fa:status")
	}

	status, err := h.service.TwoFactorStatus(c, userID)
	if err != nil {
		return httpx.FromDomain(c, err, "failed to get two-factor status", "auth:2fa:status")
	}

	return httpx.OK(c, "two-factor status retrieved", "valid access token", status)
}

// twoFactorSetup starts (or restarts) enrollment: it returns the secret and
// otpauth URL. 2FA stays disabled until the user confirms a code.
func (h *handler) twoFactorSetup(c fiber.Ctx) error {
	userID, _, _, err := getUserIDTokenRole(c)
	if err != nil {
		return httpx.BadRequest(c, "invalid user id", "auth:2fa:setup")
	}

	var req TwoFactorSetupRequestDTO
	if err := c.Bind().Body(&req); err != nil {
		return httpx.BadRequest(c, "invalid request body", "auth:2fa:setup")
	}

	setup, err := h.service.BeginTwoFactorSetup(c, userID, req.Password)
	if err != nil {
		if errors.Is(err, ErrTwoFactorAlreadyEnabled) {
			return httpx.ErrorAction(c, fiber.StatusConflict, "two-factor already enabled", "disable it first to enroll a new authenticator", "auth:2fa:already-enabled")
		}

		return httpx.FromDomain(c, err, "failed to start two-factor setup", "auth:2fa:setup")
	}

	return httpx.OK(c, "two-factor setup started", "scan the QR code and confirm with a code", setup)
}

// twoFactorEnable confirms the pending setup and returns the recovery codes.
func (h *handler) twoFactorEnable(c fiber.Ctx) error {
	userID, _, _, err := getUserIDTokenRole(c)
	if err != nil {
		return httpx.BadRequest(c, "invalid user id", "auth:2fa:enable")
	}

	var req TwoFactorEnableRequestDTO
	if err := c.Bind().Body(&req); err != nil {
		return httpx.BadRequest(c, "invalid request body", "auth:2fa:enable")
	}

	result, err := h.service.ConfirmTwoFactorSetup(c, userID, req.Code, c.IP(), c.Get("User-Agent"))
	if err != nil {
		switch {
		case errors.Is(err, ErrTwoFactorInvalidCode):
			return httpx.BadRequest(c, "invalid two-factor code", "auth:2fa:enable")
		case errors.Is(err, ErrTwoFactorSetupMissing):
			return httpx.BadRequest(c, "two-factor setup not started", "auth:2fa:enable")
		case errors.Is(err, ErrTwoFactorAlreadyEnabled):
			return httpx.BadRequest(c, "two-factor already enabled", "auth:2fa:enable")
		}

		return httpx.FromDomain(c, err, "failed to enable two-factor", "auth:2fa:enable")
	}

	return httpx.OK(c, "two-factor enabled", "store the recovery codes in a safe place", result)
}

// twoFactorDisable turns 2FA off after checking password + current code.
func (h *handler) twoFactorDisable(c fiber.Ctx) error {
	userID, _, _, err := getUserIDTokenRole(c)
	if err != nil {
		return httpx.BadRequest(c, "invalid user id", "auth:2fa:disable")
	}

	var req TwoFactorDisableRequestDTO
	if err := c.Bind().Body(&req); err != nil {
		return httpx.BadRequest(c, "invalid request body", "auth:2fa:disable")
	}

	if err := h.service.DisableTwoFactor(c, userID, req.Password, req.Code, c.IP(), c.Get("User-Agent")); err != nil {
		switch {
		case errors.Is(err, ErrTwoFactorInvalidCode):
			return httpx.BadRequest(c, "invalid two-factor code", "auth:2fa:disable")
		case errors.Is(err, ErrTwoFactorNotEnabled):
			return httpx.BadRequest(c, "two-factor not enabled", "auth:2fa:disable")
		}

		return httpx.FromDomain(c, err, "failed to disable two-factor", "auth:2fa:disable")
	}

	return httpx.OK(c, "two-factor disabled", "two-factor authentication was turned off", nil)
}

// twoFactorRecoveryCodes regenerates the recovery code batch.
func (h *handler) twoFactorRecoveryCodes(c fiber.Ctx) error {
	userID, _, _, err := getUserIDTokenRole(c)
	if err != nil {
		return httpx.BadRequest(c, "invalid user id", "auth:2fa:recovery")
	}

	var req TwoFactorDisableRequestDTO
	if err := c.Bind().Body(&req); err != nil {
		return httpx.BadRequest(c, "invalid request body", "auth:2fa:recovery")
	}

	result, err := h.service.RegenerateTwoFactorRecoveryCodes(c, userID, req.Password, req.Code)
	if err != nil {
		switch {
		case errors.Is(err, ErrTwoFactorInvalidCode):
			return httpx.BadRequest(c, "invalid two-factor code", "auth:2fa:recovery")
		case errors.Is(err, ErrTwoFactorNotEnabled):
			return httpx.BadRequest(c, "two-factor not enabled", "auth:2fa:recovery")
		}

		return httpx.FromDomain(c, err, "failed to regenerate recovery codes", "auth:2fa:recovery")
	}

	return httpx.OK(c, "recovery codes regenerated", "store the new codes; the old ones no longer work", result)
}
