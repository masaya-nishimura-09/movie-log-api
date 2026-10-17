package user

import (
	"context"
	"errors"
	"testing"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/auth"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/media"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
	"golang.org/x/crypto/bcrypt"
)

type fakeRepository struct {
	user    *user.User
	getErr  error
	updated *user.User
}

func (r *fakeRepository) GetByID(
	ctx context.Context,
	userID user.ID,
) (*user.User, error) {
	return r.user, r.getErr
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

func (r *fakeRepository) UpdateUsername(ctx context.Context, u *user.User) error {
	r.updated = u
	return nil
}

func (r *fakeRepository) UpdateEmail(ctx context.Context, u *user.User) error {
	r.updated = u
	return nil
}

func (r *fakeRepository) UpdatePassword(ctx context.Context, u *user.User) error {
	r.updated = u
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

func (r *fakeRefreshTokenRepo) DeleteExpired(
	ctx context.Context,
) error {
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

func TestUpdateUsername(t *testing.T) {
	userID := user.ID(1)
	newUsername := user.Username("NewName")

	t.Run(
		"returns the user with the new username and keeps the other fields",
		func(t *testing.T) {
			existing := &user.User{
				ID:             userID,
				Username:       user.Username("Test"),
				Email:          user.Email("test@example.com"),
				HashedPassword: hashPassword(t, user.Password("testpassword")),
				Role:           user.RoleAdmin,
			}
			repo := &fakeRepository{user: existing}
			uu := NewUserUsecase(repo, &fakeRefreshTokenRepo{}, &fakeMediaService{})

			ctx := context.Background()

			got, err := uu.UpdateUsername(ctx, userID, newUsername)
			if err != nil {
				t.Fatalf(
					"UpdateUsername(ctx, %v, %v) (*user.User, error) = %v, %v",
					userID, newUsername, got, err,
				)
			}

			if got.ID != userID ||
				got.Username != newUsername ||
				got.Email != existing.Email ||
				got.HashedPassword != existing.HashedPassword ||
				got.Role != existing.Role {
				t.Errorf(
					"UpdateUsername(ctx, %v, %v) (*user.User, error) = %v, want Username %v and the other fields of %v",
					userID, newUsername, got, newUsername, existing,
				)
			}
		},
	)

	t.Run(
		"returns ErrNotFound when the user does not exist",
		func(t *testing.T) {
			repo := &fakeRepository{getErr: exception.ErrNotFound}
			uu := NewUserUsecase(repo, &fakeRefreshTokenRepo{}, &fakeMediaService{})

			ctx := context.Background()

			got, err := uu.UpdateUsername(ctx, userID, newUsername)
			if !errors.Is(err, exception.ErrNotFound) {
				t.Fatalf(
					"UpdateUsername(ctx, %v, %v) (*user.User, error) = %v, %v, want %v",
					userID, newUsername, got, err, exception.ErrNotFound,
				)
			}
			if got != nil || repo.updated != nil {
				t.Errorf(
					"UpdateUsername(ctx, %v, %v) returned %v and updated %v, want nil for both",
					userID, newUsername, got, repo.updated,
				)
			}
		},
	)
}

func TestUpdateEmail(t *testing.T) {
	userID := user.ID(1)
	password := user.Password("testpassword")
	newEmail := user.Email("new@example.com")

	t.Run(
		"returns the user with the new email and keeps the other fields",
		func(t *testing.T) {
			existing := &user.User{
				ID:             userID,
				Username:       user.Username("Test"),
				Email:          user.Email("test@example.com"),
				HashedPassword: hashPassword(t, password),
				Role:           user.RoleAdmin,
			}
			repo := &fakeRepository{user: existing}
			uu := NewUserUsecase(repo, &fakeRefreshTokenRepo{}, &fakeMediaService{})

			ctx := context.Background()

			got, err := uu.UpdateEmail(ctx, userID, newEmail, password)
			if err != nil {
				t.Fatalf(
					"UpdateEmail(ctx, %v, %v, %v) (*user.User, error) = %v, %v",
					userID, newEmail, password, got, err,
				)
			}

			if got.ID != userID ||
				got.Username != existing.Username ||
				got.Email != newEmail ||
				got.HashedPassword != existing.HashedPassword ||
				got.Role != existing.Role {
				t.Errorf(
					"UpdateEmail(ctx, %v, %v, %v) (*user.User, error) = %v, want Email %v and the other fields of %v",
					userID, newEmail, password, got, newEmail, existing,
				)
			}
		},
	)

	t.Run(
		"returns ErrInvalid and does not update when the password is incorrect",
		func(t *testing.T) {
			repo := &fakeRepository{
				user: &user.User{
					ID:             userID,
					HashedPassword: hashPassword(t, password),
				},
			}
			uu := NewUserUsecase(repo, &fakeRefreshTokenRepo{}, &fakeMediaService{})

			ctx := context.Background()
			wrongPassword := user.Password("wrongpassword")

			got, err := uu.UpdateEmail(ctx, userID, newEmail, wrongPassword)
			if !errors.Is(err, exception.ErrInvalid) {
				t.Fatalf(
					"UpdateEmail(ctx, %v, %v, %v) (*user.User, error) = %v, %v, want %v",
					userID, newEmail, wrongPassword, got, err, exception.ErrInvalid,
				)
			}
			if got != nil || repo.updated != nil {
				t.Errorf(
					"UpdateEmail(ctx, %v, %v, %v) returned %v and updated %v, want nil for both",
					userID, newEmail, wrongPassword, got, repo.updated,
				)
			}
		},
	)

	t.Run(
		"returns ErrNotFound when the user does not exist",
		func(t *testing.T) {
			repo := &fakeRepository{getErr: exception.ErrNotFound}
			uu := NewUserUsecase(repo, &fakeRefreshTokenRepo{}, &fakeMediaService{})

			ctx := context.Background()

			got, err := uu.UpdateEmail(ctx, userID, newEmail, password)
			if !errors.Is(err, exception.ErrNotFound) {
				t.Fatalf(
					"UpdateEmail(ctx, %v, %v, %v) (*user.User, error) = %v, %v, want %v",
					userID, newEmail, password, got, err, exception.ErrNotFound,
				)
			}
			if got != nil || repo.updated != nil {
				t.Errorf(
					"UpdateEmail(ctx, %v, %v, %v) returned %v and updated %v, want nil for both",
					userID, newEmail, password, got, repo.updated,
				)
			}
		},
	)
}

func TestUpdatePassword(t *testing.T) {
	userID := user.ID(1)
	oldPassword := user.Password("oldpassword")
	newPassword := user.Password("newpassword")

	t.Run(
		"hashes the new password and keeps the other fields",
		func(t *testing.T) {
			existing := &user.User{
				ID:             userID,
				Username:       user.Username("Test"),
				Email:          user.Email("test@example.com"),
				HashedPassword: hashPassword(t, oldPassword),
				Role:           user.RoleAdmin,
			}
			repo := &fakeRepository{user: existing}
			uu := NewUserUsecase(repo, &fakeRefreshTokenRepo{}, &fakeMediaService{})

			ctx := context.Background()

			got, err := uu.UpdatePassword(ctx, userID, oldPassword, newPassword)
			if err != nil {
				t.Fatalf(
					"UpdatePassword(ctx, %v, %v, %v) (*user.User, error) = %v, %v",
					userID, oldPassword, newPassword, got, err,
				)
			}

			if err := bcrypt.CompareHashAndPassword(
				[]byte(got.HashedPassword),
				[]byte(newPassword),
			); err != nil {
				t.Errorf(
					"UpdatePassword(ctx, %v, %v, %v) HashedPassword = %v, want a bcrypt hash of %v",
					userID, oldPassword, newPassword, got.HashedPassword, newPassword,
				)
			}
			if got.ID != userID ||
				got.Username != existing.Username ||
				got.Email != existing.Email ||
				got.Role != existing.Role {
				t.Errorf(
					"UpdatePassword(ctx, %v, %v, %v) (*user.User, error) = %v, want the other fields of %v",
					userID, oldPassword, newPassword, got, existing,
				)
			}
		},
	)

	t.Run(
		"revokes all refresh tokens of the user",
		func(t *testing.T) {
			repo := &fakeRepository{
				user: &user.User{
					ID:             userID,
					HashedPassword: hashPassword(t, oldPassword),
				},
			}
			refreshTokenRepo := &fakeRefreshTokenRepo{}
			uu := NewUserUsecase(repo, refreshTokenRepo, &fakeMediaService{})

			ctx := context.Background()

			got, err := uu.UpdatePassword(ctx, userID, oldPassword, newPassword)
			if err != nil {
				t.Fatalf(
					"UpdatePassword(ctx, %v, %v, %v) (*user.User, error) = %v, %v",
					userID, oldPassword, newPassword, got, err,
				)
			}

			if refreshTokenRepo.revokedUserID != userID {
				t.Errorf(
					"UpdatePassword(ctx, %v, %v, %v) revoked user id = %v, want %v",
					userID, oldPassword, newPassword,
					refreshTokenRepo.revokedUserID, userID,
				)
			}
		},
	)

	t.Run(
		"returns ErrInvalid and neither updates nor revokes when the old password is incorrect",
		func(t *testing.T) {
			repo := &fakeRepository{
				user: &user.User{
					ID:             userID,
					HashedPassword: hashPassword(t, oldPassword),
				},
			}
			refreshTokenRepo := &fakeRefreshTokenRepo{}
			uu := NewUserUsecase(repo, refreshTokenRepo, &fakeMediaService{})

			ctx := context.Background()
			wrongPassword := user.Password("wrongpassword")

			got, err := uu.UpdatePassword(ctx, userID, wrongPassword, newPassword)
			if !errors.Is(err, exception.ErrInvalid) {
				t.Fatalf(
					"UpdatePassword(ctx, %v, %v, %v) (*user.User, error) = %v, %v, want %v",
					userID, wrongPassword, newPassword, got, err, exception.ErrInvalid,
				)
			}
			if got != nil ||
				repo.updated != nil ||
				refreshTokenRepo.revokedUserID != 0 {
				t.Errorf(
					"UpdatePassword(ctx, %v, %v, %v) returned %v, updated %v and revoked user id %v, want no result, update, or revocation",
					userID, wrongPassword, newPassword,
					got, repo.updated, refreshTokenRepo.revokedUserID,
				)
			}
		},
	)

	t.Run(
		"returns ErrNotFound when the user does not exist",
		func(t *testing.T) {
			repo := &fakeRepository{getErr: exception.ErrNotFound}
			refreshTokenRepo := &fakeRefreshTokenRepo{}
			uu := NewUserUsecase(repo, refreshTokenRepo, &fakeMediaService{})

			ctx := context.Background()

			got, err := uu.UpdatePassword(ctx, userID, oldPassword, newPassword)
			if !errors.Is(err, exception.ErrNotFound) {
				t.Fatalf(
					"UpdatePassword(ctx, %v, %v, %v) (*user.User, error) = %v, %v, want %v",
					userID, oldPassword, newPassword, got, err, exception.ErrNotFound,
				)
			}
			if got != nil ||
				repo.updated != nil ||
				refreshTokenRepo.revokedUserID != 0 {
				t.Errorf(
					"UpdatePassword(ctx, %v, %v, %v) returned %v, updated %v and revoked user id %v, want no result, update, or revocation",
					userID, oldPassword, newPassword,
					got, repo.updated, refreshTokenRepo.revokedUserID,
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
