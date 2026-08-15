// Package httpx holds the HTTP response conventions and generic middlewares
// shared by every module. The envelope shapes are a frozen contract: clients
// depend on them byte for byte, so they are documented in docs/API.md §1.1–§1.2
// and must not drift.
package httpx

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/paginate"

	"github.com/yeferson59/finexia-app/pkg/dtos"
	"github.com/yeferson59/finexia-app/pkg/helpers"
)

// Success writes the success envelope with an explicit status.
func Success(c fiber.Ctx, status int, message, details string, data any) error {
	return c.Status(status).JSON(dtos.Response{
		Success:   true,
		Message:   message,
		Details:   details,
		Data:      data,
		Timestamp: time.Now(),
	})
}

// OK writes the standard 200 success envelope.
func OK(c fiber.Ctx, message, details string, data any) error {
	return Success(c, fiber.StatusOK, message, details, data)
}

// Error writes the error envelope with an explicit status.
func Error(c fiber.Ctx, status int, message, details string) error {
	return c.Status(status).JSON(dtos.Response{
		Success:   false,
		Message:   message,
		Details:   details,
		Timestamp: time.Now(),
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

func Forbidden(c fiber.Ctx, message, details string) error {
	return Error(c, fiber.StatusForbidden, message, details)
}

// FromDomain maps a service error to an HTTP status and writes an error
// envelope carrying an action code (docs/API.md §1.2). The status comes from
// the error's typed Kind — domains tag their errors via httpx.AsNotFound and
// friends — and an untagged error is a 500.
func FromDomain(c fiber.Ctx, err error, message, action string) error {
	return c.Status(domainStatus(err)).JSON(dtos.Response{
		Success:   false,
		Message:   message,
		Action:    action,
		Timestamp: time.Now(),
	})
}

// ErrorAction writes an error envelope that carries an "action" code alongside
// free-form details, for statuses FromDomain doesn't cover (403, 409, 410, …).
func ErrorAction(c fiber.Ctx, status int, message, details, action string) error {
	return c.Status(status).JSON(dtos.Response{
		Success:   false,
		Message:   message,
		Details:   details,
		Action:    action,
		Timestamp: time.Now(),
	})
}

// SuccessAction is the success counterpart of ErrorAction, for responses that
// carry both an action code and a data payload.
func SuccessAction(c fiber.Ctx, status int, message, details, action string, data any) error {
	return c.Status(status).JSON(dtos.Response{
		Success:   true,
		Message:   message,
		Details:   details,
		Action:    action,
		Data:      data,
		Timestamp: time.Now(),
	})
}

// PaginationMetadata builds the standard "MetaData" block shared by every
// paginated list response.
func PaginationMetadata(paginateInfo *paginate.PageInfo, count uint) dtos.MetaDataPagination {
	totalPages := helpers.CalculateTotalPages(count, uint(paginateInfo.Limit))

	return dtos.MetaDataPagination{
		CurrentPage: uint(paginateInfo.Page),
		Offset:      uint(paginateInfo.Offset),
		Limit:       uint(paginateInfo.Limit),
		Total:       count,
		TotalPages:  totalPages,
		Previous:    paginateInfo.Page > 1,
		Next:        paginateInfo.Page < int(totalPages),
	}
}
