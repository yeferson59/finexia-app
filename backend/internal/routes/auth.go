package routes

func (r *Routes) Auth() {
	auth := r.app.Group("/auth")

	auth.Post("/register", r.middlewares.AuthLimiter(), r.handlers.Register)
	auth.Post("/login", r.middlewares.AuthLimiter(), r.handlers.Login)
	auth.Post("/refresh", r.middlewares.AuthLimiter(), r.handlers.Refresh)
	auth.Use(r.middlewares.JWT())
	auth.Get("/session", r.handlers.GetSession)
	auth.Get("/sessions", r.handlers.ListSessions)
	auth.Delete("/sessions/:id", r.handlers.RevokeSession)
	auth.Post("/sessions/revoke-others", r.handlers.RevokeOtherSessions)
	auth.Post("/logout", r.handlers.Logout)
}
