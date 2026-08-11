package user

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
	"github.com/masaya-nishimura-09/movie-log-api/internal/testutil"
	"gorm.io/gorm"
)

var testDB *gorm.DB

func TestMain(m *testing.M) {
	testDB = testutil.NewTestDB()
	code := m.Run()
	os.Exit(code)
}

func newTestRepo(t *testing.T) user.UserRepository {
	t.Helper()
	tx := testDB.Begin()
	if tx.Error != nil {
		t.Fatalf("Begin() error = %v", tx.Error)
	}
	t.Cleanup(func() { tx.Rollback() })
	return NewUserRepo(tx)
}

func TestGetByID(t *testing.T) {
	t.Run(
		"returns the user when the ID exists",
		func(t *testing.T) {
			ur := newTestRepo(t)

			ctx := context.Background()
			want := user.User{
				Username:       user.Username("Test"),
				Email:          user.Email("test@example.com"),
				HashedPassword: user.HashedPassword("testpassword"),
				Role:           user.RoleAdmin,
			}

			if err := ur.Create(ctx, &want); err != nil {
				t.Fatalf(
					"Create(ctx, %v) error = %v",
					want, err,
				)
			}

			got, err := ur.GetByID(ctx, want.ID)
			if err != nil {
				t.Fatalf(
					"GetByID(ctx, %d) (user.User, error) = %v, %v",
					want.ID, got, err,
				)
			}
			if got.ID != want.ID ||
				got.Username != want.Username ||
				got.Email != want.Email ||
				got.HashedPassword != want.HashedPassword ||
				got.Role != want.Role {
				t.Errorf(
					"GetByID(ctx, %d) (user.User, error) = %v, want %v",
					want.ID, got, want,
				)
			}
		},
	)

	t.Run(
		"returns ErrNotFound when the ID does not exist",
		func(t *testing.T) {
			ur := newTestRepo(t)

			ctx := context.Background()
			fakeID := user.ID(0)

			got, err := ur.GetByID(ctx, fakeID)
			if !errors.Is(err, exception.ErrNotFound) {
				t.Fatalf(
					"GetByID(ctx, %d) (user.User, error) = %v, %v, want %v",
					fakeID, got, err, exception.ErrNotFound,
				)
			}
			if got != nil {
				t.Errorf(
					"GetByID(ctx, %d) (user.User, error) = %v, %v, want %v",
					fakeID, got, err, exception.ErrNotFound,
				)
			}
		},
	)

	t.Run(
		"returns a wrapped error when the context is canceled",
		func(t *testing.T) {
			ur := newTestRepo(t)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			id := user.ID(1)
			got, err := ur.GetByID(ctx, id)

			if !errors.Is(err, context.Canceled) {
				t.Fatalf(
					"GetByID(ctx, %d) (user.User, error) = %v, %v, want %v",
					id, got, err, context.Canceled,
				)
			}
			if got != nil {
				t.Errorf(
					"GetByID(ctx, %d) (user.User, error) = %v, %v, want nil",
					id, got, err,
				)
			}
		},
	)
}

func TestGetByEmail(t *testing.T) {
	t.Run(
		"returns the user when the email exists",
		func(t *testing.T) {
			ur := newTestRepo(t)

			ctx := context.Background()
			want := user.User{
				Username:       user.Username("Test"),
				Email:          user.Email("test@example.com"),
				HashedPassword: user.HashedPassword("testpassword"),
				Role:           user.RoleAdmin,
			}

			if err := ur.Create(ctx, &want); err != nil {
				t.Fatalf(
					"Create(ctx, %v) error = %v",
					want, err,
				)
			}

			got, err := ur.GetByEmail(ctx, want.Email)
			if err != nil {
				t.Fatalf(
					"GetByEmail(ctx, %v) (user.User, error) = %v, %v",
					want.Email, got, err,
				)
			}
			if got.ID != want.ID ||
				got.Username != want.Username ||
				got.Email != want.Email ||
				got.HashedPassword != want.HashedPassword ||
				got.Role != want.Role {
				t.Errorf(
					"GetByEmail(ctx, %v) (user.User, error) = %v, want %v",
					want.Email, got, want,
				)
			}
		},
	)

	t.Run(
		"returns ErrNotFound when the email does not exist",
		func(t *testing.T) {
			ur := newTestRepo(t)

			ctx := context.Background()
			fakeEmail := user.Email("fake@example.com")

			got, err := ur.GetByEmail(ctx, fakeEmail)
			if !errors.Is(err, exception.ErrNotFound) {
				t.Fatalf(
					"GetByEmail(ctx, %v) (user.User, error) = %v, %v, want %v",
					fakeEmail, got, err, exception.ErrNotFound,
				)
			}
			if got != nil {
				t.Errorf(
					"GetByEmail(ctx, %v) (user.User, error) = %v, %v, want %v",
					fakeEmail, got, err, exception.ErrNotFound,
				)
			}
		},
	)

	t.Run(
		"returns a wrapped error when the context is canceled",
		func(t *testing.T) {
			ur := newTestRepo(t)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			email := user.Email("test@example.com")
			got, err := ur.GetByEmail(ctx, email)

			if !errors.Is(err, context.Canceled) {
				t.Fatalf(
					"GetByEmail(ctx, %v) (user.User, error) = %v, %v, want %v",
					email, got, err, context.Canceled,
				)
			}
			if got != nil {
				t.Errorf(
					"GetByEmail(ctx, %v) (user.User, error) = %v, %v, want nil",
					email, got, err,
				)
			}
		},
	)
}

