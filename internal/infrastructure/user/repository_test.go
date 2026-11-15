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

func TestReplaceUnverified(t *testing.T) {
	t.Run(
		"replaces the username and hashed password of the unverified user",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			r := NewRepository(tx)

			ctx := context.Background()

			u1 := user.User{
				Username:       user.Username("Test"),
				Email:          user.Email("test@example.com"),
				HashedPassword: user.HashedPassword("testpassword"),
				Role:           user.RoleUser,
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
				Role:           user.RoleAdmin,
			}
			if err := r.ReplaceUnverified(ctx, &u2); err != nil {
				t.Fatalf(
					"ReplaceUnverified(ctx, %v) error = %v",
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
			if got.Username != u2.Username ||
				got.HashedPassword != u2.HashedPassword {
				t.Errorf(
					"GetByID(ctx, %d) (user.User, error) = %v, want Username and HashedPassword of %v",
					u1.ID, got, u2,
				)
			}
			if got.Email != u1.Email ||
				got.Role != u1.Role ||
				got.IsEmailVerified() {
				t.Errorf(
					"GetByID(ctx, %d) (user.User, error) = %v, want Email and Role of %v and not verified",
					u1.ID, got, u1,
				)
			}
			if !got.CreatedAt.After(u1.CreatedAt) {
				t.Errorf(
					"GetByID(ctx, %d) (user.User, error) = %v, want CreatedAt after %v",
					u1.ID, got, u1.CreatedAt,
				)
			}
		},
	)

	t.Run(
		"returns ErrNotFound when the user is already verified",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			r := NewRepository(tx)

			ctx := context.Background()

			verifiedAt := time.Now()
			u1 := user.User{
				Username:        user.Username("Test"),
				Email:           user.Email("test@example.com"),
				HashedPassword:  user.HashedPassword("testpassword"),
				Role:            user.RoleUser,
				EmailVerifiedAt: &verifiedAt,
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
				HashedPassword: user.HashedPassword("test2password"),
			}
			if err := r.ReplaceUnverified(ctx, &u2); !errors.Is(err, exception.ErrNotFound) {
				t.Fatalf(
					"ReplaceUnverified(ctx, %v) error = %v, want %v",
					u2, err, exception.ErrNotFound,
				)
			}

			got, err := r.GetByID(ctx, u1.ID)
			if err != nil {
				t.Fatalf(
					"GetByID(ctx, %d) (user.User, error) = %v, %v",
					u1.ID, got, err,
				)
			}
			if got.Username != u1.Username || got.HashedPassword != u1.HashedPassword {
				t.Errorf(
					"GetByID(ctx, %d) (user.User, error) = %v, want the user unchanged",
					u1.ID, got,
				)
			}
		},
	)
}

func TestUpdateEmailVerifiedAt(t *testing.T) {
	t.Run(
		"updates only the email verified at",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			r := NewRepository(tx)

			ctx := context.Background()

			u1 := user.User{
				Username:       user.Username("Test"),
				Email:          user.Email("test@example.com"),
				HashedPassword: user.HashedPassword("testpassword"),
				Role:           user.RoleUser,
			}
			if err := r.Create(ctx, &u1); err != nil {
				t.Fatalf(
					"Create(ctx, %v) error = %v",
					u1, err,
				)
			}

			verifiedAt := time.Now()
			u2 := user.User{
				ID:              u1.ID,
				Username:        user.Username("Test2"),
				EmailVerifiedAt: &verifiedAt,
			}
			if err := r.UpdateEmailVerifiedAt(ctx, &u2); err != nil {
				t.Fatalf(
					"UpdateEmailVerifiedAt(ctx, %v) error = %v",
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
			if got.EmailVerifiedAt == nil ||
				!got.EmailVerifiedAt.Equal(verifiedAt.Truncate(time.Microsecond)) {
				t.Errorf(
					"GetByID(ctx, %d) (user.User, error) = %v, want EmailVerifiedAt %v",
					u1.ID, got, verifiedAt.Truncate(time.Microsecond),
				)
			}
			if got.Username != u1.Username {
				t.Errorf(
					"GetByID(ctx, %d) (user.User, error) = %v, want Username %v",
					u1.ID, got, u1.Username,
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

			verifiedAt := time.Now()
			u := user.User{
				ID:              user.ID(999999),
				EmailVerifiedAt: &verifiedAt,
			}
			if err := r.UpdateEmailVerifiedAt(ctx, &u); !errors.Is(err, exception.ErrNotFound) {
				t.Fatalf(
					"UpdateEmailVerifiedAt(ctx, %v) error = %v, want %v",
					u, err, exception.ErrNotFound,
				)
			}
		},
	)
}

func TestDeleteUnverifiedCreatedBefore(t *testing.T) {
	t.Run(
		"deletes only the unverified users created before the given time",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			r := NewRepository(tx)

			ctx := context.Background()

			verifiedAt := time.Now()
			oldUnverified := user.User{
				Username:       user.Username("Old"),
				Email:          user.Email("old@example.com"),
				HashedPassword: user.HashedPassword("testpassword"),
				Role:           user.RoleUser,
			}
			oldVerified := user.User{
				Username:        user.Username("Verified"),
				Email:           user.Email("verified@example.com"),
				HashedPassword:  user.HashedPassword("testpassword"),
				Role:            user.RoleUser,
				EmailVerifiedAt: &verifiedAt,
			}
			if err := r.Create(ctx, &oldUnverified); err != nil {
				t.Fatalf(
					"Create(ctx, %v) error = %v",
					oldUnverified, err,
				)
			}
			if err := r.Create(ctx, &oldVerified); err != nil {
				t.Fatalf(
					"Create(ctx, %v) error = %v",
					oldVerified, err,
				)
			}

			before := time.Now()

			newUnverified := user.User{
				Username:       user.Username("New"),
				Email:          user.Email("new@example.com"),
				HashedPassword: user.HashedPassword("testpassword"),
				Role:           user.RoleUser,
			}
			if err := r.Create(ctx, &newUnverified); err != nil {
				t.Fatalf(
					"Create(ctx, %v) error = %v",
					newUnverified, err,
				)
			}

			if err := r.DeleteUnverifiedCreatedBefore(ctx, before); err != nil {
				t.Fatalf(
					"DeleteUnverifiedCreatedBefore(ctx, %v) error = %v",
					before, err,
				)
			}

			if got, err := r.GetByID(ctx, oldUnverified.ID); !errors.Is(err, exception.ErrNotFound) {
				t.Errorf(
					"GetByID(ctx, %d) (user.User, error) = %v, %v, want %v",
					oldUnverified.ID, got, err, exception.ErrNotFound,
				)
			}
			if got, err := r.GetByID(ctx, oldVerified.ID); err != nil {
				t.Errorf(
					"GetByID(ctx, %d) (user.User, error) = %v, %v, want the verified user to remain",
					oldVerified.ID, got, err,
				)
			}
			if got, err := r.GetByID(ctx, newUnverified.ID); err != nil {
				t.Errorf(
					"GetByID(ctx, %d) (user.User, error) = %v, %v, want the new unverified user to remain",
					newUnverified.ID, got, err,
				)
			}
		},
	)
}
