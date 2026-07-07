package handlers

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/paginate"
	"github.com/google/uuid"

	"github.com/yeferson59/finexia-app/internal/middlewares"
	"github.com/yeferson59/finexia-app/pkg/helpers"
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

func (handler *Handlers) responseStatusOk(c fiber.Ctx, message, details string, data any) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success":   true,
		"message":   message,
		"details":   details,
		"data":      data,
		"timestamp": time.Now(),
	})
}

func (handler *Handlers) responseSuccess(c fiber.Ctx, status int, message, details string, data any) error {
	return c.Status(status).JSON(fiber.Map{
		"success":   true,
		"message":   message,
		"details":   details,
		"data":      data,
		"timestamp": time.Now(),
	})
}

func (handler *Handlers) responseBadRequest(c fiber.Ctx, message, details string) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"success":   false,
		"message":   message,
		"details":   details,
		"timestamp": time.Now(),
	})
}

func (handler *Handlers) responseInternalServerError(c fiber.Ctx, message, details string) error {
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"success":   false,
		"message":   message,
		"details":   details,
		"timestamp": time.Now(),
	})
}

func (handler *Handlers) responseFromDomain(c fiber.Ctx, err error, message, action string) error {
	if strings.Contains(err.Error(), "too many") {
		return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
			"success":   false,
			"message":   message,
			"action":    action,
			"timestamp": time.Now(),
		})
	}

	if strings.Contains(err.Error(), "failed") || strings.Contains(err.Error(), "invalid") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success":   false,
			"message":   message,
			"action":    action,
			"timestamp": time.Now(),
		})
	}

	if strings.Contains(err.Error(), "not found") {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success":   false,
			"message":   message,
			"action":    action,
			"timestamp": time.Now(),
		})
	}

	if strings.Contains(err.Error(), "already exist") || strings.Contains(err.Error(), "already found") || strings.Contains(err.Error(), "duplicate") {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"success":   false,
			"message":   message,
			"action":    action,
			"timestamp": time.Now(),
		})
	}

	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"success":   false,
		"message":   message,
		"action":    action,
		"timestamp": time.Now(),
	})
}

// responseErrorAction is for error cases that need an "action" code instead
// of (or alongside) free-form details, mirroring responseFromDomain's shape
// for statuses that domain-error mapping doesn't cover (403, 409, 410, ...).
func (handler *Handlers) responseErrorAction(c fiber.Ctx, status int, message, details, action string) error {
	return c.Status(status).JSON(fiber.Map{
		"success":   false,
		"message":   message,
		"details":   details,
		"action":    action,
		"timestamp": time.Now(),
	})
}

// responseSuccessAction is the success counterpart of responseErrorAction,
// for responses that carry both an action code and a data payload.
func (handler *Handlers) responseSuccessAction(c fiber.Ctx, status int, message, details, action string, data any) error {
	return c.Status(status).JSON(fiber.Map{
		"success":   true,
		"message":   message,
		"details":   details,
		"action":    action,
		"data":      data,
		"timestamp": time.Now(),
	})
}

func (handler *Handlers) responseUnauthorized(c fiber.Ctx, message, details string) error {
	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
		"success":   false,
		"message":   message,
		"details":   details,
		"timestamp": time.Now(),
	})
}

// paginationMetadata builds the standard "MetaData" block shared by every
// paginated list response. limitKey and totalKey let callers keep their
// existing field names (e.g. "usersForPage"/"totalUsers") without repeating
// the rest of the pagination-math boilerplate.
func paginationMetadata(paginateInfo *paginate.PageInfo, count uint, limitKey, totalKey string) fiber.Map {
	totalPages := helpers.CalculateTotalPages(count, uint(paginateInfo.Limit))

	return fiber.Map{
		"currentPage": paginateInfo.Page,
		limitKey:      paginateInfo.Limit,
		"offset":      paginateInfo.Offset,
		totalKey:      count,
		"totalPages":  totalPages,
		"previous":    paginateInfo.Page > 1,
		"next":        paginateInfo.Page < int(totalPages),
	}
}

func (handler *Handlers) getUserIDTokenRole(c fiber.Ctx) (uuid.UUID, string, string, error) {
	userIDStr, _ := c.Locals(middlewares.LocalUserID).(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, "", "", err
	}

	token, _ := c.Locals(middlewares.LocalToken).(string)
	role, _ := c.Locals(middlewares.LocalRole).(string)
	if token == "" || role == "" {
		return uuid.Nil, "", "", errors.New("missing authenticated identity")
	}

	return userID, token, role, nil
}
