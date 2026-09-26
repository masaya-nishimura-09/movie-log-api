package record

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	recorddomain "github.com/masaya-nishimura-09/movie-log-api/internal/domain/record"
	userdomain "github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
	"github.com/masaya-nishimura-09/movie-log-api/internal/handler/response"
	recordusecase "github.com/masaya-nishimura-09/movie-log-api/internal/usecase/record"
)

type CreditReq struct {
	PersonName string `json:"person_name"`
	CreditRole string `json:"credit_role"`
}

type RecordReq struct {
	Title       string      `json:"title"`
	ReleaseYear uint        `json:"release_year"`
	Runtime     uint        `json:"runtime"`
	Genres      []string    `json:"genres"`
	Countries   []string    `json:"countries"`
	Language    string      `json:"language"`
	Credits     []CreditReq `json:"credits"`
	PosterURL   string      `json:"poster_url"`
	WatchedAt   time.Time   `json:"watched_at"`
	Platform    string      `json:"platform"`
	Score       uint        `json:"score"`
	MoodTags    []string    `json:"mood_tags"`
	Memo        string      `json:"memo"`
}

func (r RecordReq) toDomain(userID userdomain.ID) (recorddomain.Record, error) {
	title, err := recorddomain.NewTitle(r.Title)
	if err != nil {
		return recorddomain.Record{}, err
	}
	releaseYear, err := recorddomain.NewReleaseYear(r.ReleaseYear)
	if err != nil {
		return recorddomain.Record{}, err
	}
	runtime, err := recorddomain.NewRuntime(r.Runtime)
	if err != nil {
		return recorddomain.Record{}, err
	}
	genres, err := recorddomain.NewGenres(r.Genres)
	if err != nil {
		return recorddomain.Record{}, err
	}
	countries, err := recorddomain.NewCountries(r.Countries)
	if err != nil {
		return recorddomain.Record{}, err
	}
	language, err := recorddomain.NewLanguage(r.Language)
	if err != nil {
		return recorddomain.Record{}, err
	}
	credits, err := toCredits(r.Credits)
	if err != nil {
		return recorddomain.Record{}, err
	}
	posterURL, err := recorddomain.NewPosterURL(r.PosterURL)
	if err != nil {
		return recorddomain.Record{}, err
	}
	watchedAt, err := recorddomain.NewWatchedAt(r.WatchedAt)
	if err != nil {
		return recorddomain.Record{}, err
	}
	platform, err := recorddomain.NewPlatform(r.Platform)
	if err != nil {
		return recorddomain.Record{}, err
	}
	score, err := recorddomain.NewScore(r.Score)
	if err != nil {
		return recorddomain.Record{}, err
	}
	moodTags, err := recorddomain.NewMoodTags(r.MoodTags)
	if err != nil {
		return recorddomain.Record{}, err
	}
	memo, err := recorddomain.NewMemo(r.Memo)
	if err != nil {
		return recorddomain.Record{}, err
	}

	return recorddomain.NewRecord(
		userID,
		title,
		releaseYear,
		runtime,
		genres,
		countries,
		language,
		credits,
		posterURL,
		watchedAt,
		platform,
		score,
		moodTags,
		memo,
	), nil
}

func toCredits(reqs []CreditReq) ([]recorddomain.Credit, error) {
	credits := make([]recorddomain.Credit, 0, len(reqs))

	for _, req := range reqs {
		personName, err := recorddomain.NewPersonName(req.PersonName)
		if err != nil {
			return nil, err
		}
		creditRole, err := recorddomain.NewCreditRole(req.CreditRole)
		if err != nil {
			return nil, err
		}
		credits = append(credits, recorddomain.NewCredit(personName, creditRole))
	}

	return credits, nil
}

func toResponse(r *recorddomain.Record) gin.H {
	genres := make([]string, 0, len(r.Genres))
	for _, genre := range r.Genres {
		genres = append(genres, string(genre))
	}

	countries := make([]string, 0, len(r.Countries))
	for _, country := range r.Countries {
		countries = append(countries, string(country))
	}

	credits := make([]gin.H, 0, len(r.Credits))
	for _, credit := range r.Credits {
		credits = append(credits, gin.H{
			"person_name": string(credit.PersonName),
			"credit_role": string(credit.CreditRole),
		})
	}

	moodTags := make([]string, 0, len(r.MoodTags))
	for _, moodTag := range r.MoodTags {
		moodTags = append(moodTags, string(moodTag))
	}

	return gin.H{
		"record_id":    strconv.FormatUint(uint64(r.ID), 10),
		"title":        string(r.Title),
		"release_year": uint(r.ReleaseYear),
		"runtime":      uint(r.Runtime),
		"genres":       genres,
		"countries":    countries,
		"language":     string(r.Language),
		"credits":      credits,
		"poster_url":   string(r.PosterURL),
		"watched_at":   r.WatchedAt.UTC(),
		"platform":     string(r.Platform),
		"score":        uint(r.Score),
		"mood_tags":    moodTags,
		"memo":         string(r.Memo),
	}
}

