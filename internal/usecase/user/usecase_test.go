package user

import (
	"context"
	"testing"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/auth"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
	"golang.org/x/crypto/bcrypt"
)

type fakeRepository struct {
	user *user.User
}

func (r *fakeRepository) GetByID(
	ctx context.Context,
	userID user.ID,
) (*user.User, error) {
	return r.user, nil
}

func (r *fakeRepository) GetByEmail(
	ctx context.Context,
	email user.Email,
) (*user.User, error) {
	return r.user, nil
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

type fakeRefreshTokenRepo struct {
	revokedUserID user.ID
}

func (r *fakeRefreshTokenRepo) Create(
	ctx context.Context,
	principal *auth.Principal,
) (*auth.RefreshToken, error) {
	return nil, nil
}

func (r *fakeRefreshTokenRepo) FindValidByValue(
	ctx context.Context,
	value auth.RefreshTokenValue,
) (*auth.RefreshToken, error) {
	return nil, nil
}

func (r *fakeRefreshTokenRepo) Revoke(
	ctx context.Context,
	id auth.RefreshTokenID,
) error {
	return nil
}

func (r *fakeRefreshTokenRepo) RevokeAllForUser(
	ctx context.Context,
	userID user.ID,
) error {
	r.revokedUserID = userID
	return nil
}

func hashPassword(t *testing.T, password user.Password) user.HashedPassword {
	t.Helper()
	hashed, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		t.Fatalf(
			"GenerateFromPassword(%v, %v) = %v",
			password, bcrypt.DefaultCost, err,
		)
	}
	return user.HashedPassword(hashed)
}

func TestRegister(t *testing.T) {
	t.Run(
		"hashes the password and sets the role to user",
		func(t *testing.T) {
			repo := &fakeRepository{}
			uu := NewUserUsecase(repo, &fakeRefreshTokenRepo{})

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
	userID := user.ID(1)
	username := user.Username("Test")
	email := user.Email("test@example.com")

	t.Run(
		"hashes the password",
		func(t *testing.T) {
			password := user.Password("testpassword")
			repo := &fakeRepository{
				user: &user.User{
					ID:             userID,
					HashedPassword: hashPassword(t, password),
				},
			}
			uu := NewUserUsecase(repo, &fakeRefreshTokenRepo{})

			ctx := context.Background()

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

	t.Run(
		"revokes all refresh tokens when the password changes",
		func(t *testing.T) {
			password := user.Password("newpassword")
			repo := &fakeRepository{
				user: &user.User{
					ID:             userID,
					HashedPassword: hashPassword(t, user.Password("oldpassword")),
				},
			}
			refreshTokenRepo := &fakeRefreshTokenRepo{}
			uu := NewUserUsecase(repo, refreshTokenRepo)

			ctx := context.Background()

			got, err := uu.UpdateUser(ctx, userID, username, email, password)
			if err != nil {
				t.Fatalf(
					"UpdateUser(ctx, %v, %v, %v, %v) (*user.User, error) = %v, %v",
					userID, username, email, password, got, err,
				)
			}

			if refreshTokenRepo.revokedUserID != userID {
				t.Errorf(
					"UpdateUser(ctx, %v, %v, %v, %v) revoked user id = %v, want %v",
					userID, username, email, password,
					refreshTokenRepo.revokedUserID, userID,
				)
			}
		},
	)

	t.Run(
		"keeps refresh tokens when the password is unchanged",
		func(t *testing.T) {
			password := user.Password("testpassword")
			repo := &fakeRepository{
				user: &user.User{
					ID:             userID,
					HashedPassword: hashPassword(t, password),
				},
			}
			refreshTokenRepo := &fakeRefreshTokenRepo{}
			uu := NewUserUsecase(repo, refreshTokenRepo)

			ctx := context.Background()

			got, err := uu.UpdateUser(ctx, userID, username, email, password)
			if err != nil {
				t.Fatalf(
					"UpdateUser(ctx, %v, %v, %v, %v) (*user.User, error) = %v, %v",
					userID, username, email, password, got, err,
				)
			}

			if refreshTokenRepo.revokedUserID != 0 {
				t.Errorf(
					"UpdateUser(ctx, %v, %v, %v, %v) revoked user id = %v, want no revocation",
					userID, username, email, password,
					refreshTokenRepo.revokedUserID,
				)
			}
		},
	)
}
