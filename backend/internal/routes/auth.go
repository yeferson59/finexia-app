package routes

// Auth registers the /auth routes still served by the legacy layers. The
// core (login, register, refresh, 2FA, sessions, email verification) and the
// password-reset flow moved to the auth module; the public invitation flow
// migrates in PR C of Fase 4, deleting this file.
func (r *Routes) Auth() {
	auth := r.app.Group("/auth")

	// Public invitation flow: validate a token, then accept it by setting a
	// password. Rate-limited to blunt token guessing.
	auth.Get("/invitations", r.middlewares.AuthLimiter(), r.handlers.ValidateInvitation)
	auth.Post("/invitations/accept", r.middlewares.AuthLimiter(), r.handlers.AcceptInvitation)
}
