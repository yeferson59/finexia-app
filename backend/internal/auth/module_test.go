package auth

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/yeferson59/finexia-app/internal/identity"
	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

// stubUserReader is a UserReader that always resolves to the same user, enough
// to prove SetUsers reaches the service's store slot.
type stubUserReader struct{ user identity.User }

func (s stubUserReader) GetUserByID(context.Context, uuid.UUID) (identity.User, error) {
	return s.user, nil
}

func (s stubUserReader) GetUserByEmail(context.Context, string) (identity.User, error) {
	return s.user, nil
}

// TestSetUsers covers the late injection auth needs because it and the user
// module depend on each other (docs/TECH_DEBT.md #9): the reader must land in
// the service's Users slot, and a nil one must fail loudly at wiring time
// rather than on the first password reset.
func TestSetUsers(t *testing.T) {
	newTestModule := func() *Module {
		cfg := testConfig()
		stores := testStores(&fakeRepository{})
		stores.Users = nil

		return newModule(
			Deps{Cfg: cfg, Storage: newMemStorage(), Log: logger.Noop()},
			NewService(stores, cfg, newMemStorage(), nil, nil, logger.Noop()),
		)
	}

	t.Run("injects the reader into the service", func(t *testing.T) {
		m := newTestModule()
		want := identity.User{ID: uuid.New(), Email: "ana@finexia.test"}

		m.SetUsers(stubUserReader{user: want})

		got, err := m.Service().stores.Users.GetUserByEmail(context.Background(), want.Email)
		if err != nil {
			t.Fatalf("GetUserByEmail: %v", err)
		}
		if got.ID != want.ID {
			t.Errorf("got user %s, want %s", got.ID, want.ID)
		}
	})

	t.Run("panics on a nil reader", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Error("SetUsers(nil) returned normally, want panic")
			}
		}()

		newTestModule().SetUsers(nil)
	})
}
