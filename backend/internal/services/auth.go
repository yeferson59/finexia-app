package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/yeferson59/finexia-app/internal/dtos/auth"
	"github.com/yeferson59/finexia-app/internal/mail"
	"github.com/yeferson59/finexia-app/internal/repositories"
	"github.com/yeferson59/finexia-app/pkg/helpers"
)

// ErrAccountUnverified signals a login attempt against an account whose email
// has not been verified yet, letting the handler point the client at the
// resend-verification flow instead of a generic credentials error.
var ErrAccountUnverified = errors.New("invalid account")

// ErrEmailAlreadyExists signals a registration attempt with an email that is
// already tied to an account. Unlike password reset / email verification
// requests (which never confirm whether an address exists), registration is
// already an oracle by nature: the caller is asserting they own the address
// and attempting to create an account with it, so returning a precise
// message here reveals nothing an attacker couldn't already infer from the
// request itself, and the endpoint stays behind the same rate limiter as
// login.
var ErrEmailAlreadyExists = errors.New("email already exists")

func generateRefreshToken() (raw, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return
	}
	raw = base64.URLEncoding.EncodeToString(b)
	h := sha256.Sum256(b)
	hash = hex.EncodeToString(h[:])
	return
}

func hashRefreshToken(raw string) (string, error) {
	b, err := base64.URLEncoding.DecodeString(raw)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:]), nil
}

func refreshCacheKey(hash string) string {
	return "refresh:" + hash
}

func validateTokenCacheKey(token string) string {
	return "validateToken-" + token
}

// truncate keeps a string within the column limits (ip VARCHAR(45),
// user_agent VARCHAR(255)) so an oversized header can never fail the insert.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

// loginAttemptsCacheKey tracks consecutive failed logins per email. The email
// is hashed so raw addresses never appear as cache keys.
func loginAttemptsCacheKey(email string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(email))))
	return "login_fail:" + hex.EncodeToString(sum[:])
}

// isLoginLocked reports whether the account accumulated too many failed
// attempts. It counts by email regardless of whether the account exists, so
// the lockout response never leaks which emails are registered.
func (s *Services) isLoginLocked(ctx context.Context, email string) bool {
	if s.cfg.MaxLoginAttempts <= 0 {
		return false
	}
	data, _ := s.storage.GetWithContext(ctx, loginAttemptsCacheKey(email))
	if len(data) == 0 {
		return false
	}
	count, err := strconv.Atoi(string(data))
	return err == nil && count >= s.cfg.MaxLoginAttempts
}

// recordLoginFailure increments the failure counter. Each failure renews the
// TTL, so an ongoing brute-force attempt keeps the lock alive.
func (s *Services) recordLoginFailure(ctx context.Context, email string) {
	if s.cfg.MaxLoginAttempts <= 0 {
		return
	}
	key := loginAttemptsCacheKey(email)
	count := 0
	if data, _ := s.storage.GetWithContext(ctx, key); len(data) > 0 {
		count, _ = strconv.Atoi(string(data))
	}
	_ = s.storage.SetWithContext(ctx, key, []byte(strconv.Itoa(count+1)), s.cfg.LoginLockout)
}

func (s *Services) clearLoginFailures(ctx context.Context, email string) {
	_ = s.storage.DeleteWithContext(ctx, loginAttemptsCacheKey(email))
}

// revokedFamilyCacheKey marks a whole refresh-token family as revoked in cache.
// The per-token cache entries skip the database, so without this marker a
// revoked family's newest token would keep refreshing from cache until its TTL
// ran out.
func revokedFamilyCacheKey(familyID uuid.UUID) string {
	return "revoked_family:" + familyID.String()
}

// revokeRefreshFamily revokes the family in the database and purges every
// cached token entry, then sets the revocation marker as a backstop in case a
// cache delete fails or races with a concurrent refresh.
func (s *Services) revokeRefreshFamily(ctx context.Context, familyID uuid.UUID) {
	if hashes, err := s.repos.RevokeRefreshTokenFamily(ctx, familyID); err == nil {
		for _, hash := range hashes {
			_ = s.storage.DeleteWithContext(ctx, refreshCacheKey(hash))
		}
	}
	_ = s.storage.SetWithContext(ctx, revokedFamilyCacheKey(familyID), []byte("1"), s.cfg.JWTRefreshDuration)
}

