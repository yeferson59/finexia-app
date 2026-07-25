package user

// Config is the user module's own configuration surface: exactly the settings
// this domain reads, decoupled from the platform-wide *config.Env. The
// composition root (internal/app) populates it from the environment, so the
// module — and its tests — depend on a small, explicit struct instead of the
// full Env (see docs/TECH_DEBT.md #8).
type Config struct {
	// PublicURL is the API's own base URL, used to build the absolute avatar
	// link returned after an upload.
	PublicURL string
	// FrontendURL is the base URL used to build links in emails.
	FrontendURL string
}
