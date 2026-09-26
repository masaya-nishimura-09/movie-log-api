package user

import (
	"context"
	"fmt"
	"log"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/auth"
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
	Update(
		ctx context.Context,
		userID user.ID,
		username user.Username,
		email user.Email,
		password user.Password,
	) (*user.User, error)
	Delete(ctx context.Context, userID user.ID) error
}

type UserUsecase struct {
	userRepo         user.UserRepository
	refreshTokenRepo auth.RefreshTokenRepository
	mediaService     media.Service
}

func NewUserUsecase(
	userRepo user.UserRepository,
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

func (uu *UserUsecase) Update(
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

	u := user.NewUser(
		username,
		email,
		user.HashedPassword(hashed),
		existingUser.Role,
	)
	u.ID = userID

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

func (uu *UserUsecase) Delete(ctx context.Context, userID user.ID) error {
	if err := uu.userRepo.Delete(ctx, userID); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	if err := uu.mediaService.DeleteAllForUser(ctx, userID); err != nil {
		log.Printf("delete media: %v", err)
	}

	return nil
}
