package handlers

import (
	"bytes"
	"io"
	"path/filepath"
	"strings"

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

func (handler *Handlers) BanUser(c fiber.Ctx) error {
	userID, err := handler.getParamUUID(c, "id")
	if err != nil {
		return handler.responseBadRequest(c, "Invalid user ID", err.Error())
	}

	var req user.BanUserDTO
	if err := c.Bind().JSON(&req); err != nil {
		return handler.responseBadRequest(c, "Invalid request", err.Error())
	}

	if err := handler.services.BanUser(handler.ctx, userID, req.Ban); err != nil {
		return handler.responseFromDomain(c, err, "Error updating ban status", "users:ban")
	}

	msg := "User banned"
	if !req.Ban {
		msg = "User unbanned"
	}
	return handler.responseStatusOk(c, msg, msg, nil)
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

func (handler *Handlers) UploadAvatar(c fiber.Ctx) error {
	userID, _, _, err := handler.getUserIDTokenRole(c)
	if err != nil {
		return handler.responseBadRequest(c, "Invalid user ID", err.Error())
	}

	fileHeader, err := c.FormFile("avatar")
	if err != nil {
		return handler.responseBadRequest(c, "Missing file", "avatar file is required")
	}

	const maxSize = 5 << 20 // 5 MB
	if fileHeader.Size > maxSize {
		return handler.responseBadRequest(c, "File too large", "avatar must be smaller than 5 MB")
	}

	contentType := fileHeader.Header.Get("Content-Type")
	allowed := map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/webp": ".webp",
	}
	_, ok := allowed[strings.ToLower(contentType)]
	if !ok {
		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
		switch ext {
		case ".jpg", ".jpeg":
			contentType = "image/jpeg"
		case ".png":
			contentType = "image/png"
		case ".webp":
			contentType = "image/webp"
		default:
			return handler.responseBadRequest(c, "Invalid file type", "only JPEG, PNG and WebP are allowed")
		}
	}

	f, err := fileHeader.Open()
	if err != nil {
		return handler.responseInternalServerError(c, "File open error", err.Error())
	}
	defer func() { _ = f.Close() }()

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(f); err != nil {
		return handler.responseInternalServerError(c, "File read error", err.Error())
	}

	u, err := handler.services.UploadAvatarToS3(handler.ctx, userID, &buf, contentType)
	if err != nil {
		return handler.responseInternalServerError(c, "Upload failed", err.Error())
	}

	return handler.responseStatusOk(c, "Avatar uploaded", "Avatar uploaded successfully", u)
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

func (handler *Handlers) GetUserAvatar(c fiber.Ctx) error {
	userID, err := handler.getParamUUID(c, "id")
	if err != nil {
		return handler.responseBadRequest(c, "Invalid user ID", err.Error())
	}

	body, contentType, err := handler.services.GetAvatarFromS3(handler.ctx, userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString("avatar not found")
	}
	defer func() { _ = body.Close() }()

	c.Set("Content-Type", contentType)
	c.Set("Cache-Control", "public, max-age=86400")
	c.Set("Cross-Origin-Resource-Policy", "cross-origin")
	_, err = io.Copy(c.Response().BodyWriter(), body)
	return err
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
