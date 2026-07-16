package routes

import "github.com/gofiber/fiber/v3/middleware/paginate"

func (r *Routes) Users() {
	users := r.router.Group("/users")

	users.Get("", r.auth.RequireAdmin(), paginate.New(), r.handlers.GetListUsers)
	users.Post("", r.auth.RequireAdmin(), r.handlers.CreateUser)

	// Invitations (admin). Static "/invitations" segment is registered before the
	// "/:id" routes so it is never captured as a user id.
	users.Get("/invitations", r.auth.RequireAdmin(), paginate.New(), r.handlers.ListInvitations)
	users.Post("/invitations", r.auth.RequireAdmin(), r.handlers.CreateInvitation)
	users.Post("/invitations/:id/resend", r.auth.RequireAdmin(), r.handlers.ResendInvitation)
	users.Delete("/invitations/:id", r.auth.RequireAdmin(), r.handlers.RevokeInvitation)

	users.Get("/waitlist", r.auth.RequireAdmin(), paginate.New(), r.handlers.ListWaitlist)

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
