package routes

func (r *Routes) Auth() {
	auth := r.app.Group("/auth")

	auth.Post("/register", r.handlers.Register)
	auth.Post("/login", r.handlers.Login)
	auth.Post("/refresh", r.handlers.Refresh)
	auth.Use(r.middlewares.Session(), r.middlewares.JWT())
	auth.Get("/session", r.handlers.GetSession)
	auth.Post("/logout", r.handlers.Logout)
}
