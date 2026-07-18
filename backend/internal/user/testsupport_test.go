package user

import (
	"context"

	"github.com/google/uuid"
)

// fakeRepository embeds the Repository interface so tests only override the
// methods a scenario needs; calling anything else panics loudly.
type fakeRepository struct {
	Repository

	getUserByEmail            func(ctx context.Context, email string) (User, error)
	getUserByID               func(ctx context.Context, id uuid.UUID) (User, error)
	updateUser                func(ctx context.Context, id uuid.UUID, name, email, image string) (User, error)
	updateUserProfile         func(ctx context.Context, id uuid.UUID, name, preferredCurrency, image string) (User, error)
	updateUserPassword        func(ctx context.Context, userID uuid.UUID, hashedPassword string) error
	getUserPreferences        func(ctx context.Context, userID uuid.UUID) (UserPreferences, error)
	getUsersWithWeeklySummary func(ctx context.Context) ([]User, error)
}

func (f *fakeRepository) GetPreferences(ctx context.Context, userID uuid.UUID) (UserPreferences, error) {
	return f.getUserPreferences(ctx, userID)
}

func (f *fakeRepository) GetWeeklySummary(ctx context.Context) ([]User, error) {
	return f.getUsersWithWeeklySummary(ctx)
}

func (f *fakeRepository) GetByEmail(ctx context.Context, email string) (User, error) {
	return f.getUserByEmail(ctx, email)
}

func (f *fakeRepository) GetByID(ctx context.Context, id uuid.UUID) (User, error) {
	return f.getUserByID(ctx, id)
}

func (f *fakeRepository) Update(ctx context.Context, id uuid.UUID, name, email, image string) (User, error) {
	return f.updateUser(ctx, id, name, email, image)
}
