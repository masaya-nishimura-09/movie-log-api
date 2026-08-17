package record

import (
	"time"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
)

type Record struct {
	ID          ID
	UserID      user.ID
	Title       Title
	ReleaseYear ReleaseYear
	Runtime     Runtime
	Genres      []Genre
	Countries   []Country
	Language    Language
	Credits     []Credit
	PosterURL   PosterURL
	WatchedAt   time.Time
	Platform    Platform
	Score       Score
	MoodTags    []MoodTag
	Memo        Memo
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ID uint

func NewRecord(
	userID user.ID,
	title Title,
	releaseYear ReleaseYear,
	runtime Runtime,
	genres []Genre,
	countries []Country,
	language Language,
	credits []Credit,
	posterURL PosterURL,
	watchedAt time.Time,
	platform Platform,
	score Score,
	moodTags []MoodTag,
	memo Memo,
) Record {
	return Record{
		UserID:      userID,
		Title:       title,
		ReleaseYear: releaseYear,
		Runtime:     runtime,
		Genres:      genres,
		Countries:   countries,
		Language:    language,
		Credits:     credits,
		PosterURL:   posterURL,
		WatchedAt:   watchedAt,
		Platform:    platform,
		Score:       score,
		MoodTags:    moodTags,
		Memo:        memo,
	}
}
