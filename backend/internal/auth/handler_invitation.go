package auth

import (
	"errors"

	"github.com/gofiber/fiber/v3"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

// validateInvitation (public) reports whether a token is redeemable and returns
// the safe fields the accept page needs to prefill.
func (h *handler) validateInvitation(c fiber.Ctx) error {
	token := c.Query("token")

	inv, err := h.service.ValidateInvitation(c, token)
	if err != nil {
		return h.invitationError(c, err, "invitations:validate")
	}

	return httpx.OK(c, "invitation valid", "invitation is valid", fiber.Map{
		"email":     inv.Email,
		"name":      inv.Name,
		"expiresAt": inv.ExpiresAt,
	})
}

// acceptInvitation (public) provisions the account from a valid invitation and
// the password the invitee chose.
func (h *handler) acceptInvitation(c fiber.Ctx) error {
	var req AcceptInvitationDTO
	if err := c.Bind().Body(&req); err != nil {
		return httpx.BadRequest(c, "invalid request body", err.Error())
	}

	u, err := h.service.AcceptInvitation(c, req.Token, req.Name, req.Password)
	if err != nil {
		return h.invitationError(c, err, "invitations:accept")
	}

	return httpx.Success(c, fiber.StatusCreated, "account created", "invitation accepted successfully", fiber.Map{
		"id":    u.ID,
		"name":  u.Name,
		"email": u.Email,
	})
}

// invitationError maps the invitation sentinels to precise statuses; anything
// else falls through to the shared domain-error mapping.
func (h *handler) invitationError(c fiber.Ctx, err error, action string) error {
	switch {
	case errors.Is(err, ErrInvitationExpired):
		return httpx.ErrorAction(c, fiber.StatusGone, "invitation expired", "the invitation link has expired; ask an admin to resend it", action)
	case errors.Is(err, ErrInvitationInvalid):
		return httpx.BadRequest(c, "invalid invitation", "the invitation link is invalid or has already been used")
	default:
		return httpx.FromDomain(c, err, "failed to process invitation", action)
	}
}
