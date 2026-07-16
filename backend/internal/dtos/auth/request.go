// Package auth keeps the request DTOs of the password-reset flow, the last
// auth sub-area still served by the legacy layers. It migrates to the auth
// module in PR B of Fase 4, deleting this package.
package auth

// RequestPasswordResetDTO is the public payload to ask for a reset link.
type RequestPasswordResetDTO struct {
	Email string `json:"email" validate:"required,email,max=254"`
}

// ConfirmPasswordResetDTO is the public payload that consumes a reset token
// and sets a new password. Password bounds mirror the login DTO so login
// never rejects a password set here.
type ConfirmPasswordResetDTO struct {
	Token    string `json:"token"    validate:"required"`
	Password string `json:"password" validate:"required,min=8,max=20"`
}
