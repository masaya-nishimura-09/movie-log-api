package user

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
	"github.com/masaya-nishimura-09/movie-log-api/internal/testutil"
)

func TestGetByID(t *testing.T) {
	t.Run(
		"returns the user when the ID exists",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			r := NewRepository(tx)

			ctx := context.Background()
			want := user.User{
				Username:       user.Username("Test"),
				Email:          user.Email("test@example.com"),
				HashedPassword: user.HashedPassword("testpassword"),
				Role:           user.RoleAdmin,
			}

			if err := r.Create(ctx, &want); err != nil {
				t.Fatalf(
					"Create(ctx, %v) error = %v",
					want, err,
				)
			}

			got, err := r.GetByID(ctx, want.ID)
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
			tx := testutil.BeginTx(t, testDB)
			r := NewRepository(tx)

			ctx := context.Background()
			fakeID := user.ID(999999)

			got, err := r.GetByID(ctx, fakeID)
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
			tx := testutil.BeginTx(t, testDB)
			r := NewRepository(tx)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			id := user.ID(1)
			got, err := r.GetByID(ctx, id)

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
			tx := testutil.BeginTx(t, testDB)
			r := NewRepository(tx)

			ctx := context.Background()
			want := user.User{
				Username:       user.Username("Test"),
				Email:          user.Email("test@example.com"),
				HashedPassword: user.HashedPassword("testpassword"),
				Role:           user.RoleAdmin,
			}

			if err := r.Create(ctx, &want); err != nil {
				t.Fatalf(
					"Create(ctx, %v) error = %v",
					want, err,
				)
			}

			got, err := r.GetByEmail(ctx, want.Email)
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
			tx := testutil.BeginTx(t, testDB)
			r := NewRepository(tx)

			ctx := context.Background()
			fakeEmail := user.Email("fake@example.com")

			got, err := r.GetByEmail(ctx, fakeEmail)
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
			tx := testutil.BeginTx(t, testDB)
			r := NewRepository(tx)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			email := user.Email("test@example.com")
			got, err := r.GetByEmail(ctx, email)

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
			tx := testutil.BeginTx(t, testDB)
			r := NewRepository(tx)

			ctx := context.Background()
			want := user.User{
				Username:       user.Username("Test"),
				Email:          user.Email("test@example.com"),
				HashedPassword: user.HashedPassword("testpassword"),
				Role:           user.RoleAdmin,
			}

			if err := r.Create(ctx, &want); err != nil {
				t.Fatalf(
					"Create(ctx, %v) error = %v",
					want, err,
				)
			}
			if want.CreatedAt.IsZero() || want.UpdatedAt.IsZero() {
				t.Fatalf(
					"Create(ctx, %v) sets CreatedAt = %v, UpdatedAt = %v, want non-zero",
					want, want.CreatedAt, want.UpdatedAt,
				)
			}

			got, err := r.GetByID(ctx, want.ID)
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
			if !got.CreatedAt.Equal(want.CreatedAt.Truncate(time.Microsecond)) ||
				!got.UpdatedAt.Equal(want.UpdatedAt.Truncate(time.Microsecond)) {
				t.Errorf(
					"GetByID(ctx, %d) (user.User, error) = %v, want CreatedAt %v, UpdatedAt %v",
					want.ID, got,
					want.CreatedAt.Truncate(time.Microsecond),
					want.UpdatedAt.Truncate(time.Microsecond),
				)
			}
		},
	)

	t.Run(
		"returns ErrAlreadyExists when the user already exists",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			r := NewRepository(tx)

			ctx := context.Background()

			u1 := user.User{
				Username:       user.Username("Test"),
				Email:          user.Email("test@example.com"),
				HashedPassword: user.HashedPassword("testpassword"),
				Role:           user.RoleAdmin,
			}
			if err := r.Create(ctx, &u1); err != nil {
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
			if err := r.Create(ctx, &u2); !errors.Is(err, exception.ErrAlreadyExists) {
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
			tx := testutil.BeginTx(t, testDB)
			r := NewRepository(tx)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			u := user.User{
				Username:       user.Username("Test"),
				Email:          user.Email("test@example.com"),
				HashedPassword: user.HashedPassword("testpassword"),
				Role:           user.RoleAdmin,
			}

			err := r.Create(ctx, &u)
			if !errors.Is(err, context.Canceled) {
				t.Fatalf(
					"Create(ctx, %v) error = %v, want %v",
					u, err, context.Canceled,
				)
			}
		},
	)
}

