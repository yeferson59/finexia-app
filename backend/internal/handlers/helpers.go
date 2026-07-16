package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/paginate"
	"github.com/google/uuid"

	"github.com/yeferson59/finexia-app/internal/auth"
	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

func (handler *Handlers) getParamUUID(c fiber.Ctx, paramName string) (uuid.UUID, error) {
	id := c.Params(paramName)
	idUUID, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, err
	}

	return idUUID, nil
}

func (handler *Handlers) GetParamID(c fiber.Ctx, paramName string) (string, error) {
	id := c.Params(paramName)

	return id, nil
}

// The response helpers below delegate to platform/httpx since Fase 1 of the
// architecture migration; the envelope shapes and the domain-error mapping
// live there now (frozen contract, docs/API.md §1.1–§1.2). These wrappers
// only keep the legacy handler surface stable until each handler moves to
// its module.

func (handler *Handlers) responseStatusOk(c fiber.Ctx, message, details string, data any) error {
	return httpx.OK(c, message, details, data)
}

func (handler *Handlers) responseSuccess(c fiber.Ctx, status int, message, details string, data any) error {
	return httpx.Success(c, status, message, details, data)
}

func (handler *Handlers) responseBadRequest(c fiber.Ctx, message, details string) error {
	return httpx.BadRequest(c, message, details)
}

func (handler *Handlers) responseInternalServerError(c fiber.Ctx, message, details string) error {
	return httpx.InternalServerError(c, message, details)
}

func (handler *Handlers) responseFromDomain(c fiber.Ctx, err error, message, action string) error {
	return httpx.FromDomain(c, err, message, action)
}

func paginationMetadata(paginateInfo *paginate.PageInfo, count uint, limitKey, totalKey string) fiber.Map {
	return httpx.PaginationMetadata(paginateInfo, count, limitKey, totalKey)
}

func (handler *Handlers) getUserIDTokenRole(c fiber.Ctx) (uuid.UUID, string, string, error) {
	userIDStr, _ := c.Locals(auth.LocalUserID).(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, "", "", err
	}

	token, _ := c.Locals(auth.LocalToken).(string)
	role, _ := c.Locals(auth.LocalRole).(string)
	if token == "" || role == "" {
		return uuid.Nil, "", "", errors.New("missing authenticated identity")
	}

	return userID, token, role, nil
}
