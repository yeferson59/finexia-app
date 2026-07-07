package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/paginate"

	"github.com/yeferson59/finexia-app/internal/dtos/user"
	"github.com/yeferson59/finexia-app/internal/entities"
	"github.com/yeferson59/finexia-app/internal/services"
	"github.com/yeferson59/finexia-app/pkg/dtos"
)

// CreateInvitation (admin) issues an invitation and emails the recipient a
// single-use link to set their password.
func (handler *Handlers) CreateInvitation(c fiber.Ctx) error {
	var req user.InviteUserDTO
	if err := c.Bind().Body(&req); err != nil {
		return handler.responseBadRequest(c, "invalid request body", err.Error())
	}

	invitedBy, _, _, err := handler.getUserIDTokenRole(c)
	if err != nil {
		return handler.responseBadRequest(c, "invalid user id", err.Error())
	}

	inv, err := handler.services.CreateInvitation(c, req.Email, req.Name, req.Role, invitedBy)
	if err != nil {
		return handler.responseFromDomain(c, err, "failed to create invitation", "invitations:create")
	}

	return handler.responseSuccess(c, fiber.StatusCreated, "invitation sent", "invitation created successfully", inv)
}

// ListInvitations (admin) returns the still-open invitations for the dashboard.
func (handler *Handlers) ListInvitations(c fiber.Ctx) error {
	paginateInfo, ok := paginate.FromContext(c)
	if !ok {
		return handler.responseInternalServerError(c, "", "paginate info not found")
	}

	invitations, count, err := handler.services.ListInvitations(c, uint(paginateInfo.Offset), uint(paginateInfo.Limit))
	if err != nil {
		return handler.responseFromDomain(c, err, "failed to list invitations", "invitations:list")
	}

	return handler.responseStatusOk(c, "invitations", "invitations retrieved successfully", dtos.FilterPagination[[]entities.Invitation, fiber.Map]{
		Items:    invitations,
		MetaData: paginationMetadata(paginateInfo, count, "limit", "total"),
	})
}

// ResendInvitation (admin) rotates the token of a pending invitation and resends.
func (handler *Handlers) ResendInvitation(c fiber.Ctx) error {
	id, err := handler.getParamUUID(c, "id")
	if err != nil {
		return handler.responseBadRequest(c, "invalid invitation id", err.Error())
	}

	invitedBy, _, _, err := handler.getUserIDTokenRole(c)
	if err != nil {
		return handler.responseBadRequest(c, "invalid user id", err.Error())
	}

	inv, err := handler.services.ResendInvitation(c, id, invitedBy)
	if err != nil {
		return handler.responseFromDomain(c, err, "failed to resend invitation", "invitations:resend")
	}

	return handler.responseStatusOk(c, "invitation resent", "invitation resent successfully", inv)
}

// RevokeInvitation (admin) invalidates a pending invitation.
func (handler *Handlers) RevokeInvitation(c fiber.Ctx) error {
	id, err := handler.getParamUUID(c, "id")
	if err != nil {
		return handler.responseBadRequest(c, "invalid invitation id", err.Error())
	}

	if err := handler.services.RevokeInvitation(c, id); err != nil {
		return handler.responseFromDomain(c, err, "failed to revoke invitation", "invitations:revoke")
	}

	return handler.responseStatusOk(c, "invitation revoked", "invitation revoked successfully", nil)
}

// ListWaitlist (admin) returns the waitlist so admins can invite from it.
func (handler *Handlers) ListWaitlist(c fiber.Ctx) error {
	paginateInfo, ok := paginate.FromContext(c)
	if !ok {
		return handler.responseInternalServerError(c, "", "paginate info not found")
	}

	waitlist, count, err := handler.services.ListWaitlist(c, uint(paginateInfo.Offset), uint(paginateInfo.Limit))
	if err != nil {
		return handler.responseFromDomain(c, err, "failed to list waitlist", "waitlist:list")
	}

	return handler.responseStatusOk(c, "waitlist", "waitlist retrieved successfully", dtos.FilterPagination[[]entities.Waitlist, fiber.Map]{
		Items:    waitlist,
		MetaData: paginationMetadata(paginateInfo, count, "limit", "total"),
	})
}

// ValidateInvitation (public) reports whether a token is redeemable and returns
// the safe fields the accept page needs to prefill.
func (handler *Handlers) ValidateInvitation(c fiber.Ctx) error {
	token := c.Query("token")

	inv, err := handler.services.ValidateInvitation(c, token)
	if err != nil {
		return handler.invitationError(c, err, "invitations:validate")
	}

	return handler.responseStatusOk(c, "invitation valid", "invitation is valid", fiber.Map{
		"email":     inv.Email,
		"name":      inv.Name,
		"expiresAt": inv.ExpiresAt,
	})
}

// AcceptInvitation (public) provisions the account from a valid invitation and
// the password the invitee chose.
func (handler *Handlers) AcceptInvitation(c fiber.Ctx) error {
	var req user.AcceptInvitationDTO
	if err := c.Bind().Body(&req); err != nil {
		return handler.responseBadRequest(c, "invalid request body", err.Error())
	}

	u, err := handler.services.AcceptInvitation(c, req.Token, req.Name, req.Password)
	if err != nil {
		return handler.invitationError(c, err, "invitations:accept")
	}

	return handler.responseSuccess(c, fiber.StatusCreated, "account created", "invitation accepted successfully", fiber.Map{
		"id":    u.ID,
		"name":  u.Name,
		"email": u.Email,
	})
}

// invitationError maps the invitation sentinels to precise statuses; anything
// else falls through to the shared domain-error mapping.
func (handler *Handlers) invitationError(c fiber.Ctx, err error, action string) error {
	switch {
	case errors.Is(err, services.ErrInvitationExpired):
		return handler.responseErrorAction(c, fiber.StatusGone, "invitation expired", "the invitation link has expired; ask an admin to resend it", action)
	case errors.Is(err, services.ErrInvitationInvalid):
		return handler.responseBadRequest(c, "invalid invitation", "the invitation link is invalid or has already been used")
	default:
		return handler.responseFromDomain(c, err, "failed to process invitation", action)
	}
}
