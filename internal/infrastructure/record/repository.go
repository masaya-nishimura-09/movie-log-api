package record

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/record"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
	"gorm.io/gorm"
)

type recordRepository struct {
	db *gorm.DB
}

func NewRecordRepo(db *gorm.DB) record.RecordRepository {
	return &recordRepository{db}
}

type recordDTO struct {
	ID          uint `gorm:"primaryKey"`
	UserID      uint
	Title       string
	ReleaseYear uint
	Runtime     uint
	Genres      []genreDTO   `gorm:"foreignKey:RecordID"`
	Countries   []countryDTO `gorm:"foreignKey:RecordID"`
	Language    string
	Credits     []creditDTO `gorm:"foreignKey:RecordID"`
	PosterURL   string
	WatchedAt   time.Time
	Platform    string
	Score       uint
	MoodTags    []moodTagDTO `gorm:"foreignKey:RecordID"`
	Memo        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type genreDTO struct {
	ID       uint `gorm:"primaryKey"`
	RecordID uint
	Value    string
}

type countryDTO struct {
	ID       uint `gorm:"primaryKey"`
	RecordID uint
	Value    string
}

type creditDTO struct {
	ID         uint `gorm:"primaryKey"`
	RecordID   uint
	PersonName string
	CreditRole string
}

type moodTagDTO struct {
	ID       uint `gorm:"primaryKey"`
	RecordID uint
	Value    string
}

func (recordDTO) TableName() string {
	return "records"
}

func (genreDTO) TableName() string {
	return "record_genres"
}

func (countryDTO) TableName() string {
	return "record_countries"
}

func (creditDTO) TableName() string {
	return "record_credits"
}

func (moodTagDTO) TableName() string {
	return "record_mood_tags"
}

func toDTO(r *record.Record) recordDTO {
	genres := make([]genreDTO, 0, len(r.Genres))
	for _, genre := range r.Genres {
		genres = append(genres, genreDTO{RecordID: uint(r.ID), Value: string(genre)})
	}

	countries := make([]countryDTO, 0, len(r.Countries))
	for _, country := range r.Countries {
		countries = append(countries, countryDTO{RecordID: uint(r.ID), Value: string(country)})
	}

	credits := make([]creditDTO, 0, len(r.Credits))
	for _, credit := range r.Credits {
		credits = append(credits, creditDTO{
			RecordID:   uint(r.ID),
			PersonName: string(credit.PersonName),
			CreditRole: string(credit.CreditRole),
		})
	}

	moodTags := make([]moodTagDTO, 0, len(r.MoodTags))
	for _, moodTag := range r.MoodTags {
		moodTags = append(moodTags, moodTagDTO{RecordID: uint(r.ID), Value: string(moodTag)})
	}

	return recordDTO{
		ID:          uint(r.ID),
		UserID:      uint(r.UserID),
		Title:       string(r.Title),
		ReleaseYear: uint(r.ReleaseYear),
		Runtime:     uint(r.Runtime),
		Genres:      genres,
		Countries:   countries,
		Language:    string(r.Language),
		Credits:     credits,
		PosterURL:   string(r.PosterURL),
		WatchedAt:   r.WatchedAt,
		Platform:    string(r.Platform),
		Score:       uint(r.Score),
		MoodTags:    moodTags,
		Memo:        string(r.Memo),
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

func fromDTO(dto *recordDTO) *record.Record {
	genres := make([]record.Genre, 0, len(dto.Genres))
	for _, genre := range dto.Genres {
		genres = append(genres, record.Genre(genre.Value))
	}

	countries := make([]record.Country, 0, len(dto.Countries))
	for _, country := range dto.Countries {
		countries = append(countries, record.Country(country.Value))
	}

	credits := make([]record.Credit, 0, len(dto.Credits))
	for _, credit := range dto.Credits {
		credits = append(credits, record.Credit{
			PersonName: record.PersonName(credit.PersonName),
			CreditRole: record.CreditRole(credit.CreditRole),
		})
	}

	moodTags := make([]record.MoodTag, 0, len(dto.MoodTags))
	for _, moodTag := range dto.MoodTags {
		moodTags = append(moodTags, record.MoodTag(moodTag.Value))
	}

	return &record.Record{
		ID:          record.ID(dto.ID),
		UserID:      user.ID(dto.UserID),
		Title:       record.Title(dto.Title),
		ReleaseYear: record.ReleaseYear(dto.ReleaseYear),
		Runtime:     record.Runtime(dto.Runtime),
		Genres:      genres,
		Countries:   countries,
		Language:    record.Language(dto.Language),
		Credits:     credits,
		PosterURL:   record.PosterURL(dto.PosterURL),
		WatchedAt:   dto.WatchedAt,
		Platform:    record.Platform(dto.Platform),
		Score:       record.Score(dto.Score),
		MoodTags:    moodTags,
		Memo:        record.Memo(dto.Memo),
		CreatedAt:   dto.CreatedAt,
		UpdatedAt:   dto.UpdatedAt,
	}
}

func (r *recordRepository) GetByID(
	ctx context.Context,
	recordID record.ID,
) (*record.Record, error) {
	var dto recordDTO
	result := r.db.WithContext(ctx).
		Preload("Genres").
		Preload("Countries").
		Preload("Credits").
		Preload("MoodTags").
		First(&dto, recordID)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, exception.ErrNotFound
	}
	if result.Error != nil {
		return nil, fmt.Errorf("get record by id: %w", result.Error)
	}
	return fromDTO(&dto), nil
}

func (r *recordRepository) ListByUserID(
	ctx context.Context,
	userID user.ID,
) ([]*record.Record, error) {
	var dtos []recordDTO
	result := r.db.WithContext(ctx).
		Preload("Genres").
		Preload("Countries").
		Preload("Credits").
		Preload("MoodTags").
		Where("user_id = ?", uint(userID)).
		Order("watched_at DESC, id DESC").
		Find(&dtos)
	if result.Error != nil {
		return nil, fmt.Errorf("list records by user id: %w", result.Error)
	}

	records := make([]*record.Record, 0, len(dtos))
	for i := range dtos {
		records = append(records, fromDTO(&dtos[i]))
	}
	return records, nil
}

func (r *recordRepository) Create(
	ctx context.Context,
	rec *record.Record,
) error {
	now := time.Now()
	rec.CreatedAt = now
	rec.UpdatedAt = now

	dto := toDTO(rec)
	result := r.db.WithContext(ctx).Create(&dto)
	if result.Error != nil {
		return fmt.Errorf("create record: %w", result.Error)
	}
	rec.ID = record.ID(dto.ID)
	return nil
}

func (r *recordRepository) Update(
	ctx context.Context,
	rec *record.Record,
) error {
	rec.UpdatedAt = time.Now()

	dto := toDTO(rec)
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&recordDTO{}).
			Where("id = ? AND user_id = ?", dto.ID, dto.UserID).
			Select(
				"title",
				"release_year",
				"runtime",
				"language",
				"poster_url",
				"watched_at",
				"platform",
				"score",
				"memo",
				"updated_at",
			).
			Updates(recordDTO{
				Title:       dto.Title,
				ReleaseYear: dto.ReleaseYear,
				Runtime:     dto.Runtime,
				Language:    dto.Language,
				PosterURL:   dto.PosterURL,
				WatchedAt:   dto.WatchedAt,
				Platform:    dto.Platform,
				Score:       dto.Score,
				Memo:        dto.Memo,
				UpdatedAt:   dto.UpdatedAt,
			})
		if result.Error != nil {
			return fmt.Errorf("update record: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return exception.ErrNotFound
		}

		for _, child := range []any{&genreDTO{}, &countryDTO{}, &creditDTO{}, &moodTagDTO{}} {
			if err := tx.Where("record_id = ?", dto.ID).Delete(child).Error; err != nil {
				return fmt.Errorf("delete record associations: %w", err)
			}
		}

		if len(dto.Genres) > 0 {
			if err := tx.Create(&dto.Genres).Error; err != nil {
				return fmt.Errorf("create record genres: %w", err)
			}
		}
		if len(dto.Countries) > 0 {
			if err := tx.Create(&dto.Countries).Error; err != nil {
				return fmt.Errorf("create record countries: %w", err)
			}
		}
		if len(dto.Credits) > 0 {
			if err := tx.Create(&dto.Credits).Error; err != nil {
				return fmt.Errorf("create record credits: %w", err)
			}
		}
		if len(dto.MoodTags) > 0 {
			if err := tx.Create(&dto.MoodTags).Error; err != nil {
				return fmt.Errorf("create record mood tags: %w", err)
			}
		}
		return nil
	})
}

func (r *recordRepository) Delete(ctx context.Context, recordID record.ID) error {
	result := r.db.WithContext(ctx).Delete(&recordDTO{}, uint(recordID))
	if result.Error != nil {
		return fmt.Errorf("delete record: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return exception.ErrNotFound
	}
	return nil
}
