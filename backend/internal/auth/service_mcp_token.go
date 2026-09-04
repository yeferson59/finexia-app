package auth

// Personal access token use cases: create, list, rotate, delete, and the
// authentication the /mcp guard performs on every request.

import (
	"context"
	"time"

	"uuid"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

// CreateMCPToken mints a token for the user and returns it with its secret —
// the only time that value exists outside the client's configuration file.
func (s *service) CreateMCPToken(ctx context.Context, userID uuid.UUID, name string, expiresInDays int) (MCPTokenSecret, error) {
	name, err := normalizeMCPTokenName(name)
	if err != nil {
		return MCPTokenSecret{}, err
	}

	expiresAt, err := resolveMCPTokenExpiry(expiresInDays, time.Now().UTC())
	if err != nil {
		return MCPTokenSecret{}, err
	}

	// Checked before minting rather than enforced by the insert: the limit is
	// about how many tokens a person can keep track of, so it deserves an
	// answer that says so instead of a constraint violation.
	count, err := s.stores.MCPTokens.CountMCPTokens(ctx, userID)
	if err != nil {
		return MCPTokenSecret{}, err
	}

	if count >= maxMCPTokensPerUser {
		return MCPTokenSecret{}, httpx.AsConflict(ErrTooManyMCPTokens)
	}

	raw, hash, last4, err := generateMCPToken()
	if err != nil {
		return MCPTokenSecret{}, err
	}

	token, err := s.stores.MCPTokens.CreateMCPToken(ctx, userID, name, hash, last4, expiresAt)
	if err != nil {
		return MCPTokenSecret{}, mcpTokenDomainError(err)
	}

	s.log.Info(ctx, "mcp token created",
		logger.Str("user_id", userID.String()),
		logger.Str("token_id", token.ID.String()),
	)

	return MCPTokenSecret{MCPToken: withExpiredFlag(token), Token: raw}, nil
}

// ListMCPTokens returns the user's tokens, none of them carrying a secret.
func (s *service) ListMCPTokens(ctx context.Context, userID uuid.UUID) ([]MCPToken, error) {
	tokens, err := s.stores.MCPTokens.ListMCPTokens(ctx, userID)
	if err != nil {
		return nil, err
	}

	for i := range tokens {
		tokens[i] = withExpiredFlag(tokens[i])
	}

	return tokens, nil
}

// RotateMCPToken replaces a token's secret in place. The previous secret stops
// working the moment this returns — there is no grace window on purpose:
// rotation is what a user reaches for when a token may have leaked, and a
// window would keep the leaked one alive for exactly as long.
func (s *service) RotateMCPToken(ctx context.Context, userID, tokenID uuid.UUID, expiresInDays int) (MCPTokenSecret, error) {
	expiresAt, err := resolveMCPTokenExpiry(expiresInDays, time.Now().UTC())
	if err != nil {
		return MCPTokenSecret{}, err
	}

	raw, hash, last4, err := generateMCPToken()
	if err != nil {
		return MCPTokenSecret{}, err
	}

	token, err := s.stores.MCPTokens.RotateMCPToken(ctx, userID, tokenID, hash, last4, expiresAt)
	if err != nil {
		return MCPTokenSecret{}, mcpTokenDomainError(err)
	}

	s.log.Info(ctx, "mcp token rotated",
		logger.Str("user_id", userID.String()),
		logger.Str("token_id", token.ID.String()),
	)

	return MCPTokenSecret{MCPToken: withExpiredFlag(token), Token: raw}, nil
}

// DeleteMCPToken revokes a token permanently.
func (s *service) DeleteMCPToken(ctx context.Context, userID, tokenID uuid.UUID) error {
	if err := s.stores.MCPTokens.DeleteMCPToken(ctx, userID, tokenID); err != nil {
		return mcpTokenDomainError(err)
	}

	s.log.Info(ctx, "mcp token deleted",
		logger.Str("user_id", userID.String()),
		logger.Str("token_id", tokenID.String()),
	)

	return nil
}

// AuthenticateMCPToken resolves a presented bearer token to the identity the
// /mcp guard puts in the request locals.
//
// It is not cached, unlike access-token validation. An access token expires in
// minutes, so a stale cache entry outlives its subject by very little; these
// live for months, and the whole point of the delete button is that it takes
// effect on the next call. One indexed lookup per MCP request — a model calling
// a tool, not a page rendering — is the cheaper side of that trade.
func (s *service) AuthenticateMCPToken(ctx context.Context, raw string) (uuid.UUID, string, error) {
	if !looksLikeMCPToken(raw) {
		return uuid.Nil(), "", httpx.AsBadRequest(ErrInvalidMCPToken)
	}

	token, err := s.stores.MCPTokens.GetMCPTokenByHash(ctx, hashMCPToken(raw))
	if err != nil {
		// An unknown hash is the ordinary case — a revoked token, a typo, a
		// probe — so it is not logged as a failure. A store that broke is.
		if !isMCPTokenNotFound(err) {
			s.log.Error(ctx, "mcp token lookup failed", logger.Err(err))
		}

		return uuid.Nil(), "", httpx.AsBadRequest(ErrInvalidMCPToken)
	}

	now := time.Now().UTC()
	if token.ExpiresAt != nil && now.After(*token.ExpiresAt) {
		return uuid.Nil(), "", httpx.AsBadRequest(ErrInvalidMCPToken)
	}

	s.touchMCPToken(ctx, token, now)

	return token.UserID, token.Role, nil
}

// touchMCPToken keeps last_used_at fresh enough to answer "is this token still
// in use?" without writing on every call. A failed write is swallowed: the
// column is for the settings screen, and losing one update must never turn a
// working tool call into a 401.
func (s *service) touchMCPToken(ctx context.Context, token mcpTokenIdentity, now time.Time) {
	if token.LastUsedAt != nil && now.Sub(*token.LastUsedAt) < mcpTokenTouchInterval {
		return
	}

	if err := s.stores.MCPTokens.TouchMCPToken(ctx, token.ID); err != nil {
		s.log.Warn(ctx, "mcp token touch failed",
			logger.Str("token_id", token.ID.String()),
			logger.Err(err),
		)
	}
}

// withExpiredFlag derives the field the store does not hold.
func withExpiredFlag(t MCPToken) MCPToken {
	t.Expired = t.ExpiresAt != nil && time.Now().UTC().After(*t.ExpiresAt)

	return t
}
