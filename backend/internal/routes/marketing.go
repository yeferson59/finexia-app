package routes

func (r *Routes) Marketing() {
	waitlists := r.app.Group("/marketing")

	waitlists.Post("/waitlists", r.handlers.CreateWaitlistMarketing)
}
