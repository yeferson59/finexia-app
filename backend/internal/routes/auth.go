package routes

func (r *Routes) Auth() {
	auth := r.app.Group("/auth")

	auth.Post("/register", r.middlewares.AuthLimiter(), r.handlers.Register)
	auth.Post("/login", r.middlewares.AuthLimiter(), r.handlers.Login)
	auth.Post("/refresh", r.middlewares.AuthLimiter(), r.handlers.Refresh)

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

	// Public email verification flow: (re)send a link, validate its token,
	// then confirm to mark the email verified. Rate-limited to blunt both
	// mail-bombing an address and token guessing.
	auth.Post("/verify-email", r.middlewares.AuthLimiter(), r.handlers.RequestEmailVerification)
	auth.Get("/verify-email", r.middlewares.AuthLimiter(), r.handlers.ValidateEmailVerification)
	auth.Post("/verify-email/confirm", r.middlewares.AuthLimiter(), r.handlers.ConfirmEmailVerification)

	auth.Use(r.middlewares.JWT())
	auth.Get("/session", r.handlers.GetSession)
	auth.Get("/sessions", r.handlers.ListSessions)
	auth.Delete("/sessions/:id", r.handlers.RevokeSession)
	auth.Post("/sessions/revoke-others", r.handlers.RevokeOtherSessions)
	auth.Post("/logout", r.handlers.Logout)
}
