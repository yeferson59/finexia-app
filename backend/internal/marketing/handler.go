package marketing

import (
	"github.com/gofiber/fiber/v3"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

type handler struct {
	service *Service
}

// createWaitlist keeps the exact contract of the legacy
// POST /marketing/waitlists handler (docs/API.md §2.2).
func (h *handler) createWaitlist(c fiber.Ctx) error {
	var req waitlistRequest

	if err := c.Bind().Body(&req); err != nil {
		return httpx.BadRequest(c, "invalid email", "email is required and must be a valid email address")
	}

	if err := h.service.SaveWaitlistEmail(c, req.Email); err != nil {
		return httpx.FromDomain(c, err, "error saving waitlist email", err.Error())
	}

	return httpx.OK(c, "waitlist created successfully", "", fiber.Map{
		"email": req.Email,
	})
}
