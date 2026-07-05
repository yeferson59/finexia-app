package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v3"

	"github.com/yeferson59/finexia-app/internal/dtos/auth"
	"github.com/yeferson59/finexia-app/internal/services"
)

// RequestEmailVerification (public) (re)sends a verification link if the
// address belongs to an unverified account. The response is identical
// whether or not the email exists or is already verified, so the endpoint
// never confirms which addresses are registered.
func (handler *Handlers) RequestEmailVerification(c fiber.Ctx) error {
	var req auth.RequestEmailVerificationDTO
	if err := c.Bind().Body(&req); err != nil {
		return handler.responseBadRequest(c, "invalid request body", "auth:emailVerification:request")
	}

	if err := handler.services.RequestEmailVerification(c.Context(), req.Email); err != nil {
		return handler.responseFromDomain(c, err, "failed to process request", "auth:emailVerification:request")
	}

	return handler.responseStatusOk(c, "if the email exists and is unverified, a verification link was sent",
		"email verification requested", nil)
}

// ValidateEmailVerification (public) reports whether a token is still
// redeemable, so the verify page can reject dead links before confirming.
func (handler *Handlers) ValidateEmailVerification(c fiber.Ctx) error {
	token := c.Query("token")

	if err := handler.services.ValidateEmailVerification(c.Context(), token); err != nil {
		return handler.emailVerificationError(c, err, "auth:emailVerification:validate")
	}

	return handler.responseStatusOk(c, "token valid", "email verification token is valid", nil)
}

// ConfirmEmailVerification (public) consumes a valid token and marks the
// account's email as verified.
func (handler *Handlers) ConfirmEmailVerification(c fiber.Ctx) error {
	var req auth.ConfirmEmailVerificationDTO
	if err := c.Bind().Body(&req); err != nil {
		return handler.responseBadRequest(c, "invalid request body", "auth:emailVerification:confirm")
	}

	if err := handler.services.VerifyEmail(c.Context(), req.Token); err != nil {
		return handler.emailVerificationError(c, err, "auth:emailVerification:confirm")
	}

	return handler.responseStatusOk(c, "email verified", "email verified successfully", nil)
}

// emailVerificationError maps the email-verification sentinels to precise
// statuses; anything else falls through to the shared domain-error mapping.
func (handler *Handlers) emailVerificationError(c fiber.Ctx, err error, action string) error {
	switch {
	case errors.Is(err, services.ErrEmailVerificationExpired):
		return c.Status(fiber.StatusGone).JSON(fiber.Map{
			"success": false,
			"message": "email verification link expired",
			"details": "the link has expired; request a new one",
			"action":  action,
		})
	case errors.Is(err, services.ErrEmailVerificationInvalid):
		return handler.responseBadRequest(c, "invalid email verification link", "the link is invalid or has already been used")
	default:
		return handler.responseFromDomain(c, err, "failed to process email verification", action)
	}
}