func (s *Services) Login(ctx context.Context, email, password, ipAddress, userAgent string) (auth.LoginInternalDTO, error) {
	ipAddress = truncate(ipAddress, 45)
	userAgent = truncate(userAgent, 255)

	if s.isLoginLocked(ctx, email) {
		return auth.LoginInternalDTO{}, errors.New("too many failed login attempts")
	}

	user, err := s.repos.GetAccountByEmail(ctx, email)
	if err != nil {
		s.recordLoginFailure(ctx, email)
		return auth.LoginInternalDTO{}, errors.New("invalid credentials")
	}

	if !user.EmailVerified {
		return auth.LoginInternalDTO{}, ErrAccountUnverified
	}

	if err := user.Accounts[0].ComparePassword(password); err != nil {
		s.recordLoginFailure(ctx, email)
		return auth.LoginInternalDTO{}, errors.New("invalid credentials")
	}

	s.clearLoginFailures(ctx, email)

	// Two-factor gate: with a confirmed 2FA enrollment the password alone
	// must never yield a session. A lookup failure fails closed — silently
	// skipping the check on a database hiccup would be a 2FA bypass.
	tf, tfFound, err := s.getTwoFactor(ctx, user.ID)
	if err != nil {
		return auth.LoginInternalDTO{}, err
	}
	if tfFound && tf.Enabled {
		pendingToken, err := s.createTwoFactorPending(ctx, user.ID)
		if err != nil {
			return auth.LoginInternalDTO{}, err
		}
		return auth.LoginInternalDTO{ID: user.ID, TwoFactorToken: pendingToken}, ErrTwoFactorRequired
	}

	// GetAccountByEmail does not select u.email; the login email is the same
	// address, so pass it through for the new-device alert.
	return s.issueSession(ctx, user.ID, user.Role.Name, user.Name, email, ipAddress, userAgent)
}

// issueSession creates the access token, session row, refresh token, and
// new-device alert for a fully authenticated user. Shared by password-only
// logins and logins completed through the two-factor step.
func (s *Services) issueSession(ctx context.Context, userID uuid.UUID, roleName, userName, userEmail, ipAddress, userAgent string) (auth.LoginInternalDTO, error) {
	// Decide whether this login comes from a device the user has used before.
	// Must run before CreateSession records this login's IP, and a lookup
	// failure must not block the login nor trigger a false alarm.
	knownIP := true
	if ipAddress != "" {
		if known, ipErr := s.repos.HasSessionFromIP(ctx, userID, ipAddress); ipErr == nil {
			knownIP = known
		}
	}

	accessExpiresAt := time.Now().UTC().Add(s.cfg.JWTAccessDuration)

	jwToken, err := s.CreateJWToken(userID, roleName, accessExpiresAt)
	if err != nil {
		return auth.LoginInternalDTO{}, err
	}

	var ip, ua *string
	if ipAddress != "" {
		ip = &ipAddress
	}
	if userAgent != "" {
		ua = &userAgent
	}

	sessionID, err := s.repos.CreateSession(ctx, userID, jwToken, ip, ua, accessExpiresAt)
	if err != nil {
		return auth.LoginInternalDTO{}, err
	}

	if !knownIP {
		go s.sendLoginAlert(userName, userEmail, ipAddress, userAgent)
	}

	rawRefresh, refreshHash, err := generateRefreshToken()
	if err != nil {
		return auth.LoginInternalDTO{}, err
	}

	familyID := uuid.New()
	refreshExpiresAt := time.Now().UTC().Add(s.cfg.JWTRefreshDuration)

	rtID, err := s.repos.CreateRefreshToken(ctx, userID, refreshHash, familyID, sessionID, ip, ua, refreshExpiresAt)
	if err != nil {
		return auth.LoginInternalDTO{}, err
	}

	cacheValue := fmt.Sprintf("%s|%s|%s|%s|%s|%d",
		rtID, userID, roleName, familyID, sessionID, refreshExpiresAt.Unix(),
	)
	cacheTTL := time.Until(refreshExpiresAt)
	if cacheTTL > 0 {
		_ = s.storage.SetWithContext(ctx, refreshCacheKey(refreshHash), []byte(cacheValue), cacheTTL)
	}

	return auth.LoginInternalDTO{
		ID:               userID,
		AccessToken:      jwToken,
		RawRefreshToken:  rawRefresh,
		RefreshExpiresAt: refreshExpiresAt,
	}, nil
}

