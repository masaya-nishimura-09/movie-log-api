package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/auth"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/mail"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
)

type fakeAuthUsecase struct {
	called chan struct{}
}

func (au *fakeAuthUsecase) ValidateAccessToken(
	ctx context.Context,
	accessToken *auth.AccessToken,
) (*auth.Principal, error) {
	return nil, nil
}

func (au *fakeAuthUsecase) Login(
	ctx context.Context,
	email user.Email,
	password user.Password,
) (*auth.AccessToken, *auth.RefreshToken, error) {
	return nil, nil, nil
}

func (au *fakeAuthUsecase) Logout(
	ctx context.Context,
	refreshTokenValue auth.RefreshTokenValue,
) error {
	return nil
}

func (au *fakeAuthUsecase) Refresh(
	ctx context.Context,
	refreshTokenValue auth.RefreshTokenValue,
) (*auth.AccessToken, *auth.RefreshToken, error) {
	return nil, nil, nil
}

func (au *fakeAuthUsecase) DeleteExpiredRefreshTokens(
	ctx context.Context,
) error {
	au.called <- struct{}{}
	return nil
}

type fakeUserUsecase struct {
	called chan struct{}
}

func (uu *fakeUserUsecase) GetByID(
	ctx context.Context,
	userID user.ID,
) (*user.User, error) {
	return nil, nil
}

func (uu *fakeUserUsecase) Create(
	ctx context.Context,
	username user.Username,
	email user.Email,
	password user.Password,
	language mail.Language,
) error {
	return nil
}

func (uu *fakeUserUsecase) VerifyEmail(
	ctx context.Context,
	value user.EmailVerificationTokenValue,
) error {
	return nil
}

func (uu *fakeUserUsecase) ResendVerificationEmail(
	ctx context.Context,
	email user.Email,
	language mail.Language,
) error {
	return nil
}

func (uu *fakeUserUsecase) UpdateUsername(
	ctx context.Context,
	userID user.ID,
	username user.Username,
) (*user.User, error) {
	return nil, nil
}

func (uu *fakeUserUsecase) UpdateEmail(
	ctx context.Context,
	userID user.ID,
	email user.Email,
	password user.Password,
) (*user.User, error) {
	return nil, nil
}

func (uu *fakeUserUsecase) UpdatePassword(
	ctx context.Context,
	userID user.ID,
	oldPassword user.Password,
	newPassword user.Password,
) (*user.User, error) {
	return nil, nil
}

func (uu *fakeUserUsecase) Delete(
	ctx context.Context,
	userID user.ID,
) error {
	return nil
}

func (uu *fakeUserUsecase) DeleteUnverifiedUsers(
	ctx context.Context,
) error {
	uu.called <- struct{}{}
	return nil
}

func TestRun(t *testing.T) {
	t.Run(
		"calls DeleteExpiredRefreshTokens as soon as it starts",
		func(t *testing.T) {
			usecase := &fakeAuthUsecase{called: make(chan struct{}, 1)}
			userUsecase := &fakeUserUsecase{called: make(chan struct{}, 1)}
			scheduler := NewScheduler(usecase, userUsecase, time.Minute)

			ctx := t.Context()

			go scheduler.Run(ctx)
			select {
			case <-usecase.called:
			case <-time.After(time.Second):
				t.Fatalf("Run(ctx) did not call DeleteExpiredRefreshTokens within 1s")
			}
		},
	)

	t.Run(
		"calls DeleteExpiredRefreshTokens again after the interval",
		func(t *testing.T) {
			usecase := &fakeAuthUsecase{called: make(chan struct{}, 2)}
			userUsecase := &fakeUserUsecase{called: make(chan struct{}, 2)}
			scheduler := NewScheduler(usecase, userUsecase, 10*time.Millisecond)

			ctx := t.Context()

			go scheduler.Run(ctx)
			for range 2 {
				select {
				case <-usecase.called:
				case <-time.After(time.Second):
					t.Fatalf("Run(ctx) did not call DeleteExpiredRefreshTokens again within 1s")
				}
			}
		},
	)

	t.Run(
		"calls DeleteUnverifiedUsers as soon as it starts",
		func(t *testing.T) {
			usecase := &fakeAuthUsecase{called: make(chan struct{}, 1)}
			userUsecase := &fakeUserUsecase{called: make(chan struct{}, 1)}
			scheduler := NewScheduler(usecase, userUsecase, time.Minute)

			ctx := t.Context()

			go scheduler.Run(ctx)
			select {
			case <-userUsecase.called:
			case <-time.After(time.Second):
				t.Fatalf("Run(ctx) did not call DeleteUnverifiedUsers within 1s")
			}
		},
	)

	t.Run(
		"calls DeleteUnverifiedUsers again after the interval",
		func(t *testing.T) {
			usecase := &fakeAuthUsecase{called: make(chan struct{}, 2)}
			userUsecase := &fakeUserUsecase{called: make(chan struct{}, 2)}
			scheduler := NewScheduler(usecase, userUsecase, 10*time.Millisecond)

			ctx := t.Context()

			go scheduler.Run(ctx)
			for range 2 {
				select {
				case <-userUsecase.called:
				case <-time.After(time.Second):
					t.Fatalf("Run(ctx) did not call DeleteUnverifiedUsers again within 1s")
				}
			}
		},
	)

	t.Run(
		"returns when the context is canceled",
		func(t *testing.T) {
			usecase := &fakeAuthUsecase{called: make(chan struct{}, 1)}
			userUsecase := &fakeUserUsecase{called: make(chan struct{}, 1)}
			scheduler := NewScheduler(usecase, userUsecase, time.Hour)

			ctx, cancel := context.WithCancel(t.Context())

			done := make(chan struct{})
			go func() {
				scheduler.Run(ctx)
				close(done)
			}()

			cancel()

			select {
			case <-done:
			case <-time.After(time.Second):
				t.Fatalf("Run(ctx) did not return within 1s after the context was canceled")
			}
		},
	)
}