type RecordHandler struct {
	recordUsecase recordusecase.Usecase
}

func NewRecordHandler(usecase recordusecase.Usecase) *RecordHandler {
	return &RecordHandler{recordUsecase: usecase}
}

func getUserID(c *gin.Context) (userdomain.ID, bool) {
	v, exists := c.Get("userID")
	id, ok := v.(userdomain.ID)
	if !exists || !ok {
		log.Println("userID in context is missing or not of type user.ID")
		response.InternalServerError(c)
		return 0, false
	}
	return id, true
}

func getScores(c *gin.Context) ([]recorddomain.Score, bool) {
	s := c.QueryArray("scores")
	values := make([]uint, 0, len(s))

	for _, v := range s {
		value, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			response.InvalidInput(c, fmt.Errorf("%w: score must be a number", exception.ErrInvalid))
			return nil, false
		}
		values = append(values, uint(value))
	}

	ds, err := recorddomain.NewScores(values)
	if err != nil {
		response.InvalidInput(c, err)
		return nil, false
	}
	return ds, true
}

func getPlatforms(c *gin.Context) ([]recorddomain.Platform, bool) {
	p := c.QueryArray("platforms")
	dp, err := recorddomain.NewPlatforms(p)
	if err != nil {
		response.InvalidInput(c, err)
		return nil, false
	}
	return dp, true
}

func getMoodTags(c *gin.Context) ([]recorddomain.MoodTag, bool) {
	m := c.QueryArray("mood_tags")
	dm, err := recorddomain.NewMoodTags(m)
	if err != nil {
		response.InvalidInput(c, err)
		return nil, false
	}
	return dm, true
}

func getGenres(c *gin.Context) ([]recorddomain.Genre, bool) {
	g := c.QueryArray("genres")
	dg, err := recorddomain.NewGenres(g)
	if err != nil {
		response.InvalidInput(c, err)
		return nil, false
	}
	return dg, true
}

func getTitleKeyword(c *gin.Context) (recorddomain.TitleKeyword, bool) {
	t := c.Query("title")
	dt, err := recorddomain.NewTitleKeyword(t)
	if err != nil {
		response.InvalidInput(c, err)
		return "", false
	}
	return dt, true
}

func getSortField(c *gin.Context) (recorddomain.SortField, bool) {
	t := c.Query("sort_field")
	dsf, err := recorddomain.NewSortField(t)
	if err != nil {
		response.InvalidInput(c, err)
		return "", false
	}
	return dsf, true
}

func getSortOrder(c *gin.Context) (recorddomain.SortOrder, bool) {
	t := c.Query("sort_order")
	dso, err := recorddomain.NewSortOrder(t)
	if err != nil {
		response.InvalidInput(c, err)
		return "", false
	}
	return dso, true
}

func getPage(c *gin.Context) (recorddomain.Page, bool) {
	p := c.Query("page")
	if p == "" {
		p = "1"
	}
	page, err := strconv.ParseUint(p, 10, 64)
	if err != nil {
		response.InvalidInput(c, fmt.Errorf("%w: page must be a number", exception.ErrInvalid))
		return 0, false
	}

	dp, err := recorddomain.NewPage(uint(page))
	if err != nil {
		response.InvalidInput(c, err)
		return 0, false
	}
	return dp, true
}

func getPerPage(c *gin.Context) (recorddomain.PerPage, bool) {
	p := c.Query("per_page")
	if p == "" {
		p = "20"
	}
	perPage, err := strconv.ParseUint(p, 10, 64)
	if err != nil {
		response.InvalidInput(c, fmt.Errorf("%w: per page must be a number", exception.ErrInvalid))
		return 0, false
	}

	dpp, err := recorddomain.NewPerPage(uint(perPage))
	if err != nil {
		response.InvalidInput(c, err)
		return 0, false
	}
	return dpp, true
}