// sendLoginAlert emails the user that their account was accessed from an IP
// with no prior session. Security notices are sent regardless of the marketing
// email preferences: knowing about an unexpected login protects the account
// itself. Best-effort — a mail failure never affects the login.
func (s *Services) sendLoginAlert(userName, email, ipAddress, userAgent string) {
	if s.mail == nil {
		return
	}
	if ipAddress == "" {
		ipAddress = "desconocida"
	}
	if userAgent == "" {
		userAgent = "desconocido"
	}

	_ = s.mail.SendSecurityAlert(email, mail.SecurityAlertData{
		UserName:    userName,
		Event:       "nuevo inicio de sesión",
		Detail:      "Se inició sesión en tu cuenta desde una dirección que no habíamos visto antes.",
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		When:        time.Now().UTC().Format("02 Jan 2006 15:04 UTC"),
		SecurityURL: s.cfg.FrontendURL + "/dashboard/settings",
	})
}

func (s *Services) CreateJWToken(userID uuid.UUID, role string, expiresAt time.Time) (string, error) {
	claims := jwt.MapClaims{
		"id":   userID,
		"role": role,
		"exp":  expiresAt.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(s.cfg.JWTSecret))
}

func (s *Services) Register(ctx context.Context, name, email, password string) (auth.RegisterResponseDTO, error) {
	_, err := s.repos.GetUserByEmail(ctx, email)
	if err == nil {
		return auth.RegisterResponseDTO{}, ErrEmailAlreadyExists
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return auth.RegisterResponseDTO{}, err
	}

	user, err := s.repos.Register(ctx, helpers.NormalizateNames(name), email, string(passwordHash))
	if err != nil {
		return auth.RegisterResponseDTO{}, err
	}

	s.issueEmailVerification(ctx, user.Name, user.Email)

	return auth.RegisterResponseDTO{
		Name:  user.Name,
		Email: user.Email,
		Image: user.Image,
	}, nil
}

func (s *Services) GetSession(ctx context.Context, userID uuid.UUID, token string) (auth.UserSessionResponseDTO, error) {
	user, err := s.repos.GetSessionByUserIDToken(ctx, userID, token)
	if err != nil {
		return auth.UserSessionResponseDTO{}, err
	}

	return auth.UserSessionResponseDTO{
		User: auth.UserResponseDTO{
			Name:              user.Name,
			Email:             user.Email,
			EmailVerified:     user.EmailVerified,
			Image:             user.Image,
			Role:              user.Role.Name,
			PreferredCurrency: user.PreferredCurrency,
			CreatedAt:         user.CreatedAt,
			UpdatedAt:         user.UpdatedAt,
		},
		Session: auth.SessionResponseDTO{
			ID:        user.Sessions[0].ID,
			UserID:    user.Sessions[0].UserID,
			ExpiresAt: user.Sessions[0].ExpiresAt,
			IPAddress: user.Sessions[0].IPAddress,
			UserAgent: user.Sessions[0].UserAgent,
			CreatedAt: user.Sessions[0].CreatedAt,
		},
	}, nil
}

