package auth

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/paginate"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
	"github.com/yeferson59/finexia-app/pkg/dtos"
)

// Admin side of the invitation flow. It answers under /users/invitations
// because it is part of the admin user dashboard (docs/API.md §2.6), but the
// domain — the Invitation type, its store and its service — is this module's,
// so the handlers live here rather than in the user module, which used to hold
// four pass-through wrappers around this same service.

// InviteUserDTO is the admin-side payload to invite someone. Name is optional
// (derived from the email when absent, and the invitee can set their real name
// on accept); Role defaults to "customer" and is whitelisted server-side.
type InviteUserDTO struct {
	Email string `json:"email" validate:"required,email,max=254"`
	Name  string `json:"name"  validate:"omitempty,max=254"`
	Role  string `json:"role"  validate:"omitempty,oneof=customer admin"`
}

// createInvitation (admin) issues an invitation and emails the recipient a
// single-use link to set their password.
func (h *handler) createInvitation(c fiber.Ctx) error {
	req, err := httpx.Bind[InviteUserDTO](c)
	if err != nil {
		return httpx.BadRequest(c, "invalid request body", err.Error())
	}

	invitedBy, _, _, err := httpx.Identity(c)
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

	return httpx.OK(c, "invitations", "invitations retrieved successfully", dtos.FilterPagination[[]Invitation]{
		Items:    invitations,
		MetaData: httpx.PaginationMetadata(paginateInfo, count),
	})
}

// resendInvitation (admin) rotates the token of a pending invitation and resends.
func (h *handler) resendInvitation(c fiber.Ctx) error {
	id, err := httpx.ParamUUID(c, "id")
	if err != nil {
		return httpx.BadRequest(c, "invalid invitation id", err.Error())
	}

	invitedBy, _, _, err := httpx.Identity(c)
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
	id, err := httpx.ParamUUID(c, "id")
	if err != nil {
		return httpx.BadRequest(c, "invalid invitation id", err.Error())
	}

	if err := h.service.RevokeInvitation(c, id); err != nil {
		return httpx.FromDomain(c, err, "failed to revoke invitation", "invitations:revoke")
	}

	return httpx.OK(c, "invitation revoked", "invitation revoked successfully", nil)
}
