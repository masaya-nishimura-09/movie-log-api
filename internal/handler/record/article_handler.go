package record

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	recorddomain "github.com/masaya-nishimura-09/movie-log-api/internal/domain/record"
	userdomain "github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
	"github.com/masaya-nishimura-09/movie-log-api/internal/handler/response"
	"github.com/masaya-nishimura-09/movie-log-api/internal/logging"
	recordusecase "github.com/masaya-nishimura-09/movie-log-api/internal/usecase/record"
)

type CreateArticleReq struct {
	RecordID string          `json:"record_id"`
	Title    string          `json:"title"`
	Content  json.RawMessage `json:"content"`
}

func (r CreateArticleReq) toDomain(userID userdomain.ID) (recorddomain.Article, error) {
	if r.RecordID == "" {
		return recorddomain.Article{}, fmt.Errorf("%w: record id is required", exception.ErrInvalid)
	}
	recordID, err := strconv.ParseUint(r.RecordID, 10, 64)
	if err != nil {
		return recorddomain.Article{}, fmt.Errorf("%w: record id must be a number", exception.ErrInvalid)
	}
	title, err := recorddomain.NewArticleTitle(r.Title)
	if err != nil {
		return recorddomain.Article{}, err
	}
	content, err := recorddomain.NewArticleContent(r.Content)
	if err != nil {
		return recorddomain.Article{}, err
	}

	return recorddomain.NewArticle(
		userID,
		recorddomain.ID(recordID),
		title,
		content,
	), nil
}

type UpdateArticleReq struct {
	Title   string          `json:"title"`
	Content json.RawMessage `json:"content"`
}

func (r UpdateArticleReq) toDomain(userID userdomain.ID) (recorddomain.Article, error) {
	title, err := recorddomain.NewArticleTitle(r.Title)
	if err != nil {
		return recorddomain.Article{}, err
	}
	content, err := recorddomain.NewArticleContent(r.Content)
	if err != nil {
		return recorddomain.Article{}, err
	}

	return recorddomain.NewArticle(userID, 0, title, content), nil
}

func toArticleRecordSummaryResponse(s recorddomain.ArticleRecordSummary) gin.H {
	return gin.H{
		"record_id":    strconv.FormatUint(uint64(s.ID), 10),
		"title":        string(s.Title),
		"release_year": uint(s.ReleaseYear),
		"poster_url":   string(s.PosterURL),
	}
}

func toArticleResponse(d *recorddomain.ArticleDetail) gin.H {
	return gin.H{
		"article_id": strconv.FormatUint(uint64(d.Article.ID), 10),
		"record":     toArticleRecordSummaryResponse(d.Record),
		"title":      string(d.Article.Title),
		"content":    json.RawMessage(d.Article.Content),
		"created_at": d.Article.CreatedAt.UTC(),
		"updated_at": d.Article.UpdatedAt.UTC(),
	}
}

func toArticleSummaryResponse(s recorddomain.ArticleSummary) gin.H {
	return gin.H{
		"article_id": strconv.FormatUint(uint64(s.ID), 10),
		"record":     toArticleRecordSummaryResponse(s.Record),
		"title":      string(s.Title),
		"created_at": s.CreatedAt.UTC(),
		"updated_at": s.UpdatedAt.UTC(),
	}
}

type ArticleHandler struct {
	articleUsecase recordusecase.ArticleUsecase
}

func NewArticleHandler(usecase recordusecase.ArticleUsecase) *ArticleHandler {
	return &ArticleHandler{articleUsecase: usecase}
}

func getArticleRecordID(c *gin.Context) (recorddomain.ID, bool) {
	s := c.Query("record_id")
	if s == "" {
		return 0, true
	}
	id, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		response.InvalidInput(c, fmt.Errorf("%w: record id must be a number", exception.ErrInvalid))
		return 0, false
	}
	if id == 0 {
		response.InvalidInput(c, fmt.Errorf("%w: record id must be greater than 0", exception.ErrInvalid))
		return 0, false
	}
	return recorddomain.ID(id), true
}

func getArticleKeyword(c *gin.Context) (recorddomain.ArticleKeyword, bool) {
	k := c.Query("keyword")
	dk, err := recorddomain.NewArticleKeyword(k)
	if err != nil {
		response.InvalidInput(c, err)
		return "", false
	}
	return dk, true
}

func getArticleSortField(c *gin.Context) (recorddomain.ArticleSortField, bool) {
	t := c.Query("sort_field")
	if t == "" {
		t = "created_at"
	}
	dsf, err := recorddomain.NewArticleSortField(t)
	if err != nil {
		response.InvalidInput(c, err)
		return "", false
	}
	return dsf, true
}

