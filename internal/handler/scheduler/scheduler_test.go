package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/auth"
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

func TestRun(t *testing.T) {
	t.Run(
		"calls DeleteExpiredRefreshTokens as soon as it starts",
		func(t *testing.T) {
			usecase := &fakeAuthUsecase{called: make(chan struct{}, 1)}
			scheduler := NewScheduler(usecase, time.Minute)

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
			scheduler := NewScheduler(usecase, 10*time.Millisecond)

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
		"returns when the context is canceled",
		func(t *testing.T) {
			usecase := &fakeAuthUsecase{called: make(chan struct{}, 1)}
			scheduler := NewScheduler(usecase, time.Hour)

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
