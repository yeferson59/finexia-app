package user

import (
	"context"

	"uuid"

	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/identity"
	"github.com/yeferson59/finexia-app/internal/platform/objectstore"
)

// fakeRepository embeds the Repository interface so tests only override the
// methods a scenario needs; calling anything else panics loudly.
type fakeRepository struct {
	Repository

	getUserByEmail            func(ctx context.Context, email string) (identity.User, error)
	getUserByID               func(ctx context.Context, id uuid.UUID) (identity.User, error)
	updateUser                func(ctx context.Context, id uuid.UUID, name, email, image string) (identity.User, error)
	updateUserProfile         func(ctx context.Context, id uuid.UUID, name, image string, preferredCurrency money.Currency) (identity.User, error)
	updateImage               func(ctx context.Context, id uuid.UUID, image string) (identity.User, error)
	getUserPreferences        func(ctx context.Context, userID uuid.UUID) (UserPreferences, error)
	getUsersWithWeeklySummary func(ctx context.Context) ([]identity.User, error)
}

func (f *fakeRepository) GetPreferences(ctx context.Context, userID uuid.UUID) (UserPreferences, error) {
	return f.getUserPreferences(ctx, userID)
}

func (f *fakeRepository) GetWeeklySummary(ctx context.Context) ([]identity.User, error) {
	return f.getUsersWithWeeklySummary(ctx)
}

func (f *fakeRepository) GetByEmail(ctx context.Context, email string) (identity.User, error) {
	return f.getUserByEmail(ctx, email)
}

func (f *fakeRepository) GetByID(ctx context.Context, id uuid.UUID) (identity.User, error) {
	return f.getUserByID(ctx, id)
}

func (f *fakeRepository) Update(ctx context.Context, id uuid.UUID, name, email, image string) (identity.User, error) {
	return f.updateUser(ctx, id, name, email, image)
}

func (f *fakeRepository) UpdateProfile(ctx context.Context, id uuid.UUID, name, image string, preferredCurrency money.Currency) (identity.User, error) {
	return f.updateUserProfile(ctx, id, name, image, preferredCurrency)
}

func (f *fakeRepository) UpdateImage(ctx context.Context, id uuid.UUID, image string) (identity.User, error) {
	return f.updateImage(ctx, id, image)
}

// fakeStore embeds objectstore.Store for the same reason fakeRepository embeds
// Repository: a scenario only writes the calls it makes.
type fakeStore struct {
	objectstore.Store

	put func(ctx context.Context, name, contentType string, body []byte) error
}

func (f *fakeStore) Put(ctx context.Context, name, contentType string, body []byte) error {
	return f.put(ctx, name, contentType, body)
}
