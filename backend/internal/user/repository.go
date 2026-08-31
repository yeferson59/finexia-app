package user

import (
	"context"

	"uuid"

	"github.com/yeferson59/finexia-app/internal/identity"
	"github.com/yeferson59/gofinance/v2/money"
)

type Repository interface {
	List(ctx context.Context, offset, limit uint) ([]identity.User, uint, error)
	GetByID(ctx context.Context, id uuid.UUID) (identity.User, error)
	Create(ctx context.Context, name, email string) (identity.User, error)
	Update(ctx context.Context, id uuid.UUID, name, email, image string) (identity.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Ban(ctx context.Context, id uuid.UUID, ban bool) error
	GetByEmail(ctx context.Context, email string) (identity.User, error)
	UpdateProfile(ctx context.Context, id uuid.UUID, name, image string, preferredCurrency money.Currency) (identity.User, error)
	UpdateImage(ctx context.Context, id uuid.UUID, image string) (identity.User, error)
	GetPreferences(ctx context.Context, userID uuid.UUID) (UserPreferences, error)
	UpsertPreferences(ctx context.Context, userID uuid.UUID, emailAlerts, weeklySummary bool) (UserPreferences, error)
	GetWeeklySummary(ctx context.Context) ([]identity.User, error)
}
