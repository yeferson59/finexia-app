package user

import (
	"bytes"
	"io"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/paginate"

	"github.com/yeferson59/finexia-app/internal/identity"
	"github.com/yeferson59/finexia-app/internal/platform/httpx"
	"github.com/yeferson59/finexia-app/pkg/dtos"
)

type handler struct {
	service  *Service
	sessions sessionRevoker
}

func newHandler(svc *Service, sessions sessionRevoker) *handler {
	return new(handler{svc, sessions})
}

func (h *handler) GetListUsers(c fiber.Ctx) error {
	paginateInfo, ok := paginate.FromContext(c)
	if !ok {
		return httpx.InternalServerError(c, "", "paginate info not found")
	}

	users, count, err := h.service.GetListUsers(c, uint(paginateInfo.Offset), uint(paginateInfo.Limit))
	if err != nil {
		return httpx.FromDomain(c, err, "get product pagination", "users:list")
	}

	return httpx.OK(c, "product pagination", "get products successfully", dtos.FilterPagination[[]identity.User]{
		Items:    users,
		MetaData: httpx.PaginationMetadata(paginateInfo, count),
	})
}

func (h *handler) GetUserByID(c fiber.Ctx) error {
	userID, err := httpx.ParamUUID(c, "id")
	if err != nil {
		return httpx.BadRequest(c, "validate id", "invalid user id")
	}

	user, err := h.service.GetUserByID(c, userID)
	if err != nil {
		return httpx.FromDomain(c, err, "get user by id", "users:id")
	}

	return httpx.OK(c, "get user by id", "get user successfully", user)
}

func (h *handler) CreateUser(c fiber.Ctx) error {
	createUserDto, err := httpx.Bind[CreateDTO](c)
	if err != nil {
		return httpx.BadRequest(c, "", err.Error())
	}

	user, err := h.service.CreateUser(c, createUserDto.Name, createUserDto.Email)
	if err != nil {
		return httpx.FromDomain(c, err, "", "users:create")
	}

	return httpx.Success(c, fiber.StatusCreated, "", "", user)
}

func (h *handler) UpdateUser(c fiber.Ctx) error {
	userID, err := httpx.ParamUUID(c, "id")
	if err != nil {
		return httpx.BadRequest(c, "", err.Error())
	}

	updateUser, err := httpx.Bind[UpdateDTO](c)
	if err != nil {
		return httpx.BadRequest(c, "", err.Error())
	}

	user, err := h.service.UpdateUser(c, userID, updateUser.Name, updateUser.Email, updateUser.Image)
	if err != nil {
		return httpx.FromDomain(c, err, "", "users:update")
	}

	return httpx.OK(c, "", "", user)
}

func (h *handler) DeleteUser(c fiber.Ctx) error {
	userID, err := httpx.ParamUUID(c, "id")
	if err != nil {
		return httpx.BadRequest(c, "", err.Error())
	}

	if err := h.service.DeleteUser(c, userID); err != nil {
		return httpx.FromDomain(c, err, "", "users:delete")
	}

	// The delete is a soft one: the row stays, so every session and refresh
	// token pointing at it stays valid too unless they are closed here.
	if _, err := h.sessions.RevokeAllSessions(c, userID); err != nil {
		return httpx.FromDomain(c, err, "", "users:delete:sessions")
	}

	return httpx.Success(c, fiber.StatusNoContent, "", "", "")
}

func (h *handler) BanUser(c fiber.Ctx) error {
	userID, err := httpx.ParamUUID(c, "id")
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	req, err := httpx.Bind[BanUserDTO](c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid request", err.Error())
	}

	if err := h.service.BanUser(c, userID, req.Ban); err != nil {
		return httpx.FromDomain(c, err, "Error updating ban status", "users:ban")
	}

	// Banning is meant to cut access off now, not at the next token expiry, so
	// the user's live sessions and refresh-token families go with the flag.
	// Unbanning leaves them closed: the user logs in again.
	if req.Ban {
		if _, err := h.sessions.RevokeAllSessions(c, userID); err != nil {
			return httpx.FromDomain(c, err, "Error updating ban status", "users:ban:sessions")
		}
	}

	msg := "User banned"
	if !req.Ban {
		msg = "User unbanned"
	}

	return httpx.OK(c, msg, msg, nil)
}

func (h *handler) GetMe(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	u, err := h.service.GetCurrentUser(c, userID)
	if err != nil {
		return httpx.FromDomain(c, err, "Error retrieving user", "users:me:get")
	}

	return httpx.OK(c, "User retrieved", "User retrieved successfully", u)
}

func (h *handler) UpdateMe(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	req, err := httpx.Bind[UpdateProfileDTO](c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid request", err.Error())
	}

	u, err := h.service.UpdateCurrentUser(c, userID, req.Name, req.Image, req.PreferredCurrency)
	if err != nil {
		return httpx.FromDomain(c, err, "Error updating user", "users:me:update")
	}

	return httpx.OK(c, "User updated", "User updated successfully", u)
}

func (h *handler) UploadAvatar(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	fileHeader, err := c.FormFile("avatar")
	if err != nil {
		return httpx.BadRequest(c, "Missing file", "avatar file is required")
	}

	const maxSize = 5 << 20 // 5 MB
	if fileHeader.Size > maxSize {
		return httpx.BadRequest(c, "File too large", "avatar must be smaller than 5 MB")
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
			return httpx.BadRequest(c, "Invalid file type", "only JPEG, PNG and WebP are allowed")
		}
	}

	f, err := fileHeader.Open()
	if err != nil {
		return httpx.InternalServerError(c, "File open error", err.Error())
	}
	defer func() { _ = f.Close() }()

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(f); err != nil {
		return httpx.InternalServerError(c, "File read error", err.Error())
	}

	u, err := h.service.UploadAvatarToS3(c, userID, &buf, contentType)
	if err != nil {
		return httpx.InternalServerError(c, "Upload failed", err.Error())
	}

	return httpx.OK(c, "Avatar uploaded", "Avatar uploaded successfully", u)
}

func (h *handler) GetMyPreferences(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	prefs, err := h.service.GetUserPreferences(c, userID)
	if err != nil {
		return httpx.FromDomain(c, err, "Error retrieving preferences", "users:me:preferences:get")
	}

	return httpx.OK(c, "Preferences retrieved", "Preferences retrieved successfully", prefs)
}

func (h *handler) UpdateMyPreferences(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	req, err := httpx.Bind[UpdatePreferencesDTO](c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid request", err.Error())
	}

	prefs, err := h.service.UpdateUserPreferences(c, userID, req.EmailAlerts, req.WeeklySummary)
	if err != nil {
		return httpx.FromDomain(c, err, "Error updating preferences", "users:me:preferences:update")
	}

	return httpx.OK(c, "Preferences updated", "Preferences updated successfully", prefs)
}

func (h *handler) GetUserAvatar(c fiber.Ctx) error {
	userID, err := httpx.ParamUUID(c, "id")
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	body, contentType, err := h.service.GetAvatarFromS3(c, userID)
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