func TestUpdateUsername(t *testing.T) {
	t.Run(
		"updates only the username",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			r := NewRepository(tx)

			ctx := context.Background()

			u1 := user.User{
				Username:       user.Username("Test"),
				Email:          user.Email("test@example.com"),
				HashedPassword: user.HashedPassword("testpassword"),
				Role:           user.RoleAdmin,
			}
			if err := r.Create(ctx, &u1); err != nil {
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
			if err := r.UpdateUsername(ctx, &u2); err != nil {
				t.Fatalf(
					"UpdateUsername(ctx, %v) error = %v",
					u2, err,
				)
			}

			got, err := r.GetByID(ctx, u1.ID)
			if err != nil {
				t.Fatalf(
					"GetByID(ctx, %d) (user.User, error) = %v, %v",
					u1.ID, got, err,
				)
			}
			if got.Username != u2.Username {
				t.Errorf(
					"GetByID(ctx, %d) (user.User, error) = %v, want Username %v",
					u1.ID, got, u2.Username,
				)
			}
			if got.Email != u1.Email ||
				got.HashedPassword != u1.HashedPassword ||
				got.Role != u1.Role {
				t.Errorf(
					"GetByID(ctx, %d) (user.User, error) = %v, want Email, HashedPassword and Role of %v",
					u1.ID, got, u1,
				)
			}
			if !got.CreatedAt.Equal(u1.CreatedAt.Truncate(time.Microsecond)) {
				t.Errorf(
					"GetByID(ctx, %d) (user.User, error) = %v, want CreatedAt %v",
					u1.ID, got, u1.CreatedAt.Truncate(time.Microsecond),
				)
			}
			if !got.UpdatedAt.After(u1.UpdatedAt) {
				t.Errorf(
					"GetByID(ctx, %d) (user.User, error) = %v, want UpdatedAt after %v",
					u1.ID, got, u1.UpdatedAt,
				)
			}
		},
	)

	t.Run(
		"returns ErrNotFound when the user does not exist",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			r := NewRepository(tx)

			ctx := context.Background()

			u1 := user.User{
				ID:       user.ID(999999),
				Username: user.Username("Test"),
			}
			if err := r.UpdateUsername(ctx, &u1); !errors.Is(err, exception.ErrNotFound) {
				t.Fatalf(
					"UpdateUsername(ctx, %v) error = %v, want %v",
					u1, err, exception.ErrNotFound,
				)
			}
		},
	)

	t.Run(
		"returns a wrapped error when the context is canceled",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			r := NewRepository(tx)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			u := user.User{
				ID:       user.ID(1),
				Username: user.Username("Test"),
			}

			err := r.UpdateUsername(ctx, &u)
			if !errors.Is(err, context.Canceled) {
				t.Fatalf(
					"UpdateUsername(ctx, %v) error = %v, want %v",
					u, err, context.Canceled,
				)
			}
		},
	)
}

