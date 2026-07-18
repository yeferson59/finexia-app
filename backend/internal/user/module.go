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
	DB    *pgxpool.Pool
	Cfg   *config.Env
	Store objectstore.Store
	Mail  mailer
	Geo   geoService
	Log   logger.Logger
	Auth  authService
}

type Module struct {
	cfg     *config.Env
	service *Service
	handler *handler
}

func New(deps Deps) *Module {
	pg := NewPostgresRepository(deps.DB)
	service := NewService(pg, deps.Mail, deps.Auth, deps.Store, deps.Geo, deps.Log, deps.Cfg)

	return newModule(deps, service)
}

func newModule(deps Deps, service *Service) *Module {
	return new(Module{
		cfg:     deps.Cfg,
		service: service,
		handler: new(handler{service}),
	})
}

// Service exposes the module's use cases to the composition root and other
// modules (always consumed through interfaces declared by the consumer).
func (m *Module) Service() *Service {
	return m.service
}

func (m *Module) Routes(router fiber.Router) {
	users := router.Group("/users")

	users.Get("/users/:id/avatar", m.handler.GetUserAvatar)
	users.Get("", m.RequireAdmin(), paginate.New(), m.handler.GetListUsers)
	users.Post("", m.RequireAdmin(), m.handler.CreateUser)

	// The admin invitation/waitlist routes live in the auth module
	// (Module.AdminRoutes); they register before this group so the static
	// "/invitations" and "/waitlist" segments are never captured by "/:id".

	// Self-service routes — must be registered before /:id to avoid shadowing.
	users.Get("/me", m.handler.GetMe)
	users.Patch("/me", m.handler.UpdateMe)
	users.Post("/me/avatar", m.handler.UploadAvatar)
	users.Get("/me/preferences", m.handler.GetMyPreferences)
	users.Patch("/me/preferences", m.handler.UpdateMyPreferences)
	users.Patch("/me/password", m.handler.ChangeMyPassword)

	users.Get("/:id", m.RequireAdmin(), m.handler.GetUserByID)
	users.Patch("/:id", m.RequireAdmin(), m.handler.UpdateUser)
	users.Patch("/:id/ban", m.RequireAdmin(), m.handler.BanUser)
	users.Delete("/:id", m.RequireAdmin(), m.handler.DeleteUser)
}
