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
	"github.com/yeferson59/finexia-app/pkg/helpers"
)

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

func (s *Services) Login(ctx context.Context, email, password string) (auth.LoginInternalDTO, error) {
	user, err := s.repos.GetAccountByEmail(ctx, email)
	if err != nil {
		return auth.LoginInternalDTO{}, err
	}

	if !user.EmailVerified {
		return auth.LoginInternalDTO{}, errors.New("invalid account")
	}

	if err := user.Accounts[0].ComparePassword(password); err != nil {
		return auth.LoginInternalDTO{}, errors.New("invalid credentials")
	}

	accessExpiresAt := time.Now().UTC().Add(s.cfg.JWTAccessDuration)

	jwToken, err := s.CreateJWToken(user.ID, user.Role.Name, accessExpiresAt)
	if err != nil {
		return auth.LoginInternalDTO{}, err
	}

	sessionID, err := s.repos.CreateSession(ctx, user.ID, jwToken, accessExpiresAt)
	if err != nil {
		return auth.LoginInternalDTO{}, err
	}

	rawRefresh, refreshHash, err := generateRefreshToken()
	if err != nil {
		return auth.LoginInternalDTO{}, err
	}

	familyID := uuid.New()
	refreshExpiresAt := time.Now().UTC().Add(s.cfg.JWTRefreshDuration)

	rtID, err := s.repos.CreateRefreshToken(ctx, user.ID, refreshHash, familyID, sessionID, nil, nil, refreshExpiresAt)
	if err != nil {
		return auth.LoginInternalDTO{}, err
	}

	cacheValue := fmt.Sprintf("%s|%s|%s|%s|%s|%d",
		rtID, user.ID, user.Role.Name, familyID, sessionID, refreshExpiresAt.Unix(),
	)
	cacheTTL := time.Until(refreshExpiresAt)
	if cacheTTL > 0 {
		_ = s.storage.SetWithContext(ctx, refreshCacheKey(refreshHash), []byte(cacheValue), cacheTTL)
	}

	return auth.LoginInternalDTO{
		ID:               user.ID,
		AccessToken:      jwToken,
		RawRefreshToken:  rawRefresh,
		RefreshExpiresAt: refreshExpiresAt,
	}, nil
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
		return auth.RegisterResponseDTO{}, errors.New("user existing")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return auth.RegisterResponseDTO{}, err
	}

	user, err := s.repos.Register(ctx, helpers.NormalizateNames(name), email, string(passwordHash))
	if err != nil {
		return auth.RegisterResponseDTO{}, err
	}

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
	cacheKey := "validateToken" + "-" + token

	data, err := s.storage.GetWithContext(ctx, cacheKey)
	if err != nil {
		return "", err
	}

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
		if err := s.storage.SetWithContext(ctx, cacheKey, []byte("true"), cacheTTL); err != nil {
			return "", err
		}
	}

	return token, nil
}

func (s *Services) RefreshToken(ctx context.Context, rawToken, ipAddress, userAgent string) (auth.LoginInternalDTO, error) {
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

	cached, err := s.storage.GetWithContext(ctx, oldCacheKey)
	if err != nil {
		return auth.LoginInternalDTO{}, err
	}

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
	} else {
		rt, dbErr := s.repos.GetRefreshTokenByHash(ctx, oldHash)
		if dbErr != nil {
			return auth.LoginInternalDTO{}, errors.New("invalid refresh token")
		}
		// Token reuse attack: token already consumed → revoke entire family
		if rt.UsedAt != nil {
			_ = s.repos.RevokeRefreshTokenFamily(ctx, rt.FamilyID)
			return auth.LoginInternalDTO{}, errors.New("invalid refresh token")
		}
		if rt.RevokedAt != nil {
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

	// Update session with new access token
	if err := s.repos.UpdateSessionToken(ctx, sessionID, newJWT, newAccessExpiresAt); err != nil {
		return auth.LoginInternalDTO{}, err
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
	if err := s.storage.DeleteWithContext(ctx, "validateToken"+"-"+accessToken); err != nil {
		return err
	}

	if rawRefreshToken != "" {
		if hash, err := hashRefreshToken(rawRefreshToken); err == nil {
			_ = s.storage.DeleteWithContext(ctx, refreshCacheKey(hash))
		}
	}

	if err := s.repos.DeleteSessionByUserIDToken(ctx, userID, accessToken); err != nil {
		return err
	}

	return nil
}
