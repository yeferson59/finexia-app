package services

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/yeferson59/finexia-app/internal/dtos/auth"
	"github.com/yeferson59/finexia-app/pkg/helpers"
)

func (s *Services) Login(ctx context.Context, email, password string) (auth.LoginResponseDTO, error) {
	user, err := s.repos.GetAccountByEmail(ctx, email)
	if err != nil {
		return auth.LoginResponseDTO{}, err
	}

	if !user.EmailVerified {
		return auth.LoginResponseDTO{}, errors.New("invalid account")
	}

	if err := user.Accounts[0].ComparePassword(password); err != nil {
		return auth.LoginResponseDTO{}, errors.New("invalid credentials")
	}

	expiresAt := time.Now().UTC().Add(s.cfg.JWTDuration)

	jwToken, err := s.CreateJWToken(user.ID, user.Role.Name, expiresAt)
	if err != nil {
		return auth.LoginResponseDTO{}, err
	}

	if err := s.repos.CreateSession(ctx, user.ID, jwToken, expiresAt); err != nil {
		return auth.LoginResponseDTO{}, err
	}

	return auth.LoginResponseDTO{
		ID:          user.ID,
		AccessToken: jwToken,
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

func (s *Services) Logout(ctx context.Context, userID uuid.UUID, token string) error {
	if err := s.storage.DeleteWithContext(ctx, "validateToken"+"-"+token); err != nil {
		return err
	}

	if err := s.repos.DeleteSessionByUserIDToken(ctx, userID, token); err != nil {
		return err
	}

	return nil
}