func TestCreate(t *testing.T) {
	t.Run(
		"persists the user when a valid user is given",
		func(t *testing.T) {
			ur := newTestRepo(t)

			ctx := context.Background()
			want := user.User{
				Username:       user.Username("Test"),
				Email:          user.Email("test@example.com"),
				HashedPassword: user.HashedPassword("testpassword"),
				Role:           user.RoleAdmin,
			}

			if err := ur.Create(ctx, &want); err != nil {
				t.Fatalf(
					"Create(ctx, %v) error = %v",
					want, err,
				)
			}

			got, err := ur.GetByID(ctx, want.ID)
			if err != nil {
				t.Fatalf(
					"GetByID(ctx, %d) (user.User, error) = %v, %v",
					want.ID, got, err,
				)
			}
			if got.ID != want.ID ||
				got.Username != want.Username ||
				got.Email != want.Email ||
				got.HashedPassword != want.HashedPassword ||
				got.Role != want.Role {
				t.Errorf(
					"GetByID(ctx, %d) (user.User, error) = %v, want %v",
					want.ID, got, want,
				)
			}
		},
	)

	t.Run(
		"returns ErrAlreadyExists when the user already exists",
		func(t *testing.T) {
			ur := newTestRepo(t)

			ctx := context.Background()

			u1 := user.User{
				Username:       user.Username("Test"),
				Email:          user.Email("test@example.com"),
				HashedPassword: user.HashedPassword("testpassword"),
				Role:           user.RoleAdmin,
			}
			if err := ur.Create(ctx, &u1); err != nil {
				t.Fatalf(
					"Create(ctx, %v) error = %v",
					u1, err,
				)
			}

			u2 := user.User{
				Username:       user.Username("Test2"),
				Email:          user.Email("test@example.com"),
				HashedPassword: user.HashedPassword("testpassword"),
				Role:           user.RoleAdmin,
			}
			if err := ur.Create(ctx, &u2); !errors.Is(err, exception.ErrAlreadyExists) {
				t.Fatalf(
					"Create(ctx, %v) error = %v, want %v",
					u2, err, exception.ErrAlreadyExists,
				)
			}
		},
	)

	t.Run(
		"returns a wrapped error when the context is canceled",
		func(t *testing.T) {
			ur := newTestRepo(t)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			u := user.User{
				Username:       user.Username("Test"),
				Email:          user.Email("test@example.com"),
				HashedPassword: user.HashedPassword("testpassword"),
				Role:           user.RoleAdmin,
			}

			err := ur.Create(ctx, &u)
			if !errors.Is(err, context.Canceled) {
				t.Fatalf(
					"Create(ctx, %v) error = %v, want %v",
					u, err, context.Canceled,
				)
			}
		},
	)
}

