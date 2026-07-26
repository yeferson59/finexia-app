package portfolio

// Config is the portfolio module's own configuration surface: exactly the
// settings this domain reads, decoupled from the platform-wide *config.Env.
// The composition root (internal/app) populates it from the environment, so
// the module — and its tests — depend on a small, explicit struct instead of
// the full Env.
type Config struct {
	// FrontendURL is the base URL used to build the dashboard link in the
	// transaction-import summary email.
	FrontendURL string
}
