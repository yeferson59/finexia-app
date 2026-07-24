package auth

// Token and session lifecycle: access-token validation, refresh-token
// rotation, logout and expired-record cleanup. Split out of service.go
// to keep each file under ~500 lines.

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

func (s *Service) ValidateToken(ctx context.Context, token string) (string, error) {
	cacheKey := validateTokenCacheKey(token)

	data, _ := s.storage.GetWithContext(ctx, cacheKey)

	if len(data) > 0 {
		isValidToken, err := strconv.ParseBool(string(data))
		if err != nil {
			return "", err
		}

		if !isValidToken {
			return "", httpx.AsBadRequest(errors.New("invalid access token"))
		}

		return token, nil
	}

	jwtoken, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		return []byte(s.cfg.JWTSecret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil || jwtoken == nil || !jwtoken.Valid {
		return "", httpx.AsBadRequest(errors.New("invalid access token"))
	}

	claims, ok := jwtoken.Claims.(jwt.MapClaims)
	if !ok {
		return "", httpx.AsBadRequest(errors.New("invalid access token"))
	}

	userIDValue, ok := claims["id"]
	if !ok {
		return "", httpx.AsBadRequest(errors.New("invalid access token"))
	}

	var userID string
	switch v := userIDValue.(type) {
	case string:
		userID = v
	case fmt.Stringer:
		userID = v.String()
	default:
		userID = fmt.Sprint(v)
	}

	role, ok := claims["role"].(string)
	if !ok || role == "" {
		return "", httpx.AsBadRequest(errors.New("invalid access token"))
	}

	user, err := s.stores.Sessions.GetSessionByToken(ctx, token)
	if err != nil {
		return "", httpx.AsBadRequest(errors.New("invalid access token"))
	}

	if userID != user.ID.String() {
		return "", httpx.AsBadRequest(errors.New("invalid access token"))
	}

	if token != user.Sessions[0].Token {
		return "", httpx.AsBadRequest(errors.New("invalid access token"))
	}

	if role != user.Role.Name {
		return "", httpx.AsBadRequest(errors.New("invalid access token"))
	}

	expValue, ok := claims["exp"]
	if !ok {
		return "", httpx.AsBadRequest(errors.New("invalid access token"))
	}

	var expUnix int64
	switch v := expValue.(type) {
	case float64:
		expUnix = int64(v)
	case int64:
		expUnix = v
	case string:
		expUnix, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			return "", httpx.AsBadRequest(errors.New("invalid access token"))
		}
	default:
		return "", httpx.AsBadRequest(errors.New("invalid access token"))
	}

	expTime := time.Unix(expUnix, 0).UTC()
	now := time.Now().UTC()

	const expirationLeeway = 30 * time.Second
	if now.After(expTime.Add(expirationLeeway)) {
		return "", httpx.AsBadRequest(errors.New("invalid access token"))
	}

	cacheTTL := expTime.Sub(now)
	cacheTTL = min(cacheTTL, 24*time.Hour)
	if cacheTTL > 0 {
		_ = s.storage.SetWithContext(ctx, cacheKey, []byte("true"), cacheTTL)
	}

	return token, nil
}

