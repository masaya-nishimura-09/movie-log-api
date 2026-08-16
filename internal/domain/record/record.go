package record

import (
	"time"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/movie"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
)

type Record struct {
	ID         ID
	UserID     user.ID
	MovieID    movie.ID
	WatchedAt  time.Time
	Memo       Memo
	WatchCount WatchCount
	Score      Score
	Platform   Platform
	MoodTags   []MoodTag
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type ID uint
type WatchCount uint
