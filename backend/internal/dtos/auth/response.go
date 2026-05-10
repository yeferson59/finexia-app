package auth

import (
	"github.com/google/uuid"
	"github.com/yeferson59/finexia-app/internal/entities"
)

type RegisterResponseDTO struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Image string `json:"image"`
}

type LoginResponseDTO struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	AccessToken string    `json:"accessToken"`
}

type SessionResponseDTO struct {
	User    entities.User    `json:"user"`
	Session entities.Session `json:"session"`
}
