package routes

// Auth registers the /auth routes still served by the legacy layers. The
// core (login, register, refresh, 2FA, sessions, email verification) moved
// to the auth module; password reset migrates in PR B of Fase 4 and the
// public invitation flow in PR C, deleting this file.
func (r *Routes) Auth() {
	auth := r.app.Group("/auth")

	// Public invitation flow: validate a token, then accept it by setting a
	// password. Rate-limited to blunt token guessing.
	auth.Get("/invitations", r.middlewares.AuthLimiter(), r.handlers.ValidateInvitation)
	auth.Post("/invitations/accept", r.middlewares.AuthLimiter(), r.handlers.AcceptInvitation)

	// Public password recovery flow: request a reset link, validate its
	// token, then confirm with a new password. Rate-limited to blunt both
	// mail-bombing an address and token guessing.
	auth.Post("/password-reset", r.middlewares.AuthLimiter(), r.handlers.RequestPasswordReset)
	auth.Get("/password-reset", r.middlewares.AuthLimiter(), r.handlers.ValidatePasswordReset)
	auth.Post("/password-reset/confirm", r.middlewares.AuthLimiter(), r.handlers.ConfirmPasswordReset)
}
