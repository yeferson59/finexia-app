package marketing

type Waitlist struct {
	Email string `json:"email" validate:"required,email"`
}
