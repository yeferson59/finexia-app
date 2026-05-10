package services

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/yeferson59/finexia-app/internal/dtos/auth"
	"github.com/yeferson59/finexia-app/pkg/helpers"
)

func (s *Services) Login(ctx context.Context, email, password string) (auth.LoginResponseDTO, error) {
	user, role, account, err := s.repos.GetAccountByEmail(ctx, email)
	if err != nil {
		return auth.LoginResponseDTO{}, err
	}

	if !user.EmailVerified {
		return auth.LoginResponseDTO{}, errors.New("invalid account")
	}

	if err := account.ComparePassword(password); err != nil {
		return auth.LoginResponseDTO{}, errors.New("invalid credentials")
	}

	expiresAt := time.Now().Add(s.cfg.JWTDuration)

	jwToken, err := s.CreateJWToken(user.ID, role, expiresAt)
	if err != nil {
		return auth.LoginResponseDTO{}, err
	}

	if err := s.repos.CreateSession(ctx, user.ID, jwToken, expiresAt); err != nil {
		return auth.LoginResponseDTO{}, err
	}

	return auth.LoginResponseDTO{
		ID:          user.ID,
		Name:        user.Name,
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

func (s *Services) GetSession(ctx context.Context, userID uuid.UUID, role, token string) (auth.UserSessionResponseDTO, error) {
	user, session, err := s.repos.GetSessionByUserIDToken(ctx, userID, token)
	if err != nil {
		return auth.UserSessionResponseDTO{}, err
	}

	return auth.UserSessionResponseDTO{
		User: auth.UserResponseDTO{
			Name:              user.Name,
			Email:             user.Email,
			EmailVerified:     user.EmailVerified,
			Image:             user.Image,
			Role:              role,
			PreferredCurrency: user.PreferredCurrency,
			CreatedAt:         user.CreatedAt,
			UpdatedAt:         user.UpdatedAt,
		},
		Session: auth.SessionResponseDTO{
			ID:        session.ID,
			UserID:    session.UserID,
			ExpiresAt: session.ExpiresAt,
			IPAddress: session.IPAddress,
			UserAgent: session.UserAgent,
			CreatedAt: session.CreatedAt,
		},
	}, nil
}
