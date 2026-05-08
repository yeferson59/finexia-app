package routes

import (
	"github.com/gofiber/fiber/v3/middleware/healthcheck"
)

func (r *Routes) Health() {
	health := r.app.Group("/health")

	health.Get(healthcheck.LivenessEndpoint, healthcheck.New())
	health.Get(healthcheck.ReadinessEndpoint, healthcheck.New())
	health.Get(healthcheck.StartupEndpoint, healthcheck.New())
}
