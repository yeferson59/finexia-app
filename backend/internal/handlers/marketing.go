package handlers

import (
	"github.com/gofiber/fiber/v3"

	"github.com/yeferson59/finexia-app/internal/dtos/marketing"
)

func (h *Handlers) CreateWaitlistMarketing(c fiber.Ctx) error {
	var waitlist marketing.Waitlist

	if err := c.Bind().Body(&waitlist); err != nil {
		return h.responseBadRequest(c, "invalid email", "email is required and must be a valid email address")
	}

	if err := h.services.SaveWaitlistEmail(c.Context(), waitlist.Email); err != nil {
		return h.responseFromDomain(c, err, "error saving waitlist email", err.Error())
	}

	return h.responseStatusOk(c, "waitlist created successfully", "", fiber.Map{
		"email": waitlist.Email,
	})
}
