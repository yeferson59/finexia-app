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

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
	"github.com/yeferson59/finexia-app/internal/platform/logger"
	"github.com/yeferson59/finexia-app/internal/platform/mail"
	"github.com/yeferson59/finexia-app/pkg/helpers"
)

// Service holds the auth domain use cases. It depends only on the module's
// own consumer-defined interfaces plus platform infrastructure.
type Service struct {
	stores  Stores
	cfg     Config
	storage fiber.Storage
	mail    Mailer
	geo     GeoLocator
	log     logger.Logger
}

func NewService(stores Stores, cfg Config, storage fiber.Storage, mailService Mailer, geo GeoLocator, log logger.Logger) *Service {
	return new(Service{
		stores:  stores,
		cfg:     cfg,
		storage: storage,
		mail:    mailService,
		geo:     geo,
		log:     log,
	})
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
// while that flow waits for Fase 5. The account-lookup miss is tagged NotFound
// so httpx.FromDomain maps it to 404 by type (docs/TECH_DEBT.md #1); the
// wrong-password case keeps the frozen substring mapping ("invalid" → 400).
func (s *Service) VerifyPassword(ctx context.Context, userID uuid.UUID, currentPassword string) error {
	account, err := s.stores.Accounts.GetAccountByUserID(ctx, userID)
	if err != nil {
		return httpx.AsNotFound(errors.New("not found account"))
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
