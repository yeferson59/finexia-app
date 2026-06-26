package handlers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/paginate"

	"github.com/yeferson59/finexia-app/internal/dtos/user"
	"github.com/yeferson59/finexia-app/internal/entities"
	"github.com/yeferson59/finexia-app/pkg/dtos"
	"github.com/yeferson59/finexia-app/pkg/helpers"
)

func (handler *Handlers) GetListUsers(c fiber.Ctx) error {
	paginateInfo, ok := paginate.FromContext(c)
	if !ok {
		return handler.responseInternalServerError(c, "", "paginate info not found")
	}

	users, count, err := handler.services.GetListUsers(handler.ctx, uint(paginateInfo.Offset), uint(paginateInfo.Limit))
	if err != nil {
		return handler.responseFromDomain(c, err, "get product pagination", "users:list")
	}

	totalPages := helpers.CalculateTotalPages(count, uint(paginateInfo.Limit))

	return handler.responseStatusOk(c, "product pagination", "get products successfully", dtos.FilterPagination[[]entities.User, fiber.Map]{
		Items: users,
		MetaData: fiber.Map{
			"currentPage":  paginateInfo.Page,
			"usersForPage": paginateInfo.Limit,
			"offset":       paginateInfo.Offset,
			"totalUsers":   count,
			"totalPages":   totalPages,
			"previous":     paginateInfo.Page > 1,
			"next":         paginateInfo.Page < int(totalPages),
		},
	})
}

func (handler *Handlers) GetUserByID(c fiber.Ctx) error {
	userID, err := handler.getParamUUID(c, "id")
	if err != nil {
		return handler.responseBadRequest(c, "validate id", "invalid user id")
	}

	user, err := handler.services.GetUserByID(handler.ctx, userID)
	if err != nil {
		return handler.responseFromDomain(c, err, "get user by id", "users:id")
	}

	return handler.responseStatusOk(c, "get user by id", "get user successfully", user)
}

func (handler *Handlers) CreateUser(c fiber.Ctx) error {
	var createUserDto user.CreateDTO

	if err := c.Bind().Body(&createUserDto); err != nil {
		return handler.responseBadRequest(c, "", err.Error())
	}

	user, err := handler.services.CreateUser(handler.ctx, createUserDto.Name, createUserDto.Email)
	if err != nil {
		return handler.responseFromDomain(c, err, "", "users:create")
	}

	return handler.responseSuccess(c, fiber.StatusCreated, "", "", user)
}

func (handler *Handlers) UpdateUser(c fiber.Ctx) error {
	userID, err := handler.getParamUUID(c, "id")
	if err != nil {
		return handler.responseBadRequest(c, "", err.Error())
	}

	var updateUser user.UpdateDTO

	if err := c.Bind().Body(&updateUser); err != nil {
		return handler.responseBadRequest(c, "", err.Error())
	}

	user, err := handler.services.UpdateUser(handler.ctx, userID, updateUser.Name, updateUser.Email, updateUser.Image)
	if err != nil {
		return handler.responseFromDomain(c, err, "", "users:update")
	}

	return handler.responseStatusOk(c, "", "", user)
}

func (handler *Handlers) DeleteUser(c fiber.Ctx) error {
	userID, err := handler.getParamUUID(c, "id")
	if err != nil {
		return handler.responseBadRequest(c, "", err.Error())
	}

	if err := handler.services.DeleteUser(handler.ctx, userID); err != nil {
		return handler.responseFromDomain(c, err, "", "users:delete")
	}

	return handler.responseSuccess(c, fiber.StatusNoContent, "", "", "")
}

func (handler *Handlers) GetMe(c fiber.Ctx) error {
	userID, _, _, err := handler.getUserIDTokenRole(c)
	if err != nil {
		return handler.responseBadRequest(c, "Invalid user ID", err.Error())
	}

	u, err := handler.services.GetCurrentUser(handler.ctx, userID)
	if err != nil {
		return handler.responseFromDomain(c, err, "Error retrieving user", "users:me:get")
	}

	return handler.responseStatusOk(c, "User retrieved", "User retrieved successfully", u)
}

func (handler *Handlers) UpdateMe(c fiber.Ctx) error {
	userID, _, _, err := handler.getUserIDTokenRole(c)
	if err != nil {
		return handler.responseBadRequest(c, "Invalid user ID", err.Error())
	}

	var req user.UpdateProfileDTO
	if err := c.Bind().JSON(&req); err != nil {
		return handler.responseBadRequest(c, "Invalid request", err.Error())
	}

	u, err := handler.services.UpdateCurrentUser(handler.ctx, userID, req.Name, req.PreferredCurrency, req.Image)
	if err != nil {
		return handler.responseFromDomain(c, err, "Error updating user", "users:me:update")
	}

	return handler.responseStatusOk(c, "User updated", "User updated successfully", u)
}

func (handler *Handlers) GetMyPreferences(c fiber.Ctx) error {
	userID, _, _, err := handler.getUserIDTokenRole(c)
	if err != nil {
		return handler.responseBadRequest(c, "Invalid user ID", err.Error())
	}

	prefs, err := handler.services.GetUserPreferences(handler.ctx, userID)
	if err != nil {
		return handler.responseFromDomain(c, err, "Error retrieving preferences", "users:me:preferences:get")
	}

	return handler.responseStatusOk(c, "Preferences retrieved", "Preferences retrieved successfully", prefs)
}

func (handler *Handlers) UpdateMyPreferences(c fiber.Ctx) error {
	userID, _, _, err := handler.getUserIDTokenRole(c)
	if err != nil {
		return handler.responseBadRequest(c, "Invalid user ID", err.Error())
	}

	var req user.UpdatePreferencesDTO
	if err := c.Bind().JSON(&req); err != nil {
		return handler.responseBadRequest(c, "Invalid request", err.Error())
	}

	prefs, err := handler.services.UpdateUserPreferences(handler.ctx, userID, req.EmailAlerts, req.WeeklySummary)
	if err != nil {
		return handler.responseFromDomain(c, err, "Error updating preferences", "users:me:preferences:update")
	}

	return handler.responseStatusOk(c, "Preferences updated", "Preferences updated successfully", prefs)
}

func (handler *Handlers) ChangeMyPassword(c fiber.Ctx) error {
	userID, _, _, err := handler.getUserIDTokenRole(c)
	if err != nil {
		return handler.responseBadRequest(c, "Invalid user ID", err.Error())
	}

	var req user.ChangePasswordDTO
	if err := c.Bind().JSON(&req); err != nil {
		return handler.responseBadRequest(c, "Invalid request", err.Error())
	}

	if len(req.NewPassword) < 8 {
		return handler.responseBadRequest(c, "Invalid password", "New password must be at least 8 characters")
	}

	if err := handler.services.ChangePassword(handler.ctx, userID, req.CurrentPassword, req.NewPassword); err != nil {
		return handler.responseFromDomain(c, err, "Error changing password", "users:me:password")
	}

	return handler.responseStatusOk(c, "Password changed", "Password changed successfully", nil)
}
