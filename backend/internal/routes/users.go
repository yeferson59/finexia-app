package routes

import "github.com/gofiber/fiber/v3/middleware/paginate"

func (r *Routes) Users() {
	users := r.router.Group("/users")

	users.Get("", paginate.New(), r.handlers.GetListUsers)
	users.Post("", r.handlers.CreateUser)

	// Self-service routes — must be registered before /:id to avoid shadowing.
	users.Get("/me", r.handlers.GetMe)
	users.Patch("/me", r.handlers.UpdateMe)
	users.Get("/me/preferences", r.handlers.GetMyPreferences)
	users.Patch("/me/preferences", r.handlers.UpdateMyPreferences)
	users.Patch("/me/password", r.handlers.ChangeMyPassword)

	users.Get("/:id", r.handlers.GetUserByID)
	users.Patch("/:id", r.handlers.UpdateUser)
	users.Delete("/:id", r.handlers.DeleteUser)
}
