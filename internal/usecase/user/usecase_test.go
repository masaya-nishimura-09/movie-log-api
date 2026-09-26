package user

import (
	"context"
	"errors"
	"testing"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/auth"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/media"
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

type fakeMediaService struct {
	deletedAllUserID user.ID
	err              error
}

func (s *fakeMediaService) Upload(
	ctx context.Context,
	userID user.ID,
	m media.Media,
) (media.URL, error) {
	return "", s.err
}

func (s *fakeMediaService) Delete(
	ctx context.Context,
	userID user.ID,
	url media.URL,
) error {
	return s.err
}

func (s *fakeMediaService) DeleteAllForUser(
	ctx context.Context,
	userID user.ID,
) error {
	s.deletedAllUserID = userID
	return s.err
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

func TestCreate(t *testing.T) {
	t.Run(
		"hashes the password and sets the role to user",
		func(t *testing.T) {
			repo := &fakeRepository{}
			uu := NewUserUsecase(repo, &fakeRefreshTokenRepo{}, &fakeMediaService{})

			ctx := context.Background()
			username := user.Username("Test")
			email := user.Email("test@example.com")
			password := user.Password("testpassword")

			got, err := uu.Create(ctx, username, email, password)
			if err != nil {
				t.Fatalf(
					"Create(ctx, %v, %v, %v) (*user.User, error) = %v, %v",
					username, email, password, got, err,
				)
			}

			if err := bcrypt.CompareHashAndPassword(
				[]byte(got.HashedPassword),
				[]byte(password),
			); err != nil {
				t.Errorf(
					"Create(ctx, %v, %v, %v) HashedPassword = %v, want a bcrypt hash of %v",
					username, email, password, got.HashedPassword, password,
				)
			}

			if got.Role != user.RoleUser {
				t.Errorf(
					"Create(ctx, %v, %v, %v) Role = %v, want %v",
					username, email, password, got.Role, user.RoleUser,
				)
			}
		},
	)
}

func TestUpdate(t *testing.T) {
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
			uu := NewUserUsecase(repo, &fakeRefreshTokenRepo{}, &fakeMediaService{})

			ctx := context.Background()

			got, err := uu.Update(ctx, userID, username, email, password)
			if err != nil {
				t.Fatalf(
					"Update(ctx, %v, %v, %v, %v) (*user.User, error) = %v, %v",
					userID, username, email, password, got, err,
				)
			}

			if err := bcrypt.CompareHashAndPassword(
				[]byte(got.HashedPassword),
				[]byte(password),
			); err != nil {
				t.Errorf(
					"Update(ctx, %v, %v, %v, %v) HashedPassword = %v, want a bcrypt hash of %v",
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
			uu := NewUserUsecase(repo, refreshTokenRepo, &fakeMediaService{})

			ctx := context.Background()

			got, err := uu.Update(ctx, userID, username, email, password)
			if err != nil {
				t.Fatalf(
					"Update(ctx, %v, %v, %v, %v) (*user.User, error) = %v, %v",
					userID, username, email, password, got, err,
				)
			}

			if refreshTokenRepo.revokedUserID != userID {
				t.Errorf(
					"Update(ctx, %v, %v, %v, %v) revoked user id = %v, want %v",
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
			uu := NewUserUsecase(repo, refreshTokenRepo, &fakeMediaService{})

			ctx := context.Background()

			got, err := uu.Update(ctx, userID, username, email, password)
			if err != nil {
				t.Fatalf(
					"Update(ctx, %v, %v, %v, %v) (*user.User, error) = %v, %v",
					userID, username, email, password, got, err,
				)
			}

			if refreshTokenRepo.revokedUserID != 0 {
				t.Errorf(
					"Update(ctx, %v, %v, %v, %v) revoked user id = %v, want no revocation",
					userID, username, email, password,
					refreshTokenRepo.revokedUserID,
				)
			}
		},
	)
}

func TestDelete(t *testing.T) {
	userID := user.ID(1)

	t.Run(
		"deletes all media of the user",
		func(t *testing.T) {
			mediaService := &fakeMediaService{}
			uu := NewUserUsecase(&fakeRepository{}, &fakeRefreshTokenRepo{}, mediaService)

			ctx := context.Background()

			if err := uu.Delete(ctx, userID); err != nil {
				t.Fatalf(
					"Delete(ctx, %v) error = %v",
					userID, err,
				)
			}
			if mediaService.deletedAllUserID != userID {
				t.Errorf(
					"Delete(ctx, %v) deletes media of user %v, want %v",
					userID, mediaService.deletedAllUserID, userID,
				)
			}
		},
	)

	t.Run(
		"succeeds even when deleting the media fails",
		func(t *testing.T) {
			mediaService := &fakeMediaService{
				err: errors.New("delete media"),
			}
			uu := NewUserUsecase(&fakeRepository{}, &fakeRefreshTokenRepo{}, mediaService)

			ctx := context.Background()

			if err := uu.Delete(ctx, userID); err != nil {
				t.Fatalf(
					"Delete(ctx, %v) error = %v, want nil",
					userID, err,
				)
			}
		},
	)
}
