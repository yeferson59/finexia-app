package handlers

import (
	"github.com/gofiber/fiber/v3"

	"github.com/yeferson59/finexia-app/internal/dtos/auth"
)

func (handler *Handlers) Login(c fiber.Ctx) error {
	var loginDto auth.LoginRequestDTO

	if err := c.Bind().Body(&loginDto); err != nil {
		return handler.responseBadRequest(c, "invalid request body", "auth:login")
	}

	result, err := handler.services.Login(handler.ctx, loginDto.Email, loginDto.Password)
	if err != nil {
		return handler.responseFromDomain(c, err, "failed to login", "auth:login")
	}

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    result.RawRefreshToken,
		Path:     "/auth",
		HTTPOnly: true,
		Secure:   handler.cfg.Environment == "production",
		SameSite: "Strict",
		MaxAge:   int(handler.cfg.JWTRefreshDuration.Seconds()),
	})

	return handler.responseStatusOk(c, "successfully logged in", "valid credentials",
		auth.LoginResponseDTO{ID: result.ID, AccessToken: result.AccessToken})
}

func (handler *Handlers) Register(c fiber.Ctx) error {
	var registerDto auth.RegisterRequestDTO

	if err := c.Bind().Body(&registerDto); err != nil {
		return handler.responseBadRequest(c, "invalid request body", "auth:register")
	}

	user, err := handler.services.Register(handler.ctx, registerDto.Name, registerDto.Email, registerDto.Password)
	if err != nil {
		return handler.responseFromDomain(c, err, "failed to register", "auth:register")
	}

	return handler.responseStatusOk(c, "successfully registered", "valid registration data", user)
}

func (handler *Handlers) Refresh(c fiber.Ctx) error {
	rawToken := c.Cookies("refresh_token")
	if rawToken == "" {
		return handler.responseUnauthorized(c, "missing refresh token", "auth:refresh")
	}

	result, err := handler.services.RefreshToken(handler.ctx, rawToken, c.IP(), c.Get("User-Agent"))
	if err != nil {
		return handler.responseUnauthorized(c, "invalid refresh token", "auth:refresh")
	}

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    result.RawRefreshToken,
		Path:     "/auth",
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

	userSession, err := handler.services.GetSession(handler.ctx, userID, jwtoken)
	if err != nil {
		return handler.responseFromDomain(c, err, "failed to get session", "auth:getSession")
	}

	return handler.responseStatusOk(c, "successfully retrieved session", "valid access token", userSession)
}

func (handler *Handlers) Logout(c fiber.Ctx) error {
	userID, jwtoken, _, err := handler.getUserIDTokenRole(c)
	if err != nil {
		return handler.responseBadRequest(c, "invalid user id", "auth:logout")
	}

	rawRefreshToken := c.Cookies("refresh_token")

	if err := handler.services.Logout(handler.ctx, userID, jwtoken, rawRefreshToken); err != nil {
		return handler.responseFromDomain(c, err, "failed to logout", "auth:logout")
	}

	c.Cookie(&fiber.Cookie{
		Name:   "refresh_token",
		Value:  "",
		MaxAge: -1,
	})

	return handler.responseStatusOk(c, "successfully logged out", "valid access token", nil)
}
