package user

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	List(ctx context.Context, offset, limit uint) ([]User, uint, error)
	GetByID(ctx context.Context, id uuid.UUID) (User, error)
	Create(ctx context.Context, name, email string) (User, error)
	Update(ctx context.Context, id uuid.UUID, name, email, image string) (User, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Ban(ctx context.Context, id uuid.UUID, ban bool) error
	GetByEmail(ctx context.Context, email string) (User, error)
	UpdateProfile(ctx context.Context, id uuid.UUID, name, preferredCurrency, image string) (User, error)
	UpdateImage(ctx context.Context, id uuid.UUID, image string) (User, error)
	GetPreferences(ctx context.Context, userID uuid.UUID) (UserPreferences, error)
	UpsertPreferences(ctx context.Context, userID uuid.UUID, emailAlerts, weeklySummary bool) (UserPreferences, error)
	UpdatePassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error
	GetWeeklySummary(ctx context.Context) ([]User, error)
}
