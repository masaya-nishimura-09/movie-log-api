package record

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/record"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
	"gorm.io/gorm"
)

type articleRepository struct {
	db *gorm.DB
}

func NewArticleRepository(db *gorm.DB) record.ArticleRepository {
	return &articleRepository{db}
}

type articleDTO struct {
	ID          uint `gorm:"primaryKey"`
	UserID      uint
	RecordID    uint
	Title       string
	Content     string
	ContentText string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (articleDTO) TableName() string {
	return "record_articles"
}

func toArticleDTO(a *record.Article, contentText string) articleDTO {
	return articleDTO{
		ID:          uint(a.ID),
		UserID:      uint(a.UserID),
		RecordID:    uint(a.RecordID),
		Title:       string(a.Title),
		Content:     string(a.Content),
		ContentText: contentText,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
	}
}

type articleDetailDTO struct {
	ID                uint
	UserID            uint
	RecordID          uint
	Title             string
	Content           string
	RecordTitle       string
	RecordReleaseYear uint
	RecordPosterURL   string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func fromArticleDetailDTO(dto *articleDetailDTO) *record.ArticleDetail {
	return &record.ArticleDetail{
		Article: record.Article{
			ID:        record.ArticleID(dto.ID),
			UserID:    user.ID(dto.UserID),
			RecordID:  record.ID(dto.RecordID),
			Title:     record.ArticleTitle(dto.Title),
			Content:   record.ArticleContent(dto.Content),
			CreatedAt: dto.CreatedAt,
			UpdatedAt: dto.UpdatedAt,
		},
		Record: record.ArticleRecordSummary{
			ID:          record.ID(dto.RecordID),
			Title:       record.Title(dto.RecordTitle),
			ReleaseYear: record.ReleaseYear(dto.RecordReleaseYear),
			PosterURL:   record.PosterURL(dto.RecordPosterURL),
		},
	}
}

var articleDetailColumns = []string{
	"record_articles.id",
	"record_articles.user_id",
	"record_articles.record_id",
	"record_articles.title",
	"record_articles.content",
	"records.title AS record_title",
	"records.release_year AS record_release_year",
	"records.poster_url AS record_poster_url",
	"record_articles.created_at",
	"record_articles.updated_at",
}

type articleSummaryDTO struct {
	ID                uint
	RecordID          uint
	Title             string
	RecordTitle       string
	RecordReleaseYear uint
	RecordPosterURL   string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func fromArticleSummaryDTO(dto *articleSummaryDTO) record.ArticleSummary {
	return record.ArticleSummary{
		ID: record.ArticleID(dto.ID),
		Record: record.ArticleRecordSummary{
			ID:          record.ID(dto.RecordID),
			Title:       record.Title(dto.RecordTitle),
			ReleaseYear: record.ReleaseYear(dto.RecordReleaseYear),
			PosterURL:   record.PosterURL(dto.RecordPosterURL),
		},
		Title:     record.ArticleTitle(dto.Title),
		CreatedAt: dto.CreatedAt,
		UpdatedAt: dto.UpdatedAt,
	}
}

var articleSummaryColumns = []string{
	"record_articles.id",
	"record_articles.record_id",
	"record_articles.title",
	"records.title AS record_title",
	"records.release_year AS record_release_year",
	"records.poster_url AS record_poster_url",
	"record_articles.created_at",
	"record_articles.updated_at",
}

func (r *articleRepository) GetByID(
	ctx context.Context,
	articleID record.ArticleID,
) (*record.ArticleDetail, error) {
	var dto articleDetailDTO
	result := r.db.WithContext(ctx).
		Model(&articleDTO{}).
		Joins("JOIN records ON records.id = record_articles.record_id").
		Select(articleDetailColumns).
		Where("record_articles.id = ?", uint(articleID)).
		Take(&dto)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, exception.ErrNotFound
	}
	if result.Error != nil {
		return nil, fmt.Errorf("get article by id: %w", result.Error)
	}
	return fromArticleDetailDTO(&dto), nil
}

func (r *articleRepository) ListByUserID(
	ctx context.Context,
	userID user.ID,
	query record.ArticleQuery,
) (record.ArticleListResult, error) {
	db := r.db.WithContext(ctx).
		Where("record_articles.user_id = ?", uint(userID))

	if query.RecordID != 0 {
		db = db.Where("record_articles.record_id = ?", uint(query.RecordID))
	}
	if query.Keyword != "" {
		escaped := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(string(query.Keyword))
		pattern := "%" + escaped + "%"
		db = db.Where(
			"(record_articles.title ILIKE ? OR record_articles.content_text ILIKE ?)",
			pattern, pattern,
		)
	}

	var filtered int64
	if err := db.Model(&articleDTO{}).Count(&filtered).Error; err != nil {
		return record.ArticleListResult{}, fmt.Errorf("count filtered articles: %w", err)
	}

	var total int64
	if err := r.db.WithContext(ctx).
		Model(&articleDTO{}).
		Where("user_id = ?", uint(userID)).
		Count(&total).Error; err != nil {
		return record.ArticleListResult{}, fmt.Errorf("count total articles: %w", err)
	}

	var dtos []articleSummaryDTO
	result := db.
		Model(&articleDTO{}).
		Joins("JOIN records ON records.id = record_articles.record_id").
		Select(articleSummaryColumns).
		Order(
			"record_articles." + string(query.SortField) + " " + string(query.SortOrder) +
				", record_articles.id DESC",
		).
		Limit(int(query.PerPage)).
		Offset(int((uint(query.Page) - 1) * uint(query.PerPage))).
		Find(&dtos)
	if result.Error != nil {
		return record.ArticleListResult{}, fmt.Errorf("list articles by user id: %w", result.Error)
	}

	summaries := make([]record.ArticleSummary, 0, len(dtos))
	for i := range dtos {
		summaries = append(summaries, fromArticleSummaryDTO(&dtos[i]))
	}

	filteredCount, err := record.NewFilteredCount(int(filtered))
	if err != nil {
		return record.ArticleListResult{}, fmt.Errorf("create filtered count: %w", err)
	}
	totalCount, err := record.NewTotalCount(int(total))
	if err != nil {
		return record.ArticleListResult{}, fmt.Errorf("create total count: %w", err)
	}
	return record.NewArticleListResult(summaries, filteredCount, totalCount), nil
}

func (r *articleRepository) Create(
	ctx context.Context,
	a *record.Article,
) error {
	contentText, err := extractPlainText(a.Content)
	if err != nil {
		return fmt.Errorf("create article: %w", err)
	}

	now := time.Now()
	a.CreatedAt = now
	a.UpdatedAt = now

	dto := toArticleDTO(a, contentText)
	result := r.db.WithContext(ctx).Create(&dto)
	if errors.Is(result.Error, gorm.ErrForeignKeyViolated) {
		return exception.ErrNotFound
	}
	if result.Error != nil {
		return fmt.Errorf("create article: %w", result.Error)
	}
	a.ID = record.ArticleID(dto.ID)
	return nil
}

func (r *articleRepository) Update(
	ctx context.Context,
	a *record.Article,
) error {
	contentText, err := extractPlainText(a.Content)
	if err != nil {
		return fmt.Errorf("update article: %w", err)
	}

	a.UpdatedAt = time.Now()

	dto := toArticleDTO(a, contentText)
	result := r.db.WithContext(ctx).
		Model(&articleDTO{}).
		Where("id = ? AND user_id = ?", dto.ID, dto.UserID).
		Select("title", "content", "content_text", "updated_at").
		Updates(articleDTO{
			Title:       dto.Title,
			Content:     dto.Content,
			ContentText: dto.ContentText,
			UpdatedAt:   dto.UpdatedAt,
		})
	if result.Error != nil {
		return fmt.Errorf("update article: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return exception.ErrNotFound
	}
	return nil
}

func (r *articleRepository) Delete(
	ctx context.Context,
	articleID record.ArticleID,
) error {
	result := r.db.WithContext(ctx).Delete(&articleDTO{}, uint(articleID))
	if result.Error != nil {
		return fmt.Errorf("delete article: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return exception.ErrNotFound
	}
	return nil
}
