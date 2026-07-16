package routes

import "github.com/gofiber/fiber/v3/middleware/paginate"

func (r *Routes) Users() {
	users := r.router.Group("/users")

	users.Get("", r.auth.RequireAdmin(), paginate.New(), r.handlers.GetListUsers)
	users.Post("", r.auth.RequireAdmin(), r.handlers.CreateUser)

	// The admin invitation/waitlist routes live in the auth module
	// (Module.AdminRoutes); they register before this group so the static
	// "/invitations" and "/waitlist" segments are never captured by "/:id".

	// Self-service routes — must be registered before /:id to avoid shadowing.
	users.Get("/me", r.handlers.GetMe)
	users.Patch("/me", r.handlers.UpdateMe)
	users.Post("/me/avatar", r.handlers.UploadAvatar)
	users.Get("/me/preferences", r.handlers.GetMyPreferences)
	users.Patch("/me/preferences", r.handlers.UpdateMyPreferences)
	users.Patch("/me/password", r.handlers.ChangeMyPassword)

	users.Get("/:id", r.auth.RequireAdmin(), r.handlers.GetUserByID)
	users.Patch("/:id", r.auth.RequireAdmin(), r.handlers.UpdateUser)
	users.Patch("/:id/ban", r.auth.RequireAdmin(), r.handlers.BanUser)
	users.Delete("/:id", r.auth.RequireAdmin(), r.handlers.DeleteUser)
}