func (s *Service) RefreshToken(ctx context.Context, rawToken, ipAddress, userAgent string) (LoginInternalDTO, error) {
	ipAddress = sanitizeIP(truncate(ipAddress, 45))
	userAgent = truncate(userAgent, 255)

	oldHash, err := hashRefreshToken(rawToken)
	if err != nil {
		return LoginInternalDTO{}, httpx.AsBadRequest(errors.New("invalid refresh token"))
	}
	oldCacheKey := refreshCacheKey(oldHash)

	var (
		tokenID   uuid.UUID
		userID    uuid.UUID
		role      string
		familyID  uuid.UUID
		sessionID uuid.UUID
	)

	cached, _ := s.storage.GetWithContext(ctx, oldCacheKey)

	if len(cached) > 0 {
		// format: tokenID|userID|role|familyID|sessionID|expiresUnix
		parts := strings.SplitN(string(cached), "|", 6)
		if len(parts) != 6 {
			return LoginInternalDTO{}, httpx.AsBadRequest(errors.New("invalid refresh token"))
		}
		tokenID, err = uuid.Parse(parts[0])
		if err != nil {
			return LoginInternalDTO{}, httpx.AsBadRequest(errors.New("invalid refresh token"))
		}
		userID, err = uuid.Parse(parts[1])
		if err != nil {
			return LoginInternalDTO{}, httpx.AsBadRequest(errors.New("invalid refresh token"))
		}
		role = parts[2]
		familyID, err = uuid.Parse(parts[3])
		if err != nil {
			return LoginInternalDTO{}, httpx.AsBadRequest(errors.New("invalid refresh token"))
		}
		sessionID, err = uuid.Parse(parts[4])
		if err != nil {
			return LoginInternalDTO{}, httpx.AsBadRequest(errors.New("invalid refresh token"))
		}
		expiresUnix, parseErr := strconv.ParseInt(parts[5], 10, 64)
		if parseErr != nil {
			return LoginInternalDTO{}, httpx.AsBadRequest(errors.New("invalid refresh token"))
		}
		if time.Now().UTC().After(time.Unix(expiresUnix, 0).UTC()) {
			return LoginInternalDTO{}, httpx.AsBadRequest(errors.New("invalid refresh token"))
		}
		// The cache entry knows nothing about revocations done through the
		// database, so a revoked family must be rejected here explicitly.
		if revoked, _ := s.storage.GetWithContext(ctx, revokedFamilyCacheKey(familyID)); len(revoked) > 0 {
			return LoginInternalDTO{}, httpx.AsBadRequest(errors.New("invalid refresh token"))
		}
	} else {
		rt, dbErr := s.stores.RefreshTokens.GetRefreshTokenByHash(ctx, oldHash)
		if dbErr != nil {
			return LoginInternalDTO{}, httpx.AsBadRequest(errors.New("invalid refresh token"))
		}
		if rt.RevokedAt != nil {
			return LoginInternalDTO{}, httpx.AsBadRequest(errors.New("invalid refresh token"))
		}
		// A consumed token may be a real reuse attack or a benign concurrent
		// refresh (e.g. link preload + click racing with the same cookie). If it
		// was used within the grace period and the family is still alive, treat it
		// as benign and re-issue without revoking. Outside the window, revoke the
		// whole family.
		if rt.UsedAt != nil && time.Since(*rt.UsedAt) > s.cfg.RefreshGracePeriod {
			s.revokeRefreshFamily(ctx, rt.FamilyID)
			return LoginInternalDTO{}, httpx.AsBadRequest(errors.New("invalid refresh token"))
		}
		if time.Now().UTC().After(rt.ExpiresAt) {
			return LoginInternalDTO{}, httpx.AsBadRequest(errors.New("invalid refresh token"))
		}
		tokenID = rt.ID
		userID = rt.UserID
		role = rt.Role
		familyID = rt.FamilyID
		sessionID = rt.SessionID
	}

	// Rotation: mark current token as used before issuing new pair
	if err := s.stores.RefreshTokens.MarkRefreshTokenUsed(ctx, tokenID); err != nil {
		return LoginInternalDTO{}, err
	}
	_ = s.storage.DeleteWithContext(ctx, oldCacheKey)

	// Issue new access token
	newAccessExpiresAt := time.Now().UTC().Add(s.cfg.JWTAccessDuration)
	newJWT, err := s.CreateJWToken(userID, role, newAccessExpiresAt)
	if err != nil {
		return LoginInternalDTO{}, err
	}

	// Update session with new access token. If the session row is gone (the
	// user logged out), this refresh token family is orphaned: revoke it so the
	// cookie can't keep rotating tokens forever.
	oldSessionToken, err := s.stores.Sessions.UpdateSessionToken(ctx, sessionID, newJWT, newAccessExpiresAt)
	if err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			s.revokeRefreshFamily(ctx, familyID)
			return LoginInternalDTO{}, httpx.AsBadRequest(errors.New("invalid refresh token"))
		}
		return LoginInternalDTO{}, err
	}

	// The replaced access token may still be cached as valid; drop it so it
	// stops being accepted the moment it is rotated out.
	if oldSessionToken != "" && oldSessionToken != newJWT {
		_ = s.storage.DeleteWithContext(ctx, validateTokenCacheKey(oldSessionToken))
	}

	// Generate new refresh token (same family)
	rawNew, newHash, err := generateRefreshToken()
	if err != nil {
		return LoginInternalDTO{}, err
	}

	refreshExpiresAt := time.Now().UTC().Add(s.cfg.JWTRefreshDuration)

	var ip, ua *string
	if ipAddress != "" {
		ip = &ipAddress
	}
	if userAgent != "" {
		ua = &userAgent
	}

	newRTID, err := s.stores.RefreshTokens.CreateRefreshToken(ctx, userID, newHash, familyID, sessionID, ip, ua, refreshExpiresAt)
	if err != nil {
		return LoginInternalDTO{}, err
	}

	newCacheValue := fmt.Sprintf("%s|%s|%s|%s|%s|%d",
		newRTID, userID, role, familyID, sessionID, refreshExpiresAt.Unix(),
	)
	if ttl := time.Until(refreshExpiresAt); ttl > 0 {
		_ = s.storage.SetWithContext(ctx, refreshCacheKey(newHash), []byte(newCacheValue), ttl)
	}

	return LoginInternalDTO{
		AccessToken:      newJWT,
		RawRefreshToken:  rawNew,
		RefreshExpiresAt: refreshExpiresAt,
	}, nil
}

