package user

import (
	"bytes"
	"errors"
	"io"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/paginate"
	"github.com/google/uuid"

	"github.com/yeferson59/finexia-app/internal/auth"
	"github.com/yeferson59/finexia-app/internal/identity"
	"github.com/yeferson59/finexia-app/internal/marketing"
	"github.com/yeferson59/finexia-app/internal/platform/httpx"
	"github.com/yeferson59/finexia-app/pkg/dtos"
)

type handler struct {
	service *Service
}

const (
	LocalUserID = "auth_user_id"
	LocalToken  = "auth_token"
	LocalRole   = "auth_role"
)

// getUserIDTokenRole extracts the authenticated identity the JWT middleware
// stored in the request locals.
func getUserIDTokenRole(c fiber.Ctx) (uuid.UUID, string, string, error) {
	userIDStr, _ := c.Locals(LocalUserID).(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, "", "", err
	}

	token, _ := c.Locals(LocalToken).(string)
	role, _ := c.Locals(LocalRole).(string)
	if token == "" || role == "" {
		return uuid.Nil, "", "", errors.New("missing authenticated identity")
	}

	return userID, token, role, nil
}

func getParamUUID(c fiber.Ctx, paramName string) (uuid.UUID, error) {
	return uuid.Parse(c.Params(paramName))
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

	return httpx.OK(c, "product pagination", "get products successfully", dtos.FilterPagination[[]identity.User, fiber.Map]{
		Items:    users,
		MetaData: httpx.PaginationMetadata(paginateInfo, count, "usersForPage", "totalUsers"),
	})
}

func (h *handler) GetUserByID(c fiber.Ctx) error {
	userID, err := getParamUUID(c, "id")
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
	var createUserDto CreateDTO

	if err := c.Bind().Body(&createUserDto); err != nil {
		return httpx.BadRequest(c, "", err.Error())
	}

	user, err := h.service.CreateUser(c, createUserDto.Name, createUserDto.Email)
	if err != nil {
		return httpx.FromDomain(c, err, "", "users:create")
	}

	return httpx.Success(c, fiber.StatusCreated, "", "", user)
}

func (h *handler) UpdateUser(c fiber.Ctx) error {
	userID, err := getParamUUID(c, "id")
	if err != nil {
		return httpx.BadRequest(c, "", err.Error())
	}

	var updateUser UpdateDTO

	if err := c.Bind().Body(&updateUser); err != nil {
		return httpx.BadRequest(c, "", err.Error())
	}

	user, err := h.service.UpdateUser(c, userID, updateUser.Name, updateUser.Email, updateUser.Image)
	if err != nil {
		return httpx.FromDomain(c, err, "", "users:update")
	}

	return httpx.OK(c, "", "", user)
}

func (h *handler) DeleteUser(c fiber.Ctx) error {
	userID, err := getParamUUID(c, "id")
	if err != nil {
		return httpx.BadRequest(c, "", err.Error())
	}

	if err := h.service.DeleteUser(c, userID); err != nil {
		return httpx.FromDomain(c, err, "", "users:delete")
	}

	return httpx.Success(c, fiber.StatusNoContent, "", "", "")
}

func (h *handler) BanUser(c fiber.Ctx) error {
	userID, err := getParamUUID(c, "id")
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	var req BanUserDTO
	if err := c.Bind().JSON(&req); err != nil {
		return httpx.BadRequest(c, "Invalid request", err.Error())
	}

	if err := h.service.BanUser(c, userID, req.Ban); err != nil {
		return httpx.FromDomain(c, err, "Error updating ban status", "users:ban")
	}

	msg := "User banned"
	if !req.Ban {
		msg = "User unbanned"
	}

	return httpx.OK(c, msg, msg, nil)
}

