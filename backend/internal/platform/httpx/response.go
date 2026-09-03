// Package httpx holds the HTTP response conventions and generic middlewares
// shared by every module. The envelope shapes are a frozen contract: clients
// depend on them byte for byte, so they are documented in docs/API.md §1.1–§1.2
// and must not drift.
package httpx

import (
	"os"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/paginate"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/rs/zerolog"

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
//
// On a 4xx the error's own text is returned as Details, on a 5xx it is not.
// The split is the whole point: a 4xx says the caller sent something this
// server will never accept, and naming what — "USD does not convert into itself
// at 1.0638" — is the difference between fixing one field and retrying blind
// against the same refusal. A 5xx is the server's problem, its text is written
// for the operator reading logs, and it can carry a query, a constraint name or
// a connection string, so it stays in.
//
// Handlers keep passing a generic message: it is the headline a client shows
// when it has nowhere to put the detail, and the detail is additive.
func FromDomain(c fiber.Ctx, err error, message, action string) error {
	status := domainStatus(err)

	details := ""
	if status >= fiber.StatusBadRequest && status < fiber.StatusInternalServerError && err != nil {
		details = err.Error()
	}

	// A 5xx's cause goes to the operator, since it is deliberately kept out of
	// the response. Without this it went nowhere at all: FromDomain consumed the
	// error and returned only the envelope, so the request logger recorded a
	// bare "500" and the text — a failing constraint, a missing column after a
	// migration that was never run — existed in no log on either side.
	if status >= fiber.StatusInternalServerError && err != nil {
		serverErrors.Error().
			Err(err).
			Str("method", c.Method()).
			Str("path", c.Path()).
			Str("requestId", requestid.FromContext(c)).
			Msg(message)
	}

	return c.Status(status).JSON(dtos.Response{
		Success:   false,
		Message:   message,
		Details:   details,
		Action:    action,
		Timestamp: time.Now(),
	})
}

// serverErrors writes the causes FromDomain refuses to put in a response body.
// It builds its own stderr writer for the same reason the Logger middleware
// does: this package is called from every handler in the app and threading a
// logger through all of them, to be used on one branch, would cost more than it
// buys.
var serverErrors = zerolog.New(os.Stderr).With().Timestamp().Str("component", "httpx").Logger()

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
