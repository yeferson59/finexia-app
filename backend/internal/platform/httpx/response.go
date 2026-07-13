// Package httpx holds the HTTP response conventions and generic middlewares
// shared by every module. The envelope shapes and the domain-error mapping
// replicate the legacy handlers/helpers.go behavior byte for byte — they are
// the frozen contract documented in docs/API.md §1.1–§1.2.
package httpx

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/paginate"

	"github.com/yeferson59/finexia-app/pkg/helpers"
)

// OK writes the standard 200 success envelope.
func OK(c fiber.Ctx, message, details string, data any) error {
	return Success(c, fiber.StatusOK, message, details, data)
}

// Success writes the success envelope with an explicit status.
func Success(c fiber.Ctx, status int, message, details string, data any) error {
	return c.Status(status).JSON(fiber.Map{
		"success":   true,
		"message":   message,
		"details":   details,
		"data":      data,
		"timestamp": time.Now(),
	})
}

// Error writes the error envelope with an explicit status.
func Error(c fiber.Ctx, status int, message, details string) error {
	return c.Status(status).JSON(fiber.Map{
		"success":   false,
		"message":   message,
		"details":   details,
		"timestamp": time.Now(),
	})
}

// BadRequest writes the 400 error envelope.
func BadRequest(c fiber.Ctx, message, details string) error {
	return Error(c, fiber.StatusBadRequest, message, details)
}

// Unauthorized writes the 401 error envelope.
func Unauthorized(c fiber.Ctx, message, details string) error {
	return Error(c, fiber.StatusUnauthorized, message, details)
}

// InternalServerError writes the 500 error envelope.
func InternalServerError(c fiber.Ctx, message, details string) error {
	return Error(c, fiber.StatusInternalServerError, message, details)
}

// FromDomain maps a service error to an HTTP status by matching substrings of
// its message, and writes an error envelope carrying an action code. The
// matching order is part of the frozen contract: "failed"/"invalid" win over
// "not found" (see docs/TECH_DEBT.md #1 for the planned typed-error redesign).
func FromDomain(c fiber.Ctx, err error, message, action string) error {
	status := fiber.StatusInternalServerError

	switch msg := err.Error(); {
	case strings.Contains(msg, "too many"):
		status = fiber.StatusTooManyRequests
	case strings.Contains(msg, "failed") || strings.Contains(msg, "invalid"):
		status = fiber.StatusBadRequest
	case strings.Contains(msg, "not found"):
		status = fiber.StatusNotFound
	case strings.Contains(msg, "already exist") || strings.Contains(msg, "already found") || strings.Contains(msg, "duplicate"):
		status = fiber.StatusConflict
	}

	return c.Status(status).JSON(fiber.Map{
		"success":   false,
		"message":   message,
		"action":    action,
		"timestamp": time.Now(),
	})
}

// ErrorAction writes an error envelope that carries an "action" code alongside
// free-form details, for statuses FromDomain doesn't cover (403, 409, 410, …).
func ErrorAction(c fiber.Ctx, status int, message, details, action string) error {
	return c.Status(status).JSON(fiber.Map{
		"success":   false,
		"message":   message,
		"details":   details,
		"action":    action,
		"timestamp": time.Now(),
	})
}

// SuccessAction is the success counterpart of ErrorAction, for responses that
// carry both an action code and a data payload.
func SuccessAction(c fiber.Ctx, status int, message, details, action string, data any) error {
	return c.Status(status).JSON(fiber.Map{
		"success":   true,
		"message":   message,
		"details":   details,
		"action":    action,
		"data":      data,
		"timestamp": time.Now(),
	})
}

// PaginationMetadata builds the standard "MetaData" block shared by every
// paginated list response. limitKey and totalKey let callers keep their
// historical field names (e.g. "usersForPage"/"totalUsers").
func PaginationMetadata(paginateInfo *paginate.PageInfo, count uint, limitKey, totalKey string) fiber.Map {
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