func (s *Services) ValidateToken(ctx context.Context, token string) (string, error) {
	cacheKey := validateTokenCacheKey(token)

	data, _ := s.storage.GetWithContext(ctx, cacheKey)

	if len(data) > 0 {
		isValidToken, err := strconv.ParseBool(string(data))
		if err != nil {
			return "", err
		}

		if !isValidToken {
			return "", errors.New("invalid access token")
		}

		return token, nil
	}

	jwtoken, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		return []byte(s.cfg.JWTSecret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil || jwtoken == nil || !jwtoken.Valid {
		return "", errors.New("invalid access token")
	}

	claims, ok := jwtoken.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid access token")
	}

	userIDValue, ok := claims["id"]
	if !ok {
		return "", errors.New("invalid access token")
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
		return "", errors.New("invalid access token")
	}

	user, err := s.repos.GetSessionByToken(ctx, token)
	if err != nil {
		return "", errors.New("invalid access token")
	}

	if userID != user.ID.String() {
		return "", errors.New("invalid access token")
	}

	if token != user.Sessions[0].Token {
		return "", errors.New("invalid access token")
	}

	if role != user.Role.Name {
		return "", errors.New("invalid access token")
	}

	expValue, ok := claims["exp"]
	if !ok {
		return "", errors.New("invalid access token")
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
			return "", errors.New("invalid access token")
		}
	default:
		return "", errors.New("invalid access token")
	}

	expTime := time.Unix(expUnix, 0).UTC()
	now := time.Now().UTC()

	const expirationLeeway = 30 * time.Second
	if now.After(expTime.Add(expirationLeeway)) {
		return "", errors.New("invalid access token")
	}

	cacheTTL := expTime.Sub(now)
	cacheTTL = min(cacheTTL, 24*time.Hour)
	if cacheTTL > 0 {
		_ = s.storage.SetWithContext(ctx, cacheKey, []byte("true"), cacheTTL)
	}

	return token, nil
}

