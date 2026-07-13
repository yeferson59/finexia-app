package marketing

// waitlistRequest is the body of POST /marketing/waitlists.
type waitlistRequest struct {
	Email string `json:"email" validate:"required,email"`
}
