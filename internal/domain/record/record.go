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

func NewRecord(
	userID user.ID,
	movieID movie.ID,
	watchedAt time.Time,
	memo Memo,
	watchCount WatchCount,
	score Score,
	platform Platform,
	moodTags []MoodTag,
) Record {
	return Record{
		UserID:     userID,
		MovieID:    movieID,
		WatchedAt:  watchedAt,
		Memo:       memo,
		WatchCount: watchCount,
		Score:      score,
		Platform:   platform,
		MoodTags:   moodTags,
	}
}
