package media

import (
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/media"
	userdomain "github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
	"github.com/masaya-nishimura-09/movie-log-api/internal/handler/response"
	mediausecase "github.com/masaya-nishimura-09/movie-log-api/internal/usecase/media"
)

type MediaHandler struct {
	mediaUsecase mediausecase.Usecase
}

func NewMediaHandler(mediaUsecase mediausecase.Usecase) *MediaHandler {
	return &MediaHandler{mediaUsecase: mediaUsecase}
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

func toMedia(data []byte) (media.Media, error) {
	mediaData, err := media.NewData(data)
	if err != nil {
		return media.Media{}, err
	}

	return media.NewMedia(mediaData)
}

func (mh *MediaHandler) Upload(c *gin.Context) {
	ctx := c.Request.Context()

	authUserID, ok := getUserID(c)
	if !ok {
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.MalformedBody(c)
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		log.Println(err)
		response.InternalServerError(c)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, media.MaxBytes+1))
	if err != nil {
		log.Println(err)
		response.InternalServerError(c)
		return
	}

	m, err := toMedia(data)
	if errors.Is(err, exception.ErrInvalid) {
		response.InvalidInput(c, err)
		return
	}
	if err != nil {
		log.Println(err)
		response.InternalServerError(c)
		return
	}

	url, err := mh.mediaUsecase.Upload(ctx, authUserID, m)
	if err != nil {
		log.Println(err)
		response.InternalServerError(c)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"url": string(url)})
}
