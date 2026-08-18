package record

import (
	"context"
	"fmt"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/record"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
)

type Usecase interface {
	GetByID(
		ctx context.Context,
		userID user.ID,
		recordID record.ID,
	) (*record.Record, error)
	ListByUserID(ctx context.Context, userID user.ID) ([]*record.Record, error)
	CreateRecord(
		ctx context.Context,
		userID user.ID,
		r record.Record,
	) (*record.Record, error)
	UpdateRecord(
		ctx context.Context,
		userID user.ID,
		recordID record.ID,
		r record.Record,
	) (*record.Record, error)
	DeleteRecord(ctx context.Context, userID user.ID, recordID record.ID) error
}

type RecordUsecase struct {
	recordRepo record.RecordRepository
}

func NewRecordUsecase(recordRepo record.RecordRepository) *RecordUsecase {
	return &RecordUsecase{recordRepo: recordRepo}
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
) ([]*record.Record, error) {
	records, err := ru.recordRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list records by user id: %w", err)
	}

	return records, nil
}

func (ru *RecordUsecase) CreateRecord(
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

func (ru *RecordUsecase) UpdateRecord(
	ctx context.Context,
	userID user.ID,
	recordID record.ID,
	r record.Record,
) (*record.Record, error) {
	r.ID = recordID
	r.UserID = userID

	if err := ru.recordRepo.Update(ctx, &r); err != nil {
		return nil, fmt.Errorf("update record: %w", err)
	}

	return &r, nil
}

func (ru *RecordUsecase) DeleteRecord(
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

	return nil
}
