package httpx

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// The authenticated caller travels from the auth middleware to every handler
// through the request locals. Both ends of that convention live here: the auth
// module writes the keys in its RequireAuth success handler, and each module's
// handlers read them back with Identity. Keeping the keys and the reader in
// one place is what stops them drifting — the same three constants and the
// same accessor used to be copy-pasted into auth, user and portfolio, so a
// change to the convention had to be made three times to stay correct.
//
// This is transport plumbing, not domain: httpx only moves the identity across
// the request, it does not decide who is authenticated (that is auth's job).

const (
	// LocalUserID holds the caller's user id, as its string form.
	LocalUserID = "auth_user_id"
	// LocalToken holds the raw access token the caller presented, needed by
	// the flows that must spare the current session (password change, logout).
	LocalToken = "auth_token"
	// LocalRole holds the caller's role name, which the admin guard reads.
	LocalRole = "auth_role"
)

// RoleAdmin is the privileged role's name. It lives beside LocalRole because
// the two are one convention: a handler that reads the role from the locals to
// decide what to do — rather than being gated by the guard — compares against
// this, and the guard itself is built from it.
const RoleAdmin = "admin"

// ErrNoIdentity reports that the locals carry no usable identity: the request
// reached the handler without passing the auth middleware, or passed it with
// an incomplete claim set. Handlers translate it to a 400.
var ErrNoIdentity = errors.New("missing authenticated identity")

// Identity returns the authenticated caller's id, access token and role, as
// stored in the request locals by the auth middleware. It errors when the user
// id is missing or malformed, or when the token or role are absent.
func Identity(c fiber.Ctx) (uuid.UUID, string, string, error) {
	userIDStr, _ := c.Locals(LocalUserID).(string)

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, "", "", err
	}

	token, _ := c.Locals(LocalToken).(string)
	role, _ := c.Locals(LocalRole).(string)

	if token == "" || role == "" {
		return uuid.Nil, "", "", ErrNoIdentity
	}

	return userID, token, role, nil
}

// ParamUUID parses a path parameter as a UUID. The error is the caller's to
// map: every handler answers a malformed id with a 400 naming the parameter.
func ParamUUID(c fiber.Ctx, paramName string) (uuid.UUID, error) {
	return uuid.Parse(c.Params(paramName))
}
