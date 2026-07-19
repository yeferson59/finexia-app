package user

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/paginate"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yeferson59/finexia-app/internal/platform/config"
	"github.com/yeferson59/finexia-app/internal/platform/logger"
	"github.com/yeferson59/finexia-app/internal/platform/objectstore"
)

type Deps struct {
	DB        *pgxpool.Pool
	Cfg       *config.Env
	Store     objectstore.Store
	Mail      mailer
	Geo       geoService
	Log       logger.Logger
	Auth      authService
	AuthMiddl authMiddleware
	// Limiter is the per-user rate limiter the legacy /users routes had via
	// the app-wide gate; the module keeps it now that it registers in the
	// public zone with its own RequireAuth.
	Limiter fiber.Handler
}

type authMiddleware interface {
	RequireAuth() fiber.Handler
	RequireAdmin() fiber.Handler
}

type Module struct {
	cfg       *config.Env
	service   *Service
	handler   *handler
	authMiddl authMiddleware
	limiter   fiber.Handler
}

func New(deps Deps) *Module {
	pg := NewPostgresRepository(deps.DB)
	service := NewService(pg, deps.Mail, deps.Auth, deps.Store, deps.Geo, deps.Log, deps.Cfg)

	return newModule(deps, service)
}

func newModule(deps Deps, service *Service) *Module {
	return new(Module{
		cfg:       deps.Cfg,
		service:   service,
		handler:   new(handler{service}),
		authMiddl: deps.AuthMiddl,
		limiter:   deps.Limiter,
	})
}

// Service exposes the module's use cases to the composition root and other
// modules (always consumed through interfaces declared by the consumer).
func (m *Module) Service() *Service {
	return m.service
}

func (m *Module) Routes(router fiber.Router) {
	// Public avatar (docs/API.md §2.3): registered as a terminal handler
	// before the group's RequireAuth so it stays outside the auth gate.
	router.Get("/users/:id/avatar", m.handler.GetUserAvatar)

	users := router.Group("/users")

	users.Use(m.authMiddl.RequireAuth(), m.limiter)

	// Self-service routes — must be registered before /:id to avoid shadowing.
	users.Get("/me", m.handler.GetMe)
	users.Patch("/me", m.handler.UpdateMe)
	users.Post("/me/avatar", m.handler.UploadAvatar)
	users.Get("/me/preferences", m.handler.GetMyPreferences)
	users.Patch("/me/preferences", m.handler.UpdateMyPreferences)
	users.Patch("/me/password", m.handler.ChangeMyPassword)

	// Admin guards go inline per route (never group.Use) so unmatched
	// /users/* requests fall through to a 404 instead of a 403.
	admin := m.authMiddl.RequireAdmin()
	users.Get("", admin, paginate.New(), m.handler.GetListUsers)
	users.Post("", admin, m.handler.CreateUser)
	users.Get("/:id", admin, m.handler.GetUserByID)
	users.Patch("/:id", admin, m.handler.UpdateUser)
	users.Patch("/:id/ban", admin, m.handler.BanUser)
	users.Delete("/:id", admin, m.handler.DeleteUser)
}
