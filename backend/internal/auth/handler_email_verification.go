package auth

import (
	"errors"

	"github.com/gofiber/fiber/v3"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

// requestEmailVerification (public) (re)sends a verification link if the
// address belongs to an unverified account. The response is identical
// whether or not the email exists or is already verified, so the endpoint
// never confirms which addresses are registered.
func (h *handler) requestEmailVerification(c fiber.Ctx) error {
	var req RequestEmailVerificationDTO
	if err := c.Bind().Body(&req); err != nil {
		return httpx.BadRequest(c, "invalid request body", "auth:emailVerification:request")
	}

	if err := h.service.RequestEmailVerification(c, req.Email); err != nil {
		return httpx.FromDomain(c, err, "failed to process request", "auth:emailVerification:request")
	}

	return httpx.OK(c, "if the email exists and is unverified, a verification link was sent",
		"email verification requested", nil)
}

// validateEmailVerification (public) reports whether a token is still
// redeemable, so the verify page can reject dead links before confirming.
func (h *handler) validateEmailVerification(c fiber.Ctx) error {
	token := c.Query("token")

	if err := h.service.ValidateEmailVerification(c, token); err != nil {
		return h.emailVerificationError(c, err, "auth:emailVerification:validate")
	}

	return httpx.OK(c, "token valid", "email verification token is valid", nil)
}

// confirmEmailVerification (public) consumes a valid token and marks the
// account's email as verified.
func (h *handler) confirmEmailVerification(c fiber.Ctx) error {
	var req ConfirmEmailVerificationDTO
	if err := c.Bind().Body(&req); err != nil {
		return httpx.BadRequest(c, "invalid request body", "auth:emailVerification:confirm")
	}

	if err := h.service.VerifyEmail(c, req.Token); err != nil {
		return h.emailVerificationError(c, err, "auth:emailVerification:confirm")
	}

	return httpx.OK(c, "email verified", "email verified successfully", nil)
}

// emailVerificationError maps the email-verification sentinels to precise
// statuses; anything else falls through to the shared domain-error mapping.
func (h *handler) emailVerificationError(c fiber.Ctx, err error, action string) error {
	switch {
	case errors.Is(err, ErrEmailVerificationExpired):
		return httpx.ErrorAction(c, fiber.StatusGone, "email verification link expired", "the link has expired; request a new one", action)
	case errors.Is(err, ErrEmailVerificationInvalid):
		return httpx.BadRequest(c, "invalid email verification link", "the link is invalid or has already been used")
	default:
		return httpx.FromDomain(c, err, "failed to process email verification", action)
	}
}
