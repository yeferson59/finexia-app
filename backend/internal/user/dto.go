package user

type CreateDTO struct {
	Name  string `json:"name" validate:"required,min=2"`
	Email string `json:"email" validate:"required,email,min=2"`
}

type UpdateDTO struct {
	Name  string `json:"name,omitzero"`
	Email string `json:"email,omitzero"`
	Image string `json:"image,omitzero"`
}

type UpdateProfileDTO struct {
	Name              string `json:"name,omitzero"`
	PreferredCurrency string `json:"preferredCurrency,omitzero"`
	Image             string `json:"image,omitzero"`
}

type UpdatePreferencesDTO struct {
	EmailAlerts   bool `json:"emailAlerts"`
	WeeklySummary bool `json:"weeklySummary"`
}

type ChangePasswordDTO struct {
	// NewPassword keeps the same bounds as RegisterRequestDTO/LoginRequestDTO
	// (min=8,max=20); otherwise a user could set a password login would reject.
	CurrentPassword string `json:"currentPassword" validate:"required,min=8"`
	NewPassword     string `json:"newPassword"     validate:"required,min=8,max=20"`
}

type BanUserDTO struct {
	Ban bool `json:"ban"`
}

// InviteUserDTO is the admin-side payload to invite someone. Name is optional
// (derived from the email when absent, and the invitee can set their real name
// on accept); Role defaults to "customer" and is whitelisted server-side.
type InviteUserDTO struct {
	Email string `json:"email" validate:"required,email,max=254"`
	Name  string `json:"name"  validate:"omitempty,max=254"`
	Role  string `json:"role"  validate:"omitempty,oneof=customer admin"`
}