func getArticleQuery(c *gin.Context) (recorddomain.ArticleQuery, bool) {
	recordID, ok := getArticleRecordID(c)
	if !ok {
		return recorddomain.ArticleQuery{}, false
	}

	keyword, ok := getArticleKeyword(c)
	if !ok {
		return recorddomain.ArticleQuery{}, false
	}

	sortField, ok := getArticleSortField(c)
	if !ok {
		return recorddomain.ArticleQuery{}, false
	}

	sortOrder, ok := getSortOrder(c)
	if !ok {
		return recorddomain.ArticleQuery{}, false
	}

	page, ok := getPage(c)
	if !ok {
		return recorddomain.ArticleQuery{}, false
	}

	perPage, ok := getPerPage(c)
	if !ok {
		return recorddomain.ArticleQuery{}, false
	}

	query := recorddomain.NewArticleQuery(
		recordID,
		keyword,
		sortField,
		sortOrder,
		page,
		perPage,
	)

	return query, true
}

func getArticleID(c *gin.Context) (recorddomain.ArticleID, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidInput(c, fmt.Errorf("%w: article id must be a number", exception.ErrInvalid))
		return 0, false
	}
	return recorddomain.ArticleID(id), true
}

func (ah *ArticleHandler) GetByID(c *gin.Context) {
	ctx := c.Request.Context()

	authUserID, ok := getUserID(c)
	if !ok {
		return
	}

	articleID, ok := getArticleID(c)
	if !ok {
		return
	}

	a, err := ah.articleUsecase.GetByID(ctx, authUserID, articleID)
	if errors.Is(err, exception.ErrNotFound) {
		response.ArticleNotFound(c)
		return
	}
	if err != nil {
		logging.Error(c, "request_failed", err)
		response.InternalServerError(c)
		return
	}

	c.JSON(http.StatusOK, toArticleResponse(a))
}

func (ah *ArticleHandler) ListByUserID(c *gin.Context) {
	ctx := c.Request.Context()

	authUserID, ok := getUserID(c)
	if !ok {
		return
	}

	query, ok := getArticleQuery(c)
	if !ok {
		return
	}

	listResult, err := ah.articleUsecase.ListByUserID(ctx, authUserID, query)
	if err != nil {
		logging.Error(c, "request_failed", err)
		response.InternalServerError(c)
		return
	}

	articles := make([]gin.H, 0, len(listResult.Articles))
	for _, s := range listResult.Articles {
		articles = append(articles, toArticleSummaryResponse(s))
	}

	c.JSON(http.StatusOK, gin.H{
		"articles":       articles,
		"filtered_count": uint(listResult.FilteredCount),
		"total_count":    uint(listResult.TotalCount),
	})
}

func (ah *ArticleHandler) Create(c *gin.Context) {
	ctx := c.Request.Context()

	authUserID, ok := getUserID(c)
	if !ok {
		return
	}

	var req CreateArticleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.MalformedBody(c)
		return
	}

	a, err := req.toDomain(authUserID)
	if errors.Is(err, exception.ErrInvalid) {
		response.InvalidInput(c, err)
		return
	}
	if err != nil {
		logging.Error(c, "request_failed", err)
		response.InternalServerError(c)
		return
	}

	createdArticle, err := ah.articleUsecase.Create(ctx, authUserID, a)
	if errors.Is(err, exception.ErrNotFound) {
		response.RecordNotFound(c)
		return
	}
	if err != nil {
		logging.Error(c, "request_failed", err)
		response.InternalServerError(c)
		return
	}

	c.JSON(http.StatusCreated, toArticleResponse(createdArticle))
}

func (ah *ArticleHandler) Update(c *gin.Context) {
	ctx := c.Request.Context()

	authUserID, ok := getUserID(c)
	if !ok {
		return
	}

	articleID, ok := getArticleID(c)
	if !ok {
		return
	}

	var req UpdateArticleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.MalformedBody(c)
		return
	}

	a, err := req.toDomain(authUserID)
	if errors.Is(err, exception.ErrInvalid) {
		response.InvalidInput(c, err)
		return
	}
	if err != nil {
		logging.Error(c, "request_failed", err)
		response.InternalServerError(c)
		return
	}

	updatedArticle, err := ah.articleUsecase.Update(ctx, authUserID, articleID, a)
	if errors.Is(err, exception.ErrNotFound) {
		response.ArticleNotFound(c)
		return
	}
	if err != nil {
		logging.Error(c, "request_failed", err)
		response.InternalServerError(c)
		return
	}

	c.JSON(http.StatusOK, toArticleResponse(updatedArticle))
}

func (ah *ArticleHandler) Delete(c *gin.Context) {
	ctx := c.Request.Context()

	authUserID, ok := getUserID(c)
	if !ok {
		return
	}

	articleID, ok := getArticleID(c)
	if !ok {
		return
	}

	err := ah.articleUsecase.Delete(ctx, authUserID, articleID)
	if errors.Is(err, exception.ErrNotFound) {
		response.ArticleNotFound(c)
		return
	}
	if err != nil {
		logging.Error(c, "request_failed", err)
		response.InternalServerError(c)
		return
	}

	c.Status(http.StatusNoContent)
}
