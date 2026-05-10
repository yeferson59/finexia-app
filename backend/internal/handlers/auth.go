package handlers

import (
	jwtware "github.com/gofiber/contrib/v3/jwt"
	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"

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
	userID := claims["id"].(string)
	name := claims["name"].(string)

	return handler.responseStatusOk(c, "", "", fiber.Map{
		"id":   userID,
		"name": name,
	})
}
