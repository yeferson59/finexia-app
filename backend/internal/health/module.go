package health

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/healthcheck"
)

type Module struct{}

func New() *Module {
	return new(Module{})
}

func (m *Module) Routes(router fiber.Router) {
	health := router.Group("/health")

	health.Get(healthcheck.LivenessEndpoint, healthcheck.New())
	health.Get(healthcheck.ReadinessEndpoint, healthcheck.New())
	health.Get(healthcheck.StartupEndpoint, healthcheck.New())
}
