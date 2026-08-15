package user

import (
	"context"
	"testing"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
	"golang.org/x/crypto/bcrypt"
)

type fakeRepository struct{}

func (r *fakeRepository) GetByID(
	ctx context.Context,
	userID user.ID,
) (*user.User, error) {
	return nil, nil
}

func (r *fakeRepository) GetByEmail(
	ctx context.Context,
	email user.Email,
) (*user.User, error) {
	return nil, nil
}

func (r *fakeRepository) Create(ctx context.Context, u *user.User) error {
	return nil
}

func (r *fakeRepository) Update(ctx context.Context, u *user.User) error {
	return nil
}

func (r *fakeRepository) Delete(ctx context.Context, userID user.ID) error {
	return nil
}

func TestRegister(t *testing.T) {
	t.Run(
		"hashes the password and sets the role to user",
		func(t *testing.T) {
			repo := &fakeRepository{}
			uu := NewUserUsecase(repo)

			ctx := context.Background()
			username := user.Username("Test")
			email := user.Email("test@example.com")
			password := user.Password("testpassword")

			got, err := uu.Register(ctx, username, email, password)
			if err != nil {
				t.Fatalf(
					"Register(ctx, %v, %v, %v) (*user.User, error) = %v, %v",
					username, email, password, got, err,
				)
			}

			if err := bcrypt.CompareHashAndPassword(
				[]byte(got.HashedPassword),
				[]byte(password),
			); err != nil {
				t.Errorf(
					"Register(ctx, %v, %v, %v) HashedPassword = %v, want a bcrypt hash of %v",
					username, email, password, got.HashedPassword, password,
				)
			}

			if got.Role != user.RoleUser {
				t.Errorf(
					"Register(ctx, %v, %v, %v) Role = %v, want %v",
					username, email, password, got.Role, user.RoleUser,
				)
			}
		},
	)
}

func TestUpdateUser(t *testing.T) {
	t.Run(
		"hashes the password",
		func(t *testing.T) {
			repo := &fakeRepository{}
			uu := NewUserUsecase(repo)

			ctx := context.Background()
			userID := user.ID(1)
			username := user.Username("Test")
			email := user.Email("test@example.com")
			password := user.Password("testpassword")

			got, err := uu.UpdateUser(ctx, userID, username, email, password)
			if err != nil {
				t.Fatalf(
					"UpdateUser(ctx, %v, %v, %v, %v) (*user.User, error) = %v, %v",
					userID, username, email, password, got, err,
				)
			}

			if err := bcrypt.CompareHashAndPassword(
				[]byte(got.HashedPassword),
				[]byte(password),
			); err != nil {
				t.Errorf(
					"UpdateUser(ctx, %v, %v, %v, %v) HashedPassword = %v, want a bcrypt hash of %v",
					userID, username, email, password, got.HashedPassword, password,
				)
			}
		},
	)
}
