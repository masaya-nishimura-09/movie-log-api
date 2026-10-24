package record

import (
	"context"
	"fmt"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/record"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
)

type ArticleUsecase interface {
	GetByID(
		ctx context.Context,
		userID user.ID,
		articleID record.ArticleID,
	) (*record.ArticleDetail, error)
	ListByUserID(
		ctx context.Context,
		userID user.ID,
		query record.ArticleQuery,
	) (record.ArticleListResult, error)
	Create(
		ctx context.Context,
		userID user.ID,
		a record.Article,
	) (*record.ArticleDetail, error)
	Update(
		ctx context.Context,
		userID user.ID,
		articleID record.ArticleID,
		a record.Article,
	) (*record.ArticleDetail, error)
	Delete(ctx context.Context, userID user.ID, articleID record.ArticleID) error
}

type articleUsecase struct {
	articleRepo record.ArticleRepository
	recordRepo  record.Repository
}

func NewArticleUsecase(
	articleRepo record.ArticleRepository,
	recordRepo record.Repository,
) ArticleUsecase {
	return &articleUsecase{
		articleRepo: articleRepo,
		recordRepo:  recordRepo,
	}
}

func (au *articleUsecase) GetByID(
	ctx context.Context,
	userID user.ID,
	articleID record.ArticleID,
) (*record.ArticleDetail, error) {
	d, err := au.articleRepo.GetByID(ctx, articleID)
	if err != nil {
		return nil, fmt.Errorf("get article by id: %w", err)
	}

	if d.Article.UserID != userID {
		return nil, exception.ErrNotFound
	}

	return d, nil
}

func (au *articleUsecase) ListByUserID(
	ctx context.Context,
	userID user.ID,
	query record.ArticleQuery,
) (record.ArticleListResult, error) {
	listResult, err := au.articleRepo.ListByUserID(ctx, userID, query)
	if err != nil {
		return record.ArticleListResult{}, fmt.Errorf("list articles by user id: %w", err)
	}

	return listResult, nil
}

func (au *articleUsecase) Create(
	ctx context.Context,
	userID user.ID,
	a record.Article,
) (*record.ArticleDetail, error) {
	r, err := au.recordRepo.GetByID(ctx, a.RecordID)
	if err != nil {
		return nil, fmt.Errorf("get record by id: %w", err)
	}

	if r.UserID != userID {
		return nil, exception.ErrNotFound
	}

	a.UserID = userID

	if err := au.articleRepo.Create(ctx, &a); err != nil {
		return nil, fmt.Errorf("create article: %w", err)
	}

	return &record.ArticleDetail{
		Article: a,
		Record: record.ArticleRecordSummary{
			ID:          r.ID,
			Title:       r.Title,
			ReleaseYear: r.ReleaseYear,
			PosterURL:   r.PosterURL,
		},
	}, nil
}

func (au *articleUsecase) Update(
	ctx context.Context,
	userID user.ID,
	articleID record.ArticleID,
	a record.Article,
) (*record.ArticleDetail, error) {
	current, err := au.articleRepo.GetByID(ctx, articleID)
	if err != nil {
		return nil, fmt.Errorf("get article by id: %w", err)
	}

	if current.Article.UserID != userID {
		return nil, exception.ErrNotFound
	}

	a.ID = articleID
	a.UserID = userID
	a.RecordID = current.Article.RecordID
	a.CreatedAt = current.Article.CreatedAt

	if err := au.articleRepo.Update(ctx, &a); err != nil {
		return nil, fmt.Errorf("update article: %w", err)
	}

	return &record.ArticleDetail{
		Article: a,
		Record:  current.Record,
	}, nil
}

func (au *articleUsecase) Delete(
	ctx context.Context,
	userID user.ID,
	articleID record.ArticleID,
) error {
	d, err := au.articleRepo.GetByID(ctx, articleID)
	if err != nil {
		return fmt.Errorf("get article by id: %w", err)
	}

	if d.Article.UserID != userID {
		return exception.ErrNotFound
	}

	if err := au.articleRepo.Delete(ctx, articleID); err != nil {
		return fmt.Errorf("delete article: %w", err)
	}

	return nil
}
