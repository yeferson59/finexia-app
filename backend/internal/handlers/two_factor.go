package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v3"

	"github.com/yeferson59/finexia-app/internal/dtos/auth"
	"github.com/yeferson59/finexia-app/internal/services"
)

// TwoFactorLogin finishes a login whose password already checked out but
// whose account has 2FA enabled. Public endpoint, gated by the auth limiter.
func (handler *Handlers) TwoFactorLogin(c fiber.Ctx) error {
	var req auth.TwoFactorLoginRequestDTO
	if err := c.Bind().Body(&req); err != nil {
		return handler.responseBadRequest(c, "invalid request body", "auth:2fa:login")
	}

	result, err := handler.services.CompleteTwoFactorLogin(c, req.Token, req.Code, c.IP(), c.Get("User-Agent"))
	if err != nil {
		if errors.Is(err, services.ErrTwoFactorPendingInvalid) {
			// The pending login expired or burned its attempts: the client
			// must restart from the password step.
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "two-factor session expired",
				"details": "log in with your password again",
				"action":  "auth:2fa:expired",
			})
		}
		if errors.Is(err, services.ErrTwoFactorInvalidCode) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "invalid two-factor code",
				"details": "the code is incorrect or was already used",
				"action":  "auth:2fa:invalid-code",
			})
		}
		return handler.responseFromDomain(c, err, "failed to verify two-factor code", "auth:2fa:login")
	}

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    result.RawRefreshToken,
		Path:     "/",
		HTTPOnly: true,
		Secure:   handler.cfg.Environment == "production",
		SameSite: "Strict",
		MaxAge:   int(handler.cfg.JWTRefreshDuration.Seconds()),
	})

	return handler.responseStatusOk(c, "successfully logged in", "valid two-factor code",
		auth.LoginResponseDTO{ID: result.ID, AccessToken: result.AccessToken})
}

// TwoFactorStatus reports whether the authenticated user has 2FA enabled.
func (handler *Handlers) TwoFactorStatus(c fiber.Ctx) error {
	userID, _, _, err := handler.getUserIDTokenRole(c)
	if err != nil {
		return handler.responseBadRequest(c, "invalid user id", "auth:2fa:status")
	}

	status, err := handler.services.TwoFactorStatus(c, userID)
	if err != nil {
		return handler.responseFromDomain(c, err, "failed to get two-factor status", "auth:2fa:status")
	}

	return handler.responseStatusOk(c, "two-factor status retrieved", "valid access token", status)
}

// TwoFactorSetup starts (or restarts) enrollment: it returns the secret and
// otpauth URL. 2FA stays disabled until the user confirms a code.
func (handler *Handlers) TwoFactorSetup(c fiber.Ctx) error {
	userID, _, _, err := handler.getUserIDTokenRole(c)
	if err != nil {
		return handler.responseBadRequest(c, "invalid user id", "auth:2fa:setup")
	}

	var req auth.TwoFactorSetupRequestDTO
	if err := c.Bind().Body(&req); err != nil {
		return handler.responseBadRequest(c, "invalid request body", "auth:2fa:setup")
	}

	setup, err := handler.services.BeginTwoFactorSetup(c, userID, req.Password)
	if err != nil {
		if errors.Is(err, services.ErrTwoFactorAlreadyEnabled) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"success": false,
				"message": "two-factor already enabled",
				"details": "disable it first to enroll a new authenticator",
				"action":  "auth:2fa:already-enabled",
			})
		}
		return handler.responseFromDomain(c, err, "failed to start two-factor setup", "auth:2fa:setup")
	}

	return handler.responseStatusOk(c, "two-factor setup started", "scan the QR code and confirm with a code", setup)
}

// TwoFactorEnable confirms the pending setup and returns the recovery codes.
func (handler *Handlers) TwoFactorEnable(c fiber.Ctx) error {
	userID, _, _, err := handler.getUserIDTokenRole(c)
	if err != nil {
		return handler.responseBadRequest(c, "invalid user id", "auth:2fa:enable")
	}

	var req auth.TwoFactorEnableRequestDTO
	if err := c.Bind().Body(&req); err != nil {
		return handler.responseBadRequest(c, "invalid request body", "auth:2fa:enable")
	}

	result, err := handler.services.ConfirmTwoFactorSetup(c, userID, req.Code, c.IP(), c.Get("User-Agent"))
	if err != nil {
		switch {
		case errors.Is(err, services.ErrTwoFactorInvalidCode):
			return handler.responseBadRequest(c, "invalid two-factor code", "auth:2fa:enable")
		case errors.Is(err, services.ErrTwoFactorSetupMissing):
			return handler.responseBadRequest(c, "two-factor setup not started", "auth:2fa:enable")
		case errors.Is(err, services.ErrTwoFactorAlreadyEnabled):
			return handler.responseBadRequest(c, "two-factor already enabled", "auth:2fa:enable")
		}
		return handler.responseFromDomain(c, err, "failed to enable two-factor", "auth:2fa:enable")
	}

	return handler.responseStatusOk(c, "two-factor enabled", "store the recovery codes in a safe place", result)
}

// TwoFactorDisable turns 2FA off after checking password + current code.
func (handler *Handlers) TwoFactorDisable(c fiber.Ctx) error {
	userID, _, _, err := handler.getUserIDTokenRole(c)
	if err != nil {
		return handler.responseBadRequest(c, "invalid user id", "auth:2fa:disable")
	}

	var req auth.TwoFactorDisableRequestDTO
	if err := c.Bind().Body(&req); err != nil {
		return handler.responseBadRequest(c, "invalid request body", "auth:2fa:disable")
	}

	if err := handler.services.DisableTwoFactor(c, userID, req.Password, req.Code, c.IP(), c.Get("User-Agent")); err != nil {
		switch {
		case errors.Is(err, services.ErrTwoFactorInvalidCode):
			return handler.responseBadRequest(c, "invalid two-factor code", "auth:2fa:disable")
		case errors.Is(err, services.ErrTwoFactorNotEnabled):
			return handler.responseBadRequest(c, "two-factor not enabled", "auth:2fa:disable")
		}
		return handler.responseFromDomain(c, err, "failed to disable two-factor", "auth:2fa:disable")
	}

	return handler.responseStatusOk(c, "two-factor disabled", "two-factor authentication was turned off", nil)
}

// TwoFactorRecoveryCodes regenerates the recovery code batch.
func (handler *Handlers) TwoFactorRecoveryCodes(c fiber.Ctx) error {
	userID, _, _, err := handler.getUserIDTokenRole(c)
	if err != nil {
		return handler.responseBadRequest(c, "invalid user id", "auth:2fa:recovery")
	}

	var req auth.TwoFactorDisableRequestDTO
	if err := c.Bind().Body(&req); err != nil {
		return handler.responseBadRequest(c, "invalid request body", "auth:2fa:recovery")
	}

	result, err := handler.services.RegenerateTwoFactorRecoveryCodes(c, userID, req.Password, req.Code)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrTwoFactorInvalidCode):
			return handler.responseBadRequest(c, "invalid two-factor code", "auth:2fa:recovery")
		case errors.Is(err, services.ErrTwoFactorNotEnabled):
			return handler.responseBadRequest(c, "two-factor not enabled", "auth:2fa:recovery")
		}
		return handler.responseFromDomain(c, err, "failed to regenerate recovery codes", "auth:2fa:recovery")
	}

	return handler.responseStatusOk(c, "recovery codes regenerated", "store the new codes; the old ones no longer work", result)
}
