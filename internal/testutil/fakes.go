package testutil

import (
	"context"
	"fmt"
	"time"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/auth"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
)

type FakeUserRepository struct {
	Users     map[user.ID]user.User
	NextID    user.ID
	CreateErr error
	UpdateErr error
	DeleteErr error
}

func NewFakeUserRepository() *FakeUserRepository {
	return &FakeUserRepository{Users: make(map[user.ID]user.User)}
}

func (f *FakeUserRepository) GetByID(ctx context.Context, id user.ID) (*user.User, error) {
	u, ok := f.Users[id]
	if !ok {
		return nil, exception.ErrNotFound
	}
	return &u, nil
}

func (f *FakeUserRepository) GetByEmail(ctx context.Context, email user.Email) (*user.User, error) {
	for _, u := range f.Users {
		if u.Email == email {
			return &u, nil
		}
	}
	return nil, exception.ErrNotFound
}

func (f *FakeUserRepository) Create(ctx context.Context, u *user.User) error {
	if f.CreateErr != nil {
		return f.CreateErr
	}
	for _, existing := range f.Users {
		if existing.Email == u.Email {
			return exception.ErrAlreadyExists
		}
	}
	f.NextID++
	u.ID = f.NextID
	f.Users[u.ID] = *u
	return nil
}

func (f *FakeUserRepository) Update(ctx context.Context, u *user.User) error {
	if f.UpdateErr != nil {
		return f.UpdateErr
	}
	if _, ok := f.Users[u.ID]; !ok {
		return exception.ErrNotFound
	}
	f.Users[u.ID] = *u
	return nil
}

func (f *FakeUserRepository) Delete(ctx context.Context, id user.ID) error {
	if f.DeleteErr != nil {
		return f.DeleteErr
	}
	if _, ok := f.Users[id]; !ok {
		return exception.ErrNotFound
	}
	delete(f.Users, id)
	return nil
}

type FakeRefreshTokenRepository struct {
	Tokens                map[auth.RefreshTokenID]auth.RefreshToken
	NextID                auth.RefreshTokenID
	RevokeAllForUserCalls []user.ID
}

func NewFakeRefreshTokenRepository() *FakeRefreshTokenRepository {
	return &FakeRefreshTokenRepository{Tokens: make(map[auth.RefreshTokenID]auth.RefreshToken)}
}

func (f *FakeRefreshTokenRepository) Create(ctx context.Context, principal *auth.Principal) (*auth.RefreshToken, error) {
	f.NextID++
	rt := auth.RefreshToken{
		ID:        f.NextID,
		Principal: *principal,
		Value:     auth.RefreshTokenValue(fmt.Sprintf("token-%d", f.NextID)),
		Hash:      auth.RefreshTokenHash(fmt.Sprintf("hash-%d", f.NextID)),
		ExpiresAt: time.Now().Add(time.Hour),
		CreatedAt: time.Now(),
	}
	f.Tokens[rt.ID] = rt
	return &rt, nil
}

func (f *FakeRefreshTokenRepository) FindValidByValue(ctx context.Context, value auth.RefreshTokenValue) (*auth.RefreshToken, error) {
	for _, rt := range f.Tokens {
		if rt.Value != value {
			continue
		}
		if rt.RevokedAt != nil || rt.ExpiresAt.Before(time.Now()) {
			return nil, exception.ErrInvalid
		}
		return &rt, nil
	}
	return nil, exception.ErrInvalid
}

func (f *FakeRefreshTokenRepository) Revoke(ctx context.Context, id auth.RefreshTokenID) error {
	rt, ok := f.Tokens[id]
	if !ok {
		return exception.ErrInvalid
	}
	now := time.Now()
	rt.RevokedAt = &now
	f.Tokens[id] = rt
	return nil
}

func (f *FakeRefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID user.ID) error {
	f.RevokeAllForUserCalls = append(f.RevokeAllForUserCalls, userID)
	now := time.Now()
	for id, rt := range f.Tokens {
		if rt.Principal.UserID == userID && rt.RevokedAt == nil {
			rt.RevokedAt = &now
			f.Tokens[id] = rt
		}
	}
	return nil
}

type FakeAccessTokenService struct {
	GenerateErr       error
	ValidateErr       error
	ValidatePrincipal *auth.Principal
}

func (f *FakeAccessTokenService) Generate(ctx context.Context, principal *auth.Principal) (*auth.AccessToken, error) {
	if f.GenerateErr != nil {
		return nil, f.GenerateErr
	}
	return &auth.AccessToken{
		Value:     auth.AccessTokenValue(fmt.Sprintf("access-token-for-%d", principal.UserID)),
		Principal: *principal,
		ExpiresAt: time.Now().Add(time.Hour),
	}, nil
}

func (f *FakeAccessTokenService) Validate(ctx context.Context, accessToken *auth.AccessToken) (*auth.Principal, error) {
	if f.ValidateErr != nil {
		return nil, f.ValidateErr
	}
	if f.ValidatePrincipal != nil {
		return f.ValidatePrincipal, nil
	}
	return &accessToken.Principal, nil
}