func TestUpdate(t *testing.T) {
	t.Run(
		"updates the user when a valid user is given",
		func(t *testing.T) {
			ur := newTestRepo(t)

			ctx := context.Background()

			u1 := user.User{
				Username:       user.Username("Test"),
				Email:          user.Email("test@example.com"),
				HashedPassword: user.HashedPassword("testpassword"),
				Role:           user.RoleAdmin,
			}
			if err := ur.Create(ctx, &u1); err != nil {
				t.Fatalf(
					"Create(ctx, %v) error = %v",
					u1, err,
				)
			}

			u2 := user.User{
				ID:             u1.ID,
				Username:       user.Username("Test2"),
				Email:          user.Email("test2@example.com"),
				HashedPassword: user.HashedPassword("test2password"),
				Role:           user.RoleUser,
			}
			if err := ur.Update(ctx, &u2); err != nil {
				t.Fatalf(
					"Update(ctx, %v) error = %v",
					u2, err,
				)
			}

			got, err := ur.GetByID(ctx, u1.ID)
			if err != nil {
				t.Fatalf(
					"GetByID(ctx, %d) (user.User, error) = %v, %v",
					u1.ID, got, err,
				)
			}
			if got.ID != u2.ID ||
				got.Username != u2.Username ||
				got.Email != u2.Email ||
				got.HashedPassword != u2.HashedPassword {
				t.Errorf(
					"GetByID(ctx, %d) (user.User, error) = %v, want %v",
					u1.ID, got, u2,
				)
			}
			if got.Role != u1.Role {
				t.Errorf(
					"GetByID(ctx, %d) (user.User, error) = %v, want Role %v",
					u1.ID, got, u1.Role,
				)
			}
		},
	)

	t.Run(
		"returns ErrNotFound when the user does not exist",
		func(t *testing.T) {
			ur := newTestRepo(t)

			ctx := context.Background()

			u1 := user.User{
				ID:             user.ID(0),
				Username:       user.Username("Test"),
				Email:          user.Email("test@example.com"),
				HashedPassword: user.HashedPassword("testpassword"),
				Role:           user.RoleAdmin,
			}
			if err := ur.Update(ctx, &u1); !errors.Is(err, exception.ErrNotFound) {
				t.Fatalf(
					"Update(ctx, %v) error = %v, want %v",
					u1, err, exception.ErrNotFound,
				)
			}
		},
	)

	t.Run(
		"returns a wrapped error when the context is canceled",
		func(t *testing.T) {
			ur := newTestRepo(t)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			u := user.User{
				Username:       user.Username("Test"),
				Email:          user.Email("test@example.com"),
				HashedPassword: user.HashedPassword("testpassword"),
				Role:           user.RoleAdmin,
			}

			err := ur.Update(ctx, &u)
			if !errors.Is(err, context.Canceled) {
				t.Fatalf(
					"Update(ctx, %v) error = %v, want %v",
					u, err, context.Canceled,
				)
			}
		},
	)
}

func TestDelete(t *testing.T) {
	t.Run(
		"deletes the user when a valid ID is given",
		func(t *testing.T) {
			ur := newTestRepo(t)

			ctx := context.Background()

			u := user.User{
				Username:       user.Username("Test"),
				Email:          user.Email("test@example.com"),
				HashedPassword: user.HashedPassword("testpassword"),
				Role:           user.RoleAdmin,
			}
			if err := ur.Create(ctx, &u); err != nil {
				t.Fatalf(
					"Create(ctx, %v) error = %v",
					u, err,
				)
			}

			if err := ur.Delete(ctx, u.ID); err != nil {
				t.Fatalf(
					"Delete(ctx, %d) error = %v",
					u.ID, err,
				)
			}

			got, err := ur.GetByID(ctx, u.ID)
			if !errors.Is(err, exception.ErrNotFound) {
				t.Fatalf(
					"GetByID(ctx, %d) (user.User, error) = %v, %v, want %v",
					u.ID, got, err, exception.ErrNotFound,
				)
			}
		},
	)

	t.Run(
		"returns ErrNotFound when the user does not exist",
		func(t *testing.T) {
			ur := newTestRepo(t)

			ctx := context.Background()

			id := user.ID(0)
			if err := ur.Delete(ctx, id); !errors.Is(err, exception.ErrNotFound) {
				t.Fatalf(
					"Delete(ctx, %d) error = %v, want %v",
					id, err, exception.ErrNotFound,
				)
			}
		},
	)

	t.Run(
		"returns a wrapped error when the context is canceled",
		func(t *testing.T) {
			ur := newTestRepo(t)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			id := user.ID(0)
			err := ur.Delete(ctx, id)
			if !errors.Is(err, context.Canceled) {
				t.Fatalf(
					"Delete(ctx, %d) error = %v, want %v",
					id, err, context.Canceled,
				)
			}
		},
	)
}
