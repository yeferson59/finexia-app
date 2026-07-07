package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v3"

	"github.com/yeferson59/finexia-app/internal/dtos/auth"
	"github.com/yeferson59/finexia-app/internal/services"
)

func (handler *Handlers) Login(c fiber.Ctx) error {
	var loginDto auth.LoginRequestDTO

	if err := c.Bind().Body(&loginDto); err != nil {
		return handler.responseBadRequest(c, "invalid request body", "auth:login")
	}

	result, err := handler.services.Login(c, loginDto.Email, loginDto.Password, c.IP(), c.Get("User-Agent"))
	if err != nil {
		if errors.Is(err, services.ErrTwoFactorRequired) {
			return handler.responseSuccessAction(c, fiber.StatusOK, "two-factor authentication required", "enter the code from your authenticator app", "auth:login:2fa", fiber.Map{
				"twoFactorToken": result.TwoFactorToken,
			})
		}

		if errors.Is(err, services.ErrAccountUnverified) {
			return handler.responseErrorAction(c, fiber.StatusForbidden, "email not verified", "verify your email before logging in", "auth:login:unverified")
		}

		return handler.responseFromDomain(c, err, "failed to login", "auth:login")
	}

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    result.RawRefreshToken,
		Path:     "/",
		HTTPOnly: true,
		Secure:   handler.cfg.Environment == "production",
		SameSite: "Strict",
		MaxAge:   int(handler.cfg.JWTRefreshDuration.Seconds()),
	})

	return handler.responseStatusOk(c, "successfully logged in", "valid credentials",
		auth.LoginResponseDTO{ID: result.ID, AccessToken: result.AccessToken})
}

func (handler *Handlers) Register(c fiber.Ctx) error {
	if !handler.cfg.SelfRegistrationEnabled {
		return handler.responseErrorAction(c, fiber.StatusForbidden, "self-registration is disabled",
			"Finexia is invite-only during the beta; ask an existing member for an invitation", "auth:register:disabled")
	}

	var registerDto auth.RegisterRequestDTO

	if err := c.Bind().Body(&registerDto); err != nil {
		return handler.responseBadRequest(c, "invalid request body", "auth:register")
	}

	user, err := handler.services.Register(c, registerDto.Name, registerDto.Email, registerDto.Password)
	if err != nil {
		if errors.Is(err, services.ErrEmailAlreadyExists) {
			return handler.responseErrorAction(c, fiber.StatusConflict, "email already registered", "an account with this email already exists", "auth:register:duplicate")
		}
		return handler.responseFromDomain(c, err, "failed to register", "auth:register")
	}

	return handler.responseStatusOk(c, "successfully registered", "valid registration data", user)
}

func (handler *Handlers) Refresh(c fiber.Ctx) error {
	rawToken := c.Cookies("refresh_token")
	if rawToken == "" {
		return handler.responseUnauthorized(c, "missing refresh token", "auth:refresh")
	}

	result, err := handler.services.RefreshToken(c, rawToken, c.IP(), c.Get("User-Agent"))
	if err != nil {
		return handler.responseUnauthorized(c, "invalid refresh token", "auth:refresh")
	}

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    result.RawRefreshToken,
		Path:     "/",
		HTTPOnly: true,
		Secure:   handler.cfg.Environment == "production",
		SameSite: "Strict",
		MaxAge:   int(handler.cfg.JWTRefreshDuration.Seconds()),
	})

	return handler.responseStatusOk(c, "token refreshed", "valid refresh token",
		auth.LoginResponseDTO{AccessToken: result.AccessToken})
}

func (handler *Handlers) GetSession(c fiber.Ctx) error {
	userID, jwtoken, _, err := handler.getUserIDTokenRole(c)
	if err != nil {
		return handler.responseBadRequest(c, "invalid user id", "auth:getSession")
	}

	userSession, err := handler.services.GetSession(c, userID, jwtoken)
	if err != nil {
		return handler.responseFromDomain(c, err, "failed to get session", "auth:getSession")
	}

	return handler.responseStatusOk(c, "successfully retrieved session", "valid access token", userSession)
}

func (handler *Handlers) ListSessions(c fiber.Ctx) error {
	userID, jwtoken, _, err := handler.getUserIDTokenRole(c)
	if err != nil {
		return handler.responseBadRequest(c, "invalid user id", "auth:sessions:list")
	}

	sessions, err := handler.services.ListSessions(c, userID, jwtoken)
	if err != nil {
		return handler.responseFromDomain(c, err, "failed to list sessions", "auth:sessions:list")
	}

	return handler.responseStatusOk(c, "active sessions retrieved", "valid access token", sessions)
}

func (handler *Handlers) RevokeSession(c fiber.Ctx) error {
	userID, jwtoken, _, err := handler.getUserIDTokenRole(c)
	if err != nil {
		return handler.responseBadRequest(c, "invalid user id", "auth:sessions:revoke")
	}

	sessionID, err := handler.getParamUUID(c, "id")
	if err != nil {
		return handler.responseBadRequest(c, "invalid session id", "auth:sessions:revoke")
	}

	if err := handler.services.RevokeSession(c, userID, sessionID, jwtoken); err != nil {
		return handler.responseFromDomain(c, err, "failed to revoke session", "auth:sessions:revoke")
	}

	return handler.responseStatusOk(c, "session revoked", "session revoked successfully", nil)
}

func (handler *Handlers) RevokeOtherSessions(c fiber.Ctx) error {
	userID, jwtoken, _, err := handler.getUserIDTokenRole(c)
	if err != nil {
		return handler.responseBadRequest(c, "invalid user id", "auth:sessions:revokeOthers")
	}

	revoked, err := handler.services.RevokeOtherSessions(c, userID, jwtoken)
	if err != nil {
		return handler.responseFromDomain(c, err, "failed to revoke sessions", "auth:sessions:revokeOthers")
	}

	return handler.responseStatusOk(c, "sessions revoked", "other sessions revoked successfully", fiber.Map{
		"revoked": revoked,
	})
}

func (handler *Handlers) Logout(c fiber.Ctx) error {
	userID, jwtoken, _, err := handler.getUserIDTokenRole(c)
	if err != nil {
		return handler.responseBadRequest(c, "invalid user id", "auth:logout")
	}

	rawRefreshToken := c.Cookies("refresh_token")

	if err := handler.services.Logout(c, userID, jwtoken, rawRefreshToken); err != nil {
		return handler.responseFromDomain(c, err, "failed to logout", "auth:logout")
	}

	c.Cookie(&fiber.Cookie{
		Name:   "refresh_token",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	return handler.responseStatusOk(c, "successfully logged out", "valid access token", nil)
}
