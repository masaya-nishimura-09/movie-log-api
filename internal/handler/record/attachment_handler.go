package record

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/media"
	recorddomain "github.com/masaya-nishimura-09/movie-log-api/internal/domain/record"
	"github.com/masaya-nishimura-09/movie-log-api/internal/handler/response"
	"github.com/masaya-nishimura-09/movie-log-api/internal/logging"
	recordusecase "github.com/masaya-nishimura-09/movie-log-api/internal/usecase/record"
)

func toAttachmentResponse(a *recorddomain.Attachment) gin.H {
	return gin.H{
		"attachment_id": strconv.FormatUint(uint64(a.ID), 10),
		"image_url":     string(a.ImageURL),
		"created_at":    a.CreatedAt.UTC(),
	}
}

type AttachmentHandler struct {
	attachmentUsecase recordusecase.AttachmentUsecase
}

func NewAttachmentHandler(usecase recordusecase.AttachmentUsecase) *AttachmentHandler {
	return &AttachmentHandler{attachmentUsecase: usecase}
}

func getAttachmentID(c *gin.Context) (recorddomain.AttachmentID, bool) {
	id, err := strconv.ParseUint(c.Param("attachmentId"), 10, 64)
	if err != nil {
		response.InvalidInput(c, fmt.Errorf("%w: attachment id must be a number", exception.ErrInvalid))
		return 0, false
	}
	return recorddomain.AttachmentID(id), true
}

func toMedia(data []byte) (media.Media, error) {
	mediaData, err := media.NewData(data)
	if err != nil {
		return media.Media{}, err
	}

	return media.NewMedia(mediaData)
}

func (ah *AttachmentHandler) ListByRecordID(c *gin.Context) {
	ctx := c.Request.Context()

	authUserID, ok := getUserID(c)
	if !ok {
		return
	}

	recordID, ok := getRecordID(c)
	if !ok {
		return
	}

	attachments, err := ah.attachmentUsecase.ListByRecordID(ctx, authUserID, recordID)
	if errors.Is(err, exception.ErrNotFound) {
		response.RecordNotFound(c)
		return
	}
	if err != nil {
		logging.Error(c, "request_failed", err)
		response.InternalServerError(c)
		return
	}

	res := make([]gin.H, 0, len(attachments))
	for i := range attachments {
		res = append(res, toAttachmentResponse(&attachments[i]))
	}

	c.JSON(http.StatusOK, gin.H{"attachments": res})
}

func (ah *AttachmentHandler) Create(c *gin.Context) {
	ctx := c.Request.Context()

	authUserID, ok := getUserID(c)
	if !ok {
		return
	}

	recordID, ok := getRecordID(c)
	if !ok {
		return
	}

	c.Request.Body = http.MaxBytesReader(
		c.Writer, c.Request.Body, media.MaxBytes+1024*1024,
	)
	fileHeader, err := c.FormFile("file")
	var maxBytesErr *http.MaxBytesError
	if errors.As(err, &maxBytesErr) {
		response.InvalidInput(
			c, fmt.Errorf("%w: media must be at most 5 megabytes", exception.ErrInvalid),
		)
		return
	}
	if err != nil {
		response.MalformedBody(c)
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		logging.Error(c, "request_failed", err)
		response.InternalServerError(c)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, media.MaxBytes+1))
	if err != nil {
		logging.Error(c, "request_failed", err)
		response.InternalServerError(c)
		return
	}

	m, err := toMedia(data)
	if errors.Is(err, exception.ErrInvalid) {
		response.InvalidInput(c, err)
		return
	}
	if err != nil {
		logging.Error(c, "request_failed", err)
		response.InternalServerError(c)
		return
	}

	a, err := ah.attachmentUsecase.Create(ctx, authUserID, recordID, m)
	if errors.Is(err, exception.ErrInvalid) {
		response.InvalidInput(c, err)
		return
	}
	if errors.Is(err, exception.ErrNotFound) {
		response.RecordNotFound(c)
		return
	}
	if err != nil {
		logging.Error(c, "request_failed", err)
		response.InternalServerError(c)
		return
	}

	c.JSON(http.StatusCreated, toAttachmentResponse(a))
}

func (ah *AttachmentHandler) Delete(c *gin.Context) {
	ctx := c.Request.Context()

	authUserID, ok := getUserID(c)
	if !ok {
		return
	}

	recordID, ok := getRecordID(c)
	if !ok {
		return
	}

	attachmentID, ok := getAttachmentID(c)
	if !ok {
		return
	}

	err := ah.attachmentUsecase.Delete(ctx, authUserID, recordID, attachmentID)
	if errors.Is(err, exception.ErrNotFound) {
		response.AttachmentNotFound(c)
		return
	}
	if err != nil {
		logging.Error(c, "request_failed", err)
		response.InternalServerError(c)
		return
	}

	c.Status(http.StatusNoContent)
}
