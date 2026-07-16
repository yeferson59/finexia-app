package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/yeferson59/finexia-app/internal/platform/config"
	"github.com/yeferson59/finexia-app/internal/platform/logger"
	"github.com/yeferson59/finexia-app/internal/platform/mail"
	"github.com/yeferson59/finexia-app/pkg/helpers"
)

// Service holds the auth domain use cases. It depends only on the module's
// own consumer-defined interfaces plus platform infrastructure.
type Service struct {
	stores  Stores
	cfg     *config.Env
	storage fiber.Storage
	mail    Mailer
	geo     GeoLocator
	log     logger.Logger
}

func NewService(stores Stores, cfg *config.Env, storage fiber.Storage, mailService Mailer, geo GeoLocator, log logger.Logger) *Service {
	return &Service{
		stores:  stores,
		cfg:     cfg,
		storage: storage,
		mail:    mailService,
		geo:     geo,
		log:     log,
	}
}

func refreshCacheKey(hash string) string {
	return "refresh:" + hash
}

func validateTokenCacheKey(token string) string {
	return "validateToken-" + token
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
func (s *Service) isLoginLocked(ctx context.Context, email string) bool {
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
func (s *Service) recordLoginFailure(ctx context.Context, email string) {
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

func (s *Service) clearLoginFailures(ctx context.Context, email string) {
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
func (s *Service) revokeRefreshFamily(ctx context.Context, familyID uuid.UUID) {
	if hashes, err := s.stores.RefreshTokens.RevokeRefreshTokenFamily(ctx, familyID); err == nil {
		for _, hash := range hashes {
			_ = s.storage.DeleteWithContext(ctx, refreshCacheKey(hash))
		}
	}
	_ = s.storage.SetWithContext(ctx, revokedFamilyCacheKey(familyID), []byte("1"), s.cfg.JWTRefreshDuration)
}

func (s *Service) Login(ctx context.Context, email, password, ipAddress, userAgent string) (LoginInternalDTO, error) {
	ipAddress = sanitizeIP(truncate(ipAddress, 45))
	userAgent = truncate(userAgent, 255)

	if s.isLoginLocked(ctx, email) {
		return LoginInternalDTO{}, errors.New("too many failed login attempts")
	}

	user, err := s.stores.Accounts.GetAccountByEmail(ctx, email)
	if err != nil {
		s.recordLoginFailure(ctx, email)
		return LoginInternalDTO{}, errors.New("invalid credentials")
	}

	if !user.EmailVerified {
		return LoginInternalDTO{}, ErrAccountUnverified
	}

	if err := comparePassword(user.Accounts[0].Password, password); err != nil {
		s.recordLoginFailure(ctx, email)
		return LoginInternalDTO{}, errors.New("invalid credentials")
	}

	s.clearLoginFailures(ctx, email)

	// Two-factor gate: with a confirmed 2FA enrollment the password alone
	// must never yield a session. A lookup failure fails closed — silently
	// skipping the check on a database hiccup would be a 2FA bypass.
	tf, tfFound, err := s.getTwoFactor(ctx, user.ID)
	if err != nil {
		return LoginInternalDTO{}, err
	}
	if tfFound && tf.Enabled {
		pendingToken, err := s.createTwoFactorPending(ctx, user.ID)
		if err != nil {
			return LoginInternalDTO{}, err
		}
		return LoginInternalDTO{ID: user.ID, TwoFactorToken: pendingToken}, ErrTwoFactorRequired
	}

	// GetAccountByEmail does not select u.email; the login email is the same
	// address, so pass it through for the new-device alert.
	return s.issueSession(ctx, user.ID, user.Role.Name, user.Name, email, ipAddress, userAgent)
}

// issueSession creates the access token, session row, refresh token, and
// new-device alert for a fully authenticated user. Shared by password-only
// logins and logins completed through the two-factor step.
func (s *Service) issueSession(ctx context.Context, userID uuid.UUID, roleName, userName, userEmail, ipAddress, userAgent string) (LoginInternalDTO, error) {
	// Decide whether this login comes from a device the user has used before.
	// Checked against known_login_ips (not the sessions table) because a
	// session is deleted on logout and swept on expiry, while this history
	// must survive both so a returning IP is never re-flagged as new. A
	// lookup failure must not block the login nor trigger a false alarm.
	knownIP := true
	if ipAddress != "" {
		if known, ipErr := s.stores.Sessions.HasKnownLoginIP(ctx, userID, ipAddress); ipErr == nil {
			knownIP = known
		}
	}

	accessExpiresAt := time.Now().UTC().Add(s.cfg.JWTAccessDuration)

	jwToken, err := s.CreateJWToken(userID, roleName, accessExpiresAt)
	if err != nil {
		return LoginInternalDTO{}, err
	}

	var ip, ua *string
	if ipAddress != "" {
		ip = &ipAddress
	}
	if userAgent != "" {
		ua = &userAgent
	}

	sessionID, err := s.stores.Sessions.CreateSession(ctx, userID, jwToken, ip, ua, accessExpiresAt)
	if err != nil {
		return LoginInternalDTO{}, err
	}

	// Resolved after the fact so the login never waits on the geo lookup.
	go s.recordSessionLocation(ctx, sessionID, ipAddress)

	if !knownIP {
		go s.sendLoginAlert(userName, userEmail, ipAddress, userAgent)
	}

	// Remembered regardless of knownIP so the next login from this address —
	// even after this session is logged out or expires — is recognized.
	if ipAddress != "" {
		go func() {
			if err := s.stores.Sessions.RecordKnownLoginIP(context.Background(), userID, ipAddress); err != nil {
				s.log.Error(ctx, "failed to record known login ip", logger.Err(err))
			}
		}()
	}

	rawRefresh, refreshHash, err := generateRefreshToken()
	if err != nil {
		return LoginInternalDTO{}, err
	}

	familyID := uuid.New()
	refreshExpiresAt := time.Now().UTC().Add(s.cfg.JWTRefreshDuration)

	rtID, err := s.stores.RefreshTokens.CreateRefreshToken(ctx, userID, refreshHash, familyID, sessionID, ip, ua, refreshExpiresAt)
	if err != nil {
		return LoginInternalDTO{}, err
	}

	cacheValue := fmt.Sprintf("%s|%s|%s|%s|%s|%d",
		rtID, userID, roleName, familyID, sessionID, refreshExpiresAt.Unix(),
	)
	cacheTTL := time.Until(refreshExpiresAt)
	if cacheTTL > 0 {
		_ = s.storage.SetWithContext(ctx, refreshCacheKey(refreshHash), []byte(cacheValue), cacheTTL)
	}

	return LoginInternalDTO{
		ID:               userID,
		AccessToken:      jwToken,
		RawRefreshToken:  rawRefresh,
		RefreshExpiresAt: refreshExpiresAt,
	}, nil
}

// recordSessionLocation stamps the session row with the approximate location
// of its IP. Best-effort: an unknown location just leaves the column empty.
func (s *Service) recordSessionLocation(ctx context.Context, sessionID uuid.UUID, ipAddress string) {
	location := s.locateIP(ipAddress)
	if location == "" {
		return
	}
	if err := s.stores.Sessions.UpdateSessionLocation(ctx, sessionID, truncate(location, 120)); err != nil {
		s.log.Error(ctx, "failed to record session location", logger.Err(err))
	}
}

// sendLoginAlert emails the user that their account was accessed from an IP
// with no prior session. Security notices are sent regardless of the marketing
// email preferences: knowing about an unexpected login protects the account
// itself. Best-effort — a mail failure never affects the login.
func (s *Service) sendLoginAlert(userName, email, ipAddress, userAgent string) {
	if s.mail == nil {
		return
	}
	location := s.locateIP(ipAddress)
	if location == "" {
		location = "desconocida"
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
		Location:    location,
		When:        time.Now().UTC().Format("02 Jan 2006 15:04 UTC"),
		SecurityURL: s.cfg.FrontendURL + "/dashboard/settings",
	})
}

func (s *Service) CreateJWToken(userID uuid.UUID, role string, expiresAt time.Time) (string, error) {
	claims := jwt.MapClaims{
		"id":   userID,
		"role": role,
		"exp":  expiresAt.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(s.cfg.JWTSecret))
}

func (s *Service) Register(ctx context.Context, name, email, password string) (RegisterResponseDTO, error) {
	_, err := s.stores.Accounts.GetUserByEmail(ctx, email)
	if err == nil {
		return RegisterResponseDTO{}, ErrEmailAlreadyExists
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return RegisterResponseDTO{}, err
	}

	user, err := s.stores.Accounts.Register(ctx, helpers.NormalizateNames(name), email, string(passwordHash))
	if err != nil {
		return RegisterResponseDTO{}, err
	}

	s.issueEmailVerification(ctx, user.Name, user.Email)

	return RegisterResponseDTO{
		Name:  user.Name,
		Email: user.Email,
		Image: user.Image,
	}, nil
}

// VerifyPassword checks the user's current password. It backs the legacy user
// service's change-password flow (via the services.AuthService interface)
// while that flow waits for Fase 5. The two error strings are part of the
// frozen HTTP contract: httpx.FromDomain maps them by substring.
func (s *Service) VerifyPassword(ctx context.Context, userID uuid.UUID, currentPassword string) error {
	account, err := s.stores.Accounts.GetAccountByUserID(ctx, userID)
	if err != nil {
		return errors.New("not found account")
	}

	if err := comparePassword(account.Password, currentPassword); err != nil {
		return errors.New("invalid current password")
	}

	return nil
}

func (s *Service) GetSession(ctx context.Context, userID uuid.UUID, token string) (UserSessionResponseDTO, error) {
	user, err := s.stores.Sessions.GetSessionByUserIDToken(ctx, userID, token)
	if err != nil {
		return UserSessionResponseDTO{}, err
	}

	return UserSessionResponseDTO{
		User: UserResponseDTO{
			Name:              user.Name,
			Email:             user.Email,
			EmailVerified:     user.EmailVerified,
			Image:             user.Image,
			Role:              user.Role.Name,
			PreferredCurrency: user.PreferredCurrency,
			CreatedAt:         user.CreatedAt,
			UpdatedAt:         user.UpdatedAt,
		},
		Session: SessionResponseDTO{
			ID:        user.Sessions[0].ID,
			UserID:    user.Sessions[0].UserID,
			ExpiresAt: user.Sessions[0].ExpiresAt,
			IPAddress: user.Sessions[0].IPAddress,
			UserAgent: user.Sessions[0].UserAgent,
			CreatedAt: user.Sessions[0].CreatedAt,
		},
	}, nil
}

func (s *Service) ValidateToken(ctx context.Context, token string) (string, error) {
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

	user, err := s.stores.Sessions.GetSessionByToken(ctx, token)
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

func (s *Service) RefreshToken(ctx context.Context, rawToken, ipAddress, userAgent string) (LoginInternalDTO, error) {
	ipAddress = sanitizeIP(truncate(ipAddress, 45))
	userAgent = truncate(userAgent, 255)

	oldHash, err := hashRefreshToken(rawToken)
	if err != nil {
		return LoginInternalDTO{}, errors.New("invalid refresh token")
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
			return LoginInternalDTO{}, errors.New("invalid refresh token")
		}
		tokenID, err = uuid.Parse(parts[0])
		if err != nil {
			return LoginInternalDTO{}, errors.New("invalid refresh token")
		}
		userID, err = uuid.Parse(parts[1])
		if err != nil {
			return LoginInternalDTO{}, errors.New("invalid refresh token")
		}
		role = parts[2]
		familyID, err = uuid.Parse(parts[3])
		if err != nil {
			return LoginInternalDTO{}, errors.New("invalid refresh token")
		}
		sessionID, err = uuid.Parse(parts[4])
		if err != nil {
			return LoginInternalDTO{}, errors.New("invalid refresh token")
		}
		expiresUnix, parseErr := strconv.ParseInt(parts[5], 10, 64)
		if parseErr != nil {
			return LoginInternalDTO{}, errors.New("invalid refresh token")
		}
		if time.Now().UTC().After(time.Unix(expiresUnix, 0).UTC()) {
			return LoginInternalDTO{}, errors.New("invalid refresh token")
		}
		// The cache entry knows nothing about revocations done through the
		// database, so a revoked family must be rejected here explicitly.
		if revoked, _ := s.storage.GetWithContext(ctx, revokedFamilyCacheKey(familyID)); len(revoked) > 0 {
			return LoginInternalDTO{}, errors.New("invalid refresh token")
		}
	} else {
		rt, dbErr := s.stores.RefreshTokens.GetRefreshTokenByHash(ctx, oldHash)
		if dbErr != nil {
			return LoginInternalDTO{}, errors.New("invalid refresh token")
		}
		if rt.RevokedAt != nil {
			return LoginInternalDTO{}, errors.New("invalid refresh token")
		}
		// A consumed token may be a real reuse attack or a benign concurrent
		// refresh (e.g. link preload + click racing with the same cookie). If it
		// was used within the grace period and the family is still alive, treat it
		// as benign and re-issue without revoking. Outside the window, revoke the
		// whole family.
		if rt.UsedAt != nil && time.Since(*rt.UsedAt) > s.cfg.RefreshGracePeriod {
			s.revokeRefreshFamily(ctx, rt.FamilyID)
			return LoginInternalDTO{}, errors.New("invalid refresh token")
		}
		if time.Now().UTC().After(rt.ExpiresAt) {
			return LoginInternalDTO{}, errors.New("invalid refresh token")
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
			return LoginInternalDTO{}, errors.New("invalid refresh token")
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
