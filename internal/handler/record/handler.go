package record

import (
	"errors"
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
		"watched_at":   r.WatchedAt,
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

func getRecordID(c *gin.Context) (recorddomain.ID, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.RecordNotFound(c)
		return 0, false
	}
	return recorddomain.ID(id), true
}

func (rh *RecordHandler) CreateRecord(c *gin.Context) {
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

	createdRecord, err := rh.recordUsecase.CreateRecord(ctx, authUserID, r)
	if err != nil {
		log.Println(err)
		response.InternalServerError(c)
		return
	}

	c.JSON(http.StatusCreated, toResponse(createdRecord))
}

func (rh *RecordHandler) GetRecord(c *gin.Context) {
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

func (rh *RecordHandler) ListRecords(c *gin.Context) {
	ctx := c.Request.Context()

	authUserID, ok := getUserID(c)
	if !ok {
		return
	}

	records, err := rh.recordUsecase.ListByUserID(ctx, authUserID)
	if err != nil {
		log.Println(err)
		response.InternalServerError(c)
		return
	}

	responses := make([]gin.H, 0, len(records))
	for _, r := range records {
		responses = append(responses, toResponse(r))
	}

	c.JSON(http.StatusOK, gin.H{"records": responses})
}

func (rh *RecordHandler) UpdateRecord(c *gin.Context) {
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

	updatedRecord, err := rh.recordUsecase.UpdateRecord(ctx, authUserID, recordID, r)
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

func (rh *RecordHandler) DeleteRecord(c *gin.Context) {
	ctx := c.Request.Context()

	authUserID, ok := getUserID(c)
	if !ok {
		return
	}

	recordID, ok := getRecordID(c)
	if !ok {
		return
	}

	err := rh.recordUsecase.DeleteRecord(ctx, authUserID, recordID)
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
