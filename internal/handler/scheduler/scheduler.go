package scheduler

import (
	"context"
	"log/slog"
	"time"

	authusecase "github.com/masaya-nishimura-09/movie-log-api/internal/usecase/auth"
	userusecase "github.com/masaya-nishimura-09/movie-log-api/internal/usecase/user"
)

type Scheduler struct {
	authUsecase authusecase.Usecase
	userUsecase userusecase.Usecase
	interval    time.Duration
}

func NewScheduler(
	authUsecase authusecase.Usecase,
	userUsecase userusecase.Usecase,
	interval time.Duration,
) *Scheduler {
	return &Scheduler{authUsecase: authUsecase, userUsecase: userUsecase, interval: interval}
}

func (s *Scheduler) Run(ctx context.Context) {
	s.cleanup(ctx)

	ticker := time.NewTicker(s.interval)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.cleanup(ctx)
		}
	}
}

func (s *Scheduler) cleanup(ctx context.Context) {
	slog.InfoContext(ctx, "refresh_token_cleanup_run", "interval", s.interval.String())
	if err := s.authUsecase.DeleteExpiredRefreshTokens(ctx); err != nil {
		slog.ErrorContext(ctx, "refresh_token_cleanup_failed", "message", err.Error())
	} else {
		slog.InfoContext(ctx, "refresh_token_cleanup_completed")
	}

	slog.InfoContext(ctx, "email_change_token_cleanup_run", "interval", s.interval.String())
	if err := s.userUsecase.DeleteExpiredEmailChangeTokens(ctx); err != nil {
		slog.ErrorContext(ctx, "email_change_token_cleanup_failed", "message", err.Error())
	} else {
		slog.InfoContext(ctx, "email_change_token_cleanup_completed")
	}

	slog.InfoContext(ctx, "unverified_user_cleanup_run", "interval", s.interval.String())
	if err := s.userUsecase.DeleteUnverifiedUsers(ctx); err != nil {
		slog.ErrorContext(ctx, "unverified_user_cleanup_failed", "message", err.Error())
	} else {
		slog.InfoContext(ctx, "unverified_user_cleanup_completed")
	}
}
