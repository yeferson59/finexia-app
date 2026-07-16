package auth

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/paginate"

	"github.com/yeferson59/finexia-app/internal/marketing"
	"github.com/yeferson59/finexia-app/internal/platform/httpx"
	"github.com/yeferson59/finexia-app/pkg/dtos"
)

// createInvitation (admin) issues an invitation and emails the recipient a
// single-use link to set their password.
func (h *handler) createInvitation(c fiber.Ctx) error {
	var req InviteUserDTO
	if err := c.Bind().Body(&req); err != nil {
		return httpx.BadRequest(c, "invalid request body", err.Error())
	}

	invitedBy, _, _, err := getUserIDTokenRole(c)
	if err != nil {
		return httpx.BadRequest(c, "invalid user id", err.Error())
	}

	inv, err := h.service.CreateInvitation(c, req.Email, req.Name, req.Role, invitedBy)
	if err != nil {
		return httpx.FromDomain(c, err, "failed to create invitation", "invitations:create")
	}

	return httpx.Success(c, fiber.StatusCreated, "invitation sent", "invitation created successfully", inv)
}

// listInvitations (admin) returns the still-open invitations for the dashboard.
func (h *handler) listInvitations(c fiber.Ctx) error {
	paginateInfo, ok := paginate.FromContext(c)
	if !ok {
		return httpx.InternalServerError(c, "", "paginate info not found")
	}

	invitations, count, err := h.service.ListInvitations(c, uint(paginateInfo.Offset), uint(paginateInfo.Limit))
	if err != nil {
		return httpx.FromDomain(c, err, "failed to list invitations", "invitations:list")
	}

	return httpx.OK(c, "invitations", "invitations retrieved successfully", dtos.FilterPagination[[]Invitation, fiber.Map]{
		Items:    invitations,
		MetaData: httpx.PaginationMetadata(paginateInfo, count, "limit", "total"),
	})
}

// resendInvitation (admin) rotates the token of a pending invitation and resends.
func (h *handler) resendInvitation(c fiber.Ctx) error {
	id, err := getParamUUID(c, "id")
	if err != nil {
		return httpx.BadRequest(c, "invalid invitation id", err.Error())
	}

	invitedBy, _, _, err := getUserIDTokenRole(c)
	if err != nil {
		return httpx.BadRequest(c, "invalid user id", err.Error())
	}

	inv, err := h.service.ResendInvitation(c, id, invitedBy)
	if err != nil {
		return httpx.FromDomain(c, err, "failed to resend invitation", "invitations:resend")
	}

	return httpx.OK(c, "invitation resent", "invitation resent successfully", inv)
}

// revokeInvitation (admin) invalidates a pending invitation.
func (h *handler) revokeInvitation(c fiber.Ctx) error {
	id, err := getParamUUID(c, "id")
	if err != nil {
		return httpx.BadRequest(c, "invalid invitation id", err.Error())
	}

	if err := h.service.RevokeInvitation(c, id); err != nil {
		return httpx.FromDomain(c, err, "failed to revoke invitation", "invitations:revoke")
	}

	return httpx.OK(c, "invitation revoked", "invitation revoked successfully", nil)
}

// listWaitlist (admin) returns the waitlist so admins can invite from it.
func (h *handler) listWaitlist(c fiber.Ctx) error {
	paginateInfo, ok := paginate.FromContext(c)
	if !ok {
		return httpx.InternalServerError(c, "", "paginate info not found")
	}

	waitlist, count, err := h.service.ListWaitlist(c, uint(paginateInfo.Offset), uint(paginateInfo.Limit))
	if err != nil {
		return httpx.FromDomain(c, err, "failed to list waitlist", "waitlist:list")
	}

	return httpx.OK(c, "waitlist", "waitlist retrieved successfully", dtos.FilterPagination[[]marketing.Waitlist, fiber.Map]{
		Items:    waitlist,
		MetaData: httpx.PaginationMetadata(paginateInfo, count, "limit", "total"),
	})
}

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
