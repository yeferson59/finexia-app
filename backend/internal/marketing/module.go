package marketing

import "github.com/gofiber/fiber/v3"

// Module is the marketing domain module: construction via New, HTTP surface
// via Routes. It receives only the dependencies it uses.
type Module struct {
	service *Service
	handler *handler
}

func New(repo Repository, mail Mailer) *Module {
	service := NewService(repo, mail)
	return &Module{
		service: service,
		handler: &handler{service: service},
	}
}

// Service exposes the module's use cases to the composition root and other
// modules (always consumed through interfaces declared by the consumer).
func (m *Module) Service() *Service {
	return m.service
}

// Routes registers the module's endpoints, replicating routes/marketing.go.
// The waitlist admin listing lives under /users/waitlist in the user module
// instead: it is gated by user's own RequireAuth/RequireAdmin, alongside the
// invitation dashboard.
func (m *Module) Routes(router fiber.Router) {
	waitlists := router.Group("/marketing")

	waitlists.Post("/waitlists", m.handler.createWaitlist)
}
