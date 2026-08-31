package user

import "github.com/yeferson59/gofinance/v2/money"

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
	Name              string         `json:"name,omitzero"`
	PreferredCurrency money.Currency `json:"preferredCurrency,omitzero"`
	Image             string         `json:"image,omitzero"`
}

type UpdatePreferencesDTO struct {
	EmailAlerts   bool `json:"emailAlerts"`
	WeeklySummary bool `json:"weeklySummary"`
}

type BanUserDTO struct {
	Ban bool `json:"ban"`
}
