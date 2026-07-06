package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v3"

	"github.com/yeferson59/finexia-app/internal/dtos/auth"
	"github.com/yeferson59/finexia-app/internal/services"
)

// RequestPasswordReset (public) emails a reset link if the address belongs to
// an account. The response is identical whether or not the email exists, so
// the endpoint never confirms which addresses are registered.
func (handler *Handlers) RequestPasswordReset(c fiber.Ctx) error {
	var req auth.RequestPasswordResetDTO
	if err := c.Bind().Body(&req); err != nil {
		return handler.responseBadRequest(c, "invalid request body", "auth:passwordReset:request")
	}

	if err := handler.services.RequestPasswordReset(c.Context(), req.Email); err != nil {
		return handler.responseFromDomain(c, err, "failed to process request", "auth:passwordReset:request")
	}

	return handler.responseStatusOk(c, "if the email exists, a reset link was sent",
		"password reset requested", nil)
}

// ValidatePasswordReset (public) reports whether a token is still redeemable,
// so the reset page can reject dead links before asking for a new password.
func (handler *Handlers) ValidatePasswordReset(c fiber.Ctx) error {
	token := c.Query("token")

	if err := handler.services.ValidatePasswordResetToken(c.Context(), token); err != nil {
		return handler.passwordResetError(c, err, "auth:passwordReset:validate")
	}

	return handler.responseStatusOk(c, "token valid", "password reset token is valid", nil)
}

// ConfirmPasswordReset (public) consumes a valid token and sets the new
// password, revoking every existing session.
func (handler *Handlers) ConfirmPasswordReset(c fiber.Ctx) error {
	var req auth.ConfirmPasswordResetDTO
	if err := c.Bind().Body(&req); err != nil {
		return handler.responseBadRequest(c, "invalid request body", "auth:passwordReset:confirm")
	}

	if err := handler.services.ResetPassword(c.Context(), req.Token, req.Password, c.IP(), c.Get("User-Agent")); err != nil {
		return handler.passwordResetError(c, err, "auth:passwordReset:confirm")
	}

	return handler.responseStatusOk(c, "password updated", "password reset successfully", nil)
}

// passwordResetError maps the password-reset sentinels to precise statuses;
// anything else falls through to the shared domain-error mapping.
func (handler *Handlers) passwordResetError(c fiber.Ctx, err error, action string) error {
	switch {
	case errors.Is(err, services.ErrPasswordResetExpired):
		return c.Status(fiber.StatusGone).JSON(fiber.Map{
			"success": false,
			"message": "password reset link expired",
			"details": "the link has expired; request a new one",
			"action":  action,
		})
	case errors.Is(err, services.ErrPasswordResetInvalid):
		return handler.responseBadRequest(c, "invalid password reset link", "the link is invalid or has already been used")
	default:
		return handler.responseFromDomain(c, err, "failed to process password reset", action)
	}
}
