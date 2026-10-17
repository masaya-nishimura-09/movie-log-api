package user

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/auth"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/media"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
	"golang.org/x/crypto/bcrypt"
)

type Usecase interface {
	GetByID(ctx context.Context, userID user.ID) (*user.User, error)
	Create(
		ctx context.Context,
		username user.Username,
		email user.Email,
		password user.Password,
	) (*user.User, error)
	UpdateUsername(
		ctx context.Context,
		userID user.ID,
		username user.Username,
	) (*user.User, error)
	UpdateEmail(
		ctx context.Context,
		userID user.ID,
		email user.Email,
		password user.Password,
	) (*user.User, error)
	UpdatePassword(
		ctx context.Context,
		userID user.ID,
		oldPassword user.Password,
		newPassword user.Password,
	) (*user.User, error)
	Delete(ctx context.Context, userID user.ID) error
}

type UserUsecase struct {
	userRepo         user.Repository
	refreshTokenRepo auth.RefreshTokenRepository
	mediaService     media.Service
}

func NewUserUsecase(
	userRepo user.Repository,
	refreshTokenRepo auth.RefreshTokenRepository,
	mediaService media.Service,
) *UserUsecase {
	return &UserUsecase{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		mediaService:     mediaService,
	}
}

func (uu *UserUsecase) GetByID(
	ctx context.Context,
	userID user.ID,
) (*user.User, error) {
	u, err := uu.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return u, nil
}

func (uu *UserUsecase) Create(
	ctx context.Context,
	username user.Username,
	email user.Email,
	password user.Password,
) (*user.User, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	u := user.NewUser(
		username,
		email,
		user.HashedPassword(hashed),
		user.RoleUser,
	)

	if err := uu.userRepo.Create(ctx, &u); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return &u, nil
}

func (uu *UserUsecase) UpdateUsername(
	ctx context.Context,
	userID user.ID,
	newUsername user.Username,
) (*user.User, error) {
	existingUser, err := uu.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	u := user.NewUser(
		newUsername,
		existingUser.Email,
		existingUser.HashedPassword,
		existingUser.Role,
	)
	u.ID = userID

	if err := uu.userRepo.UpdateUsername(ctx, &u); err != nil {
		return nil, fmt.Errorf("update username: %w", err)
	}

	return &u, nil
}

func (uu *UserUsecase) UpdateEmail(
	ctx context.Context,
	userID user.ID,
	newEmail user.Email,
	password user.Password,
) (*user.User, error) {
	existingUser, err := uu.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(existingUser.HashedPassword),
		[]byte(password),
	); err != nil {
		return nil, fmt.Errorf("login: %w", exception.ErrInvalid)
	}

	u := user.NewUser(
		existingUser.Username,
		newEmail,
		existingUser.HashedPassword,
		existingUser.Role,
	)
	u.ID = userID

	if err := uu.userRepo.UpdateEmail(ctx, &u); err != nil {
		return nil, fmt.Errorf("update email: %w", err)
	}

	return &u, nil
}

func (uu *UserUsecase) UpdatePassword(
	ctx context.Context,
	userID user.ID,
	oldPassword user.Password,
	newPassword user.Password,
) (*user.User, error) {
	existingUser, err := uu.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(existingUser.HashedPassword),
		[]byte(oldPassword),
	); err != nil {
		return nil, fmt.Errorf("login: %w", exception.ErrInvalid)
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	u := user.NewUser(
		existingUser.Username,
		existingUser.Email,
		user.HashedPassword(hashed),
		existingUser.Role,
	)
	u.ID = userID

	if err := uu.userRepo.UpdatePassword(ctx, &u); err != nil {
		return nil, fmt.Errorf("update password: %w", err)
	}

	if err := uu.refreshTokenRepo.RevokeAllForUser(ctx, userID); err != nil {
		return nil, fmt.Errorf("revoke refresh tokens: %w", err)
	}

	return &u, nil
}

func (uu *UserUsecase) Delete(ctx context.Context, userID user.ID) error {
	if err := uu.userRepo.Delete(ctx, userID); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	if err := uu.mediaService.DeleteAllForUser(ctx, userID); err != nil {
		slog.WarnContext(ctx, "media_delete_failed", "userId", fmt.Sprint(userID), "message", err.Error())
	}

	return nil
}
