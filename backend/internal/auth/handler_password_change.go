package auth

import (
	"github.com/gofiber/fiber/v3"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

// ChangePasswordDTO is the self-service password change payload.
type ChangePasswordDTO struct {
	// NewPassword keeps the same bounds as RegisterRequestDTO/LoginRequestDTO
	// (min=8,max=20); otherwise a user could set a password login would reject.
	CurrentPassword string `json:"currentPassword" validate:"required,min=8"`
	NewPassword     string `json:"newPassword"     validate:"required,min=8,max=20"`
}

// changePassword answers PATCH /users/me/password. The path belongs to the
// user dashboard, but the credentials are this module's (docs/API.md §2.3).
func (h *handler) changePassword(c fiber.Ctx) error {
	userID, jwtoken, _, err := getUserIDTokenRole(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	var req ChangePasswordDTO
	if err := c.Bind().JSON(&req); err != nil {
		return httpx.BadRequest(c, "Invalid request", err.Error())
	}

	if len(req.NewPassword) < 8 {
		return httpx.BadRequest(c, "Invalid password", "New password must be at least 8 characters")
	}

	// Same upper bound as register/login validation; without it the user could
	// set a password that the login endpoint would later reject.
	if len(req.NewPassword) > 20 {
		return httpx.BadRequest(c, "Invalid password", "New password must be at most 20 characters")
	}

	if err := h.service.ChangePassword(c, userID, jwtoken, req.CurrentPassword, req.NewPassword, c.IP(), c.Get("User-Agent")); err != nil {
		return httpx.FromDomain(c, err, "Error changing password", "users:me:password")
	}

	return httpx.OK(c, "Password changed", "Password changed successfully", nil)
}