func (s *Services) RefreshToken(ctx context.Context, rawToken, ipAddress, userAgent string) (auth.LoginInternalDTO, error) {
	ipAddress = truncate(ipAddress, 45)
	userAgent = truncate(userAgent, 255)

	oldHash, err := hashRefreshToken(rawToken)
	if err != nil {
		return auth.LoginInternalDTO{}, errors.New("invalid refresh token")
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
			return auth.LoginInternalDTO{}, errors.New("invalid refresh token")
		}
		tokenID, err = uuid.Parse(parts[0])
		if err != nil {
			return auth.LoginInternalDTO{}, errors.New("invalid refresh token")
		}
		userID, err = uuid.Parse(parts[1])
		if err != nil {
			return auth.LoginInternalDTO{}, errors.New("invalid refresh token")
		}
		role = parts[2]
		familyID, err = uuid.Parse(parts[3])
		if err != nil {
			return auth.LoginInternalDTO{}, errors.New("invalid refresh token")
		}
		sessionID, err = uuid.Parse(parts[4])
		if err != nil {
			return auth.LoginInternalDTO{}, errors.New("invalid refresh token")
		}
		expiresUnix, parseErr := strconv.ParseInt(parts[5], 10, 64)
		if parseErr != nil {
			return auth.LoginInternalDTO{}, errors.New("invalid refresh token")
		}
		if time.Now().UTC().After(time.Unix(expiresUnix, 0).UTC()) {
			return auth.LoginInternalDTO{}, errors.New("invalid refresh token")
		}
		// The cache entry knows nothing about revocations done through the
		// database, so a revoked family must be rejected here explicitly.
		if revoked, _ := s.storage.GetWithContext(ctx, revokedFamilyCacheKey(familyID)); len(revoked) > 0 {
			return auth.LoginInternalDTO{}, errors.New("invalid refresh token")
		}
	} else {
		rt, dbErr := s.repos.GetRefreshTokenByHash(ctx, oldHash)
		if dbErr != nil {
			return auth.LoginInternalDTO{}, errors.New("invalid refresh token")
		}
		if rt.RevokedAt != nil {
			return auth.LoginInternalDTO{}, errors.New("invalid refresh token")
		}
		// A consumed token may be a real reuse attack or a benign concurrent
		// refresh (e.g. link preload + click racing with the same cookie). If it
		// was used within the grace period and the family is still alive, treat it
		// as benign and re-issue without revoking. Outside the window, revoke the
		// whole family.
		if rt.UsedAt != nil && time.Since(*rt.UsedAt) > s.cfg.RefreshGracePeriod {
			s.revokeRefreshFamily(ctx, rt.FamilyID)
			return auth.LoginInternalDTO{}, errors.New("invalid refresh token")
		}
		if time.Now().UTC().After(rt.ExpiresAt) {
			return auth.LoginInternalDTO{}, errors.New("invalid refresh token")
		}
		tokenID = rt.ID
		userID = rt.UserID
		role = rt.Role
		familyID = rt.FamilyID
		sessionID = rt.SessionID
	}

	// Rotation: mark current token as used before issuing new pair
	if err := s.repos.MarkRefreshTokenUsed(ctx, tokenID); err != nil {
		return auth.LoginInternalDTO{}, err
	}
	_ = s.storage.DeleteWithContext(ctx, oldCacheKey)

	// Issue new access token
	newAccessExpiresAt := time.Now().UTC().Add(s.cfg.JWTAccessDuration)
	newJWT, err := s.CreateJWToken(userID, role, newAccessExpiresAt)
	if err != nil {
		return auth.LoginInternalDTO{}, err
	}

	// Update session with new access token. If the session row is gone (the
	// user logged out), this refresh token family is orphaned: revoke it so the
	// cookie can't keep rotating tokens forever.
	oldSessionToken, err := s.repos.UpdateSessionToken(ctx, sessionID, newJWT, newAccessExpiresAt)
	if err != nil {
		if errors.Is(err, repositories.ErrSessionNotFound) {
			s.revokeRefreshFamily(ctx, familyID)
			return auth.LoginInternalDTO{}, errors.New("invalid refresh token")
		}
		return auth.LoginInternalDTO{}, err
	}

	// The replaced access token may still be cached as valid; drop it so it
	// stops being accepted the moment it is rotated out.
	if oldSessionToken != "" && oldSessionToken != newJWT {
		_ = s.storage.DeleteWithContext(ctx, validateTokenCacheKey(oldSessionToken))
	}

	// Generate new refresh token (same family)
	rawNew, newHash, err := generateRefreshToken()
	if err != nil {
		return auth.LoginInternalDTO{}, err
	}

	refreshExpiresAt := time.Now().UTC().Add(s.cfg.JWTRefreshDuration)

	var ip, ua *string
	if ipAddress != "" {
		ip = &ipAddress
	}
	if userAgent != "" {
		ua = &userAgent
	}

	newRTID, err := s.repos.CreateRefreshToken(ctx, userID, newHash, familyID, sessionID, ip, ua, refreshExpiresAt)
	if err != nil {
		return auth.LoginInternalDTO{}, err
	}

	newCacheValue := fmt.Sprintf("%s|%s|%s|%s|%s|%d",
		newRTID, userID, role, familyID, sessionID, refreshExpiresAt.Unix(),
	)
	if ttl := time.Until(refreshExpiresAt); ttl > 0 {
		_ = s.storage.SetWithContext(ctx, refreshCacheKey(newHash), []byte(newCacheValue), ttl)
	}

	return auth.LoginInternalDTO{
		AccessToken:      newJWT,
		RawRefreshToken:  rawNew,
		RefreshExpiresAt: refreshExpiresAt,
	}, nil
}

func (s *Services) Logout(ctx context.Context, userID uuid.UUID, accessToken, rawRefreshToken string) error {
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
	if hashes, familyIDs, err := s.repos.GetRefreshTokenFamiliesBySession(ctx, userID, accessToken); err == nil {
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

	if err := s.repos.DeleteSessionByUserIDToken(ctx, userID, accessToken); err != nil {
		return err
	}

	return nil
}

// CleanupExpiredAuth prunes refresh tokens that can no longer be redeemed and
// sessions that expired with no live refresh token left. Without this, both
// tables grow unboundedly: rows are only ever deleted on explicit logout.
func (s *Services) CleanupExpiredAuth(ctx context.Context) (sessions, refreshTokens int64, err error) {
	refreshTokens, err = s.repos.DeleteExpiredRefreshTokens(ctx)
	if err != nil {
		return 0, 0, err
	}

	sessions, err = s.repos.DeleteExpiredSessions(ctx)
	if err != nil {
		return 0, refreshTokens, err
	}

	return sessions, refreshTokens, nil
}
