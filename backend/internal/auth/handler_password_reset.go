package auth

import (
	"errors"

	"github.com/gofiber/fiber/v3"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

// requestPasswordReset (public) emails a reset link if the address belongs to
// an account. The response is identical whether or not the email exists, so
// the endpoint never confirms which addresses are registered.
func (h *handler) requestPasswordReset(c fiber.Ctx) error {
	var req RequestPasswordResetDTO
	if err := c.Bind().Body(&req); err != nil {
		return httpx.BadRequest(c, "invalid request body", "auth:passwordReset:request")
	}

	if err := h.service.RequestPasswordReset(c, req.Email); err != nil {
		return httpx.FromDomain(c, err, "failed to process request", "auth:passwordReset:request")
	}

	return httpx.OK(c, "if the email exists, a reset link was sent",
		"password reset requested", nil)
}

// validatePasswordReset (public) reports whether a token is still redeemable,
// so the reset page can reject dead links before asking for a new password.
func (h *handler) validatePasswordReset(c fiber.Ctx) error {
	token := c.Query("token")

	if err := h.service.ValidatePasswordResetToken(c, token); err != nil {
		return h.passwordResetError(c, err, "auth:passwordReset:validate")
	}

	return httpx.OK(c, "token valid", "password reset token is valid", nil)
}

// confirmPasswordReset (public) consumes a valid token and sets the new
// password, revoking every existing session.
func (h *handler) confirmPasswordReset(c fiber.Ctx) error {
	var req ConfirmPasswordResetDTO
	if err := c.Bind().Body(&req); err != nil {
		return httpx.BadRequest(c, "invalid request body", "auth:passwordReset:confirm")
	}

	if err := h.service.ResetPassword(c, req.Token, req.Password, c.IP(), c.Get("User-Agent")); err != nil {
		return h.passwordResetError(c, err, "auth:passwordReset:confirm")
	}

	return httpx.OK(c, "password updated", "password reset successfully", nil)
}

// passwordResetError maps the password-reset sentinels to precise statuses;
// anything else falls through to the shared domain-error mapping.
func (h *handler) passwordResetError(c fiber.Ctx, err error, action string) error {
	switch {
	case errors.Is(err, ErrPasswordResetExpired):
		return httpx.ErrorAction(c, fiber.StatusGone, "password reset link expired", "the link has expired; request a new one", action)
	case errors.Is(err, ErrPasswordResetInvalid):
		return httpx.BadRequest(c, "invalid password reset link", "the link is invalid or has already been used")
	default:
		return httpx.FromDomain(c, err, "failed to process password reset", action)
	}
}
