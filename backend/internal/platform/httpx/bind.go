package httpx

import (
	"github.com/gofiber/fiber/v3"
)

// Bind decodes the request body into a, running the app's StructValidator.
//
// It binds the body and nothing else. Fiber's Bind().All() would additionally
// merge the query string, the request headers and the cookies into any field
// the body left unset, which turns every DTO in this codebase into an
// alternative intake for values that were only ever meant to arrive in a body:
// passwords, password-reset and invitation tokens, TOTP codes, and the
// market-data API keys users bring. A value passed that way reaches places a
// body never does — access logs, Referer headers, browser history, proxy caches
// — and a field the caller left out could be filled from a header they do not
// control either.
//
// Path parameters are not bound here either; handlers read them explicitly with
// ParamUUID or c.Params, which keeps it visible at the call site which part of
// the URL a handler trusts.
func Bind[a any](c fiber.Ctx) (a, error) {
	var result a

	if err := c.Bind().Body(&result); err != nil {
		return result, err
	}

	return result, nil
}
