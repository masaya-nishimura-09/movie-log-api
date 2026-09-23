package record

import (
	"context"
	"fmt"
	"log"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/media"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/record"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
)

type Usecase interface {
	GetByID(
		ctx context.Context,
		userID user.ID,
		recordID record.ID,
	) (*record.Record, error)
	ListByUserID(
		ctx context.Context, 
		userID user.ID,
		query record.Query,
	) (record.ListResult, error)
	Create(
		ctx context.Context,
		userID user.ID,
		r record.Record,
	) (*record.Record, error)
	Update(
		ctx context.Context,
		userID user.ID,
		recordID record.ID,
		r record.Record,
	) (*record.Record, error)
	Delete(ctx context.Context, userID user.ID, recordID record.ID) error
}

type RecordUsecase struct {
	recordRepo   record.RecordRepository
	mediaService media.Service
}

func NewRecordUsecase(
	recordRepo record.RecordRepository,
	mediaService media.Service,
) *RecordUsecase {
	return &RecordUsecase{
		recordRepo:   recordRepo,
		mediaService: mediaService,
	}
}

func (ru *RecordUsecase) GetByID(
	ctx context.Context,
	userID user.ID,
	recordID record.ID,
) (*record.Record, error) {
	r, err := ru.recordRepo.GetByID(ctx, recordID)
	if err != nil {
		return nil, fmt.Errorf("get record by id: %w", err)
	}

	if r.UserID != userID {
		return nil, exception.ErrNotFound
	}

	return r, nil
}

func (ru *RecordUsecase) ListByUserID(
	ctx context.Context,
	userID user.ID, 
	query record.Query,
) (record.ListResult, error) {
	listResult, err := ru.recordRepo.ListByUserID(ctx, userID, query)
	if err != nil {
		return record.ListResult{}, fmt.Errorf("list records by user id: %w", err)
	}

	return listResult, nil
}

func (ru *RecordUsecase) Create(
	ctx context.Context,
	userID user.ID,
	r record.Record,
) (*record.Record, error) {
	r.UserID = userID

	if err := ru.recordRepo.Create(ctx, &r); err != nil {
		return nil, fmt.Errorf("create record: %w", err)
	}

	return &r, nil
}

func (ru *RecordUsecase) Update(
	ctx context.Context,
	userID user.ID,
	recordID record.ID,
	r record.Record,
) (*record.Record, error) {
	current, err := ru.recordRepo.GetByID(ctx, recordID)
	if err != nil {
		return nil, fmt.Errorf("get record by id: %w", err)
	}

	if current.UserID != userID {
		return nil, exception.ErrNotFound
	}

	r.ID = recordID
	r.UserID = userID

	if err := ru.recordRepo.Update(ctx, &r); err != nil {
		return nil, fmt.Errorf("update record: %w", err)
	}

	if current.PosterURL != r.PosterURL {
		if err := ru.mediaService.Delete(ctx, userID, media.URL(current.PosterURL)); err != nil {
			log.Printf("delete media: %v", err)
		}
	}

	return &r, nil
}

func (ru *RecordUsecase) Delete(
	ctx context.Context,
	userID user.ID,
	recordID record.ID,
) error {
	r, err := ru.recordRepo.GetByID(ctx, recordID)
	if err != nil {
		return fmt.Errorf("get record by id: %w", err)
	}

	if r.UserID != userID {
		return exception.ErrNotFound
	}

	if err := ru.recordRepo.Delete(ctx, recordID); err != nil {
		return fmt.Errorf("delete record: %w", err)
	}

	if err := ru.mediaService.Delete(ctx, userID, media.URL(r.PosterURL)); err != nil {
		log.Printf("delete media: %v", err)
	}

	return nil
}
