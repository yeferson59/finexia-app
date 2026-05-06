package auth

import "github.com/google/uuid"

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