func (h *handler) GetMe(c fiber.Ctx) error {
	userID, _, _, err := getUserIDTokenRole(c)
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
	userID, _, _, err := getUserIDTokenRole(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	var req UpdateProfileDTO
	if err := c.Bind().JSON(&req); err != nil {
		return httpx.BadRequest(c, "Invalid request", err.Error())
	}

	u, err := h.service.UpdateCurrentUser(c, userID, req.Name, req.PreferredCurrency, req.Image)
	if err != nil {
		return httpx.FromDomain(c, err, "Error updating user", "users:me:update")
	}

	return httpx.OK(c, "User updated", "User updated successfully", u)
}

func (h *handler) UploadAvatar(c fiber.Ctx) error {
	userID, _, _, err := getUserIDTokenRole(c)
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
	userID, _, _, err := getUserIDTokenRole(c)
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
	userID, _, _, err := getUserIDTokenRole(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	var req UpdatePreferencesDTO
	if err := c.Bind().JSON(&req); err != nil {
		return httpx.BadRequest(c, "Invalid request", err.Error())
	}

	prefs, err := h.service.UpdateUserPreferences(c, userID, req.EmailAlerts, req.WeeklySummary)
	if err != nil {
		return httpx.FromDomain(c, err, "Error updating preferences", "users:me:preferences:update")
	}

	return httpx.OK(c, "Preferences updated", "Preferences updated successfully", prefs)
}

func (h *handler) GetUserAvatar(c fiber.Ctx) error {
	userID, err := getParamUUID(c, "id")
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

func (h *handler) ChangeMyPassword(c fiber.Ctx) error {
	userID, jwtoken, _, err := getUserIDTokenRole(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid user ID", err.Error())
	}

	var req ChangePasswordDTO
	if err := c.Bind().JSON(&req); err != nil {
		return httpx.BadRequest(c, "Invalid request", err.Error())
	}

	if len(req.NewPassword) < 8 {
		return httpx.BadRequest(c, "Invalid password", "New password must be at least 8 characters")
	}

	// Same upper bound as register/login validation; without it the user could
	// set a password that the login endpoint would later reject.
	if len(req.NewPassword) > 20 {
		return httpx.BadRequest(c, "Invalid password", "New password must be at most 20 characters")
	}

	if err := h.service.ChangePassword(c, userID, jwtoken, req.CurrentPassword, req.NewPassword, c.IP(), c.Get("User-Agent")); err != nil {
		return httpx.FromDomain(c, err, "Error changing password", "users:me:password")
	}

	return httpx.OK(c, "Password changed", "Password changed successfully", nil)
}

// createInvitation (admin) issues an invitation and emails the recipient a
// single-use link to set their password.
func (h *handler) createInvitation(c fiber.Ctx) error {
	var req InviteUserDTO
	if err := c.Bind().Body(&req); err != nil {
		return httpx.BadRequest(c, "invalid request body", err.Error())
	}

	invitedBy, _, _, err := getUserIDTokenRole(c)
	if err != nil {
		return httpx.BadRequest(c, "invalid user id", err.Error())
	}

	inv, err := h.service.CreateInvitation(c, req.Email, req.Name, req.Role, invitedBy)
	if err != nil {
		return httpx.FromDomain(c, err, "failed to create invitation", "invitations:create")
	}

	return httpx.Success(c, fiber.StatusCreated, "invitation sent", "invitation created successfully", inv)
}

// listInvitations (admin) returns the still-open invitations for the dashboard.
func (h *handler) listInvitations(c fiber.Ctx) error {
	paginateInfo, ok := paginate.FromContext(c)
	if !ok {
		return httpx.InternalServerError(c, "", "paginate info not found")
	}

	invitations, count, err := h.service.ListInvitations(c, uint(paginateInfo.Offset), uint(paginateInfo.Limit))
	if err != nil {
		return httpx.FromDomain(c, err, "failed to list invitations", "invitations:list")
	}

	return httpx.OK(c, "invitations", "invitations retrieved successfully", dtos.FilterPagination[[]auth.Invitation, fiber.Map]{
		Items:    invitations,
		MetaData: httpx.PaginationMetadata(paginateInfo, count, "limit", "total"),
	})
}

// resendInvitation (admin) rotates the token of a pending invitation and resends.
func (h *handler) resendInvitation(c fiber.Ctx) error {
	id, err := getParamUUID(c, "id")
	if err != nil {
		return httpx.BadRequest(c, "invalid invitation id", err.Error())
	}

	invitedBy, _, _, err := getUserIDTokenRole(c)
	if err != nil {
		return httpx.BadRequest(c, "invalid user id", err.Error())
	}

	inv, err := h.service.ResendInvitation(c, id, invitedBy)
	if err != nil {
		return httpx.FromDomain(c, err, "failed to resend invitation", "invitations:resend")
	}

	return httpx.OK(c, "invitation resent", "invitation resent successfully", inv)
}

// revokeInvitation (admin) invalidates a pending invitation.
func (h *handler) revokeInvitation(c fiber.Ctx) error {
	id, err := getParamUUID(c, "id")
	if err != nil {
		return httpx.BadRequest(c, "invalid invitation id", err.Error())
	}

	if err := h.service.RevokeInvitation(c, id); err != nil {
		return httpx.FromDomain(c, err, "failed to revoke invitation", "invitations:revoke")
	}

	return httpx.OK(c, "invitation revoked", "invitation revoked successfully", nil)
}

// listWaitlist (admin) returns the waitlist so admins can invite from it.
func (h *handler) listWaitlist(c fiber.Ctx) error {
	paginateInfo, ok := paginate.FromContext(c)
	if !ok {
		return httpx.InternalServerError(c, "", "paginate info not found")
	}

	waitlist, count, err := h.service.ListWaitlist(c, uint(paginateInfo.Offset), uint(paginateInfo.Limit))
	if err != nil {
		return httpx.FromDomain(c, err, "failed to list waitlist", "waitlist:list")
	}

	return httpx.OK(c, "waitlist", "waitlist retrieved successfully", dtos.FilterPagination[[]marketing.Waitlist, fiber.Map]{
		Items:    waitlist,
		MetaData: httpx.PaginationMetadata(paginateInfo, count, "limit", "total"),
	})
}