func getQuery(c *gin.Context) (recorddomain.Query, bool) {
	scores, ok := getScores(c)
	if !ok {
		return recorddomain.Query{}, false
	}

	platforms, ok := getPlatforms(c)
	if !ok {
		return recorddomain.Query{}, false
	}

	moodTags, ok := getMoodTags(c)
	if !ok {
		return recorddomain.Query{}, false
	}

	genres, ok := getGenres(c)
	if !ok {
		return recorddomain.Query{}, false
	}

	titleKeyword, ok := getTitleKeyword(c)
	if !ok {
		return recorddomain.Query{}, false
	}

	sortField, ok := getSortField(c)
	if !ok {
		return recorddomain.Query{}, false
	}

	sortOrder, ok := getSortOrder(c)
	if !ok {
		return recorddomain.Query{}, false
	}

	page, ok := getPage(c)
	if !ok {
		return recorddomain.Query{}, false
	}

	perPage, ok := getPerPage(c)
	if !ok {
		return recorddomain.Query{}, false
	}

	query := recorddomain.NewQuery(
		scores,
		platforms,
		moodTags,
		genres,
		titleKeyword,
		sortField,
		sortOrder,
		page,
		perPage,
	)

	return query, true
}

func getRecordID(c *gin.Context) (recorddomain.ID, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidInput(c, fmt.Errorf("%w: record id must be a number", exception.ErrInvalid))
		return 0, false
	}
	return recorddomain.ID(id), true
}

func (rh *RecordHandler) GetByID(c *gin.Context) {
	ctx := c.Request.Context()

	authUserID, ok := getUserID(c)
	if !ok {
		return
	}

	recordID, ok := getRecordID(c)
	if !ok {
		return
	}

	r, err := rh.recordUsecase.GetByID(ctx, authUserID, recordID)
	if errors.Is(err, exception.ErrNotFound) {
		response.RecordNotFound(c)
		return
	}
	if err != nil {
		log.Println(err)
		response.InternalServerError(c)
		return
	}

	c.JSON(http.StatusOK, toResponse(r))
}

func (rh *RecordHandler) ListByUserID(c *gin.Context) {
	ctx := c.Request.Context()

	authUserID, ok := getUserID(c)
	if !ok {
		return
	}

	query, ok := getQuery(c)
	if !ok {
		return
	}

	listResult, err := rh.recordUsecase.ListByUserID(ctx, authUserID, query)
	if err != nil {
		log.Println(err)
		response.InternalServerError(c)
		return
	}

	records := make([]gin.H, 0, len(listResult.Records))
	for _, r := range listResult.Records {
		records = append(records, toResponse(r))
	}

	c.JSON(http.StatusOK, gin.H{
		"records":     records,
		"total_count": uint(listResult.TotalCount),
	})
}

func (rh *RecordHandler) Create(c *gin.Context) {
	ctx := c.Request.Context()

	authUserID, ok := getUserID(c)
	if !ok {
		return
	}

	var req RecordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.MalformedBody(c)
		return
	}

	r, err := req.toDomain(authUserID)
	if errors.Is(err, exception.ErrInvalid) {
		response.InvalidInput(c, err)
		return
	}
	if err != nil {
		log.Println(err)
		response.InternalServerError(c)
		return
	}

	createdRecord, err := rh.recordUsecase.Create(ctx, authUserID, r)
	if err != nil {
		log.Println(err)
		response.InternalServerError(c)
		return
	}

	c.JSON(http.StatusCreated, toResponse(createdRecord))
}

func (rh *RecordHandler) Update(c *gin.Context) {
	ctx := c.Request.Context()

	authUserID, ok := getUserID(c)
	if !ok {
		return
	}

	recordID, ok := getRecordID(c)
	if !ok {
		return
	}

	var req RecordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.MalformedBody(c)
		return
	}

	r, err := req.toDomain(authUserID)
	if errors.Is(err, exception.ErrInvalid) {
		response.InvalidInput(c, err)
		return
	}
	if err != nil {
		log.Println(err)
		response.InternalServerError(c)
		return
	}

	updatedRecord, err := rh.recordUsecase.Update(ctx, authUserID, recordID, r)
	if errors.Is(err, exception.ErrNotFound) {
		response.RecordNotFound(c)
		return
	}
	if err != nil {
		log.Println(err)
		response.InternalServerError(c)
		return
	}

	c.JSON(http.StatusOK, toResponse(updatedRecord))
}

func (rh *RecordHandler) Delete(c *gin.Context) {
	ctx := c.Request.Context()

	authUserID, ok := getUserID(c)
	if !ok {
		return
	}

	recordID, ok := getRecordID(c)
	if !ok {
		return
	}

	err := rh.recordUsecase.Delete(ctx, authUserID, recordID)
	if errors.Is(err, exception.ErrNotFound) {
		response.RecordNotFound(c)
		return
	}
	if err != nil {
		log.Println(err)
		response.InternalServerError(c)
		return
	}

	c.Status(http.StatusNoContent)
}
