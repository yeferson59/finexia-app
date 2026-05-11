package handlers

import (
	jwtware "github.com/gofiber/contrib/v3/jwt"
	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/yeferson59/finexia-app/internal/dtos/auth"
)

func (handler *Handlers) Login(c fiber.Ctx) error {
	var loginDto auth.LoginRequestDTO

	if err := c.Bind().Body(&loginDto); err != nil {
		return handler.responseBadRequest(c, "", "")
	}

	login, err := handler.services.Login(handler.ctx, loginDto.Email, loginDto.Password)
	if err != nil {
		return handler.responseFromDomain(c, err, "", "auth:login")
	}

	return handler.responseStatusOk(c, "", "", login)
}

func (handler *Handlers) Register(c fiber.Ctx) error {
	var registerDto auth.RegisterRequestDTO

	if err := c.Bind().Body(&registerDto); err != nil {
		return handler.responseBadRequest(c, "", "")
	}

	user, err := handler.services.Register(handler.ctx, registerDto.Name, registerDto.Email, registerDto.Password)
	if err != nil {
		return handler.responseFromDomain(c, err, "", "auth:register")
	}

	return handler.responseStatusOk(c, "", "", user)
}

func (handler *Handlers) GetSession(c fiber.Ctx) error {
	jwtToken := jwtware.FromContext(c)

	claims := jwtToken.Claims.(jwt.MapClaims)
	role := claims["role"].(string)
	userID, err := uuid.Parse(claims["id"].(string))
	if err != nil {
		return handler.responseBadRequest(c, "invalid user id", "auth:getSession")
	}

	userSession, err := handler.services.GetSession(handler.ctx, userID, role, jwtToken.Raw)
	if err != nil {
		return handler.responseFromDomain(c, err, "", "auth:getSession")
	}

	return handler.responseStatusOk(c, "", "", userSession)
}

func (handler *Handlers) Logout(c fiber.Ctx) error {
	jwtToken := jwtware.FromContext(c)

	claims := jwtToken.Claims.(jwt.MapClaims)
	userID, err := uuid.Parse(claims["id"].(string))
	if err != nil {
		return handler.responseBadRequest(c, "invalid user id", "auth:logout")
	}

	if err := handler.services.Logout(handler.ctx, userID, jwtToken.Raw); err != nil {
		return handler.responseFromDomain(c, err, "", "auth:logout")
	}

	return handler.responseStatusOk(c, "successfully logged out", "valid access token", nil)
}
