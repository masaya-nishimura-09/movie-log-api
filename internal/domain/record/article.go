package record

import (
	"time"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
)

type Article struct {
	ID        ArticleID
	UserID    user.ID
	RecordID  ID
	Title     ArticleTitle
	Content   ArticleContent
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ArticleID uint

func NewArticle(
	userID user.ID,
	recordID ID,
	title ArticleTitle,
	content ArticleContent,
) Article {
	return Article{
		UserID:   userID,
		RecordID: recordID,
		Title:    title,
		Content:  content,
	}
}
