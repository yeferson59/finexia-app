package auth

import (
	"errors"

	"github.com/gofiber/fiber/v3"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

type handler struct {
	service *service
	cfg     Config
}

func newHandler(svc *service, cfg Config) *handler {
	return new(handler{svc, cfg})
}

func (h *handler) login(c fiber.Ctx) error {
	loginDto, err := httpx.Bind[LoginRequestDTO](c)
	if err != nil {
		return httpx.BadRequest(c, "invalid request body", "auth:login")
	}

	result, err := h.service.Login(c, loginDto.Email, loginDto.Password, c.IP(), c.Get("User-Agent"))
	if err != nil {
		if errors.Is(err, ErrTwoFactorRequired) {
			return httpx.SuccessAction(c, fiber.StatusOK, "two-factor authentication required", "enter the code from your authenticator app", "auth:login:2fa", fiber.Map{
				"twoFactorToken": result.TwoFactorToken,
			})
		}

		if errors.Is(err, ErrAccountUnverified) {
			return httpx.ErrorAction(c, fiber.StatusForbidden, "email not verified", "verify your email before logging in", "auth:login:unverified")
		}

		if errors.Is(err, ErrAccountDisabled) {
			return httpx.ErrorAction(c, fiber.StatusForbidden, "account disabled", "this account has been suspended; contact support", "auth:login:disabled")
		}

		return httpx.FromDomain(c, err, "failed to login", "auth:login")
	}

	c.Cookie(new(fiber.Cookie{
		Name:     "refresh_token",
		Value:    result.RawRefreshToken,
		Path:     "/",
		HTTPOnly: true,
		Secure:   h.cfg.Environment == "production",
		SameSite: "Strict",
		MaxAge:   int(h.cfg.JWTRefreshDuration.Seconds()),
	}))

	return httpx.OK(c, "successfully logged in", "valid credentials",
		LoginResponseDTO{ID: result.ID, AccessToken: result.AccessToken})
}

func (h *handler) register(c fiber.Ctx) error {
	if !h.cfg.SelfRegistrationEnabled {
		return httpx.ErrorAction(c, fiber.StatusForbidden, "self-registration is disabled",
			"Finexia is invite-only during the beta; ask an existing member for an invitation", "auth:register:disabled")
	}

	registerDto, err := httpx.Bind[RegisterRequestDTO](c)
	if err != nil {
		return httpx.BadRequest(c, "invalid request body", "auth:register")
	}

	user, err := h.service.Register(c, registerDto.Name, registerDto.Email, registerDto.Password)
	if err != nil {
		if errors.Is(err, ErrEmailAlreadyExists) {
			return httpx.ErrorAction(c, fiber.StatusConflict, "email already registered", "an account with this email already exists", "auth:register:duplicate")
		}
		return httpx.FromDomain(c, err, "failed to register", "auth:register")
	}

	return httpx.OK(c, "successfully registered", "valid registration data", user)
}

func (h *handler) refresh(c fiber.Ctx) error {
	rawToken := c.Cookies("refresh_token")
	if rawToken == "" {
		return httpx.Unauthorized(c, "missing refresh token", "auth:refresh")
	}

	result, err := h.service.RefreshToken(c, rawToken, c.IP(), c.Get("User-Agent"))
	if err != nil {
		return httpx.Unauthorized(c, "invalid refresh token", "auth:refresh")
	}

	c.Cookie(new(fiber.Cookie{
		Name:     "refresh_token",
		Value:    result.RawRefreshToken,
		Path:     "/",
		HTTPOnly: true,
		Secure:   h.cfg.Environment == "production",
		SameSite: "Strict",
		MaxAge:   int(h.cfg.JWTRefreshDuration.Seconds()),
	}))

	return httpx.OK(c, "token refreshed", "valid refresh token",
		LoginResponseDTO{AccessToken: result.AccessToken})
}

func (h *handler) getSession(c fiber.Ctx) error {
	userID, jwtoken, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "invalid user id", "auth:getSession")
	}

	userSession, err := h.service.GetSession(c, userID, jwtoken)
	if err != nil {
		return httpx.FromDomain(c, err, "failed to get session", "auth:getSession")
	}

	return httpx.OK(c, "successfully retrieved session", "valid access token", userSession)
}

func (h *handler) listSessions(c fiber.Ctx) error {
	userID, jwtoken, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "invalid user id", "auth:sessions:list")
	}

	sessions, err := h.service.ListSessions(c, userID, jwtoken)
	if err != nil {
		return httpx.FromDomain(c, err, "failed to list sessions", "auth:sessions:list")
	}

	return httpx.OK(c, "active sessions retrieved", "valid access token", sessions)
}

func (h *handler) revokeSession(c fiber.Ctx) error {
	userID, jwtoken, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "invalid user id", "auth:sessions:revoke")
	}

	sessionID, err := httpx.ParamUUID(c, "id")
	if err != nil {
		return httpx.BadRequest(c, "invalid session id", "auth:sessions:revoke")
	}

	if err := h.service.RevokeSession(c, userID, sessionID, jwtoken); err != nil {
		return httpx.FromDomain(c, err, "failed to revoke session", "auth:sessions:revoke")
	}

	return httpx.OK(c, "session revoked", "session revoked successfully", nil)
}

func (h *handler) revokeOtherSessions(c fiber.Ctx) error {
	userID, jwtoken, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "invalid user id", "auth:sessions:revokeOthers")
	}

	revoked, err := h.service.RevokeOtherSessions(c, userID, jwtoken)
	if err != nil {
		return httpx.FromDomain(c, err, "failed to revoke sessions", "auth:sessions:revokeOthers")
	}

	return httpx.OK(c, "sessions revoked", "other sessions revoked successfully", fiber.Map{
		"revoked": revoked,
	})
}

func (h *handler) logout(c fiber.Ctx) error {
	userID, jwtoken, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "invalid user id", "auth:logout")
	}

	rawRefreshToken := c.Cookies("refresh_token")

	if err := h.service.Logout(c, userID, jwtoken, rawRefreshToken); err != nil {
		return httpx.FromDomain(c, err, "failed to logout", "auth:logout")
	}

	c.Cookie(new(fiber.Cookie{
		Name:   "refresh_token",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}))

	return httpx.OK(c, "successfully logged out", "valid access token", nil)
}
