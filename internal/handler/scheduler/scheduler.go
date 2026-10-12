package scheduler

import (
	"context"
	"log/slog"
	"time"

	authusecase "github.com/masaya-nishimura-09/movie-log-api/internal/usecase/auth"
)

type Scheduler struct {
	authUsecase authusecase.Usecase
	interval    time.Duration
}

func NewScheduler(authUsecase authusecase.Usecase, interval time.Duration) *Scheduler {
	return &Scheduler{authUsecase: authUsecase, interval: interval}
}

func (s *Scheduler) Run(ctx context.Context) {
	slog.InfoContext(ctx, "refresh_token_cleanup_run", "interval", s.interval.String())
	if err := s.authUsecase.DeleteExpiredRefreshTokens(ctx); err != nil {
		slog.ErrorContext(ctx, "refresh_token_cleanup_failed", "message", err.Error())
	} else {
		slog.InfoContext(ctx, "refresh_token_cleanup_completed")
	}

	ticker := time.NewTicker(s.interval)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			slog.InfoContext(ctx, "refresh_token_cleanup_run", "interval", s.interval.String())
			if err := s.authUsecase.DeleteExpiredRefreshTokens(ctx); err != nil {
				slog.ErrorContext(ctx, "refresh_token_cleanup_failed", "message", err.Error())
			} else {
				slog.InfoContext(ctx, "refresh_token_cleanup_completed")
			}
		}
	}
}
