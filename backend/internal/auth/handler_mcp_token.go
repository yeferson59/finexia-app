package auth

import (
	"github.com/gofiber/fiber/v3"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

// The MCP token endpoints. Only two of them can ever answer with a secret —
// create and rotate — and both say so in their message: the client has one
// chance to copy it.

func (h *handler) listMCPTokens(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "invalid user id", "auth:mcpTokens:list")
	}

	tokens, err := h.service.ListMCPTokens(c, userID)
	if err != nil {
		return httpx.FromDomain(c, err, "failed to list MCP tokens", "auth:mcpTokens:list")
	}

	return httpx.OK(c, "MCP tokens retrieved", "", tokens)
}

func (h *handler) createMCPToken(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "invalid user id", "auth:mcpTokens:create")
	}

	req, err := httpx.Bind[CreateMCPTokenRequestDTO](c)
	if err != nil {
		return httpx.BadRequest(c, "invalid request body", "auth:mcpTokens:create")
	}

	token, err := h.service.CreateMCPToken(c, userID, req.Name, req.expiryDays())
	if err != nil {
		return httpx.FromDomain(c, err, "failed to create the MCP token", "auth:mcpTokens:create")
	}

	return httpx.Success(c, fiber.StatusCreated, "MCP token created",
		"copy it now: it will not be shown again", token)
}

// rotateMCPToken issues a new secret for an existing token. The body is
// optional — rotating without naming a lifetime keeps the default one — so an
// absent body is not a malformed request.
func (h *handler) rotateMCPToken(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "invalid user id", "auth:mcpTokens:rotate")
	}

	tokenID, err := httpx.ParamUUID(c, "id")
	if err != nil {
		return httpx.BadRequest(c, "invalid token id", "auth:mcpTokens:rotate")
	}

	req := RotateMCPTokenRequestDTO{}
	if len(c.Body()) > 0 {
		if req, err = httpx.Bind[RotateMCPTokenRequestDTO](c); err != nil {
			return httpx.BadRequest(c, "invalid request body", "auth:mcpTokens:rotate")
		}
	}

	token, err := h.service.RotateMCPToken(c, userID, tokenID, req.expiryDays())
	if err != nil {
		return httpx.FromDomain(c, err, "failed to rotate the MCP token", "auth:mcpTokens:rotate")
	}

	return httpx.OK(c, "MCP token rotated",
		"the previous token stopped working; copy the new one now", token)
}

func (h *handler) deleteMCPToken(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "invalid user id", "auth:mcpTokens:delete")
	}

	tokenID, err := httpx.ParamUUID(c, "id")
	if err != nil {
		return httpx.BadRequest(c, "invalid token id", "auth:mcpTokens:delete")
	}

	if err := h.service.DeleteMCPToken(c, userID, tokenID); err != nil {
		return httpx.FromDomain(c, err, "failed to delete the MCP token", "auth:mcpTokens:delete")
	}

	return httpx.OK(c, "MCP token deleted", "the token stopped working immediately", nil)
}