func TestUpdateEmail(t *testing.T) {
	t.Run(
		"updates only the email",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			r := NewRepository(tx)

			ctx := context.Background()

			u1 := user.User{
				Username:       user.Username("Test"),
				Email:          user.Email("test@example.com"),
				HashedPassword: user.HashedPassword("testpassword"),
				Role:           user.RoleAdmin,
			}
			if err := r.Create(ctx, &u1); err != nil {
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
			if err := r.UpdateEmail(ctx, &u2); err != nil {
				t.Fatalf(
					"UpdateEmail(ctx, %v) error = %v",
					u2, err,
				)
			}

			got, err := r.GetByID(ctx, u1.ID)
			if err != nil {
				t.Fatalf(
					"GetByID(ctx, %d) (user.User, error) = %v, %v",
					u1.ID, got, err,
				)
			}
			if got.Email != u2.Email {
				t.Errorf(
					"GetByID(ctx, %d) (user.User, error) = %v, want Email %v",
					u1.ID, got, u2.Email,
				)
			}
			if got.Username != u1.Username ||
				got.HashedPassword != u1.HashedPassword ||
				got.Role != u1.Role {
				t.Errorf(
					"GetByID(ctx, %d) (user.User, error) = %v, want Username, HashedPassword and Role of %v",
					u1.ID, got, u1,
				)
			}
			if !got.CreatedAt.Equal(u1.CreatedAt.Truncate(time.Microsecond)) {
				t.Errorf(
					"GetByID(ctx, %d) (user.User, error) = %v, want CreatedAt %v",
					u1.ID, got, u1.CreatedAt.Truncate(time.Microsecond),
				)
			}
			if !got.UpdatedAt.After(u1.UpdatedAt) {
				t.Errorf(
					"GetByID(ctx, %d) (user.User, error) = %v, want UpdatedAt after %v",
					u1.ID, got, u1.UpdatedAt,
				)
			}
		},
	)

	t.Run(
		"returns ErrNotFound when the user does not exist",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			r := NewRepository(tx)

			ctx := context.Background()

			u1 := user.User{
				ID:    user.ID(999999),
				Email: user.Email("test@example.com"),
			}
			if err := r.UpdateEmail(ctx, &u1); !errors.Is(err, exception.ErrNotFound) {
				t.Fatalf(
					"UpdateEmail(ctx, %v) error = %v, want %v",
					u1, err, exception.ErrNotFound,
				)
			}
		},
	)

	t.Run(
		"returns ErrAlreadyExists when the email is used by another user",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			r := NewRepository(tx)

			ctx := context.Background()

			u1 := user.User{
				Username:       user.Username("Test"),
				Email:          user.Email("test@example.com"),
				HashedPassword: user.HashedPassword("testpassword"),
				Role:           user.RoleAdmin,
			}
			if err := r.Create(ctx, &u1); err != nil {
				t.Fatalf(
					"Create(ctx, %v) error = %v",
					u1, err,
				)
			}

			u2 := user.User{
				Username:       user.Username("Test2"),
				Email:          user.Email("test2@example.com"),
				HashedPassword: user.HashedPassword("testpassword"),
				Role:           user.RoleAdmin,
			}
			if err := r.Create(ctx, &u2); err != nil {
				t.Fatalf(
					"Create(ctx, %v) error = %v",
					u2, err,
				)
			}

			u2.Email = u1.Email
			if err := r.UpdateEmail(ctx, &u2); !errors.Is(err, exception.ErrAlreadyExists) {
				t.Fatalf(
					"UpdateEmail(ctx, %v) error = %v, want %v",
					u2, err, exception.ErrAlreadyExists,
				)
			}
		},
	)

	t.Run(
		"returns a wrapped error when the context is canceled",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			r := NewRepository(tx)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			u := user.User{
				ID:    user.ID(1),
				Email: user.Email("test@example.com"),
			}

			err := r.UpdateEmail(ctx, &u)
			if !errors.Is(err, context.Canceled) {
				t.Fatalf(
					"UpdateEmail(ctx, %v) error = %v, want %v",
					u, err, context.Canceled,
				)
			}
		},
	)
}

