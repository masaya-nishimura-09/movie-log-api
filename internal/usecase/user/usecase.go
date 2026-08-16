package user

import (
	"context"
	"fmt"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/auth"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
	"golang.org/x/crypto/bcrypt"
)

type Usecase interface {
	GetByID(ctx context.Context, userID user.ID) (*user.User, error)
	Register(
		ctx context.Context,
		username user.Username,
		email user.Email,
		password user.Password,
	) (*user.User, error)
	UpdateUser(
		ctx context.Context,
		userID user.ID,
		username user.Username,
		email user.Email,
		password user.Password,
	) (*user.User, error)
	DeleteUser(ctx context.Context, userID user.ID) error
}

type UserUsecase struct {
	userRepo         user.UserRepository
	refreshTokenRepo auth.RefreshTokenRepository
}

func NewUserUsecase(
	userRepo user.UserRepository,
	refreshTokenRepo auth.RefreshTokenRepository,
) *UserUsecase {
	return &UserUsecase{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
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

func (uu *UserUsecase) Register(
	ctx context.Context,
	username user.Username,
	email user.Email,
	password user.Password,
) (*user.User, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	u := user.User{
		Username:       username,
		Email:          email,
		HashedPassword: user.HashedPassword(hashed),
		Role:           user.RoleUser,
	}

	if err := uu.userRepo.Create(ctx, &u); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return &u, nil
}

func (uu *UserUsecase) UpdateUser(
	ctx context.Context,
	userID user.ID,
	username user.Username,
	email user.Email,
	password user.Password,
) (*user.User, error) {
	existingUser, err := uu.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	passwordChanged := bcrypt.CompareHashAndPassword(
		[]byte(existingUser.HashedPassword),
		[]byte(password),
	) != nil

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	u := user.User{
		ID:             userID,
		Username:       username,
		Email:          email,
		HashedPassword: user.HashedPassword(hashed),
	}

	if err := uu.userRepo.Update(ctx, &u); err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}

	if passwordChanged {
		if err := uu.refreshTokenRepo.RevokeAllForUser(ctx, userID); err != nil {
			return nil, fmt.Errorf("revoke refresh tokens: %w", err)
		}
	}

	return &u, nil
}

func (uu *UserUsecase) DeleteUser(ctx context.Context, userID user.ID) error {
	if err := uu.userRepo.Delete(ctx, userID); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	return nil
}
