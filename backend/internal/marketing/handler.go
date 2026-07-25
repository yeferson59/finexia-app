package marketing

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/paginate"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
	"github.com/yeferson59/finexia-app/pkg/dtos"
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

// listWaitlist (admin) returns the waitlist so admins can invite from it. It
// backs the invitation dashboard, which is why it answers under /users
// (docs/API.md §2.6) even though the data is this module's.
func (h *handler) listWaitlist(c fiber.Ctx) error {
	paginateInfo, ok := paginate.FromContext(c)
	if !ok {
		return httpx.InternalServerError(c, "", "paginate info not found")
	}

	waitlist, count, err := h.service.ListWaitlist(c, uint(paginateInfo.Offset), uint(paginateInfo.Limit))
	if err != nil {
		return httpx.FromDomain(c, err, "failed to list waitlist", "waitlist:list")
	}

	return httpx.OK(c, "waitlist", "waitlist retrieved successfully", dtos.FilterPagination[[]Waitlist, fiber.Map]{
		Items:    waitlist,
		MetaData: httpx.PaginationMetadata(paginateInfo, count, "limit", "total"),
	})
}