func TestUpdatePassword(t *testing.T) {
	t.Run(
		"updates only the hashed password",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			r := NewRepository(tx)

			ctx := context.Background()

			u1 := user.User{
				Username:       user.Username("Test"),
				Email:          user.Email("test@example.com"),
				HashedPassword: user.HashedPassword("testpassword"),
				Role:           user.RoleAdmin,
			}
			if err := r.Create(ctx, &u1); err != nil {
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
			if err := r.UpdatePassword(ctx, &u2); err != nil {
				t.Fatalf(
					"UpdatePassword(ctx, %v) error = %v",
					u2, err,
				)
			}

			got, err := r.GetByID(ctx, u1.ID)
			if err != nil {
				t.Fatalf(
					"GetByID(ctx, %d) (user.User, error) = %v, %v",
					u1.ID, got, err,
				)
			}
			if got.HashedPassword != u2.HashedPassword {
				t.Errorf(
					"GetByID(ctx, %d) (user.User, error) = %v, want HashedPassword %v",
					u1.ID, got, u2.HashedPassword,
				)
			}
			if got.Username != u1.Username ||
				got.Email != u1.Email ||
				got.Role != u1.Role {
				t.Errorf(
					"GetByID(ctx, %d) (user.User, error) = %v, want Username, Email and Role of %v",
					u1.ID, got, u1,
				)
			}
			if !got.CreatedAt.Equal(u1.CreatedAt.Truncate(time.Microsecond)) {
				t.Errorf(
					"GetByID(ctx, %d) (user.User, error) = %v, want CreatedAt %v",
					u1.ID, got, u1.CreatedAt.Truncate(time.Microsecond),
				)
			}
			if !got.UpdatedAt.After(u1.UpdatedAt) {
				t.Errorf(
					"GetByID(ctx, %d) (user.User, error) = %v, want UpdatedAt after %v",
					u1.ID, got, u1.UpdatedAt,
				)
			}
		},
	)

	t.Run(
		"returns ErrNotFound when the user does not exist",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			r := NewRepository(tx)

			ctx := context.Background()

			u1 := user.User{
				ID:             user.ID(999999),
				HashedPassword: user.HashedPassword("testpassword"),
			}
			if err := r.UpdatePassword(ctx, &u1); !errors.Is(err, exception.ErrNotFound) {
				t.Fatalf(
					"UpdatePassword(ctx, %v) error = %v, want %v",
					u1, err, exception.ErrNotFound,
				)
			}
		},
	)

	t.Run(
		"returns a wrapped error when the context is canceled",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			r := NewRepository(tx)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			u := user.User{
				ID:             user.ID(1),
				HashedPassword: user.HashedPassword("testpassword"),
			}

			err := r.UpdatePassword(ctx, &u)
			if !errors.Is(err, context.Canceled) {
				t.Fatalf(
					"UpdatePassword(ctx, %v) error = %v, want %v",
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
			tx := testutil.BeginTx(t, testDB)
			r := NewRepository(tx)

			ctx := context.Background()

			u := user.User{
				Username:       user.Username("Test"),
				Email:          user.Email("test@example.com"),
				HashedPassword: user.HashedPassword("testpassword"),
				Role:           user.RoleAdmin,
			}
			if err := r.Create(ctx, &u); err != nil {
				t.Fatalf(
					"Create(ctx, %v) error = %v",
					u, err,
				)
			}

			if err := r.Delete(ctx, u.ID); err != nil {
				t.Fatalf(
					"Delete(ctx, %d) error = %v",
					u.ID, err,
				)
			}

			got, err := r.GetByID(ctx, u.ID)
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
			tx := testutil.BeginTx(t, testDB)
			r := NewRepository(tx)

			ctx := context.Background()

			id := user.ID(999999)
			if err := r.Delete(ctx, id); !errors.Is(err, exception.ErrNotFound) {
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
			tx := testutil.BeginTx(t, testDB)
			r := NewRepository(tx)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			id := user.ID(1)
			err := r.Delete(ctx, id)
			if !errors.Is(err, context.Canceled) {
				t.Fatalf(
					"Delete(ctx, %d) error = %v, want %v",
					id, err, context.Canceled,
				)
			}
		},
	)
}