func (s *Service) Logout(ctx context.Context, userID uuid.UUID, accessToken, rawRefreshToken string) error {
	// Cache invalidation is best-effort: a cache hiccup must not leave the user
	// unable to log out, and the markers below close the remaining windows.
	_ = s.storage.DeleteWithContext(ctx, validateTokenCacheKey(accessToken))

	if rawRefreshToken != "" {
		if hash, err := hashRefreshToken(rawRefreshToken); err == nil {
			_ = s.storage.DeleteWithContext(ctx, refreshCacheKey(hash))
		}
	}

	// Purge every cached refresh token tied to this session and mark its
	// families revoked BEFORE deleting the session: the delete cascades to the
	// refresh_tokens rows, but the cache entries would otherwise keep the
	// refresh cookie working long after logout.
	if hashes, familyIDs, err := s.stores.RefreshTokens.GetRefreshTokenFamiliesBySession(ctx, userID, accessToken); err == nil {
		for _, hash := range hashes {
			_ = s.storage.DeleteWithContext(ctx, refreshCacheKey(hash))
		}
		seen := make(map[uuid.UUID]struct{}, len(familyIDs))
		for _, familyID := range familyIDs {
			if _, ok := seen[familyID]; ok {
				continue
			}
			seen[familyID] = struct{}{}
			_ = s.storage.SetWithContext(ctx, revokedFamilyCacheKey(familyID), []byte("1"), s.cfg.JWTRefreshDuration)
		}
	}

	if err := s.stores.Sessions.DeleteSessionByUserIDToken(ctx, userID, accessToken); err != nil {
		return err
	}

	return nil
}

// CleanupExpiredAuth prunes refresh tokens that can no longer be redeemed and
// sessions that expired with no live refresh token left. Without this, both
// tables grow unboundedly: rows are only ever deleted on explicit logout.
func (s *Service) CleanupExpiredAuth(ctx context.Context) (sessions, refreshTokens int64, err error) {
	refreshTokens, err = s.stores.RefreshTokens.DeleteExpiredRefreshTokens(ctx)
	if err != nil {
		return 0, 0, err
	}

	sessions, err = s.stores.Sessions.DeleteExpiredSessions(ctx)
	if err != nil {
		return 0, refreshTokens, err
	}

	return sessions, refreshTokens, nil
}
