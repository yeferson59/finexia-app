package notification

// Config is the notification module's own configuration surface: exactly the
// settings this domain reads, decoupled from the platform-wide *config.Env.
// The composition root (internal/app) populates it from the environment, so
// the module — and its tests — depend on a small, explicit struct instead of
// the full Env.
type Config struct {
	// PublicURL is the base URL used to build the dashboard link in the
	// weekly summary email.
	PublicURL string
}
