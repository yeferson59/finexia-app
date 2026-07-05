package auth

type RegisterRequestDTO struct {
	Name     string `json:"name" validate:"required,min=2,max=254"`
	Email    string `json:"email" validate:"required,email,min=2,max=254"`
	Password string `json:"password" validate:"required,min=8,max=20"`
}

type LoginRequestDTO struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=20"`
}

// RequestPasswordResetDTO is the public payload to ask for a reset link.
type RequestPasswordResetDTO struct {
	Email string `json:"email" validate:"required,email,max=254"`
}

// ConfirmPasswordResetDTO is the public payload that consumes a reset token
// and sets a new password. Password bounds mirror LoginRequestDTO so login
// never rejects a password set here.
type ConfirmPasswordResetDTO struct {
	Token    string `json:"token"    validate:"required"`
	Password string `json:"password" validate:"required,min=8,max=20"`
}
