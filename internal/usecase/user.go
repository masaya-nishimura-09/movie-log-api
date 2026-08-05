package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/model"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/repository"
	"golang.org/x/crypto/bcrypt"
)

type UserUsecase struct {
	userRepo  repository.UserRepository
	tokenRepo repository.TokenRepository
}

func NewUserUsecase(userRepo repository.UserRepository, tokenRepo repository.TokenRepository) *UserUsecase {
	return &UserUsecase{userRepo: userRepo, tokenRepo: tokenRepo}
}

func (uu *UserUsecase) Login(ctx context.Context, email model.Email, password model.Password) (model.Token, error) {
	existingUser, err := uu.userRepo.GetByEmail(ctx, email)
	if errors.Is(err, model.ErrUserNotFound) {
		return model.Token(""), model.ErrInvalidCredentials
	}
	if err != nil {
		return model.Token(""), fmt.Errorf("login: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(existingUser.HashedPassword), []byte(password)); err != nil {
		return model.Token(""), model.ErrInvalidCredentials
	}

	principal := model.Principal{
		UserID: model.UserID(existingUser.ID),
		Role:   model.Role(existingUser.Role),
	}

	token, err := uu.tokenRepo.Generate(&principal)
	if err != nil {
		return model.Token(""), fmt.Errorf("generate token: %w", err)
	}

	return token, nil
}

func (uu *UserUsecase) Register(
	ctx context.Context,
	username model.Username,
	email model.Email,
	password model.Password,
) (model.Token, error) {
	if err := model.ValidateUsername(username); err != nil {
		return model.Token(""), err
	}
	if err := model.ValidateEmail(email); err != nil {
		return model.Token(""), err
	}
	if err := model.ValidatePassword(password); err != nil {
		return model.Token(""), err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return model.Token(""), fmt.Errorf("hash password: %w", err)
	}

	user := model.User{
		Username:       username,
		Email:          email,
		HashedPassword: model.HashedPassword(hashed),
		Role:           model.RoleUser,
	}

	if err := uu.userRepo.Create(ctx, &user); err != nil {
		return model.Token(""), err
	}

	principal := model.Principal{
		UserID: user.ID,
		Role:   user.Role,
	}
	token, err := uu.tokenRepo.Generate(&principal)
	if err != nil {
		return model.Token(""), fmt.Errorf("generate token: %w", err)
	}

	return token, nil
}

func (uu *UserUsecase) UpdateUser(
	ctx context.Context,
	userID model.UserID,
	username model.Username,
	email model.Email,
) error {
	if err := model.ValidateUsername(username); err != nil {
		return err
	}
	if err := model.ValidateEmail(email); err != nil {
		return err
	}

	user := model.User{
		ID:       userID,
		Username: username,
		Email:    email,
	}

	if err := uu.userRepo.Update(ctx, &user); err != nil {
		return err
	}

	return nil
}

func (uu *UserUsecase) UpdatePassword(
	ctx context.Context,
	userID model.UserID,
	password model.Password,
) error {
	if err := model.ValidatePassword(password); err != nil {
		return err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	user := model.User{
		ID:             userID,
		HashedPassword: model.HashedPassword(hashed),
	}

	if err := uu.userRepo.UpdatePassword(ctx, &user); err != nil {
		return err
	}

	return nil
}

func (uu *UserUsecase) DeleteUser(ctx context.Context, userID model.UserID) error {
	if err := uu.userRepo.Delete(ctx, userID); err != nil {
		return err
	}

	return nil
}
